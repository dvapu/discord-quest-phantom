package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Fingerprint is the device identity every request claims to come from.
//
// Discord correlates a handful of fields across headers and the base64
// X-Super-Properties blob: the Discord client version, the Electron/Chromium
// version inside its user agent, the declared OS build and the locale/timezone.
// When those drift apart the request looks synthetic even if each individual
// value is plausible. Keeping them in one struct, derived from a single source
// of truth, is the whole point: every field is written from the same device.
type Fingerprint struct {
	OS                string `json:"os"`
	OSVersion         string `json:"os_version"`
	OSArch            string `json:"os_arch"`
	AppArch           string `json:"app_arch"`
	ClientVersion     string `json:"client_version"`
	BrowserVersion    string `json:"browser_version"`
	BrowserUserAgent  string `json:"browser_user_agent"`
	ReleaseChannel    string `json:"release_channel"`
	NativeBuildNumber int    `json:"native_build_number"`
	// SystemLocale is the operating system locale. It is a property of the
	// machine and never follows a proxy or a per-request locale override; a real
	// client cannot change its OS locale between two requests.
	SystemLocale string `json:"system_locale"`
}

// DefaultFingerprint returns the last fingerprint the project shipped with.
// It stays the zero-configuration fallback so NewClient keeps working without
// any identity wiring (tests, one-off tools, cmd/test_regions).
func DefaultFingerprint() Fingerprint {
	return Fingerprint{
		OS:                "Windows",
		OSVersion:         "10.0.26100",
		OSArch:            "x64",
		AppArch:           "x64",
		ClientVersion:     "1.0.9175",
		BrowserVersion:    "32.2.7",
		BrowserUserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.9175 Chrome/128.0.6613.186 Electron/32.2.7 Safari/537.36",
		ReleaseChannel:    "stable",
		NativeBuildNumber: 59498,
		SystemLocale:      "en-US",
	}
}

// Normalize fills empty fields from the default fingerprint. A partially
// populated fingerprint is worse than the default one, because it mixes a real
// value with a placeholder.
func (f Fingerprint) Normalize() Fingerprint {
	d := DefaultFingerprint()
	if strings.TrimSpace(f.OS) == "" {
		f.OS = d.OS
	}
	if strings.TrimSpace(f.OSVersion) == "" {
		f.OSVersion = d.OSVersion
	}
	if strings.TrimSpace(f.OSArch) == "" {
		f.OSArch = d.OSArch
	}
	if strings.TrimSpace(f.AppArch) == "" {
		f.AppArch = d.AppArch
	}
	if strings.TrimSpace(f.ClientVersion) == "" {
		f.ClientVersion = d.ClientVersion
	}
	if strings.TrimSpace(f.BrowserVersion) == "" {
		f.BrowserVersion = d.BrowserVersion
	}
	if strings.TrimSpace(f.ReleaseChannel) == "" {
		f.ReleaseChannel = d.ReleaseChannel
	}
	if f.NativeBuildNumber <= 0 {
		f.NativeBuildNumber = d.NativeBuildNumber
	}
	if strings.TrimSpace(f.SystemLocale) == "" {
		f.SystemLocale = d.SystemLocale
	}
	if strings.TrimSpace(f.BrowserUserAgent) == "" {
		chrome := ChromeVersionForUserAgent(d.BrowserUserAgent)
		f.BrowserUserAgent = BuildDiscordUserAgent(f.OS, f.ClientVersion, f.BrowserVersion, chrome)
	}
	return f
}

// BuildDiscordUserAgent renders the desktop user agent. Discord's Electron
// client reports the same chromium-era string on every platform apart from the
// platform token, so the only moving parts are client, Chrome and Electron.
func BuildDiscordUserAgent(osName, clientVersion, browserVersion, chromeVersion string) string {
	if clientVersion == "" {
		clientVersion = DefaultFingerprint().ClientVersion
	}
	if browserVersion == "" {
		browserVersion = DefaultFingerprint().BrowserVersion
	}
	if chromeVersion == "" {
		chromeVersion = "128.0.6613.186"
	}
	return fmt.Sprintf(
		"Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) discord/%s Chrome/%s Electron/%s Safari/537.36",
		platformToken(osName), clientVersion, chromeVersion, browserVersion,
	)
}

