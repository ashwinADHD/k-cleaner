# K-Cleaner Roadmap

Phased delivery plan aligned with Pearcleaner capabilities and K-Cleaner scope decisions.

---

## Phase 1 — Foundation ✅ (v1.0.0)

- [x] Go project scaffold with layered packages
- [x] Category scans (browser cache, system cache, logs, trash)
- [x] Reverse orphan scanner with skip-keywords
- [x] Forward app uninstall scanner
- [x] `pearFormat` matching + sensitivity tiers
- [x] UUID container resolution
- [x] Trash-based deletion
- [x] CLI (Pearcleaner-style commands)
- [x] Fyne GUI with per-item selection
- [x] Full Disk Access check
- [x] Documentation (DECISIONS, ARCHITECTURE)

---

## Phase 2 — Robustness ✅ (v1.1.0)

- [x] Integration tests with fixture Library trees
- [x] Spotlight supplement via `mdfind` for forward scan gaps
- [x] Parent-path deduplication improvements
- [x] Exclusion list persistence (user-defined skip paths)
- [x] `--json` output flag for scripting
- [x] Code signing + notarization guide

---

## Phase 3 — Pearcleaner Parity Modules ✅ (v1.2.0)

- [ ] ~~Homebrew cache and formula cleanup~~ — **Out of scope** (not used)
- [x] Developer environment caches (npm, cargo, pip, Gradle, Xcode simulators)
- [x] Launch daemon / privileged helper tool scanner (read-only)
- [x] PKG receipt and BOM parsing (read-only via `lsbom`)
- [x] GUI uninstall tab (pick app → list related → clean)

---

## Phase 4 — Automation

- [ ] Sentinel-style Trash watcher (offer cleanup when `.app` trashed)
- [ ] Scheduled scan via `launchd` plist template
- [ ] Optional privileged helper for `/Library` owned files

---

## Non-Goals

- Windows / Linux ports (macOS-only by design)
- Cloud sync or telemetry
- Paid subscription model
- Permanent secure erase / DoD wipe
- Homebrew integration
