package inventory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"discord-quest-completer/pkg/api"
)

const snapshotVersion = 1

// Snapshot is the last successful quest inventory saved to disk.
type Snapshot struct {
	Version                int         `json:"version"`
	UserID                 string      `json:"user_id"`
	SavedAt                time.Time   `json:"saved_at"`
	Quests                 []api.Quest `json:"quests"`
	ExcludedCount          int         `json:"excluded_count"`
	EnrollmentBlockedUntil *string     `json:"enrollment_blocked_until,omitempty"`
	AccessSuspendedUntil   *string     `json:"access_suspended_until,omitempty"`
}

// Path returns the snapshot path inside a cache directory.
func Path(cacheDir string) string {
	return filepath.Join(cacheDir, "quests-snapshot.json")
}

// Load reads a snapshot from disk. A missing file returns (nil, nil) so a first
// run does not look like an error.
func Load(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read quest snapshot: %w", err)
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode quest snapshot: %w", err)
	}
	if snapshot.Version != snapshotVersion {
		return nil, nil
	}
	return &snapshot, nil
}

// Save writes a snapshot atomically so a crash cannot leave a torn file.
func Save(path string, snapshot *Snapshot) error {
	if snapshot == nil {
		return nil
	}
	snapshot.Version = snapshotVersion
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	data, err := json.MarshalIndent(snapshot, "", " ")
	if err != nil {
		return fmt.Errorf("encode quest snapshot: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "quests-snapshot-*.tmp")
	if err != nil {
		return fmt.Errorf("create snapshot temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write snapshot temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close snapshot temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replace quest snapshot: %w", err)
	}
	return nil
}

// Merge returns the current quests plus unexpired snapshot quests the live fetch
// failed to return. It bridges transient API gaps only: stale snapshots are
// ignored entirely and claimed/expired quests are never restored.
func Merge(current []api.Quest, snapshot *Snapshot, userID string, now time.Time, maxAge time.Duration) []api.Quest {
	merged := make([]api.Quest, 0, len(current)+8)
	seen := make(map[string]bool, len(current)+8)
	for _, quest := range current {
		seen[quest.ID] = true
		merged = append(merged, quest)
	}
	if snapshot == nil || snapshot.UserID == "" || (userID != "" && snapshot.UserID != userID) {
		return merged
	}
	if maxAge > 0 && now.Sub(snapshot.SavedAt) > maxAge {
		return merged
	}
	for _, quest := range snapshot.Quests {
		if quest.ID == "" || seen[quest.ID] {
			continue
		}
		if !ActiveQuest(quest, now) {
			continue
		}
		if quest.DiscoveredRegion == "" {
			quest.DiscoveredRegion = "CACHE"
		}
		quest.FromCache = true
		seen[quest.ID] = true
		merged = append(merged, quest)
	}
	return merged
}

// ActiveQuest reports whether a quest has not expired and is not already claimed.
func ActiveQuest(quest api.Quest, now time.Time) bool {
	if quest.UserStatus != nil && quest.UserStatus.ClaimedAt != nil && *quest.UserStatus.ClaimedAt != "" {
		return false
	}
	if quest.Config.ExpiresAt == "" {
		return true
	}
	expires, err := time.Parse(time.RFC3339, quest.Config.ExpiresAt)
	if err != nil {
		return true
	}
	return expires.After(now)
}

// FromSweep builds a snapshot from a successful discovery sweep.
func FromSweep(userID string, result *api.SweepResult, now time.Time) *Snapshot {
	if result == nil {
		return nil
	}
	return &Snapshot{
		UserID:                 userID,
		SavedAt:                now,
		Quests:                 result.Quests,
		ExcludedCount:          result.ExcludedCount(),
		EnrollmentBlockedUntil: result.EnrollmentBlockedUntil,
		AccessSuspendedUntil:   result.AccessSuspendedUntil,
	}
}
