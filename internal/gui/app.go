package gui

import (
	"fmt"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ashwinADHD/k-cleaner/internal/browser"
	"github.com/ashwinADHD/k-cleaner/internal/clean"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/permissions"
	"github.com/ashwinADHD/k-cleaner/internal/report"
	"github.com/ashwinADHD/k-cleaner/internal/scan"
)

type itemEntry struct {
	item  models.Item
	check *widget.Check
	row   fyne.CanvasObject
}

type categoryEntry struct {
	result      models.ScanResult
	check       *widget.Check
	expandBtn   *widget.Button
	itemsBox    *fyne.Container
	itemEntries []*itemEntry
	row         fyne.CanvasObject
	expanded    bool
}

type App struct {
	window    fyne.Window
	scanner   *scan.Scanner
	summary   *widget.Label
	warning   *widget.Label
	status    *widget.Label
	progress  *widget.ProgressBarInfinite
	list      *fyne.Container
	scanBtn   *widget.Button
	cleanBtn  *widget.Button
	selectAll *widget.Button
	entries   []*categoryEntry
	busy      bool
}

func Run() {
	sc, err := scan.New()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to initialize: %w", err), nil)
		return
	}

	st := permissions.CheckFullDiskAccess()

	a := app.NewWithID("com.ashwin.kcleaner")
	a.SetIcon(theme.ComputerIcon())

	ui := &App{scanner: sc, window: a.NewWindow("K-Cleaner")}
	ui.window.Resize(fyne.NewSize(880, 640))
	ui.build()

	if !st.FullDiskAccess {
		dialog.ShowInformation("Full Disk Access recommended",
			"Some scans (sandbox containers, protected folders) work best with Full Disk Access.\n\n"+
				"Run  kclean permissions  for setup instructions.",
			ui.window)
	}

	ui.window.ShowAndRun()
}

func (ui *App) build() {
	ui.summary = widget.NewLabel("Click Scan to find reclaimable space")
	ui.summary.TextStyle = fyne.TextStyle{Bold: true}

	ui.warning = widget.NewLabel("")
	ui.warning.Wrapping = fyne.TextWrapWord
	ui.warning.Importance = widget.WarningImportance
	ui.warning.Hide()

	ui.status = widget.NewLabel("Ready — items are moved to Trash (restorable)")
	ui.progress = widget.NewProgressBarInfinite()
	ui.progress.Hide()

	ui.list = container.NewVBox()
	scroll := container.NewVScroll(ui.list)
	scroll.SetMinSize(fyne.NewSize(840, 400))

	ui.scanBtn = widget.NewButtonWithIcon("Scan", theme.SearchIcon(), ui.onScan)
	ui.cleanBtn = widget.NewButtonWithIcon("Clean Selected", theme.DeleteIcon(), ui.onClean)
	ui.cleanBtn.Disable()
	ui.selectAll = widget.NewButton("Select All", ui.onSelectAll)

	toolbar := container.NewHBox(ui.scanBtn, ui.cleanBtn, widget.NewSeparator(), ui.selectAll)

	cleanupTab := container.NewBorder(
		container.NewVBox(ui.summary, ui.warning, widget.NewSeparator()),
		toolbar,
		nil, nil, scroll,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Cleanup", cleanupTab),
		container.NewTabItem("Uninstall", ui.buildUninstallTab()),
	)

	footer := container.NewVBox(
		widget.NewSeparator(), ui.progress, ui.status,
	)

	ui.window.SetContent(container.NewBorder(nil, footer, nil, nil, tabs))
}

func (ui *App) setBusy(busy bool, status string) {
	ui.busy = busy
	if busy {
		ui.scanBtn.Disable()
		ui.cleanBtn.Disable()
		ui.selectAll.Disable()
	} else {
		ui.scanBtn.Enable()
		ui.selectAll.Enable()
		ui.updateCleanButton()
	}
	ui.status.SetText(status)
	if busy {
		ui.progress.Show()
		ui.progress.Start()
	} else {
		ui.progress.Stop()
		ui.progress.Hide()
	}
}

func (ui *App) updateCleanButton() {
	if ui.hasSelection() {
		ui.cleanBtn.Enable()
	} else {
		ui.cleanBtn.Disable()
	}
}

