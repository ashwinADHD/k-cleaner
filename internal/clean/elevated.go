package clean

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/paths"
)

func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	if os.IsPermission(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "permission denied") || strings.Contains(msg, "operation not permitted")
}

func elevatedMoveToTrash(src, bundleName string) error {
	destDir := filepath.Join(paths.Trash(), bundleName)
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	shell := "mkdir -p " + shellQuote(destAbs) + " && mv " + shellQuote(srcAbs) + " " + shellQuote(destAbs) + "/"
	script := "do shell script " + shellQuote(shell) + " with administrator privileges"
	return exec.Command("osascript", "-e", script).Run()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
