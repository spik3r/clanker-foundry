---
name: project-memory
description: Use when starting, resuming, handing off or finishing repository changes that need durable task context. Record local implementation evidence and validate committed changes; read-only lookups do not need events. Central task records remain authoritative for ownership.
---

# Project Memory

Keep enough repository-local evidence to resume work: intent, paths, validation,
result and next action. Link the canonical task when one exists. Do not copy its
ownership or claims into a second ledger. Check that task and existing PRs before
starting overlapping work.

Never record credentials, tokens, private conversations, personal account data,
raw trading data, or raw logs. Record concise technical facts and validation
commands/results. An event is a declaration, not proof that a test actually ran.

## Start and resume

Examples assume this skill is installed at `skills/project-memory`. For a global
installation, use the installed script's full path while inside the target repo.
All `--files` values are repository-root-relative file paths.

1. Inspect Git, the canonical task and current memory:
   `node skills/project-memory/scripts/project-memory.mjs status`
2. Start one coherent task before editing:
   `node skills/project-memory/scripts/project-memory.mjs start --id TASK-ID --summary "Intended change" --next "Implement and validate" --files path/to/file`
   Add `--task https://...` for the canonical task's verified URL.
3. For interrupted work, use its latest `next` field and inspect Git before
   resuming. An active task does not need another start. To record a handoff or
   change its scope, use `checkpoint` with `--id`, `--summary`, `--next` and,
   when scope changes, `--files` containing the full new scope.

`status --stale-days 7` marks old active tasks for review. It does not abandon
them or block unrelated work. Only the canonical owner can decide a claim's
status. `abandon --id TASK-ID --summary "Why work stopped" --next "Owner action"`
closes local implementation work without providing change coverage or altering
the central task. A later `start` opens a new cycle after complete or abandon.

## Finish and commit

1. Run the relevant checks against the final code and tests.
2. Stage only this task's explicit changed paths, including deletions. Preserve
   other workers' staged changes; do not use an unscoped `git add -A`.
3. Record completion:
   `node skills/project-memory/scripts/project-memory.mjs complete --id TASK-ID --summary "Result" --validation "Command and observed result" --next "Review PR"`
   The command reuses the active scope unless `--files` gives a narrower final
   file list. Expand scope with a checkpoint before completing it.
4. Stage `.project-memory/tasks.jsonl` and commit it with the code and tests.
5. Validate the committed PR diff:
   `node skills/project-memory/scripts/project-memory.mjs validate --base origin/main`

Completion snapshots staged Git blobs and file modes. It rejects unstaged edits
in the listed paths. This handles line-ending normalization, symlinks and modes
without storing file contents. Deleted paths use a null snapshot. Include both
old and new paths for a rename. If code or tests change after completion, start a
new cycle, rerun the checks, stage the final paths and complete again.

File-to-directory changes need the former file path plus every changed child;
directory-to-file changes need the removed children plus the new file path.
Directories themselves do not cover their children. Repository paths must use
valid UTF-8; other filename bytes fail closed rather than being replaced.

The exact ledger path is exempt from file fingerprints, so appending the event
and committing everything together has no self-reference. Other files under
`.project-memory/` are ordinary changed paths and still need coverage.

## Validation contract

- Read the committed HEAD and its merge base with `--base`; working-tree ledger
  edits cannot satisfy a committed change.
- Preserve the merge-base ledger bytes as an exact prefix. Append events; do
  not rewrite or reorder history. During a rebase, keep the target ledger
  intact and append the branch's new events, retaining each task's order.
- Validate event fields, task-local time/order, scope and legal transitions.
  Start or checkpoint is active. Only a final completion supplies coverage;
  abandon and unfinished cycles do not.
- Each changed non-ledger file needs a newly appended final completion whose
  recorded blob and mode still match HEAD. Historical completed tasks cannot
  cover later edits. A current completion can finish a historical open start.
- Another task can supersede an older result on the same path, but at least
  one fresh final completion must match the actual committed file.

CI checks declarations and committed content, not truthfulness or actual wall
clock edit order. It does not prove that a start preceded a human's edits, that
declared tests ran, or that code is secure. Review and ordinary tests still
apply. The job prevents merging only when repository rules require its success;
adding a workflow does not create a branch-protection rule.

## Existing ledgers

Existing unversioned start/complete records may remain unchanged in the base
history. All newly appended records use version 1. No history is invented or
backfilled. If legacy scope included the ledger itself, record a checkpoint
with the real project paths before completing. History that already violates
event order needs an explicit reviewed migration, not automatic rewriting.
