package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/common"
)

func HandleDuplicateCommand() common.Command {
	return common.Command{
		Name:        "duplicate",
		Description: "duplicate a file with a _copy suffix",
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: duplicate <file>...")
			}
			for _, src := range args {
				dst, err := duplicatePath(src)
				if err != nil {
					return err
				}
				fmt.Printf("duplicated: %s → %s\n", src, dst)
			}
			return nil
		},
	}
}

func duplicatePath(src string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("cannot access '%s': %w", src, err)
	}

	dir := filepath.Dir(src)
	base := filepath.Base(src)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	dst := filepath.Join(dir, stem+"_copy"+ext)
	for i := 2; ; i++ {
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			break
		}
		dst = filepath.Join(dir, fmt.Sprintf("%s_copy%d%s", stem, i, ext))
	}

	if info.IsDir() {
		if err := copyDir(src, dst); err != nil {
			return "", fmt.Errorf("duplicate dir failed: %w", err)
		}
	} else {
		if err := copyFile(src, dst, info.Mode()); err != nil {
			return "", fmt.Errorf("duplicate failed: %w", err)
		}
	}
	return dst, nil
}
