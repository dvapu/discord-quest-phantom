package spoofer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLaunchGameNestedPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dq_test_launch_*")
	if err != nil {
		t.Fatal(err)
	}

	sp, err := NewSpoofer(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Case 1: Simple flat exe name
	proc1, err1 := sp.LaunchGame(context.Background(), "357607478105604096", "War Thunder", "aces.exe")
	if err1 != nil {
		t.Errorf("LaunchGame failed for flat exe name: %v", err1)
	} else {
		t.Logf("LaunchGame succeeded for aces.exe with PID %d", proc1.PID())
		if err := proc1.Stop(); err != nil {
			t.Errorf("Failed to stop proc1: %v", err)
		}
	}

	// Case 2: Nested path exe name (e.g. Helldivers 2)
	proc2, err2 := sp.LaunchGame(context.Background(), "1205090671527071784", "HELLDIVERS 2", filepath.Join("bin", "helldivers2.exe"))
	if err2 != nil {
		t.Errorf("LaunchGame failed for nested exe name: %v", err2)
	} else {
		t.Logf("LaunchGame succeeded for bin/helldivers2.exe with PID %d", proc2.PID())
		if err := proc2.Stop(); err != nil {
			t.Errorf("Failed to stop proc2: %v", err)
		}
	}
}
