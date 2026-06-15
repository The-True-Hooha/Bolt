package lscmd

import (
	"github.com/The-True-Hooha/Bolt/internal/config"
)

func GetFileTags(filename string) ([]string, error) {
	return config.GetFileTags(filename), nil
}

func AddFileTags(filename, tag string) error {
	return config.AddFileTag(filename, tag)
}

func RemoveFileTags(filename, tag string) error {
	return config.RemoveFileTag(filename, tag)
}
