package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"os/exec"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/The-True-Hooha/Bolt/internal/config"
	"github.com/The-True-Hooha/Bolt/internal/utils/archive"
	"github.com/The-True-Hooha/Bolt/internal/tui"
	"github.com/The-True-Hooha/Bolt/internal/utils/bookmarks"
	"github.com/The-True-Hooha/Bolt/internal/utils/fileops"
	"github.com/The-True-Hooha/Bolt/internal/utils/find"
	"github.com/The-True-Hooha/Bolt/internal/utils/grep"
	lscmd "github.com/The-True-Hooha/Bolt/internal/utils/ls"
	"github.com/The-True-Hooha/Bolt/internal/utils/preview"
	"github.com/The-True-Hooha/Bolt/internal/utils/search"
	"github.com/The-True-Hooha/Bolt/internal/utils/trash"
	"github.com/The-True-Hooha/Bolt/internal/utils/watch"
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

func getConfig() map[string]any {
	return viper.AllSettings()
}

func LoadInit() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (set default $HOME/.bolt.yaml)")
	rootCmd.PersistentFlags().StringP("author", "", "David Ogar", fmt.Sprintf("©%d David Ogar", time.Now().Year()))
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "", "", "Name of license for the project")
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

func (cr *CommandRecord) RunUI() error {
	p := tea.NewProgram(tui.New(""), tea.WithAltScreen())
	_, err := p.Run()
	return err
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
		Description: "change the current working directory",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: cd <directory>")
			}
			if err := os.Chdir(args[0]); err != nil {
				return fmt.Errorf("cd: %s: %w", args[0], err)
			}
			return nil
		},
	})

	cr.AddNew(common.Command{
		Name:        "ui",
		Description: "open the interactive TUI file manager",
		Execute: func(args []string) error {
			start := ""
			if len(args) > 0 {
				start = args[0]
			}
			p := tea.NewProgram(tui.New(start), tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	})

	cr.AddNew(fileops.HandlePwdCommand())
	cr.AddNew(fileops.HandleCpCommand())
	cr.AddNew(fileops.HandleMvCommand())
	cr.AddNew(fileops.HandleRmCommand())
	cr.AddNew(fileops.HandleMkdirCommand())
	cr.AddNew(fileops.HandleTouchCommand())
	cr.AddNew(find.HandleFindCommand())
	cr.AddNew(grep.HandleGrepCommand())
	cr.AddNew(preview.HandlePreviewCommand())
	cr.AddNew(trash.HandleTrashCommand())
	cr.AddNew(trash.HandleTrashListCommand())
	cr.AddNew(trash.HandleTrashRestoreCommand())
	cr.AddNew(trash.HandleTrashEmptyCommand())
	cr.AddNew(bookmarks.HandleBookmarkCommand())
	cr.AddNew(watch.HandleWatchCommand())
	cr.AddNew(archive.HandleZipCommand())
	cr.AddNew(archive.HandleUnzipCommand())
	cr.AddNew(archive.HandleTarCommand())
	cr.AddNew(search.HandleSearchCommand())
	cr.AddNew(search.HandleIndexCommand())

	// Register enabled plugins from config.toml [plugins] section
	for name, plugin := range config.GetPlugins() {
		if !plugin.Enabled {
			continue
		}
		name, plugin := name, plugin // capture loop vars
		desc := plugin.Description
		if desc == "" {
			desc = fmt.Sprintf("plugin: %s", name)
		}
		cr.AddNew(common.Command{
			Name:        name,
			Description: desc,
			Execute: func(args []string) error {
				cmd := exec.Command(plugin.Command, args...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			},
		})
	}

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
