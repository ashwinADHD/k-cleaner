package report

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ashwinADHD/k-cleaner/internal/models"
)

const brandName = "K-Cleaner"

func PrintScanSummary(w io.Writer, results []models.ScanResult, verbose bool) {
	var grandTotal int64
	var totalItems int

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  %s — System Cleanup Scan\n", brandName)
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	fmt.Fprintln(w, "")

	for _, r := range results {
		if len(r.Items) == 0 {
			continue
		}
		fmt.Fprintf(w, "  ▶ %s\n", r.Category.Label())
		fmt.Fprintf(w, "    %s\n", r.Category.Description())
		fmt.Fprintf(w, "    Reclaimable: %s (%d items)\n", FormatBytes(r.TotalSize), len(r.Items))

		if verbose {
			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			for _, item := range r.Items {
				fmt.Fprintf(tw, "      %s\t%s\n", FormatBytes(item.Size), ShortenPath(item.Path))
			}
			tw.Flush()
		}
		fmt.Fprintln(w, "")
		grandTotal += r.TotalSize
		totalItems += len(r.Items)
	}

	if grandTotal == 0 {
		fmt.Fprintln(w, "  ✓ Your Mac looks clean — nothing to reclaim right now.")
	} else {
		fmt.Fprintln(w, "  ─────────────────────────────────────────")
		fmt.Fprintf(w, "  Total reclaimable: %s across %d items\n", FormatBytes(grandTotal), totalItems)
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "  Run  kclean clean --all  to move these items to Trash.")
		fmt.Fprintln(w, "  Run  kclean clean --dry-run --all  to preview safely.")
	}
	fmt.Fprintln(w, "")
}

func PrintItemList(w io.Writer, title string, items []models.Item, verbose bool) {
	var total int64
	for _, item := range items {
		total += item.Size
	}

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  %s — %s\n", brandName, title)
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	fmt.Fprintf(w, "  Found %d items (%s)\n\n", len(items), FormatBytes(total))

	if verbose || len(items) <= 50 {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		for _, item := range items {
			fmt.Fprintf(tw, "  %s\t%s\n", FormatBytes(item.Size), ShortenPath(item.Path))
		}
		tw.Flush()
	} else {
		fmt.Fprintf(w, "  (Use --verbose to list all %d paths)\n", len(items))
	}
	fmt.Fprintln(w, "")
}

func PrintCleanSummary(w io.Writer, results []models.CleanResult, dryRun bool) {
	var totalFreed int64
	var totalDeleted, totalFailed int

	action := "Moved to Trash"
	if dryRun {
		action = "Would move to Trash"
	}

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  %s — Cleanup Results\n", brandName)
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	fmt.Fprintln(w, "")

	var trashPath string
	for _, r := range results {
		if r.Deleted == 0 && r.Failed == 0 {
			continue
		}
		fmt.Fprintf(w, "  %s\n", r.Category.Label())
		fmt.Fprintf(w, "    %s: %d items (%s)\n", action, r.Deleted, FormatBytes(r.BytesFreed))
		if r.Failed > 0 {
			fmt.Fprintf(w, "    Failed: %d\n", r.Failed)
			for _, e := range r.Errors {
				fmt.Fprintf(w, "      • %s\n", e)
			}
		}
		fmt.Fprintln(w, "")
		totalFreed += r.BytesFreed
		totalDeleted += r.Deleted
		totalFailed += r.Failed
		if r.TrashPath != "" {
			trashPath = r.TrashPath
		}
	}

	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	if dryRun {
		fmt.Fprintf(w, "  Preview: would free %s (%d items)\n", FormatBytes(totalFreed), totalDeleted)
	} else {
		fmt.Fprintf(w, "  Freed %s (%d items moved to Trash", FormatBytes(totalFreed), totalDeleted)
		if totalFailed > 0 {
			fmt.Fprintf(w, ", %d failed", totalFailed)
		}
		fmt.Fprintln(w, ")")
		if trashPath != "" {
			fmt.Fprintf(w, "  Restore from: %s\n", ShortenPath(trashPath))
		}
	}
	fmt.Fprintln(w, "")
}

func PrintCategories(w io.Writer) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  Available cleanup categories:")
	fmt.Fprintln(w, "")
	for _, cat := range models.AllCategories() {
		fmt.Fprintf(w, "    %-22s  %s\n", cat, cat.Label())
		fmt.Fprintf(w, "    %-22s  %s\n", "", cat.Description())
		fmt.Fprintln(w, "")
	}
}

func PrintPermissions(w io.Writer, fda bool, details []string) {
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  %s — Permissions\n", brandName)
	fmt.Fprintln(w, "  ─────────────────────────────────────────")
	if fda {
		fmt.Fprintln(w, "  ✓ Full Disk Access appears to be granted")
	} else {
		fmt.Fprintln(w, "  ⚠ Full Disk Access may be missing")
	}
	fmt.Fprintln(w, "")
	for _, d := range details {
		fmt.Fprintln(w, d)
	}
	fmt.Fprintln(w, "")
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ShortenPath(p string) string {
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	if idx := strings.Index(p, "/Library/"); idx > 0 {
		return "~" + p[idx:]
	}
	return p
}
