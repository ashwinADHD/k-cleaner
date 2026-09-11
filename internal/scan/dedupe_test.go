package scan

import (
	"testing"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

func TestDedupeParentPaths(t *testing.T) {
	found := map[string]models.Item{
		"/tmp/Library/Caches/com.foo.app": {
			Path: "/tmp/Library/Caches/com.foo.app", Size: 100,
		},
		"/tmp/Library/Caches/com.foo.app/sub": {
			Path: "/tmp/Library/Caches/com.foo.app/sub", Size: 50,
		},
		"/tmp/Library/Preferences/com.foo.app.plist": {
			Path: "/tmp/Library/Preferences/com.foo.app.plist", Size: 10,
		},
	}

	items := DedupeParentPaths(found)
	if len(items) != 2 {
		t.Fatalf("expected 2 items after dedupe, got %d", len(items))
	}

	for _, item := range items {
		if item.Path == "/tmp/Library/Caches/com.foo.app/sub" {
			t.Error("child path should be removed when parent is kept")
		}
	}
}

func TestDedupeParentPathsStableOrder(t *testing.T) {
	found := map[string]models.Item{
		"/b": {Path: "/b", Size: 1},
		"/a": {Path: "/a", Size: 1},
	}
	items := DedupeParentPaths(found)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Path != "/a" || items[1].Path != "/b" {
		t.Errorf("unexpected order: %v, %v", items[0].Path, items[1].Path)
	}
}

func TestIsSubpath(t *testing.T) {
	if !isSubpath("/a/b/c", "/a/b") {
		t.Error("expected subpath match")
	}
	if isSubpath("/a/b", "/a/b") {
		t.Error("same path is not a subpath")
	}
}
