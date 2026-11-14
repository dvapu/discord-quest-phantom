package pacing

import (
	"context"
	"testing"
	"time"
)

func TestDurationWithoutSpreadIsExact(t *testing.T) {
	if got := Duration(3*time.Second, 0); got != 3*time.Second {
		t.Fatalf("expected an exact delay, got %v", got)
	}
	if got := Duration(0, 0.2); got != 0 {
		t.Fatalf("expected zero to stay zero, got %v", got)
	}
}

func TestDurationStaysInsideTheBand(t *testing.T) {
	base := 20 * time.Second
	spread := 0.15
	min := time.Duration(float64(base) * (1 - spread))
	max := time.Duration(float64(base) * (1 + spread))

	for i := 0; i < 500; i++ {
		got := Duration(base, spread)
		if got < min || got > max {
			t.Fatalf("delay %v outside the %v..%v band", got, min, max)
		}
	}
}

func TestDurationVaries(t *testing.T) {
	seen := make(map[time.Duration]bool)
	for i := 0; i < 50; i++ {
		seen[Duration(time.Minute, 0.1)] = true
	}
	if len(seen) < 10 {
		t.Fatalf("expected varied delays, saw only %d distinct values", len(seen))
	}
}

func TestSleepHonoursCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if Sleep(ctx, 10*time.Second, 0.1) {
		t.Fatal("expected cancellation to be reported")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("cancellation took too long: %v", elapsed)
	}
}
