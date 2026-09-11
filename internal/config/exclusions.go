package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

const appSupportDir = "K-Cleaner"
const exclusionsFile = "exclusions.json"

// Exclusions holds user-defined paths that scans and cleans must skip.
type Exclusions struct {
	Paths []string `json:"paths"`
	path  string   // absolute path to config file
}

type filePayload struct {
	Paths []string `json:"paths"`
}

// NewExclusions creates an exclusion list with an optional config file path (for tests).
func NewExclusions(paths []string, configPath string) *Exclusions {
	return &Exclusions{Paths: normalizePaths(paths), path: configPath}
}

// ConfigDir returns ~/Library/Application Support/K-Cleaner.
func ConfigDir() string {
	return filepath.Join(paths.ApplicationSupport(), appSupportDir)
}

func exclusionsPath() string {
	return filepath.Join(ConfigDir(), exclusionsFile)
}

// Load reads exclusions from disk, or returns empty list if missing.
func LoadExclusions() (*Exclusions, error) {
	p := exclusionsPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Exclusions{Paths: []string{}, path: p}, nil
		}
		return nil, err
	}
	var payload filePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &Exclusions{Paths: normalizePaths(payload.Paths), path: p}, nil
}

// Save persists exclusions to disk.
func (e *Exclusions) Save() error {
	if err := os.MkdirAll(filepath.Dir(e.path), 0o755); err != nil {
		return err
	}
	payload := filePayload{Paths: e.Paths}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(e.path, data, 0o644)
}

// FilePath returns the on-disk config location.
func (e *Exclusions) FilePath() string {
	return e.path
}

// Add appends a path if not already present.
func (e *Exclusions) Add(raw string) error {
	abs, err := absPath(raw)
	if err != nil {
		return err
	}
	for _, p := range e.Paths {
		if p == abs {
			return nil
		}
	}
	e.Paths = append(e.Paths, abs)
	return e.Save()
}

// Remove deletes a path from the exclusion list.
func (e *Exclusions) Remove(raw string) error {
	abs, err := absPath(raw)
	if err != nil {
		return err
	}
	var kept []string
	removed := false
	for _, p := range e.Paths {
		if p == abs {
			removed = true
			continue
		}
		kept = append(kept, p)
	}
	if !removed {
		return os.ErrNotExist
	}
	e.Paths = kept
	return e.Save()
}

// IsExcluded reports whether target is excluded or inside an excluded directory.
func (e *Exclusions) IsExcluded(target string) bool {
	abs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	abs = filepath.Clean(abs)
	for _, ex := range e.Paths {
		ex = filepath.Clean(ex)
		if abs == ex {
			return true
		}
		if strings.HasPrefix(abs, ex+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// FilterItems removes excluded paths from a slice.
func (e *Exclusions) FilterItems(items []models.Item) []models.Item {
	if e == nil || len(e.Paths) == 0 {
		return items
	}
	out := make([]models.Item, 0, len(items))
	for _, item := range items {
		if !e.IsExcluded(item.Path) {
			out = append(out, item)
		}
	}
	return out
}

func absPath(raw string) (string, error) {
	if raw == "" {
		return "", os.ErrInvalid
	}
	if strings.HasPrefix(raw, "~") {
		raw = filepath.Join(paths.Home(), strings.TrimPrefix(raw, "~"))
	}
	return filepath.Abs(raw)
}

func normalizePaths(in []string) []string {
	var out []string
	seen := make(map[string]bool)
	for _, p := range in {
		abs, err := absPath(p)
		if err != nil {
			continue
		}
		if !seen[abs] {
			seen[abs] = true
			out = append(out, abs)
		}
	}
	return out
}
