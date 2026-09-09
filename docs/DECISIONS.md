# K-Cleaner — Product & Engineering Decisions

This document records **what** K-Cleaner is, **why** we built it this way, and **how** key decisions were made. It is the authoritative rationale for the project direction.

---

## 1. What is K-Cleaner?

**K-Cleaner** is a standalone, native macOS utility for reclaiming disk space safely. It combines:

1. **System cleanup** — browser caches, application caches, logs, trash
2. **Orphan detection** — leftover files from apps that were uninstalled without a proper cleaner
3. **App uninstall** — find and remove an `.app` bundle plus all related support files

It ships as:

- A **CLI** binary (`kclean`) for power users and automation
- A **GUI** app (`K-Cleaner.app`) for everyday use with Pearcleaner-style transparency (per-path visibility, checkboxes, Finder reveal)

K-Cleaner is **not** a background daemon, cloud service, or subscription product. It runs locally, shows every path before deletion, and moves files to **Trash** (restorable) rather than permanently erasing them.

---

## 2. Why K-Cleaner Exists

### Problem

macOS accumulates reclaimable data in ways users rarely see:

- Browser and app caches grow silently in `~/Library/Caches`
- Dragging an app to Trash removes the bundle but leaves preferences, containers, launch agents, and support folders
- Generic “cache cleaners” delete known folders blindly without app context
- Commercial tools (e.g. CleanMyMac) are closed-source and subscription-based

### Inspiration: Pearcleaner

