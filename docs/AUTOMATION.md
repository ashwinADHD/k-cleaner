# K-Cleaner Automation (Phase 4)

Background automation inspired by Pearcleaner Sentinel and scheduled maintenance.

---

## Sentinel — Trash Watcher

When you drag an app to Trash, Sentinel can notify you and offer to clean related files.

### Commands

```bash
kclean sentinel install      # LaunchAgent — starts at login
kclean sentinel status
kclean sentinel run          # Foreground (testing)
kclean sentinel uninstall
```

### Behavior

1. Watches `~/.Trash` for new `.app` bundles
2. Shows a macOS notification
3. Optional dialog to open Terminal with `kclean uninstall-all <path>`

Logs: `~/Library/Logs/K-Cleaner/sentinel.log`

---

## Scheduled Scan

Periodic scan summary via `launchd` (no automatic deletion).

```bash
kclean schedule install --weekly   # Sundays 9:00 AM
kclean schedule install --daily    # Daily 9:00 AM
kclean schedule status
kclean schedule run
kclean schedule uninstall
```

Writes logs to `~/Library/Logs/K-Cleaner/scan-*.log` and notifies if reclaimable ≥ 100 MB.

---

## Elevated Clean

For protected `/Library` paths that fail with permission errors:

```bash
kclean clean --all --elevated
kclean uninstall-all /Applications/SomeApp.app --elevated --yes
```

Prompts for administrator password via macOS — retries permission failures only.

---

## Remove automation

```bash
kclean sentinel uninstall
kclean schedule uninstall
```
