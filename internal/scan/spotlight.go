package scan

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashwinADHD/k-cleaner/internal/match"
	"github.com/ashwinADHD/k-cleaner/internal/models"
)

const spotlightTimeout = 5 * time.Second
const spotlightMaxResults = 500

// SpotlightSearcher finds related files via mdfind (Spotlight CLI).
type SpotlightSearcher struct {
	Run func(ctx context.Context, query string) ([]string, error)
}

func defaultSpotlightRun(ctx context.Context, query string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "mdfind", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

func (s *SpotlightSearcher) find(app *models.AppInfo, level models.Sensitivity) []string {
	if level == models.SensitivityStrict {
		return nil
	}
	run := s.Run
	if run == nil {
		run = defaultSpotlightRun
	}

	ctx, cancel := context.WithTimeout(context.Background(), spotlightTimeout)
	defer cancel()

	var queries []string
	if app.BundleID != "" {
		queries = append(queries, "kMDItemCFBundleIdentifier == '"+app.BundleID+"'")
	}
	if level >= models.SensitivityEnhanced && app.Name != "" {
		n := match.PearFormat(app.Name)
		if len(n) >= 5 {
			queries = append(queries, "kMDItemFSName == '*"+n+"*'cd")
		}
	}
	if level >= models.SensitivityDeep && app.BundleID != "" {
		n := match.PearFormat(app.BundleID)
		if len(n) >= 5 {
			queries = append(queries, "kMDItemFSName == '*"+n+"*'cd")
		}
	}

	seen := make(map[string]bool)
	var results []string
	for _, q := range queries {
		paths, err := run(ctx, q)
		if err != nil {
			continue
		}
		for _, p := range paths {
			p = filepath.Clean(p)
			if seen[p] || !spotlightPathAllowed(p) {
				continue
			}
			seen[p] = true
			results = append(results, p)
			if len(results) >= spotlightMaxResults {
				return results
			}
		}
	}
	return results
}

func spotlightPathAllowed(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return false
	}
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(p, home) {
		return true
	}
	allowed := []string{"/Applications/", "/Library/", "/usr/local/", "/private/tmp/"}
	for _, prefix := range allowed {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}