[Pearcleaner](https://github.com/alienator88/Pearcleaner) demonstrated that a **fair, transparent, app-aware cleaner** could be built with:

- Forward scanning (given an app → find related files)
- Reverse scanning (given installed apps → find orphans)
- Normalized string matching (`pearFormat`)
- Trash-based deletion with undo
- Per-item selection in the UI

Pearcleaner is Swift/Xcode-based with multiple targets (helper daemon, Sentinel, Finder extension). Development is currently on hold, but its architecture is the best open reference for this problem space.

### Why a new project instead of forking Pearcleaner?

| Factor | Pearcleaner | K-Cleaner decision |
|---|---|---|
| Language | Swift, Xcode multi-target | **Go** — single binary, no Xcode required |
| Distribution | Signed app + SMAppService helper | **Static binary + optional .app bundle** |
| Maintenance | On hold | Active, scoped MVP we control |
| Branding | Pearcleaner | **K-Cleaner** — distinct identity |
| Prior work | — | MacLean prototype validated scan/clean/GUI patterns |

### Why Go?

1. **Single static binary** — easy to build, ship, and run from Terminal
2. **Cross-compilation** — future CI builds without a Mac farm (runtime still macOS-only)
3. **CLI-first** — Pearcleaner already exposes `pear`; Go fits scripting and automation
4. **No privileged helper initially** — use Trash + user permissions; add `sudo`/helper later if needed
5. **Fyne GUI** — native-feeling macOS window without SwiftUI

---

## 3. How We Decided the Architecture

### 3.1 Dual scanner model (from Pearcleaner)

```
Forward scan:  App metadata → walk known paths → match bundle ID / name / entitlements
Reverse scan:  Installed app index → walk leftover-prone dirs → exclude matches → orphans
```

**Why:** Cache-only cleaners miss app leftovers. App-only uninstallers miss general caches. Dual scanners cover both use cases with shared matching logic.

**How:** Implemented in `internal/scan/forward.go` and `internal/scan/reverse.go`, sharing `internal/match` and `internal/appindex`.

### 3.2 Matching: `pearFormat` + sensitivity tiers

**What:** Strip non-alphanumeric characters, lowercase, then compare bundle IDs, app names, and (at deep level) entitlements.

**Why:** File paths use inconsistent naming (`com.google.Chrome`, `Google Chrome`, vendor subfolders). Pearcleaner’s normalization reduces false negatives.

**Sensitivity tiers:**

| Tier | Use case |
|---|---|
| `strict` | Exact bundle ID / name only — safest, fewest false positives |
| `enhanced` | Partial matches, version-stripped names — **default** |
| `deep` | Entitlement-based matching — most complete, higher false-positive risk |

**How:** `internal/match/match.go`, exposed via `--sensitivity` on CLI.

### 3.3 Trash-based deletion (not `rm`)

**Why:**

- Pearcleaner’s undo model builds user trust
- macOS users expect Trash as the recovery path
- Grouped folders (`K-Cleaner_AppName_timestamp`) make bulk restore easy

**How:** `internal/clean/clean.go` uses `os.Rename` into `~/.Trash/K-Cleaner_*`.

**Trade-off:** Cross-volume moves fall back to copy+remove; protected system paths may still fail without elevated permissions.

### 3.4 Safety guards

| Guard | Rationale |
|---|---|
| Path allowlist in `clean.isSafePath` | Never delete `/`, `$HOME`, or arbitrary paths |
| `ProtectedCacheNames` | Skip iCloud, CloudKit, NSURLSession caches |
| `SkipReverseKeywords` | Block `apple`, `icloud`, `coresimulator` orphan false positives |
| Per-app `conditions` | Disambiguate Xcode vs Xcodes, Chrome vs Chromium |
| Browser-running warning | Prevent tab loss when clearing live browser cache |
| Confirmation prompts | Default deny; `--yes` for automation only |
| `--dry-run` | Preview without side effects |

### 3.5 UUID sandbox container resolution

**Problem:** macOS stores sandbox data in `~/Library/Containers/<UUID>/`, not by bundle ID.

**Solution:** Read `.com.apple.containermanagerd.metadata.plist` → `MCMMetadataIdentifier` (Pearcleaner pattern).

**How:** `internal/appinfo/appinfo.go` → `ContainerBundleID()`.

### 3.6 Full Disk Access (FDA)

**Why:** Without FDA, container directories and TCC-protected paths return permission errors; scans are incomplete.

**How:** `internal/permissions/fda.go` probes known protected paths at startup / via `kclean permissions`.

**Decision:** FDA is **recommended**, not required. The app degrades gracefully and tells the user what is missing.

### 3.7 CLI command design (Pearcleaner parity)

Pearcleaner’s embedded CLI uses `pear list`, `pear uninstall-all`, `pear list-orphaned`. K-Cleaner mirrors this:

| Pearcleaner | K-Cleaner |
|---|---|
| `pear list <app>` | `kclean list <app>` |
| `pear uninstall <app>` | `kclean uninstall <app>` |
| `pear uninstall-all <app>` | `kclean uninstall-all <app>` |
| `pear list-orphaned` | `kclean list-orphaned` |
| `pear remove-orphaned` | `kclean remove-orphaned` |

Additional commands: `scan`, `clean`, `permissions`, `categories`, `gui`.

**Why mirror:** Low learning curve for Pearcleaner users; proven command surface.

### 3.8 GUI scope (v1)

**Included:** Scan all categories, expand paths, per-item checkboxes, Finder reveal, Trash-safe clean, browser warning.

**Deferred (see ROADMAP):** Dedicated uninstall tab, Sentinel trash watcher, Homebrew module, PKG receipts, binary lipo.

**Why:** Ship a trustworthy cleanup core before advanced modules.

---

## 4. What We Explicitly Did Not Build (Yet)

| Feature | Reason deferred |
|---|---|
| Privileged XPC helper (PearcleanerHelper) | Go CLI can use Trash for most user data; system paths need later design |
| Sentinel (Trash watcher) | Separate background process; v2 feature |
| Finder Sync extension | Requires `.appex` + code signing |
| Homebrew / PKG managers | Separate domains; add as plugins |
| Code signing / notarization | Requires Apple Developer account; document manual “Right-click → Open” for now |
| Spotlight (`mdfind`) supplement | Adds complexity; directory walk sufficient for v1 |

---

## 5. Naming & Branding

| Item | Value | Rationale |
|---|---|---|
| Product name | **K-Cleaner** | Distinct from MacLean prototype and Pearcleaner |
| CLI binary | `kclean` | Short, Pearcleaner `pear`-style |
| Module path | `github.com/ashwinADHD/k-cleaner` | Matches GitHub remote |
| App bundle ID | `com.ashwin.kcleaner` | macOS identity for GUI app |
| Trash folder prefix | `K-Cleaner_*` | Recognizable, restorable groups |

---

## 6. Success Criteria

K-Cleaner v1 is successful if:

1. **Transparent** — every deletion candidate shows full path and size before action
2. **Restorable** — default action moves to Trash, not permanent delete
3. **Accurate enough** — orphan/uninstall scans find real leftovers without touching Apple system data
4. **Dual interface** — same logic powers CLI and GUI
5. **Documented** — architecture and decisions are readable by future contributors

---

## 7. References

- [Pearcleaner repository](https://github.com/alienator88/Pearcleaner) — primary architectural reference
- [AppCleaner](https://freemacsoft.net/appcleaner/) — original “uninstall app + related files” UX pattern
- Apple Container Manager metadata — `MCMMetadataIdentifier` in sandbox plists
- macOS Full Disk Access — required for complete `~/Library/Containers` visibility

---

*Last updated: September 2026 — K-Cleaner v1.0.0 initial framework*
