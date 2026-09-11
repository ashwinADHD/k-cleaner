package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

// Cleaner moves items to Trash (Pearcleaner undo-safe deletion).
type Cleaner struct {
	DryRun   bool
	Bundle   string // trash subfolder name
	Elevated bool   // retry permission failures with administrator privileges
}

func (c *Cleaner) CleanItems(items []models.Item) models.CleanResult {
	if len(items) == 0 {
		return models.CleanResult{}
	}

	result := models.CleanResult{Category: items[0].Category}
	trashBundle := c.trashBundleName(items)

	for _, item := range items {
		if !isSafePath(item.Path) {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("blocked unsafe path: %s", item.Path))
			continue
		}

		if c.DryRun {
			result.Deleted++
			result.BytesFreed += item.Size
			continue
		}

		freed, err := moveToTrash(item.Path, trashBundle)
		if err != nil && c.Elevated && isPermissionError(err) {
			if err2 := elevatedMoveToTrash(item.Path, trashBundle); err2 == nil {
				result.Deleted++
				result.BytesFreed += item.Size
				continue
			}
		}
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", item.Path, err))
			continue
		}
		result.Deleted++
		result.BytesFreed += freed
	}

	if !c.DryRun && result.Deleted > 0 {
		result.TrashPath = filepath.Join(paths.Trash(), trashBundle)
	}
	return result
}

func (c *Cleaner) trashBundleName(items []models.Item) string {
	if c.Bundle != "" {
		return c.Bundle
	}
	label := "K-Cleaner"
	if len(items) > 0 && items[0].Category == models.CategoryAppRelated {
		label = "K-Cleaner_Uninstall"
	}
	ts := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s", label, ts)
}

func isSafePath(target string) bool {
	home := paths.Home()
	if home == "" {
		return false
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		return false
	}

	// Block critical system paths.
	for _, blocked := range paths.BlockedDeleteRoots {
		if abs == blocked {
			return false
		}
	}
	if abs == home {
		return false
	}

	library := filepath.Join(home, "Library")
	trash := paths.Trash()

	if strings.HasPrefix(abs, library+string(os.PathSeparator)) {
		return true
	}
	if abs == trash || strings.HasPrefix(abs, trash+string(os.PathSeparator)) {
		return true
	}

	// Allow app bundles and usr/local paths for uninstall.
	if strings.HasSuffix(abs, ".app") {
		return true
	}
	if strings.HasPrefix(abs, "/usr/local/") {
		return true
	}
	if strings.HasPrefix(abs, "/Library/") {
		return true
	}
	if strings.HasPrefix(abs, "/private/var/db/receipts/") {
		return true
	}
	if strings.HasPrefix(abs, "/private/tmp/") {
		return true
	}
	if strings.HasPrefix(abs, filepath.Join(home, "Documents")+string(os.PathSeparator)) {
		return true
	}
	if strings.HasPrefix(abs, filepath.Join(home, "Desktop")+string(os.PathSeparator)) {
		return true
	}
	if strings.HasPrefix(abs, filepath.Join(home, ".config")+string(os.PathSeparator)) {
		return true
	}
	// Developer tool caches
	devPrefixes := []string{
		".npm", ".cargo", ".gradle", ".m2", ".swiftpm", ".cursor", ".cache", ".yarn-cache", "go",
	}
	for _, prefix := range devPrefixes {
		p := filepath.Join(home, prefix)
		if abs == p || strings.HasPrefix(abs, p+string(os.PathSeparator)) {
			return true
		}
	}
	devLib := filepath.Join(home, "Library", "Developer")
	if abs == devLib || strings.HasPrefix(abs, devLib+string(os.PathSeparator)) {
		return true
	}
	pnpm := filepath.Join(home, "Library", "pnpm")
	if abs == pnpm || strings.HasPrefix(abs, pnpm+string(os.PathSeparator)) {
		return true
	}

	return false
}

func moveToTrash(src, bundleName string) (int64, error) {
	info, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	var size int64
	if info.IsDir() {
		size, _ = dirSize(src)
	} else {
		size = info.Size()
	}

	destDir := filepath.Join(paths.Trash(), bundleName)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return 0, err
	}

	dest := uniqueTrashPath(destDir, filepath.Base(src))
	if err := os.Rename(src, dest); err != nil {
		// Cross-device or permission issue: try copy+remove.
		if err2 := copyAndRemove(src, dest); err2 != nil {
			return 0, fmt.Errorf("move to trash failed: %w", err)
		}
	}

	return size, nil
}

func uniqueTrashPath(dir, base string) string {
	dest := filepath.Join(dir, base)
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest
	}
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d", base, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s_%d", base, time.Now().UnixNano()))
}

func copyAndRemove(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := out.ReadFrom(in); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func dirSize(root string) (int64, error) {
	var total int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total, err
}

// RemoveAppBundle moves only the .app bundle to trash.
func RemoveAppBundle(appPath string, dryRun bool) (int64, error) {
	if !isSafePath(appPath) {
		return 0, fmt.Errorf("blocked unsafe path: %s", appPath)
	}
	if dryRun {
		size, _ := dirSize(appPath)
		return size, nil
	}
	bundle := fmt.Sprintf("K-Cleaner_Uninstall_%s", time.Now().Format("20060102_150405"))
	return moveToTrash(appPath, bundle)
}
