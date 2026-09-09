package models

import "time"

type Category string

const (
	CategoryBrowserCache       Category = "browser-cache"
	CategorySystemCache        Category = "system-cache"
	CategoryLogs               Category = "logs"
	CategoryOrphanSupport      Category = "orphan-support"
	CategoryOrphanPrefs        Category = "orphan-preferences"
	CategoryOrphanContainers   Category = "orphan-containers"
	CategoryOrphanSavedState   Category = "orphan-saved-state"
	CategoryOrphanLaunchAgents Category = "orphan-launch-agents"
	CategoryTrash              Category = "trash"
	CategoryAppRelated         Category = "app-related"
)

func AllCategories() []Category {
	return []Category{
		CategoryBrowserCache,
		CategorySystemCache,
		CategoryLogs,
		CategoryOrphanSupport,
		CategoryOrphanPrefs,
		CategoryOrphanContainers,
		CategoryOrphanSavedState,
		CategoryOrphanLaunchAgents,
		CategoryTrash,
	}
}

func (c Category) Label() string {
	switch c {
	case CategoryBrowserCache:
		return "Browser Cache"
	case CategorySystemCache:
		return "System Cache"
	case CategoryLogs:
		return "Application Logs"
	case CategoryOrphanSupport:
		return "Orphan Application Support"
	case CategoryOrphanPrefs:
		return "Orphan Preferences"
	case CategoryOrphanContainers:
		return "Orphan Sandbox Containers"
	case CategoryOrphanSavedState:
		return "Orphan Saved Application State"
	case CategoryOrphanLaunchAgents:
		return "Orphan Launch Agents"
	case CategoryTrash:
		return "Trash"
	case CategoryAppRelated:
		return "Application Files"
	default:
		return string(c)
	}
}

func (c Category) Description() string {
	switch c {
	case CategoryBrowserCache:
		return "Cached web pages, images, and scripts from Safari, Chrome, Firefox, and other browsers"
	case CategorySystemCache:
		return "Temporary files stored by macOS and third-party applications"
	case CategoryLogs:
		return "Diagnostic and crash logs from applications"
	case CategoryOrphanSupport:
		return "Application Support folders left behind by uninstalled apps"
	case CategoryOrphanPrefs:
		return "Preference files (.plist) for apps no longer installed"
	case CategoryOrphanContainers:
		return "Sandbox containers for apps that have been removed"
	case CategoryOrphanSavedState:
		return "Window/session restore data for uninstalled applications"
	case CategoryOrphanLaunchAgents:
		return "Background startup items for apps no longer on your Mac"
	case CategoryTrash:
		return "Items currently in the Trash"
	case CategoryAppRelated:
		return "Files related to the selected application"
	default:
		return ""
	}
}

type Item struct {
	Path       string
	Size       int64
	Category   Category
	Reason     string
	ModifiedAt time.Time
}

type ScanResult struct {
	Category  Category
	Items     []Item
	TotalSize int64
	Errors    []string
}

type CleanResult struct {
	Category   Category
	Deleted    int
	Failed     int
	BytesFreed int64
	TrashPath  string
	Errors     []string
}

type AppInfo struct {
	Path            string
	Name            string
	BundleID        string
	Version         string
	Executable      string
	Entitlements    []string
	TeamID          string
}

type Sensitivity int

const (
	SensitivityStrict Sensitivity = iota
	SensitivityEnhanced
	SensitivityDeep
)

func ParseSensitivity(s string) Sensitivity {
	switch s {
	case "enhanced":
		return SensitivityEnhanced
	case "deep":
		return SensitivityDeep
	default:
		return SensitivityStrict
	}
}

func (s Sensitivity) String() string {
	switch s {
	case SensitivityEnhanced:
		return "enhanced"
	case SensitivityDeep:
		return "deep"
	default:
		return "strict"
	}
}
