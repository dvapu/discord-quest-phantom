package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Client handles authenticated requests to Discord REST API endpoints.
type Client struct {
	httpClient  *http.Client
	apiBase     string
	token       string
	userAgent   string
	superProps  string
	timezone    string
	buildNumber int
	locale      string
}

// BuildSuperPropertiesWithLocale produces a base64-encoded client fingerprint JSON string for a specific locale.
func BuildSuperPropertiesWithLocale(buildNumber int, locale string) string {
	if locale == "" {
		locale = "en-US"
	}
	props := map[string]interface{}{
		"os":                  "Windows",
		"browser":             "Discord Client",
		"release_channel":     "stable",
		"client_version":      "1.0.9175",
		"os_version":          "10.0.26100",
		"os_arch":             "x64",
		"app_arch":            "x64",
		"system_locale":       locale,
		"client_locale":       locale,
		"browser_user_agent":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9175 Chrome/128.0.6613.186 Electron/32.2.7 Safari/537.36",
		"browser_version":     "32.2.7",
		"client_build_number": buildNumber,
		"native_build_number": 59498,
		"client_event_source": nil,
	}

	raw, _ := json.Marshal(props)
	return base64.StdEncoding.EncodeToString(raw)
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
	locale := "en-US"
	timezone := "America/New_York"
	return &Client{
		httpClient:  &http.Client{Timeout: 20 * time.Second},
		apiBase:     "https://discord.com/api/v9",
		token:       token,
		userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9175 Chrome/128.0.6613.186 Electron/32.2.7 Safari/537.36",
		superProps:  BuildSuperPropertiesWithLocale(buildNumber, locale),
		timezone:    timezone,
		buildNumber: buildNumber,
		locale:      locale,
	}
}

// SetLocale dynamically reconfigures client locale, timezone, and super properties fingerprint.
func (c *Client) SetLocale(locale, timezone string) {
	if locale == "" {
		locale = "en-US"
	}
	if timezone == "" {
		timezone = "America/New_York"
	}
	c.locale = locale
	c.timezone = timezone
	c.superProps = BuildSuperPropertiesWithLocale(c.buildNumber, locale)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Super-Properties", c.superProps)
	req.Header.Set("X-Discord-Locale", c.locale)
	req.Header.Set("X-Discord-Timezone", c.timezone)
	req.Header.Set("Accept-Language", c.locale+",en;q=0.9")
	req.Header.Set("Origin", "https://discord.com")
	req.Header.Set("Referer", "https://discord.com/channels/@me")
}

// handleRateLimit checks for HTTP 429 and parses retry duration.
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

// FetchQuests queries GET /api/v9/quests/@me with automatic rate-limit backoff.
func (c *Client) FetchQuests(ctx context.Context) ([]Quest, error) {
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", c.apiBase+"/quests/@me", nil)
		if err != nil {
			return nil, fmt.Errorf("create quests request failed: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request /quests/@me failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed reading /quests/@me response: %w", err)
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

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetching quests failed: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var qResp QuestResponse
		if err := json.Unmarshal(bodyBytes, &qResp); err == nil && len(qResp.Quests) > 0 {
			return qResp.Quests, nil
		}

		var rawQuests []Quest
		if err := json.Unmarshal(bodyBytes, &rawQuests); err == nil && len(rawQuests) > 0 {
			return rawQuests, nil
		}

		if err == nil {
			return qResp.Quests, nil
		}

		return nil, fmt.Errorf("unable to decode quest response structure")
	}

	return nil, fmt.Errorf("exceeded max retries fetching quests")
}

// FetchQuestsWithLocale queries GET /api/v9/quests/@me with specified locale and timezone.
func (c *Client) FetchQuestsWithLocale(ctx context.Context, locale, timezone string) ([]Quest, error) {
	prevLocale := c.locale
	prevTz := c.timezone
	prevProps := c.superProps
	c.SetLocale(locale, timezone)
	defer func() {
		c.locale = prevLocale
		c.timezone = prevTz
		c.superProps = prevProps
	}()

	url := fmt.Sprintf("%s/quests/@me?locale=%s", c.apiBase, locale)
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create quests request failed: %w", err)
		}
		c.setHeaders(req)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request /quests/@me failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed reading /quests/@me response: %w", err)
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

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetching quests failed: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var qResp QuestResponse
		if err := json.Unmarshal(bodyBytes, &qResp); err == nil && len(qResp.Quests) > 0 {
			return qResp.Quests, nil
		}

		var rawQuests []Quest
		if err := json.Unmarshal(bodyBytes, &rawQuests); err == nil && len(rawQuests) > 0 {
			return rawQuests, nil
		}

		if err == nil {
			return qResp.Quests, nil
		}

		return nil, fmt.Errorf("unable to decode quest response structure")
	}

	return nil, fmt.Errorf("exceeded max retries fetching quests")
}

// FetchQuestsMultiRegion discovers quests across multiple regions (e.g. US, JP, VN) and deduplicates them.
func (c *Client) FetchQuestsMultiRegion(ctx context.Context, regions []string) ([]Quest, error) {
	type targetLocale struct {
		locale   string
		timezone string
	}

	var targets []targetLocale
	wantUS := false
	wantJP := false
	wantVN := false

	if len(regions) == 0 {
		wantUS = true
		wantJP = true
		wantVN = true
	} else {
		for _, r := range regions {
			switch strings.ToLower(strings.TrimSpace(r)) {
			case "all", "":
				wantUS = true
				wantJP = true
				wantVN = true
			case "us", "en", "en-us":
				wantUS = true
			case "jp", "ja", "ja-jp":
				wantJP = true
			case "vn", "vi", "vi-vn":
				wantVN = true
			}
		}
	}

	if wantUS {
		targets = append(targets, targetLocale{locale: "en-US", timezone: "America/New_York"})
	}
	if wantJP {
		targets = append(targets, targetLocale{locale: "ja-JP", timezone: "Asia/Tokyo"})
	}
	if wantVN {
		targets = append(targets, targetLocale{locale: "vi-VN", timezone: "Asia/Ho_Chi_Minh"})
	}

	seen := make(map[string]bool)
	var aggregated []Quest

	for i, t := range targets {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		quests, err := c.FetchQuestsWithLocale(ctx, t.locale, t.timezone)
		if err == nil {
			for _, q := range quests {
				if !seen[q.ID] {
					seen[q.ID] = true
					aggregated = append(aggregated, q)
				}
			}
		}

		if i < len(targets)-1 {
			time.Sleep(500 * time.Millisecond) // Polite pacing between regional sweeps
		}
	}

	if len(aggregated) > 0 {
		return aggregated, nil
	}

	return c.FetchQuests(ctx)
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

		return fmt.Errorf("enroll failed: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
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

		return nil, fmt.Errorf("video progress failed: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
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

		return nil, fmt.Errorf("heartbeat failed: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil, fmt.Errorf("exceeded max retries for heartbeat on quest %s", questID)
}
