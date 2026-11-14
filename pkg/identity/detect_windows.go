//go:build windows

package identity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const keyWow64_64Key = 0x0100

// platformOSBuild reads the real Windows NT build from the registry instead of
// hardcoding one machine's build into every installation.
func platformOSBuild() string {
	build := registryString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "CurrentBuildNumber")
	build = strings.TrimSpace(build)
	if build == "" {
		return "10.0.26100"
	}
	return "10.0." + build
}

func registryString(path, name string) string {
	var root syscall.Handle
	sub, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return ""
	}
	if err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, sub, 0, syscall.KEY_READ|keyWow64_64Key, &root); err != nil {
		return ""
	}
	defer syscall.RegCloseKey(root)

	value, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return ""
	}
	buf := make([]uint16, 256)
	size := uint32(len(buf) * 2)
	var typ uint32
	if err := syscall.RegQueryValueEx(root, value, nil, &typ, (*byte)(unsafe.Pointer(&buf[0])), &size); err != nil {
		return ""
	}
	if typ != syscall.REG_SZ {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

// installedDiscordVersion returns the highest stable Discord client version
// installed for the current user, or "" when Discord is not installed.
func installedDiscordVersion() string {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" {
		return ""
	}
	entries, err := os.ReadDir(filepath.Join(root, "Discord"))
	if err != nil {
		return ""
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
	if best == "" {
		return ""
	}

	// build_info.json is authoritative when present: it also tells us whether
	// this is the stable channel we claim to be.
	info := filepath.Join(root, "Discord", "app-"+best, "resources", "build_info.json")
	if data, err := os.ReadFile(info); err == nil {
		var parsed struct {
			Version        string `json:"version"`
			ReleaseChannel string `json:"releaseChannel"`
		}
		if json.Unmarshal(data, &parsed) == nil {
			if parsed.ReleaseChannel != "" && parsed.ReleaseChannel != "stable" {
				return ""
			}
			if looksLikeClientVersion(parsed.Version) {
				return parsed.Version
			}
		}
	}
	return best
}

func looksLikeClientVersion(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
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
			av, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bv, _ = strconv.Atoi(bs[i])
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
