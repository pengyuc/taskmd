---
title: "Add list-compatible filter shortcuts to graph command"
id: "01kpwpa5d"
status: completed
priority: medium
type: feature
tags: ["cli", "graph", "filters"]
created: "2026-04-23"
completed_at: 2026-04-23
---

# Add list-compatible filter shortcuts to graph command

## Objective

Add the `--status`, `--priority`, and `--phase` shortcut flags to the `graph` command, matching the convenience filters available on `list`. Refactor the shared filtering logic so both commands reuse the same code path instead of duplicating filter expansion.

## Context

The `graph` command already supports `--filter key=value` and `--scope`, but lacks the shortcut flags (`--status`, `--priority`, `--phase`) that `list` provides. Users must write `--filter status=pending` on `graph` while `list` allows `--status pending`. The underlying filter expansion logic in `applyListFiltersAndSort` (list.go:228) is coupled to list-specific package vars; it should be extracted into a shared helper that both commands can call.

## Tasks

- [x] Extract a reusable filter-expansion function (shortcut flags → filter expressions, apply filters, apply scope, apply phase) that accepts parameters instead of reading package-level vars
- [x] Add `--status`, `--priority`, and `--phase` shortcut flags to `graphCmd` in `graph.go`
- [x] Wire the graph command to use the shared filter function instead of its inline filtering logic
- [x] Update `list` to call the same shared function
- [x] Add tests for the new shortcut flags on the graph command
- [x] Verify existing list and graph tests still pass

## Acceptance Criteria

- `taskmd graph --status pending` produces the same result as `taskmd graph --filter status=pending`
- `taskmd graph --priority high` produces the same result as `taskmd graph --filter priority=high`
- `taskmd graph --phase <phase>` filters the graph to tasks in that phase
- The shortcut flags compose with existing graph flags (e.g. `--status pending --root 022 --downstream`)
- No duplicated filter-expansion logic between `list` and `graph` — both call a single shared function
- All existing tests for `list` and `graph` continue to pass
