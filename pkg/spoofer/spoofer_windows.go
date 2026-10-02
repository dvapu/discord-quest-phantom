//go:build windows

package spoofer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	// Normalize exeName path separators (catalog can use forward slashes, e.g. "bin/helldivers2.exe")
	destExe := filepath.Join(s.cacheDir, "stubs", appID, filepath.FromSlash(exeName))

	// Fix 1: Ensure all intermediate parent directories exist recursively
	if err := os.MkdirAll(filepath.Dir(destExe), 0755); err != nil {
		return nil, fmt.Errorf("failed creating stub directory: %w", err)
	}

	// Fix 2: Remove knownSrcWin shortcut and create self-contained stub binary by copying current executable
	if _, err := os.Stat(destExe); os.IsNotExist(err) {
		selfPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed locating self executable: %w", err)
		}
		if err := copyFile(selfPath, destExe); err != nil {
			return nil, fmt.Errorf("failed creating dummy game executable: %w", err)
		}
	}

	cmd := exec.Command(destExe, "--stub", "--title", gameTitle)
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
