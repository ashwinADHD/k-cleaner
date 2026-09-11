package launchd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

const agentsSubdir = "LaunchAgents"

// AgentDir returns ~/Library/LaunchAgents.
func AgentDir() string {
	return filepath.Join(paths.Home(), "Library", agentsSubdir)
}

// LogDir returns ~/Library/Logs/K-Cleaner.
func LogDir() string {
	return filepath.Join(paths.Home(), "Library", "Logs", "K-Cleaner")
}

// KcleanBinary resolves the current kclean executable path.
func KcleanBinary() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// PlistPath returns the path for a LaunchAgent plist by label.
func PlistPath(label string) string {
	return filepath.Join(AgentDir(), label+".plist")
}

// Install writes a plist and loads the agent.
func Install(label, plistContent string) error {
	if err := os.MkdirAll(AgentDir(), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(LogDir(), 0o755); err != nil {
		return err
	}
	path := PlistPath(label)
	if err := os.WriteFile(path, []byte(plistContent), 0o644); err != nil {
		return err
	}
	_ = Unload(label)
	return Load(label)
}

// Uninstall unloads and removes a LaunchAgent plist.
func Uninstall(label string) error {
	_ = Unload(label)
	path := PlistPath(label)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Load starts a LaunchAgent.
func Load(label string) error {
	path := PlistPath(label)
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("plist not found: %s", path)
	}
	out, err := exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%d", os.Getuid()), path).CombinedOutput()
	if err != nil {
		// Fallback for older macOS
		out2, err2 := exec.Command("launchctl", "load", path).CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("launchctl load failed: %s (%v); bootstrap: %s (%v)", string(out2), err2, string(out), err)
		}
	}
	return nil
}

// Unload stops a LaunchAgent.
func Unload(label string) error {
	path := PlistPath(label)
	out, err := exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d", os.Getuid()), path).CombinedOutput()
	if err != nil {
		_, _ = exec.Command("launchctl", "unload", path).CombinedOutput()
		_ = out
	}
	return nil
}

// IsInstalled reports whether the plist file exists.
func IsInstalled(label string) bool {
	_, err := os.Stat(PlistPath(label))
	return err == nil
}

// IsRunning checks if launchctl knows about the job.
func IsRunning(label string) bool {
	out, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), label)
}

// EscapeXML escapes characters for plist string values.
func EscapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
