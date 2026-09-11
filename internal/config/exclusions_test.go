package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

func TestExclusionsAddRemove(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ex, err := LoadExclusions()
	if err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(home, "Library", "Caches", "keep-me")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := ex.Add(target); err != nil {
		t.Fatal(err)
	}
	if !ex.IsExcluded(target) {
		t.Error("expected path to be excluded after add")
	}
	child := filepath.Join(target, "nested", "file.dat")
	if !ex.IsExcluded(child) {
		t.Error("expected child path to be excluded")
	}

	if err := ex.Remove(target); err != nil {
		t.Fatal(err)
	}
	if ex.IsExcluded(target) {
		t.Error("expected path to not be excluded after remove")
	}
}

func TestExclusionsPersistence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	ex, err := LoadExclusions()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "Library", "Logs")
	if err := ex.Add(path); err != nil {
		t.Fatal(err)
	}

	reloaded, err := LoadExclusions()
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded.IsExcluded(path) {
		t.Error("expected exclusion to persist across reload")
	}
}

func TestFilterItems(t *testing.T) {
	home := t.TempDir()
	ex := &Exclusions{Paths: []string{filepath.Join(home, "skip")}}

	items := []models.Item{
		{Path: filepath.Join(home, "skip"), Size: 1},
		{Path: filepath.Join(home, "skip", "nested"), Size: 2},
		{Path: filepath.Join(home, "keep"), Size: 3},
	}
	filtered := ex.FilterItems(items)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 item after filter, got %d", len(filtered))
	}
	if filtered[0].Path != filepath.Join(home, "keep") {
		t.Errorf("unexpected kept path: %s", filtered[0].Path)
	}
}
