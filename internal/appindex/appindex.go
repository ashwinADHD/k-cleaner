package appindex

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/appinfo"
	"github.com/ashwinADHD/k-cleaner/internal/match"
	"github.com/ashwinADHD/k-cleaner/internal/models"
)

type Index struct {
	Apps       []models.AppInfo
	BundleIDs  map[string]bool
	AppNames   map[string]bool
	Formatted  []match.AppIdentifiers
}

func Build() (*Index, error) {
	idx := &Index{
		BundleIDs: make(map[string]bool),
		AppNames:  make(map[string]bool),
	}

	dirs := []string{
		"/Applications",
		"/System/Applications",
		"/Applications/Utilities",
	}
	if home := os.Getenv("HOME"); home != "" {
		dirs = append(dirs, filepath.Join(home, "Applications"))
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
				continue
			}
			appPath := filepath.Join(dir, entry.Name())
			app, err := appinfo.FromPath(appPath)
			if err != nil {
				idx.indexBasic(appPath)
				continue
			}
			idx.addApp(*app)
		}
	}

	for _, id := range systemBundleIDs {
		idx.BundleIDs[id] = true
	}

	return idx, nil
}

var systemBundleIDs = []string{
	"com.apple.Safari", "com.apple.finder", "com.apple.systempreferences",
	"com.apple.mail", "com.apple.MobileSMS", "com.apple.Notes",
	"com.apple.Preview", "com.apple.calculator", "com.apple.iCal",
	"com.apple.dt.Xcode", "com.apple.Terminal", "com.apple.Music",
	"com.apple.Photos", "com.apple.TV",
}

func (idx *Index) indexBasic(appPath string) {
	base := strings.TrimSuffix(filepath.Base(appPath), ".app")
	idx.AppNames[base] = true
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	if id := plutilExtract(plistPath, "CFBundleIdentifier"); id != "" {
		idx.BundleIDs[id] = true
		idx.Formatted = append(idx.Formatted, match.AppIdentifiers{
			BundleID: id, AppName: base,
		})
	}
}

func (idx *Index) addApp(app models.AppInfo) {
	idx.Apps = append(idx.Apps, app)
	idx.BundleIDs[app.BundleID] = true
	idx.AppNames[app.Name] = true
	idx.Formatted = append(idx.Formatted, match.AppIdentifiers{
		BundleID: app.BundleID, AppName: app.Name, Entitlements: app.Entitlements,
	})
}

func plutilExtract(plistPath, key string) string {
	out, err := exec.Command("plutil", "-extract", key, "raw", "-o", "-", plistPath).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (idx *Index) HasBundleID(id string) bool {
	if idx.BundleIDs[id] {
		return true
	}
	for bid := range idx.BundleIDs {
		if strings.HasPrefix(id, bid+".") {
			return true
		}
	}
	return false
}

func (idx *Index) HasAppName(name string) bool {
	if idx.AppNames[name] {
		return true
	}
	lower := strings.ToLower(name)
	for n := range idx.AppNames {
		if strings.ToLower(n) == lower {
			return true
		}
	}
	return false
}

func (idx *Index) Identifiers() []match.AppIdentifiers {
	return idx.Formatted
}

// InstalledApps returns user-facing apps sorted by name (excludes system-only entries).
func (idx *Index) InstalledApps() []models.AppInfo {
	apps := make([]models.AppInfo, len(idx.Apps))
	copy(apps, idx.Apps)
	sort.Slice(apps, func(i, j int) bool {
		return strings.ToLower(apps[i].Name) < strings.ToLower(apps[j].Name)
	})
	return apps
}
