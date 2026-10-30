---
name: task-management
description: Create, scope, update or close actionable project tasks in the existing task store. Use when turning agreed work into tracked tasks or recording task progress and integration evidence; not for ordinary status replies or session checkpoints.
---

# Task Management

Use the project's existing tracker, task files and lifecycle. Do not create a second
backlog because this skill is installed. If no convention exists and tracking is
requested, use repository-local `tasks/` Markdown files as a small default.

## Define the task

1. Read project instructions and find the canonical task store. Search existing tasks
   and open PRs for the same problem, including recently completed work.
2. Reuse the matching record. For new work, name an observable outcome, scope and
   exclusions, acceptance checks, dependencies, and how the result will be validated.
   Keep unselected ideas separate from work ready to implement.
3. Use the tracker's stable identity. For file-backed tasks, choose a project prefix
   and unused ID, then use `tasks/ID-description.md`. The bundled
   [task template](templates/task.md) is a starting point, not a required schema.
   In parallel, reserve the ID through the shared source before relying on it;
   checking a private branch alone does not reserve an ID.
4. Record the accountable owner, affected repositories and integration responsibility.
   Leave unclaimed work unassigned. A worker brief or checkpoint may link the task,
   but must not become another authority for its status or ownership.

## Track progress

Follow existing states. If choosing new ones, use:

| State | Meaning |
|---|---|
| `ready` | Scoped and available; prerequisites are met |
| `in-progress` | An owner has claimed a bounded slice |
| `blocked` | A named dependency prevents the next action |
| `review` | Implementation and validation evidence are ready for review |
| `done` | Acceptance and required integration checks passed |
| `cancelled` | Work was declined, abandoned or superseded; retain the reason |

Record the blocker and its resolver, rather than treating inactivity as a blocker.
When a claim is needed, record owner, exact write scope, branch/PR if any, intended
outcome and next handoff in the canonical task. Verify other active work before
editing. Local claims remain proposals until visible through the shared source.

## Close or hand off

- Record changed behaviour, checked revision, validation commands and actual results,
  limitations, and implementation links. Link evidence rather than copying raw logs.
- Distinguish implementation ready from integrated or deployed. A merged component
  does not finish a task whose acceptance requires another repo, release or rollout.
- Update dependencies and the next owner/action. Release only the finished claim;
  do not clear another worker's active scope. For a transfer, record agreement with
  the receiving owner before treating the transfer as effective.
- Keep stable IDs and completed records. Superseded work links to its replacement.
- Apply only actions authorised by the request. Editing a task does not authorise
  publishing, merging, deploying or changing an external tracker.

## Optional implementation evidence

If a project uses Project Memory, follow its installed source and event contract.
It requires a committed Markdown task under `tasks/ID-description.md` and a nonempty
claim heading; a tracker URL alone is insufficient. Agree on a supported canonical
source before enabling that helper, rather than silently mirroring a live tracker.
Pinned snapshots and local events do not prove current ownership or test execution.

Return the canonical task reference, changed state, evidence and next action. State
unpublished updates or inaccessible tracker checks explicitly.
