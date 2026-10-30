---
name: backlog-maintenance
description: Review or reconcile an existing backlog for stale, duplicate, completed, blocked or superseded work. Use for backlog upkeep and task closeout audits; not to implement tasks or generate speculative project work.
---

# Backlog Maintenance

Review the existing canonical task store. Default to findings and proposed changes
when asked to review; change records only when upkeep or edits are requested.
Follow project states and keep ownership in task records rather than a new index.

## Reconcile evidence

1. Read the requested queue, current project direction and relevant task records.
   Bound the review to the requested area; do not expand it into a project rewrite.
2. For each candidate, compare its stated outcome with current implementation,
   acceptance checks, validation and integration evidence. Record the revision or
   source checked. Old checkboxes and merged PR titles alone are not completion proof.
3. Classify using the existing lifecycle. Distinguish actionable work, an unresolved
   blocker with a resolver, completed work, superseded/cancelled work and reference
   material. Age flags a record for review; it does not release ownership.
4. Search before adding a missing task. Add one only for a concrete problem supported
   by evidence, with scope, dependencies, acceptance checks and a next action.
   Leave unselected ideas outside the ready queue.

## Update within scope

- Close completed tasks with outcome, implementation links, actual validation and
  remaining limits. Keep a task open when required integrated checks remain.
- For duplicates or superseded work, choose the canonical survivor and link both
  ways where useful. Preserve IDs, useful negative findings and provenance.
- Check code, docs, tests and task links before moving or deleting records. Prefer
  retaining stable completed records; use redirects when a move is necessary.
- Do not clear active claims or reassign owners without checking with those owners.
  Record unresolved ownership as a finding if coordination is unavailable.
- Keep summaries as links or generated views, not a second hand-maintained source
  for status. Do not reorganise an external tracker without authorisation.

## Validate and report

Check changed links, dependencies, IDs and lifecycle fields with the project's
available checks. State evidence gaps rather than inferring completion from absence.
Report records changed or proposed, evidence, unresolved items and next owner/action.
Distinguish backlog edits from product implementation and from unpublished proposals.
