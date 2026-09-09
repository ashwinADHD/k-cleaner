package paths

import (
	"os"
	"path/filepath"
)

// StandardLibrarySubdirs are macOS Library folders where depth-2 vendor matching applies.
var StandardLibrarySubdirs = map[string]bool{
	"Application Scripts": true, "Application Support": true, "Caches": true,
	"Containers": true, "Group Containers": true, "HTTPStorages": true,
	"Internet Plug-Ins": true, "LaunchAgents": true, "LaunchDaemons": true,
	"Logs": true, "Preferences": true, "PreferencePanes": true,
	"PrivilegedHelperTools": true, "Saved Application State": true,
	"Services": true, "WebKit": true, "Extensions": true, "Frameworks": true,
}

// ForwardScanPaths are searched when uninstalling an app (Pearcleaner Locations.apps).
func ForwardScanPaths() []string {
	home := Home()
	paths := []string{
		home,
		filepath.Join(home, ".config"),
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Desktop"),
		filepath.Join(home, "Applications"),
		filepath.Join(home, "Library"),
		filepath.Join(home, "Library", "Application Scripts"),
		filepath.Join(home, "Library", "Application Support"),
		filepath.Join(home, "Library", "Application Support", "CrashReporter"),
		filepath.Join(home, "Library", "Containers"),
		filepath.Join(home, "Library", "Caches"),
		filepath.Join(home, "Library", "Group Containers"),
		filepath.Join(home, "Library", "HTTPStorages"),
		filepath.Join(home, "Library", "Internet Plug-Ins"),
		filepath.Join(home, "Library", "LaunchAgents"),
		filepath.Join(home, "Library", "Logs"),
		filepath.Join(home, "Library", "Logs", "DiagnosticReports"),
		filepath.Join(home, "Library", "Preferences"),
		filepath.Join(home, "Library", "PreferencePanes"),
		filepath.Join(home, "Library", "Preferences", "ByHost"),
		filepath.Join(home, "Library", "Saved Application State"),
		filepath.Join(home, "Library", "Services"),
		filepath.Join(home, "Library", "WebKit"),
		"/Applications",
		"/Users/Shared",
		"/Users/Shared/Library/Application Support",
		"/Library",
		"/Library/Application Support",
		"/Library/Application Support/CrashReporter",
		"/Library/Caches",
		"/Library/Extensions",
		"/Library/Internet Plug-Ins",
		"/Library/LaunchAgents",
		"/Library/LaunchDaemons",
		"/Library/Logs",
		"/Library/Logs/DiagnosticReports",
		"/Library/Preferences",
		"/Library/PrivilegedHelperTools",
		"/private/var/db/receipts",
		"/private/tmp",
		"/usr/local/bin",
		"/usr/local/etc",
		"/usr/local/opt",
		"/usr/local/sbin",
		"/usr/local/share",
		"/usr/local/var",
	}

	// Dynamic Application Support subfolders (vendor dirs).
	supportDir := filepath.Join(home, "Library", "Application Support")
	if entries, err := os.ReadDir(supportDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				paths = append(paths, filepath.Join(supportDir, e.Name()))
			}
		}
	}

	return paths
}

// ReverseScanPaths are searched for orphaned leftovers (Pearcleaner Locations.reverse).
func ReverseScanPaths() []string {
	home := Home()
	return []string{
		filepath.Join(home, "Library", "Application Scripts"),
		filepath.Join(home, "Library", "Application Support"),
		filepath.Join(home, "Library", "Application Support", "Caches"),
		filepath.Join(home, "Library", "Containers"),
		filepath.Join(home, "Library", "Caches"),
		filepath.Join(home, "Library", "HTTPStorages"),
		filepath.Join(home, "Library", "Internet Plug-Ins"),
		filepath.Join(home, "Library", "LaunchAgents"),
		filepath.Join(home, "Library", "Logs"),
		filepath.Join(home, "Library", "Preferences"),
		filepath.Join(home, "Library", "PreferencePanes"),
		filepath.Join(home, "Library", "Preferences", "ByHost"),
		filepath.Join(home, "Library", "Saved Application State"),
		filepath.Join(home, "Library", "WebKit"),
		"/Users/Shared/Library/Application Support",
		"/Library/Application Support",
		"/Library/Caches",
		"/Library/Internet Plug-Ins",
		"/Library/LaunchAgents",
		"/Library/Logs",
		"/Library/Preferences",
	}
}
