//go:build windows

package spoofer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type windowsSpoofer struct {
	cacheDir string
	registry *Registry
}

func newPlatformSpoofer(cacheDir string, reg *Registry) (Spoofer, error) {
	return &windowsSpoofer{
		cacheDir: cacheDir,
		registry: reg,
	}, nil
}

type windowsProcess struct {
	cmd     *exec.Cmd
	exeName string
	title   string
}

func (p *windowsProcess) PID() int               { return p.cmd.Process.Pid }
func (p *windowsProcess) ExecutableName() string { return p.exeName }
func (p *windowsProcess) GameTitle() string      { return p.title }
func (p *windowsProcess) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

func (s *windowsSpoofer) ResolveExecutable(appID string) (string, string, error) {
	return s.registry.ResolveExecutable(appID, "windows")
}

func (s *windowsSpoofer) LaunchGame(ctx context.Context, appID string, gameTitle string, exeName string) (GameProcess, error) {
	// Build an install-shaped path: <cache>\stubs\<Game Name>\<relative exe path>.
	// Discord resolves process image names, but a path full of opaque app ids is
	// an unnecessary giveaway to anything that inspects the process list.
	folder := slug(gameTitle)
	if folder == "" {
		folder = appID
	}
	destExe := filepath.Join(s.cacheDir, "stubs", folder, filepath.FromSlash(exeName))

	// Ensure all intermediate parent directories exist recursively
	if err := os.MkdirAll(filepath.Dir(destExe), 0755); err != nil {
		return nil, fmt.Errorf("failed creating stub directory: %w", err)
	}

	if _, err := os.Stat(destExe); os.IsNotExist(err) {
		selfPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed locating self executable: %w", err)
		}
		// A hard link costs no disk space and leaves one binary on the volume
		// instead of one 10 MB copy per game, which is also quieter for on-access
		// antivirus scanners. Fall back to a copy when the volume refuses links.
		if linkErr := os.Link(selfPath, destExe); linkErr != nil {
			if err := copyFile(selfPath, destExe); err != nil {
				return nil, fmt.Errorf("failed creating dummy game executable: %w", err)
			}
		}
	}

	cmd := exec.Command(destExe, "--stub", "--title", gameTitle)
	// Real games run with their install directory as the working directory.
	cmd.Dir = filepath.Dir(destExe)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: false,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to launch game process %s: %w", exeName, err)
	}

	return &windowsProcess{
		cmd:     cmd,
		exeName: exeName,
		title:   gameTitle,
	}, nil
}

// slug turns a game title into a filesystem-safe folder name.
func slug(title string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range title {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
		if b.Len() >= 48 {
			break
		}
	}
	return strings.Trim(b.String(), "-")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
