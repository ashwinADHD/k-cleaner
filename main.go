package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ashwinADHD/k-cleaner/internal/appinfo"
	"github.com/ashwinADHD/k-cleaner/internal/browser"
	"github.com/ashwinADHD/k-cleaner/internal/clean"
	"github.com/ashwinADHD/k-cleaner/internal/config"
	"github.com/ashwinADHD/k-cleaner/internal/gui"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/modules/daemons"
	"github.com/ashwinADHD/k-cleaner/internal/modules/pkgscan"
	"github.com/ashwinADHD/k-cleaner/internal/permissions"
	"github.com/ashwinADHD/k-cleaner/internal/report"
	"github.com/ashwinADHD/k-cleaner/internal/scan"
)

const version = "1.2.0"

func main() {
	if len(os.Args) < 2 {
		gui.Run()
		return
	}

	switch os.Args[1] {
	case "scan":
		runScan(os.Args[2:])
	case "clean":
		runClean(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "uninstall":
		runUninstall(os.Args[2:], false)
	case "uninstall-all":
		runUninstall(os.Args[2:], true)
	case "list-orphaned":
		runListOrphaned(os.Args[2:])
	case "remove-orphaned":
		runRemoveOrphaned(os.Args[2:])
	case "permissions":
		runPermissions()
	case "exclusions":
		runExclusions(os.Args[2:])
	case "daemons":
		runDaemons(os.Args[2:])
	case "pkg":
		runPkg(os.Args[2:])
	case "gui", "ui":
		gui.Run()
	case "categories":
		report.PrintCategories(os.Stdout)
	case "version", "-v", "--version":
		fmt.Printf("kclean %s (K-Cleaner for macOS)\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`K-Cleaner — macOS cleanup utility (Pearcleaner-inspired)

Safely scan caches, find orphaned app leftovers, and uninstall apps
with full related-file discovery. Items are moved to Trash (restorable).

Usage:
  kclean                                Open graphical interface
  kclean gui                            Open graphical interface
  kclean scan [--verbose]               Scan for reclaimable space
  kclean clean [options]                Move scanned items to Trash
  kclean list <app-path>                List files related to an app
  kclean uninstall <app-path>           Move app bundle to Trash
  kclean uninstall-all <app-path>       Uninstall app + related files
  kclean list-orphaned [--verbose]      List orphaned leftover files
  kclean remove-orphaned                Move orphaned files to Trash
  kclean permissions                    Check Full Disk Access
  kclean exclusions list|add|remove     Manage user exclusion paths
  kclean daemons list [--json]          List launch agents/daemons (read-only)
  kclean pkg list [--json]              List .pkg install receipts (read-only)
  kclean categories                     List cleanup categories
  kclean version                        Show version

Clean options:
  --all                                 Clean all categories
  --category <name>                     Clean one category (comma-separated)
  --dry-run                             Preview without moving files
  --yes, -y                             Skip confirmation prompt
  --json                                Output results as JSON

App / orphan options:
  --sensitivity strict|enhanced|deep    Match sensitivity (default: enhanced)
  --yes, -y                             Skip confirmation
  --json                                Output results as JSON

Examples:
  kclean scan --verbose
  kclean clean --category browser-cache --dry-run
  kclean list /Applications/MyApp.app
  kclean uninstall-all /Applications/MyApp.app --yes
  kclean list-orphaned --verbose
  kclean remove-orphaned --yes

`)
}

func runScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show individual items")
	asJSON := fs.Bool("json", false, "Output as JSON")
	_ = fs.Parse(args)

	sc, err := scan.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing scanner: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Scanning your Mac…")
	results := sc.ScanAll()
	if *asJSON {
		if err := report.WriteScanJSON(os.Stdout, results, *verbose); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
			os.Exit(1)
		}
		return
	}
	report.PrintScanSummary(os.Stdout, results, *verbose)
}

func runClean(args []string) {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	all := fs.Bool("all", false, "Clean all categories")
	dryRun := fs.Bool("dry-run", false, "Preview without deleting")
	yes := fs.Bool("yes", false, "Skip confirmation")
	yesShort := fs.Bool("y", false, "Skip confirmation")
	categories := fs.String("category", "", "Category to clean (comma-separated)")
	asJSON := fs.Bool("json", false, "Output as JSON")
	_ = fs.Parse(args)

	skipConfirm := *yes || *yesShort

	var cats []models.Category
	if *all {
		cats = models.AllCategories()
	} else if *categories != "" {
		for _, raw := range strings.Split(*categories, ",") {
			cats = append(cats, models.Category(strings.TrimSpace(raw)))
		}
	} else {
		fmt.Fprintln(os.Stderr, "Error: specify --all or --category <name>")
		os.Exit(1)
	}

	sc, err := scan.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing scanner: %v\n", err)
		os.Exit(1)
	}

	var allItems []models.Item
	var scanResults []models.ScanResult
	for _, cat := range cats {
		r := sc.ScanCategory(cat)
		scanResults = append(scanResults, r)
		allItems = append(allItems, r.Items...)
	}

	if len(allItems) == 0 {
		fmt.Fprint(os.Stdout, "\n  Nothing to clean in the selected categories.\n")
		return
	}

	var totalSize int64
	for _, item := range allItems {
		totalSize += item.Size
	}

	report.PrintScanSummary(os.Stdout, scanResults, false)

	if affected := browser.AffectedBrowsers(allItems); len(affected) > 0 && !*dryRun {
		fmt.Fprintf(os.Stderr, "\n  WARNING: %s\n", browser.WarningMessage(affected))
		if !confirm(skipConfirm, "Continue anyway?") {
			if !*asJSON {
				fmt.Fprint(os.Stdout, "\n  Cancelled.\n")
			}
			return
		}
	}

	msg := fmt.Sprintf("Move %d items to Trash (%s)?", len(allItems), report.FormatBytes(totalSize))
	if !confirm(skipConfirm || *dryRun, msg) {
		if !*asJSON {
			fmt.Fprint(os.Stdout, "\n  Cancelled.\n")
		}
		return
	}

	cl := &clean.Cleaner{DryRun: *dryRun}
	var cleanResults []models.CleanResult
	for _, sr := range scanResults {
		if len(sr.Items) == 0 {
			continue
		}
		cr := cl.CleanItems(sr.Items)
		cr.Category = sr.Category
		cleanResults = append(cleanResults, cr)
	}
	if *asJSON {
		_ = report.WriteCleanJSON(os.Stdout, cleanResults, *dryRun)
		return
	}
	report.PrintCleanSummary(os.Stdout, cleanResults, *dryRun)
}

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	sensitivity := fs.String("sensitivity", "enhanced", "Match sensitivity: strict, enhanced, deep")
	verbose := fs.Bool("verbose", true, "Show all paths")
	asJSON := fs.Bool("json", false, "Output as JSON")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: kclean list <app-path>")
		os.Exit(1)
	}

	app, err := appinfo.FromPath(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	level := models.ParseSensitivity(*sensitivity)
	finder := scan.NewForwardScanner(app, level)
	if ex, err := config.LoadExclusions(); err == nil {
		finder.Exclusions = ex
	}
	items := finder.FindRelated()

	// Include the app bundle itself.
	if appItem, ok := makeAppBundleItem(app.Path); ok {
		items = append([]models.Item{appItem}, items...)
	}

	title := fmt.Sprintf("Files for %s (%s)", app.Name, app.BundleID)
	if *asJSON {
		_ = report.WriteItemListJSON(os.Stdout, title, items)
		return
	}
	report.PrintItemList(os.Stdout, title, items, *verbose)
}

func runUninstall(args []string, includeRelated bool) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	sensitivity := fs.String("sensitivity", "enhanced", "Match sensitivity")
	yes := fs.Bool("yes", false, "Skip confirmation")
	dryRun := fs.Bool("dry-run", false, "Preview only")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: kclean uninstall[-all] <app-path>")
		os.Exit(1)
	}

	app, err := appinfo.FromPath(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var items []models.Item
	if includeRelated {
		level := models.ParseSensitivity(*sensitivity)
		finder := scan.NewForwardScanner(app, level)
		if ex, err := config.LoadExclusions(); err == nil {
			finder.Exclusions = ex
		}
		items = finder.FindRelated()
	}

	appItem, _ := makeAppBundleItem(app.Path)
	allItems := append([]models.Item{appItem}, items...)

	var total int64
	for _, item := range allItems {
		total += item.Size
	}

	action := "uninstall"
	if includeRelated {
		action = "uninstall-all"
	}
	report.PrintItemList(os.Stdout, fmt.Sprintf("%s: %s", action, app.Name), allItems, true)

	if !confirm(*yes || *dryRun, fmt.Sprintf("Move %d items to Trash (%s)?", len(allItems), report.FormatBytes(total))) {
		fmt.Fprint(os.Stdout, "\n  Cancelled.\n")
		return
	}

	if !*dryRun {
		_ = appinfo.KillRunningApp(app)
	}

	bundle := fmt.Sprintf("K-Cleaner_%s", strings.ReplaceAll(app.Name, " ", "_"))
	cl := &clean.Cleaner{DryRun: *dryRun, Bundle: bundle}
	cr := cl.CleanItems(allItems)
	cr.Category = models.CategoryAppRelated
	report.PrintCleanSummary(os.Stdout, []models.CleanResult{cr}, *dryRun)
}

func runListOrphaned(args []string) {
	fs := flag.NewFlagSet("list-orphaned", flag.ExitOnError)
	sensitivity := fs.String("sensitivity", "enhanced", "Match sensitivity")
	verbose := fs.Bool("verbose", false, "Show all paths")
	asJSON := fs.Bool("json", false, "Output as JSON")
	_ = fs.Parse(args)

	sc, err := scan.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	level := models.ParseSensitivity(*sensitivity)
	result := sc.ScanOrphans(level)
	if *asJSON {
		_ = report.WriteItemListJSON(os.Stdout, "Orphaned Files", result.Items)
		return
	}
	report.PrintItemList(os.Stdout, "Orphaned Files", result.Items, *verbose)
}

func runRemoveOrphaned(args []string) {
	fs := flag.NewFlagSet("remove-orphaned", flag.ExitOnError)
	sensitivity := fs.String("sensitivity", "enhanced", "Match sensitivity")
	yes := fs.Bool("yes", false, "Skip confirmation")
	dryRun := fs.Bool("dry-run", false, "Preview only")
	_ = fs.Parse(args)

	sc, err := scan.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	level := models.ParseSensitivity(*sensitivity)
	result := sc.ScanOrphans(level)
	if len(result.Items) == 0 {
		fmt.Fprint(os.Stdout, "\n  No orphaned files found.\n")
		return
	}

	report.PrintItemList(os.Stdout, "Orphaned Files", result.Items, false)

	if !confirm(*yes || *dryRun, fmt.Sprintf("Move %d orphaned items to Trash (%s)?",
		len(result.Items), report.FormatBytes(result.TotalSize))) {
		fmt.Fprint(os.Stdout, "\n  Cancelled.\n")
		return
	}

	cl := &clean.Cleaner{DryRun: *dryRun, Bundle: "K-Cleaner_Orphans"}
	cr := cl.CleanItems(result.Items)
	cr.Category = models.CategoryOrphanSupport
	report.PrintCleanSummary(os.Stdout, []models.CleanResult{cr}, *dryRun)
}

func runExclusions(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kclean exclusions list|add <path>|remove <path>")
		os.Exit(1)
	}

	ex, err := config.LoadExclusions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading exclusions: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		fmt.Printf("Exclusion file: %s\n\n", ex.FilePath())
		if len(ex.Paths) == 0 {
			fmt.Println("  (no exclusions configured)")
			return
		}
		for _, p := range ex.Paths {
			fmt.Printf("  %s\n", report.ShortenPath(p))
		}
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: kclean exclusions add <path>")
			os.Exit(1)
		}
		if err := ex.Add(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added exclusion: %s\n", report.ShortenPath(args[1]))
	case "remove":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: kclean exclusions remove <path>")
			os.Exit(1)
		}
		if err := ex.Remove(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed exclusion: %s\n", report.ShortenPath(args[1]))
	default:
		fmt.Fprintf(os.Stderr, "Unknown exclusions command: %s\n", args[0])
		os.Exit(1)
	}
}

func runPermissions() {
	st := permissions.CheckFullDiskAccess()
	report.PrintPermissions(os.Stdout, st.FullDiskAccess, st.Details)
	if !st.FullDiskAccess {
		fmt.Fprintln(os.Stdout, permissions.FDAInstructions())
	}
}

func runDaemons(args []string) {
	fs := flag.NewFlagSet("daemons", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "Output as JSON")
	_ = fs.Parse(args)

	if fs.NArg() < 1 || fs.Arg(0) != "list" {
		fmt.Fprintln(os.Stderr, "Usage: kclean daemons list [--json]")
		os.Exit(1)
	}

	sc, err := scan.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	entries := daemons.List(sc.Apps())
	if *asJSON {
		type row struct {
			Path       string `json:"path"`
			Label      string `json:"label"`
			Program    string `json:"program,omitempty"`
			Scope      string `json:"scope"`
			OrphanHint bool   `json:"orphan_hint"`
		}
		var rows []row
		for _, e := range entries {
			rows = append(rows, row{e.Path, e.Label, e.Program, e.Scope, e.OrphanHint})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{"count": len(rows), "entries": rows})
		return
	}

	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "  K-Cleaner — Launch Agents & Daemons (read-only)")
	fmt.Fprintln(os.Stdout, "  ─────────────────────────────────────────")
	fmt.Fprintf(os.Stdout, "  Found %d entries\n\n", len(entries))
	for _, e := range entries {
		flag := ""
		if e.OrphanHint {
			flag = "  [possible orphan]"
		}
		fmt.Fprintf(os.Stdout, "  %s%s\n", e.Label, flag)
		fmt.Fprintf(os.Stdout, "    %s · %s\n", e.Scope, report.ShortenPath(e.Path))
		if e.Program != "" {
			fmt.Fprintf(os.Stdout, "    Program: %s\n", e.Program)
		}
		fmt.Fprintln(os.Stdout, "")
	}
}

func runPkg(args []string) {
	fs := flag.NewFlagSet("pkg", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "Output as JSON")
	_ = fs.Parse(args)

	if fs.NArg() < 1 || fs.Arg(0) != "list" {
		fmt.Fprintln(os.Stderr, "Usage: kclean pkg list [--json]")
		os.Exit(1)
	}

	packages, err := pkgscan.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading PKG receipts: %v\n", err)
		os.Exit(1)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{"count": len(packages), "packages": packages})
		return
	}

	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "  K-Cleaner — PKG Install Receipts (read-only)")
	fmt.Fprintln(os.Stdout, "  ─────────────────────────────────────────")
	fmt.Fprintf(os.Stdout, "  Found %d packages\n\n", len(packages))
	for _, p := range packages {
		fmt.Fprintf(os.Stdout, "  %s", p.ID)
		if p.Version != "" {
			fmt.Fprintf(os.Stdout, "  v%s", p.Version)
		}
		fmt.Fprintln(os.Stdout, "")
		if p.FileCount > 0 {
			fmt.Fprintf(os.Stdout, "    BOM files: %d\n", p.FileCount)
		}
		fmt.Fprintf(os.Stdout, "    %s\n\n", report.ShortenPath(p.ReceiptPath))
	}
}

func confirm(skip bool, prompt string) bool {
	if skip {
		return true
	}
	fmt.Fprintf(os.Stdout, "  %s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func makeAppBundleItem(appPath string) (models.Item, bool) {
	info, err := os.Stat(appPath)
	if err != nil {
		return models.Item{}, false
	}
	size, _ := scan.DirSize(appPath)
	return models.Item{
		Path: appPath, Size: size, Category: models.CategoryAppRelated,
		Reason: "Application bundle", ModifiedAt: info.ModTime(),
	}, true
}
