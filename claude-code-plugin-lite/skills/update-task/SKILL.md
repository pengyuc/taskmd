---
name: update-task
description: Update an existing task's fields (status, priority, title, tags, dependencies, etc.). Use when the user wants to modify a task's properties.
allowed-tools: Glob, Read, Edit
---

# Update Task

Update fields of an existing task — no CLI required.

## Instructions

The user's query is in `$ARGUMENTS` (e.g. "set task 042 to high priority and in-progress", "rename task 15 to Fix auth bug", "add tag backend to 042").

1. **Parse the user's input** from `$ARGUMENTS` to extract:
   - The **task ID** (required)
   - The **fields to update** and their new values

2. **Find the task file**:
   - Read `.taskmd.yaml` for custom `dir` (default: `tasks`)
   - Use `Glob` for `<task-dir>/**/*<ID>*.md`
   - If multiple matches, read frontmatter to find the exact ID match
   - If not found, list available tasks

3. **Read the task file** with the `Read` tool

4. **Apply updates using the `Edit` tool**:

   ### Frontmatter field updates
   For each field to change, edit the YAML frontmatter:
   - **status**: Replace the `status: <old>` line with `status: <new>` — valid values: pending, in-progress, completed, in-review, blocked, cancelled
   - **priority**: Replace `priority: <old>` with `priority: <new>` — valid: low, medium, high, critical
   - **effort**: Replace `effort: <old>` with `effort: <new>` — valid: small, medium, large
   - **type**: Replace `type: <old>` with `type: <new>` — valid: feature, bug, improvement, chore, docs
   - **owner**: Replace or add `owner: "<value>"`
   - **phase**: Replace or add `phase: "<value>"` — must match a phase `id` in `.taskmd.yaml`. If the phase doesn't exist yet, add it to the `phases` array in `.taskmd.yaml` first (see `SPEC_REFERENCE.md`)
   - **parent**: Replace or add `parent: "<value>"`
   - **title**: Replace the `title: "..."` line

   ### Array field updates (tags, dependencies, pr)
   - **Add tag**: Add the value to the `tags` array in frontmatter
   - **Remove tag**: Remove the value from the `tags` array
   - **Add dependency**: Add the ID to the `dependencies` array
   - **Remove dependency**: Remove the ID from the array
   - **Add PR**: Add the URL to the `pr` array
   - **Remove PR**: Remove the URL from the array

   If a field doesn't exist in the frontmatter yet, add it before the closing `---`

   ### Body updates
   - **description/subtasks/acceptance criteria**: Edit the markdown body directly

5. **Validate** the changes:
   - Ensure enum values are valid (see SPEC_REFERENCE.md)
   - Ensure the frontmatter is still valid YAML
   - If a field value is invalid, explain the valid options to the user

6. **Confirm** the changes to the user, showing what was updated

See `SPEC_REFERENCE.md` (in the plugin root) for valid field values and frontmatter schema.
