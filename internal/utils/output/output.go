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
	"time"

	"github.com/fatih/color"
	"github.com/The-True-Hooha/NimbleFiles/internal/utils/logger"
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
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	headers := make([]string, len(opts.Columns))
	for i, col := range opts.Columns {
		headers[i] = col.Header
	}
	fmt.Println(tw, strings.Join(headers, "\t"))

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
			values[i] = col.Value(info)
		}

		if opts.ShouldColor {
			values[len(values)-1] = getColoredFileNames(file)
		}

		fmt.Fprintf(tw, strings.Join(values, "\t")+"\n")
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
		name = color.BlueString(name)
	} else if file.Type()&fs.ModeSymlink != 0 {
		name = color.CyanString(name)
		if pathDestination, err := os.Readlink(filepath.Join(".", name)); err == nil {
			return name + " -> " + pathDestination
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
		{Header: "Links", Width: 5, Format: "%5s", Value: func(info fs.FileInfo) string { return "1" }},                         // Placeholder
		{Header: "Owner", Width: 8, Format: "%-8s", Value: func(info fs.FileInfo) string { return strconv.Itoa(os.Geteuid()) }}, // Placeholder
		{Header: "Group", Width: 8, Format: "%-8s", Value: func(info fs.FileInfo) string { return strconv.Itoa(os.Getegid()) }}, // Placeholder
		{Header: "Size", Width: 8, Format: "%8d", Value: func(info fs.FileInfo) string { return strconv.FormatInt(info.Size(), 10) }},
		{Header: "Modified", Width: 20, Format: "%-20s", Value: func(info fs.FileInfo) string { return info.ModTime().Format(time.RFC3339) }},
		{Header: "Name", Width: 0, Format: "%s", Value: func(info fs.FileInfo) string { return info.Name() }},
	}
}
