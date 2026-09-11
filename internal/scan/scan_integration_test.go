package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ashwinADHD/k-cleaner/internal/appindex"
	"github.com/ashwinADHD/k-cleaner/internal/config"
	"github.com/ashwinADHD/k-cleaner/internal/match"
	"github.com/ashwinADHD/k-cleaner/internal/models"
)

func setupTestLibrary(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	dirs := []string{
		"Library/Application Support/com.installed.app",
		"Library/Application Support/com.orphan.removed",
		"Library/Preferences",
		"Library/Caches/com.installed.app",
		"Library/Caches/com.orphan.removed",
		"Library/Containers/com.orphan.removed",
		"Library/LaunchAgents",
		"Library/Saved Application State",
		"Library/Logs",
		"Library/WebKit",
	}
	for _, d := range dirs {
		p := filepath.Join(home, d)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, "data.bin"), []byte("testdata"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(home, "Library/Preferences/com.orphan.removed.plist"),
		[]byte("plist"), 0o644); err != nil {
		t.Fatal(err)
	}

	return home
}

func testInstalledIndex() *appindex.Index {
	return &appindex.Index{
		BundleIDs: map[string]bool{"com.installed.app": true},
		AppNames:  map[string]bool{"InstalledApp": true},
		Formatted: []match.AppIdentifiers{
			{BundleID: "com.installed.app", AppName: "InstalledApp"},
		},
	}
}

func TestReverseScannerFindsOrphans(t *testing.T) {
	setupTestLibrary(t)
	rs := &ReverseScanner{
		Apps:        testInstalledIndex(),
		Sensitivity: models.SensitivityEnhanced,
	}
	items := rs.FindOrphans()

	found := map[string]bool{}
	for _, item := range items {
		found[filepath.Base(item.Path)] = true
	}

	if found["com.installed.app"] {
		t.Error("installed app support folder should not be orphan")
	}
	if !found["com.orphan.removed"] {
		t.Error("expected orphan support folder com.orphan.removed")
	}
	if !found["com.orphan.removed.plist"] {
		t.Error("expected orphan preference file")
	}
}

func TestForwardScannerFindsRelatedFiles(t *testing.T) {
	home := setupTestLibrary(t)
	app := &models.AppInfo{
		Path:     filepath.Join(home, "Applications", "InstalledApp.app"),
		Name:     "InstalledApp",
		BundleID: "com.installed.app",
	}

	fs := NewForwardScanner(app, models.SensitivityStrict)
	items := fs.FindRelated()

	cacheFound := false
	for _, item := range items {
		if filepath.Base(item.Path) == "com.installed.app" &&
			filepath.Base(filepath.Dir(item.Path)) == "Caches" {
			cacheFound = true
		}
	}
	if !cacheFound {
		t.Error("expected forward scan to find app cache folder")
	}
}

func TestForwardScannerSpotlightIntegration(t *testing.T) {
	home := setupTestLibrary(t)
	spotlightPath := filepath.Join(home, "Library", "Caches", "com.installed.app.spotlight")
	if err := os.MkdirAll(spotlightPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(spotlightPath, "extra.dat"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := &models.AppInfo{
		Name: "InstalledApp", BundleID: "com.installed.app",
	}
	fs := &ForwardScanner{
		App: app, Sensitivity: models.SensitivityEnhanced,
		Spotlight: &SpotlightSearcher{
			Run: func(ctx context.Context, query string) ([]string, error) {
				return []string{spotlightPath}, nil
			},
		},
	}
	items := fs.FindRelated()
	found := false
	for _, item := range items {
		if item.Path == spotlightPath {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected spotlight-supplemented path in results")
	}
}

func TestReverseScannerRespectsExclusions(t *testing.T) {
	home := setupTestLibrary(t)
	orphanPath := filepath.Join(home, "Library", "Application Support", "com.orphan.removed")

	ex := config.NewExclusions([]string{orphanPath}, filepath.Join(home, "Library", "Application Support", "K-Cleaner", "exclusions.json"))

	rs := &ReverseScanner{
		Apps: testInstalledIndex(), Sensitivity: models.SensitivityEnhanced,
		Exclusions: ex,
	}
	items := rs.FindOrphans()
	for _, item := range items {
		if item.Path == orphanPath {
			t.Error("excluded orphan path should not appear in results")
		}
	}
}
