package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client handles authenticated requests to Discord REST API endpoints.
type Client struct {
	httpClient  *http.Client
	transport   *RegionTransport
	apiBase     string
	token       string
	timezone    string
	buildNumber int
	locale      string
	fp          Fingerprint
	stateMu     sync.RWMutex
	fpMu        sync.RWMutex
	propsMu     sync.Mutex
	propsCache  map[string]string
}

// BuildSuperPropertiesWithLocale produces a base64-encoded client fingerprint
// JSON string for a specific locale using the default device fingerprint.
func BuildSuperPropertiesWithLocale(buildNumber int, locale string) string {
	return DefaultFingerprint().SuperProperties(buildNumber, locale)
}

// BuildSuperProperties produces a base64-encoded client fingerprint JSON string.
func BuildSuperProperties(buildNumber int) string {
	return BuildSuperPropertiesWithLocale(buildNumber, "en-US")
}

// FetchLatestBuildNumber scrapes Discord CDN to discover the current client_build_number.
func FetchLatestBuildNumber() int {
	const fallback = 626571
	client := &http.Client{Timeout: 15 * time.Second}
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

	req, err := http.NewRequest("GET", "https://discord.com/app", nil)
	if err != nil {
		return fallback
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fallback
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fallback
	}
	bodyStr := string(bodyBytes)

	reAsset := regexp.MustCompile(`/assets/([a-f0-9]+)\.js`)
	matches := reAsset.FindAllStringSubmatch(bodyStr, -1)
	if len(matches) == 0 {
		return fallback
	}

	reBuild := regexp.MustCompile(`buildNumber["\s:]+["\s]*(\d{5,7})`)
	startIdx := len(matches) - 5
	if startIdx < 0 {
		startIdx = 0
	}

	for i := len(matches) - 1; i >= startIdx; i-- {
		assetHash := matches[i][1]
		assetReq, err := http.NewRequest("GET", "https://discord.com/assets/"+assetHash+".js", nil)
		if err != nil {
			continue
		}
		assetReq.Header.Set("User-Agent", ua)

		assetResp, err := client.Do(assetReq)
		if err != nil || assetResp.StatusCode != http.StatusOK {
			continue
		}

		assetBody, err := io.ReadAll(assetResp.Body)
		assetResp.Body.Close()
		if err != nil {
			continue
		}

		if buildMatch := reBuild.FindSubmatch(assetBody); len(buildMatch) > 1 {
			if bn, err := strconv.Atoi(string(buildMatch[1])); err == nil && bn > 0 {
				return bn
			}
		}
	}

	return fallback
}

// NewClient initializes a new Discord API client with realistic headers and specified build number.
func NewClient(token string, buildNumber int) *Client {
	if buildNumber <= 0 {
		buildNumber = 626571
	}
	regionTransport := NewRegionTransport()
	fp := DefaultFingerprint()
	return &Client{
		httpClient:  &http.Client{Timeout: 20 * time.Second, Transport: regionTransport},
		transport:   regionTransport,
		apiBase:     "https://discord.com/api/v9",
		token:       token,
		timezone:    "America/New_York",
		locale:      "en-US",
		fp:          fp,
		buildNumber: buildNumber,
		propsCache:  make(map[string]string),
	}
}

// SetFingerprint replaces the device identity used by every authenticated
// request. The user agent, X-Super-Properties and all derived fields are
// re-derived together so they can never disagree.
func (c *Client) SetFingerprint(fp Fingerprint) {
	fp = fp.Normalize()
	c.fpMu.Lock()
	c.fp = fp
	c.fpMu.Unlock()

	c.propsMu.Lock()
	c.propsCache = make(map[string]string)
	c.propsMu.Unlock()
}

// Fingerprint returns the device identity currently in use.
func (c *Client) Fingerprint() Fingerprint {
	c.fpMu.RLock()
	defer c.fpMu.RUnlock()
	return c.fp
}

