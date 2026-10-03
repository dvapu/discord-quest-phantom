package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseRetryAfterCases(t *testing.T) {
	// Case 1: JSON body with retry_after
	resp1 := &http.Response{Header: make(http.Header)}
	body1 := []byte(`{"message": "The resource is being rate limited.", "retry_after": 2.5, "global": false}`)
	dur1 := parseRetryAfter(resp1, body1)
	if dur1 != 2500*time.Millisecond {
		t.Errorf("expected 2500ms, got %v", dur1)
	}

	// Case 2: Header Retry-After fallback
	resp2 := &http.Response{Header: http.Header{"Retry-After": []string{"3.2"}}}
	body2 := []byte(`{"message": "Rate limited without retry_after in body"}`)
	dur2 := parseRetryAfter(resp2, body2)
	if dur2 != 3200*time.Millisecond {
		t.Errorf("expected 3200ms, got %v", dur2)
	}

	// Case 3: Default fallback (5s)
	resp3 := &http.Response{Header: make(http.Header)}
	body3 := []byte(`non-json response`)
	dur3 := parseRetryAfter(resp3, body3)
	if dur3 != 5*time.Second {
		t.Errorf("expected 5s default, got %v", dur3)
	}
}

func TestValidateTokenMockHTTP401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "401: Unauthorized", "code": 0}`))
	}))
	defer server.Close()

	client := NewClient("fake_token", 504649)
	client.apiBase = server.URL

	user, err := client.ValidateToken(context.Background())
	if err == nil {
		t.Errorf("expected error on 401, got user %v", user)
	}
	if user != nil {
		t.Errorf("expected nil user on 401, got %v", user)
	}
}

func TestValidateTokenMockHTTP429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message": "Rate limited", "retry_after": 0.1}`))
	}))
	defer server.Close()

	client := NewClient("fake_token", 504649)
	client.apiBase = server.URL

	user, err := client.ValidateToken(context.Background())
	if err == nil {
		t.Errorf("expected error on 429, got user %v", user)
	}
}

func TestFetchQuestsRetryOnHTTP429(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := atomic.AddInt32(&attempts, 1)
		if att == 1 {
			// First attempt returns 429 with 0.1s retry_after
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message": "Rate limited", "retry_after": 0.1}`))
			return
		}
		// Second attempt succeeds
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"quests": [{"id": "quest_retry_success"}]}`))
	}))
	defer server.Close()

	client := NewClient("test_token", 504649)
	client.apiBase = server.URL

	quests, err := client.FetchQuests(context.Background())
	if err != nil {
		t.Fatalf("expected successful retry, got error: %v", err)
	}
	if len(quests) != 1 || quests[0].ID != "quest_retry_success" {
		t.Errorf("expected 1 quest with ID quest_retry_success, got %v", quests)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestEnrollQuestRetryOnHTTP429(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		att := atomic.AddInt32(&attempts, 1)
		if att == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message": "Rate limited", "retry_after": 0.1}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("test_token", 504649)
	client.apiBase = server.URL

	err := client.EnrollQuest(context.Background(), Quest{ID: "quest_enroll_test"})
	if err != nil {
		t.Fatalf("expected successful enroll on retry, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
