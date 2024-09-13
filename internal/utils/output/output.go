package output

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"
	// "time"

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
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)

	// Print headers
	headers := make([]string, len(opts.Columns))
	for i, col := range opts.Columns {
		headers[i] = color.New(color.Bold).Sprintf(col.Format, col.Header)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	// Print file information
	for _, file := range files {
		if !opts.ShowHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			logger.Warn("failed to get the file info", file.Name(), "err", err)
			continue
		}

		values := make([]string, len(opts.Columns))
		for i, col := range opts.Columns {
			value := col.Value(info)
			if i == len(opts.Columns)-1 && opts.ShouldColor { // Name column
				value = getColoredFileNames(file)
			}
			values[i] = fmt.Sprintf(col.Format, value)
		}

		fmt.Fprintln(tw, strings.Join(values, "\t"))
	}

	tw.Flush()
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
		{Header: "Mode", Width: 10, Format: "%-10s", Value: func(info fs.FileInfo) string { return info.Mode().String() }},
		{Header: "Links", Width: 5, Format: "%5s", Value: func(info fs.FileInfo) string { return "1" }},
		{Header: "Owner", Width: 8, Format: "%-8s", Value: func(info fs.FileInfo) string { return strconv.Itoa(os.Geteuid()) }},
		{Header: "Group", Width: 8, Format: "%-8s", Value: func(info fs.FileInfo) string { return strconv.Itoa(os.Getegid()) }},
		{Header: "Size", Width: 10, Format: "%10s", Value: func(info fs.FileInfo) string { 
			return humanizeSize(info.Size()) 
		}},
		{Header: "Modified", Width: 12, Format: "%-12s", Value: func(info fs.FileInfo) string { 
			return info.ModTime().Format("Jan _2 15:04") 
		}},
		{Header: "Name", Width: 0, Format: "%s", Value: func(info fs.FileInfo) string { return info.Name() }},
	}
}

// humanizeSize converts a file size in bytes to a human-readable string
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