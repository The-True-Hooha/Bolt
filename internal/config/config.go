package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// V is the shared viper instance used across the app.
var V = viper.New()

type Config struct {
	Core    CoreConfig              `mapstructure:"core"`
	Ls      LsConfig                `mapstructure:"ls"`
	Tags    TagsConfig              `mapstructure:"tags"`
	Plugins map[string]PluginConfig `mapstructure:"plugins"`
}

type CoreConfig struct {
	CacheDir  string `mapstructure:"cache_dir"`
	ConfigDir string `mapstructure:"config_dir"`
	DataDir   string `mapstructure:"data_dir"`
}

type LsConfig struct {
	DefaultSort string `mapstructure:"default_sort"`
	ShowHidden  bool   `mapstructure:"show_hidden"`
	Color       bool   `mapstructure:"color"`
}

type TagsConfig struct {
	Files map[string][]string `mapstructure:"files"`
}

type PluginConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Command     string `mapstructure:"command"`
	Description string `mapstructure:"description"`
}

// CacheDir / ConfigDir / DataDir kept for backwards compat with main.go
func (c *Config) GetCacheDir() string  { return c.Core.CacheDir }
func (c *Config) GetConfigDir() string { return c.Core.ConfigDir }
func (c *Config) GetDataDir() string   { return c.Core.DataDir }

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bolt"), nil
}

func Load() (*Config, error) {
	dir, err := configPath()
	if err != nil {
		return nil, err
	}

	V.SetConfigName("config")
	V.SetConfigType("toml")
	V.AddConfigPath(dir)

	setDefaults(dir)

	if err := V.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// No config file yet — write defaults so user has something to edit
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		if err := V.WriteConfigAs(filepath.Join(dir, "config.toml")); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
	}

	var cfg Config
	if err := V.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

func DefaultDirectory() *Config {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "bolt")
	setDefaults(dir)
	var cfg Config
	_ = V.Unmarshal(&cfg)
	return &cfg
}

func setDefaults(cfgDir string) {
	home, _ := os.UserHomeDir()

	V.SetDefault("core.cache_dir", filepath.Join(home, ".cache", "bolt"))
	V.SetDefault("core.config_dir", cfgDir)
	V.SetDefault("core.data_dir", filepath.Join(home, ".local", "share", "bolt"))

	V.SetDefault("ls.default_sort", "name")
	V.SetDefault("ls.show_hidden", false)
	V.SetDefault("ls.color", true)

	V.SetDefault("tags.files", map[string][]string{})
	V.SetDefault("plugins", map[string]PluginConfig{})
}


func GetFileTags(filename string) []string {
	key := "tags.files." + sanitizeKey(filename)
	return V.GetStringSlice(key)
}

func AddFileTag(filename, tag string) error {
	existing := GetFileTags(filename)
	for _, t := range existing {
		if t == tag {
			return nil // already exists
		}
	}
	V.Set("tags.files."+sanitizeKey(filename), append(existing, tag))
	return V.WriteConfig()
}

func RemoveFileTag(filename, tag string) error {
	existing := GetFileTags(filename)
	updated := existing[:0]
	for _, t := range existing {
		if t != tag {
			updated = append(updated, t)
		}
	}
	V.Set("tags.files."+sanitizeKey(filename), updated)
	return V.WriteConfig()
}

func sanitizeKey(path string) string {
	result := []byte(path)
	for i, b := range result {
		if b == '.' || b == '/' || b == '\\' {
			result[i] = '_'
		}
	}
	return string(result)
}


func GetPlugins() map[string]PluginConfig {
	var cfg Config
	if err := V.Unmarshal(&cfg); err != nil {
		return nil
	}
	return cfg.Plugins
}
