package dupes

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Group struct {
	Hash  string
	Size  int64
	Paths []string
}

// Find walks root and returns groups of duplicate files (same sha256).
func Find(root string) ([]Group, error) {
	sizeMap := make(map[int64][]string)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() == 0 {
			return nil
		}
		sizeMap[info.Size()] = append(sizeMap[info.Size()], path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	hashMap := make(map[string][]string)
	for _, paths := range sizeMap {
		if len(paths) < 2 {
			continue
		}
		for _, p := range paths {
			h, err := hashFile(p)
			if err != nil {
				continue
			}
			hashMap[h] = append(hashMap[h], p)
		}
	}

	var groups []Group
	for hash, paths := range hashMap {
		if len(paths) < 2 {
			continue
		}
		info, _ := os.Stat(paths[0])
		sz := int64(0)
		if info != nil {
			sz = info.Size()
		}
		groups = append(groups, Group{Hash: hash[:12], Size: sz, Paths: paths})
	}
	return groups, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
