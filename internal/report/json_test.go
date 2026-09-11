package report

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

func TestWriteScanJSON(t *testing.T) {
	results := []models.ScanResult{{
		Category: models.CategorySystemCache, TotalSize: 1024,
		Items: []models.Item{{Path: "/tmp/cache", Size: 1024, Category: models.CategorySystemCache}},
	}}
	var buf bytes.Buffer
	if err := WriteScanJSON(&buf, results, true); err != nil {
		t.Fatal(err)
	}
	var out jsonScanOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.TotalItems != 1 || out.TotalSize != 1024 {
		t.Errorf("unexpected output: %+v", out)
	}
}
