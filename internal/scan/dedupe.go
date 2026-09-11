package scan

import (
	"sort"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

// DedupeParentPaths keeps the shortest path when a parent directory is also matched.
// Returns a stable, sorted list (Pearcleaner finalizeCollection behavior).
func DedupeParentPaths(found map[string]models.Item) []models.Item {
	if len(found) == 0 {
		return nil
	}

	pathList := make([]string, 0, len(found))
	for p := range found {
		pathList = append(pathList, cleanPath(p))
	}
	sort.Strings(pathList)

	kept := make(map[string]bool)
	for i, p := range pathList {
		if kept[p] {
			continue
		}
		isChild := false
		for j := 0; j < i; j++ {
			parent := pathList[j]
			if parent != p && isSubpath(p, parent) {
				isChild = true
				break
			}
		}
		if !isChild {
			kept[p] = true
		}
	}

	items := make([]models.Item, 0, len(kept))
	for p := range kept {
		// Prefer original map key item (may differ only by Clean).
		item, ok := found[p]
		if !ok {
			for k, v := range found {
				if cleanPath(k) == p {
					item = v
					ok = true
					break
				}
			}
		}
		if ok {
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Path < items[j].Path
	})
	return items
}

func cleanPath(p string) string {
	return strings.TrimRight(p, "/")
}

func isSubpath(child, parent string) bool {
	if parent == "" || child == parent {
		return false
	}
	return strings.HasPrefix(child, parent+"/")
}
