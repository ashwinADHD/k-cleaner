package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// MacOS displays a native notification via osascript.
func MacOS(title, message string) error {
	script := fmt.Sprintf(
		`display notification %s with title %s`,
		appleQuote(message), appleQuote(title),
	)
	return exec.Command("osascript", "-e", script).Run()
}

// ConfirmUninstall shows a dialog; returns true if user chose to clean.
func ConfirmUninstall(appName, appPath string) (bool, error) {
	script := fmt.Sprintf(`
set r to display dialog %s with title %s buttons {"Not Now", "Show Command"} default button 2
if button returned of r is "Show Command" then
  return "yes"
else
  return "no"
end if`,
		appleQuote(fmt.Sprintf("%s was moved to Trash.\n\nClean related files with K-Cleaner?", appName)),
		appleQuote("K-Cleaner Sentinel"),
	)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "yes", nil
}

// OpenTerminalCommand opens Terminal with a suggested kclean command.
func OpenTerminalCommand(appPath string) error {
	cmd := fmt.Sprintf("kclean uninstall-all %q", appPath)
	script := fmt.Sprintf(
		`tell application "Terminal" to do script %s`,
		appleQuote(cmd),
	)
	return exec.Command("osascript", "-e", script).Run()
}

func appleQuote(s string) string {
	return fmt.Sprintf("%q", s)
}
