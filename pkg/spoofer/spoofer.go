package spoofer

import (
	"context"
	"fmt"
)

// GameProcess represents an active OS spoofed game process.
type GameProcess interface {
	PID() int
	ExecutableName() string
	GameTitle() string
	Stop() error
}

// Spoofer provides game process emulation and executable resolution.
type Spoofer interface {
	LaunchGame(ctx context.Context, appID string, gameTitle string, exeName string) (GameProcess, error)
	ResolveExecutable(appID string) (exeName string, gameTitle string, err error)
}

// NewSpoofer instantiates the platform-appropriate process spoofer.
func NewSpoofer(cacheDir string) (Spoofer, error) {
	reg, err := LoadRegistry(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detectable registry: %w", err)
	}

	return newPlatformSpoofer(cacheDir, reg)
}
