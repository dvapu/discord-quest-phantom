//go:build windows

package identity

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

var (
	modCrypt32             = syscall.NewLazyDLL("crypt32.dll")
	modKernel32            = syscall.NewLazyDLL("kernel32.dll")
	procCryptProtectData   = modCrypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = modCrypt32.NewProc("CryptUnprotectData")
	procLocalFree          = modKernel32.NewProc("LocalFree")
)

// cryptProtectUIForbidden keeps DPAPI from ever showing a prompt in a console
// session; the credential must be readable unattended.
const cryptProtectUIForbidden = 0x1

// Protect seals data with the Windows Data Protection API, bound to the current
// Windows user profile.
func Protect(data []byte) ([]byte, error) {
	return dpapiCall(procCryptProtectData, data, "CryptProtectData")
}

// Unprotect opens data sealed by Protect for the same Windows user.
func Unprotect(data []byte) ([]byte, error) {
	return dpapiCall(procCryptUnprotectData, data, "CryptUnprotectData")
}

// ProtectionLabel describes the at-rest protection for the startup banner.
func ProtectionLabel() string { return "DPAPI, current Windows user" }

// ProtectionStrong reports whether secrets are actually encrypted at rest.
func ProtectionStrong() bool { return true }

func dpapiCall(proc *syscall.LazyProc, data []byte, name string) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("nothing to encrypt")
	}
	in := dataBlob{cbData: uint32(len(data)), pbData: &data[0]}
	var out dataBlob
	r, _, err := proc.Call(
		uintptr(unsafe.Pointer(&in)),
		0, // description
		0, // optional entropy
		0, // reserved
		0, // prompt struct
		cryptProtectUIForbidden,
		uintptr(unsafe.Pointer(&out)),
	)
	if r == 0 {
		return nil, fmt.Errorf("%s failed: %w", name, err)
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	sealed := make([]byte, out.cbData)
	copy(sealed, unsafe.Slice(out.pbData, int(out.cbData)))
	return sealed, nil
}
