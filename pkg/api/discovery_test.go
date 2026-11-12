package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newTestSweepClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client := NewClient("test-token", 500000)
	client.SetAPIBaseForTesting(serverURL)
	return client
}

func fastSweepOptions() SweepOptions {
	return SweepOptions{
		MinInterval:    time.Millisecond,
		MaxInterval:    2 * time.Millisecond,
		MaxAttempts:    2,
		RetryBaseDelay: time.Millisecond,
		ErrorPenalty:   time.Millisecond,
		RetryBuffer:    time.Millisecond,
	}
}

func TestSweepQuestsUnionDedupAndRegions(t *testing.T) {
	payloads := map[string]string{
		"en-US": `{"quests":[{"id":"q1"},{"id":"q2"}],"excluded_quests":[{"id":"e1"}]}`,
		"ja":    `{"quests":[{"id":"q2"},{"id":"q3"}],"excluded_quests":[{"id":"e2","replacement_id":"q1"}]}`,
		"vi":    `{"quests":[],"excluded_quests":[]}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := payloads[r.Header.Get("X-Discord-Locale")]
		if !ok {
			body = `{"quests":[]}`
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	targets := []RegionTarget{
		{Code: "LOCAL", Locale: "en-US", Timezone: "America/New_York"},
		{Code: "L:ja", Locale: "ja", Timezone: "Asia/Tokyo"},
		{Code: "L:vi", Locale: "vi", Timezone: "Asia/Ho_Chi_Minh"},
	}
	result, err := client.SweepQuests(context.Background(), targets, fastSweepOptions())
	if err != nil {
		t.Fatalf("unexpected sweep error: %v", err)
	}
	if len(result.Quests) != 3 {
		t.Fatalf("expected union of 3 quests, got %d: %+v", len(result.Quests), result.Quests)
	}
	byID := make(map[string]Quest, len(result.Quests))
	for _, quest := range result.Quests {
		byID[quest.ID] = quest
	}
	if byID["q2"].DiscoveredRegion != "LOCAL" {
		t.Fatalf("expected q2 to be owned by the first region, got %s", byID["q2"].DiscoveredRegion)
	}
	if strings.Join(byID["q2"].SeenIn, ",") != "LOCAL,L:JA" {
		t.Fatalf("expected q2 provenance LOCAL,L:JA, got %v", byID["q2"].SeenIn)
	}
	if byID["q3"].DiscoveredRegion != "L:JA" {
		t.Fatalf("expected q3 region L:JA, got %s", byID["q3"].DiscoveredRegion)
	}
	if byID["q3"].DiscoveredLocale != "ja" || byID["q3"].DiscoveredTimezone != "Asia/Tokyo" {
		t.Fatalf("expected q3 locale metadata to be preserved, got %+v", byID["q3"])
	}
	if len(result.Excluded) != 2 {
		t.Fatalf("expected 2 distinct excluded quests, got %d", len(result.Excluded))
	}
	if result.Excluded[1].ReplacementID == nil || *result.Excluded[1].ReplacementID != "q1" {
		t.Fatalf("expected replacement_id q1 on excluded quest, got %+v", result.Excluded[1])
	}
	if result.Succeeded != 3 || result.Failed != 0 {
		t.Fatalf("expected 3 successful targets, got %d ok / %d failed", result.Succeeded, result.Failed)
	}
	if result.Requests != 3 {
		t.Fatalf("expected 3 requests, got %d", result.Requests)
	}
}

func TestSweepQuestsRetriesTransientHTTP5xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
			return
		}
		_, _ = w.Write([]byte(`{"quests":[{"id":"q1"}]}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	result, err := client.SweepQuests(context.Background(), []RegionTarget{{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"}}, fastSweepOptions())
	if err != nil {
		t.Fatalf("unexpected sweep error: %v", err)
	}
	if len(result.Quests) != 1 || result.Quests[0].ID != "q1" {
		t.Fatalf("expected q1 after a transient 500, got %+v", result.Quests)
	}
	if result.Targets[0].Attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", result.Targets[0].Attempts)
	}
}

func TestSweepQuestsRetriesHTTP429(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limited","retry_after":0.01}`))
			return
		}
		_, _ = w.Write([]byte(`{"quests":[{"id":"q-after-429"}]}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	result, err := client.SweepQuests(context.Background(), []RegionTarget{{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"}}, fastSweepOptions())
	if err != nil {
		t.Fatalf("unexpected sweep error: %v", err)
	}
	if len(result.Quests) != 1 || result.Quests[0].ID != "q-after-429" {
		t.Fatalf("expected discovery to recover from 429, got %+v", result.Quests)
	}
	if result.Targets[0].Attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", result.Targets[0].Attempts)
	}
}

func TestSweepQuestsReportsPartialFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Discord-Locale") == "de" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"always broken"}`))
			return
		}
		questID := "q-" + r.Header.Get("X-Discord-Locale")
		_, _ = w.Write([]byte(`{"quests":[{"id":"` + questID + `"}]}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	opts := fastSweepOptions()
	opts.MaxAttempts = 1
	targets := []RegionTarget{
		{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"},
		{Code: "L:de", Locale: "de", Timezone: "Europe/Berlin"},
		{Code: "L:ja", Locale: "ja", Timezone: "Asia/Tokyo"},
	}
	result, err := client.SweepQuests(context.Background(), targets, opts)
	if err != nil {
		t.Fatalf("partial failure must not fail the whole sweep: %v", err)
	}
	if result.Succeeded != 2 || result.Failed != 1 {
		t.Fatalf("expected 2 ok / 1 failed, got %d / %d", result.Succeeded, result.Failed)
	}
	if len(result.Quests) != 2 {
		t.Fatalf("expected 2 quests from healthy targets, got %d", len(result.Quests))
	}
	failed := result.FailedTargets()
	if len(failed) != 1 || failed[0] != "L:DE" {
		t.Fatalf("expected L:DE in the failure report, got %v", failed)
	}
	for _, target := range result.Targets {
		if target.Code == "L:DE" && target.Attempts < 2 {
			t.Fatalf("failed target should be retried in the recovery pass, attempts=%d", target.Attempts)
		}
	}
}

func TestSweepQuestsAllTargetsFailingReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"message":"bad gateway"}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	opts := fastSweepOptions()
	opts.MaxAttempts = 1
	result, err := client.SweepQuests(context.Background(), []RegionTarget{{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"}}, opts)
	if err == nil {
		t.Fatal("expected an error when every target fails")
	}
	if result == nil || result.Requests == 0 {
		t.Fatalf("expected a partial result with request accounting, got %+v", result)
	}
}

func TestSweepQuestsHandlesEmptyQuestListAndGates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"quests":[],"excluded_quests":[{"id":"e1"}],"quest_access_suspended_until":"2099-01-02T03:04:05Z","quest_enrollment_blocked_until":"2099-01-02T03:04:05Z"}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	result, err := client.SweepQuests(context.Background(), []RegionTarget{{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"}}, fastSweepOptions())
	if err != nil {
		t.Fatalf("an empty quest list is a valid response: %v", err)
	}
	if len(result.Quests) != 0 || len(result.Excluded) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.AccessSuspendedUntil == nil || result.EnrollmentBlockedUntil == nil {
		t.Fatal("expected both gate timestamps to be parsed")
	}
	state := result.AccessState(time.Now().UTC())
	if !state.AccessSuspended || !state.EnrollmentBlocked {
		t.Fatalf("expected both gates to be reported as future blocks, got %+v", state)
	}
}

func TestFutureTimestamp(t *testing.T) {
	future := "2099-01-02T03:04:05Z"
	past := "2000-01-02T03:04:05Z"
	now := time.Now().UTC()
	if _, ok := FutureTimestamp(nil, now); ok {
		t.Fatal("nil timestamp must not be future")
	}
	if _, ok := FutureTimestamp(&future, now); !ok {
		t.Fatal("expected future timestamp")
	}
	if _, ok := FutureTimestamp(&past, now); ok {
		t.Fatal("expected past timestamp to be ignored")
	}
	invalid := "not-a-timestamp"
	if _, ok := FutureTimestamp(&invalid, now); ok {
		t.Fatal("expected invalid timestamp to be ignored")
	}
}

func TestDecodeQuestPayloadRejectsNonQuestJSON(t *testing.T) {
	for _, body := range []string{`{}`, `null`, `{"message":"blocked"}`} {
		if payload, err := decodeQuestPayload([]byte(body)); err == nil {
			t.Fatalf("body %q must not be accepted as an empty inventory: %+v", body, payload)
		}
	}
	if payload, err := decodeQuestPayload([]byte(`{"quests":[]}`)); err != nil || payload == nil {
		t.Fatalf("an explicit empty quests array is valid: payload=%+v err=%v", payload, err)
	}
	if payload, err := decodeQuestPayload([]byte(`[]`)); err != nil || payload == nil {
		t.Fatalf("a legacy empty array is valid: payload=%+v err=%v", payload, err)
	}
}

func TestSweepQuests403DoesNotAbortRemainingTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Discord-Locale") == "de" {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"blocked"}`))
			return
		}
		questID := "q-" + r.Header.Get("X-Discord-Locale")
		_, _ = w.Write([]byte(`{"quests":[{"id":"` + questID + `"}]}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	opts := fastSweepOptions()
	opts.MaxAttempts = 1
	result, err := client.SweepQuests(context.Background(), []RegionTarget{
		{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"},
		{Code: "L:de", Locale: "de", Timezone: "Europe/Berlin"},
		{Code: "L:ja", Locale: "ja", Timezone: "Asia/Tokyo"},
	}, opts)
	if err != nil {
		t.Fatalf("a 403 on one target must not abort the sweep: %v", err)
	}
	if result.Succeeded != 2 || result.Failed != 1 {
		t.Fatalf("expected 2 ok / 1 failed after the 403, got %d / %d", result.Succeeded, result.Failed)
	}
	for _, target := range result.Targets {
		if target.Code == "L:JA" && !target.Ok {
			t.Fatalf("targets after the 403 were not attempted: %+v", target)
		}
	}
}

func TestSweepQuests401MarksUnattemptedTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"401: Unauthorized"}`))
	}))
	defer server.Close()

	client := newTestSweepClient(t, server.URL)
	opts := fastSweepOptions()
	opts.MaxAttempts = 1
	result, err := client.SweepQuests(context.Background(), []RegionTarget{
		{Code: "LOCAL", Locale: "en-US", Timezone: "UTC"},
		{Code: "L:ja", Locale: "ja", Timezone: "Asia/Tokyo"},
	}, opts)
	if err == nil {
		t.Fatal("expected a 401 to abort the sweep")
	}
	if result == nil || len(result.Targets) != 2 {
		t.Fatalf("expected both targets in the report, got %+v", result)
	}
	if result.Targets[1].Attempts != 0 || !strings.Contains(result.Targets[1].Err, "not attempted") {
		t.Fatalf("unattempted target should be marked, got %+v", result.Targets[1])
	}
}

func TestRedactTargetError(t *testing.T) {
	target := RegionTarget{Code: "US", ProxyURL: "socks5://user:secret@proxy.example:1080"}
	message := "request failed for socks5://user:secret@proxy.example:1080 (auth rejected)"
	redacted := redactTargetError(message, target)
	if strings.Contains(redacted, "secret") {
		t.Fatalf("proxy credentials leaked: %s", redacted)
	}
	if !strings.Contains(redacted, "socks5://proxy.example:1080") {
		t.Fatalf("redacted message lost the proxy host: %s", redacted)
	}
	if got := redactTargetError("plain error", RegionTarget{Code: "LOCAL"}); got != "plain error" {
		t.Fatalf("direct target error must pass through, got %s", got)
	}
}
