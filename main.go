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
	// fmt.Println("does this line even work??")
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
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
