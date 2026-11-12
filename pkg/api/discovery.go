package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// SweepOptions tunes the reliability and pace of quest discovery.
type SweepOptions struct {
	MinInterval    time.Duration
	MaxInterval    time.Duration
	MaxAttempts    int
	RetryBaseDelay time.Duration
	ErrorPenalty   time.Duration
	RetryBuffer    time.Duration
}

// DefaultSweepOptions returns conservative defaults that respect the tight
// per-user rate limit Discord applies to GET /quests/@me.
func DefaultSweepOptions() SweepOptions {
	return SweepOptions{
		MinInterval:    3 * time.Second,
		MaxInterval:    15 * time.Second,
		MaxAttempts:    3,
		RetryBaseDelay: 1500 * time.Millisecond,
		ErrorPenalty:   time.Second,
		RetryBuffer:    1500 * time.Millisecond,
	}
}

func (o SweepOptions) withDefaults() SweepOptions {
	def := DefaultSweepOptions()
	if o.MinInterval <= 0 {
		o.MinInterval = def.MinInterval
	}
	if o.MaxInterval <= 0 {
		o.MaxInterval = def.MaxInterval
	}
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = def.MaxAttempts
	}
	if o.RetryBaseDelay <= 0 {
		o.RetryBaseDelay = def.RetryBaseDelay
	}
	if o.ErrorPenalty <= 0 {
		o.ErrorPenalty = def.ErrorPenalty
	}
	if o.RetryBuffer <= 0 {
		o.RetryBuffer = def.RetryBuffer
	}
	if o.MaxInterval < o.MinInterval {
		o.MaxInterval = o.MinInterval
	}
	return o
}

// TargetReport describes the outcome of one sweep target.
type TargetReport struct {
	Code     string
	Locale   string
	Timezone string
	Proxy    string
	Attempts int
	Status   int
	Ok       bool
	Err      string
	Quests   int
	Excluded int
	Latency  time.Duration
}

// SweepResult is the union of every successful target in a discovery sweep.
type SweepResult struct {
	Quests                 []Quest
	Excluded               []ExcludedQuest
	EnrollmentBlockedUntil *string
	AccessSuspendedUntil   *string
	Targets                []TargetReport
	Requests               int
	Succeeded              int
	Failed                 int
	Duration               time.Duration
}

// AccessState summarizes the account-level quest gates returned by Discord.
type AccessState struct {
	AccessSuspendedUntil   time.Time
	AccessSuspended        bool
	EnrollmentBlockedUntil time.Time
	EnrollmentBlocked      bool
}

// AccessState evaluates the account-level gate timestamps at a point in time.
func (r *SweepResult) AccessState(now time.Time) AccessState {
	state := AccessState{}
	if r == nil {
		return state
	}
	if until, future := FutureTimestamp(r.AccessSuspendedUntil, now); future {
		state.AccessSuspended = true
		state.AccessSuspendedUntil = until
	}
	if until, future := FutureTimestamp(r.EnrollmentBlockedUntil, now); future {
		state.EnrollmentBlocked = true
		state.EnrollmentBlockedUntil = until
	}
	return state
}

// FailedTargets returns the codes of targets that produced no data.
func (r *SweepResult) FailedTargets() []string {
	if r == nil {
		return nil
	}
	failed := make([]string, 0, r.Failed)
	for _, target := range r.Targets {
		if !target.Ok {
			failed = append(failed, target.Code)
		}
	}
	return failed
}

// ExcludedCount returns how many distinct quests Discord excluded for the account.
func (r *SweepResult) ExcludedCount() int {
	if r == nil {
		return 0
	}
	return len(r.Excluded)
}

// HasProxyTargets reports whether any target used a regional proxy.
func (r *SweepResult) HasProxyTargets() bool {
	if r == nil {
		return false
	}
	for _, target := range r.Targets {
		if target.Proxy != "" {
			return true
		}
	}
	return false
}

func (r *SweepResult) summarize() {
	r.Succeeded = 0
	r.Failed = 0
	for _, target := range r.Targets {
		if target.Ok {
			r.Succeeded++
		} else {
			r.Failed++
		}
	}
}

