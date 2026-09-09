package permissions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

type Status struct {
	FullDiskAccess bool
	Details        []string
}

// CheckFullDiskAccess probes TCC-protected paths (Pearcleaner PermissionsSheetView pattern).
func CheckFullDiskAccess() Status {
	st := Status{Details: []string{}}

	probes := []struct {
		path string
		desc string
	}{
		{filepath.Join(paths.Containers(), "com.apple.stocks"), "Containers (sandbox data)"},
		{filepath.Join(paths.Home(), "Library", "Safari", "Bookmarks.plist"), "Safari data"},
		{filepath.Join(paths.Home(), "Library", "Mail"), "Mail data"},
	}

	allOK := true
	for _, p := range probes {
		if _, err := os.Stat(p.path); err != nil {
			if os.IsNotExist(err) {
				st.Details = append(st.Details, fmt.Sprintf("  • %s — path not found (may be OK)", p.desc))
				continue
			}
			allOK = false
			st.Details = append(st.Details, fmt.Sprintf("  • %s — access denied", p.desc))
		} else {
			st.Details = append(st.Details, fmt.Sprintf("  • %s — accessible", p.desc))
		}
	}

	// Try reading a container directory listing.
	containerDir := paths.Containers()
	if entries, err := os.ReadDir(containerDir); err != nil {
		allOK = false
		st.Details = append(st.Details, "  • Container listing — access denied (Full Disk Access likely required)")
	} else if len(entries) > 0 {
		st.Details = append(st.Details, fmt.Sprintf("  • Container listing — %d entries visible", len(entries)))
	}

	st.FullDiskAccess = allOK
	return st
}

func FDAInstructions() string {
	return `Full Disk Access lets K-Cleaner scan sandbox containers and protected Library folders.

To enable:
  1. Open System Settings → Privacy & Security → Full Disk Access
  2. Click + and add K-Cleaner (or kclean from Terminal)
  3. Restart K-Cleaner

Without FDA, orphan container scanning and some app uninstall paths may be incomplete.`
}
