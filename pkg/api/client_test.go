package api

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestBuildSuperProperties(t *testing.T) {
	b64Props := BuildSuperProperties(504649)
	decoded, err := base64.StdEncoding.DecodeString(b64Props)
	if err != nil {
		t.Fatalf("failed decoding super properties base64: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(decoded, &m); err != nil {
		t.Fatalf("failed unmarshaling json from decoded super props: %v", err)
	}

	if m["os"] != "Windows" {
		t.Errorf("expected os to be Windows, got %v", m["os"])
	}
	if m["browser"] != "Discord Client" {
		t.Errorf("expected browser to be Discord Client, got %v", m["browser"])
	}
	if m["client_build_number"] != float64(504649) {
		t.Errorf("expected build number 504649, got %v", m["client_build_number"])
	}
}

func TestUnmarshalQuestResponse(t *testing.T) {
	jsonPayload := `{
		"quests": [
			{
				"id": "1234567890",
				"config": {
					"id": "1234567890",
					"expires_at": "2026-10-31T00:00:00Z",
					"application": {
						"id": "999888",
						"name": "Test Game"
					},
					"messages": {
						"quest_name": "Test Quest"
					},
					"task_config_v2": {
						"tasks": {
							"PLAY_ON_DESKTOP": {
								"type": "PLAY_ON_DESKTOP",
								"target": 900,
								"applications": [{"id": "999888"}]
							}
						}
					}
				},
				"user_status": {
					"user_id": "555",
					"quest_id": "1234567890",
					"progress": {
						"PLAY_ON_DESKTOP": {
							"value": 450,
							"event_name": "PLAY_ON_DESKTOP"
						}
					}
				}
			}
		]
	}`

	var resp QuestResponse
	if err := json.Unmarshal([]byte(jsonPayload), &resp); err != nil {
		t.Fatalf("failed unmarshaling test quest: %v", err)
	}

	if len(resp.Quests) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(resp.Quests))
	}

	q := resp.Quests[0]
	if q.ID != "1234567890" {
		t.Errorf("expected quest ID 1234567890, got %s", q.ID)
	}
	if q.Config.Application.Name != "Test Game" {
		t.Errorf("expected application name Test Game, got %s", q.Config.Application.Name)
	}
	if q.UserStatus == nil || q.UserStatus.Progress["PLAY_ON_DESKTOP"].Value != 450 {
		t.Errorf("expected progress value 450")
	}
}
