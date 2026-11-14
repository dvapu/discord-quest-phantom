// Package identity keeps the device identity the tool claims to be: the
// Discord client build, OS build, locale and timezone. Discord correlates all
// of these across headers and X-Super-Properties, so they are derived from one
// source of truth instead of being hardcoded in three places.
package identity

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/i18n"
)

// KnownClient couples a Discord stable client release with the Electron and
// Chromium versions it ships and the native module build it reports.
//
// Only add an entry when all four values come from the same release. Mixing a
// current client_version with an older Electron version is exactly the
// inconsistency this package exists to prevent, and a stale client_version with
// a correct user agent is a far weaker signal than a self-contradictory pair.
type KnownClient struct {
	ClientVersion string
	Electron      string
	Chrome        string
	NativeBuild   int
}

// knownClients is ordered oldest to newest. Extend it when a Discord release is
// verified locally (app-<version> folder plus its Electron/Chromium pair).
var knownClients = []KnownClient{
	{ClientVersion: "1.0.9175", Electron: "32.2.7", Chrome: "128.0.6613.186", NativeBuild: 59498},
}

// Device is the resolved device identity plus the provenance of every value.
type Device struct {
	Fingerprint     api.Fingerprint
	SystemLocale    string
	Timezone        string
	InstalledClient string // version found on disk, empty when none was found
	ClientSource    string // installed-catalog | catalog | fallback
	Notes           []string
}

// KnownClientVersions returns the catalog entries, newest last.
func KnownClientVersions() []KnownClient {
	out := make([]KnownClient, len(knownClients))
	copy(out, knownClients)
	return out
}

// Detect resolves the local device identity.
func Detect() Device {
	var notes []string

	installed := installedDiscordVersion()
	known, ok := lookupKnown(installed)
	source := "installed-catalog"
	switch {
	case installed == "":
		source = "catalog"
		notes = append(notes, "no Discord install found; using the newest known client build")
	case !ok:
		source = "catalog"
		notes = append(notes, fmt.Sprintf(
			"installed Discord %s is not in the version catalog; pinned to %s so client_version, user agent and native build stay consistent",
			installed, known.ClientVersion))
	}
	if known.ClientVersion == "" {
		d := api.DefaultFingerprint()
		known = KnownClient{
			ClientVersion: d.ClientVersion,
			Electron:      d.BrowserVersion,
			Chrome:        api.ChromeVersionForUserAgent(d.BrowserUserAgent),
			NativeBuild:   d.NativeBuildNumber,
		}
		source = "fallback"
	}

	locale := systemLocale()
	osName := osName()
	arch := archFor(runtime.GOOS, runtime.GOARCH)
	fp := api.Fingerprint{
		OS:                osName,
		OSVersion:         platformOSBuild(),
		OSArch:            arch,
		AppArch:           arch,
		ClientVersion:     known.ClientVersion,
		BrowserVersion:    known.Electron,
		BrowserUserAgent:  api.BuildDiscordUserAgent(osName, known.ClientVersion, known.Electron, known.Chrome),
		ReleaseChannel:    "stable",
		NativeBuildNumber: known.NativeBuild,
		SystemLocale:      locale,
	}

	return Device{
		Fingerprint:     fp.Normalize(),
		SystemLocale:    locale,
		Timezone:        systemTimezone(locale),
		InstalledClient: installed,
		ClientSource:    source,
		Notes:           notes,
	}
}

// Summary is a one-line human readable description for the startup banner.
func (d Device) Summary() string {
	fp := d.Fingerprint
	return fmt.Sprintf("%s %s %s | client %s (%s) | locale %s | tz %s",
		fp.OS, fp.OSVersion, fp.OSArch, fp.ClientVersion, fp.ReleaseChannel, d.SystemLocale, d.Timezone)
}

