package scan

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashwinADHD/k-cleaner/internal/appindex"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

type Scanner struct {
	apps *appindex.Index
}

func New() (*Scanner, error) {
	idx, err := appindex.Build()
	if err != nil {
		return nil, err
	}
	return &Scanner{apps: idx}, nil
}

func (s *Scanner) Apps() *appindex.Index {
	return s.apps
}

func (s *Scanner) ScanCategory(cat models.Category) models.ScanResult {
	switch cat {
	case models.CategoryBrowserCache:
		return s.scanBrowserCache()
	case models.CategorySystemCache:
		return s.scanSystemCache()
	case models.CategoryLogs:
		return s.scanLogs()
	case models.CategoryOrphanSupport:
		return s.scanOrphansByCategory(models.CategoryOrphanSupport)
	case models.CategoryOrphanPrefs:
		return s.scanOrphansByCategory(models.CategoryOrphanPrefs)
	case models.CategoryOrphanContainers:
		return s.scanOrphansByCategory(models.CategoryOrphanContainers)
	case models.CategoryOrphanSavedState:
		return s.scanOrphansByCategory(models.CategoryOrphanSavedState)
	case models.CategoryOrphanLaunchAgents:
		return s.scanOrphansByCategory(models.CategoryOrphanLaunchAgents)
	case models.CategoryTrash:
		return s.scanTrash()
	default:
		return models.ScanResult{Category: cat}
	}
}

func (s *Scanner) ScanAll() []models.ScanResult {
	results := make([]models.ScanResult, 0, len(models.AllCategories()))
	for _, cat := range models.AllCategories() {
		results = append(results, s.ScanCategory(cat))
	}
	return results
}

func (s *Scanner) ScanOrphans(level models.Sensitivity) models.ScanResult {
	rs := NewReverseScanner(s.apps, level)
	items := rs.FindOrphans()
	var total int64
	for _, item := range items {
		total += item.Size
	}
	return models.ScanResult{
		Category: models.CategoryOrphanSupport,
		Items:    items,
		TotalSize: total,
	}
}

func (s *Scanner) scanOrphansByCategory(want models.Category) models.ScanResult {
	rs := NewReverseScanner(s.apps, models.SensitivityEnhanced)
	all := rs.FindOrphans()
	var items []models.Item
	var total int64
	for _, item := range all {
		if item.Category == want {
			items = append(items, item)
			total += item.Size
		}
	}
	return models.ScanResult{Category: want, Items: items, TotalSize: total}
}

func (s *Scanner) scanBrowserCache() models.ScanResult {
	result := models.ScanResult{Category: models.CategoryBrowserCache}
	for _, bc := range paths.BrowserCachePaths() {
		item, ok := dirItem(bc.Path, models.CategoryBrowserCache, bc.Name+" browser cache")
		if ok {
			result.Items = append(result.Items, item)
			result.TotalSize += item.Size
		}
	}
	return result
}

func (s *Scanner) scanSystemCache() models.ScanResult {
	result := models.ScanResult{Category: models.CategorySystemCache}
	cacheDir := paths.Caches()
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	for _, entry := range entries {
		name := entry.Name()
		if paths.ProtectedCacheNames[name] || isBrowserCacheName(name) {
			continue
		}
		fullPath := filepath.Join(cacheDir, name)
		item, ok := dirItem(fullPath, models.CategorySystemCache, "Application cache data")
		if ok {
			result.Items = append(result.Items, item)
			result.TotalSize += item.Size
		}
	}
	return result
}

func isBrowserCacheName(name string) bool {
	prefixes := []string{
		"com.apple.Safari", "com.apple.WebKit", "Google", "Firefox",
		"org.mozilla", "com.microsoft.edgemac", "com.brave", "com.opera",
		"company.thebrowser", "com.vivaldi",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func (s *Scanner) scanLogs() models.ScanResult {
	result := models.ScanResult{Category: models.CategoryLogs}
	logDir := paths.Logs()
	item, ok := dirItem(logDir, models.CategoryLogs, "Application and system logs")
	if ok {
		result.Items = append(result.Items, item)
		result.TotalSize += item.Size
	}
	return result
}

func (s *Scanner) scanTrash() models.ScanResult {
	result := models.ScanResult{Category: models.CategoryTrash}
	trashDir := paths.Trash()
	item, ok := dirItem(trashDir, models.CategoryTrash, "Items in Trash")
	if ok && item.Size > 0 {
		result.Items = append(result.Items, item)
		result.TotalSize += item.Size
	}
	return result
}

func dirItem(path string, cat models.Category, reason string) (models.Item, bool) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return models.Item{}, false
	}
	size, _ := dirSize(path)
	if size == 0 {
		return models.Item{}, false
	}
	return models.Item{
		Path: path, Size: size, Category: cat, Reason: reason, ModifiedAt: info.ModTime(),
	}, true
}

func SkipRecent(mod time.Time, within time.Duration) bool {
	return time.Since(mod) < within
}
