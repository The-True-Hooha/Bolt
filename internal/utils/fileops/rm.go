package fileops

import (
	"fmt"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

func HandleRmCommand() common.Command {
	flags := pflag.NewFlagSet("rm", pflag.ContinueOnError)
	recursive := flags.BoolP("recursive", "r", false, "remove directories and their contents recursively")
	force := flags.BoolP("force", "f", false, "ignore nonexistent files, never prompt")

	return common.Command{
		Name:        "rm",
		Description: "remove files or directories",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing operand: specify at least one file or directory")
			}
			for _, target := range args {
				info, err := os.Lstat(target)
				if err != nil {
					if os.IsNotExist(err) && *force {
						continue
					}
					return fmt.Errorf("cannot remove '%s': %w", target, err)
				}

				if info.IsDir() && !*recursive {
					return fmt.Errorf("cannot remove '%s': is a directory (use -r to remove recursively)", target)
				}

				if err := os.RemoveAll(target); err != nil {
					return fmt.Errorf("cannot remove '%s': %w", target, err)
				}
			}
			return nil
		},
	}
}
