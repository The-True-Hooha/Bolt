package diskusage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

type Entry struct {
	Path  string
	Size  int64
	IsDir bool
}

// TopLevel returns sizes of immediate children of root, sorted by size desc.
func TopLevel(root string) ([]Entry, error) {
	children, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, c := range children {
		full := filepath.Join(root, c.Name())
		sz, err := dirSize(full)
		if err != nil {
			sz = 0
		}
		entries = append(entries, Entry{Path: full, Size: sz, IsDir: c.IsDir()})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Size > entries[j].Size })
	return entries, nil
}

func dirSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total, err
}

func HandleDuCommand() common.Command {
	flags := pflag.NewFlagSet("du", pflag.ContinueOnError)
	top := flags.IntP("top", "n", 20, "show top N entries")

	return common.Command{
		Name:        "du",
		Description: "show disk usage of directory contents",
		Flags:       flags,
		Execute: func(args []string) error {
			root := "."
			if len(args) > 0 {
				root = args[0]
			}
			entries, err := TopLevel(root)
			if err != nil {
				return fmt.Errorf("cannot read '%s': %w", root, err)
			}
			limit := *top
			if limit > len(entries) {
				limit = len(entries)
			}
			maxSz := int64(1)
			for _, e := range entries[:limit] {
				if e.Size > maxSz {
					maxSz = e.Size
				}
			}
			fmt.Printf("%-10s  %s\n", "SIZE", "PATH")
			for _, e := range entries[:limit] {
				bar := renderBar(e.Size, maxSz, 20)
				fmt.Printf("%-10s  %s  %s\n", humanBytes(e.Size), bar, filepath.Base(e.Path))
			}
			return nil
		},
	}
}

func renderBar(size, max int64, width int) string {
	if max == 0 {
		return ""
	}
	filled := int(int64(width) * size / max)
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func HumanBytes(size int64) string { return humanBytes(size) }

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
