package sentinel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

func TestIsUnderTrash(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	trash := filepath.Join(home, ".Trash")
	if err := os.MkdirAll(trash, 0o755); err != nil {
		t.Fatal(err)
	}
	app := filepath.Join(trash, "MyApp.app")
	if !isUnderTrash(app) {
		t.Error("expected app in trash to match")
	}
	if isUnderTrash("/Applications/MyApp.app") {
		t.Error("expected app outside trash to not match")
	}
	_ = paths.Trash()
}
