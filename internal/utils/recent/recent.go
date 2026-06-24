package recent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/The-True-Hooha/Bolt/internal/config"
)

const maxRecent = 50

type Entry struct {
	Path string    `json:"path"`
	Name string    `json:"name"`
	IsDir bool     `json:"is_dir"`
	At   time.Time `json:"at"`
}

func dataPath() string {
	cfg := config.DefaultDirectory()
	return filepath.Join(cfg.Core.DataDir, "recent.json")
}

func Load() ([]Entry, error) {
	path := dataPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func Push(path string, isDir bool) error {
	entries, _ := Load()

	// remove existing entry for same path
	entries = slices.DeleteFunc(entries, func(e Entry) bool { return e.Path == path })

	entry := Entry{
		Path:  path,
		Name:  filepath.Base(path),
		IsDir: isDir,
		At:    time.Now(),
	}
	entries = append([]Entry{entry}, entries...)

	if len(entries) > maxRecent {
		entries = entries[:maxRecent]
	}

	dir := filepath.Dir(dataPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(dataPath(), data, 0644)
}
