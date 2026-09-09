package scan

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func dirSize(root string) (int64, error) {
	out, err := exec.Command("du", "-sk", root).Output()
	if err == nil {
		fields := strings.Fields(string(out))
		if len(fields) >= 1 {
			if kb, err := strconv.ParseInt(fields[0], 10, 64); err == nil {
				return kb * 1024, nil
			}
		}
	}

	var total int64
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total, nil
}

func DirSize(root string) (int64, error) {
	return dirSize(root)
}
