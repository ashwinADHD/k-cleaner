# Git Workflow

K-Cleaner uses a simple feature-branch workflow:

```
feat/<name>  →  main
```

## Process

1. Create a feature branch from `main`:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feat/my-feature
   ```

2. Commit changes on the feature branch.

3. Push the feature branch:
   ```bash
   git push -u origin feat/my-feature
   ```

4. Merge into `main` (via pull request or local merge):
   ```bash
   git checkout main
   git pull origin main
   git merge feat/my-feature --no-ff
   git push origin main
   ```

## Branches

| Branch | Purpose |
|---|---|
| `main` | Stable, release-ready code |
| `feat/*` | Feature and framework work |

Do not push directly to `main` for new work — use `feat/*` first, then merge.
