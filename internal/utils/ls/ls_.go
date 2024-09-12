package lscmd

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/The-True-Hooha/NimbleFiles/internal/common"
	"github.com/The-True-Hooha/NimbleFiles/internal/utils/logger"
	"github.com/The-True-Hooha/NimbleFiles/internal/utils/output"
	"github.com/spf13/pflag"
)

type LsOptions struct {
	longFormat bool
	showHidden bool
	sortBy     string
	reverse    bool
	tags       string
}

func HandleLsCommandTags() common.Command {
	opts := &LsOptions{}

	// DoSomething()

	flags := pflag.NewFlagSet("ls", pflag.ContinueOnError)
	flags.BoolVarP(&opts.longFormat, "long", "l", false, "uses the long listing format")
	flags.BoolVarP(&opts.longFormat, "all", "a", false, "show hidden files")
	flags.BoolVarP(&opts.longFormat, "reverse", "r", false, "reverses the order of files")
	flags.StringVarP(&opts.sortBy, "sort", "s", "name", "sorts by: name, size, createdDate")
	flags.StringVarP(&opts.tags, "tag", "t", "", "filter files by tags")

	return common.Command{
		Name:        "ls",
		Description: "list the directory contents",
		Flags:       flags,
		Execute:     executeLsCommand(opts),
	}
}

func executeLsCommand(opts *LsOptions) func(args []string) error {
	return func(args []string) error {
		// current_working_dir, err := os.Getwd()
		currentPath := "."
		if len(args) > 0 {
			currentPath = args[0]
		}

		logger.Info("listing directory contents", "dir", currentPath)

		files, err := sortDirectory(currentPath, opts)
		if err != nil {
			logger.Error("failed to read the directory", "dir", currentPath, "error", err)
			return fmt.Errorf("failed to read the directory %s: %w", currentPath, err)
		}

		if opts.tags != "" {
			files, err := filterFilesByTags(files, opts.tags)
			if err != nil{
				logger.Debug("failed to get file tags for some unknown reason %v\n: %w\n", len(files), err)
				return err
			}
		}

		printOpts := output.PrintOptions{
			LongFormat:  opts.longFormat,
			ShowHidden:  opts.showHidden,
			ShouldColor: true,
			Columns:     output.GetDefaultColumns(),
		}
		output.PrintFileInfo(os.Stdout, files, printOpts)
		logger.Info("successfully listed directory contents", "dir", currentPath, "fileCount", len(files))
		return nil
	}

}

func sortDirectory(path string, opts *LsOptions) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []fs.DirEntry

	for _, entry := range entries {
		if !opts.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		files = append(files, entry)

	}

	sort.Slice(files, func(first, second int) bool {
		less := false

		switch opts.sortBy {
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

		if opts.reverse {
			return !less
		}
		return less
	})

	return files, nil

}

func filterFilesByTags(files []fs.DirEntry, tag string) ([]fs.DirEntry, error) {
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
