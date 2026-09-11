package scan

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/appinfo"
	"github.com/ashwinADHD/k-cleaner/internal/conditions"
	"github.com/ashwinADHD/k-cleaner/internal/config"
	"github.com/ashwinADHD/k-cleaner/internal/match"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

// ForwardScanner finds files related to a specific app (Pearcleaner AppPathFinder).
type ForwardScanner struct {
	App         *models.AppInfo
	Sensitivity models.Sensitivity
	Exclusions  *config.Exclusions
	Spotlight   *SpotlightSearcher
}

func NewForwardScanner(app *models.AppInfo, level models.Sensitivity) *ForwardScanner {
	return &ForwardScanner{App: app, Sensitivity: level, Spotlight: &SpotlightSearcher{}}
}

func (f *ForwardScanner) FindRelated() []models.Item {
	found := make(map[string]models.Item)

	for _, root := range paths.ForwardScanPaths() {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		f.scanLocation(root, found)
	}

	f.scanContainers(found)
	f.scanSpotlight(found)

	items := DedupeParentPaths(found)
	if f.Exclusions != nil {
		items = f.Exclusions.FilterItems(items)
	}
	return items
}

func (f *ForwardScanner) scanSpotlight(found map[string]models.Item) {
	if f.Spotlight == nil || f.App == nil {
		return
	}
	for _, p := range f.Spotlight.find(f.App, f.Sensitivity) {
		if f.Exclusions != nil && f.Exclusions.IsExcluded(p) {
			continue
		}
		name := filepath.Base(p)
		if !f.matchesItem(name, p) {
			continue
		}
		item, ok := makeItem(p, models.CategoryAppRelated, "Spotlight: related to "+f.App.Name)
		if ok {
			found[p] = item
		}
	}
}

func (f *ForwardScanner) scanLocation(root string, found map[string]models.Item) {
	depth := scanDepth(root)

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}
		depthLevel := strings.Count(rel, string(os.PathSeparator)) + 1
		if depthLevel > depth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		if strings.HasPrefix(name, ".") && name != ".com.apple.containermanagerd.metadata.plist" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if f.matchesItem(name, path) {
			item, ok := makeItem(path, models.CategoryAppRelated, "Related to "+f.App.Name)
			if ok {
				found[path] = item
			}
			if d.IsDir() {
				return filepath.SkipDir
			}
		}
		return nil
	})
}

func scanDepth(root string) int {
	base := filepath.Base(root)
	if base == "Library" || strings.HasSuffix(root, "/Library") {
		return 2
	}
	return 1
}

func (f *ForwardScanner) matchesItem(name, path string) bool {
	if conditions.MatchesConditionExclude(name, path, f.App.BundleID) {
		return false
	}
	if conditions.MatchesConditionInclude(name, path, f.App.BundleID) {
		return true
	}
	return match.MatchesApp(name, path, f.App.BundleID, f.App.Name, f.App.Entitlements, f.Sensitivity)
}

func (f *ForwardScanner) scanContainers(found map[string]models.Item) {
	containerDir := paths.Containers()
	entries, err := os.ReadDir(containerDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(containerDir, entry.Name())
		bundleID := appinfo.ContainerBundleID(fullPath)
		if bundleID == "" {
			bundleID = entry.Name()
		}
		if match.MatchesApp(entry.Name(), fullPath, f.App.BundleID, f.App.Name, f.App.Entitlements, f.Sensitivity) ||
			match.PearFormat(bundleID) == match.PearFormat(f.App.BundleID) {
			item, ok := makeItem(fullPath, models.CategoryAppRelated, "Sandbox container for "+f.App.Name)
			if ok {
				found[fullPath] = item
			}
		}
	}
}

func makeItem(path string, cat models.Category, reason string) (models.Item, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return models.Item{}, false
	}
	var size int64
	if info.IsDir() {
		size, _ = dirSize(path)
	} else {
		size = info.Size()
	}
	if size == 0 && !info.IsDir() {
		return models.Item{}, false
	}
	return models.Item{
		Path: path, Size: size, Category: cat, Reason: reason, ModifiedAt: info.ModTime(),
	}, true
}
