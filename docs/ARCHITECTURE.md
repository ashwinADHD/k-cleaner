# K-Cleaner — System Architecture

This document defines the **framework** for K-Cleaner: layers, packages, data flow, and extension points.

---

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Presentation Layer                        │
│  ┌──────────────────────┐    ┌──────────────────────────────┐ │
│  │  main.go (CLI router) │    │  internal/gui (Fyne UI)      │ │
│  └──────────┬───────────┘    └──────────────┬───────────────┘ │
└─────────────┼─────────────────────────────────┼─────────────────┘
              │                                 │
┌─────────────▼─────────────────────────────────▼─────────────────┐
│                      Application Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ scan        │  │ clean       │  │ report / browser        │ │
│  │ forward     │  │ trash move  │  │ output & UX warnings    │ │
│  │ reverse     │  │ safety      │  │                         │ │
│  │ category    │  │             │  │                         │ │
│  └──────┬──────┘  └──────┬──────┘  └─────────────────────────┘ │
└─────────┼────────────────┼──────────────────────────────────────┘
          │                │
┌─────────▼────────────────▼──────────────────────────────────────┐
│                        Domain Layer                              │
│  ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌──────────────────┐ │
│  │ models   │ │ match    │ │ conditions│ │ appinfo          │ │
│  │ Category │ │ PearFormat│ │ skip rules│ │ plist/codesign   │ │
│  │ Item     │ │ sensitivity│ │ per-app  │ │ container UUID   │ │
│  └──────────┘ └──────────┘ └───────────┘ └──────────────────┘ │
└─────────┼────────────────────────────────────────────────────────┘
          │
┌─────────▼────────────────────────────────────────────────────────┐
│                     Infrastructure Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────────┐  │
│  │ paths        │  │ appindex     │  │ permissions             │  │
│  │ locations    │  │ installed    │  │ Full Disk Access probe  │  │
│  │ macOS dirs   │  │ app catalog  │  │                         │  │
│  └──────────────┘  └──────────────┘  └─────────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
          │
