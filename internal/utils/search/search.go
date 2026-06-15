package search

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/pflag"

	"github.com/The-True-Hooha/Bolt/internal/common"
)

func HandleSearchCommand() common.Command {
	flags := pflag.NewFlagSet("search", pflag.ContinueOnError)
	recursive := flags.BoolP("recursive", "r", false, "search subdirectories")
	hidden := flags.BoolP("hidden", "H", false, "include hidden files")
	fileType := flags.StringP("type", "t", "", "filter type: f=file, d=dir, l=symlink")
	exts := flags.StringP("ext", "e", "", "comma-separated extensions: .go,.ts")
	ignore := flags.StringArrayP("ignore", "I", nil, "ignore pattern (repeatable)")
	depth := flags.IntP("depth", "d", 0, "max depth (0=unlimited)")
	limit := flags.IntP("limit", "n", 200, "max results")
	exact := flags.Bool("exact", false, "exact substring match")
	glob := flags.Bool("glob", false, "glob pattern match")
	regex := flags.Bool("regex", false, "regex match")
	caseSensitive := flags.BoolP("case-sensitive", "s", false, "case sensitive")
	long := flags.BoolP("long", "l", false, "long format (size + mtime)")
	jsonOut := flags.Bool("json", false, "JSON output")
	forceIndex := flags.Bool("index", false, "force use of trigram index")
	forceWalk := flags.Bool("no-index", false, "force live walk")
	path := flags.StringP("path", "p", "", "search root (default: cwd)")

	return common.Command{
		Name:        "search",
		Description: "search for files: search <pattern> [flags]",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: search <pattern> [flags]")
			}
			pattern := args[0]

			root := *path
			if root == "" {
				var err error
				root, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			mode := ModeFuzzy
			if *exact {
				mode = ModeExact
			} else if *glob {
				mode = ModeGlob
			} else if *regex {
				mode = ModeRegex
			}

			var extList []string
			if *exts != "" {
				for e := range strings.SplitSeq(*exts, ",") {
					e = strings.TrimSpace(e)
					if !strings.HasPrefix(e, ".") {
						e = "." + e
					}
					extList = append(extList, strings.ToLower(e))
				}
			}

			out := make(chan Result, 256)
			var results []Result

			go func() {
				defer close(out)
				_ = Search(Options{
					Root:          root,
					Pattern:       pattern,
					Mode:          mode,
					CaseSensitive: *caseSensitive,
					Recursive:     *recursive,
					MaxDepth:      *depth,
					Hidden:        *hidden,
					Type:          *fileType,
					Exts:          extList,
					Ignore:        *ignore,
					Limit:         *limit,
					ForceIndex:    *forceIndex,
					ForceWalk:     *forceWalk,
				}, out)
			}()

			for r := range out {
				results = append(results, r)
			}

			rank(results)

			if *jsonOut {
				return printJSON(results)
			}

			printResults(results, *long)
			fmt.Printf("\n%d result(s)\n", len(results))
			return nil
		},
	}
}

func HandleIndexCommand() common.Command {
	return common.Command{
		Name:        "index",
		Description: "manage search index: index <build|status|clean> [path]",
		Execute: func(args []string) error {
			sub := "status"
			if len(args) > 0 {
				sub = args[0]
			}
			root := "."
			if len(args) > 1 {
				root = args[1]
			}
			abs, err := absRoot(root)
			if err != nil {
				return err
			}

			switch sub {
			case "build":
				return buildIndex(abs)
			case "status":
				return indexStatus(abs)
			case "clean":
				return cleanIndex(abs)
			default:
				return fmt.Errorf("unknown subcommand '%s': use build, status, clean", sub)
			}
		},
	}
}

func buildIndex(root string) error {
	fmt.Printf("building index for %s...\n", root)
	start := time.Now()
	var last int
	idx, err := BuildIndex(root, func(n int) {
		if n-last >= 10000 {
			fmt.Printf("\r  indexed %d files...", n)
			last = n
		}
	})
	if err != nil {
		return err
	}
	fmt.Printf("\r  indexed %d files in %s\n", idx.meta.FileCount, time.Since(start).Round(time.Millisecond))

	idxPath := IndexPath(root)
	if err := SaveIndex(idx, idxPath); err != nil {
		return fmt.Errorf("save index: %w", err)
	}
	fmt.Printf("index saved: %s\n", idxPath)
	return nil
}

func indexStatus(root string) error {
	idxPath := IndexPath(root)
	info, err := os.Stat(idxPath)
	if err != nil {
		fmt.Printf("no index for %s\n", root)
		return nil
	}
	age := time.Since(info.ModTime()).Round(time.Second)
	size := info.Size()
	fmt.Printf("root:  %s\npath:  %s\nage:   %s\nsize:  %s\n", root, idxPath, age, formatSize(size))
	return nil
}

func cleanIndex(root string) error {
	idxPath := IndexPath(root)
	if err := os.Remove(idxPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("no index to clean")
			return nil
		}
		return err
	}
	fmt.Println("index removed")
	return nil
}

func absRoot(path string) (string, error) {
	if path == "." {
		return os.Getwd()
	}
	return path, nil
}

func printResults(results []Result, long bool) {
	dirColor := color.New(color.FgCyan, color.Bold)
	fileColor := color.New(color.FgWhite)
	for _, r := range results {
		if long && r.Info != nil {
			mod := r.Info.ModTime().Format("2006-01-02 15:04")
			size := formatSize(r.Info.Size())
			fmt.Printf("%-8s  %s  ", size, mod)
		}
		if r.IsDir {
			dirColor.Println(r.Path)
		} else {
			fileColor.Println(r.Path)
		}
	}
}

func printJSON(results []Result) error {
	fmt.Println("[")
	for i, r := range results {
		comma := ","
		if i == len(results)-1 {
			comma = ""
		}
		fmt.Printf(`  {"path":%q,"name":%q,"score":%d,"is_dir":%v}%s`+"\n",
			r.Path, r.Name, r.Score, r.IsDir, comma)
	}
	fmt.Println("]")
	return nil
}

func formatSize(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fK", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
