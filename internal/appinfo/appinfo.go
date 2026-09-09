package appinfo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}$`)

// FromPath reads metadata from a .app bundle.
func FromPath(appPath string) (*models.AppInfo, error) {
	appPath = strings.TrimSpace(appPath)
	if !strings.HasSuffix(appPath, ".app") {
		appPath += ".app"
	}

	info, err := os.Stat(appPath)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not an application bundle: %s", appPath)
	}

	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	app := &models.AppInfo{
		Path:     appPath,
		Name:     strings.TrimSuffix(filepath.Base(appPath), ".app"),
		BundleID: plutilExtract(plistPath, "CFBundleIdentifier"),
		Version:  plutilExtract(plistPath, "CFBundleShortVersionString"),
	}

	if display := plutilExtract(plistPath, "CFBundleDisplayName"); display != "" {
		app.Name = display
	} else if name := plutilExtract(plistPath, "CFBundleName"); name != "" {
		app.Name = name
	}

	app.Executable = plutilExtract(plistPath, "CFBundleExecutable")
	app.Entitlements = extractEntitlements(appPath)
	app.TeamID = extractTeamID(appPath)

	if app.BundleID == "" {
		return nil, fmt.Errorf("could not read bundle identifier from %s", appPath)
	}

	return app, nil
}

func plutilExtract(plistPath, key string) string {
	out, err := exec.Command("plutil", "-extract", key, "raw", "-o", "-", plistPath).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func extractEntitlements(appPath string) []string {
	out, err := exec.Command("codesign", "-d", "--entitlements", ":-", appPath).CombinedOutput()
	if err != nil {
		return nil
	}
	var ents []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "<key>") {
			start := strings.Index(line, "<key>") + 5
			end := strings.Index(line, "</key>")
			if end > start {
				key := strings.TrimSpace(line[start:end])
				if key != "" && !strings.HasPrefix(key, "com.apple.security") {
					ents = append(ents, key)
				}
			}
		}
	}
	return ents
}

func extractTeamID(appPath string) string {
	out, err := exec.Command("codesign", "-dv", appPath).CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "TeamIdentifier=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// ContainerBundleID resolves a UUID-named container folder to its bundle ID.
func ContainerBundleID(containerPath string) string {
	name := filepath.Base(containerPath)
	if !uuidPattern.MatchString(name) {
		return name // already a bundle ID folder
	}

	metaPath := filepath.Join(containerPath, ".com.apple.containermanagerd.metadata.plist")
	id := plutilExtract(metaPath, "MCMMetadataIdentifier")
	if id != "" {
		return id
	}
	return ""
}

// KillRunningApp terminates the app if it is running.
func KillRunningApp(app *models.AppInfo) error {
	if app.Executable == "" {
		return nil
	}
	return exec.Command("pkill", "-x", app.Executable).Run()
}
