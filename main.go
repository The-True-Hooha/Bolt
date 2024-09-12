package main

import (
	"flag"
	"fmt"
	"github.com/The-True-Hooha/NimbleFiles/internal/cmd"
	"github.com/The-True-Hooha/NimbleFiles/internal/config"
	"log"
	"os"
)

var (
	CurrentPath string
	appConfig   *config.Config
)

func init() {
	var err error

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// get the current working directory
	CurrentPath, err = os.Getwd()
	if err != nil {
		log.Fatalf("failed to get the current working directory: %v\n", err)
	}

	appConfig, err = config.Load()
	if err != nil {
		log.Printf("failed to load the config from system source: %v\n", err)
		appConfig = config.DefaultDirectory()
	}

	checkAppDirectoriesExist()
}

func checkAppDirectoriesExist() {
	dir := []string{
		appConfig.CacheDir,
		appConfig.ConfigDir,
		appConfig.DataDir,
	}

	for _, dir := range dir {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf("failed to create config directories: %s\n: %v\n", dir, err)
		}
	}
}

func main() {

	command := cmd.InitCommands()
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("welcome, see the available commands")
		for _, cmd := range command.DisplayCommands() {
			fmt.Printf("  %s: %s\n", cmd.Name, cmd.Description)
		}
		return
	}
	name, args := cmd.ParseCommand(flag.Arg(0))
	err := command.Execute(name, args)
	if err != nil {
		fmt.Printf("some error here %s", err)
	}

}
