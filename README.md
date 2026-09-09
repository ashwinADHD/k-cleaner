# K-Cleaner

A standalone macOS cleanup utility — Pearcleaner-inspired, built in Go.

Scan caches, find orphaned app leftovers, and uninstall applications with full related-file discovery. Transparent, safe, and restorable: everything goes to **Trash**, not permanent deletion.

Inspired by [Pearcleaner](https://github.com/alienator88/Pearcleaner) and CleanMyMac.

**Repository:** [github.com/ashwinADHD/k-cleaner](https://github.com/ashwinADHD/k-cleaner)

## Documentation

| Document | Description |
|---|---|
| [docs/DECISIONS.md](docs/DECISIONS.md) | **What, why, and how** — product and engineering decisions |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System framework, layers, and data flow |
| [docs/CLI.md](docs/CLI.md) | Full command reference |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Phased delivery plan |
| [docs/WORKFLOW.md](docs/WORKFLOW.md) | Git workflow (`feat/*` → `main`) |

## Features

| Category | What it cleans |
|---|---|
| **Browser Cache** | Safari, Chrome, Firefox, Edge, Brave, Arc, Vivaldi |
| **System Cache** | `~/Library/Caches` (protected Apple paths skipped) |
| **Application Logs** | `~/Library/Logs` |
| **Orphan Detection** | Leftovers from uninstalled apps (reverse scan) |
| **App Uninstall** | `.app` bundle + related prefs, caches, containers |
| **Trash** | Empty reclaimable space from Trash |

### Pearcleaner-inspired capabilities

- **Forward scanner** — given an app, finds related files across 50+ locations
- **Reverse scanner** — finds orphans by comparing filesystem to installed apps
- **`pearFormat` matching** — normalized bundle ID / app name matching
- **Sensitivity tiers** — `strict`, `enhanced` (default), `deep`
- **UUID container resolution** — reads Apple sandbox metadata plists
- **Trash-based deletion** — grouped under `K-Cleaner_*` folders (undo-safe)
- **Full Disk Access check** — `kclean permissions`

## Safety

- Scan first — always shows paths and sizes before cleanup
- Dry-run mode — preview without moving anything
- Confirmation prompts — explicit approval required
- Browser warnings — alerts when Safari/Chrome/Firefox are running
- Apple system files protected — skips critical `com.apple.*` paths
- Blocked roots — never touches `/`, `/System`, home directory root

## Requirements

- macOS 12 or later
- Go 1.22+ (to build from source)

## Install

### Native Mac app

```bash
cd kcleaner
make app              # builds K-Cleaner.app
make install-app      # copies to /Applications
```

**First launch:** Right-click **K-Cleaner → Open** (unsigned app gate).

### Command line

```bash
cd kcleaner
go build -o kclean .
sudo mv kclean /usr/local/bin/   # optional
```

## CLI Usage (Pearcleaner-style)

```bash
# Graphical interface
kclean
kclean gui

# System cleanup scan
kclean scan
kclean scan --verbose

# Clean caches and orphans
kclean clean --all --dry-run
kclean clean --category browser-cache,system-cache
kclean clean --all --yes

# App uninstall (Pearcleaner pear commands)
kclean list /Applications/MyApp.app
kclean uninstall /Applications/MyApp.app
kclean uninstall-all /Applications/MyApp.app --yes

# Orphan management
kclean list-orphaned --verbose
kclean remove-orphaned --yes

# Permissions
kclean permissions

# Help
kclean categories
kclean version
```

## Graphical interface

Launch with no arguments or `kclean gui`:

- **Scan** — finds reclaimable space across all categories
- **Show paths** — expand categories to see every file/folder (Pearcleaner-style transparency)
- **Per-item checkboxes** — deselect anything before cleaning
- **Reveal in Finder** — folder icon opens exact location
- **Trash-safe** — items moved to Trash, not permanently deleted

## How orphan detection works

K-Cleaner indexes apps in `/Applications`, `/System/Applications`, and `~/Applications` (bundle IDs, names, entitlements). It then reverse-scans Library folders and flags items that don't match any installed app — using Pearcleaner's `pearFormat` normalization and skip-keyword blocklists.

## Full Disk Access

For complete container and sandbox scanning:

1. System Settings → Privacy & Security → Full Disk Access
2. Add **K-Cleaner** or **kclean**
3. Restart the app

Run `kclean permissions` to check status.

## License

MIT
