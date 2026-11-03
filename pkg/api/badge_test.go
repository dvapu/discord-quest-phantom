package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserMe_Badges(t *testing.T) {
	// Case 1: No badges
	userEmpty := &UserMe{}
	if userEmpty.HasHypeSquad() {
		t.Errorf("expected no hypesquad for empty user")
	}
	if userEmpty.FormatBadges() != "None" {
		t.Errorf("expected 'None', got %s", userEmpty.FormatBadges())
	}

	// Case 2: HypeSquad Bravery
	userBravery := &UserMe{
		Flags: FlagHouseBravery,
	}
	if !userBravery.HasHypeSquad() {
		t.Errorf("expected user to have hypesquad")
	}
	houseID, houseName := userBravery.GetHypeSquadHouse()
	if houseID != HypeSquadHouseBravery || houseName != "House of Bravery" {
		t.Errorf("expected Bravery (1), got %d (%s)", houseID, houseName)
	}
	if !strings.Contains(userBravery.FormatBadges(), "HypeSquad Bravery") {
		t.Errorf("expected format to contain HypeSquad Bravery: %s", userBravery.FormatBadges())
	}

	// Case 3: Brilliance + Active Developer + Nitro in PublicFlags
	userMulti := &UserMe{
		PublicFlags: FlagHouseBrilliance | FlagActiveDeveloper,
		PremiumType: 2, // Nitro
	}
	houseID, houseName = userMulti.GetHypeSquadHouse()
	if houseID != HypeSquadHouseBrilliance || houseName != "House of Brilliance" {
		t.Errorf("expected Brilliance (2), got %d (%s)", houseID, houseName)
	}
	formatted := userMulti.FormatBadges()
	if !strings.Contains(formatted, "HypeSquad Brilliance") || !strings.Contains(formatted, "Active Developer") || !strings.Contains(formatted, "Nitro") {
		t.Errorf("expected badges to contain Brilliance, Active Developer and Nitro, got: %s", formatted)
	}

	// Case 4: Balance
	userBalance := &UserMe{
		Flags: FlagHouseBalance,
	}
	houseID, houseName = userBalance.GetHypeSquadHouse()
	if houseID != HypeSquadHouseBalance || houseName != "House of Balance" {
		t.Errorf("expected Balance (3), got %d (%s)", houseID, houseName)
	}
}

func TestClient_JoinAndLeaveHypeSquad(t *testing.T) {
	var receivedMethod string
	var receivedPath string
	var receivedBody HypeSquadPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("mock_token", 500000)
	client.apiBase = server.URL

	ctx := context.Background()

	// 1. Join House of Brilliance (2)
	err := client.JoinHypeSquad(ctx, HypeSquadHouseBrilliance)
	if err != nil {
		t.Fatalf("unexpected error joining hypesquad: %v", err)
	}
	if receivedMethod != http.MethodPost || receivedPath != "/hypesquad/online" {
		t.Errorf("expected POST /hypesquad/online, got %s %s", receivedMethod, receivedPath)
	}
	if receivedBody.HouseID != 2 {
		t.Errorf("expected house_id 2, got %d", receivedBody.HouseID)
	}

	// 2. Invalid house ID
	if err := client.JoinHypeSquad(ctx, 99); err == nil {
		t.Errorf("expected error for invalid house ID 99")
	}

	// 3. Leave HypeSquad
	err = client.LeaveHypeSquad(ctx)
	if err != nil {
		t.Fatalf("unexpected error leaving hypesquad: %v", err)
	}
	if receivedMethod != http.MethodDelete || receivedPath != "/hypesquad/online" {
		t.Errorf("expected DELETE /hypesquad/online, got %s %s", receivedMethod, receivedPath)
	}
}
