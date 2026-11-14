package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"discord-quest-completer/pkg/api"
)

const (
	profileVersion = 1
	profileMagic   = "DQC1"
)

// Profile is the stored identity of one Discord account. It is deliberately
// stable: the fingerprint only changes when the device itself changes (a new
// Discord client build), never per run. That stability is what makes an account
// look like "the same person on the same machine", which is the opposite of
// what a rotating, per-run fingerprint does.
type Profile struct {
	Version      int             `json:"version"`
	AccountID    string          `json:"account_id"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Fingerprint  api.Fingerprint `json:"fingerprint"`
	SystemLocale string          `json:"system_locale"`
	ClientLocale string          `json:"client_locale"`
	Timezone     string          `json:"timezone"`
	HomeEgress   string          `json:"home_egress"`
	ClientSource string          `json:"client_source"`
	DeviceTag    string          `json:"device_tag"`
	Notes        []string        `json:"notes,omitempty"`
}

// Store keeps one profile file per account under the cache directory. Files are
// sealed with the platform secret store (DPAPI on Windows) and named by a hash
// so a stray backup never reveals which Discord account they belong to.
type Store struct {
	dir string
}

// NewStore returns a profile store rooted at <cacheDir>/profiles.
func NewStore(cacheDir string) *Store {
	return &Store{dir: filepath.Join(cacheDir, "profiles")}
}

// Dir returns the directory that holds the profile files.
func (s *Store) Dir() string { return s.dir }

// Path returns the profile file for an account.
func (s *Store) Path(accountID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(accountID)))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:8])+".profile")
}

// Load reads a profile. A missing file is not an error: it returns (nil, nil).
func (s *Store) Load(accountID string) (*Profile, error) {
	raw, err := os.ReadFile(s.Path(accountID))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read profile: %w", err)
	}
	if !strings.HasPrefix(string(raw), profileMagic) {
		return nil, fmt.Errorf("profile file has an unknown format")
	}
	payload, err := Unprotect(raw[len(profileMagic):])
	if err != nil {
		return nil, fmt.Errorf("unseal profile (created by another Windows user?): %w", err)
	}
	var profile Profile
	if err := json.Unmarshal(payload, &profile); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}
	if profile.AccountID != accountID {
		return nil, fmt.Errorf("profile belongs to a different account")
	}
	return &profile, nil
}

// Save writes the profile atomically, sealed at rest.
func (s *Store) Save(profile *Profile) error {
	if profile == nil || profile.AccountID == "" {
		return fmt.Errorf("profile has no account id")
	}
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("create profile directory: %w", err)
	}
	profile.Version = profileVersion
	profile.UpdatedAt = time.Now().UTC()

	payload, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile: %w", err)
	}
	sealed, err := Protect(payload)
	if err != nil {
		return fmt.Errorf("seal profile: %w", err)
	}

	target := s.Path(profile.AccountID)
	temp := target + ".tmp"
	if err := os.WriteFile(temp, append([]byte(profileMagic), sealed...), 0o600); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}
	return os.Rename(temp, target)
}

// Enforce loads the account profile or creates it from the detected device.
// The returned bool reports whether a new profile was created.
func (s *Store) Enforce(accountID string, device Device) (*Profile, bool, error) {
	existing, err := s.Load(accountID)
	if err != nil {
		// A profile from another Windows user or a corrupt file must not block
		// the run: rebuild it and say so.
		existing = nil
	}

	now := time.Now().UTC()
	tag := deviceTag()
	if existing == nil {
		profile := &Profile{
			Version:      profileVersion,
			AccountID:    accountID,
			CreatedAt:    now,
			UpdatedAt:    now,
			Fingerprint:  device.Fingerprint,
			SystemLocale: device.SystemLocale,
			ClientLocale: device.SystemLocale,
			Timezone:     device.Timezone,
			HomeEgress:   "direct",
			ClientSource: device.ClientSource,
			DeviceTag:    tag,
			Notes:        append([]string(nil), device.Notes...),
		}
		if saveErr := s.Save(profile); saveErr != nil {
			return profile, true, saveErr
		}
		return profile, true, err
	}

	changed := false
	if existing.Fingerprint.ClientVersion != device.Fingerprint.ClientVersion ||
		existing.Fingerprint.BrowserUserAgent != device.Fingerprint.BrowserUserAgent ||
		existing.Fingerprint.SystemLocale != device.Fingerprint.SystemLocale {
		existing.Fingerprint = device.Fingerprint
		existing.ClientSource = device.ClientSource
		changed = true
	}
	if existing.SystemLocale != device.SystemLocale {
		existing.SystemLocale = device.SystemLocale
		changed = true
	}
	if existing.ClientLocale == "" {
		existing.ClientLocale = device.SystemLocale
		changed = true
	}
	if existing.Timezone != device.Timezone {
		existing.Timezone = device.Timezone
		changed = true
	}
	if existing.HomeEgress == "" {
		existing.HomeEgress = "direct"
		changed = true
	}
	if changed {
		if saveErr := s.Save(existing); saveErr != nil {
			return existing, false, saveErr
		}
	}
	return existing, false, nil
}

// DeviceTag is a local, non-transmitted marker of the machine the profile was
// first created on. It exists so moving a profile to another computer is
// visible in the audit instead of silently blending two devices.
func deviceTag() string {
	host, _ := os.Hostname()
	sum := sha256.Sum256([]byte(strings.ToLower(host) + "|" + osName() + "|" + osArch()))
	return hex.EncodeToString(sum[:6])
}

// Apply installs the profile fingerprint and locale defaults on an API client.
func (p *Profile) Apply(client *api.Client) {
	if p == nil || client == nil {
		return
	}
	locale := p.ClientLocale
	if locale == "" {
		locale = p.SystemLocale
	}
	client.SetFingerprint(p.Fingerprint)
	client.SetLocale(locale, p.Timezone)
}

// Issue is one finding of the profile audit.
type Issue struct {
	Level   string // "error", "warn" or "info"
	Message string
}

// Audit compares the stored profile with the device and the session egress.
func Audit(profile *Profile, device Device, egressRegion, egressTimezone string) []Issue {
	var issues []Issue
	if profile == nil {
		return []Issue{{Level: "warn", Message: "no account profile available"}}
	}

	for _, problem := range profile.Fingerprint.Validate() {
		issues = append(issues, Issue{Level: "error", Message: "fingerprint: " + problem})
	}

	if profile.DeviceTag != "" && profile.DeviceTag != deviceTag() {
		issues = append(issues, Issue{
			Level:   "warn",
			Message: "this account profile was created on a different machine",
		})
	}

	if device.InstalledClient != "" && device.InstalledClient != profile.Fingerprint.ClientVersion {
		issues = append(issues, Issue{
			Level: "info",
			Message: fmt.Sprintf("installed Discord is %s, pinned client is %s (catalog keeps every field consistent)",
				device.InstalledClient, profile.Fingerprint.ClientVersion),
		})
	}

	switch profile.ClientSource {
	case "catalog":
		issues = append(issues, Issue{
			Level:   "info",
			Message: "client version came from the catalog because the installed build is not indexed",
		})
	case "fallback":
		issues = append(issues, Issue{
			Level:   "warn",
			Message: "using the built-in fallback fingerprint",
		})
	}

	if !ProtectionStrong() {
		issues = append(issues, Issue{
			Level:   "warn",
			Message: "profile is not encrypted at rest on this platform (permissions only)",
		})
	}

	if egressRegion != "" && !strings.EqualFold(egressRegion, "direct") &&
		egressTimezone != "" && egressTimezone != profile.Timezone {
		issues = append(issues, Issue{
			Level: "info",
			Message: fmt.Sprintf("session egress %s uses timezone %s while the home identity stays %s",
				egressRegion, egressTimezone, profile.Timezone),
		})
	}

	return issues
}
