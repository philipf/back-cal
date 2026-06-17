//go:build windows

package monitor

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfo      = user32.NewProc("GetMonitorInfoW")
)

const monitorInfoFPrimary = 0x00000001

type monitorInfo struct {
	cbSize    uint32
	rcMonitor rect
	rcWork    rect
	dwFlags   uint32
}

type rect struct {
	left, top, right, bottom int32
}

type primaryResult struct {
	width  int
	height int
	found  bool
}

func enumProc(hMonitor, _ uintptr, _ *rect, data uintptr) uintptr {
	result := (*primaryResult)(unsafe.Pointer(data))
	var mi monitorInfo
	mi.cbSize = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfo.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))
	if mi.dwFlags&monitorInfoFPrimary != 0 {
		result.width = int(mi.rcMonitor.right - mi.rcMonitor.left)
		result.height = int(mi.rcMonitor.bottom - mi.rcMonitor.top)
		result.found = true
	}
	return 1
}

func PrimaryMonitor() (width, height int, err error) {
	result := &primaryResult{}
	cb := windows.NewCallback(func(hMonitor, hdc uintptr, lpRect *rect, data uintptr) uintptr {
		return enumProc(hMonitor, hdc, lpRect, data)
	})
	procEnumDisplayMonitors.Call(0, 0, cb, uintptr(unsafe.Pointer(result)))
	if !result.found {
		return 0, 0, fmt.Errorf("primary monitor not found")
	}
	return result.width, result.height, nil
}
