---
name: persistent-iteration
description: Continue authorised work until the user's stated outcome or time boundary is reached. Use when asked to keep iterating, work until a deadline, resume an ongoing task, or monitor an in-progress result. Do not use to invent unrelated work, broaden permissions, or schedule recurring jobs without authorisation.
---

# Persistent Iteration

Treat a useful slice or commit as a checkpoint when the user requested a larger outcome
or an explicit period of continued work.

## Establish the contract

1. Read the latest request and current project instructions.
2. Identify the bounded objective, authorised actions, observable success criteria,
   and stopping condition. Resolve an ambiguous deadline or timezone before relying on it.
3. Check a trustworthy current time source for time-bounded work.
4. Inspect current task state, repository status when relevant, and active commands.
   Preserve existing user changes and other workers' sessions.
5. Read an existing checkpoint or plan, then verify it against the current source.
   Do not repeat completed work merely because a handoff is incomplete.
6. For monitoring, confirm the result can actually be observed. Distinguish permission
   to watch from permission to change, publish, merge, or deploy.

Do not promise unattended or future execution unless the environment supports it.
Use existing execution or waiting facilities for in-progress work. Creating a new
recurring job requires the user's authorisation and is not a substitute for finishing
the current task.

## Work loop

Repeat while the requested outcome or time boundary remains unmet:

1. Choose the smallest useful next step that advances the authorised objective.
2. Investigate or implement it according to the active task mode.
3. Run meaningful validation and record its result against the state checked.
4. Diagnose recoverable failures and retry within scope. Do not repeat identical
   attempts without new evidence, weaken checks, or suppress failures to appear finished.
5. Recheck current state, dependencies, and time when relevant.
6. Refresh an existing checkpoint after meaningful milestones or before a long pause.
   Keep it concise: completed work, evidence, blocker, and first next action.
7. Continue to the next useful step or wait for the external result being monitored.

Send updates for meaningful results, changed scope, blockers, or decisions. Avoid
narrating every iteration. Commit or publish only when that action is authorised.

## Waiting and resuming

- A pending or unchanged external result is not completion.
- Wait at a cadence suited to the event, its cost, and any deadline. Recheck the
  relevant revision or run so that an old success does not satisfy a new task.
- On resume, verify the actual task, repository, process, and time state before
  continuing. An old checkpoint does not override current evidence.
- Continue independent authorised work while a dependent step is blocked.
- Ask for missing permission or a required decision; elapsed time is not approval.

## Stop discipline

Stop when the requested outcome is established, the user's stated deadline or other
stopping condition is reached, the user cancels or replaces the task, or the next
necessary step requires unavailable access, information, or authorisation.

If no safe useful step remains, report why. Do not manufacture activity to fill a
timebox, and do not claim the requested outcome was achieved when it remains blocked.
Do not stop at an arbitrary iteration count or after one successful commit.

Before closing:

- record passed, failed, blocked, and unrun checks;
- inspect the final working state and note protected or unfinished changes;
- report relevant active processes and their ownership;
- leave a clear next action for any unresolved work.

A deadline does not justify killing another worker's process, abandoning an unsafe
partial change without a handoff, or describing incomplete validation as complete.
