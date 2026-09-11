package pkgscan

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const receiptsDir = "/private/var/db/receipts"

// Package describes an macOS .pkg install receipt.
type Package struct {
	ID          string
	Version     string
	ReceiptPath string
	BOMPath     string
	FileCount   int
}

// List reads installed PKG receipts from /private/var/db/receipts.
func List() ([]Package, error) {
	entries, err := os.ReadDir(receiptsDir)
	if err != nil {
		return nil, err
	}

	var packages []Package
	seen := make(map[string]bool)

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".plist") {
			continue
		}
		id := strings.TrimSuffix(name, ".plist")
		if seen[id] {
			continue
		}
		seen[id] = true

		receiptPath := filepath.Join(receiptsDir, name)
		pkg := Package{
			ID:          id,
			Version:     plutilExtract(receiptPath, "PackageVersion"),
			ReceiptPath: receiptPath,
		}
		bomPath := filepath.Join(receiptsDir, id+".bom")
		if _, err := os.Stat(bomPath); err == nil {
			pkg.BOMPath = bomPath
			pkg.FileCount = bomFileCount(bomPath)
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

// Files lists paths installed by a PKG using lsbom (read-only).
func Files(bomPath string) ([]string, error) {
	if bomPath == "" {
		return nil, os.ErrNotExist
	}
	out, err := exec.Command("lsbom", bomPath).Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			files = append(files, fields[0])
		}
	}
	return files, nil
}

func bomFileCount(bomPath string) int {
	files, err := Files(bomPath)
	if err != nil {
		return 0
	}
	return len(files)
}

func plutilExtract(plistPath, key string) string {
	out, err := exec.Command("plutil", "-extract", key, "raw", "-o", "-", plistPath).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
