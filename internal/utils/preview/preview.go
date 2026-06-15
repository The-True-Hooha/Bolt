package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

const maxPreviewBytes = 512 * 1024 // 512 KB

func HandlePreviewCommand() common.Command {
	flags := pflag.NewFlagSet("preview", pflag.ContinueOnError)
	theme := flags.StringP("theme", "t", "monokai", "syntax highlight theme (monokai, dracula, github, solarized-dark, etc.)")
	plain := flags.BoolP("plain", "p", false, "disable syntax highlighting, print raw content")
	lines := flags.IntP("lines", "n", 0, "limit output to N lines (0 = all)")

	return common.Command{
		Name:        "preview",
		Description: "preview file contents with syntax highlighting",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: preview [options] <file>")
			}

			path := args[0]
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("cannot access '%s': %w", path, err)
			}
			if info.IsDir() {
				return fmt.Errorf("'%s' is a directory", path)
			}
			if info.Size() > maxPreviewBytes {
				fmt.Fprintf(os.Stderr, "warning: file is large (%.1f MB), showing first 512 KB\n",
					float64(info.Size())/1024/1024)
			}

			data, err := readCapped(path)
			if err != nil {
				return fmt.Errorf("cannot read '%s': %w", path, err)
			}

			content := string(data)
			if *lines > 0 {
				content = headLines(content, *lines)
			}

			if *plain {
				fmt.Print(content)
				return nil
			}

			return printHighlighted(path, content, *theme)
		},
	}
}

func readCapped(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, maxPreviewBytes)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return nil, err
	}
	return buf[:n], nil
}

func headLines(content string, n int) string {
	lines := strings.SplitN(content, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

func printHighlighted(path, content, themeName string) error {
	lexer := lexers.Match(filepath.Base(path))
	if lexer == nil {
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get(themeName)
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return fmt.Errorf("tokenise error: %w", err)
	}

	return formatter.Format(os.Stdout, style, iterator)
}
