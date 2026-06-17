//go:build windows

package monitor

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW     = user32.NewProc("GetMonitorInfoW")
)

const monitorInfoFPrimary = 0x00000001

// matches MONITORINFO from winuser.h
type monitorInfo struct {
	cbSize    uint32
	rcMonitor rect32
	rcWork    rect32
	dwFlags   uint32
}

type rect32 struct {
	Left, Top, Right, Bottom int32
}

// primaryResult is passed by pointer through the EnumDisplayMonitors callback.
// EnumDisplayMonitors is synchronous so the GC will not move it during the call.
type primaryResult struct {
	width, height int
	found         bool
}

// PrimaryMonitor returns the pixel dimensions of the Windows primary display.
func PrimaryMonitor() (width, height int, err error) {
	res := &primaryResult{}

	cb := windows.NewCallback(func(hMonitor, _, _, data uintptr) uintptr {
		r := (*primaryResult)(unsafe.Pointer(data))
		var mi monitorInfo
		mi.cbSize = uint32(unsafe.Sizeof(mi))
		procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
		if mi.dwFlags&monitorInfoFPrimary != 0 {
			r.width = int(mi.rcMonitor.Right - mi.rcMonitor.Left)
			r.height = int(mi.rcMonitor.Bottom - mi.rcMonitor.Top)
			r.found = true
		}
		return 1 // continue enumeration
	})

	procEnumDisplayMonitors.Call(0, 0, cb, uintptr(unsafe.Pointer(res)))

	if !res.found {
		return 0, 0, fmt.Errorf("primary monitor not found")
	}
	return res.width, res.height, nil
}