func (r *SweepResult) failureSummary() string {
	parts := make([]string, 0, len(r.Targets))
	for _, target := range r.Targets {
		if target.Ok {
			continue
		}
		detail := target.Err
		if detail == "" {
			detail = "unknown error"
		}
		parts = append(parts, target.Code+": "+detail)
	}
	return strings.Join(parts, "; ")
}

// FutureTimestamp parses an ISO8601 quest gate timestamp and reports whether it
// is still in the future.
func FutureTimestamp(raw *string, now time.Time) (time.Time, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, parsed.After(now)
}

// sweepPacer serializes quest requests and adapts its floor to HTTP 429
// feedback so a multi-region sweep does not silently lose targets.
type sweepPacer struct {
	floor  time.Duration
	min    time.Duration
	max    time.Duration
	buffer time.Duration
	next   time.Time
}

func newSweepPacer(opts SweepOptions) *sweepPacer {
	return &sweepPacer{floor: opts.MinInterval, min: opts.MinInterval, max: opts.MaxInterval, buffer: opts.RetryBuffer}
}

func (p *sweepPacer) wait(ctx context.Context) error {
	if p == nil {
		return nil
	}
	if delay := time.Until(p.next); delay > 0 {
		if !sleepContext(ctx, delay) {
			return ctx.Err()
		}
	}
	p.next = time.Now().Add(p.floor)
	return nil
}

func (p *sweepPacer) success() {
	if p == nil || p.floor <= p.min {
		return
	}
	p.floor = p.floor * 85 / 100
	if p.floor < p.min {
		p.floor = p.min
	}
}

func (p *sweepPacer) penalty(retryAfter time.Duration) {
	if p == nil {
		return
	}
	raised := p.floor * 3 / 2
	if retryAfter > 0 {
		needed := retryAfter + p.buffer
		if needed > raised {
			raised = needed
		}
	}
	if raised < p.min {
		raised = p.min
	}
	if p.max > 0 && raised > p.max {
		raised = p.max
	}
	p.floor = raised
	p.next = time.Now().Add(p.floor)
}

func sleepContext(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func sweepJitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(max)))
}

// doQuestFetch performs exactly one GET /quests/@me request.
func (c *Client) doQuestFetch(ctx context.Context) (*QuestResponse, int, []byte, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/quests/@me", nil)
	if err != nil {
		return nil, 0, nil, 0, fmt.Errorf("create quests request failed: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, nil, 0, fmt.Errorf("request /quests/@me failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, nil, 0, fmt.Errorf("failed reading /quests/@me response: %w", readErr)
	}
	if resp.StatusCode != http.StatusOK {
		retryAfter := parseRetryAfter(resp, bodyBytes)
		return nil, resp.StatusCode, bodyBytes, retryAfter, nil
	}
	payload, decodeErr := decodeQuestPayload(bodyBytes)
	if decodeErr != nil {
		return nil, resp.StatusCode, bodyBytes, 0, decodeErr
	}
	return payload, resp.StatusCode, bodyBytes, 0, nil
}

// decodeQuestPayload accepts both the current object payload and the legacy
// bare-array payload some older builds returned.
func decodeQuestPayload(body []byte) (*QuestResponse, error) {
	var payload QuestResponse
	objectErr := json.Unmarshal(body, &payload)
	if objectErr == nil && (payload.Quests != nil || payload.ExcludedQuests != nil ||
		payload.QuestEnrollmentBlockedUntil != nil || payload.QuestAccessSuspendedUntil != nil) {
		return &payload, nil
	}
	var legacy []Quest
	if err := json.Unmarshal(body, &legacy); err == nil && legacy != nil {
		return &QuestResponse{Quests: legacy}, nil
	}
	if objectErr != nil {
		return nil, fmt.Errorf("unable to decode quest response structure: %w", objectErr)
	}
	// A 200 whose body is not a quest payload (schema change, gateway/error page
	// wrapped in JSON) must not be mistaken for an empty inventory.
	return nil, fmt.Errorf("quest response did not contain a quests field")
}

// mergeSweepPayload unions one target response into the sweep result. The first
// target to see a quest owns its traffic metadata and therefore its proxy
// binding; later targets only add provenance.
func mergeSweepPayload(result *SweepResult, payload *QuestResponse, target RegionTarget) {
	if result == nil || payload == nil {
		return
	}
	index := make(map[string]int, len(result.Quests))
	for i := range result.Quests {
		index[result.Quests[i].ID] = i
	}
	for _, quest := range payload.Quests {
		if quest.ID == "" {
			continue
		}
		if i, ok := index[quest.ID]; ok {
			result.Quests[i].SeenIn = appendRegion(result.Quests[i].SeenIn, target.Code)
			continue
		}
		quest.DiscoveredRegion = target.Code
		quest.DiscoveredLocale = target.Locale
		quest.DiscoveredTimezone = target.Timezone
		quest.RegionProxied = target.ProxyURL != ""
		quest.SeenIn = appendRegion(quest.SeenIn, target.Code)
		index[quest.ID] = len(result.Quests)
		result.Quests = append(result.Quests, quest)
	}

	excludedIndex := make(map[string]int, len(result.Excluded))
	for i := range result.Excluded {
		excludedIndex[result.Excluded[i].ID] = i
	}
	for _, excluded := range payload.ExcludedQuests {
		if excluded.ID == "" {
			continue
		}
		if _, ok := excludedIndex[excluded.ID]; ok {
			continue
		}
		excluded.DiscoveredRegion = target.Code
		excludedIndex[excluded.ID] = len(result.Excluded)
		result.Excluded = append(result.Excluded, excluded)
	}

	result.EnrollmentBlockedUntil = laterTimestamp(result.EnrollmentBlockedUntil, payload.QuestEnrollmentBlockedUntil)
	result.AccessSuspendedUntil = laterTimestamp(result.AccessSuspendedUntil, payload.QuestAccessSuspendedUntil)
}

func appendRegion(regions []string, code string) []string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return regions
	}
	for _, existing := range regions {
		if existing == code {
			return regions
		}
	}
	return append(regions, code)
}

