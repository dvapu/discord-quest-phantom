package api

import (
	"strings"
	"testing"
)

func TestBuildTargetsAllWithProxy(t *testing.T) {
	targets, warnings, err := BuildTargets("all", "us=socks5://127.0.0.1:1080", "", "off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d (%v)", len(targets), targets)
	}
	if targets[0].Code != "LOCAL" || targets[0].Kind != TargetDirect {
		t.Fatalf("expected LOCAL direct target first, got %+v", targets[0])
	}
	if targets[1].Code != "US" || targets[1].Kind != TargetProxy || targets[1].ProxyURL == "" {
		t.Fatalf("expected US proxy target second, got %+v", targets[1])
	}
	if targets[1].Locale != "en-US" {
		t.Fatalf("expected en-US locale for US proxy, got %s", targets[1].Locale)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings when a proxy is configured, got %v", warnings)
	}
}

func TestBuildTargetsWarnsWithoutProxy(t *testing.T) {
	targets, warnings, err := BuildTargets("all", "", "", "off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 || targets[0].Code != "LOCAL" {
		t.Fatalf("expected only the LOCAL target, got %+v", targets)
	}
	if len(warnings) == 0 {
		t.Fatal("expected a warning explaining that no regional proxy is configured")
	}
	found := false
	for _, warning := range warnings {
		if strings.Contains(warning, "no regional proxy") {
			found = true
		}
	}
	if !found {
		t.Fatalf("warning does not mention the missing proxy: %v", warnings)
	}
}

func TestBuildTargetsLocaleSweepTrio(t *testing.T) {
	targets, _, err := BuildTargets("all", "", "", "trio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	codes := make([]string, 0, len(targets))
	for _, target := range targets {
		codes = append(codes, target.Code)
	}
	joined := strings.Join(codes, ",")
	if joined != "LOCAL,L:JA,L:VI" {
		t.Fatalf("unexpected trio target list: %s", joined)
	}
}

func TestBuildTargetsExplicitRegionWithoutProxy(t *testing.T) {
	targets, warnings, err := BuildTargets("br", "", "", "off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected LOCAL + L:BR, got %+v", targets)
	}
	if targets[1].Locale != "pt-BR" || targets[1].Kind != TargetLocale {
		t.Fatalf("expected a pt-BR locale target, got %+v", targets[1])
	}
	if len(warnings) == 0 || !strings.Contains(warnings[0], "no proxy configured") {
		t.Fatalf("expected a locale-only warning, got %v", warnings)
	}
}

func TestBuildTargetsUnknownFails(t *testing.T) {
	if _, _, err := BuildTargets("zzz-not-a-region", "", "", "off"); err == nil {
		t.Fatal("expected an error for an unknown region")
	}
}

func TestBuildTargetsRejectsBadLocaleSweep(t *testing.T) {
	if _, _, err := BuildTargets("all", "", "", "sometimes"); err == nil {
		t.Fatal("expected an error for an invalid locale sweep mode")
	}
}

func TestBuildTargetsFullIncludesLocales(t *testing.T) {
	targets, _, err := BuildTargets("full", "", "", "off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) < 20 {
		t.Fatalf("expected full sweep to include all locales, got %d targets", len(targets))
	}
}

func TestParseProxyTargetsVariants(t *testing.T) {
	targets, _, err := ParseProxyTargets("us=socks5://10.0.0.1:1080, http://10.0.0.2:8080", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 proxies, got %d", len(targets))
	}
	if targets[0].Code != "US" || targets[1].Code != "PROXY2" {
		t.Fatalf("unexpected proxy codes: %+v", targets)
	}
	if targets[1].ProxyURL != "http://10.0.0.2:8080" {
		t.Fatalf("unexpected proxy URL: %s", targets[1].ProxyURL)
	}
	if _, _, err := ParseProxyTargets("us=http://a:1,us=http://b:2", ""); err == nil {
		t.Fatal("expected duplicate proxy codes to be rejected")
	}
	if _, _, err := ParseProxyTargets("us=ftp://a:1", ""); err == nil {
		t.Fatal("expected an unsupported proxy scheme to be rejected")
	}
}