// SetRegionProxy registers a regional proxy for a sweep target code. Every
// request made with api.WithRegionInfo for that code is tunneled through it.
func (c *Client) SetRegionProxy(code, proxyURL string) error {
	if c.transport == nil {
		return fmt.Errorf("client transport is unavailable")
	}
	return c.transport.SetProxy(code, proxyURL)
}

// RemoveRegionProxy unregisters a regional proxy.
func (c *Client) RemoveRegionProxy(code string) {
	if c.transport != nil {
		c.transport.RemoveProxy(code)
	}
}

// HasRegionProxy reports whether a proxy is registered for a region code.
func (c *Client) HasRegionProxy(code string) bool {
	if c.transport == nil {
		return false
	}
	return c.transport.HasProxy(code)
}

// CloseIdleConnections releases pooled connections held by every route.
func (c *Client) CloseIdleConnections() {
	if c.transport != nil {
		c.transport.CloseIdleConnections()
	}
}

// SetLocale changes the default locale/timezone used when a request has no
// explicit region or locale override.
func (c *Client) SetLocale(locale, timezone string) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if locale != "" {
		c.locale = locale
	}
	if timezone != "" {
		c.timezone = timezone
	}
}

// SetAPIBaseForTesting overrides apiBase for unit tests using httptest.Server.
func (c *Client) SetAPIBaseForTesting(url string) {
	c.apiBase = url
}

// requestSettings resolves the effective locale/timezone for one request.
func (c *Client) requestSettings(ctx context.Context) (string, string) {
	c.stateMu.RLock()
	locale := c.locale
	timezone := c.timezone
	c.stateMu.RUnlock()
	if override, ok := localeOverrideFromContext(ctx); ok {
		if override.Locale != "" {
			locale = override.Locale
		}
		if override.Timezone != "" {
			timezone = override.Timezone
		}
	}
	if info, ok := RegionInfoFromContext(ctx); ok {
		if info.Locale != "" {
			locale = info.Locale
		}
		if info.Timezone != "" {
			timezone = info.Timezone
		}
	}
	if locale == "" {
		locale = "en-US"
	}
	if timezone == "" {
		timezone = "America/New_York"
	}
	return locale, timezone
}

// superPropsFor caches the base64 client fingerprint per locale.
func (c *Client) superPropsFor(locale string) string {
	c.propsMu.Lock()
	defer c.propsMu.Unlock()
	if c.propsCache == nil {
		c.propsCache = make(map[string]string)
	}
	if cached, ok := c.propsCache[locale]; ok {
		return cached
	}
	c.fpMu.RLock()
	fp := c.fp
	c.fpMu.RUnlock()
	props := fp.SuperProperties(c.buildNumber, locale)
	c.propsCache[locale] = props
	return props
}

func (c *Client) setHeaders(req *http.Request) {
	locale, timezone := c.requestSettings(req.Context())
	fp := c.Fingerprint()
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", fp.UserAgent())
	req.Header.Set("X-Super-Properties", c.superPropsFor(locale))
	req.Header.Set("X-Discord-Locale", locale)
	req.Header.Set("X-Discord-Timezone", timezone)
	req.Header.Set("Accept-Language", acceptLanguageFor(locale))
	req.Header.Set("Origin", "https://discord.com")
	req.Header.Set("Referer", "https://discord.com/channels/@me")
}

// acceptLanguageFor mirrors Chromium's language negotiation: the exact locale
// first, then the base language, then English. A bare "ja" with no fallback is
// not something the desktop client ever sends.
func acceptLanguageFor(locale string) string {
	if locale == "" {
		return "en-US,en;q=0.9"
	}
	base := locale
	if idx := strings.IndexAny(locale, "-_"); idx > 0 {
		base = locale[:idx]
	} else {
		// Chromium expands a language-only locale to a regional default.
		if locale != "en" {
			return locale + ",en-US;q=0.9,en;q=0.8"
		}
		return "en-US,en;q=0.9"
	}
	if base == "en" {
		return locale + ",en;q=0.9"
	}
	return locale + "," + base + ";q=0.9,en;q=0.8"
}

