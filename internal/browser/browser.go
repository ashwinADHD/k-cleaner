package browser

import (
	"os/exec"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

type Info struct {
	Name        string
	ProcessName string
}

var catalog = []Info{
	{Name: "Safari", ProcessName: "Safari"},
	{Name: "Chrome", ProcessName: "Google Chrome"},
	{Name: "Chrome Canary", ProcessName: "Google Chrome Canary"},
	{Name: "Firefox", ProcessName: "firefox"},
	{Name: "Edge", ProcessName: "Microsoft Edge"},
	{Name: "Brave", ProcessName: "Brave Browser"},
	{Name: "Opera", ProcessName: "Opera"},
	{Name: "Arc", ProcessName: "Arc"},
	{Name: "Vivaldi", ProcessName: "Vivaldi"},
}

func Running() []string {
	var open []string
	seen := make(map[string]bool)
	for _, b := range catalog {
		if isProcessRunning(b.ProcessName) && !seen[b.Name] {
			open = append(open, b.Name)
			seen[b.Name] = true
		}
	}
	if !seen["Firefox"] && isProcessRunning("Firefox") {
		open = append(open, "Firefox")
	}
	return open
}

func BrowserForCachePath(cachePath string) string {
	cachePath = strings.ToLower(cachePath)
	for _, bc := range paths.BrowserCachePaths() {
		if strings.EqualFold(cachePath, bc.Path) || strings.HasPrefix(cachePath, strings.ToLower(bc.Path)+"/") {
			name := bc.Name
			if strings.Contains(name, "Safari") {
				return "Safari"
			}
			if strings.Contains(name, "Chrome") {
				if strings.Contains(name, "Canary") {
					return "Chrome Canary"
				}
				return "Chrome"
			}
			if strings.Contains(name, "Firefox") {
				return "Firefox"
			}
			return name
		}
	}
	return ""
}

func AffectedBrowsers(items []models.Item) []string {
	running := make(map[string]bool)
	for _, name := range Running() {
		running[name] = true
	}

	var affected []string
	seen := make(map[string]bool)
	for _, item := range items {
		if item.Category != models.CategoryBrowserCache {
			continue
		}
		browser := BrowserForCachePath(item.Path)
		if browser == "" {
			continue
		}
		if running[browser] && !seen[browser] {
			affected = append(affected, browser)
			seen[browser] = true
		}
	}
	return affected
}

func WarningMessage(browsers []string) string {
	if len(browsers) == 0 {
		return ""
	}
	names := strings.Join(browsers, ", ")
	return "These browsers are currently open: " + names + ".\n\n" +
		"Clearing their cache while running can cause tabs to reload, lose unsaved form data, " +
		"or make the browser unstable.\n\n" +
		"Quit the browsers first for the safest cleanup, or skip browser cache items."
}

func isProcessRunning(name string) bool {
	return exec.Command("pgrep", "-x", name).Run() == nil
}
