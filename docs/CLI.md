# K-Cleaner CLI Reference

Binary name: **`kclean`**

Pearcleaner users: map `pear` → `kclean` for equivalent commands.

---

## General

| Command | Description |
|---|---|
| `kclean` | Open GUI |
| `kclean gui` | Open GUI |
| `kclean version` | Print version |
| `kclean help` | Print usage |
| `kclean categories` | List cleanup categories |
| `kclean permissions` | Check Full Disk Access status |

---

## System Cleanup

### Scan

```bash
kclean scan [--verbose]
```

Scans all categories and prints reclaimable space summary.

### Clean

```bash
kclean clean --all [--dry-run] [--yes] [--elevated]
kclean clean --category browser-cache,system-cache [--dry-run] [--yes] [--elevated]
```

Moves matched items to Trash. Requires `--all` or `--category`.

**Categories:** `browser-cache`, `system-cache`, `logs`, `orphan-support`, `orphan-preferences`, `orphan-containers`, `orphan-saved-state`, `orphan-launch-agents`, `trash`

---

## App Uninstall (Pearcleaner-style)

### List related files

```bash
kclean list /Applications/MyApp.app [--sensitivity enhanced] [--verbose]
```

### Uninstall app bundle only

```bash
kclean uninstall /Applications/MyApp.app [--yes] [--dry-run] [--elevated]
```

### Uninstall app + all related files

```bash
kclean uninstall-all /Applications/MyApp.app [--yes] [--dry-run] [--sensitivity enhanced] [--elevated]
```

**Sensitivity:** `strict` | `enhanced` (default) | `deep`

---

## Orphan Management

```bash
kclean list-orphaned [--verbose] [--sensitivity enhanced]
kclean remove-orphaned [--yes] [--dry-run] [--sensitivity enhanced]
```

---

## Automation (Phase 4)

```bash
kclean sentinel install|uninstall|status|run
kclean schedule install [--daily|--weekly]|uninstall|status|run
```

See [AUTOMATION.md](AUTOMATION.md) for behavior and logs.

---

## Flags

| Flag | Commands | Description |
|---|---|---|
| `--verbose` | scan, list, list-orphaned | Show individual file paths |
| `--dry-run` | clean, uninstall*, remove-orphaned | Preview without moving files |
| `--yes`, `-y` | clean, uninstall*, remove-orphaned | Skip confirmation prompt |
| `--sensitivity` | list, uninstall*, list-orphaned, remove-orphaned | Match strictness |
| `--all` | clean | All categories |
| `--elevated` | clean, uninstall* | Retry permission failures with admin password |

---

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Error (invalid args, scan failure, user cancel not an error) |

---

## Examples

```bash
# Preview full system cleanup
kclean clean --all --dry-run

# Clean browser cache only, auto-confirm
kclean clean --category browser-cache --yes

# Deep scan for app leftovers
kclean list /Applications/Slack.app --sensitivity deep --verbose

# Remove all orphans
kclean remove-orphaned --yes
```
