package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"discord-quest-completer/pkg/api"
	"discord-quest-completer/pkg/config"
	"discord-quest-completer/pkg/i18n"
)

func TestHandleBadgeMaximizer_DryRunGuards(t *testing.T) {
	i18n.InitLanguage("en")
	var httpCallCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&httpCallCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient("mock_token", 500000)
	client.SetAPIBaseForTesting(server.URL)

	ctx := context.Background()

	// 1. DryRun with explicit hypesquad join
	cfgDryJoin := &config.Config{
		DryRun:    true,
		HypeSquad: "bravery",
		MaxBadge:  true,
	}
	user := &api.UserMe{}
	handleBadgeMaximizer(ctx, cfgDryJoin, client, user)
	if count := atomic.LoadInt32(&httpCallCount); count != 0 {
		t.Fatalf("DryRun joined HypeSquad via HTTP! callCount = %d", count)
	}

	// 2. DryRun with explicit hypesquad leave
	cfgDryLeave := &config.Config{
		DryRun:    true,
		HypeSquad: "leave",
		MaxBadge:  true,
	}
	handleBadgeMaximizer(ctx, cfgDryLeave, client, user)
	if count := atomic.LoadInt32(&httpCallCount); count != 0 {
		t.Fatalf("DryRun left HypeSquad via HTTP! callCount = %d", count)
	}

	// 3. DryRun with max-badge auto enroll
	cfgDryAuto := &config.Config{
		DryRun:   true,
		MaxBadge: true,
	}
	handleBadgeMaximizer(ctx, cfgDryAuto, client, user)
	if count := atomic.LoadInt32(&httpCallCount); count != 0 {
		t.Fatalf("DryRun auto-enrolled HypeSquad via HTTP! callCount = %d", count)
	}
}

func TestHandleBadgeMaximizer_LiveExecution(t *testing.T) {
	i18n.InitLanguage("en")
	var hypeMethod, hypePath string
	var lastHouseID int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/@me" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           "12345",
				"username":     "tester",
				"flags":        int64(api.FlagHouseBrilliance),
				"public_flags": int64(api.FlagHouseBrilliance),
			})
			return
		}
		if r.URL.Path == "/hypesquad/online" {
			hypeMethod = r.Method
			hypePath = r.URL.Path
			if r.Method == http.MethodPost {
				var body api.HypeSquadPayload
				_ = json.NewDecoder(r.Body).Decode(&body)
				lastHouseID = body.HouseID
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient("mock_token", 500000)
	client.SetAPIBaseForTesting(server.URL)

	ctx := context.Background()

	// 1. Explicit Join Brilliance
	cfgJoin := &config.Config{
		DryRun:    false,
		HypeSquad: "brilliance",
	}
	user := &api.UserMe{}
	user = handleBadgeMaximizer(ctx, cfgJoin, client, user)
	if hypeMethod != http.MethodPost || hypePath != "/hypesquad/online" || lastHouseID != 2 {
		t.Errorf("expected POST /hypesquad/online with house 2, got %s %s house %d", hypeMethod, hypePath, lastHouseID)
	}
	if !user.HasHypeSquad() {
		t.Errorf("expected user to reflect refreshed hypesquad")
	}

	// 2. Already owned does not call Join
	hypeMethod = ""
	hypePath = ""
	cfgAuto := &config.Config{
		DryRun:   false,
		MaxBadge: true,
	}
	user = handleBadgeMaximizer(ctx, cfgAuto, client, user)
	if hypeMethod != "" || hypePath != "" {
		t.Errorf("expected no HTTP call for already owned house, got %s %s", hypeMethod, hypePath)
	}
}
