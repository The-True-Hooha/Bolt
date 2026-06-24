package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallDir returns the default directory where bolt should be installed.
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

// BinaryName returns the platform-appropriate binary name.
func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "bolt.exe"
	}
	return "bolt"
}

// Install copies the running binary to dir and adds dir to the user PATH.
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

// Uninstall removes the bolt binary from dir and removes dir from user PATH.
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

// Windows: use PowerShell to update User-level PATH in registry (no admin needed)
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

// Unix: append export line to shell rc files if not already present
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
