package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	
)

type Config struct {
	CacheDir  string `json:"cache_dir"`
	ConfigDir string `json:"config_dir"`
	DataDir   string `json:"data_dir"`
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(homeDir, ".bolt", "config.json")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultDirectory(), nil
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}


func DefaultDirectory() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		CacheDir:  filepath.Join(homeDir, ".cache", "bolt"),
		ConfigDir: filepath.Join(homeDir, ".config", "bolt"),
		DataDir:   filepath.Join(homeDir, ".local", "share", "bolt"),
	}

}
