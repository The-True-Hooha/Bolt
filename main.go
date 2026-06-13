package main

import (
	"fmt"
	"log"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/cmd"
	"github.com/The-True-Hooha/Bolt/internal/config"
)

func init() {
	cmd.LoadInit()
}

func ensureAppDirectories(cfg *config.Config) {
	for _, dir := range []string{cfg.CacheDir, cfg.ConfigDir, cfg.DataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("failed to create directory %s: %v\n", dir, err)
		}
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("failed to load config: %v\n", err)
		cfg = config.DefaultDirectory()
	}
	ensureAppDirectories(cfg)

	command := cmd.InitCommands()
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