// SweepQuests queries every target, unions the responses, deduplicates by quest
// ID and reports per-target failures instead of silently dropping a region.
func (c *Client) SweepQuests(ctx context.Context, targets []RegionTarget, opts SweepOptions) (*SweepResult, error) {
	opts = opts.withDefaults()
	startedAt := time.Now()
	result := &SweepResult{Quests: []Quest{}, Excluded: []ExcludedQuest{}}
	if len(targets) == 0 {
		return result, fmt.Errorf("quest sweep has no targets")
	}

	seen := make(map[string]bool, len(targets))
	unique := make([]RegionTarget, 0, len(targets))
	for _, target := range targets {
		code := strings.ToUpper(strings.TrimSpace(target.Code))
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		target.Code = code
		if target.Locale == "" {
			target.Locale = "en-US"
		}
		if target.Timezone == "" {
			target.Timezone = "America/New_York"
		}
		unique = append(unique, target)
		result.Targets = append(result.Targets, TargetReport{
			Code:     target.Code,
			Locale:   target.Locale,
			Timezone: target.Timezone,
			Proxy:    RedactProxyURL(target.ProxyURL),
		})
	}
	if len(unique) == 0 {
		return result, fmt.Errorf("quest sweep has no usable targets")
	}

	for i := range unique {
		if unique[i].ProxyURL == "" || c.HasRegionProxy(unique[i].Code) {
			continue
		}
		if err := c.SetRegionProxy(unique[i].Code, unique[i].ProxyURL); err != nil {
			result.Targets[i].Err = "proxy setup failed: " + err.Error()
		}
	}

	pacer := newSweepPacer(opts)

	attemptTarget := func(index int, allowRetry bool) bool {
		target := unique[index]
		report := &result.Targets[index]
		if report.Attempts == 0 && report.Err != "" {
			// Proxy registration failed before any request.
			return false
		}
		attempts := opts.MaxAttempts
		if !allowRetry {
			attempts = 1
		}
		for attempt := 1; attempt <= attempts; attempt++ {
			if err := ctx.Err(); err != nil {
				report.Err = err.Error()
				return true
			}
			if attempt > 1 {
				backoff := opts.RetryBaseDelay * time.Duration(1<<(attempt-2))
				backoff += sweepJitter(800 * time.Millisecond)
				if !sleepContext(ctx, backoff) {
					report.Err = ctx.Err().Error()
					return true
				}
			}
			if err := pacer.wait(ctx); err != nil {
				report.Err = err.Error()
				return true
			}

			requestStarted := time.Now()
			payload, status, body, retryAfter, err := c.doQuestFetch(WithRegionInfo(ctx, target.Info()))
			result.Requests++
			report.Attempts++
			report.Latency = time.Since(requestStarted)
			report.Status = status

			if err != nil {
				report.Err = err.Error()
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return true
				}
				pacer.penalty(opts.ErrorPenalty)
				continue
			}
			if status == http.StatusOK && payload != nil {
				pacer.success()
				report.Ok = true
				report.Err = ""
				report.Quests = len(payload.Quests)
				report.Excluded = len(payload.ExcludedQuests)
				mergeSweepPayload(result, payload, target)
				return false
			}

			switch status {
			case http.StatusUnauthorized:
				report.Err = "HTTP 401: token is invalid or revoked"
				return true
			case http.StatusForbidden:
				// A proxy blocked by Cloudflare/proxy auth, or a region-specific
				// block, must not kill the remaining regions.
				report.Err = "HTTP 403: request blocked (Cloudflare challenge, proxy auth or account restriction)"
				return false
			case http.StatusTooManyRequests:
				if retryAfter <= 0 {
					retryAfter = 3 * time.Second
				}
				pacer.penalty(retryAfter)
				report.Err = fmt.Sprintf("HTTP 429: rate limited (retry_after %s)", retryAfter.Round(100*time.Millisecond))
				continue
			}

			report.Err = fmt.Sprintf("HTTP %d: %s", status, truncateBody(body))
			if status >= 500 || status == http.StatusRequestTimeout {
				pacer.penalty(opts.ErrorPenalty)
				continue
			}
			return false
		}
		if report.Err == "" {
			report.Err = "target failed after retries"
		}
		return false
	}

	failed := make([]int, 0, len(unique))
	for i := range unique {
		if result.Targets[i].Attempts == 0 && result.Targets[i].Err != "" {
			continue
		}
		fatal := attemptTarget(i, true)
		if fatal {
			abortErr := redactTargetError(result.Targets[i].Err, unique[i])
			for j := range result.Targets {
				if j == i || result.Targets[j].Attempts > 0 || result.Targets[j].Err != "" {
					continue
				}
				result.Targets[j].Err = "not attempted: sweep aborted (" + abortErr + ")"
			}
			result.Duration = time.Since(startedAt)
			result.summarize()
			return result, fmt.Errorf("quest discovery aborted: %s", abortErr)
		}
		if !result.Targets[i].Ok {
			failed = append(failed, i)
		}
	}

	// Recovery pass: a 429 storm during the first pass usually clears once the
	// other targets have drained the bucket, so failed targets get one more try.
	for _, index := range failed {
		if ctx.Err() != nil {
			break
		}
		if !sleepContext(ctx, opts.RetryBaseDelay+sweepJitter(time.Second)) {
			break
		}
		attemptTarget(index, false)
	}

	for i := range result.Targets {
		result.Targets[i].Err = redactTargetError(result.Targets[i].Err, unique[i])
	}
	result.Duration = time.Since(startedAt)
	result.summarize()
	if result.Succeeded == 0 {
		return result, fmt.Errorf("quest discovery failed on every target (%s)", result.failureSummary())
	}
	return result, nil
}

// laterTimestamp keeps the later of two ISO8601 gate timestamps.
func laterTimestamp(current *string, candidate *string) *string {
	if candidate == nil {
		return current
	}
	if current == nil {
		value := *candidate
		return &value
	}
	currentTime, currentErr := time.Parse(time.RFC3339, *current)
	candidateTime, candidateErr := time.Parse(time.RFC3339, *candidate)
	if candidateErr != nil {
		return current
	}
	if currentErr != nil || candidateTime.After(currentTime) {
		value := *candidate
		return &value
	}
	return current
}

// redactTargetError removes proxy credentials from any error that embeds the
// proxy URL before it reaches logs.
func redactTargetError(message string, target RegionTarget) string {
	if message == "" || target.ProxyURL == "" {
		return message
	}
	return strings.ReplaceAll(message, target.ProxyURL, RedactProxyURL(target.ProxyURL))
}
