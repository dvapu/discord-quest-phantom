package identity

import (
	"strings"
	"testing"

	"discord-quest-completer/pkg/api"
)

func TestDetectProducesConsistentFingerprint(t *testing.T) {
	device := Detect()
	t.Logf("device: %s", device.Summary())
	for _, note := range device.Notes {
		t.Logf("note: %s", note)
	}

	if issues := device.Fingerprint.Validate(); len(issues) != 0 {
		t.Fatalf("detected fingerprint is self-contradictory: %v", issues)
	}
	if device.Timezone == "" {
		t.Fatal("timezone must never be empty")
	}
	if device.SystemLocale == "" || !strings.Contains(device.SystemLocale, "-") {
		t.Fatalf("system locale %q is not a BCP-47 tag", device.SystemLocale)
	}
	if device.Fingerprint.BrowserUserAgent != device.Fingerprint.UserAgent() {
		t.Fatal("fingerprint user agent is not stable")
	}
	if device.ClientSource == "" {
		t.Fatal("client source must explain where the version came from")
	}
}

func TestDetectedFingerprintDoesNotMixUserAgentVersions(t *testing.T) {
	fp := Detect().Fingerprint
	if !strings.Contains(fp.BrowserUserAgent, "discord/"+fp.ClientVersion) {
		t.Fatalf("user agent %q does not carry client version %s", fp.BrowserUserAgent, fp.ClientVersion)
	}
	if !strings.Contains(fp.BrowserUserAgent, "Electron/"+fp.BrowserVersion) {
		t.Fatalf("user agent %q does not carry Electron version %s", fp.BrowserUserAgent, fp.BrowserVersion)
	}
	if !strings.Contains(fp.BrowserUserAgent, "Chrome/"+api.ChromeVersionForUserAgent(fp.BrowserUserAgent)) {
		t.Fatalf("user agent %q has no usable Chrome version", fp.BrowserUserAgent)
	}
}

func TestCatalogEntriesAreInternallyConsistent(t *testing.T) {
	if len(KnownClientVersions()) == 0 {
		t.Fatal("version catalog is empty")
	}
	for _, entry := range KnownClientVersions() {
		fp := api.Fingerprint{
			OS:                "Windows",
			OSVersion:         "10.0.26100",
			OSArch:            "x64",
			AppArch:           "x64",
			ClientVersion:     entry.ClientVersion,
			BrowserVersion:    entry.Electron,
			BrowserUserAgent:  api.BuildDiscordUserAgent("Windows", entry.ClientVersion, entry.Electron, entry.Chrome),
			ReleaseChannel:    "stable",
			NativeBuildNumber: entry.NativeBuild,
		}
		if issues := fp.Validate(); len(issues) != 0 {
			t.Fatalf("catalog entry %s is inconsistent: %v", entry.ClientVersion, issues)
		}
	}
}

func TestNormalizeFillsMissingValuesFromOneSource(t *testing.T) {
	fp := api.Fingerprint{ClientVersion: "1.0.9175", BrowserVersion: "32.2.7"}.Normalize()
	if fp.OS == "" || fp.OSVersion == "" || fp.OSArch == "" {
		t.Fatalf("normalize left required fields empty: %+v", fp)
	}
	if !strings.Contains(fp.BrowserUserAgent, "discord/1.0.9175") {
		t.Fatalf("normalize built the wrong user agent: %q", fp.BrowserUserAgent)
	}
	if issues := fp.Validate(); len(issues) != 0 {
		t.Fatalf("normalized fingerprint is inconsistent: %v", issues)
	}
}

func TestValidateCatchesMismatchedUserAgent(t *testing.T) {
	fp := api.DefaultFingerprint()
	fp.ClientVersion = "1.0.9268" // user agent still says 1.0.9175
	issues := fp.Validate()
	if len(issues) == 0 {
		t.Fatal("expected a version mismatch to be reported")
	}
	joined := strings.Join(issues, " | ")
	if !strings.Contains(joined, "discord/1.0.9268") {
		t.Fatalf("mismatch report is not specific: %s", joined)
	}
}

func TestDefaultTimezoneForLocale(t *testing.T) {
	cases := map[string]string{
		"vi-VN": "Asia/Ho_Chi_Minh",
		"ja-JP": "Asia/Tokyo",
		"en-US": "America/New_York",
		"de-DE": "Europe/Berlin",
	}
	for locale, want := range cases {
		if got := defaultTimezoneForLocale(locale); got != want {
			t.Fatalf("timezone for %s = %q, want %q", locale, got, want)
		}
	}
}

func TestIsIANAZone(t *testing.T) {
	if !isIANAZone("Asia/Ho_Chi_Minh") {
		t.Fatal("expected Asia/Ho_Chi_Minh to be accepted")
	}
	for _, bad := range []string{"", "Local", "SE Asia Standard Time", "not a zone"} {
		if isIANAZone(bad) {
			t.Fatalf("expected %q to be rejected", bad)
		}
	}
}
