package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/The-True-Hooha/Bolt/internal/config"
	"github.com/spf13/pflag"
)

func trashDir() string {
	cfg, err := config.Load()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "bolt", "trash")
	}
	return filepath.Join(cfg.Core.DataDir, "trash")
}

func HandleTrashCommand() common.Command {
	return common.Command{
		Name:        "trash",
		Description: "move files to trash instead of deleting permanently",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: trash <file>...")
			}
			dir := trashDir()
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("cannot create trash directory: %w", err)
			}
			for _, target := range args {
				if err := moveToTrash(target, dir); err != nil {
					return err
				}
				fmt.Printf("trashed: %s\n", target)
			}
			return nil
		},
	}
}

func HandleTrashListCommand() common.Command {
	return common.Command{
		Name:        "trash-list",
		Description: "list files in trash",
		Execute: func(args []string) error {
			dir := trashDir()
			entries, err := os.ReadDir(dir)
			if os.IsNotExist(err) {
				fmt.Println("trash is empty")
				return nil
			}
			if err != nil {
				return fmt.Errorf("cannot read trash: %w", err)
			}
			if len(entries) == 0 {
				fmt.Println("trash is empty")
				return nil
			}
			for _, e := range entries {
				info, _ := e.Info()
				if info != nil {
					fmt.Printf("%-40s  %s\n", e.Name(), info.ModTime().Format("2006-01-02 15:04:05"))
				} else {
					fmt.Println(e.Name())
				}
			}
			return nil
		},
	}
}

func HandleTrashRestoreCommand() common.Command {
	return common.Command{
		Name:        "trash-restore",
		Description: "restore a file from trash",
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: trash-restore <trashed-name> <destination>")
			}
			src := filepath.Join(trashDir(), args[0])
			dst := args[1]
			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("cannot restore '%s': %w", args[0], err)
			}
			fmt.Printf("restored: %s → %s\n", args[0], dst)
			return nil
		},
	}
}

func HandleTrashEmptyCommand() common.Command {
	flags := pflag.NewFlagSet("trash-empty", pflag.ContinueOnError)
	force := flags.BoolP("force", "f", false, "skip confirmation")

	return common.Command{
		Name:        "trash-empty",
		Description: "permanently delete all files in trash",
		Flags:       flags,
		Execute: func(args []string) error {
			dir := trashDir()
			if !*force {
				fmt.Print("permanently delete all trash? [y/N] ")
				var answer string
				fmt.Scanln(&answer)
				if answer != "y" && answer != "Y" {
					fmt.Println("aborted")
					return nil
				}
			}
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("cannot empty trash: %w", err)
			}
			fmt.Println("trash emptied")
			return nil
		},
	}
}

func moveToTrash(target, trashDir string) error {
	if _, err := os.Lstat(target); err != nil {
		return fmt.Errorf("cannot access '%s': %w", target, err)
	}

	base := filepath.Base(target)
	// stamp name to avoid collisions
	stamp := time.Now().Format("20060102-150405")
	dest := filepath.Join(trashDir, fmt.Sprintf("%s.%s", base, stamp))

	if err := os.Rename(target, dest); err != nil {
		// cross-device: copy then remove
		return crossDeviceMoveToTrash(target, dest)
	}
	return nil
}

func crossDeviceMoveToTrash(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, info.Mode()); err != nil {
		return err
	}
	return os.Remove(src)
}
