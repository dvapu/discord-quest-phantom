//go:build linux

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

type linuxSpoofer struct {
	cacheDir string
	registry *Registry
}

func newPlatformSpoofer(cacheDir string, reg *Registry) (Spoofer, error) {
	return &linuxSpoofer{
		cacheDir: cacheDir,
		registry: reg,
	}, nil
}

type linuxProcess struct {
	cmd     *exec.Cmd
	exeName string
	title   string
}

func (p *linuxProcess) PID() int               { return p.cmd.Process.Pid }
func (p *linuxProcess) ExecutableName() string { return p.exeName }
func (p *linuxProcess) GameTitle() string      { return p.title }
func (p *linuxProcess) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	err := p.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

func (s *linuxSpoofer) ResolveExecutable(appID string) (string, string, error) {
	return s.registry.ResolveExecutable(appID, "linux")
}

func (s *linuxSpoofer) LaunchGame(ctx context.Context, appID string, gameTitle string, exeName string) (GameProcess, error) {
	// Normalize exeName path separators (catalog can use forward slashes, e.g. "bin/helldivers2.exe")
	destExe := filepath.Join(s.cacheDir, "stubs", appID, filepath.FromSlash(exeName))

	// Fix 1: Ensure all intermediate parent directories exist recursively
	if err := os.MkdirAll(filepath.Dir(destExe), 0755); err != nil {
		return nil, fmt.Errorf("failed creating linux stub directory: %w", err)
	}

	if _, err := os.Stat(destExe); os.IsNotExist(err) {
		selfPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed locating current executable: %w", err)
		}

		// Try symlink first, fallback to copy
		if err := os.Symlink(selfPath, destExe); err != nil {
			if err := copyFile(selfPath, destExe); err != nil {
				return nil, fmt.Errorf("failed creating linux stub link/copy: %w", err)
			}
		}
	}

	cmd := exec.Command(destExe, "--stub", "--title", gameTitle)
	cmd.Args[0] = exeName
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to launch linux game process: %w", err)
	}

	return &linuxProcess{
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
