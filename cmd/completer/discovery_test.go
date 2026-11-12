package main

import (
	"testing"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/scanner"
)

func TestMergeRefreshedKeepsProxyOnlyQuests(t *testing.T) {
	existing := []scanner.AnalyzedQuest{
		{ID: "q1", Quest: api.Quest{ID: "q1"}},
		{ID: "proxy-only", Quest: api.Quest{ID: "proxy-only"}},
	}
	refreshed := []api.Quest{{ID: "q1", DiscoveredRegion: "LOCAL"}}
	merged := mergeRefreshed(existing, refreshed)
	if len(merged) != 2 {
		t.Fatalf("expected 2 merged quests, got %+v", merged)
	}
	byID := make(map[string]api.Quest, len(merged))
	for _, quest := range merged {
		byID[quest.ID] = quest
	}
	if byID["q1"].DiscoveredRegion != "LOCAL" {
		t.Fatalf("expected the refreshed q1 to win, got %+v", byID["q1"])
	}
	if _, ok := byID["proxy-only"]; !ok {
		t.Fatal("proxy-only quest must survive a direct refresh")
	}
}

func TestRegionInfoFor(t *testing.T) {
	previous := sweepTargets
	defer func() { sweepTargets = previous }()
	sweepTargets = []api.RegionTarget{{Code: "US", Locale: "en-US", Timezone: "America/New_York"}}

	info := regionInfoFor("us")
	if info.Code != "US" || info.Locale != "en-US" || info.Timezone != "America/New_York" {
		t.Fatalf("unexpected region info: %+v", info)
	}
	fallback := regionInfoFor("UNKNOWN")
	if fallback.Code != "UNKNOWN" || fallback.Locale != "en-US" {
		t.Fatalf("unexpected fallback region info: %+v", fallback)
	}
}
