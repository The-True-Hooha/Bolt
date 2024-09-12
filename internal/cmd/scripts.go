package cmd

import (
	"fmt"
	"strings"

	"github.com/The-True-Hooha/NimbleFiles/internal/utils/ls"
	"github.com/The-True-Hooha/NimbleFiles/internal/common"
)

type CommandRecord struct {
	command map[string]common.Command
}

func CommandRegistry() *CommandRecord {
	return &CommandRecord{
		command: make(map[string]common.Command),
	}
}

func (cr *CommandRecord) AddNew(cmd common.Command) {
	cr.command[cmd.Name] = cmd
}

func (cr *CommandRecord) DisplayCommands() []common.Command {
	lists := make([]common.Command, 0, len(cr.command))
	for _, cmd := range cr.command {
		lists = append(lists, cmd)
	}
	return lists
}

// run the command by it's name with the argument attached
func (cr *CommandRecord) Execute(name string, args []string) error {
	cmd, isFound := cr.command[name]
	if !isFound {
		return fmt.Errorf("you entered an unknown command: %s", name)
	}
	return cmd.Execute(args)
}

func InitCommands() *CommandRecord {
	cr := CommandRegistry()

	cr.AddNew(lscmd.HandleLsCommandTags())

	cr.AddNew(common.Command{
		Name:        "cd",
		Description: "change the current directory",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("please add a file path or directory you want to switch to")
			}
			fmt.Printf("change directory %s\n", args[0])
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "pwd",
		Description: "prints the current working directory",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "cp",
		Description: "copy files and directories",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "mv",
		Description: "move files or directories",
		Execute: func(args []string) error {
			return nil
		},
	})
	cr.AddNew(common.Command{
		Name:        "rm",
		Description: "remove files or directories",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "mkdir",
		Description: "make directories",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "touch",
		Description: "create a new empty file",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "find",
		Description: "search for files in a directory",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "grep",
		Description: "search for a pattern within files",
		Execute: func(args []string) error {
			return nil
		},
	})

	return cr
}

// split the command to return the name and args
func ParseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
