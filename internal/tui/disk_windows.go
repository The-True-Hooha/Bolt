//go:build windows

package tui

import (
	"path/filepath"
	"syscall"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"
)

func loadDiskUsage(path string) tea.Cmd {
	return func() tea.Msg {
		abs, err := filepath.Abs(path)
		if err != nil {
			return diskUsageMsg{}
		}

		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		getDiskFreeEx := kernel32.NewProc("GetDiskFreeSpaceExW")

		absPtr, err := syscall.UTF16PtrFromString(abs)
		if err != nil {
			return diskUsageMsg{}
		}

		var freeBytesAvailable, totalBytes, totalFreeBytes uint64
		ret, _, _ := getDiskFreeEx.Call(
			uintptr(unsafe.Pointer(absPtr)),
			uintptr(unsafe.Pointer(&freeBytesAvailable)),
			uintptr(unsafe.Pointer(&totalBytes)),
			uintptr(unsafe.Pointer(&totalFreeBytes)),
		)
		if ret == 0 {
			return diskUsageMsg{}
		}
		return diskUsageMsg{free: freeBytesAvailable, total: totalBytes}
	}
}
