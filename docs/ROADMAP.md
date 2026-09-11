# K-Cleaner Roadmap

---

## Phase 1 — Foundation ✅ (v1.0.0)

Complete.

## Phase 2 — Robustness ✅ (v1.1.0)

Complete.

## Phase 3 — Pearcleaner Parity Modules ✅ (v1.2.0)

Complete (Homebrew out of scope).

## Phase 4 — Automation ✅ (v1.3.0)

- [x] Sentinel-style Trash watcher (offer cleanup when `.app` trashed)
- [x] Scheduled scan via `launchd` plist
- [x] Elevated clean for `/Library` owned files (`--elevated` + admin prompt)

---

## Future ideas

- URL scheme handler (`kclean://`) for Sentinel deep links
- Notarized `.app` bundle with embedded Sentinel
- Threshold configuration for scheduled notifications

## Non-Goals

- Homebrew integration
- Windows / Linux ports
- Cloud sync or telemetry
