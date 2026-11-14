//go:build windows

package spoofer

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var (
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW            = modUser32.NewProc("RegisterClassExW")
	procCreateWindowExW             = modUser32.NewProc("CreateWindowExW")
	procDefWindowProcW              = modUser32.NewProc("DefWindowProcW")
	procShowWindow                  = modUser32.NewProc("ShowWindow")
	procUpdateWindow                = modUser32.NewProc("UpdateWindow")
	procPeekMessageW                = modUser32.NewProc("PeekMessageW")
	procTranslateMessage            = modUser32.NewProc("TranslateMessage")
	procDispatchMessageW            = modUser32.NewProc("DispatchMessageW")
	procDestroyWindow               = modUser32.NewProc("DestroyWindow")
	procMsgWaitForMultipleObjectsEx = modUser32.NewProc("MsgWaitForMultipleObjectsEx")
	procGetModuleHandleW            = modKernel32.NewProc("GetModuleHandleW")
)

type wndClassExW struct {
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

type point struct {
	x, y int32
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

const (
	wsOverlappedWindow = 0x00CF0000
	wsMinimize         = 0x20000000
	swShowMinNoActive  = 7
	pmRemove           = 0x0001
	wmQuit             = 0x0012
	qsAllInput         = 0x04FF
	mwmoInputAvailable = 0x0004
)

// gameWindowClasses are window class names used by real game engines. Discord
// reads window titles today, but a class named after this tool is a permanent
// fingerprint for anything that looks deeper.
var gameWindowClasses = []string{
	"UnityWndClass",
	"UnrealWindow",
	"SDL_app",
	"Valve001",
	"GAMEFACE",
	"Chrome_WidgetWin_0",
}

// windowClassFor picks a stable class per game title.
func windowClassFor(title string) string {
	var hash uint32 = 2166136261
	for _, r := range title {
		hash ^= uint32(r)
		hash *= 16777619
	}
	return gameWindowClasses[int(hash)%len(gameWindowClasses)]
}

// workingSetMB derives a plausible resident size for a modern game, bounded so
// a spoofed quest never costs more than half a gigabyte of real memory.
func workingSetMB(title string) int {
	if raw := os.Getenv("DQC_STUB_MEM_MB"); raw != "" {
		if mb, err := strconv.Atoi(raw); err == nil && mb >= 0 && mb <= 8192 {
			return mb
		}
	}
	var hash uint32 = 2166136261
	for _, r := range title {
		hash ^= uint32(r)
		hash *= 16777619
	}
	return 256 + int(hash%257) // 256..512 MB
}

// stubResident keeps the pages referenced so the Go collector cannot release
// them: Discord's native module can query private working set size.
var stubResident []byte

// reserveWorkingSet touches one byte per page so the committed size is real,
// not just a reservation.
func reserveWorkingSet(mb int) {
	if mb <= 0 {
		return
	}
	stubResident = make([]byte, mb<<20)
	for i := 0; i < len(stubResident); i += 4096 {
		stubResident[i] = byte(i)
	}
}

// startCPUNoise burns a fraction of a percent of one core, the way a game idling
// on a menu does. A process at exactly 0.0% CPU for 15 minutes is a signature.
func startCPUNoise() chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				deadline := time.Now().Add(2 * time.Millisecond)
				var acc float64
				for time.Now().Before(deadline) {
					acc += 1.000001
				}
				if acc < 0 {
					fmt.Print("")
				}
			}
		}
	}()
	return stop
}

// RunStub creates a visible, minimized Win32 GUI window with the specified title
// and executes a message pump dispatching messages until an OS interrupt or termination signal.
// This enables Discord Desktop's native module (EnumWindows and GetWindowTextW) to detect
// the game window and initiate quest progression without user-visible focus disruption.
func RunStub(title string) error {
	if title == "" {
		title = "Discord Quest Stub Window"
	}
	fmt.Printf("[Dummy Game Process] Running Win32 GUI stub for: %s (PID: %d)\n", title, os.Getpid())

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := windowClassFor(title)

	clsNamePtr, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return fmt.Errorf("failed encoding window class name: %w", err)
	}

	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return fmt.Errorf("failed encoding window title: %w", err)
	}

	var wc wndClassExW
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	wc.style = 0
	wc.lpfnWndProc = procDefWindowProcW.Addr()
	wc.hInstance = hInstance
	wc.lpszClassName = clsNamePtr

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(clsNamePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		wsOverlappedWindow|wsMinimize,
		0x80000000, 0x80000000, // CW_USEDEFAULT
		640, 480,
		0, 0,
		hInstance,
		0,
	)

	if hwnd == 0 {
		return fmt.Errorf("CreateWindowExW failed for title %q: %w", title, err)
	}
	defer procDestroyWindow.Call(hwnd)

	// Display window in minimized state without stealing keyboard or focus
	procShowWindow.Call(hwnd, swShowMinNoActive)
	procUpdateWindow.Call(hwnd)

	reserveWorkingSet(workingSetMB(title))
	cpuStop := startCPUNoise()
	defer close(cpuStop)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var m msg
	for {
		// Drain message queue
		for {
			r, _, _ := procPeekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
			if r == 0 {
				break
			}
			if m.message == wmQuit {
				return nil
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}

		select {
		case sig := <-sigChan:
			fmt.Printf("[Dummy Game Process] Received %v, terminating window...\n", sig)
			return nil
		default:
		}

		// Wait up to 100ms for incoming messages or timeouts, maintaining 0% CPU consumption
		procMsgWaitForMultipleObjectsEx.Call(0, 0, 100, qsAllInput, mwmoInputAvailable)
	}
}