// parseRetryAfter parses the retry duration from a 429 response.
func parseRetryAfter(resp *http.Response, bodyBytes []byte) time.Duration {
	var rl RateLimitResponse
	if err := json.Unmarshal(bodyBytes, &rl); err == nil && rl.RetryAfter > 0 {
		return time.Duration(rl.RetryAfter*1000) * time.Millisecond
	}

	if headerVal := resp.Header.Get("Retry-After"); headerVal != "" {
		if sec, err := strconv.ParseFloat(headerVal, 64); err == nil && sec > 0 {
			return time.Duration(sec*1000) * time.Millisecond
		}
	}

	return 5 * time.Second
}

// terminalQuestStatuses are Discord statuses that can never succeed on retry
// because the quest campaign itself is done (ended, claimed, or removed).
var terminalQuestStatuses = map[int]bool{
	http.StatusNotFound:            true, // 404 - quest no longer exists
	http.StatusGone:                true, // 410 - campaign ended
	http.StatusForbidden:           true, // 403 - no longer eligible / unenrolled
	http.StatusUnprocessableEntity: true, // 422 - quest closed for submissions
}

// classifyQuestFailure converts a non-retryable quest response into a typed
// error so the runner can drop the quest quietly instead of hammering it.
func classifyQuestFailure(status int, bodyBytes []byte, questID string) error {
	if terminalQuestStatuses[status] {
		return &QuestExpiredError{QuestID: questID, Status: status, Detail: truncateBody(bodyBytes)}
	}
	return nil
}

// truncateBody keeps error messages readable in the console.
func truncateBody(bodyBytes []byte) string {
	const max = 180
	s := strings.TrimSpace(string(bodyBytes))
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// ValidateToken verifies authentication and retrieves current account identity.
func (c *Client) ValidateToken(ctx context.Context) (*UserMe, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiBase+"/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("create user request failed: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request /users/@me failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading /users/@me response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryDuration := parseRetryAfter(resp, bodyBytes)
		return nil, fmt.Errorf("rate limited (HTTP 429), retry after %v", retryDuration)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed: HTTP %d (response: %s)", resp.StatusCode, string(bodyBytes))
	}

	var user UserMe
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		return nil, fmt.Errorf("failed to parse /users/@me json: %w", err)
	}

	return &user, nil
}

// FetchQuests queries GET /api/v9/quests/@me with retry/backoff handling.
func (c *Client) FetchQuests(ctx context.Context) ([]Quest, error) {
	return c.fetchQuestsRetrying(ctx, 3)
}

// fetchQuestsRetrying fetches the quest payload, retrying transient failures.
func (c *Client) fetchQuestsRetrying(ctx context.Context, maxAttempts int) ([]Quest, error) {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			if !sleepContext(ctx, time.Duration(attempt-1)*800*time.Millisecond) {
				return nil, ctx.Err()
			}
		}
		payload, status, body, retryAfter, err := c.doQuestFetch(ctx)
		if err != nil {
			lastErr = err
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			continue
		}
		if status == http.StatusOK && payload != nil {
			return payload.Quests, nil
		}
		if status == http.StatusTooManyRequests {
			if retryAfter <= 0 {
				retryAfter = 3 * time.Second
			}
			lastErr = fmt.Errorf("rate limited (HTTP 429), retry after %v", retryAfter)
			if !sleepContext(ctx, retryAfter+time.Second) {
				return nil, ctx.Err()
			}
			continue
		}
		return nil, fmt.Errorf("fetching quests failed: HTTP %d: %s", status, truncateBody(body))
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("fetching quests failed")
	}
	return nil, lastErr
}

// FetchQuestsWithLocale queries GET /api/v9/quests/@me with explicit locale
// headers without mutating the shared client state.
func (c *Client) FetchQuestsWithLocale(ctx context.Context, locale, timezone string) ([]Quest, error) {
	return c.fetchQuestsRetrying(withLocaleOverride(ctx, locale, timezone), 3)
}

