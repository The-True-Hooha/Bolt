package bookmarks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/viper"

	"github.com/The-True-Hooha/Bolt/internal/config"
)

func bookmarkKey(name string) string {
	return "bookmarks." + name
}

func HandleBookmarkCommand() common.Command {
	return common.Command{
		Name:        "bm",
		Description: "manage bookmarks: bm <add|go|list|rm> [name] [path]",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return listBookmarks()
			}
			switch args[0] {
			case "add":
				return addBookmark(args[1:])
			case "go":
				return goBookmark(args[1:])
			case "rm", "remove":
				return removeBookmark(args[1:])
			case "list", "ls":
				return listBookmarks()
			default:
				return fmt.Errorf("unknown subcommand '%s': use add, go, list, rm", args[0])
			}
		},
	}
}

func addBookmark(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bm add <name> [path]")
	}
	name := args[0]
	path := "."
	if len(args) > 1 {
		path = args[1]
	}

	abs, err := resolveAndValidate(path)
	if err != nil {
		return err
	}

	config.V.Set(bookmarkKey(name), abs)
	if err := config.V.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save bookmark: %w", err)
	}
	fmt.Printf("bookmark added: %s → %s\n", name, abs)
	return nil
}

func goBookmark(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bm go <name>")
	}
	name := args[0]
	path := config.V.GetString(bookmarkKey(name))
	if path == "" {
		return fmt.Errorf("no bookmark named '%s'", name)
	}

	// Print the path so shell integration can cd to it.
	// With eval $(bolt shell-init), the shell will intercept this and cd.
	fmt.Println(path)
	return nil
}

func removeBookmark(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: bm rm <name>")
	}
	name := args[0]
	if config.V.GetString(bookmarkKey(name)) == "" {
		return fmt.Errorf("no bookmark named '%s'", name)
	}

	// viper doesn't support deleting keys directly; rebuild bookmarks map without the key
	all := config.V.GetStringMapString("bookmarks")
	delete(all, name)
	config.V.Set("bookmarks", all)

	if err := config.V.WriteConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("bookmark removed: %s\n", name)
	return nil
}

func listBookmarks() error {
	all := viper.GetStringMapString("bookmarks")
	if len(all) == 0 {
		fmt.Println("no bookmarks saved")
		return nil
	}
	fmt.Println("bookmarks:")
	for name, path := range all {
		fmt.Printf("  %-20s %s\n", name, path)
	}
	return nil
}

func resolveAndValidate(path string) (string, error) {
	abs, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if path != "." {
		abs = path
	}
	abs, err = absPath(abs)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("path does not exist: %s", abs)
	}
	return abs, nil
}

func absPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home + path[1:], nil
	}
	return filepath.Abs(path)
}
