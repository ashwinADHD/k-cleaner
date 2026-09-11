package scan

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashwinADHD/k-cleaner/internal/appindex"
	"github.com/ashwinADHD/k-cleaner/internal/config"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

type Scanner struct {
	apps       *appindex.Index
	exclusions *config.Exclusions
}

func New() (*Scanner, error) {
	idx, err := appindex.Build()
	if err != nil {
		return nil, err
	}
	ex, err := config.LoadExclusions()
	if err != nil {
		return nil, err
	}
	return &Scanner{apps: idx, exclusions: ex}, nil
}

func (s *Scanner) Exclusions() *config.Exclusions {
	return s.exclusions
}

func (s *Scanner) filterResult(r models.ScanResult) models.ScanResult {
	if s.exclusions == nil {
		return r
	}
	r.Items = s.exclusions.FilterItems(r.Items)
	r.TotalSize = 0
	for _, item := range r.Items {
		r.TotalSize += item.Size
	}
	return r
}

func (s *Scanner) Apps() *appindex.Index {
	return s.apps
}

func (s *Scanner) ScanCategory(cat models.Category) models.ScanResult {
	var r models.ScanResult
	switch cat {
	case models.CategoryBrowserCache:
		r = s.scanBrowserCache()
	case models.CategorySystemCache:
		r = s.scanSystemCache()
	case models.CategoryLogs:
		r = s.scanLogs()
	case models.CategoryOrphanSupport:
		r = s.scanOrphansByCategory(models.CategoryOrphanSupport)
	case models.CategoryOrphanPrefs:
		r = s.scanOrphansByCategory(models.CategoryOrphanPrefs)
	case models.CategoryOrphanContainers:
		r = s.scanOrphansByCategory(models.CategoryOrphanContainers)
	case models.CategoryOrphanSavedState:
		r = s.scanOrphansByCategory(models.CategoryOrphanSavedState)
	case models.CategoryOrphanLaunchAgents:
		r = s.scanOrphansByCategory(models.CategoryOrphanLaunchAgents)
	case models.CategoryTrash:
		r = s.scanTrash()
	default:
		return models.ScanResult{Category: cat}
	}
	return s.filterResult(r)
}

func (s *Scanner) ScanAll() []models.ScanResult {
	results := make([]models.ScanResult, 0, len(models.AllCategories()))
	for _, cat := range models.AllCategories() {
		results = append(results, s.ScanCategory(cat))
	}
	return results
}

func (s *Scanner) ScanOrphans(level models.Sensitivity) models.ScanResult {
	rs := &ReverseScanner{Apps: s.apps, Sensitivity: level, Exclusions: s.exclusions}
	items := rs.FindOrphans()
	var total int64
	for _, item := range items {
		total += item.Size
	}
	return models.ScanResult{
		Category:  models.CategoryOrphanSupport,
		Items:     items,
		TotalSize: total,
	}
}

func (s *Scanner) scanOrphansByCategory(want models.Category) models.ScanResult {
	rs := &ReverseScanner{Apps: s.apps, Sensitivity: models.SensitivityEnhanced, Exclusions: s.exclusions}
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