func (ui *App) onScan() {
	ui.setBusy(true, "Scanning your Mac…")
	go func() {
		results := ui.scanner.ScanAll()
		fyne.Do(func() {
			ui.applyScanResults(results)
			ui.setBusy(false, "Scan complete — expand a category to see exact paths")
		})
	}()
}

func (ui *App) applyScanResults(results []models.ScanResult) {
	ui.entries = nil
	ui.list.RemoveAll()

	var total int64
	var itemCount int

	for _, r := range results {
		entry := ui.makeCategoryRow(r)
		ui.entries = append(ui.entries, entry)
		ui.list.Add(entry.row)
		total += r.TotalSize
		itemCount += len(r.Items)
	}

	if itemCount == 0 {
		ui.summary.SetText("Your Mac looks clean — nothing to reclaim")
		ui.warning.Hide()
		ui.cleanBtn.Disable()
	} else {
		ui.summary.SetText(fmt.Sprintf("Total reclaimable: %s  ·  %d items",
			report.FormatBytes(total), itemCount))
		ui.updateBrowserWarning()
		ui.updateCleanButton()
	}
	ui.list.Refresh()
}

func (ui *App) updateBrowserWarning() {
	running := browser.Running()
	if len(running) == 0 {
		ui.warning.Hide()
		return
	}
	ui.warning.SetText(fmt.Sprintf(
		"⚠ %s currently running — quit before cleaning browser cache.",
		joinNames(running),
	))
	ui.warning.Show()
}

func (ui *App) makeCategoryRow(r models.ScanResult) *categoryEntry {
	entry := &categoryEntry{result: r}
	sizeText := report.FormatBytes(r.TotalSize)
	count := len(r.Items)

	entry.check = widget.NewCheck(
		fmt.Sprintf("%s  —  %s (%d)", r.Category.Label(), sizeText, count),
		func(checked bool) {
			for _, ie := range entry.itemEntries {
				ie.check.SetChecked(checked)
			}
			if !ui.busy {
				ui.updateCleanButton()
			}
		},
	)
	entry.check.SetChecked(r.TotalSize > 0)
	if r.TotalSize == 0 {
		entry.check.SetChecked(false)
		entry.check.Disable()
	}

	desc := widget.NewLabel(r.Category.Description())
	desc.Wrapping = fyne.TextWrapWord
	desc.Importance = widget.LowImportance

	entry.itemsBox = container.NewVBox()
	colHeader := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Size", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Path", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{}),
	)
	entry.itemsBox.Add(colHeader)
	for _, item := range r.Items {
		ie := ui.makeItemRow(entry, item)
		entry.itemEntries = append(entry.itemEntries, ie)
		entry.itemsBox.Add(ie.row)
	}

	expandLabel := fmt.Sprintf("Show %d paths", count)
	if count == 0 {
		expandLabel = "No paths"
	}
	entry.expandBtn = widget.NewButtonWithIcon(expandLabel, theme.MenuExpandIcon(), func() {
		entry.toggleExpand()
	})
	if count == 0 {
		entry.expandBtn.Disable()
	}

	header := container.NewBorder(nil, nil, nil, entry.expandBtn, entry.check)
	entry.row = container.NewVBox(header, desc, entry.itemsBox, widget.NewSeparator())
	entry.itemsBox.Hide()
	return entry
}

func (ui *App) makeItemRow(cat *categoryEntry, item models.Item) *itemEntry {
	pathLabel := widget.NewLabel(report.ShortenPath(item.Path))
	pathLabel.Wrapping = fyne.TextWrapWord
	sizeLabel := widget.NewLabel(report.FormatBytes(item.Size))

	revealBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		_ = exec.Command("open", "-R", item.Path).Start()
	})
	revealBtn.Importance = widget.LowImportance

	ie := &itemEntry{item: item}
	ie.check = widget.NewCheck("", func(bool) {
		ui.syncCategoryCheck(cat)
		if !ui.busy {
			ui.updateCleanButton()
		}
	})
	ie.check.SetChecked(true)

	ie.row = container.NewBorder(nil, nil,
		container.NewHBox(ie.check, sizeLabel), revealBtn, pathLabel)
	return ie
}

