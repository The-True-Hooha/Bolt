package grep

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

var (
	matchColor    = color.New(color.FgRed, color.Bold)
	fileColor     = color.New(color.FgMagenta, color.Bold)
	lineNumColor  = color.New(color.FgCyan)
)

func HandleGrepCommand() common.Command {
	flags := pflag.NewFlagSet("grep", pflag.ContinueOnError)
	recursive := flags.BoolP("recursive", "r", false, "search recursively in directories")
	ignoreCase := flags.BoolP("ignore-case", "i", false, "case-insensitive matching")
	lineNumbers := flags.BoolP("line-number", "n", true, "show line numbers")
	invertMatch := flags.BoolP("invert-match", "v", false, "select non-matching lines")
	countOnly := flags.BoolP("count", "c", false, "print only count of matching lines")

	return common.Command{
		Name:        "grep",
		Description: "search for a pattern within files",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: grep [options] <pattern> [file...]")
			}

			pattern := args[0]
			if *ignoreCase {
				pattern = "(?i)" + pattern
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern '%s': %w", args[0], err)
			}

			targets := args[1:]
			if len(targets) == 0 {
				targets = []string{"."}
			}

			multiFile := len(targets) > 1
			for _, target := range targets {
				info, err := os.Stat(target)
				if err != nil {
					return fmt.Errorf("cannot access '%s': %w", target, err)
				}

				if info.IsDir() {
					if !*recursive {
						return fmt.Errorf("'%s' is a directory (use -r to search recursively)", target)
					}
					if err := walkAndGrep(target, re, *lineNumbers, *invertMatch, *countOnly); err != nil {
						return err
					}
				} else {
					if err := grepFile(target, re, *lineNumbers, *invertMatch, *countOnly, multiFile); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

func walkAndGrep(root string, re *regexp.Regexp, lineNumbers, invertMatch, countOnly bool) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot access '%s': %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		return grepFile(path, re, lineNumbers, invertMatch, countOnly, true)
	})
}

func grepFile(path string, re *regexp.Regexp, showLineNums, invertMatch, countOnly, showFilename bool) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open '%s': %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	matchCount := 0

	var output strings.Builder

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matched := re.MatchString(line)

		if invertMatch {
			matched = !matched
		}

		if !matched {
			continue
		}

		matchCount++
		if countOnly {
			continue
		}

		var sb strings.Builder
		if showFilename {
			sb.WriteString(fileColor.Sprint(path))
			sb.WriteString(":")
		}
		if showLineNums {
			sb.WriteString(lineNumColor.Sprintf("%d", lineNum))
			sb.WriteString(":")
		}
		sb.WriteString(highlightMatches(line, re))
		output.WriteString(sb.String())
		output.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading '%s': %w", path, err)
	}

	if countOnly {
		if showFilename {
			fmt.Printf("%s:%d\n", fileColor.Sprint(path), matchCount)
		} else {
			fmt.Println(matchCount)
		}
		return nil
	}

	if output.Len() > 0 {
		fmt.Print(output.String())
	}
	return nil
}

func highlightMatches(line string, re *regexp.Regexp) string {
	return re.ReplaceAllStringFunc(line, func(m string) string {
		return matchColor.Sprint(m)
	})
}
