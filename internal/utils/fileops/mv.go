package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

func HandleMvCommand() common.Command {
	flags := pflag.NewFlagSet("mv", pflag.ContinueOnError)
	force := flags.BoolP("force", "f", false, "overwrite destination without prompt")

	return common.Command{
		Name:        "mv",
		Description: "move or rename files and directories",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: mv [options] <source>... <destination>")
			}

			dst := args[len(args)-1]
			srcs := args[:len(args)-1]

			dstInfo, dstErr := os.Stat(dst)
			dstIsDir := dstErr == nil && dstInfo.IsDir()

			if len(srcs) > 1 && !dstIsDir {
				return fmt.Errorf("target '%s' is not a directory", dst)
			}

			for _, src := range srcs {
				target := dst
				if dstIsDir {
					target = filepath.Join(dst, filepath.Base(src))
				}

				if !*force {
					if _, err := os.Lstat(target); err == nil {
						return fmt.Errorf("'%s' already exists (use -f to overwrite)", target)
					}
				}

				if err := os.Rename(src, target); err != nil {
					// os.Rename fails across filesystems; fall back to copy+delete
					if err := crossDeviceMove(src, target); err != nil {
						return fmt.Errorf("cannot move '%s' to '%s': %w", src, target, err)
					}
				}
			}
			return nil
		},
	}
}

func crossDeviceMove(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return crossDeviceMoveDir(src, dst)
	}
	if err := copyFile(src, dst, srcInfo.Mode()); err != nil {
		return err
	}
	return os.Remove(src)
}

func crossDeviceMoveDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if err := crossDeviceMove(s, d); err != nil {
			return err
		}
	}
	return os.Remove(src)
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
