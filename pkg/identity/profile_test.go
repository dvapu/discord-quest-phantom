package identity

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"discord-quest-completer/pkg/api"
)

func testDevice() Device {
	return Device{
		Fingerprint: api.Fingerprint{
			OS:                "Windows",
			OSVersion:         "10.0.26100",
			OSArch:            "x64",
			AppArch:           "x64",
			ClientVersion:     "1.0.9175",
			BrowserVersion:    "32.2.7",
			BrowserUserAgent:  api.BuildDiscordUserAgent("Windows", "1.0.9175", "32.2.7", "128.0.6613.186"),
			ReleaseChannel:    "stable",
			NativeBuildNumber: 59498,
		},
		SystemLocale: "vi-VN",
		Timezone:     "Asia/Ho_Chi_Minh",
		ClientSource: "catalog",
	}
}

func TestProfileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	profile, created, err := store.Enforce("123456789012345678", testDevice())
	if err != nil {
		t.Fatalf("enforce failed: %v", err)
	}
	if !created {
		t.Fatal("expected a new profile on first use")
	}
	if profile.Fingerprint.ClientVersion != "1.0.9175" {
		t.Fatalf("stored the wrong client version: %s", profile.Fingerprint.ClientVersion)
	}

	info, err := os.Stat(store.Path("123456789012345678"))
	if err != nil {
		t.Fatalf("profile file is missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("profile file is empty")
	}

	raw, err := os.ReadFile(store.Path("123456789012345678"))
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if len(raw) < len(profileMagic) || string(raw[:len(profileMagic)]) != profileMagic {
		t.Fatal("profile file must start with the format magic")
	}
	if string(raw) == string(mustJSON(t, profile)) {
		t.Fatal("profile must not be stored as plaintext")
	}

	loaded, err := store.Load("123456789012345678")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded == nil || loaded.AccountID != profile.AccountID {
		t.Fatalf("loaded profile does not match: %+v", loaded)
	}
	if loaded.Fingerprint != profile.Fingerprint {
		t.Fatalf("fingerprint did not survive the round trip: %+v", loaded.Fingerprint)
	}
}

func TestProfileStoreIsolatesAccounts(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, _, err := store.Enforce("111", testDevice()); err != nil {
		t.Fatalf("enforce failed: %v", err)
	}
	if _, _, err := store.Enforce("222", testDevice()); err != nil {
		t.Fatalf("enforce failed: %v", err)
	}
	if store.Path("111") == store.Path("222") {
		t.Fatal("two accounts must not share a profile file")
	}
	if missing, err := store.Load("333"); err != nil || missing != nil {
		t.Fatalf("unknown account must be nil, nil; got %v, %v", missing, err)
	}
}

func TestProfileStoreKeepsCreationTime(t *testing.T) {
	store := NewStore(t.TempDir())
	first, _, err := store.Enforce("42", testDevice())
	if err != nil {
		t.Fatalf("enforce failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	updated := testDevice()
	updated.Fingerprint.ClientVersion = "1.0.9500"
	updated.Fingerprint.BrowserUserAgent = api.BuildDiscordUserAgent("Windows", "1.0.9500", "35.0.0", "135.0.0.0")
	updated.Fingerprint.BrowserVersion = "35.0.0"

	second, created, err := store.Enforce("42", updated)
	if err != nil {
		t.Fatalf("second enforce failed: %v", err)
	}
	if created {
		t.Fatal("second enforce must reuse the existing profile")
	}
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("creation time changed: %s -> %s", first.CreatedAt, second.CreatedAt)
	}
	if second.Fingerprint.ClientVersion != "1.0.9500" {
		t.Fatalf("new client build was not adopted: %s", second.Fingerprint.ClientVersion)
	}
	if issues := second.Fingerprint.Validate(); len(issues) != 0 {
		t.Fatalf("adopted fingerprint is inconsistent: %v", issues)
	}
}

func TestProfileStoreRejectsTamperedAccount(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if _, _, err := store.Enforce("owner", testDevice()); err != nil {
		t.Fatalf("enforce failed: %v", err)
	}

	// A profile copied over another account's path must not be accepted.
	target := store.Path("intruder")
	if err := os.Rename(store.Path("owner"), target); err != nil {
		t.Fatalf("rename failed: %v", err)
	}
	loaded, err := store.Load("intruder")
	if err == nil || loaded != nil {
		t.Fatalf("expected the account mismatch to be rejected, got %+v, %v", loaded, err)
	}
}

func TestAuditFlagsBrokenFingerprint(t *testing.T) {
	profile := &Profile{
		AccountID:    "1",
		Fingerprint:  api.DefaultFingerprint(),
		DeviceTag:    deviceTag(),
		ClientSource: "fallback",
	}
	profile.Fingerprint.ClientVersion = "9.9.9" // user agent disagrees

	issues := Audit(profile, testDevice(), "local", "")
	if len(issues) == 0 {
		t.Fatal("audit must report the contradiction")
	}
	foundError := false
	for _, issue := range issues {
		if issue.Level == "error" {
			foundError = true
		}
	}
	if !foundError {
		t.Fatalf("audit did not raise an error: %+v", issues)
	}
}

func TestProfileApplySetsFingerprintAndLocale(t *testing.T) {
	client := api.NewClient("token", 500000)
	profile := &Profile{
		Fingerprint:  testDevice().Fingerprint,
		SystemLocale: "vi-VN",
		Timezone:     "Asia/Ho_Chi_Minh",
	}
	profile.Apply(client)

	got := client.Fingerprint()
	if got.ClientVersion != "1.0.9175" {
		t.Fatalf("client did not adopt the profile fingerprint: %+v", got)
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}
