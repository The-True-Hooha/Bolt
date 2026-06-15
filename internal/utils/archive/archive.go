package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/The-True-Hooha/Bolt/internal/common"
	"github.com/spf13/pflag"
)

func HandleZipCommand() common.Command {
	flags := pflag.NewFlagSet("zip", pflag.ContinueOnError)
	output := flags.StringP("output", "o", "", "output archive name (default: <first source>.zip)")

	return common.Command{
		Name:        "zip",
		Description: "compress files or directories into a zip archive",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: zip [options] <source>...")
			}

			dest := *output
			if dest == "" {
				dest = strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0])) + ".zip"
			}
			if !strings.HasSuffix(dest, ".zip") {
				dest += ".zip"
			}

			if err := createZip(dest, args); err != nil {
				return fmt.Errorf("zip failed: %w", err)
			}

			info, _ := os.Stat(dest)
			fmt.Printf("created: %s (%s)\n", dest, humanizeSize(info.Size()))
			return nil
		},
	}
}

func HandleUnzipCommand() common.Command {
	flags := pflag.NewFlagSet("unzip", pflag.ContinueOnError)
	dest := flags.StringP("dest", "d", ".", "destination directory")
	list := flags.BoolP("list", "l", false, "list contents without extracting")

	return common.Command{
		Name:        "unzip",
		Description: "extract a zip archive",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: unzip [options] <archive.zip>")
			}

			if *list {
				return listZip(args[0])
			}

			count, err := extractZip(args[0], *dest)
			if err != nil {
				return fmt.Errorf("unzip failed: %w", err)
			}
			fmt.Printf("extracted %d files to %s\n", count, *dest)
			return nil
		},
	}
}

func HandleTarCommand() common.Command {
	flags := pflag.NewFlagSet("tar", pflag.ContinueOnError)
	output := flags.StringP("output", "o", "", "output archive name (default: <first source>.tar.gz)")
	extract := flags.BoolP("extract", "x", false, "extract instead of create")
	dest := flags.StringP("dest", "d", ".", "destination directory when extracting")
	list := flags.BoolP("list", "l", false, "list contents without extracting")

	return common.Command{
		Name:        "tar",
		Description: "create or extract tar.gz archives",
		Flags:       flags,
		Execute: func(args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: tar [options] <source>...")
			}

			if *list {
				return listTar(args[0])
			}

			if *extract {
				count, err := extractTar(args[0], *dest)
				if err != nil {
					return fmt.Errorf("tar extract failed: %w", err)
				}
				fmt.Printf("extracted %d files to %s\n", count, *dest)
				return nil
			}

			out := *output
			if out == "" {
				out = strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0])) + ".tar.gz"
			}

			if err := createTar(out, args); err != nil {
				return fmt.Errorf("tar failed: %w", err)
			}

			info, _ := os.Stat(out)
			fmt.Printf("created: %s (%s)\n", out, humanizeSize(info.Size()))
			return nil
		},
	}
}


func createZip(dest string, sources []string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, src := range sources {
		if err := addToZip(w, src, filepath.Dir(src)); err != nil {
			return err
		}
	}
	return nil
}

func addToZip(w *zip.Writer, path, base string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := addToZip(w, filepath.Join(path, entry.Name()), base); err != nil {
				return err
			}
		}
		return nil
	}

	rel, err := filepath.Rel(base, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = rel
	header.Method = zip.Deflate

	fw, err := w.CreateHeader(header)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(fw, f)
	return err
}

func extractZip(src, dest string) (int, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		target := filepath.Join(dest, filepath.FromSlash(f.Name))

		// guard against zip slip
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator),
			filepath.Clean(dest)+string(os.PathSeparator)) {
			return count, fmt.Errorf("illegal path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return count, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return count, err
		}

		if err := writeZipEntry(f, target); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func writeZipEntry(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func listZip(src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	fmt.Printf("%-12s  %-20s  %s\n", "Size", "Modified", "Name")
	fmt.Println(strings.Repeat("─", 60))
	for _, f := range r.File {
		fmt.Printf("%-12s  %-20s  %s\n",
			humanizeSize(int64(f.UncompressedSize64)),
			f.Modified.Format("2006-01-02 15:04"),
			f.Name,
		)
	}
	fmt.Printf("\n%d files\n", len(r.File))
	return nil
}


func createTar(dest string, sources []string) error {
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	for _, src := range sources {
		if err := addToTar(tw, src, filepath.Dir(src)); err != nil {
			return err
		}
	}
	return nil
}

func addToTar(tw *tar.Writer, path, base string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := addToTar(tw, filepath.Join(path, entry.Name()), base); err != nil {
				return err
			}
		}
		return nil
	}

	rel, err := filepath.Rel(base, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = rel

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

func extractTar(src, dest string) (int, error) {
	f, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return 0, fmt.Errorf("not a valid gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	count := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}

		target := filepath.Join(dest, filepath.FromSlash(header.Name))

		// guard against tar slip
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator),
			filepath.Clean(dest)+string(os.PathSeparator)) {
			return count, fmt.Errorf("illegal path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return count, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return count, err
			}
			if err := writeTarEntry(tr, target, os.FileMode(header.Mode)); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

func writeTarEntry(tr *tar.Reader, target string, mode os.FileMode) error {
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, tr)
	return err
}

func listTar(src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("not a valid gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	fmt.Printf("%-12s  %-20s  %s\n", "Size", "Modified", "Name")
	fmt.Println(strings.Repeat("─", 60))

	count := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("%-12s  %-20s  %s\n",
			humanizeSize(header.Size),
			header.ModTime.Format("2006-01-02 15:04"),
			header.Name,
		)
		count++
	}
	fmt.Printf("\n%d entries\n", count)
	return nil
}


func humanizeSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