┌─────────▼────────────────────────────────────────────────────────┐
│                     macOS Filesystem / Tools                       │
│  ~/Library/*  /Applications  plutil  codesign  du  pgrep  open   │
└──────────────────────────────────────────────────────────────────┘
```

---

## Package Reference

| Package | Responsibility | Key types / functions |
|---|---|---|
| `main` | CLI routing, flags, confirmation prompts | `runScan`, `runUninstall`, … |
| `internal/models` | Shared domain types | `Category`, `Item`, `AppInfo`, `Sensitivity` |
| `internal/paths` | macOS path constants | `Home()`, `Caches()`, `ForwardScanPaths()` |
| `internal/match` | String normalization & matching | `PearFormat()`, `MatchesApp()` |
| `internal/conditions` | Safety blocklists & per-app rules | `ShouldSkipReverse()`, `AppConditions` |
| `internal/appinfo` | Read `.app` metadata | `FromPath()`, `ContainerBundleID()` |
| `internal/appindex` | Installed application catalog | `Build()`, `HasBundleID()` |
| `internal/scan` | All scan strategies | `Scanner`, `ForwardScanner`, `ReverseScanner` |
| `internal/clean` | Trash-safe deletion | `Cleaner.CleanItems()`, `isSafePath()` |
| `internal/permissions` | FDA detection | `CheckFullDiskAccess()` |
| `internal/browser` | Running browser detection | `Running()`, `AffectedBrowsers()` |
| `internal/report` | Formatted CLI output | `PrintScanSummary()`, `FormatBytes()` |
| `internal/gui` | Fyne graphical interface | `Run()`, category/item rows |

---

## Core Data Flow

### A. Category scan (cache / logs / trash)

```
Scanner.ScanAll()
  → for each Category:
      scanBrowserCache | scanSystemCache | scanLogs | scanTrash
      scanOrphansByCategory (reverse engine filtered by category)
  → []ScanResult { Items[], TotalSize }
  → report.PrintScanSummary / GUI render
```

### B. App uninstall (forward scan)

```
appinfo.FromPath("/Applications/Foo.app")
  → bundle ID, name, entitlements
ForwardScanner.FindRelated()
  → walk ForwardScanPaths() with depth 1–2
  → match via match.MatchesApp + conditions
  → resolve UUID containers via ContainerBundleID
  → dedupe parent/child paths
  → []Item
Cleaner.CleanItems() → ~/.Trash/K-Cleaner_Foo/
```

### C. Orphan scan (reverse scan)

```
appindex.Build() → installed app identifiers
ReverseScanner.FindOrphans()
  → for each ReverseScanPath top-level entry:
      skip if conditions.ShouldSkipReverse
      skip if match.MatchesAnyApp (any installed app)
      categorize → Item
```

---

## Scanning Framework

### Path lists (`internal/paths/locations.go`)

Two canonical lists ported from Pearcleaner `Locations.swift`:

- **ForwardScanPaths** — ~50 locations for app-related file discovery
- **ReverseScanPaths** — subset focused on orphan-prone directories

**Depth rule:** `Library` roots scan depth 2 (vendor subfolders like `/Library/Microsoft/`); other roots depth 1.

### Matching pipeline

```
Input: item name + full path + target app (or installed app list)
  1. conditions.MatchesConditionExclude → reject
  2. conditions.MatchesConditionInclude → accept
  3. match.MatchesApp / MatchesAnyApp with Sensitivity tier
  4. Minimum 5-character normalized match length
Output: include or exclude
```

### Size calculation

Primary: macOS `du -sk` (fast). Fallback: `filepath.Walk` byte sum.

---

## Cleaning Framework

### Safety pipeline

```
For each Item:
  1. isSafePath() — block root, $HOME, non-allowlisted prefixes
  2. DryRun? → count only
  3. moveToTrash() → ~/.Trash/<bundle>/<basename>
  4. Record CleanResult { Deleted, Failed, TrashPath }
```

### Allowed deletion prefixes

- `~/Library/**`
- `~/.Trash/**`
- `*.app` bundles under `/Applications`, `~/Applications`
- `/usr/local/**`, `/Library/**`, `/private/var/db/receipts/**`, `/private/tmp/**`
- `~/Documents/**`, `~/Desktop/**`, `~/.config/**`

### Blocked paths

- `/`, `$HOME`, `/System`, `/Applications` (root itself), `/Users`, `/private`

---

## Extension Points (Future Modules)

Design new features as pluggable scan modules implementing:

```go
type Module interface {
    Name() string
    Scan(ctx context.Context) ([]models.Item, error)
}
```

Candidate modules (ROADMAP):

| Module | Package (proposed) |
|---|---|
| Homebrew cleanup | `internal/modules/brew` |
| Developer caches (npm, cargo, Xcode) | `internal/modules/dev` |
| PKG receipt uninstall | `internal/modules/pkg` |
| Trash Sentinel watcher | `internal/automation/sentinel` |
| Scheduled scan | `internal/automation/schedule` |
| LaunchAgent helpers | `internal/automation/launchd` |
| macOS notifications | `internal/automation/notify` |

Register modules in `Scanner.ScanAll()` when ready.

---

## GUI Architecture

Fyne single-window app (`internal/gui/app.go`):

- **State:** `Scanner`, category entries, item checkboxes, busy flag
- **Scan:** background goroutine → `fyne.Do()` UI update
- **Clean:** selected items → `Cleaner` → rescan → status message with Trash path
- **Patterns:** Pearcleaner transparency — expandable path lists, Finder reveal, select-all

CLI and GUI **must share** `scan` and `clean` packages — no duplicated logic.

---

## Build & Distribution

```
make build     → kclean binary
make app       → K-Cleaner.app (Info.plist + icon + embedded binary)
make test      → go test ./...
```

CI (`.github/workflows/test.yml`): `go test` + `go build` on `macos-latest`.

---

## Testing Strategy

| Layer | Approach |
|---|---|
| `match` | Unit tests for `PearFormat`, strict matching |
| `clean` | Temp dir tests for `isSafePath`, dry-run |
| `scan` | Integration tests with fixture dirs (future) |
| `appinfo` | Mock plist paths (future) |

Run: `go test ./...`

---

## Dependencies

| Dependency | Purpose |
|---|---|
| `fyne.io/fyne/v2` | Cross-platform GUI (macOS target) |
| macOS CLI tools | `plutil`, `codesign`, `du`, `pgrep`, `open`, `iconutil` |

No runtime dependency on Pearcleaner or Swift.

---

*See also: [DECISIONS.md](./DECISIONS.md) · [ROADMAP.md](./ROADMAP.md) · [CLI.md](./CLI.md)*
