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

## Phase 2 — Robustness

- [ ] Integration tests with fixture Library trees
- [ ] Spotlight supplement via `mdfind` for forward scan gaps
- [ ] Parent-path deduplication improvements
- [ ] Exclusion list persistence (user-defined skip paths)
- [ ] `--json` output flag for scripting
- [ ] Code signing + notarization guide

---

## Phase 3 — Pearcleaner Parity Modules

- [ ] Homebrew cache and formula cleanup (`brew cleanup` wrapper)
- [ ] Developer environment caches (npm, cargo, pip, Gradle, Xcode simulators)
- [ ] Launch daemon / privileged helper tool scanner (read-only first)
- [ ] PKG receipt and BOM parsing
- [ ] GUI uninstall tab (pick app → list related → clean)

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
