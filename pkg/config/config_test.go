package config

import (
	"os"
	"testing"
	"time"
)

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "******"},
		{"short", "******"},
		{"123456789012", "******"},
		{"NTk5ODExNjkwODExNTY3MzE1MQ.Gy_abc.XYZ123456789Xk", "NTk5OD...89Xk"},
	}

	for _, tc := range tests {
		actual := MaskToken(tc.input)
		if actual != tc.expected {
			t.Errorf("MaskToken(%q) = %q; want %q", tc.input, actual, tc.expected)
		}
	}
}

func TestResolveTokenFromEnv(t *testing.T) {
	testToken := "TEST_ENV_TOKEN_12345"
	os.Setenv("DISCORD_TOKEN", testToken)
	defer os.Unsetenv("DISCORD_TOKEN")

	resolved, err := ResolveToken("")
	if err != nil {
		t.Fatalf("unexpected error resolving token from env: %v", err)
	}
	if resolved != testToken {
		t.Errorf("got %q, want %q", resolved, testToken)
	}
}

func TestResolveTokenCliPrecedence(t *testing.T) {
	os.Setenv("DISCORD_TOKEN", "ENV_TOKEN")
	defer os.Unsetenv("DISCORD_TOKEN")

	resolved, err := ResolveToken("CLI_TOKEN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != "CLI_TOKEN" {
		t.Errorf("CLI token did not take precedence: got %q", resolved)
	}
}

func TestConfigStructFields(t *testing.T) {
	cfg := Config{
		Token:        "MOCK_TOKEN",
		PollInterval: 45 * time.Second,
		AutoAccept:   true,
		DryRun:       true,
		QuestID:      "123456789",
		Duration:     15 * time.Minute,
		Concurrency:  5,
		Region:       "all",
		EnablePortal: true,
		PortalPort:   8080,
	}

	if cfg.QuestID != "123456789" {
		t.Errorf("expected QuestID 123456789, got %s", cfg.QuestID)
	}
	if cfg.Concurrency != 5 {
		t.Errorf("expected Concurrency 5, got %d", cfg.Concurrency)
	}
	if cfg.Region != "all" {
		t.Errorf("expected Region all, got %s", cfg.Region)
	}
	if !cfg.EnablePortal || cfg.PortalPort != 8080 {
		t.Errorf("expected portal enabled on port 8080")
	}
	if cfg.Duration != 15*time.Minute {
		t.Errorf("expected Duration 15m, got %v", cfg.Duration)
	}
	if !cfg.AutoAccept {
		t.Errorf("expected AutoAccept true")
	}
}

func TestParseHypeSquadAction(t *testing.T) {
	tests := []struct {
		input       string
		wantHouse   int
		wantLeave   bool
		wantErr     bool
	}{
		{"", 0, false, false},
		{"bravery", 1, false, false},
		{"1", 1, false, false},
		{"house1", 1, false, false},
		{"brilliance", 2, false, false},
		{"2", 2, false, false},
		{"house2", 2, false, false},
		{"balance", 3, false, false},
		{"3", 3, false, false},
		{"house3", 3, false, false},
		{"leave", 0, true, false},
		{"0", 0, true, false},
		{"none", 0, true, false},
		{"remove", 0, true, false},
		{"invalid", 0, false, true},
	}

	for _, tt := range tests {
		houseID, isLeave, err := ParseHypeSquadAction(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseHypeSquadAction(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
		}
		if houseID != tt.wantHouse {
			t.Errorf("ParseHypeSquadAction(%q) houseID = %d, want %d", tt.input, houseID, tt.wantHouse)
		}
		if isLeave != tt.wantLeave {
			t.Errorf("ParseHypeSquadAction(%q) isLeave = %v, want %v", tt.input, isLeave, tt.wantLeave)
		}
	}
}

