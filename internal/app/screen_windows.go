//go:build windows

package app

import (
	"syscall"
	"unsafe"
)

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

var (
	user32                    = syscall.NewLazyDLL("user32.dll")
	procSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	procGetDpiForSystem       = user32.NewProc("GetDpiForSystem")
)

const spiGetWorkArea = 0x0030

// GetAvailableWorkArea returns the primary monitor's usable work area dimensions
// (excluding taskbars and docked bars) converted to logical pixels.
func GetAvailableWorkArea() (int, int, error) {
	var dpi uint32 = 96
	if err := procGetDpiForSystem.Find(); err == nil {
		r, _, _ := procGetDpiForSystem.Call()
		if r > 0 {
			dpi = uint32(r)
		}
	}

	var rc rect
	ret, _, err := procSystemParametersInfoW.Call(
		uintptr(spiGetWorkArea),
		0,
		uintptr(unsafe.Pointer(&rc)),
		0,
	)
	if ret == 0 {
		return 0, 0, err
	}

	physWidth := int(rc.Right - rc.Left)
	physHeight := int(rc.Bottom - rc.Top)
	if dpi == 0 {
		dpi = 96
	}

	logicalWidth := (physWidth * 96) / int(dpi)
	logicalHeight := (physHeight * 96) / int(dpi)

	return logicalWidth, logicalHeight, nil
}

