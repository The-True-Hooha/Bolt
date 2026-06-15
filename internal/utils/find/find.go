package find

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

func HandleFindCommand() common.Command {
	flags := pflag.NewFlagSet("find", pflag.ContinueOnError)
	name := flags.StringP("name", "n", "", "match files by name pattern (e.g. *.go)")
	fileType := flags.StringP("type", "t", "", "filter by type: f=file, d=directory")
	maxDepth := flags.IntP("depth", "d", -1, "maximum directory depth (-1 = unlimited)")

	return common.Command{
		Name:        "find",
		Description: "search for files in a directory hierarchy",
		Flags:       flags,
		Execute: func(args []string) error {
			root := "."
			if len(args) > 0 {
				root = args[0]
			}

			if _, err := os.Stat(root); err != nil {
				return fmt.Errorf("cannot access '%s': %w", root, err)
			}

			rootDepth := strings.Count(filepath.Clean(root), string(os.PathSeparator))

			return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: cannot access '%s': %v\n", path, err)
					return nil
				}

				if path == root {
					return nil
				}

				if *maxDepth >= 0 {
					depth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - rootDepth
					if depth > *maxDepth {
						if d.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}
				}

				if *fileType != "" {
					switch *fileType {
					case "f":
						if d.IsDir() {
							return nil
						}
					case "d":
						if !d.IsDir() {
							return nil
						}
					default:
						return fmt.Errorf("invalid type '%s': use f or d", *fileType)
					}
				}

				if *name != "" {
					matched, err := filepath.Match(*name, d.Name())
					if err != nil {
						return fmt.Errorf("invalid pattern '%s': %w", *name, err)
					}
					if !matched {
						return nil
					}
				}

				if d.IsDir() {
					fmt.Println(color.BlueString(path))
				} else {
					fmt.Println(path)
				}
				return nil
			})
		},
	}
}
