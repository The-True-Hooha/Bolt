package tagcmd

import (
	"fmt"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/The-True-Hooha/Bolt/internal/config"
)

func HandleTagCommand() common.Command {
	return common.Command{
		Name:        "tag",
		Description: "manage file tags: tag <add|remove|list> <path> [tag]",
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: tag <add|remove|list> <path> [tag]")
			}
			sub, path := args[0], args[1]
			switch sub {
			case "list":
				tags := config.GetFileTags(path)
				if len(tags) == 0 {
					fmt.Printf("%s: no tags\n", path)
					return nil
				}
				fmt.Printf("%s: %s\n", path, strings.Join(tags, ", "))
			case "add":
				if len(args) < 3 {
					return fmt.Errorf("usage: tag add <path> <tag>")
				}
				if err := config.AddFileTag(path, args[2]); err != nil {
					return err
				}
				fmt.Printf("tagged '%s' with '%s'\n", path, args[2])
			case "remove":
				if len(args) < 3 {
					return fmt.Errorf("usage: tag remove <path> <tag>")
				}
				if err := config.RemoveFileTag(path, args[2]); err != nil {
					return err
				}
				fmt.Printf("removed tag '%s' from '%s'\n", args[2], path)
			default:
				return fmt.Errorf("unknown subcommand '%s': use add, remove, or list", sub)
			}
			return nil
		},
	}
}
