//go:build !windows

package spoofer

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RunStub handles dummy process execution on non-Windows platforms (Linux, macOS).
func RunStub(title string) error {
	fmt.Printf("[Dummy Game Process] Running stub for: %s (PID: %d)\n", title, os.Getpid())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case sig := <-sigChan:
			fmt.Printf("[Dummy Game Process] Received %v, terminating cleanly...\n", sig)
			return nil
		case <-ticker.C:
		}
	}
}