// ChromeVersionForUserAgent extracts the Chromium version already present in a
// user agent, so rebuilding the string never invents a different browser build.
func ChromeVersionForUserAgent(ua string) string {
	const marker = "Chrome/"
	idx := strings.Index(ua, marker)
	if idx < 0 {
		return ""
	}
	rest := ua[idx+len(marker):]
	end := strings.IndexAny(rest, " ;")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

// ElectronVersionForUserAgent extracts the Electron version from a user agent.
func ElectronVersionForUserAgent(ua string) string {
	const marker = "Electron/"
	idx := strings.Index(ua, marker)
	if idx < 0 {
		return ""
	}
	rest := ua[idx+len(marker):]
	end := strings.IndexAny(rest, " ;")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func platformToken(osName string) string {
	switch strings.ToLower(strings.TrimSpace(osName)) {
	case "windows":
		return "Windows NT 10.0; Win64; x64"
	case "darwin", "macos", "mac":
		return "Macintosh; Intel Mac OS X 10_15_7"
	case "linux":
		return "X11; Linux x86_64"
	default:
		return "Windows NT 10.0; Win64; x64"
	}
}

// UserAgent returns the desktop user agent matching the fingerprint.
func (f Fingerprint) UserAgent() string {
	f = f.Normalize()
	return f.BrowserUserAgent
}

// SuperProperties encodes the fingerprint for one locale.
func (f Fingerprint) SuperProperties(buildNumber int, locale string) string {
	f = f.Normalize()
	if locale == "" {
		locale = "en-US"
	}
	systemLocale := f.SystemLocale
	if systemLocale == "" {
		systemLocale = locale
	}
	props := map[string]interface{}{
		"os":                  f.OS,
		"browser":             "Discord Client",
		"release_channel":     f.ReleaseChannel,
		"client_version":      f.ClientVersion,
		"os_version":          f.OSVersion,
		"os_arch":             f.OSArch,
		"app_arch":            f.AppArch,
		"system_locale":       systemLocale,
		"client_locale":       locale,
		"browser_user_agent":  f.BrowserUserAgent,
		"browser_version":     f.BrowserVersion,
		"client_build_number": buildNumber,
		"native_build_number": f.NativeBuildNumber,
		"client_event_source": nil,
	}

	raw, _ := json.Marshal(props)
	return base64.StdEncoding.EncodeToString(raw)
}

// Validate reports the internal contradictions of a fingerprint. Callers surface
// these before the first request instead of discovering them in a ban wave.
func (f Fingerprint) Validate() []string {
	f = f.Normalize()
	var issues []string

	if !strings.Contains(f.BrowserUserAgent, "discord/"+f.ClientVersion) {
		issues = append(issues,
			fmt.Sprintf("user agent does not carry discord/%s", f.ClientVersion))
	}
	if electron := ElectronVersionForUserAgent(f.BrowserUserAgent); electron != f.BrowserVersion {
		issues = append(issues, fmt.Sprintf(
			"user agent Electron/%s does not match browser_version %s", electron, f.BrowserVersion))
	}
	if !strings.Contains(f.BrowserUserAgent, "Electron/") {
		issues = append(issues, "user agent is not an Electron desktop client string")
	}
	if f.ReleaseChannel != "stable" && f.ReleaseChannel != "ptb" && f.ReleaseChannel != "canary" {
		issues = append(issues, fmt.Sprintf("unknown release channel %q", f.ReleaseChannel))
	}
	if f.NativeBuildNumber <= 0 {
		issues = append(issues, "native_build_number is missing; real desktop clients always send it")
	}
	if strings.EqualFold(f.OS, "Windows") {
		if _, err := checkOSVersion(f.OSVersion); err != nil {
			issues = append(issues, err.Error())
		}
	}
	return issues
}

// checkOSVersion is a sanity check on the declared OS build: Discord ships
// Windows NT versions in the 10.0.x form, never a bare major version.
func checkOSVersion(osVersion string) (string, error) {
	if osVersion == "" {
		return "", fmt.Errorf("os_version is empty")
	}
	if strings.HasPrefix(osVersion, "10.0.") {
		return osVersion, nil
	}
	return "", fmt.Errorf("os_version %q is not a Windows NT 10.0.x build", osVersion)
}
