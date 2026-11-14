//go:build !windows

package presence

// Snapshot is a no-op on platforms without the Windows presence checks. The
// guards stay disabled instead of blocking the spoofer on a guess.
func Snapshot() State { return State{Supported: false} }

// TryLock has no named-mutex equivalent here; concurrent instances are left to
// the caller's own coordination.
func TryLock(name string) (release func(), ok bool) { return func() {}, true }
