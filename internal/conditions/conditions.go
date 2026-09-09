package conditions

import (
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/match"
)

// SkipReverseKeywords block orphan matches (Pearcleaner skipReverse).
var SkipReverseKeywords = []string{
	"apple", "icloud", "coresimulator", "simulator", "xcode",
	"homekit", "safari", "mail", "photos", "music", "tv",
	"finder", "system", "cloudkit", "bird", "nsurlsessiond",
}

// SkipReverseAllowPrefixes allow matches even when keyword would block.
var SkipReverseAllowPrefixes = []string{
	"com.apple.dt.xcode",
}

type AppCondition struct {
	BundleID string
	Include  []string
	Exclude  []string
}

// Per-app rules adapted from Pearcleaner Conditions.swift (subset).
var AppConditions = []AppCondition{
	{BundleID: "com.apple.dt.Xcode", Include: []string{"comappledt", "xcode", "simulator"},
		Exclude: []string{"comrobotsandpencilsxcodesapp", "comxcodesorgxcodesapp", "xcodes", "xcodecleaner"}},
	{BundleID: "com.brave.Browser", Include: []string{"brave"}, Exclude: []string{}},
	{BundleID: "com.google.Chrome", Include: []string{"google", "chrome"}, Exclude: []string{"chromium"}},
	{BundleID: "org.mozilla.firefox", Include: []string{"firefox", "mozilla"}, Exclude: []string{"thunderbird"}},
	{BundleID: "com.microsoft.VSCode", Include: []string{"vscode", "visualstudio", "code"}, Exclude: []string{}},
	{BundleID: "com.tinyspeck.slackmacgap", Include: []string{"slack"}, Exclude: []string{}},
	{BundleID: "us.zoom.xos", Include: []string{"zoom"}, Exclude: []string{}},
	{BundleID: "com.spotify.client", Include: []string{"spotify"}, Exclude: []string{}},
	{BundleID: "com.docker.docker", Include: []string{"docker", "comdocker"}, Exclude: []string{}},
}

func ShouldSkipReverse(itemName, itemPath string) bool {
	nName := match.PearFormat(itemName)
	nPath := match.PearFormat(itemPath)

	for _, allow := range SkipReverseAllowPrefixes {
		if strings.Contains(nName, match.PearFormat(allow)) || strings.Contains(nPath, match.PearFormat(allow)) {
			return false
		}
	}

	for _, kw := range SkipReverseKeywords {
		if strings.Contains(nName, kw) || strings.Contains(nPath, kw) {
			return true
		}
	}
	return false
}

func ConditionFor(bundleID string) *AppCondition {
	n := match.PearFormat(bundleID)
	for i := range AppConditions {
		if match.PearFormat(AppConditions[i].BundleID) == n {
			return &AppConditions[i]
		}
	}
	return nil
}

func MatchesConditionInclude(itemName, itemPath, bundleID string) bool {
	cond := ConditionFor(bundleID)
	if cond == nil {
		return false
	}
	nName := match.PearFormat(itemName)
	nPath := match.PearFormat(itemPath)
	for _, inc := range cond.Include {
		if strings.Contains(nName, inc) || strings.Contains(nPath, inc) {
			return true
		}
	}
	return false
}

func MatchesConditionExclude(itemName, itemPath, bundleID string) bool {
	cond := ConditionFor(bundleID)
	if cond == nil {
		return false
	}
	nName := match.PearFormat(itemName)
	nPath := match.PearFormat(itemPath)
	for _, exc := range cond.Exclude {
		if strings.Contains(nName, exc) || strings.Contains(nPath, exc) {
			return true
		}
	}
	return false
}
