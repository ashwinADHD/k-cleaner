package gui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ashwinADHD/k-cleaner/internal/appinfo"
	"github.com/ashwinADHD/k-cleaner/internal/clean"
	"github.com/ashwinADHD/k-cleaner/internal/config"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/report"
	"github.com/ashwinADHD/k-cleaner/internal/scan"
)

type uninstallUI struct {
	parent       *App
	appSelect    *widget.Select
	summary      *widget.Label
	list         *fyne.Container
	scanBtn      *widget.Button
	uninstallBtn *widget.Button
	selectAll    *widget.Button
	entries      []*itemEntry
	selectedApp  *models.AppInfo
	busy         bool
}

func (ui *App) buildUninstallTab() fyne.CanvasObject {
	u := &uninstallUI{parent: ui}
	apps := ui.scanner.Apps().InstalledApps()
	names := make([]string, 0, len(apps))
	appByName := make(map[string]models.AppInfo)
	for _, a := range apps {
		label := fmt.Sprintf("%s  (%s)", a.Name, a.BundleID)
		names = append(names, label)
		appByName[label] = a
	}

	u.summary = widget.NewLabel("Select an application, then scan for related files")
	u.summary.Wrapping = fyne.TextWrapWord

	u.list = container.NewVBox()
	scroll := container.NewVScroll(u.list)
	scroll.SetMinSize(fyne.NewSize(840, 380))

	u.appSelect = widget.NewSelect(names, func(s string) {
		if app, ok := appByName[s]; ok {
			copy := app
			u.selectedApp = &copy
		}
	})
	if len(names) > 0 {
		u.appSelect.SetSelected(names[0])
		if app, ok := appByName[names[0]]; ok {
			copy := app
			u.selectedApp = &copy
		}
	} else {
		u.appSelect.PlaceHolder = "No applications found in /Applications"
	}

	u.scanBtn = widget.NewButtonWithIcon("Scan Related Files", theme.SearchIcon(), u.onScan)
	u.uninstallBtn = widget.NewButtonWithIcon("Uninstall Selected", theme.DeleteIcon(), u.onUninstall)
	u.uninstallBtn.Disable()
	u.selectAll = widget.NewButton("Select All", u.onSelectAll)

	toolbar := container.NewHBox(u.scanBtn, u.uninstallBtn, widget.NewSeparator(), u.selectAll)

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Uninstall Application", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			u.appSelect,
			u.summary,
			widget.NewSeparator(),
		),
		toolbar,
		nil, nil, scroll,
	)
}

func (u *uninstallUI) setBusy(busy bool) {
	u.busy = busy
	if busy {
		u.scanBtn.Disable()
		u.uninstallBtn.Disable()
		u.selectAll.Disable()
		u.appSelect.Disable()
	} else {
		u.scanBtn.Enable()
		u.selectAll.Enable()
		u.appSelect.Enable()
		u.updateUninstallButton()
	}
}

func (u *uninstallUI) onScan() {
	if u.selectedApp == nil {
		dialog.ShowInformation("No app selected", "Choose an application from the list.", u.parent.window)
		return
	}
	u.setBusy(true)
	u.parent.status.SetText("Scanning related files…")
	u.parent.progress.Show()
	u.parent.progress.Start()

	app := u.selectedApp
	go func() {
		finder := scan.NewForwardScanner(app, models.SensitivityEnhanced)
		if ex, err := config.LoadExclusions(); err == nil {
			finder.Exclusions = ex
		}
		items := finder.FindRelated()
		if info, err := os.Stat(app.Path); err == nil {
			size, _ := scan.DirSize(app.Path)
			items = append([]models.Item{{
				Path: app.Path, Size: size, Category: models.CategoryAppRelated,
				Reason: "Application bundle", ModifiedAt: info.ModTime(),
			}}, items...)
		}

		fyne.Do(func() {
			u.applyItems(items, app)
			u.setBusy(false)
			u.parent.progress.Stop()
			u.parent.progress.Hide()
			u.parent.status.SetText(fmt.Sprintf("Found %d related items for %s", len(items), app.Name))
		})
	}()
}

