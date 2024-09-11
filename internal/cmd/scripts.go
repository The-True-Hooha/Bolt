package cmd

import (
	"fmt"
	"os"
	"io/fs"
	"sort"
	"strings"
	
	"github.com/spf13/pflag"
	"github.com/The-True-Hooha/NimbleFiles/internal/utils"
	"github.com/The-True-Hooha/NimbleFiles/internal/utils/output"
)


type lsOptions struct {
	longFormat bool
	showHidden bool
	sortBy string
	reverse bool
}

func handleLsCommand() Command {
	opts := &lsOptions{}

	flags := pflag.NewFlagSet("ls", pflag.ContinueOnError)
	flags.BoolVarP(&opts.longFormat, "long", "l", false, "uses the long listing format")
	flags.BoolVarP(&opts.longFormat, "all", "a", false, "show hidden files")
	flags.BoolVarP(&opts.longFormat, "reverse", "r", false, "reverses the order of files")
	flags.StringVarP(&opts.sortBy, "sort", "s", "name", "sorts by: name, size, createdDate")
	
	return Command {
		Name: "ls",
		Description: "list the directory contents",
		Flags: flags,
		// Execute: ,
	}
}

func executeLsCommand(opts *lsOptions) func(args []string) error{
	return func(args []string) error {
		// current_working_dir, err := os.Getwd()
		currentPath  := "."
		if len(args) > 0{
			currentPath = args[0]
		}

		logger.Info("listing directory contents", "dir", currentPath)

		files, err := sortDirectory(currentPath, opts)
		if err != nil{
			logger.Error("failed to read the directory", "dir", currentPath, "error", err)
			return fmt.Errorf("failed to read the directory %s: %w", currentPath, err)
		}

		printOpts := output.PrintOptions {
			LongFormat: opts.longFormat,
			ShowHidden: opts.showHidden,
			ShouldColor: true,
			Columns: output.GetDefaultColumns(),
		}
		output.PrintFileInfo(os.Stdout, files, printOpts)
		logger.Info("successfully listed directory contents", "dir", currentPath, "fileCount", len(files))
		return nil
	}

}


func sortDirectory(path string, opts *lsOptions) ([]fs.DirEntry, error){
	entries, err := os.ReadDir(path)
	if err != nil{
		return nil, err
	}

	var files []fs.DirEntry

	for _, entry := range entries {
		if !opts.showHidden && strings.HasPrefix(entry.Name(), "."){
			continue
		}
		files = append(files, entry)
		
	}

	sort.Slice(files, func (first, second int) bool {
		less := false

		switch opts.sortBy{
		case "size":
			firstInfo, _ := files[first].Info()
			secondInfo, _ := files[second].Info()
			less = firstInfo.Size() < secondInfo.Size()
		case "createdDate":
			firstInfo, _ := files[first].Info()
			secondInfo, _ := files[second].Info()
			less = firstInfo.ModTime().Before(secondInfo.ModTime())
		default:
			less = files[first].Name() < files[second].Name()
		}

		if opts.reverse {
			return !less
		}
		return less
	})

	return files, nil

}


type Command struct {
	Name        string
	Description string
	Flags interface{}
	Execute     func(args []string) error
}

type CommandRecord struct {
	command map[string]Command
}

func CommandRegistry() *CommandRecord {
	return &CommandRecord{
		command: make(map[string]Command),
	}
}

func (cr *CommandRecord) AddNew(cmd Command) {
	cr.command[cmd.Name] = cmd
}

func (cr *CommandRecord) DisplayCommands() []Command {
	lists := make([]Command, 0, len(cr.command))
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

	// TODO: how about joining tags together like ls -a or handling sub commands
	// a new command
	// cr.AddNew(Command{
	// 	Name:        "ls",
	// 	Description: "list all the contents in a directory",
	// 	Tag: "",
	// 	Execute: func(args []string) error {
	// 		logger.Info("listing all the files in the directory")
	// 		files, err := os.ReadDir(".")
	// 		if err != nil {
	// 			logger.Debug("some error occurred", err)

	// 		}
	// 		for _, file := range files {
	// 			fmt.Println(file.Name())
	// 		}
	// 		return nil
	// 	},
	// })

	cr.AddNew(Command{
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

	cr.AddNew(Command{
		Name:        "pwd",
		Description: "prints the current working directory",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(Command{
		Name:        "cp",
		Description: "copy files and directories",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(Command{
		Name:        "mv",
		Description: "move files or directories",
		Execute: func(args []string) error {
			return nil
		},
	})
	cr.AddNew(Command{
		Name:        "rm",
		Description: "remove files or directories",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(Command{
		Name:        "mkdir",
		Description: "make directories",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(Command{
		Name:        "touch",
		Description: "create a new empty file",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(Command{
		Name:        "find",
		Description: "search for files in a directory",
		Execute: func(args []string) error {
			return nil
		},
	})

	cr.AddNew(Command{
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
