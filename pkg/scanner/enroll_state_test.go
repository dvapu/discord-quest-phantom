package scanner

import (
	"testing"
)

func TestApplyEnrolledState(t *testing.T) {
	quests := []AnalyzedQuest{
		{ID: "q-available", State: StateAvailable},
		{ID: "q-claimed", State: StateClaimed},
		{ID: "q-other", State: StateAvailable},
	}
	updated := ApplyEnrolledState(quests, []string{"q-available", "q-claimed", "q-missing"})
	if updated != 1 {
		t.Fatalf("expected exactly 1 state update, got %d", updated)
	}
	if quests[0].State != StateEnrolled {
		t.Fatalf("available quest should become ENROLLED, got %s", quests[0].State)
	}
	if quests[1].State != StateClaimed {
		t.Fatalf("claimed quest must keep its state, got %s", quests[1].State)
	}
	if quests[2].State != StateAvailable {
		t.Fatalf("unrelated quest must keep its state, got %s", quests[2].State)
	}
}

func TestApplyEnrolledStateEmptyInputs(t *testing.T) {
	quests := []AnalyzedQuest{{ID: "q1", State: StateAvailable}}
	if updated := ApplyEnrolledState(nil, []string{"q1"}); updated != 0 {
		t.Fatalf("expected no updates for empty quest list, got %d", updated)
	}
	if updated := ApplyEnrolledState(quests, nil); updated != 0 {
		t.Fatalf("expected no updates for empty id list, got %d", updated)
	}
}
