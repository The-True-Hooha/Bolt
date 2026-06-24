package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/The-True-Hooha/Bolt/internal/common"
)

func HandleInfoCommand() common.Command {
	return common.Command{
		Name:        "info",
		Description: "show detailed file or directory information",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: info <path>...")
			}
			for _, path := range args {
				if err := printInfo(path); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func printInfo(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve '%s': %w", path, err)
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return fmt.Errorf("cannot stat '%s': %w", path, err)
	}

	fmt.Printf("\nPath:        %s\n", abs)
	fmt.Printf("Name:        %s\n", info.Name())
	fmt.Printf("Type:        %s\n", fileType(info))
	fmt.Printf("Size:        %s (%d bytes)\n", humanBytes(info.Size()), info.Size())
	fmt.Printf("Permissions: %s  (%04o)\n", info.Mode().String(), info.Mode().Perm())
	fmt.Printf("Modified:    %s\n", info.ModTime().Format(time.RFC1123))
	fmt.Printf("Mode bits:   %s\n", info.Mode().String())

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(abs)
		if err == nil {
			fmt.Printf("Symlink →    %s\n", target)
		}
	}

	if info.IsDir() {
		entries, err := os.ReadDir(abs)
		if err == nil {
			dirs, files := 0, 0
			for _, e := range entries {
				if e.IsDir() {
					dirs++
				} else {
					files++
				}
			}
			fmt.Printf("Contents:    %d directories, %d files\n", dirs, files)
		}
	}

	fmt.Println()
	return nil
}

func fileType(info os.FileInfo) string {
	m := info.Mode()
	switch {
	case m&os.ModeSymlink != 0:
		return "symlink"
	case m&os.ModeDir != 0:
		return "directory"
	case m&os.ModeNamedPipe != 0:
		return "named pipe"
	case m&os.ModeSocket != 0:
		return "socket"
	case m&os.ModeDevice != 0:
		return "device"
	default:
		return "regular file"
	}
}

func humanBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
