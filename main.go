package main

import (
	"fmt"
	"log"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/cmd"
	"github.com/The-True-Hooha/Bolt/internal/config"
)

var (
	CurrentPath string
	appConfig   *config.Config
)

func init() {
	cmd.LoadInit()
}

func CheckAppDirectoriesExist() {
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
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
