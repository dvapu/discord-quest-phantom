package api

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// TargetKind describes how a region target reaches Discord.
type TargetKind int

const (
	// TargetDirect uses the host machine's own network egress.
	TargetDirect TargetKind = iota
	// TargetProxy tunnels quest requests through a regional proxy.
	TargetProxy
	// TargetLocale only changes locale headers. It does not change eligibility:
	// Discord decides quest availability from the request IP and account state.
	TargetLocale
)

// RegionInfo is carried in a request context so the client can align locale,
// timezone and the proxy transport for every call that belongs to a target.
type RegionInfo struct {
	Code     string
	Locale   string
	Timezone string
}

// RegionTarget is one egress/locale unit used by the multi-region sweep.
type RegionTarget struct {
	Code     string
	Locale   string
	Timezone string
	ProxyURL string
	Kind     TargetKind
}

// Info returns the context payload used to route a target's requests.
func (t RegionTarget) Info() RegionInfo {
	return RegionInfo{Code: t.Code, Locale: t.Locale, Timezone: t.Timezone}
}

// discordLocaleInfo lists the locales the Discord client ships with.
var discordLocaleInfo = []struct {
	Locale   string
	Timezone string
}{
	{"da", "Europe/Copenhagen"},
	{"de", "Europe/Berlin"},
	{"en-GB", "Europe/London"},
	{"en-US", "America/New_York"},
	{"es-419", "America/Mexico_City"},
	{"es-ES", "Europe/Madrid"},
	{"fr", "Europe/Paris"},
	{"hr", "Europe/Zagreb"},
	{"hu", "Europe/Budapest"},
	{"it", "Europe/Rome"},
	{"lt", "Europe/Vilnius"},
	{"nl", "Europe/Amsterdam"},
	{"no", "Europe/Oslo"},
	{"pl", "Europe/Warsaw"},
	{"pt-BR", "America/Sao_Paulo"},
	{"ro", "Europe/Bucharest"},
	{"fi", "Europe/Helsinki"},
	{"sv-SE", "Europe/Stockholm"},
	{"vi", "Asia/Ho_Chi_Minh"},
	{"tr", "Europe/Istanbul"},
	{"cs", "Europe/Prague"},
	{"el", "Europe/Athens"},
	{"bg", "Europe/Sofia"},
	{"ru", "Europe/Moscow"},
	{"uk", "Europe/Kyiv"},
	{"hi", "Asia/Kolkata"},
	{"th", "Asia/Bangkok"},
	{"zh-CN", "Asia/Shanghai"},
	{"ja", "Asia/Tokyo"},
	{"zh-TW", "Asia/Taipei"},
	{"ko", "Asia/Seoul"},
}

// regionPresets maps ISO country codes to a Discord locale + IANA timezone.
// These values are used to align headers when a regional proxy is configured.
var regionPresets = map[string]struct {
	Locale   string
	Timezone string
}{
	"US": {"en-US", "America/New_York"},
	"CA": {"en-US", "America/Toronto"},
	"MX": {"es-419", "America/Mexico_City"},
	"BR": {"pt-BR", "America/Sao_Paulo"},
	"AR": {"es-419", "America/Argentina/Buenos_Aires"},
	"CL": {"es-419", "America/Santiago"},
	"CO": {"es-419", "America/Bogota"},
	"PE": {"es-419", "America/Lima"},
	"GB": {"en-GB", "Europe/London"},
	"IE": {"en-GB", "Europe/Dublin"},
	"DE": {"de", "Europe/Berlin"},
	"FR": {"fr", "Europe/Paris"},
	"ES": {"es-ES", "Europe/Madrid"},
	"IT": {"it", "Europe/Rome"},
	"NL": {"nl", "Europe/Amsterdam"},
	"BE": {"nl", "Europe/Brussels"},
	"CH": {"de", "Europe/Zurich"},
	"AT": {"de", "Europe/Vienna"},
	"SE": {"sv-SE", "Europe/Stockholm"},
	"NO": {"no", "Europe/Oslo"},
	"DK": {"da", "Europe/Copenhagen"},
	"FI": {"fi", "Europe/Helsinki"},
	"PL": {"pl", "Europe/Warsaw"},
	"CZ": {"cs", "Europe/Prague"},
	"HU": {"hu", "Europe/Budapest"},
	"RO": {"ro", "Europe/Bucharest"},
	"BG": {"bg", "Europe/Sofia"},
	"GR": {"el", "Europe/Athens"},
	"HR": {"hr", "Europe/Zagreb"},
	"LT": {"lt", "Europe/Vilnius"},
	"TR": {"tr", "Europe/Istanbul"},
	"RU": {"ru", "Europe/Moscow"},
	"UA": {"uk", "Europe/Kyiv"},
	"JP": {"ja", "Asia/Tokyo"},
	"KR": {"ko", "Asia/Seoul"},
	"TW": {"zh-TW", "Asia/Taipei"},
	"HK": {"zh-TW", "Asia/Hong_Kong"},
	"CN": {"zh-CN", "Asia/Shanghai"},
	"SG": {"en-GB", "Asia/Singapore"},
	"MY": {"en-GB", "Asia/Kuala_Lumpur"},
	"TH": {"th", "Asia/Bangkok"},
	"VN": {"vi", "Asia/Ho_Chi_Minh"},
	"ID": {"en-US", "Asia/Jakarta"},
	"PH": {"en-US", "Asia/Manila"},
	"IN": {"hi", "Asia/Kolkata"},
	"AU": {"en-GB", "Australia/Sydney"},
	"NZ": {"en-GB", "Pacific/Auckland"},
	"ZA": {"en-GB", "Africa/Johannesburg"},
	"AE": {"en-US", "Asia/Dubai"},
	"IL": {"en-US", "Asia/Jerusalem"},
}

