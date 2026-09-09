package clean

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

func TestIsSafePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	library := filepath.Join(home, "Library", "Caches", "com.example.app")
	if err := os.MkdirAll(library, 0o755); err != nil {
		t.Fatal(err)
	}

	if !isSafePath(library) {
		t.Error("Library cache path should be safe")
	}
	if isSafePath("/") {
		t.Error("root should not be safe")
	}
	if isSafePath(home) {
		t.Error("home root should not be safe")
	}
}

func TestCleanerDryRun(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	target := filepath.Join(home, "Library", "Caches", "testapp")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "cache.dat"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cl := &Cleaner{DryRun: true}
	cr := cl.CleanItems([]models.Item{
		{Path: target, Size: 4, Category: models.CategorySystemCache},
	})

	if cr.Deleted != 1 {
		t.Errorf("expected 1 deleted in dry run, got %d", cr.Deleted)
	}
	if _, err := os.Stat(target); err != nil {
		t.Error("dry run should not remove files")
	}
}
