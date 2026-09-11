package dev

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsDevCaches(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	npmCache := filepath.Join(home, ".npm", "cache")
	if err := os.MkdirAll(npmCache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(npmCache, "data.bin"), []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	items := Scan()
	if len(items) == 0 {
		t.Fatal("expected at least one dev cache item")
	}
	found := false
	for _, item := range items {
		if item.Path == filepath.Join(home, ".npm") {
			found = true
			if item.Size == 0 {
				t.Error("expected non-zero size for npm cache")
			}
		}
	}
	if !found {
		t.Error("expected npm cache in scan results")
	}
}
