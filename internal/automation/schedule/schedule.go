package schedule

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ashwinADHD/k-cleaner/internal/automation/launchd"
	"github.com/ashwinADHD/k-cleaner/internal/automation/notify"
	"github.com/ashwinADHD/k-cleaner/internal/models"
	"github.com/ashwinADHD/k-cleaner/internal/report"
	"github.com/ashwinADHD/k-cleaner/internal/scan"
)

const Label = "com.ashwin.kcleaner.schedule"

type Interval string

const (
	Daily  Interval = "daily"
	Weekly Interval = "weekly"
)

// Install registers a launchd job for periodic scan summaries.
func Install(interval Interval) error {
	bin, err := launchd.KcleanBinary()
	if err != nil {
		return err
	}
	stdout := filepath.Join(launchd.LogDir(), "schedule.log")
	stderr := filepath.Join(launchd.LogDir(), "schedule.error.log")

	calendar := weeklyCalendar()
	if interval == Daily {
		calendar = dailyCalendar()
	}

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>schedule</string>
		<string>run</string>
	</array>
	<key>StartCalendarInterval</key>
	%s
	<key>StandardOutPath</key>
	<string>%s</string>
	<key>StandardErrorPath</key>
	<string>%s</string>
	<key>ProcessType</key>
	<string>Background</string>
</dict>
</plist>
`, Label, launchd.EscapeXML(bin), calendar, launchd.EscapeXML(stdout), launchd.EscapeXML(stderr))

	return launchd.Install(Label, plist)
}

func dailyCalendar() string {
	return `<dict>
		<key>Hour</key>
		<integer>9</integer>
		<key>Minute</key>
		<integer>0</integer>
	</dict>`
}

func weeklyCalendar() string {
	return `<dict>
		<key>Weekday</key>
		<integer>0</integer>
		<key>Hour</key>
		<integer>9</integer>
		<key>Minute</key>
		<integer>0</integer>
	</dict>`
}

// Uninstall removes the scheduled scan LaunchAgent.
func Uninstall() error {
	return launchd.Uninstall(Label)
}

// Status returns installation state.
func Status() (installed, running bool, plistPath string) {
	return launchd.IsInstalled(Label), launchd.IsRunning(Label), launchd.PlistPath(Label)
}

// Run performs a scan, logs results, and notifies if reclaimable space exceeds threshold.
func Run(notifyThreshold int64) error {
	sc, err := scan.New()
	if err != nil {
		return err
	}
	results := sc.ScanAll()

	var total int64
	var items int
	for _, r := range results {
		total += r.TotalSize
		items += len(r.Items)
	}

	logPath := filepath.Join(launchd.LogDir(),
		fmt.Sprintf("scan-%s.log", time.Now().Format("20060102-150405")))
	_ = os.MkdirAll(launchd.LogDir(), 0o755)

	f, err := os.Create(logPath)
	if err == nil {
		fmt.Fprintf(f, "K-Cleaner scheduled scan %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(f, "Total: %s (%d items)\n\n", report.FormatBytes(total), items)
		for _, r := range results {
			if len(r.Items) == 0 {
				continue
			}
			fmt.Fprintf(f, "%s: %s (%d items)\n", r.Category.Label(), report.FormatBytes(r.TotalSize), len(r.Items))
		}
		f.Close()
	}

	if total >= notifyThreshold {
		_ = notify.MacOS("K-Cleaner Scan",
			fmt.Sprintf("%s reclaimable across %d items. Run kclean scan for details.", report.FormatBytes(total), items))
	}
	return nil
}

// ParseInterval parses daily|weekly.
func ParseInterval(s string) Interval {
	if s == "daily" {
		return Daily
	}
	return Weekly
}

// DefaultNotifyThreshold is 100 MB.
const DefaultNotifyThreshold int64 = 100 * 1024 * 1024

func FormatResultsSummary(results []models.ScanResult) (int64, int) {
	var total int64
	var items int
	for _, r := range results {
		total += r.TotalSize
		items += len(r.Items)
	}
	return total, items
}