func TestRedactProxyURL(t *testing.T) {
	got := RedactProxyURL("socks5://user:secret@1.2.3.4:1080")
	if strings.Contains(got, "user") || strings.Contains(got, "secret") {
		t.Fatalf("proxy credentials leaked into %q", got)
	}
	if !strings.Contains(got, "1.2.3.4:1080") {
		t.Fatalf("proxy host missing from %q", got)
	}
}

func TestNormalizeDiscordLocale(t *testing.T) {
	cases := map[string]string{
		"ja-JP": "ja",
		"vi-VN": "vi",
		"en-us": "en-US",
		"pt-BR": "pt-BR",
		"EN":    "en-US",
		"zh-cn": "zh-CN",
	}
	for input, want := range cases {
		got, ok := NormalizeDiscordLocale(input)
		if !ok || got != want {
			t.Errorf("NormalizeDiscordLocale(%q) = %q, %v; want %q", input, got, ok, want)
		}
	}
	if _, ok := NormalizeDiscordLocale("xx-YY"); ok {
		t.Fatal("expected an unknown locale to be rejected")
	}
}

func TestBuildTargetsDedupesEquivalentRoutes(t *testing.T) {
	targets, _, err := BuildTargets("full", "", "", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen := make(map[string]bool, len(targets))
	for _, target := range targets {
		route := target.Locale + "|" + target.Timezone + "|" + target.ProxyURL
		if seen[route] {
			t.Fatalf("duplicate route in targets: %+v", target)
		}
		seen[route] = true
	}
}

func TestBuildTargetsExplicitMultiRegionWithoutProxies(t *testing.T) {
	targets, warnings, err := BuildTargets("us,br", "", "", "off")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	codes := make([]string, 0, len(targets))
	for _, target := range targets {
		codes = append(codes, target.Code)
	}
	joined := strings.Join(codes, ",")
	if joined != "LOCAL,L:BR" {
		// The US preset maps to en-US/America/New_York, which the local egress
		// already covers, so it is deduplicated instead of queried twice.
		t.Fatalf("unexpected target list: %s", joined)
	}
	if len(warnings) != 2 {
		t.Fatalf("expected warnings for the covered US region and the BR locale-only region, got %v", warnings)
	}
	if !strings.Contains(strings.Join(warnings, " "), "already covered by the local egress") {
		t.Fatalf("expected an explicit note that the US view is already covered, got %v", warnings)
	}
}

func TestParseProxyTargetsKeepsUrlsWithEquals(t *testing.T) {
	targets, _, err := ParseProxyTargets("http://proxy.example:8080/?token=abc", "")
	if err != nil {
		t.Fatalf("proxy URL with a query token must be accepted: %v", err)
	}
	if len(targets) != 1 || targets[0].ProxyURL != "http://proxy.example:8080/?token=abc" {
		t.Fatalf("unexpected parsed proxy: %+v", targets)
	}
	// Userinfo containing '=' must also survive.
	targets, _, err = ParseProxyTargets("socks5://user:p=ss@proxy.example:1080", "")
	if err != nil {
		t.Fatalf("proxy URL with '=' in userinfo must be accepted: %v", err)
	}
	if !strings.Contains(targets[0].ProxyURL, "p=ss") {
		t.Fatalf("userinfo was truncated: %s", targets[0].ProxyURL)
	}
	// A real CODE=url entry still splits correctly.
	targets, _, err = ParseProxyTargets("br=http://proxy.example:8080/?token=abc", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targets[0].Code != "BR" || !strings.Contains(targets[0].ProxyURL, "token=abc") {
		t.Fatalf("CODE=url entry parsed incorrectly: %+v", targets[0])
	}
}

func TestParseProxyTargetsWarnsOnUnknownCode(t *testing.T) {
	_, warnings, err := ParseProxyTargets("jap=http://proxy.example:8080", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) == 0 || !strings.Contains(warnings[0], "not a known country code") {
		t.Fatalf("expected an unknown-code warning, got %v", warnings)
	}
}
