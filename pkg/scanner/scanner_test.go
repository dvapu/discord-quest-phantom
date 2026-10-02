package scanner

import (
	"testing"

	"discord-quest-completer/pkg/api"
)

func TestAnalyzeQuestStates(t *testing.T) {
	timeStr := "2026-10-02T12:00:00Z"

	// Case 1: Claimed
	qClaimed := api.Quest{
		ID: "q1",
		UserStatus: &api.UserStatus{
			ClaimedAt: &timeStr,
		},
	}
	aqClaimed := AnalyzeQuest(qClaimed)
	if aqClaimed.State != StateClaimed {
		t.Errorf("expected state CLAIMED, got %v", aqClaimed.State)
	}

	// Case 2: Completed (unclaimed)
	qCompleted := api.Quest{
		ID: "q2",
		UserStatus: &api.UserStatus{
			CompletedAt: &timeStr,
		},
	}
	aqCompleted := AnalyzeQuest(qCompleted)
	if aqCompleted.State != StateCompleted {
		t.Errorf("expected state COMPLETED, got %v", aqCompleted.State)
	}

	// Case 3: Enrolled (in progress)
	qEnrolled := api.Quest{
		ID: "q3",
		UserStatus: &api.UserStatus{
			EnrolledAt: &timeStr,
		},
	}
	aqEnrolled := AnalyzeQuest(qEnrolled)
	if aqEnrolled.State != StateEnrolled {
		t.Errorf("expected state ENROLLED, got %v", aqEnrolled.State)
	}

	// Case 4: Available (un-enrolled)
	qAvailable := api.Quest{
		ID:         "q4",
		UserStatus: nil,
	}
	aqAvailable := AnalyzeQuest(qAvailable)
	if aqAvailable.State != StateAvailable {
		t.Errorf("expected state AVAILABLE, got %v", aqAvailable.State)
	}
}

func TestAnalyzeQuestTargetAndProgress(t *testing.T) {
	name := "War Thunder"
	q := api.Quest{
		ID: "1550081499628441600",
		Config: api.QuestConfig{
			Messages: api.QuestMessages{
				GameTitle: &name,
			},
			TaskConfigV2: &api.TaskConfig{
				Tasks: map[string]api.TaskRequirement{
					"PLAY_ON_DESKTOP": {
						Type:   "PLAY_ON_DESKTOP",
						Target: 900,
						Applications: []api.ApplicationRef{
							{ID: "357607478105604096"},
						},
					},
				},
			},
		},
		UserStatus: &api.UserStatus{
			Progress: map[string]api.TaskProgress{
				"PLAY_ON_DESKTOP": {
					Value: 900,
				},
			},
		},
	}

	aq := AnalyzeQuest(q)
	if aq.Title != "War Thunder" {
		t.Errorf("expected title 'War Thunder', got %s", aq.Title)
	}
	if aq.TaskType != "PLAY_ON_DESKTOP" {
		t.Errorf("expected task type PLAY_ON_DESKTOP, got %s", aq.TaskType)
	}
	if aq.TargetSec != 900 {
		t.Errorf("expected target 900, got %d", aq.TargetSec)
	}
	if aq.CurrentSec != 900 {
		t.Errorf("expected current 900, got %f", aq.CurrentSec)
	}
	if aq.AppID != "357607478105604096" {
		t.Errorf("expected app ID 357607478105604096, got %s", aq.AppID)
	}
}
