// Package presence answers one question before the spoofer pretends to play a
// game: is somebody actually using this machine right now?
//
// The failure this exists to prevent: the tool farms a quest for game A while
// the owner is playing game B in the foreground. Discord's client can only
// report one game at a time, so the account's presence flips between the two,
// and progress for A stops while B is focused. Real users never produce that
// pattern. Pausing is both safer and more effective than running anyway.
package presence

import "strings"

// Process is one running process. Image names are lower-case so comparisons
// never depend on how Windows capitalized the file.
type Process struct {
	PID int
	Exe string
}

// Window is the foreground window.
type Window struct {
	PID   int
	Exe   string
	Title string
}

// State is a snapshot of the machine.
type State struct {
	// Supported is false on platforms where these checks do not exist; the
	// guards then stay out of the way instead of blocking the tool.
	Supported  bool
	Processes  []Process
	Foreground Window
	// Fullscreen is true when the OS reports a full-screen app (exclusive
	// game, presentation, full-screen video) or the foreground window covers
	// the whole screen with no title bar (borderless full-screen game).
	Fullscreen bool
	Discord    bool
}

// Lower normalizes an executable name for comparison.
func Lower(name string) string { return strings.ToLower(strings.TrimSpace(name)) }

// HasExe reports whether any process with this image name is running.
func (s State) HasExe(exe string) bool {
	exe = Lower(exe)
	for _, p := range s.Processes {
		if p.Exe == exe {
			return true
		}
	}
	return false
}

// Kind classifies the verdict.
type Kind string

const (
	// Ready means the machine is quiet and the stub may be shown.
	Ready Kind = "ready"
	// NoDiscord means the Discord desktop client is not running, so no local
	// game detection can happen and spoofing a desktop quest is pointless.
	NoDiscord Kind = "no-discord"
	// WaitFullscreen means the owner is in a full-screen app.
	WaitFullscreen Kind = "wait-fullscreen"
	// WaitOtherGame means a different, quest-relevant game is running.
	WaitOtherGame Kind = "wait-other-game"
	// RealSession means the quest's own game is already running for real.
	RealSession Kind = "real-session"
)

// Verdict is the guard's answer plus a human readable reason.
type Verdict struct {
	Kind   Kind
	Detail string
}

// Target describes the game the spoofer wants to show.
type Target struct {
	Exe     string
	Title   string
	StubPID int // PID of the stub already launched for this target, 0 if none
}

// Evaluate decides whether the machine allows showing target right now.
//
// other maps executable names (lower-case) of *other* eligible quest games to
// their display names. Those are the only foreign games the tool can name with
// confidence; anything else is covered by the full-screen check.
func Evaluate(s State, target Target, other map[string]string) Verdict {
	if !s.Supported {
		return Verdict{Kind: Ready}
	}
	if !s.Discord {
		return Verdict{Kind: NoDiscord, Detail: "Discord desktop client is not running"}
	}

	exe := Lower(target.Exe)
	if exe != "" {
		for _, p := range s.Processes {
			if p.Exe == exe && p.PID != target.StubPID {
				return Verdict{
					Kind:   RealSession,
					Detail: strings.TrimSpace(target.Title) + " is already running on this machine",
				}
			}
		}
	}

	if s.Fullscreen {
		return Verdict{Kind: WaitFullscreen, Detail: "a full-screen application is active"}
	}

	for _, p := range s.Processes {
		if p.PID == target.StubPID || p.Exe == exe {
			continue
		}
		if name, ok := other[p.Exe]; ok {
			return Verdict{Kind: WaitOtherGame, Detail: name + " is running"}
		}
	}

	return Verdict{Kind: Ready}
}
