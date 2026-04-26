---
title: "Add history command to show task status transitions"
id: "01kp0pxrz"
status: pending
priority: medium
type: feature
tags: ["cli"]
created: "2026-04-12"
---

# Add history command to show task status transitions

## Objective

Add a `taskmd history <task-id>` command that derives a timeline of status transitions from git history. Instead of storing transition data in frontmatter (which bloats task files and duplicates git), this command parses `git log -p` diffs on the task file to extract status field changes and presents them as a clean, chronological timeline.

Example output:
```
$ taskmd history cli-049
2026-04-12  pending ← cancelled     (by driangle)
2026-04-10  cancelled ← in-progress (by driangle)
2026-04-08  in-progress ← pending   (by driangle)
2026-04-05  created                  (by driangle)
```

## Tasks

- [ ] Add `internal/cli/history.go` with `historyCmd` registered on `rootCmd`
- [ ] Parse `git log -p --follow -- <task-file>` output to extract frontmatter status changes per commit
- [ ] Extract commit author and date for each transition
- [ ] Detect the initial "created" event from the first commit that introduced the file
- [ ] Support `--format` flag (table default, json, yaml) using existing output helpers
- [ ] Support `--field` flag to track fields other than status (e.g. priority, effort)
- [ ] Handle edge cases: task file not tracked by git, no status changes, file renames
- [ ] Add comprehensive tests in `internal/cli/history_test.go`

## Acceptance Criteria

- `taskmd history <task-id>` shows status transitions with dates and authors
- Output includes the initial "created" event
- Reopening events (completed→pending, cancelled→pending) are clearly visible
- `--format json` produces machine-readable output
- `--field priority` tracks priority changes instead of status
- Graceful error when run outside a git repo or on an untracked file
- Tests cover: happy path, multiple transitions, file with no changes, JSON output
