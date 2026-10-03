package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ReleaseInfo holds metadata for a GitHub release.
type ReleaseInfo struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Name        string `json:"name"`
	Body        string `json:"body"`
}

// CheckUpdate queries GitHub releases for repository dvapu/discord-quest-phantom
// and reports whether a newer version than currentVersion is available.
func CheckUpdate(ctx context.Context, currentVersion string) (*ReleaseInfo, bool) {
	return checkUpdateFromURL(ctx, "https://api.github.com/repos/dvapu/discord-quest-phantom/releases/latest", currentVersion)
}

func checkUpdateFromURL(ctx context.Context, apiURL, currentVersion string) (*ReleaseInfo, bool) {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, false
	}
	req.Header.Set("User-Agent", "Discord-Quest-Phantom/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, false
	}
	defer resp.Body.Close()

	var rel ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, false
	}

	cleanLatest := strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
	cleanCurrent := strings.TrimPrefix(strings.TrimSpace(currentVersion), "v")

	if cleanLatest != "" && cleanLatest != cleanCurrent && cleanLatest > cleanCurrent {
		return &rel, true
	}

	return &rel, false
}
