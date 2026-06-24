package diffcmd

import (
	"fmt"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/The-True-Hooha/Bolt/internal/utils/diff"
	"github.com/spf13/pflag"
)

func HandleDiffCommand() common.Command {
	flags := pflag.NewFlagSet("diff", pflag.ContinueOnError)
	noColor := flags.Bool("no-color", false, "disable colored output")

	return common.Command{
		Name:        "diff",
		Description: "show unified diff between two files",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: diff <file-a> <file-b>")
			}
			a, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("cannot read '%s': %w", args[0], err)
			}
			b, err := os.ReadFile(args[1])
			if err != nil {
				return fmt.Errorf("cannot read '%s': %w", args[1], err)
			}

			result := diff.Unified(args[0], args[1], string(a), string(b))

			if *noColor {
				fmt.Print(result)
				return nil
			}

			const red = "\x1b[31m"
			const green = "\x1b[32m"
			const cyan = "\x1b[36m"
			const reset = "\x1b[0m"

			for _, line := range splitLines(result) {
				switch {
				case len(line) > 0 && line[0] == '+':
					fmt.Printf("%s%s%s\n", green, line, reset)
				case len(line) > 0 && line[0] == '-':
					fmt.Printf("%s%s%s\n", red, line, reset)
				case len(line) > 2 && line[:3] == "---" || len(line) > 2 && line[:3] == "+++":
					fmt.Printf("%s%s%s\n", cyan, line, reset)
				default:
					fmt.Println(line)
				}
			}
			return nil
		},
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, b := range s {
		if b == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