// FetchQuestsMultiRegion is the legacy multi-locale entry point kept for
// compatibility. New code should call SweepQuests with BuildTargets instead.
func (c *Client) FetchQuestsMultiRegion(ctx context.Context, regions []string) ([]Quest, error) {
	targets := make([]RegionTarget, 0, len(regions)+1)
	if len(regions) == 0 {
		targets = append(targets, DirectTarget())
	}
	for _, raw := range regions {
		token := strings.TrimSpace(raw)
		if token == "" || strings.EqualFold(token, "all") {
			targets = append(targets, DirectTarget())
			continue
		}
		if preset, ok := regionPresets[strings.ToUpper(token)]; ok {
			targets = append(targets, RegionTarget{
				Code:     "L:" + strings.ToUpper(token),
				Locale:   preset.Locale,
				Timezone: preset.Timezone,
				Kind:     TargetLocale,
			})
			continue
		}
		if locale, ok := NormalizeDiscordLocale(token); ok {
			targets = append(targets, RegionTarget{
				Code:     "L:" + locale,
				Locale:   locale,
				Timezone: timezoneForLocale(locale),
				Kind:     TargetLocale,
			})
		}
	}
	if len(targets) == 0 {
		targets = append(targets, DirectTarget())
	}
	opts := DefaultSweepOptions()
	opts.MinInterval = 500 * time.Millisecond
	opts.MaxInterval = 5 * time.Second
	result, err := c.SweepQuests(ctx, targets, opts)
	if err != nil && (result == nil || len(result.Quests) == 0) {
		return c.FetchQuests(ctx)
	}
	if result == nil {
		return c.FetchQuests(ctx)
	}
	return result.Quests, nil
}

// EnrollQuestWithCaptcha sends enrollment payload to POST /api/v9/quests/{qid}/enroll, optionally with captcha solution.
func (c *Client) EnrollQuestWithCaptcha(ctx context.Context, quest Quest, captchaKey, captchaRqtoken string) error {
	payload := EnrollPayload{
		Location:              11,
		IsTargeted:            false,
		TrafficMetadataRaw:    quest.TrafficMetadataRaw,
		TrafficMetadataSealed: quest.TrafficMetadataSealed,
	}
	if captchaKey != "" {
		payload.CaptchaKey = &captchaKey
	}
	if captchaRqtoken != "" {
		payload.CaptchaRqtoken = &captchaRqtoken
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode enroll payload: %w", err)
	}

	url := fmt.Sprintf("%s/quests/%s/enroll", c.apiBase, quest.ID)
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create enroll request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("enroll request failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed reading enroll response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryDuration := parseRetryAfter(resp, bodyBytes)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDuration + 1*time.Second):
				continue
			}
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			return nil
		}

		// Detect Discord Captcha challenge on HTTP 400
		if resp.StatusCode == http.StatusBadRequest {
			var captchaResp DiscordCaptchaResponse
			if err := json.Unmarshal(bodyBytes, &captchaResp); err == nil && (captchaResp.CaptchaSitekey != "" || len(captchaResp.CaptchaKey) > 0) {
				key := ""
				if len(captchaResp.CaptchaKey) > 0 {
					key = captchaResp.CaptchaKey[0]
				}
				return &CaptchaRequiredError{
					QuestID:        quest.ID,
					CaptchaKey:     key,
					CaptchaService: captchaResp.CaptchaService,
					CaptchaSitekey: captchaResp.CaptchaSitekey,
					CaptchaRqdata:  captchaResp.CaptchaRqdata,
					CaptchaRqtoken: captchaResp.CaptchaRqtoken,
				}
			}
		}

		if err := classifyQuestFailure(resp.StatusCode, bodyBytes, quest.ID); err != nil {
			return err
		}

		return fmt.Errorf("enroll failed: HTTP %d: %s", resp.StatusCode, truncateBody(bodyBytes))
	}

	return fmt.Errorf("exceeded max retries enrolling quest %s", quest.ID)
}

// EnrollQuest sends enrollment payload to POST /api/v9/quests/{qid}/enroll.
func (c *Client) EnrollQuest(ctx context.Context, quest Quest) error {
	return c.EnrollQuestWithCaptcha(ctx, quest, "", "")
}

