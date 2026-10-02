package spoofer

import (
	"testing"
)

func TestResolveExecutable(t *testing.T) {
	reg := &Registry{
		apps: map[string]DetectableApp{
			"app1": {
				ID:   "app1",
				Name: "Test Win Game",
				Executables: []DetectableExecutable{
					{Name: "launcher.exe", OS: "win32", IsLauncher: true},
					{Name: "game.exe", OS: "win32", IsLauncher: false},
				},
			},
			"app2": {
				ID:   "app2",
				Name: "Test Linux Game",
				Executables: []DetectableExecutable{
					{Name: "linux_binary", OS: "linux", IsLauncher: false},
					{Name: "win_binary.exe", OS: "win32", IsLauncher: false},
				},
			},
		},
	}

	// Test Windows resolution: should pick non-launcher game.exe
	exe, title, err := reg.ResolveExecutable("app1", "windows")
	if err != nil {
		t.Fatalf("unexpected error resolving app1 on windows: %v", err)
	}
	if exe != "game.exe" {
		t.Errorf("expected game.exe, got %s", exe)
	}
	if title != "Test Win Game" {
		t.Errorf("expected title 'Test Win Game', got %s", title)
	}

	// Test Linux resolution: should pick linux_binary
	exeLinux, _, err := reg.ResolveExecutable("app2", "linux")
	if err != nil {
		t.Fatalf("unexpected error resolving app2 on linux: %v", err)
	}
	if exeLinux != "linux_binary" {
		t.Errorf("expected linux_binary, got %s", exeLinux)
	}

	// Test Linux fallback for Windows-only games (Wine/Proton)
	exeWine, _, err := reg.ResolveExecutable("app1", "linux")
	if err != nil {
		t.Fatalf("unexpected error resolving app1 on linux with wine: %v", err)
	}
	if exeWine != "game.exe" {
		t.Errorf("expected game.exe fallback on linux, got %s", exeWine)
	}
}
