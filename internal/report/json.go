package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

const jsonVersion = "1.1"

type jsonItem struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	SizeHuman  string `json:"size_human"`
	Category   string `json:"category"`
	Reason     string `json:"reason,omitempty"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

type jsonScanResult struct {
	Category    string     `json:"category"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	TotalSize   int64      `json:"total_size"`
	SizeHuman   string     `json:"size_human"`
	ItemCount   int        `json:"item_count"`
	Items       []jsonItem `json:"items,omitempty"`
	Errors      []string   `json:"errors,omitempty"`
}

type jsonScanOutput struct {
	Tool        string           `json:"tool"`
	Version     string           `json:"version"`
	TotalSize   int64            `json:"total_size"`
	SizeHuman   string           `json:"size_human"`
	TotalItems  int            `json:"total_items"`
	Results     []jsonScanResult `json:"results"`
}

type jsonCleanResult struct {
	Category   string   `json:"category"`
	Deleted    int      `json:"deleted"`
	Failed     int      `json:"failed"`
	BytesFreed int64    `json:"bytes_freed"`
	SizeHuman  string   `json:"size_human"`
	TrashPath  string   `json:"trash_path,omitempty"`
	Errors     []string `json:"errors,omitempty"`
}

type jsonCleanOutput struct {
	Tool       string            `json:"tool"`
	Version    string            `json:"version"`
	DryRun     bool              `json:"dry_run"`
	BytesFreed int64             `json:"bytes_freed"`
	SizeHuman  string            `json:"size_human"`
	Deleted    int               `json:"deleted"`
	Failed     int               `json:"failed"`
	Results    []jsonCleanResult `json:"results"`
}

type jsonItemListOutput struct {
	Tool       string     `json:"tool"`
	Version    string     `json:"version"`
	Title      string     `json:"title"`
	TotalSize  int64      `json:"total_size"`
	SizeHuman  string     `json:"size_human"`
	TotalItems int        `json:"total_items"`
	Items      []jsonItem `json:"items"`
}

func WriteScanJSON(w io.Writer, results []models.ScanResult, includeItems bool) error {
	out := jsonScanOutput{
		Tool: brandName, Version: jsonVersion, Results: make([]jsonScanResult, 0),
	}
	for _, r := range results {
		if len(r.Items) == 0 && len(r.Errors) == 0 {
			continue
		}
		jr := jsonScanResult{
			Category: string(r.Category), Label: r.Category.Label(),
			Description: r.Category.Description(), TotalSize: r.TotalSize,
			SizeHuman: FormatBytes(r.TotalSize), ItemCount: len(r.Items), Errors: r.Errors,
		}
		if includeItems {
			for _, item := range r.Items {
				jr.Items = append(jr.Items, toJSONItem(item))
			}
		}
		out.Results = append(out.Results, jr)
		out.TotalSize += r.TotalSize
		out.TotalItems += len(r.Items)
	}
	out.SizeHuman = FormatBytes(out.TotalSize)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func WriteItemListJSON(w io.Writer, title string, items []models.Item) error {
	var total int64
	jsonItems := make([]jsonItem, 0, len(items))
	for _, item := range items {
		total += item.Size
		jsonItems = append(jsonItems, toJSONItem(item))
	}
	out := jsonItemListOutput{
		Tool: brandName, Version: jsonVersion, Title: title,
		TotalSize: total, SizeHuman: FormatBytes(total),
		TotalItems: len(items), Items: jsonItems,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func WriteCleanJSON(w io.Writer, results []models.CleanResult, dryRun bool) error {
	out := jsonCleanOutput{
		Tool: brandName, Version: jsonVersion, DryRun: dryRun,
		Results: make([]jsonCleanResult, 0),
	}
	for _, r := range results {
		if r.Deleted == 0 && r.Failed == 0 {
			continue
		}
		out.Results = append(out.Results, jsonCleanResult{
			Category: string(r.Category), Deleted: r.Deleted, Failed: r.Failed,
			BytesFreed: r.BytesFreed, SizeHuman: FormatBytes(r.BytesFreed),
			TrashPath: r.TrashPath, Errors: r.Errors,
		})
		out.BytesFreed += r.BytesFreed
		out.Deleted += r.Deleted
		out.Failed += r.Failed
	}
	out.SizeHuman = FormatBytes(out.BytesFreed)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func toJSONItem(item models.Item) jsonItem {
	ji := jsonItem{
		Path: item.Path, Size: item.Size, SizeHuman: FormatBytes(item.Size),
		Category: string(item.Category), Reason: item.Reason,
	}
	if !item.ModifiedAt.IsZero() {
		ji.ModifiedAt = item.ModifiedAt.UTC().Format(time.RFC3339)
	}
	return ji
}
