package install

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const githubAPI = "https://api.github.com/repos/The-True-Hooha/Bolt/releases/latest"

func SelfUpdate(currentVersion string) (latestVersion string, upToDate bool, err error) {
	resp, err := http.Get(githubAPI)
	if err != nil {
		return "", false, fmt.Errorf("cannot reach GitHub: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", false, fmt.Errorf("cannot parse release info: %w", err)
	}

	latestVersion = strings.TrimPrefix(release.TagName, "v")
	if latestVersion == currentVersion {
		return latestVersion, true, nil
	}

	os_ := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	}

	// find matching asset
	var assetURL, assetName string
	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, os_) && strings.Contains(name, arch) {
			assetURL = a.BrowserDownloadURL
			assetName = a.Name
			break
		}
	}
	if assetURL == "" {
		return latestVersion, false, fmt.Errorf("no prebuilt for %s/%s in release %s", os_, arch, release.TagName)
	}

	// download asset
	dlResp, err := http.Get(assetURL)
	if err != nil {
		return latestVersion, false, fmt.Errorf("download failed: %w", err)
	}
	defer dlResp.Body.Close()

	tmp, err := os.MkdirTemp("", "bolt-update-*")
	if err != nil {
		return latestVersion, false, err
	}
	defer os.RemoveAll(tmp)

	archivePath := filepath.Join(tmp, assetName)
	f, err := os.Create(archivePath)
	if err != nil {
		return latestVersion, false, err
	}
	if _, err := io.Copy(f, dlResp.Body); err != nil {
		f.Close()
		return latestVersion, false, err
	}
	f.Close()

	// extract binary
	var binPath string
	if strings.HasSuffix(assetName, ".zip") {
		binPath, err = extractZip(archivePath, tmp)
	} else {
		binPath, err = extractTarGz(archivePath, tmp)
	}
	if err != nil {
		return latestVersion, false, fmt.Errorf("extract failed: %w", err)
	}

	// replace running binary
	exe, err := os.Executable()
	if err != nil {
		return latestVersion, false, fmt.Errorf("cannot find current executable: %w", err)
	}
	exe, _ = filepath.EvalSymlinks(exe)

	if err := copyFile(binPath, exe); err != nil {
		return latestVersion, false, fmt.Errorf("cannot replace binary: %w", err)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(exe, 0755)
	}
	return latestVersion, false, nil
}

func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		out := filepath.Join(destDir, filepath.Base(hdr.Name))
		of, err := os.Create(out)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(of, tr); err != nil {
			of.Close()
			return "", err
		}
		of.Close()
		return out, nil
	}
	return "", fmt.Errorf("no file found in archive")
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out := filepath.Join(destDir, filepath.Base(f.Name))
		of, err := os.Create(out)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, err = io.Copy(of, rc)
		of.Close()
		rc.Close()
		if err != nil {
			return "", err
		}
		return out, nil
	}
	return "", fmt.Errorf("no file found in archive")
}

func InstallDir() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "bolt", "bin")
	}
	return filepath.Join(home, ".local", "bin")
}

func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "bolt.exe"
	}
	return "bolt"
}

func Install(dir string) error {
	src, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate current binary: %w", err)
	}
	// resolve symlinks
	src, err = filepath.EvalSymlinks(src)
	if err != nil {
		return fmt.Errorf("cannot resolve binary path: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create install dir: %w", err)
	}

	dst := filepath.Join(dir, BinaryName())
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("cannot copy binary: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dst, 0755); err != nil {
			return fmt.Errorf("cannot set executable bit: %w", err)
		}
	}

	if err := addToPath(dir); err != nil {
		fmt.Printf("warning: could not add %s to PATH automatically: %v\n", dir, err)
		fmt.Printf("Add this to your shell profile manually:\n  export PATH=\"%s:$PATH\"\n", dir)
	}
	return nil
}

func Uninstall(dir string) error {
	dst := filepath.Join(dir, BinaryName())
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove binary: %w", err)
	}
	_ = removeFromPath(dir)
	return nil
}

func addToPath(dir string) error {
	if runtime.GOOS == "windows" {
		return addToPathWindows(dir)
	}
	return addToPathUnix(dir)
}

func removeFromPath(dir string) error {
	if runtime.GOOS == "windows" {
		return removeFromPathWindows(dir)
	}
	return removeFromPathUnix(dir)
}

func addToPathWindows(dir string) error {
	script := fmt.Sprintf(`
$dir = '%s'
$scope = 'User'
$current = [Environment]::GetEnvironmentVariable('PATH', $scope)
if ($current -split ';' -notcontains $dir) {
    [Environment]::SetEnvironmentVariable('PATH', "$current;$dir", $scope)
}
`, dir)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func removeFromPathWindows(dir string) error {
	script := fmt.Sprintf(`
$dir = '%s'
$scope = 'User'
$parts = [Environment]::GetEnvironmentVariable('PATH', $scope) -split ';' | Where-Object { $_ -ne $dir }
[Environment]::SetEnvironmentVariable('PATH', ($parts -join ';'), $scope)
`, dir)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Run()
}

func addToPathUnix(dir string) error {
	home, _ := os.UserHomeDir()
	line := fmt.Sprintf(`export PATH="%s:$PATH"`, dir)

	// already in PATH?
	currentPath := os.Getenv("PATH")
	for _, p := range strings.Split(currentPath, ":") {
		if p == dir {
			return nil // already active
		}
	}

	rcFiles := shellRCFiles(home)
	added := false
	for _, rc := range rcFiles {
		if _, err := os.Stat(rc); os.IsNotExist(err) {
			continue
		}
		data, _ := os.ReadFile(rc)
		if strings.Contains(string(data), dir) {
			added = true
			continue
		}
		f, err := os.OpenFile(rc, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		fmt.Fprintf(f, "\n# added by bolt install\n%s\n", line)
		f.Close()
		fmt.Printf("  added PATH entry to %s\n", rc)
		added = true
	}

	if !added {
		// fallback: write to ~/.profile
		profile := filepath.Join(home, ".profile")
		f, err := os.OpenFile(profile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		fmt.Fprintf(f, "\n# added by bolt install\n%s\n", line)
		f.Close()
		fmt.Printf("  added PATH entry to %s\n", profile)
	}
	return nil
}

func removeFromPathUnix(dir string) error {
	home, _ := os.UserHomeDir()
	for _, rc := range append(shellRCFiles(home), filepath.Join(home, ".profile")) {
		data, err := os.ReadFile(rc)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		var kept []string
		for _, l := range lines {
			if strings.Contains(l, dir) || l == "# added by bolt install" {
				continue
			}
			kept = append(kept, l)
		}
		_ = os.WriteFile(rc, []byte(strings.Join(kept, "\n")), 0644)
	}
	return nil
}

func shellRCFiles(home string) []string {
	return []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".config", "fish", "config.fish"),
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// write to a temp file first, then rename for atomicity
	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	out.Close()
	return os.Rename(tmp, dst)
}
