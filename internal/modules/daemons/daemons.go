package daemons

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/appindex"
	"github.com/ashwinADHD/k-cleaner/internal/match"
	"github.com/ashwinADHD/k-cleaner/internal/models"
)

// Entry describes a launch agent or daemon plist (read-only inventory).
type Entry struct {
	Path       string
	Label      string
	Program    string
	Scope      string
	OrphanHint bool
	Size       int64
}

var systemScanDirs = []struct {
	rel   string
	scope string
}{
	{"Library/LaunchAgents", "User Launch Agent"},
	{"/Library/LaunchAgents", "System Launch Agent"},
	{"/Library/LaunchDaemons", "System Launch Daemon"},
}

// List scans launch agent/daemon plists (read-only — no deletion).
func List(apps *appindex.Index) []Entry {
	var entries []Entry
	identifiers := apps.Identifiers()
	home := os.Getenv("HOME")

	for _, dir := range systemScanDirs {
		path := dir.rel
		if !strings.HasPrefix(path, "/") {
			path = filepath.Join(home, path)
		}
		plistEntries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for _, e := range plistEntries {
			if !strings.HasSuffix(e.Name(), ".plist") {
				continue
			}
			fullPath := filepath.Join(path, e.Name())
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			label := plutilExtract(fullPath, "Label")
			if label == "" {
				label = strings.TrimSuffix(e.Name(), ".plist")
			}
			program := plutilExtract(fullPath, "Program")
			if program == "" {
				program = plutilExtract(fullPath, "ProgramArguments")
			}

			orphan := !match.MatchesAnyApp(label, fullPath, identifiers, models.SensitivityEnhanced) &&
				!strings.HasPrefix(label, "com.apple.")

			entries = append(entries, Entry{
				Path: fullPath, Label: label, Program: program,
				Scope: dir.scope, OrphanHint: orphan, Size: info.Size(),
			})
		}
	}
	return entries
}

func plutilExtract(plistPath, key string) string {
	out, err := exec.Command("plutil", "-extract", key, "raw", "-o", "-", plistPath).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
