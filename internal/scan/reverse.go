package scan

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/appindex"
	"github.com/ashwinADHD/k-cleaner/internal/appinfo"
	"github.com/ashwinADHD/k-cleaner/internal/conditions"
	"github.com/ashwinADHD/k-cleaner/internal/match"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

// ReverseScanner finds orphaned files not tied to any installed app.
type ReverseScanner struct {
	Apps        *appindex.Index
	Sensitivity models.Sensitivity
}

func NewReverseScanner(apps *appindex.Index, level models.Sensitivity) *ReverseScanner {
	return &ReverseScanner{Apps: apps, Sensitivity: level}
}

func (r *ReverseScanner) FindOrphans() []models.Item {
	var items []models.Item
	seen := make(map[string]bool)
	identifiers := r.Apps.Identifiers()

	for _, location := range paths.ReverseScanPaths() {
		if _, err := os.Stat(location); err != nil {
			continue
		}
		entries, err := os.ReadDir(location)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			fullPath := filepath.Join(location, name)

			if seen[fullPath] {
				continue
			}

			if conditions.ShouldSkipReverse(name, fullPath) {
				continue
			}

			// Resolve UUID containers.
			resolvedName := name
			if entry.IsDir() {
				if bundleID := appinfo.ContainerBundleID(fullPath); bundleID != "" {
					resolvedName = bundleID
				}
			}

			if match.MatchesAnyApp(resolvedName, fullPath, identifiers, r.Sensitivity) {
				continue
			}

			cat := categorizeOrphan(location, name)
			item, ok := makeItem(fullPath, cat, "No matching installed application found")
			if !ok {
				continue
			}
			items = append(items, item)
			seen[fullPath] = true
		}
	}
	return items
}

func categorizeOrphan(location, name string) models.Category {
	switch {
	case strings.Contains(location, "Preferences"):
		return models.CategoryOrphanPrefs
	case strings.Contains(location, "Containers"):
		return models.CategoryOrphanContainers
	case strings.Contains(location, "Saved Application State"):
		return models.CategoryOrphanSavedState
	case strings.Contains(location, "LaunchAgents"):
		return models.CategoryOrphanLaunchAgents
	default:
		return models.CategoryOrphanSupport
	}
}
