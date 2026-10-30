---
name: agent-coordination
description: Coordinate ownership and safe handoffs when several people or agents may work on the same project. Use before concurrent edits, when checking overlapping work, or when transferring a claimed slice. Does not itself launch workers or authorise publication.
---

# Agent Coordination

Coordinate through the project's canonical task record. A plan, local event log,
worker message or saved snapshot is context, not a live ownership lock.

## Before editing

1. Inspect the branch, staged and unstaged diffs, and untracked files. Preserve
   unrelated work, including edits within a path you also need.
2. Find the matching task, current claims and open PRs at the shared source. Check
   both outcome and write scope; different filenames can still change one interface.
   If access is unavailable, report what could not be checked. Do not claim a shared
   reservation from a private file or stale snapshot.
3. Reuse existing work. Review the current PR or choose a disjoint slice instead of
   rebuilding it. Resolve overlapping ownership before making competing edits.
4. Record a bounded claim in the canonical task: owner, intended outcome, exact paths,
   exclusions, branch/base, dependencies, acceptance checks and integration owner.
   Use the project's shared update or claim mechanism; recheck after publication.
   Where simultaneous claims cannot be resolved atomically, resolve the race with
   the owners rather than treating either local claim as exclusive.

Tiny solo changes and read-only checks do not need a new claim unless project rules
require one. Do not add a shared claims table alongside the canonical task records.

## Work and handoff

- Use an isolated worktree from a verified base for independent work when the shared
  checkout is dirty. Do not stash, reset, clean or move someone else's edits.
  Git stash is shared across worktrees.
- Give a worker the task reference, outcome, exact writable paths, exclusions,
  inputs, dependencies, validation, allowed actions and integration owner. Delegation
  must already be authorised; this procedure does not require spawning workers.
- Return changed paths, branch/revision, checks and results, unresolved risks, running
  processes and the first next action. Keep transcripts and secrets out of handoffs.
- For transfer, confirm the receiving owner's acceptance through an authorised
  channel, then update the shared record. Do not send messages without authority or
  silently expire another owner's claim based on age.

## Integration and closeout

Compare the complete diff with the intended current base and task scope. Check for
already-landed changes, unrelated commits and unexplained test deletions. If the
change is superseded, record that instead of merging a duplicate.

Stage named paths and inspect the full staged diff; a named file can contain another
worker's edits. Run whitespace and relevant checks, recording the revision tested.
Follow required CI and branch rules. Publication, merge and deployment each need
applicable authority; do not bypass checks to close a task.

Record integration evidence in the canonical task and release only the completed
scope. Preserve dependencies that still need verification. A local completion event
or worker report alone cannot close an integrated outcome.

Return the current claim reference, scope, validation, unresolved overlap or blocker,
and next owner/action. Clearly identify any update that remains local.