func (cat *categoryEntry) toggleExpand() {
	cat.expanded = !cat.expanded
	if cat.expanded {
		cat.itemsBox.Show()
		cat.expandBtn.SetText(fmt.Sprintf("Hide %d paths", len(cat.result.Items)))
		cat.expandBtn.SetIcon(theme.MenuDropDownIcon())
	} else {
		cat.itemsBox.Hide()
		cat.expandBtn.SetText(fmt.Sprintf("Show %d paths", len(cat.result.Items)))
		cat.expandBtn.SetIcon(theme.MenuExpandIcon())
	}
	cat.itemsBox.Refresh()
}

func (ui *App) syncCategoryCheck(cat *categoryEntry) {
	all, any := true, false
	for _, ie := range cat.itemEntries {
		if ie.check.Checked {
			any = true
		} else {
			all = false
		}
	}
	if any && all {
		cat.check.SetChecked(true)
	} else if !any {
		cat.check.SetChecked(false)
	}
}

func (ui *App) selectedItems() []models.Item {
	var items []models.Item
	for _, cat := range ui.entries {
		for _, ie := range cat.itemEntries {
			if ie.check.Checked {
				items = append(items, ie.item)
			}
		}
	}
	return items
}

func (ui *App) hasSelection() bool {
	return len(ui.selectedItems()) > 0
}

func (ui *App) onSelectAll() {
	for _, cat := range ui.entries {
		if len(cat.itemEntries) == 0 {
			continue
		}
		cat.check.SetChecked(true)
		for _, ie := range cat.itemEntries {
			ie.check.SetChecked(true)
		}
	}
	ui.updateCleanButton()
}

func (ui *App) onClean() {
	items := ui.selectedItems()
	if len(items) == 0 {
		dialog.ShowInformation("Nothing selected", "Select at least one item to clean.", ui.window)
		return
	}

	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}

	affected := browser.AffectedBrowsers(items)
	if len(affected) > 0 {
		ui.showBrowserWarning(items, totalSize, affected)
		return
	}
	ui.confirmClean(items, totalSize)
}

func (ui *App) showBrowserWarning(items []models.Item, totalSize int64, browsers []string) {
	msg := browser.WarningMessage(browsers) + fmt.Sprintf(
		"\n\nYou selected %d items (%s) including browser cache.",
		len(items), report.FormatBytes(totalSize))

	d := dialog.NewCustomConfirm("Browsers are running", "Skip browser cache", "Clean anyway",
		widget.NewLabel(msg), func(skipBrowser bool) {
			if skipBrowser {
				filtered := filterOutBrowserCache(items)
				if len(filtered) == 0 {
					dialog.ShowInformation("Nothing left", "All selected items were browser cache.", ui.window)
					return
				}
				var size int64
				for _, it := range filtered {
					size += it.Size
				}
				ui.confirmClean(filtered, size)
			} else {
				ui.confirmClean(items, totalSize)
			}
		}, ui.window)
	d.SetDismissText("Cancel")
	d.Show()
}

func (ui *App) confirmClean(items []models.Item, totalSize int64) {
	msg := fmt.Sprintf(
		"Move %d items to Trash and free %s?\n\nItems can be restored from Trash. Grouped under K-Cleaner_* in Trash.",
		len(items), report.FormatBytes(totalSize))
	dialog.ShowConfirm("Confirm cleanup", msg, func(ok bool) {
		if ok {
			ui.runClean(items)
		}
	}, ui.window)
}

func (ui *App) runClean(items []models.Item) {
	ui.setBusy(true, "Moving items to Trash…")
	go func() {
		cl := &clean.Cleaner{DryRun: false}
		cr := cl.CleanItems(items)
		results := ui.scanner.ScanAll()
		fyne.Do(func() {
			ui.applyScanResults(results)
			msg := fmt.Sprintf("Freed %s (%d items) — restore from %s",
				report.FormatBytes(cr.BytesFreed), cr.Deleted, report.ShortenPath(cr.TrashPath))
			ui.setBusy(false, msg)
		})
	}()
}

func filterOutBrowserCache(items []models.Item) []models.Item {
	var out []models.Item
	for _, item := range items {
		if item.Category != models.CategoryBrowserCache {
			out = append(out, item)
		}
	}
	return out
}

func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		return fmt.Sprintf("%s, and %s", names[0], names[len(names)-1])
	}
}
