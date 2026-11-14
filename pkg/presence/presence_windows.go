//go:build windows

package presence

import (
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modUser32   = syscall.NewLazyDLL("user32.dll")
	modShell32  = syscall.NewLazyDLL("shell32.dll")

	procCreateToolhelp32Snapshot   = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW            = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW             = modKernel32.NewProc("Process32NextW")
	procOpenProcess                = modKernel32.NewProc("OpenProcess")
	procQueryFullProcessImageNameW = modKernel32.NewProc("QueryFullProcessImageNameW")
	procCloseHandle                = modKernel32.NewProc("CloseHandle")
	procCreateMutexW               = modKernel32.NewProc("CreateMutexW")
	procWaitForSingleObject        = modKernel32.NewProc("WaitForSingleObject")
	procReleaseMutex               = modKernel32.NewProc("ReleaseMutex")

	procGetForegroundWindow      = modUser32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessID = modUser32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextW           = modUser32.NewProc("GetWindowTextW")
	procGetWindowLongW           = modUser32.NewProc("GetWindowLongW")
	procGetWindowRect            = modUser32.NewProc("GetWindowRect")
	procGetSystemMetrics         = modUser32.NewProc("GetSystemMetrics")
	procMonitorFromWindow        = modUser32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW          = modUser32.NewProc("GetMonitorInfoW")

	procSHQueryUserNotificationState = modShell32.NewProc("SHQueryUserNotificationState")
)

const (
	th32csSnapProcess              = 0x00000002
	invalidHandleValue             = ^uintptr(0)
	processQueryLimitedInformation = 0x1000

	gwlStyle   = ^uintptr(15) // -16
	wsCaption  = 0x00C00000
	smCxScreen = 0
	smCyScreen = 1

	waitObject0   = 0x00000000
	waitAbandoned = 0x00000080
	waitTimeout   = 0x00000102

	monitorDefaultToNearest = 2
)

type processEntry32W struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	CntThreads      uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

type rect struct {
	Left, Top, Right, Bottom int32
}

// Snapshot reads processes, the foreground window and full-screen state.
func Snapshot() State {
	state := State{Supported: true}
	state.Processes = listProcesses()
	for _, p := range state.Processes {
		if isDiscordExe(p.Exe) {
			state.Discord = true
			break
		}
	}
	state.Foreground = foregroundWindow()
	state.Fullscreen = fullscreenActive() || foregroundCoversScreen()
	return state
}

func listProcesses() []Process {
	snap, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snap == 0 || snap == invalidHandleValue {
		return nil
	}
	defer procCloseHandle.Call(snap)

	var entry processEntry32W
	entry.Size = uint32(unsafe.Sizeof(entry))
	out := make([]Process, 0, 256)

	ok, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		exe := strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:]))
		if exe != "" {
			out = append(out, Process{PID: int(entry.ProcessID), Exe: exe})
		}
		ok, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}
	return out
}

func processPath(pid uint32) string {
	handle, _, _ := procOpenProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if handle == 0 {
		return ""
	}
	defer procCloseHandle.Call(handle)

	buf := make([]uint16, syscall.MAX_PATH)
	size := uint32(len(buf))
	ok, _, _ := procQueryFullProcessImageNameW.Call(handle, 0,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:size])
}

func foregroundWindow() Window {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return Window{}
	}
	var pid uint32
	procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	var title [512]uint16
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&title[0])), uintptr(len(title)))

	win := Window{PID: int(pid), Title: syscall.UTF16ToString(title[:n])}
	if path := processPath(pid); path != "" {
		win.Exe = strings.ToLower(filepath.Base(path))
	}
	return win
}

// notificationState maps SHQueryUserNotificationState to a boolean. Values 2-4
// mean a full-screen app, a D3D exclusive game or presentation mode is active.
func fullscreenActive() bool {
	var state int32
	hr, _, _ := procSHQueryUserNotificationState.Call(uintptr(unsafe.Pointer(&state)))
	if int32(hr) != 0 {
		return false
	}
	return state == 2 || state == 3 || state == 4
}

// foregroundCoversScreen catches borderless full-screen games, which report
// QUNS_ACCEPTS_NOTIFICATIONS because Windows does not consider them exclusive.
func foregroundCoversScreen() bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}
	style, _, _ := procGetWindowLongW.Call(hwnd, gwlStyle)
	if style&wsCaption != 0 {
		return false
	}

	var r rect
	if ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ok == 0 {
		return false
	}
	// Measure against the monitor the window actually lives on: SM_CXSCREEN only
	// describes the primary display, so a borderless game on a second screen
	// would be missed.
	bounds, ok := windowMonitorBounds(hwnd)
	if !ok {
		cx, _, _ := procGetSystemMetrics.Call(smCxScreen)
		cy, _, _ := procGetSystemMetrics.Call(smCyScreen)
		bounds = rect{Right: int32(cx), Bottom: int32(cy)}
	}
	width := int(bounds.Right - bounds.Left)
	height := int(bounds.Bottom - bounds.Top)
	if width <= 0 || height <= 0 {
		return false
	}
	return int(r.Right-r.Left) >= width-2 && int(r.Bottom-r.Top) >= height-2
}

type monitorInfoW struct {
	cbSize    uint32
	rcMonitor rect
	rcWork    rect
	dwFlags   uint32
}

// windowMonitorBounds returns the bounds of the monitor containing the window.
func windowMonitorBounds(hwnd uintptr) (rect, bool) {
	monitor, _, _ := procMonitorFromWindow.Call(hwnd, monitorDefaultToNearest)
	if monitor == 0 {
		return rect{}, false
	}
	var info monitorInfoW
	info.cbSize = uint32(unsafe.Sizeof(info))
	if ok, _, _ := procGetMonitorInfoW.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
		return rect{}, false
	}
	return info.rcMonitor, true
}

func isDiscordExe(exe string) bool {
	switch exe {
	case "discord.exe", "discordptb.exe", "discordcanary.exe":
		return true
	}
	return false
}

// TryLock takes a per-session named mutex so two tool instances on the same
// machine can never show two different games at once. The mutex is released by
// the OS when the process exits, so a crash cannot leave it stuck.
func TryLock(name string) (release func(), ok bool) {
	ptr, err := syscall.UTF16PtrFromString(`Local\` + name)
	if err != nil {
		return func() {}, true
	}
	handle, _, _ := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(ptr)))
	if handle == 0 {
		return func() {}, true
	}
	result, _, _ := procWaitForSingleObject.Call(handle, 0)
	switch result {
	case waitTimeout:
		// Another instance owns the slot.
		procCloseHandle.Call(handle)
		return func() {}, false
	case waitObject0, waitAbandoned:
		// Ours. WAIT_ABANDONED means the previous owner died without releasing;
		// Windows already reset ownership, so the mutex is usable.
	default:
		// WAIT_FAILED: ownership is unknown. Fail open so an API error can never
		// wedge the tool, but never release a mutex we do not own.
		procCloseHandle.Call(handle)
		return func() {}, true
	}

	var once sync.Once
	release = func() {
		once.Do(func() {
			procReleaseMutex.Call(handle)
			procCloseHandle.Call(handle)
		})
	}
	return release, true
}