// normalizeLanguageAliases maps bare language tags to a preferred full locale.
var normalizeLanguageAliases = map[string]string{
	"en": "en-US",
	"es": "es-ES",
	"pt": "pt-BR",
	"zh": "zh-CN",
	"sv": "sv-SE",
}

// proxyCodePattern limits what may appear before the '=' of a "CODE=url" entry.
var proxyCodePattern = regexp.MustCompile("^[A-Za-z0-9_-]+$")

// NormalizeDiscordLocale converts common spellings (ja-JP, EN-us, ja_JP) to a
// canonical Discord locale identifier. Unknown locales report false.
func NormalizeDiscordLocale(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	lower := strings.ToLower(trimmed)
	for _, info := range discordLocaleInfo {
		if strings.ToLower(info.Locale) == lower {
			return info.Locale, true
		}
	}
	language := lower
	if idx := strings.IndexAny(language, "-_"); idx > 0 {
		language = language[:idx]
	}
	if canonical, ok := normalizeLanguageAliases[language]; ok {
		return canonical, true
	}
	for _, info := range discordLocaleInfo {
		if strings.ToLower(info.Locale) == language {
			return info.Locale, true
		}
	}
	return "", false
}

// timezoneForLocale returns the default IANA timezone for a Discord locale.
func timezoneForLocale(locale string) string {
	for _, info := range discordLocaleInfo {
		if strings.EqualFold(info.Locale, locale) {
			return info.Timezone
		}
	}
	return "America/New_York"
}

// DirectTarget returns the local egress target. The local view is always swept
// first because it is the fastest and most authentic one for the signed-in user.
func DirectTarget() RegionTarget {
	return RegionTarget{Code: "LOCAL", Locale: "en-US", Timezone: "America/New_York", Kind: TargetDirect}
}

// DiscordLocaleTargets returns one locale-only target per Discord client locale.
func DiscordLocaleTargets() []RegionTarget {
	targets := make([]RegionTarget, 0, len(discordLocaleInfo))
	for _, info := range discordLocaleInfo {
		targets = append(targets, RegionTarget{
			Code:     "L:" + info.Locale,
			Locale:   info.Locale,
			Timezone: info.Timezone,
			Kind:     TargetLocale,
		})
	}
	return targets
}

// TrioLocaleTargets returns the extra locales used by the legacy golden trio
// sweep (en-US is already covered by the direct target).
func TrioLocaleTargets() []RegionTarget {
	return []RegionTarget{
		{Code: "L:ja", Locale: "ja", Timezone: "Asia/Tokyo", Kind: TargetLocale},
		{Code: "L:vi", Locale: "vi", Timezone: "Asia/Ho_Chi_Minh", Kind: TargetLocale},
	}
}

