# Workflow

## Issue Tracking (bd / beads)

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress
bd close <id>         # Complete work
bd sync               # Sync with git
```

Run `bd onboard` to get started.

## Session Completion

Work is NOT complete until `git push` succeeds.

1. File issues for remaining work
2. Run quality gates: `make test && make lint && make build`
3. Update issue status
4. Push:
   ```bash
   git pull --rebase && bd sync && git push
   ```
5. Verify `git status` shows "up to date with origin"
