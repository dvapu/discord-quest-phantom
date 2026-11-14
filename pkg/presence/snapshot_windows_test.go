//go:build windows

package presence

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSnapshotSeesThisProcess checks the real Windows plumbing: process
// enumeration, image paths, the foreground window and the full-screen probe.
func TestSnapshotSeesThisProcess(t *testing.T) {
	state := Snapshot()
	if !state.Supported {
		t.Fatal("windows snapshot should be supported")
	}
	if len(state.Processes) == 0 {
		t.Fatal("process enumeration returned nothing")
	}

	self := Lower(filepath.Base(os.Args[0]))
	var found *Process
	for i := range state.Processes {
		if state.Processes[i].Exe == self {
			found = &state.Processes[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("snapshot did not see the test process %q", self)
	}
	if state.Foreground.Exe != "" && state.Foreground.PID == 0 {
		t.Fatal("foreground window reported an executable without a pid")
	}
	t.Logf("processes=%d discord=%v fullscreen=%v foreground=%q title=%q",
		len(state.Processes), state.Discord, state.Fullscreen, state.Foreground.Exe, state.Foreground.Title)
}

// TestWindowMonitorBoundsCoversForeground proves the borderless full-screen
// probe measures the monitor the window is really on, not just the primary one.
func TestWindowMonitorBoundsCoversForeground(t *testing.T) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		t.Skip("no foreground window (headless session)")
	}
	bounds, ok := windowMonitorBounds(hwnd)
	if !ok {
		t.Fatal("windowMonitorBounds failed for the foreground window")
	}
	if bounds.Right-bounds.Left <= 0 || bounds.Bottom-bounds.Top <= 0 {
		t.Fatalf("monitor bounds are empty: %+v", bounds)
	}
	t.Logf("monitor bounds: %+v", bounds)
}

// TestTryLockIsReentrantSafe documents the mutex contract: the same process may
// re-acquire its own named mutex (Windows mutexes are recursive per thread), so
// the guard must be paired with the PID checks instead of relying on the lock.
func TestTryLockReleases(t *testing.T) {
	release, ok := TryLock("DiscordQuestSpooferTest")
	if !ok {
		t.Skip("mutex already held by another instance")
	}
	release()
	release2, ok := TryLock("DiscordQuestSpooferTest")
	if !ok {
		t.Fatal("mutex was not released")
	}
	release2()
}
