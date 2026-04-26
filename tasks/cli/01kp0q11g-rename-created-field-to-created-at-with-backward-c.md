---
id: "01kp0q11g"
title: "Rename 'created' field to 'created_at' with backward compatibility"
status: completed
priority: medium
dependencies: []
tags: []
created: 2026-04-12
---

# Rename 'created' field to 'created_at' with backward compatibility

## Objective

Rename the `created` frontmatter field to `created_at` across the taskmd specification and CLI codebase. The old `created` field must remain supported for backward compatibility — existing task files should continue to work without modification.

## Tasks

- [x] Update the taskmd specification (`docs/taskmd_specification.md`) to document `created_at` as the canonical field name, with `created` noted as a deprecated alias
- [x] Update the parser to accept both `created_at` and `created` (prefer `created_at` when both are present)
- [x] Update the `taskmd add` command to emit `created_at` in newly created task files
- [x] Update any templates that reference the `created` field to use `created_at`
- [x] Update the validator to accept both field names without errors
- [x] Update output/display logic (list, board, etc.) to use `created_at` internally
- [x] Sync spec copies: run `make sync-spec` after updating the canonical spec
- [x] Add tests for backward compatibility (task files with `created` still parse correctly)
- [x] Add tests for new `created_at` field parsing and output
- [x] Update CLAUDE.md task format examples if they reference `created`

## Acceptance Criteria

- `created_at` is the documented canonical field name in the specification
- `created` is listed as a deprecated alias that still works
- Newly created tasks (via `taskmd add`) use `created_at`
- Existing task files with `created` parse and validate without errors
- All existing tests pass; new tests cover both field names
- `taskmd validate` accepts both `created` and `created_at`
- Spec copies are in sync (`make sync-spec` has been run)
