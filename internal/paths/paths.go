package paths

import (
	"os"
	"os/user"
	"path/filepath"
)

func Home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir
}

func Library() string  { return filepath.Join(Home(), "Library") }
func Caches() string   { return filepath.Join(Library(), "Caches") }
func ApplicationSupport() string {
	return filepath.Join(Library(), "Application Support")
}
func Preferences() string { return filepath.Join(Library(), "Preferences") }
func Containers() string  { return filepath.Join(Library(), "Containers") }
func GroupContainers() string {
	return filepath.Join(Library(), "Group Containers")
}
func SavedApplicationState() string {
	return filepath.Join(Library(), "Saved Application State")
}
func Logs() string        { return filepath.Join(Library(), "Logs") }
func LaunchAgents() string { return filepath.Join(Library(), "LaunchAgents") }
func Trash() string       { return filepath.Join(Home(), ".Trash") }
func WebKit() string      { return filepath.Join(Library(), "WebKit") }
func HTTPStorages() string {
	return filepath.Join(Library(), "HTTPStorages")
}

func ApplicationDirs() []string {
	home := Home()
	return []string{
		"/Applications",
		"/System/Applications",
		"/Applications/Utilities",
		filepath.Join(home, "Applications"),
	}
}

func BrowserCachePaths() []struct {
	Name string
	Path string
} {
	caches := Caches()
	return []struct {
		Name string
		Path string
	}{
		{"Safari", filepath.Join(caches, "com.apple.Safari")},
		{"Safari WebKit", filepath.Join(caches, "com.apple.WebKit.Networking")},
		{"Chrome", filepath.Join(caches, "Google", "Chrome")},
		{"Chrome Canary", filepath.Join(caches, "Google", "Chrome Canary")},
		{"Firefox", filepath.Join(caches, "Firefox")},
		{"Firefox (Mozilla)", filepath.Join(caches, "org.mozilla.firefox")},
		{"Edge", filepath.Join(caches, "com.microsoft.edgemac")},
		{"Brave", filepath.Join(caches, "com.brave.Browser")},
		{"Opera", filepath.Join(caches, "com.operasoftware.Opera")},
		{"Arc", filepath.Join(caches, "company.thebrowser.Browser")},
		{"Vivaldi", filepath.Join(caches, "com.vivaldi.Vivaldi")},
	}
}

var ProtectedCacheNames = map[string]bool{
	"com.apple.bird":          true,
	"com.apple.nsurlsessiond": true,
	"CloudKit":                true,
	"com.apple.HomeKit":       true,
	"com.apple.Safari":        true,
}

var BlockedDeleteRoots = []string{
	"/",
	"/Applications",
	"/System",
	"/Library",
	"/Users",
	"/private",
}