// SendVideoProgress updates video watch timestamp on POST /api/v9/quests/{qid}/video-progress.
func (c *Client) SendVideoProgress(ctx context.Context, questID string, timestamp float64) (*ProgressUpdateResponse, error) {
	payload := VideoProgressPayload{Timestamp: timestamp}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode video progress payload: %w", err)
	}

	url := fmt.Sprintf("%s/quests/%s/video-progress", c.apiBase, questID)
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create video progress request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("video progress request failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed reading video progress response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryDuration := parseRetryAfter(resp, bodyBytes)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDuration + 1*time.Second):
				continue
			}
		}

		if resp.StatusCode == http.StatusOK {
			var update ProgressUpdateResponse
			_ = json.Unmarshal(bodyBytes, &update)
			return &update, nil
		}

		if err := classifyQuestFailure(resp.StatusCode, bodyBytes, questID); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("video progress failed: HTTP %d: %s", resp.StatusCode, truncateBody(bodyBytes))
	}

	return nil, fmt.Errorf("exceeded max retries for video progress on quest %s", questID)
}

// SendHeartbeat sends a game activity heartbeat to POST /api/v9/quests/{qid}/heartbeat.
func (c *Client) SendHeartbeat(ctx context.Context, questID string, streamKey string, terminal bool) (*ProgressUpdateResponse, error) {
	payload := HeartbeatPayload{
		StreamKey: streamKey,
		Terminal:  terminal,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode heartbeat payload: %w", err)
	}

	url := fmt.Sprintf("%s/quests/%s/heartbeat", c.apiBase, questID)
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create heartbeat request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("heartbeat request failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed reading heartbeat response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryDuration := parseRetryAfter(resp, bodyBytes)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDuration + 1*time.Second):
				continue
			}
		}

		if resp.StatusCode == http.StatusOK {
			var update ProgressUpdateResponse
			_ = json.Unmarshal(bodyBytes, &update)
			return &update, nil
		}

		if err := classifyQuestFailure(resp.StatusCode, bodyBytes, questID); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("heartbeat failed: HTTP %d: %s", resp.StatusCode, truncateBody(bodyBytes))
	}

	return nil, fmt.Errorf("exceeded max retries for heartbeat on quest %s", questID)
}

// HypeSquadPayload models the JSON payload sent to POST /api/v9/hypesquad/online.
type HypeSquadPayload struct {
	HouseID int `json:"house_id"`
}

// JoinHypeSquad joins or switches the authenticated user's HypeSquad house (1: Bravery, 2: Brilliance, 3: Balance).
func (c *Client) JoinHypeSquad(ctx context.Context, houseID int) error {
	if houseID < 1 || houseID > 3 {
		return fmt.Errorf("invalid house ID %d: must be 1 (Bravery), 2 (Brilliance), or 3 (Balance)", houseID)
	}

	payload := HypeSquadPayload{HouseID: houseID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode hypesquad payload: %w", err)
	}

	url := c.apiBase + "/hypesquad/online"
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create hypesquad request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("hypesquad request failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed reading hypesquad response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryDuration := parseRetryAfter(resp, bodyBytes)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDuration + 1*time.Second):
				continue
			}
		}

		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			return nil
		}

		return fmt.Errorf("hypesquad join failed: HTTP %d: %s", resp.StatusCode, truncateBody(bodyBytes))
	}

	return fmt.Errorf("exceeded max retries joining hypesquad house %d", houseID)
}

// LeaveHypeSquad removes the authenticated user from their HypeSquad house.
func (c *Client) LeaveHypeSquad(ctx context.Context) error {
	url := c.apiBase + "/hypesquad/online"
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create hypesquad leave request: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("hypesquad leave request failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed reading hypesquad leave response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryDuration := parseRetryAfter(resp, bodyBytes)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDuration + 1*time.Second):
				continue
			}
		}

		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
			return nil
		}

		return fmt.Errorf("hypesquad leave failed: HTTP %d: %s", resp.StatusCode, truncateBody(bodyBytes))
	}

	return fmt.Errorf("exceeded max retries leaving hypesquad")
}