// ParseProxyTargets parses inline and file based proxy definitions. Each entry
// is either "code=socks5://host:port" or a bare proxy URL.
func ParseProxyTargets(inline, file string) ([]RegionTarget, []string, error) {
	var warnings []string
	var entries []string
	if strings.TrimSpace(inline) != "" {
		entries = append(entries, strings.Split(inline, ",")...)
	}
	if strings.TrimSpace(file) != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, warnings, fmt.Errorf("read proxy file: %w", err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if idx := strings.Index(line, " #"); idx > 0 {
				line = strings.TrimSpace(line[:idx])
			}
			entries = append(entries, line)
		}
	}
	targets := make([]RegionTarget, 0, len(entries))
	seen := make(map[string]bool, len(entries))
	for index, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		target, err := parseProxyEntry(entry, index+1)
		if err != nil {
			return nil, warnings, err
		}
		key := strings.ToUpper(target.Code)
		if seen[key] {
			return nil, warnings, fmt.Errorf("duplicate proxy code %q", target.Code)
		}
		seen[key] = true
		targets = append(targets, target)
		if _, known := regionPresets[strings.ToUpper(target.Code)]; !known {
			warnings = append(warnings, fmt.Sprintf("proxy code %q is not a known country code; defaulting its locale/timezone to en-US/America/New_York", target.Code))
		}
	}
	return targets, warnings, nil
}

func parseProxyEntry(entry string, index int) (RegionTarget, error) {
	code := ""
	rawURL := entry
	// Only split on '=' when the left side looks like a code and the right side
	// is a URL. URLs may legitimately contain '=' (tokens, userinfo, queries).
	if idx := strings.Index(entry, "="); idx > 0 {
		candidateCode := strings.TrimSpace(entry[:idx])
		candidateURL := strings.TrimSpace(entry[idx+1:])
		if proxyCodePattern.MatchString(candidateCode) && strings.Contains(candidateURL, "://") {
			code = candidateCode
			rawURL = candidateURL
		}
	}
	if code == "" {
		code = fmt.Sprintf("PROXY%d", index)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return RegionTarget{}, fmt.Errorf("invalid proxy %q: %w", rawURL, err)
	}
	scheme := strings.ToLower(parsed.Scheme)
	switch scheme {
	case "http", "https", "socks5", "socks5h":
	default:
		return RegionTarget{}, fmt.Errorf("unsupported proxy scheme %q (use http, https, socks5 or socks5h)", parsed.Scheme)
	}
	if parsed.Host == "" {
		return RegionTarget{}, fmt.Errorf("proxy %q has no host", rawURL)
	}
	preset, ok := regionPresets[strings.ToUpper(code)]
	if !ok {
		preset = regionPresets["US"]
	}
	return RegionTarget{
		Code:     strings.ToUpper(code),
		Locale:   preset.Locale,
		Timezone: preset.Timezone,
		ProxyURL: rawURL,
		Kind:     TargetProxy,
	}, nil
}

// RedactProxyURL hides credentials so proxy addresses can be logged safely.
func RedactProxyURL(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "(invalid proxy url)"
	}
	host := parsed.Hostname()
	if port := parsed.Port(); port != "" {
		host = host + ":" + port
	}
	if host == "" {
		return "(invalid proxy url)"
	}
	return strings.ToLower(parsed.Scheme) + "://" + host
}

