package lscmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/The-True-Hooha/Bolt/internal/utils/logger"
	"github.com/The-True-Hooha/Bolt/internal/utils/output"
	"github.com/spf13/pflag"
)

const (
	Sort_Name        = "name"
	Sort_size        = "size"
	Sort_CreatedDate = "createdDate"
)

type LsOptions struct {
	LongFormat bool
	ShowHidden bool
	SortBy     string
	Reverse    bool
	Filter     string
}

func HandleLsCommandTags() common.Command {
	opts := &LsOptions{}

	flags := pflag.NewFlagSet("ls", pflag.ContinueOnError)
	flags.BoolVarP(&opts.LongFormat, "long", "l", false, "uses the long listing format")
	flags.BoolVarP(&opts.LongFormat, "all", "a", false, "show hidden files")
	flags.BoolVarP(&opts.LongFormat, "reverse", "r", false, "reverses the order of files")
	flags.StringVarP(&opts.SortBy, "sort", "s", "name", "sorts by: name, size, createdDate")
	flags.StringVarP(&opts.Filter, "tag", "t", "", "filter files by tags or extension")

	return common.Command{
		Name:        "ls",
		Description: "list the directory contents",
		Flags:       flags,
		Execute:     executeLsCommand(opts),
	}
}

func executeLsCommand(opts *LsOptions) func(args []string) error {
	return func(args []string) error {

		logger.Debug("LS command executed with options",
			"LongFormat", opts.LongFormat,
			"ShowHidden", opts.ShowHidden,
			"Reverse", opts.Reverse,
			"SortBy", opts.SortBy,
			"Filter", opts.Filter)
		logger.Debug("Arguments passed", "args", args)

		currentPath := "."
		if len(args) > 0 {
			currentPath = args[0]
		}

		fmt.Println(currentPath, "the second part")

		logger.Info("listing directory contents", "dir", currentPath)

		files, err := getFilteredFiles(currentPath, opts)
		if err != nil {
			logger.Error("failed to read the directory", "dir", currentPath, "error", err)
			return fmt.Errorf("failed to read the directory %s: %w", currentPath, err)
		}

		printOpts := output.PrintOptions{
			LongFormat:  opts.LongFormat,
			ShowHidden:  opts.ShowHidden,
			ShouldColor: true,
			Columns:     output.GetDefaultColumns(),
		}
		output.PrintFileInfo(os.Stdout, files, printOpts)
		// logger.Info("successfully listed directory contents", "dir", currentPath, "fileCount", len(files))
		return nil
	}

}

func SortDirectory_(path string, opts *LsOptions) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []fs.DirEntry

	for _, entry := range entries {
		if !opts.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		files = append(files, entry)

	}

	sort.Slice(files, func(first, second int) bool {
		less := false

		switch opts.SortBy {
		case "size":
			firstInfo, _ := files[first].Info()
			secondInfo, _ := files[second].Info()
			less = firstInfo.Size() < secondInfo.Size()
		case "createdDate":
			firstInfo, _ := files[first].Info()
			secondInfo, _ := files[second].Info()
			less = firstInfo.ModTime().Before(secondInfo.ModTime())
		default:
			less = files[first].Name() < files[second].Name()
		}

		if opts.Reverse {
			return !less
		}
		return less
	})

	return files, nil

}

func FilterFilesByTags_(files []fs.DirEntry, tag string) ([]fs.DirEntry, error) {
	var filtered []fs.DirEntry
	for _, file := range files {
		tags, err := GetFileTags(file.Name())
		if err != nil {
			logger.Debug("failed to get tag for this file %s: %w", file.Name(), err)
			return nil, err
		}
		for _, fileTag := range tags {
			if fileTag == tag {
				filtered = append(filtered, file)
				break
			}
		}
		fmt.Println(file)
	}
	return filtered, nil
}

func getFilteredFiles(path string, opts *LsOptions) ([]fs.DirEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var filteredFiles []fs.DirEntry
	for _, file := range files {
		if !opts.ShowHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if opts.Filter != "" {
			if strings.HasPrefix(opts.Filter, ".") {
				if filepath.Ext(file.Name()) != opts.Filter {
					continue
				}
			} else {
				tags, err := GetFileTags(filepath.Join(path, file.Name()))
				if err != nil {
					logger.Warn("failed to get file tags", "file", file.Name(), "error", err)
					continue
				}

				if !doesFileContainsExt(tags, opts.Filter) {
					continue
				}
			}
		}
		filteredFiles = append(filteredFiles, file)
	}

	if opts.SortBy != "" {
		sort.Slice(filteredFiles, func(i, j int) bool {
			less := false
			var iData fs.FileInfo
			var jData fs.FileInfo

			switch opts.SortBy {
			case Sort_size:
				iData, _ = filteredFiles[i].Info()
				jData, _ = filteredFiles[j].Info()
				less = iData.Size() < jData.Size()
			case Sort_CreatedDate:
				iData, _ = filteredFiles[i].Info()
				jData, _ = filteredFiles[j].Info()
				less = iData.ModTime().Before(jData.ModTime())
			default:
				less = filteredFiles[i].Name() < filteredFiles[j].Name()
			}

			if opts.Reverse {
				return !less
			}
			return less
		})
	}

	return filteredFiles, nil

}

func doesFileContainsExt(ext []string, item string) bool {
	for _, word := range ext {
		if word == item {
			return true
		}
	}
	return false
}
