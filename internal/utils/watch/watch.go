package watch

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
)

var (
	colorCreate = color.New(color.FgGreen, color.Bold)
	colorWrite  = color.New(color.FgYellow, color.Bold)
	colorRemove = color.New(color.FgRed, color.Bold)
	colorRename = color.New(color.FgCyan, color.Bold)
	colorChmod  = color.New(color.FgMagenta)
)

func HandleWatchCommand() common.Command {
	flags := pflag.NewFlagSet("watch", pflag.ContinueOnError)
	recursive := flags.BoolP("recursive", "r", false, "watch directory recursively")

	return common.Command{
		Name:        "watch",
		Description: "watch a directory for file system changes",
		Flags:       flags,
		Execute: func(args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			abs, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("cannot resolve path: %w", err)
			}
			if _, err := os.Stat(abs); err != nil {
				return fmt.Errorf("cannot access '%s': %w", abs, err)
			}

			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				return fmt.Errorf("cannot create watcher: %w", err)
			}
			defer watcher.Close()

			if *recursive {
				if err := addRecursive(watcher, abs); err != nil {
					return err
				}
			} else {
				if err := watcher.Add(abs); err != nil {
					return fmt.Errorf("cannot watch '%s': %w", abs, err)
				}
			}

			fmt.Printf("watching %s (ctrl+c to stop)\n", abs)

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return nil
					}
					printEvent(event)

					// auto-watch new directories in recursive mode
					if *recursive && event.Has(fsnotify.Create) {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							_ = watcher.Add(event.Name)
						}
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return nil
					}
					fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)

				case <-quit:
					fmt.Println("\nstopped watching")
					return nil
				}
			}
		},
	}
}

func printEvent(event fsnotify.Event) {
	ts := time.Now().Format("15:04:05")
	rel := event.Name

	switch {
	case event.Has(fsnotify.Create):
		fmt.Printf("[%s] %s %s\n", ts, colorCreate.Sprint("CREATE"), rel)
	case event.Has(fsnotify.Write):
		fmt.Printf("[%s] %s  %s\n", ts, colorWrite.Sprint("WRITE "), rel)
	case event.Has(fsnotify.Remove):
		fmt.Printf("[%s] %s %s\n", ts, colorRemove.Sprint("REMOVE"), rel)
	case event.Has(fsnotify.Rename):
		fmt.Printf("[%s] %s %s\n", ts, colorRename.Sprint("RENAME"), rel)
	case event.Has(fsnotify.Chmod):
		fmt.Printf("[%s] %s  %s\n", ts, colorChmod.Sprint("CHMOD "), rel)
	}
}

func addRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible dirs
		}
		if d.IsDir() {
			if err := watcher.Add(path); err != nil {
				fmt.Fprintf(os.Stderr, "warning: cannot watch '%s': %v\n", path, err)
			}
		}
		return nil
	})
}
