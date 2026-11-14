package presence

import "testing"

func baseState() State {
	return State{
		Supported: true,
		Discord:   true,
		Processes: []Process{
			{PID: 10, Exe: "explorer.exe"},
			{PID: 20, Exe: "discord.exe"},
		},
	}
}

func TestEvaluateUnsupportedIsReady(t *testing.T) {
	if got := Evaluate(State{}, Target{Exe: "game.exe"}, nil); got.Kind != Ready {
		t.Fatalf("unsupported platform should not block, got %q", got.Kind)
	}
}

func TestEvaluateNoDiscord(t *testing.T) {
	s := baseState()
	s.Discord = false
	if got := Evaluate(s, Target{Exe: "game.exe"}, nil); got.Kind != NoDiscord {
		t.Fatalf("want no-discord, got %q", got.Kind)
	}
}

func TestEvaluateRealSessionWins(t *testing.T) {
	s := baseState()
	s.Processes = append(s.Processes, Process{PID: 30, Exe: "game.exe"})
	got := Evaluate(s, Target{Exe: "Game.exe", Title: "Game"}, nil)
	if got.Kind != RealSession {
		t.Fatalf("want real-session, got %q (%s)", got.Kind, got.Detail)
	}
}

func TestEvaluateOwnStubIsNotRealSession(t *testing.T) {
	s := baseState()
	s.Processes = append(s.Processes, Process{PID: 30, Exe: "game.exe"})
	got := Evaluate(s, Target{Exe: "game.exe", StubPID: 30}, nil)
	if got.Kind != Ready {
		t.Fatalf("our own stub must not count as a real session, got %q", got.Kind)
	}
}

func TestEvaluateFullscreenPauses(t *testing.T) {
	s := baseState()
	s.Fullscreen = true
	if got := Evaluate(s, Target{Exe: "game.exe"}, nil); got.Kind != WaitFullscreen {
		t.Fatalf("want wait-fullscreen, got %q", got.Kind)
	}
}

func TestEvaluateOtherGamePauses(t *testing.T) {
	s := baseState()
	s.Processes = append(s.Processes, Process{PID: 40, Exe: "other.exe"})
	got := Evaluate(s, Target{Exe: "game.exe"}, map[string]string{"other.exe": "Other Game"})
	if got.Kind != WaitOtherGame {
		t.Fatalf("want wait-other-game, got %q", got.Kind)
	}
	if got.Detail != "Other Game is running" {
		t.Fatalf("unexpected detail %q", got.Detail)
	}
}

func TestEvaluateStubPidIsIgnoredForOtherGames(t *testing.T) {
	s := baseState()
	s.Processes = append(s.Processes, Process{PID: 40, Exe: "other.exe"})
	// A process with our own stub PID must not be treated as a foreign game.
	got := Evaluate(s, Target{Exe: "game.exe", StubPID: 40}, map[string]string{"other.exe": "Other Game"})
	if got.Kind != Ready {
		t.Fatalf("own stub should be ignored, got %q", got.Kind)
	}
}

func TestEvaluateQuietMachineIsReady(t *testing.T) {
	if got := Evaluate(baseState(), Target{Exe: "game.exe"}, nil); got.Kind != Ready {
		t.Fatalf("want ready, got %q", got.Kind)
	}
}

func TestHasExeIsCaseInsensitive(t *testing.T) {
	s := baseState()
	if !s.HasExe("Discord.EXE") {
		t.Fatal("HasExe should compare case-insensitively")
	}
	if s.HasExe("steam.exe") {
		t.Fatal("HasExe reported a process that is not running")
	}
}