func (u *uninstallUI) applyItems(items []models.Item, app *models.AppInfo) {
	u.entries = nil
	u.list.RemoveAll()

	var total int64
	for _, item := range items {
		total += item.Size
		ie := u.makeItemRow(item)
		u.entries = append(u.entries, ie)
		u.list.Add(ie.row)
	}

	if len(items) == 0 {
		u.summary.SetText(fmt.Sprintf("No related files found for %s", app.Name))
		u.uninstallBtn.Disable()
	} else {
		u.summary.SetText(fmt.Sprintf("%s — %d items, %s reclaimable (includes app bundle)",
			app.Name, len(items), report.FormatBytes(total)))
		u.updateUninstallButton()
	}
	u.list.Refresh()
}

func (u *uninstallUI) makeItemRow(item models.Item) *itemEntry {
	pathLabel := widget.NewLabel(report.ShortenPath(item.Path))
	pathLabel.Wrapping = fyne.TextWrapWord
	sizeLabel := widget.NewLabel(report.FormatBytes(item.Size))

	revealBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		_ = exec.Command("open", "-R", item.Path).Start()
	})
	revealBtn.Importance = widget.LowImportance

	ie := &itemEntry{item: item}
	ie.check = widget.NewCheck("", func(bool) {
		if !u.busy {
			u.updateUninstallButton()
		}
	})
	ie.check.SetChecked(true)

	ie.row = container.NewBorder(nil, nil,
		container.NewHBox(ie.check, sizeLabel), revealBtn, pathLabel)
	return ie
}

func (u *uninstallUI) selectedItems() []models.Item {
	var items []models.Item
	for _, ie := range u.entries {
		if ie.check.Checked {
			items = append(items, ie.item)
		}
	}
	return items
}

func (u *uninstallUI) updateUninstallButton() {
	if len(u.selectedItems()) > 0 {
		u.uninstallBtn.Enable()
	} else {
		u.uninstallBtn.Disable()
	}
}

func (u *uninstallUI) onSelectAll() {
	for _, ie := range u.entries {
		ie.check.SetChecked(true)
	}
	u.updateUninstallButton()
}

func (u *uninstallUI) onUninstall() {
	items := u.selectedItems()
	if len(items) == 0 || u.selectedApp == nil {
		return
	}
	var total int64
	for _, item := range items {
		total += item.Size
	}

	msg := fmt.Sprintf(
		"Move %d items for %s to Trash (%s)?\n\nIncludes the app bundle and related support files.",
		len(items), u.selectedApp.Name, report.FormatBytes(total))
	dialog.ShowConfirm("Confirm uninstall", msg, func(ok bool) {
		if ok {
			u.runUninstall(items)
		}
	}, u.parent.window)
}

func (u *uninstallUI) runUninstall(items []models.Item) {
	u.setBusy(true)
	u.parent.status.SetText("Moving items to Trash…")
	app := u.selectedApp

	go func() {
		_ = appinfo.KillRunningApp(app)
		bundle := "K-Cleaner_" + strings.ReplaceAll(app.Name, " ", "_")
		cl := &clean.Cleaner{DryRun: false, Bundle: bundle}
		cr := cl.CleanItems(items)

		fyne.Do(func() {
			u.entries = nil
			u.list.RemoveAll()
			u.summary.SetText(fmt.Sprintf("Uninstalled %s — freed %s (restore from %s)",
				app.Name, report.FormatBytes(cr.BytesFreed), report.ShortenPath(cr.TrashPath)))
			u.uninstallBtn.Disable()
			u.setBusy(false)
			u.parent.status.SetText("Uninstall complete")
			u.parent.progress.Stop()
			u.parent.progress.Hide()
		})
	}()
}
