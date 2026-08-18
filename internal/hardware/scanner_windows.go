//go:build windows
// +build windows

package hardware

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = modkernel32.NewProc("GlobalMemoryStatusEx")
)

type memoryStatusEx struct {
	cbLength               uint32
	dwMemoryLoad           uint32
	ullTotalPhys           uint64
	ullAvailPhys           uint64
	ullTotalPageFile       uint64
	ullAvailPageFile       uint64
	ullTotalVirtual        uint64
	ullAvailVirtual        uint64
	ullAvailExtendedVirtual uint64
}

func getTotalRAMGB() (int, error) {
	var memInfo memoryStatusEx
	memInfo.cbLength = uint32(unsafe.Sizeof(memInfo))

	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	if ret == 0 {
		return 0, err
	}

	// Convert bytes to GB
	gb := int(memInfo.ullTotalPhys / (1024 * 1024 * 1024))
	return gb, nil
}
