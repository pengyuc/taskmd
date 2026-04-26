---
title: "Auto-set completed date in frontmatter on status change"
id: "01kp071jv"
status: completed
priority: medium
type: feature
tags: []
created: "2026-04-12"
---

# Auto-set completed date in frontmatter on status change

## Objective

Automatically set a `completed` date field in task frontmatter when a task transitions to a terminal status (`completed`, `cancelled`), and clear it when transitioning back to a non-terminal status. This mirrors the existing `created` field, giving tasks a visible lifecycle in their metadata without relying on git history.

## Context

- The `created` field already tracks when a task was made, but there's no `completed` counterpart
- After archiving, the only way to find completion dates is git history — which isn't always available or queryable
- Enables reporting and metrics (cycle time, throughput) from frontmatter alone

## Tasks

- [x] Add `completed` field to the task frontmatter schema (using `FlexibleTime`, same as `created`)
- [x] Update the `set` command: when status changes to `completed` or `cancelled`, auto-write `completed: YYYY-MM-DD`
- [x] Update the `set` command: when status changes from a terminal status back to a non-terminal status (`pending`, `in-progress`, `blocked`), clear the `completed` field
- [x] Update the task parser to read/write the `completed` field
- [x] Update the specification (`docs/taskmd_specification.md`) to document the `completed` field
- [x] Run `make sync-spec` to propagate spec changes
- [x] Add unit tests for the set command covering completion date lifecycle
- [x] Add e2e tests verifying `completed` field is set/cleared on status transitions

## Acceptance Criteria

- Running `taskmd set <id> --status completed` writes `completed: YYYY-MM-DD` (today's date) to frontmatter
- Running `taskmd set <id> --done` also writes the `completed` date
- Running `taskmd set <id> --status cancelled` also writes the `completed` date
- Transitioning from `completed` → `pending` (or any non-terminal status) removes the `completed` field
- The `completed` field uses `FlexibleTime`, consistent with `created`
- Existing tasks without a `completed` field continue to parse correctly
- `taskmd validate` accepts tasks with or without the `completed` field
- All new and existing tests pass
