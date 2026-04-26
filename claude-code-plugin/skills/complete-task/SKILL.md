---
name: complete-task
description: Mark a task as completed. Use when the user wants to mark a task as done or complete.
allowed-tools: Bash, Read, Edit
---

# Complete Task

Mark a task as completed using the `taskmd` CLI.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077`). If `$ARGUMENTS` is empty or does not contain a task ID, infer the task from conversation context (e.g., the task currently being worked on). If the task cannot be determined, ask the user which task to complete.

1. **Read the task file** to understand the full task scope:
   - Run `taskmd get <ID>` to get the task contents
   - Identify all **subtask checklists** (`- [ ]` / `- [x]` items) in the task body
   - Identify any **acceptance criteria** section

2. **Verify subtasks and acceptance criteria are met**:
   - Review each subtask checklist item — confirm the work has been done
   - Review each acceptance criterion — confirm it is satisfied
   - **Check off** (`- [x]`) any items that are complete but not yet checked off by editing the task file
   - If any items are genuinely incomplete, report them to the user and ask how to proceed — do NOT mark the task as completed

3. **Add a final worklog entry** (if worklogs are enabled):
   - Check `.taskmd.yaml` for `worklogs: true` -- only create worklogs if explicitly enabled; skip this step otherwise
   - Otherwise, find the worklog file at `tasks/<group>/.worklogs/<ID>.md` (or `tasks/.worklogs/<ID>.md`)
   - If a worklog exists, append a timestamped completion summary

4. **Check the workflow mode** in `.taskmd.yaml`:
   - If `workflow: pr-review` is set, use `taskmd set $ARGUMENTS --status in-review` instead of `completed` (note: in pr-review mode, tasks are completed by merging the PR, not by setting status directly)
   - Otherwise (default `solo` mode), run `taskmd set $ARGUMENTS --status completed --verify`
   - The `--verify` flag runs any verification checks defined in the task before applying the status change
   - If verification fails, report the failures to the user

5. Confirm the status change to the user
