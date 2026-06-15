package fileops

import (
	"fmt"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/common"
)

func HandlePwdCommand() common.Command {
	return common.Command{
		Name:        "pwd",
		Description: "print the current working directory",
		Execute: func(args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			fmt.Println(dir)
			return nil
		},
	}
}