// BuildTargets converts CLI region/proxy/locale options into sweep targets.
// The local egress is always included first so the sweep can fall back to the
// signed-in user's own region even when every configured proxy is unusable.
func BuildTargets(regionSpec, proxySpec, proxyFile, localeSweep string) ([]RegionTarget, []string, error) {
	proxies, warnings, err := ParseProxyTargets(proxySpec, proxyFile)
	if err != nil {
		return nil, warnings, err
	}
	proxyByCode := make(map[string]RegionTarget, len(proxies))
	for _, proxy := range proxies {
		proxyByCode[strings.ToUpper(proxy.Code)] = proxy
	}

	seen := make(map[string]bool, len(proxies)+2)
	seenRoute := make(map[string]bool, len(proxies)+2)
	targets := make([]RegionTarget, 0, len(proxies)+2)
	add := func(target RegionTarget) error {
		code := strings.ToUpper(strings.TrimSpace(target.Code))
		if code == "" {
			return fmt.Errorf("sweep target has an empty code")
		}
		target.Code = code
		if target.Locale == "" {
			target.Locale = "en-US"
		}
		if target.Timezone == "" {
			target.Timezone = "America/New_York"
		}
		// Two targets with the same locale/timezone/egress would return the same
		// payload and only burn rate-limit budget.
		route := strings.ToLower(target.Locale) + "|" + target.Timezone + "|" + target.ProxyURL
		if seen[code] || seenRoute[route] {
			return nil
		}
		seen[code] = true
		seenRoute[route] = true
		targets = append(targets, target)
		return nil
	}

	if err := add(DirectTarget()); err != nil {
		return nil, warnings, err
	}

	spec := strings.ToLower(strings.TrimSpace(regionSpec))
	full := spec == "full" || spec == "max"
	useAllProxies := spec == "" || spec == "all" || spec == "auto" || full

	if !useAllProxies && spec != "direct" && spec != "local" && spec != "none" {
		for _, raw := range strings.Split(regionSpec, ",") {
			token := strings.TrimSpace(raw)
			if token == "" {
				continue
			}
			lower := strings.ToLower(token)
			if lower == "all" || lower == "auto" {
				useAllProxies = true
				continue
			}
			if lower == "direct" || lower == "local" || lower == "none" {
				continue
			}
			if proxy, ok := proxyByCode[strings.ToUpper(token)]; ok {
				before := len(targets)
				if err := add(proxy); err != nil {
					return nil, warnings, err
				}
				if len(targets) == before {
					warnings = append(warnings, fmt.Sprintf("target %q is already covered by another sweep route; duplicate request skipped", token))
				}
				continue
			}
			if preset, ok := regionPresets[strings.ToUpper(token)]; ok {
				before := len(targets)
				if err := add(RegionTarget{Code: "L:" + strings.ToUpper(token), Locale: preset.Locale, Timezone: preset.Timezone, Kind: TargetLocale}); err != nil {
					return nil, warnings, err
				}
				if len(targets) > before {
					warnings = append(warnings, fmt.Sprintf("region %s has no proxy configured; probing it only changes locale headers, which does not change quest eligibility", strings.ToUpper(token)))
				} else {
					warnings = append(warnings, fmt.Sprintf("region %s is already covered by the local egress (same locale/timezone); duplicate request skipped", strings.ToUpper(token)))
				}
				continue
			}
			if locale, ok := NormalizeDiscordLocale(token); ok {
				before := len(targets)
				if err := add(RegionTarget{Code: "L:" + locale, Locale: locale, Timezone: timezoneForLocale(locale), Kind: TargetLocale}); err != nil {
					return nil, warnings, err
				}
				if len(targets) == before {
					warnings = append(warnings, fmt.Sprintf("locale %q is already covered by another sweep route; duplicate request skipped", locale))
				}
				continue
			}
			return nil, warnings, fmt.Errorf("unknown region or locale %q: use a proxy code, a country code (US, BR, JP...), a Discord locale (ja, pt-BR...) or all/direct", token)
		}
	}

	if useAllProxies {
		for _, proxy := range proxies {
			if err := add(proxy); err != nil {
				return nil, warnings, err
			}
		}
	}

	switch strings.ToLower(strings.TrimSpace(localeSweep)) {
	case "", "off", "none", "false":
		// Locale headers do not change quest eligibility, so they stay opt-in.
	case "trio":
		for _, target := range TrioLocaleTargets() {
			if err := add(target); err != nil {
				return nil, warnings, err
			}
		}
	case "all", "full", "true":
		for _, target := range DiscordLocaleTargets() {
			if err := add(target); err != nil {
				return nil, warnings, err
			}
		}
	default:
		return nil, warnings, fmt.Errorf("invalid locale sweep mode %q (use off, trio or all)", localeSweep)
	}

	if full {
		for _, target := range DiscordLocaleTargets() {
			if err := add(target); err != nil {
				return nil, warnings, err
			}
		}
	}

	if len(targets) == 0 {
		return nil, warnings, fmt.Errorf("no quest discovery targets configured")
	}
	if len(proxies) == 0 && (spec == "all" || spec == "" || spec == "auto") {
		warnings = append(warnings, "no regional proxy configured: only the local egress is queried. Discord decides quest eligibility from the request IP, so region-locked quests stay hidden until a proxy is supplied with -proxy or -proxy-file")
	}
	return targets, warnings, nil
}
