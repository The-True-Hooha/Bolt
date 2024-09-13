package output

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/utils/logger"
	"github.com/fatih/color"
)

type ColumnStructure struct {
	Header string
	Width  int
	Format string
	Value  func(fs.FileInfo) string
}

type PrintOptions struct {
	LongFormat  bool
	ShowHidden  bool
	ShouldColor bool
	Columns     []ColumnStructure
}

func PrintFileInfo(w io.Writer, files []fs.DirEntry, opts PrintOptions) {
	if opts.LongFormat {
		printLongFormat(w, files, opts)
	} else {
		printShortFormat(w, files, opts)
	}
}

func printLongFormat(w io.Writer, files []fs.DirEntry, opts PrintOptions) {

	var lines []string

	headerParts := make([]string, len(opts.Columns))
	for i, col := range opts.Columns {
		if col.Width > 0 {
			headerParts[i] = fmt.Sprintf(col.Format, col.Header)
		} else {
			headerParts[i] = col.Header
		}
	}
	lines = append(lines, strings.Join(headerParts, " "))

	for _, file := range files {
		if !opts.ShowHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}
		info, err := file.Info()
		if err != nil {
			logger.Warn("failed to get the file info", file.Name(), "err", err)
			continue
		}

		parts := make([]string, len(opts.Columns))
		for i, col := range opts.Columns {
			value := col.Value(info)
			if col.Width > 0 {
				parts[i] = fmt.Sprintf(col.Format, value)
			} else {
				parts[i] = value
			}
		}

		if opts.ShouldColor {
			parts[len(parts)-1] = getColoredFileNames(file)
		}

		lines = append(lines, strings.Join(parts, " "))
	}
	fmt.Fprintln(w, strings.Join(lines, "\n"))
}

func printShortFormat(w io.Writer, files []fs.DirEntry, opts PrintOptions) {
	for _, file := range files {
		if !opts.ShowHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}
		name := file.Name()
		if opts.ShouldColor {
			name = getColoredFileNames(file)
		}
		fmt.Fprintln(w, name)
	}
}

func getColoredFileNames(file fs.DirEntry) string {
	name := file.Name()
	info, err := file.Info()
	if err != nil {
		return name
	}
	if file.IsDir() {
		return color.BlueString(name)
	} else if file.Type()&fs.ModeSymlink != 0 {
		name = color.CyanString(name)
		if pathDestination, err := os.Readlink(filepath.Join(".", name)); err == nil {
			return fmt.Sprintf("%s -> %s", name, color.MagentaString(pathDestination))
		}
		return name
	} else if fileIsExecutable(info) {
		return color.GreenString(name)
	}
	return name
}

func fileIsExecutable(info fs.FileInfo) bool {
	if runtime.GOOS == "windows" {
		return filepath.Ext(info.Name()) == ".exe" || filepath.Ext(info.Name()) == ".bat"
	}
	return info.Mode()&0111 != 0
}

func GetDefaultColumns() []ColumnStructure {
	return []ColumnStructure{
		{Header: "Permissions", Width: 11, Format: "%-11s", Value: func(info fs.FileInfo) string { return info.Mode().String() }},
		{Header: "Owner", Width: 8, Format: "%-8s", Value: getOwnerName},
		{Header: "Group", Width: 8, Format: "%-8s", Value: getGroupName},
		{Header: "Size", Width: 9, Format: "%9s", Value: func(info fs.FileInfo) string { return humanizeSize(info.Size()) }},
		{Header: "Modified", Width: 20, Format: "%-20s", Value: func(info fs.FileInfo) string { 
			return info.ModTime().Format("2006-01-02 15:04:05")
		}},
		{Header: "Name", Width: 0, Format: "%s", Value: func(info fs.FileInfo) string { return info.Name() }},
	}
}

func getOwnerName(info fs.FileInfo) string {
	if runtime.GOOS == "windows" {
		return "user"
	}
	uid := info.Sys().(interface{ Uid() uint32 }).Uid()
	if u, err := user.LookupId(fmt.Sprint(uid)); err == nil {
		return u.Username
	}
	return fmt.Sprint(uid)
}

func getGroupName(info fs.FileInfo) string {
	if runtime.GOOS == "windows" {
		return "user"
	}
	gid := info.Sys().(interface{ Gid() uint32 }).Gid()
	if g, err := user.LookupGroupId(fmt.Sprint(gid)); err == nil {
		return g.Name
	}
	return fmt.Sprint(gid)
}

func humanizeSize(size int64) string {
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