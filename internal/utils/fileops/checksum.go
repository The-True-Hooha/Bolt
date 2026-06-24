package fileops

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"os"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

func HandleChecksumCommand() common.Command {
	flags := pflag.NewFlagSet("checksum", pflag.ContinueOnError)
	algo := flags.StringP("algo", "a", "sha256", "hash algorithm: sha256, sha512, md5")

	return common.Command{
		Name:        "checksum",
		Description: "compute file checksum (sha256 by default)",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: checksum [--algo sha256|sha512|md5] <file>...")
			}
			for _, path := range args {
				sum, err := computeChecksum(path, *algo)
				if err != nil {
					return err
				}
				fmt.Printf("%s  %s\n", sum, path)
			}
			return nil
		},
	}
}

func computeChecksum(path, algo string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open '%s': %w", path, err)
	}
	defer f.Close()

	switch algo {
	case "sha512":
		h := sha512.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	case "md5":
		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	default:
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", h.Sum(nil)), nil
	}
}
