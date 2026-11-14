//go:build !windows

package identity

// Protect is a no-op on platforms without a per-user secret store. Callers use
// ProtectionStrong to warn that the file is only protected by its permissions.
func Protect(data []byte) ([]byte, error) {
	return append([]byte(nil), data...), nil
}

// Unprotect reverses Protect on platforms without a secret store.
func Unprotect(data []byte) ([]byte, error) {
	return append([]byte(nil), data...), nil
}

// ProtectionLabel describes the at-rest protection for the startup banner.
func ProtectionLabel() string { return "file permissions (0600)" }

// ProtectionStrong reports whether secrets are actually encrypted at rest.
func ProtectionStrong() bool { return false }
