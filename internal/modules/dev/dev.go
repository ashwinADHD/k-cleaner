package dev

import (
	"os"
	"path/filepath"

	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

type cacheTarget struct {
	Tool string
	Path string
}

// CacheTargets are developer tool cache locations (Pearcleaner PathLibrary subset).
func CacheTargets() []cacheTarget {
	home := paths.Home()
	lib := paths.Library()
	return []cacheTarget{
		{"npm", filepath.Join(home, ".npm")},
		{"npm", filepath.Join(home, ".cache", "yarn")},
		{"Yarn", filepath.Join(home, ".yarn-cache")},
		{"Cargo", filepath.Join(home, ".cargo", "registry")},
		{"Cargo", filepath.Join(home, ".cargo", "git")},
		{"pip", filepath.Join(lib, "Caches", "pip")},
		{"Poetry", filepath.Join(lib, "Caches", "pypoetry")},
		{"Gradle", filepath.Join(home, ".gradle", "caches")},
		{"Maven", filepath.Join(home, ".m2", "repository")},
		{"Go modules", filepath.Join(home, "go", "pkg", "mod")},
		{"Deno", filepath.Join(lib, "Caches", "deno")},
		{"Swift PM", filepath.Join(home, ".swiftpm")},
		{"VS Code", filepath.Join(lib, "Application Support", "Code", "Cache")},
		{"VS Code", filepath.Join(lib, "Application Support", "Code", "CachedData")},
		{"VS Code", filepath.Join(lib, "Application Support", "Code", "CachedExtensions")},
		{"VS Code", filepath.Join(lib, "Application Support", "Code", "Code Cache")},
		{"Cursor", filepath.Join(lib, "Application Support", "Cursor", "Cache")},
		{"Cursor", filepath.Join(lib, "Application Support", "Cursor", "CachedData")},
		{"Cursor", filepath.Join(home, ".cursor", "extensions")},
		{"Xcode DerivedData", filepath.Join(home, "Library", "Developer", "Xcode", "DerivedData")},
		{"Xcode Archives", filepath.Join(home, "Library", "Developer", "Xcode", "Archives")},
		{"Xcode DeviceSupport", filepath.Join(home, "Library", "Developer", "Xcode", "iOS DeviceSupport")},
		{"CoreSimulator", filepath.Join(home, "Library", "Developer", "CoreSimulator", "Devices")},
		{"Xcode Caches", filepath.Join(lib, "Caches", "com.apple.dt.Xcode")},
		{"CocoaPods", filepath.Join(lib, "Caches", "CocoaPods")},
		{"pnpm", filepath.Join(home, "Library", "pnpm", "store")},
	}
}

// Scan finds existing developer cache directories with reclaimable size.
func Scan() []models.Item {
	var items []models.Item
	seen := make(map[string]bool)
	for _, target := range CacheTargets() {
		if seen[target.Path] {
			continue
		}
		seen[target.Path] = true
		item, ok := scanPath(target.Path, target.Tool)
		if ok {
			items = append(items, item)
		}
	}
	return items
}

func scanPath(p, tool string) (models.Item, bool) {
	info, err := os.Stat(p)
	if err != nil {
		return models.Item{}, false
	}
	if !info.IsDir() {
		return models.Item{}, false
	}
	size := dirSize(p)
	if size == 0 {
		return models.Item{}, false
	}
	return models.Item{
		Path: p, Size: size, Category: models.CategoryDevCaches,
		Reason: tool + " cache", ModifiedAt: info.ModTime(),
	}, true
}

func dirSize(root string) int64 {
	var total int64
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}
