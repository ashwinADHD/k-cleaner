package sentinel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/ashwinADHD/k-cleaner/internal/automation/launchd"
	"github.com/ashwinADHD/k-cleaner/internal/automation/notify"
	"github.com/ashwinADHD/k-cleaner/internal/appinfo"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

const Label = "com.ashwin.kcleaner.sentinel"

var (
	recent   = make(map[string]time.Time)
	recentMu sync.Mutex
	cooldown = 2 * time.Minute
)

// Install registers the Trash watcher LaunchAgent.
func Install() error {
	bin, err := launchd.KcleanBinary()
	if err != nil {
		return err
	}
	stdout := filepath.Join(launchd.LogDir(), "sentinel.log")
	stderr := filepath.Join(launchd.LogDir(), "sentinel.error.log")

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>sentinel</string>
		<string>run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>%s</string>
	<key>StandardErrorPath</key>
	<string>%s</string>
	<key>ProcessType</key>
	<string>Background</string>
</dict>
</plist>
`, Label, launchd.EscapeXML(bin), launchd.EscapeXML(stdout), launchd.EscapeXML(stderr))

	return launchd.Install(Label, plist)
}

// Uninstall removes the Sentinel LaunchAgent.
func Uninstall() error {
	return launchd.Uninstall(Label)
}

// Status describes Sentinel installation state.
func Status() (installed, running bool, plistPath string) {
	return launchd.IsInstalled(Label), launchd.IsRunning(Label), launchd.PlistPath(Label)
}

// Run watches ~/.Trash and notifies when an .app bundle appears (foreground or via launchd).
func Run() error {
	trash := paths.Trash()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(trash); err != nil {
		return fmt.Errorf("watch trash: %w", err)
	}

	// Scan existing trashed apps on startup (optional skip - only new events)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Create|fsnotify.Rename|fsnotify.Write) != 0 {
				handlePath(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}

func handlePath(path string) {
	if !strings.HasSuffix(strings.ToLower(path), ".app") {
		return
	}
	if !isUnderTrash(path) {
		return
	}
	if shouldSkip(path) {
		return
	}

	app, err := appinfo.FromPath(path)
	name := filepath.Base(strings.TrimSuffix(path, ".app"))
	if err == nil {
		name = app.Name
		if app.BundleID == "com.ashwin.kcleaner" {
			return
		}
	}

	_ = notify.MacOS("K-Cleaner Sentinel",
		fmt.Sprintf("%s was trashed — run kclean uninstall-all to remove leftovers.", name))

	yes, err := notify.ConfirmUninstall(name, path)
	if err != nil || !yes {
		return
	}
	_ = notify.OpenTerminalCommand(path)
}

func isUnderTrash(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	trash, _ := filepath.Abs(paths.Trash())
	return abs == trash || strings.HasPrefix(abs, trash+string(os.PathSeparator))
}

func shouldSkip(path string) bool {
	recentMu.Lock()
	defer recentMu.Unlock()
	if t, ok := recent[path]; ok && time.Since(t) < cooldown {
		return true
	}
	recent[path] = time.Now()
	return false
}
