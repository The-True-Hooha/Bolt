package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/The-True-Hooha/Bolt/internal/common"
	lscmd "github.com/The-True-Hooha/Bolt/internal/utils/ls"
)

var longStory = `
A blazingly fast solution that ensures you can navigate, manage, 
and manipulate your files with unparalleled efficiency.
Whether you are a seasoned developer or a casual user, 
this terminal-based file manager offers a seamless experience
that enhances productivity and streamlines workflows. 
`

var rootCmd = &cobra.Command{
	Use:   "bolt",
	Short: "A blazingly fast modern terminal based file manager written in Go",
	Long:  longStory,
}

var (
	cfgFile     string
	userLicense string
	version = "0.1.0"
)

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".bolt")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}
}

func getConfig() map[string]interface{} {
	return viper.AllSettings()
}

func LoadInit() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (set default $HOME/.bolt.yaml)")
	rootCmd.PersistentFlags().StringP("author", "a", "David Ogar", fmt.Sprintf("©%d David Ogar", time.Now().Year()))
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "Name of license for the project")
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("Bolt version {{.Version}}\n")
	rootCmd.SuggestionsMinimumDistance = 1

	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("license", rootCmd.PersistentFlags().Lookup("license"))

	viper.SetDefault("author", fmt.Sprintf("©%d David Ogar <owogogahhero@outlook.com>", time.Now().Year()))
	viper.SetDefault("license", "MIT")

}

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

	cobra := &cobra.Command{
		Use:   cmd.Name,
		Short: cmd.Description,
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.Execute(args)
		},
	}

	if cmd.Flags != nil {
		cobra.Flags().AddFlagSet(cmd.Flags)
	}

	rootCmd.AddCommand(cobra)
}

func (cr *CommandRecord) DisplayCommands() []common.Command {
	lists := make([]common.Command, 0, len(cr.command))
	for _, cmd := range cr.command {
		lists = append(lists, cmd)
	}
	return lists
}

func (cr *CommandRecord) Execute() error {
	return rootCmd.Execute()
}

func InitCommands() *CommandRecord {
	cr := CommandRegistry()
	ls := lscmd.HandleLsCommandTags()
	cr.AddNew(ls)

	cr.AddNew(common.Command{ // prints the current device configuration
		Name: "config",
		Description: "display your current configuration",
		Execute: func(args []string) error {
			boltConfig := getConfig()
			for i, v := range boltConfig{
				fmt.Printf("%s: %v\n", i, v)
			}
			return nil
		},
	})

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
