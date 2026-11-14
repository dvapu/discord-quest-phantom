//go:build !windows

package identity

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// platformOSBuild reports the kernel version in the shape Discord expects. The
// desktop client reports the Darwin kernel version on macOS and the kernel
// release on Linux, never a Windows NT string.
func platformOSBuild() string {
	switch runtime.GOOS {
	case "darwin":
		if v := macProductVersion(); v != "" {
			return v
		}
		return "14.6.1"
	default:
		if v := unixKernelRelease(); v != "" {
			return v
		}
		return "6.8.0"
	}
}

func macProductVersion() string {
	data, err := os.ReadFile("/System/Library/CoreServices/SystemVersion.plist")
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`(?s)<key>ProductVersion</key>\s*<string>([^<]+)</string>`)
	if match := re.FindSubmatch(data); len(match) > 1 {
		return strings.TrimSpace(string(match[1]))
	}
	return ""
}

// unixKernelRelease reads the running kernel release without shelling out.
func unixKernelRelease() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// installedDiscordVersion looks for the Linux/macOS Discord install layout. Only
// stable paths are considered, because the fingerprint claims release_channel
// "stable".
func installedDiscordVersion() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".config", "discord"),
		filepath.Join(home, ".local", "share", "discord"),
	}
	for _, root := range candidates {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		best := ""
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "app-") {
				continue
			}
			version := strings.TrimPrefix(entry.Name(), "app-")
			if !looksLikeClientVersion(version) {
				continue
			}
			if best == "" || compareVersions(version, best) > 0 {
				best = version
			}
		}
		if best != "" {
			return best
		}
	}
	return ""
}

func looksLikeClientVersion(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func compareVersions(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(as) {
			for _, r := range as[i] {
				av = av*10 + int(r-'0')
			}
		}
		if i < len(bs) {
			for _, r := range bs[i] {
				bv = bv*10 + int(r-'0')
			}
		}
		if av != bv {
			if av > bv {
				return 1
			}
			return -1
		}
	}
	return 0
}
