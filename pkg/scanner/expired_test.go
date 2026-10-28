package scanner

import (
	"strings"
	"testing"

	"discord-quest-completer/pkg/api"
)

// Expired quests must be collapsed out of the table by default; otherwise a
// long-idle account spams the console with dead campaigns on every scan.
func TestFormatQuestTableHidesExpiredByDefault(t *testing.T) {
	past := "2020-01-01T00:00:00Z"
	future := "2999-01-01T00:00:00Z"

	quests := []AnalyzedQuest{
		{ID: "alive", Title: "Alive Quest", State: StateEnrolled, TargetSec: 900, CurrentSec: 10, ExpiresAt: future},
		{ID: "dead1", Title: "Dead One", State: StateExpired, IsExpired: true, ExpiresAt: past},
		{ID: "dead2", Title: "Dead Two", State: StateExpired, IsExpired: true, ExpiresAt: past},
	}

	hidden := FormatQuestTable(quests, false)
	if strings.Contains(hidden, "Dead One") || strings.Contains(hidden, "Dead Two") {
		t.Fatalf("expired quests should be hidden by default:\n%s", hidden)
	}
	if !strings.Contains(hidden, "Alive Quest") {
		t.Fatalf("active quest should still be listed:\n%s", hidden)
	}
	if !strings.Contains(hidden, "2 expired") {
		t.Fatalf("expected a single collapsed expired notice:\n%s", hidden)
	}

	// -debug (showExpired) must reveal them again.
	shown := FormatQuestTable(quests, true)
	if !strings.Contains(shown, "Dead One") || !strings.Contains(shown, "Dead Two") {
		t.Fatalf("showExpired should list expired quests:\n%s", shown)
	}
}

// The signature must change with progress so the daemon prints a new table
// when something actually happens, and stay stable when nothing does.
func TestTableSignatureStability(t *testing.T) {
	base := []AnalyzedQuest{
		{ID: "a", State: StateEnrolled, TargetSec: 900, CurrentSec: 100},
		{ID: "b", State: StateClaimed, TargetSec: 300, CurrentSec: 300},
	}
	sig1 := TableSignature(base)
	if sig1 != TableSignature(base) {
		t.Fatal("signature must be deterministic for identical input")
	}

	progressed := []AnalyzedQuest{
		{ID: "a", State: StateEnrolled, TargetSec: 900, CurrentSec: 120},
		{ID: "b", State: StateClaimed, TargetSec: 300, CurrentSec: 300},
	}
	if sig1 == TableSignature(progressed) {
		t.Fatal("signature must change when quest progress advances")
	}

	claimed := []AnalyzedQuest{
		{ID: "a", State: StateClaimed, TargetSec: 900, CurrentSec: 900},
		{ID: "b", State: StateClaimed, TargetSec: 300, CurrentSec: 300},
	}
	if sig1 == TableSignature(claimed) {
		t.Fatal("signature must change when a quest becomes claimed")
	}
}

// Auto-enrollment must never target expired quests.
func TestAutoEnrollSkipsExpired(t *testing.T) {
	past := "2020-01-01T00:00:00Z"
	q := api.Quest{
		ID: "dead",
		Config: api.QuestConfig{
			ExpiresAt: past,
		},
	}
	aq := AnalyzeQuest(q)
	if !aq.IsExpired {
		t.Fatal("quest with a past expires_at must be marked expired")
	}
	if aq.State != StateExpired {
		t.Fatalf("expected EXPIRED state, got %s", aq.State)
	}
}
