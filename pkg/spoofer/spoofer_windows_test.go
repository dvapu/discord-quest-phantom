//go:build windows

package spoofer

import (
	"fmt"
	"syscall"
	"testing"
	"unsafe"
)

var (
	user32Test   = syscall.NewLazyDLL("user32.dll")
	kernel32Test = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExWTest = user32Test.NewProc("RegisterClassExW")
	procCreateWindowExWTest  = user32Test.NewProc("CreateWindowExW")
	procDefWindowProcWTest   = user32Test.NewProc("DefWindowProcW")
	procShowWindowTest       = user32Test.NewProc("ShowWindow")
	procDestroyWindowTest    = user32Test.NewProc("DestroyWindow")
	procGetModuleHandleWTest = kernel32Test.NewProc("GetModuleHandleW")
	procGetWindowTextWTest   = user32Test.NewProc("GetWindowTextW")
	procIsWindowVisibleTest  = user32Test.NewProc("IsWindowVisible")
)

type wndClassExWTest struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

func createTestWindow(title string) (uintptr, error) {
	hInstance, _, _ := procGetModuleHandleWTest.Call(0)
	className := "TestStubWndClass2"

	clsNamePtr, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return 0, err
	}

	var wc wndClassExWTest
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.lpfnWndProc = procDefWindowProcWTest.Addr()
	wc.hInstance = hInstance
	wc.lpszClassName = clsNamePtr

	procRegisterClassExWTest.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, err := procCreateWindowExWTest.Call(
		0,
		uintptr(unsafe.Pointer(clsNamePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		wsOverlappedWindow|wsMinimize,
		0x80000000, 0x80000000,
		400, 300,
		0, 0,
		hInstance,
		0,
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowExW failed: %w", err)
	}

	procShowWindowTest.Call(hwnd, swShowMinNoActive)
	return hwnd, nil
}

func TestWin32StubWindowCreation(t *testing.T) {
	expectedTitle := "HELLDIVERS 2"
	hwnd, err := createTestWindow(expectedTitle)
	if err != nil {
		t.Fatalf("Failed to create test window: %v", err)
	}
	defer procDestroyWindowTest.Call(hwnd)

	// 1. Verify window visibility for EnumWindows
	vis, _, _ := procIsWindowVisibleTest.Call(hwnd)
	if vis == 0 {
		t.Errorf("Expected window to be visible, got not visible")
	}

	// 2. Verify window text
	buf := make([]uint16, 256)
	procGetWindowTextWTest.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	actualTitle := syscall.UTF16ToString(buf)
	if actualTitle != expectedTitle {
		t.Errorf("Window title mismatch: got %q, expected %q", actualTitle, expectedTitle)
	}

	t.Logf("Win32 window successfully validated: HWND=0x%X, Visible=%v, Title=%q", hwnd, vis != 0, actualTitle)
}
