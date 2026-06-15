//go:build !windows

package tui

import (
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func loadDiskUsage(path string) tea.Cmd {
	return func() tea.Msg {
		abs, err := filepath.Abs(path)
		if err != nil {
			return diskUsageMsg{}
		}
		var stat syscall.Statfs_t
		if err := syscall.Statfs(abs, &stat); err != nil {
			return diskUsageMsg{}
		}
		return diskUsageMsg{
			free:  stat.Bavail * uint64(stat.Bsize),
			total: stat.Blocks * uint64(stat.Bsize),
		}
	}
}