func lookupKnown(version string) (KnownClient, bool) {
	if version == "" {
		if len(knownClients) == 0 {
			return KnownClient{}, false
		}
		return knownClients[len(knownClients)-1], false
	}
	for _, entry := range knownClients {
		if entry.ClientVersion == version {
			return entry, true
		}
	}
	if len(knownClients) == 0 {
		return KnownClient{}, false
	}
	return knownClients[len(knownClients)-1], false
}

func osName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "Darwin"
	case "linux":
		return "Linux"
	default:
		return "Windows"
	}
}

// archFor returns the architecture Discord's desktop client would report.
// Discord ships its Linux desktop client for x64 only, and the Linux
// user agent we send says x86_64, so a Linux host must claim x64 even when it
// really runs on arm64. Sending "arm64" next to an x86_64 user agent is a
// self-contradiction inside one request.
func osArch() string { return archFor(runtime.GOOS, runtime.GOARCH) }

func archFor(goos, goarch string) string {
	if goos == "linux" {
		return "x64"
	}
	switch goarch {
	case "amd64":
		return "x64"
	case "arm64":
		return "arm64"
	case "386":
		return "x86"
	default:
		return "x64"
	}
}

// systemLocale returns the OS UI locale in Discord's BCP-47 shape.
func systemLocale() string {
	raw := strings.TrimSpace(i18n.OSLocale())
	if raw == "" {
		return "en-US"
	}
	raw = strings.ReplaceAll(raw, "_", "-")
	parts := strings.Split(raw, "-")
	if len(parts) == 1 {
		// A bare language is valid but Discord clients normally carry a region.
		switch strings.ToLower(parts[0]) {
		case "vi":
			return "vi-VN"
		case "ja":
			return "ja-JP"
		case "ko":
			return "ko-KR"
		case "zh":
			return "zh-CN"
		case "en":
			return "en-US"
		case "pt":
			return "pt-BR"
		case "fr":
			return "fr-FR"
		case "de":
			return "de-DE"
		case "es":
			return "es-ES"
		default:
			return raw
		}
	}
	// Normalize the region subtag to upper case (vi-vn -> vi-VN).
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) == 2 {
			parts[i] = strings.ToUpper(parts[i])
		}
	}
	return strings.Join(parts, "-")
}

// systemTimezone prefers the real OS timezone and only falls back to a
// locale-derived default when the OS cannot name one (some containers report
// "Local" or an empty value).
func systemTimezone(locale string) string {
	name := strings.TrimSpace(time.Local.String())
	if isIANAZone(name) {
		return name
	}
	if tz := defaultTimezoneForLocale(locale); tz != "" {
		return tz
	}
	return "UTC"
}

func isIANAZone(name string) bool {
	if name == "" || name == "Local" || name == "UTC" {
		return name == "UTC"
	}
	// IANA zones are Area/City (Asia/Ho_Chi_Minh). Windows timezone ids such as
	// "SE Asia Standard Time" contain spaces and are not accepted by Discord.
	if strings.ContainsAny(name, " \\") {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}

func defaultTimezoneForLocale(locale string) string {
	switch strings.ToLower(locale) {
	case "vi-vn", "vi":
		return "Asia/Ho_Chi_Minh"
	case "ja-jp", "ja":
		return "Asia/Tokyo"
	case "ko-kr", "ko":
		return "Asia/Seoul"
	case "zh-cn", "zh":
		return "Asia/Shanghai"
	case "zh-tw":
		return "Asia/Taipei"
	case "en-us", "en":
		return "America/New_York"
	case "en-gb":
		return "Europe/London"
	case "de-de", "de":
		return "Europe/Berlin"
	case "fr-fr", "fr":
		return "Europe/Paris"
	case "pt-br", "pt":
		return "America/Sao_Paulo"
	case "ru-ru", "ru":
		return "Europe/Moscow"
	case "tr-tr", "tr":
		return "Europe/Istanbul"
	case "hi-in", "hi":
		return "Asia/Kolkata"
	}
	return ""
}
