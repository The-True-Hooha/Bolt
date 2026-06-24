package fileops

import (
	"fmt"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/common"
)

func HandleSymlinkCommand() common.Command {
	return common.Command{
		Name:        "symlink",
		Description: "create a symbolic link: symlink <target> <link-name>",
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: symlink <target> <link-name>")
			}
			target, link := args[0], args[1]
			if err := os.Symlink(target, link); err != nil {
				return fmt.Errorf("symlink failed: %w", err)
			}
			fmt.Printf("created symlink: %s → %s\n", link, target)
			return nil
		},
	}
}

func HandleChmodCommand() common.Command {
	return common.Command{
		Name:        "chmod",
		Description: "change file permissions: chmod <octal> <file>...",
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: chmod <octal-perms> <file>...")
			}
			var perm uint32
			if _, err := fmt.Sscanf(args[0], "%o", &perm); err != nil {
				return fmt.Errorf("invalid permission '%s': use octal e.g. 755", args[0])
			}
			for _, path := range args[1:] {
				if err := os.Chmod(path, os.FileMode(perm)); err != nil {
					return fmt.Errorf("chmod '%s': %w", path, err)
				}
				fmt.Printf("chmod %s %s\n", args[0], path)
			}
			return nil
		},
	}
}
