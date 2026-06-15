package fileops

import (
	"fmt"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

func HandleMkdirCommand() common.Command {
	flags := pflag.NewFlagSet("mkdir", pflag.ContinueOnError)
	parents := flags.BoolP("parents", "p", false, "create parent directories as needed")

	return common.Command{
		Name:        "mkdir",
		Description: "create directories",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing operand: specify at least one directory name")
			}
			for _, dir := range args {
				var err error
				if *parents {
					err = os.MkdirAll(dir, 0755)
				} else {
					err = os.Mkdir(dir, 0755)
				}
				if err != nil {
					return fmt.Errorf("cannot create directory '%s': %w", dir, err)
				}
			}
			return nil
		},
	}
}
