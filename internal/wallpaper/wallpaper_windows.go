//go:build windows

package wallpaper

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"unsafe"
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	procSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
)

const (
	spiSetDeskWallpaper = 0x0014
	spifUpdateIniFile   = 0x0001
	spifSendChange      = 0x0002
)

func Set(path string) error {
	// Set wallpaper style to "Fit" (no scaling) since canvas matches screen resolution
	k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Desktop`, registry.SET_VALUE)
	if err == nil {
		k.SetStringValue("WallpaperStyle", "6")
		k.SetStringValue("TileWallpaper", "0")
		k.Close()
	}

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	ret, _, err := procSystemParametersInfoW.Call(
		spiSetDeskWallpaper,
		0,
		uintptr(unsafe.Pointer(pathPtr)),
		spifUpdateIniFile|spifSendChange,
	)
	if ret == 0 {
		return err
	}
	return nil
}
