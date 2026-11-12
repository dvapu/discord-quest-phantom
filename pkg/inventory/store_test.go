package inventory

import (
	"path/filepath"
	"testing"
	"time"

	"discord-quest-completer/pkg/api"
)

func stringPtr(value string) *string { return &value }

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := Path(dir)
	if path != filepath.Join(dir, "quests-snapshot.json") {
		t.Fatalf("unexpected snapshot path %q", path)
	}
	now := time.Now().UTC()
	snapshot := &Snapshot{
		UserID:  "user-1",
		SavedAt: now,
		Quests: []api.Quest{{
			ID:                 "q1",
			Config:             api.QuestConfig{ExpiresAt: now.Add(time.Hour).Format(time.RFC3339)},
			DiscoveredRegion:   "BR",
			DiscoveredLocale:   "pt-BR",
			DiscoveredTimezone: "America/Sao_Paulo",
		}},
		ExcludedCount: 7,
	}
	if err := Save(path, snapshot); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded == nil || loaded.UserID != "user-1" || len(loaded.Quests) != 1 || loaded.ExcludedCount != 7 {
		t.Fatalf("unexpected loaded snapshot: %+v", loaded)
	}
	restored := loaded.Quests[0]
	if restored.DiscoveredRegion != "BR" || restored.DiscoveredLocale != "pt-BR" || restored.DiscoveredTimezone != "America/Sao_Paulo" {
		t.Fatalf("regional routing metadata was not persisted: %+v", restored)
	}
}

func TestLoadMissingReturnsNil(t *testing.T) {
	snapshot, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("missing snapshot should not be an error: %v", err)
	}
	if snapshot != nil {
		t.Fatalf("expected nil snapshot, got %+v", snapshot)
	}
}

func TestMergeRestoresOnlyActiveUnclaimedQuests(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(time.Hour).Format(time.RFC3339)
	past := now.Add(-time.Hour).Format(time.RFC3339)
	claimed := now.Add(-time.Minute).Format(time.RFC3339)
	snapshot := &Snapshot{
		UserID:  "user-1",
		SavedAt: now,
		Quests: []api.Quest{
			{ID: "live"},
			{ID: "cachedActive", Config: api.QuestConfig{ExpiresAt: future}},
			{ID: "cachedExpired", Config: api.QuestConfig{ExpiresAt: past}},
			{ID: "cachedClaimed", Config: api.QuestConfig{ExpiresAt: future}, UserStatus: &api.UserStatus{ClaimedAt: stringPtr(claimed)}},
		},
	}
	merged := Merge([]api.Quest{{ID: "live"}}, snapshot, "user-1", now, time.Hour)
	if len(merged) != 2 {
		t.Fatalf("expected live + cachedActive, got %+v", merged)
	}
	byID := make(map[string]api.Quest, len(merged))
	for _, quest := range merged {
		byID[quest.ID] = quest
	}
	if byID["cachedActive"].DiscoveredRegion != "CACHE" {
		t.Fatalf("expected restored quest to be tagged CACHE, got %+v", byID["cachedActive"])
	}
	if _, ok := byID["cachedExpired"]; ok {
		t.Fatal("expired quest must not be restored")
	}
	if _, ok := byID["cachedClaimed"]; ok {
		t.Fatal("claimed quest must not be restored")
	}
}

func TestMergeSkipsStaleAndForeignSnapshots(t *testing.T) {
	now := time.Now().UTC()
	future := now.Add(time.Hour).Format(time.RFC3339)
	snapshot := &Snapshot{
		UserID:  "user-1",
		SavedAt: now.Add(-3 * time.Hour),
		Quests:  []api.Quest{{ID: "cached", Config: api.QuestConfig{ExpiresAt: future}}},
	}
	if merged := Merge(nil, snapshot, "user-1", now, time.Hour); len(merged) != 0 {
		t.Fatalf("stale snapshot must be ignored, got %+v", merged)
	}
	fresh := &Snapshot{UserID: "user-2", SavedAt: now, Quests: []api.Quest{{ID: "cached", Config: api.QuestConfig{ExpiresAt: future}}}}
	if merged := Merge(nil, fresh, "user-1", now, time.Hour); len(merged) != 0 {
		t.Fatalf("foreign snapshot must be ignored, got %+v", merged)
	}
}

func TestActiveQuest(t *testing.T) {
	now := time.Now().UTC()
	if !ActiveQuest(api.Quest{}, now) {
		t.Fatal("quest without expiry should be active")
	}
	expired := api.Quest{Config: api.QuestConfig{ExpiresAt: now.Add(-time.Minute).Format(time.RFC3339)}}
	if ActiveQuest(expired, now) {
		t.Fatal("expired quest should not be active")
	}
	claimed := api.Quest{UserStatus: &api.UserStatus{ClaimedAt: stringPtr(now.Format(time.RFC3339))}}
	if ActiveQuest(claimed, now) {
		t.Fatal("claimed quest should not be active")
	}
}
