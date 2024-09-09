package cmd

import (
	"fmt"
	"strings"
)

// "fmt"

type Command struct {
	Name string
	Description string
	Execute func(args [] string) error
}

type CommandRecord struct {
	command map[string] Command
}

func CommandRegistry() *CommandRecord {
	return &CommandRecord{
		command: make(map[string]Command),
	}
}


func (cr *CommandRecord) AddNew (cmd Command){
	cr.command[cmd.Name] = cmd
}

func (cr *CommandRecord) DisplayCommands() []Command {
	lists := make([] Command, 0, len(cr.command))
	for _, cmd := range cr.command {
		lists = append(lists, cmd)
	}
	return lists
}

// run the command by it's name with the argument attached
func (cr *CommandRecord) Execute(name string, args [] string) error {
	cmd, isFound := cr.command[name]
	if !isFound {
		return fmt.Errorf("you entered an unknown command: %s", name)
	}
	return cmd.Execute(args)
}

func InitCommands() *CommandRecord {
	cr:= CommandRegistry()

	// a new command
	cr.AddNew(Command{
		Name: "li",
		Description: "list all the files in a specified directory",
		Execute: func(args []string) error {
			// should do something like listing the files or so
			fmt.Println("listing all the files")
			return nil
		},
	})

	// another command
	cr.AddNew(Command{
		Name: "op",
		Description: "open or change current directory",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("please add a file path or directory you want to query")
			}
			fmt.Printf("change directory %s\n", args[0])
			return nil
		},
		
	})

	return cr
}

// split the command to return the name and args
func ParseCommand(input string) (string, []string){
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

