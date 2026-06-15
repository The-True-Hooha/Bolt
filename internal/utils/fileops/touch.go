package fileops

import (
	"fmt"
	"os"
	"time"

	"github.com/The-True-Hooha/Bolt/internal/common"
)

func HandleTouchCommand() common.Command {
	return common.Command{
		Name:        "touch",
		Description: "create empty files or update timestamps",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing operand: specify at least one file name")
			}
			now := time.Now()
			for _, name := range args {
				f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("cannot touch '%s': %w", name, err)
				}
				f.Close()
				if err := os.Chtimes(name, now, now); err != nil {
					return fmt.Errorf("cannot update timestamp for '%s': %w", name, err)
				}
			}
			return nil
		},
	}
}
