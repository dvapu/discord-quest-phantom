//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLocaleName = kernel32.NewProc("GetUserDefaultLocaleName")
	procSetConsoleOutputCP       = kernel32.NewProc("SetConsoleOutputCP")
)

func init() {
	// Enable UTF-8 code page (65001) for Windows console so emojis and unicode display correctly
	_, _, _ = procSetConsoleOutputCP.Call(uintptr(65001))
}

func detectOSLocale() string {
	buf := make([]uint16, 85)
	r, _, _ := procGetUserDefaultLocaleName.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
