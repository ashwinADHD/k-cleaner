package match

import (
	"strings"
	"unicode"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

// PearFormat strips non-alphanumeric characters and lowercases (Pearcleaner pearFormat).
func PearFormat(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	out := strings.ToLower(b.String())
	if out == "" {
		return strings.ToLower(s)
	}
	return out
}

const minMatchLen = 5

// MatchesApp reports whether an item name/path relates to an installed app identifier.
func MatchesApp(itemName, itemPath string, bundleID, appName string, entitlements []string, level models.Sensitivity) bool {
	nName := PearFormat(itemName)
	nPath := PearFormat(itemPath)
	nBundle := PearFormat(bundleID)
	nApp := PearFormat(appName)

	if nBundle != "" && len(nBundle) >= minMatchLen {
		if nName == nBundle || strings.Contains(nName, nBundle) || strings.Contains(nPath, nBundle) {
			return true
		}
	}

	if nApp != "" && len(nApp) >= minMatchLen {
		if nName == nApp || strings.Contains(nName, nApp) || strings.Contains(nPath, nApp) {
			return true
		}
	}

	if level == models.SensitivityStrict {
		return false
	}

	// Enhanced: partial match on stripped app name (drop version numbers).
	stripped := stripVersion(nApp)
	if stripped != "" && len(stripped) >= minMatchLen {
		if strings.Contains(nName, stripped) || strings.Contains(nPath, stripped) {
			return true
		}
	}

	if level == models.SensitivityEnhanced {
		return false
	}

	// Deep: entitlement-based matching.
	for _, ent := range entitlements {
		nEnt := PearFormat(ent)
		if len(nEnt) >= minMatchLen && (strings.Contains(nName, nEnt) || strings.Contains(nPath, nEnt)) {
			return true
		}
	}

	return false
}

// MatchesAnyApp checks if an item relates to any app in the inventory.
func MatchesAnyApp(itemName, itemPath string, apps []AppIdentifiers, level models.Sensitivity) bool {
	for _, app := range apps {
		if MatchesApp(itemName, itemPath, app.BundleID, app.AppName, app.Entitlements, level) {
			return true
		}
	}
	return false
}

type AppIdentifiers struct {
	BundleID     string
	AppName      string
	Entitlements []string
}

func stripVersion(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return name
	}
	last := parts[len(parts)-1]
	if len(last) > 0 && unicode.IsDigit(rune(last[0])) {
		return strings.Join(parts[:len(parts)-1], "")
	}
	return name
}
