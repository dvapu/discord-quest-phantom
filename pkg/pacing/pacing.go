// Package pacing adds realistic variance to request timing. Fixed sleeps leave
// a clockwork cadence in traffic, which is a textbook bot signature; a bell
// shaped jitter keeps the average delay while removing the regularity.
package pacing

import (
	"context"
	"math/rand"
	"time"
)

// DefaultSpread is the documented default for interactive delays.
const DefaultSpread = 0.15

// Duration returns base scaled by a bell-shaped random factor within +/- spread.
// A spread of 0.15 gives roughly 85%-115% of base, with values near the edges
// being rare. Negative or zero spread returns base unchanged.
func Duration(base time.Duration, spread float64) time.Duration {
	if base <= 0 || spread <= 0 {
		return base
	}
	if spread > 1 {
		spread = 1
	}
	// Sum of two uniforms is a cheap bell shape centred on 0.5.
	bell := (rand.Float64() + rand.Float64()) / 2
	factor := 1 + (bell-0.5)*2*spread
	return time.Duration(float64(base) * factor)
}

// Sleep waits for Duration(base, spread) and reports false when ctx ended first.
func Sleep(ctx context.Context, base time.Duration, spread float64) bool {
	delay := Duration(base, spread)
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}
