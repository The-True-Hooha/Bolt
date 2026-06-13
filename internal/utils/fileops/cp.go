package fileops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

func HandleCpCommand() common.Command {
	flags := pflag.NewFlagSet("cp", pflag.ContinueOnError)
	recursive := flags.BoolP("recursive", "r", false, "copy directories recursively")
	force := flags.BoolP("force", "f", false, "overwrite destination without prompt")

	return common.Command{
		Name:        "cp",
		Description: "copy files and directories",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: cp [options] <source>... <destination>")
			}

			dst := args[len(args)-1]
			srcs := args[:len(args)-1]

			dstInfo, dstErr := os.Stat(dst)
			dstIsDir := dstErr == nil && dstInfo.IsDir()

			if len(srcs) > 1 && !dstIsDir {
				return fmt.Errorf("target '%s' is not a directory", dst)
			}

			for _, src := range srcs {
				srcInfo, err := os.Stat(src)
				if err != nil {
					return fmt.Errorf("cannot stat '%s': %w", src, err)
				}

				target := dst
				if dstIsDir {
					target = filepath.Join(dst, filepath.Base(src))
				}

				if !*force {
					if _, err := os.Lstat(target); err == nil {
						return fmt.Errorf("'%s' already exists (use -f to overwrite)", target)
					}
				}

				if srcInfo.IsDir() {
					if !*recursive {
						return fmt.Errorf("'%s' is a directory (use -r to copy recursively)", src)
					}
					if err := copyDir(src, target); err != nil {
						return fmt.Errorf("cannot copy '%s' to '%s': %w", src, target, err)
					}
				} else {
					if err := copyFile(src, target, srcInfo.Mode()); err != nil {
						return fmt.Errorf("cannot copy '%s' to '%s': %w", src, target, err)
					}
				}
			}
			return nil
		},
	}
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := copyFile(s, d, info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}
