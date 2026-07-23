---
name: context-offload
description: Create, refresh, or resume a compact `context.md` handoff for the current task. Use when a session is becoming large, before changing models or agents, before pausing overnight, or when the user asks to compact, summarize, checkpoint, offload, or resume context. Preserve only verified task state and the next action; do not use it as a durable decision log.
---

# Context Offload

## Purpose

Write a dense, replaceable handoff that lets the next agent continue the current task
without rereading the conversation.

## Use when

- context is becoming difficult to navigate;
- switching models, agents, or work sessions;
- pausing work or resuming a prior task;
- the user asks to compact, summarize, checkpoint, or offload context.

## Do not use when

- the repository already uses `checkpoint` and needs durable decisions and assumptions;
- the request is only to report status in chat;
- the handoff would contain secrets, credentials, or sensitive production data.

## Default access and modes

Documentation-only edit mode. Use **create or refresh** to replace stale
`context.md`; use **resume** to verify it against the repository before continuing.

## Workflow

1. Read the newest request and inspect only the repository state needed to continue:
   working tree, branch, relevant files, and active validation or services.
2. Treat the repository as the source of truth. Separate facts and evidence from
   assumptions, unknowns, risks, and blockers.
3. Replace stale `context.md` at the repository root. Include the current goal,
   state to preserve, completed evidence, validation, and the first next action.
4. In resume mode, compare recorded branch, files, and validation state with the
   repository. Correct stale claims before acting.

## Validation

Verify the branch, working tree, changed files, and recorded check outcomes. Record
outcomes rather than command transcripts. State anything not run or no longer current.

## Output contract

Return:

- **Summary** — mode and `context.md` path;
- **Evidence** — verified task and repository state;
- **Validation** — current and stale results;
- **Risks** — assumptions, unknowns, and blockers;
- **Next action** — one immediate, ordered action.

## Stop or hand back

Stop when the required repository access, task scope, or sensitive-data treatment is
unclear. Hand durable project knowledge to `checkpoint` and the task itself to its
primary workflow.

## Safeguards

Keep the file factual, short, and replaceable. Preserve user edits; do not let a
checkpoint override current repository state.

## `context.md` shape

```markdown
# Context checkpoint

Updated: YYYY-MM-DD HH:mm TZ
Workspace: /absolute/path

## Current goal

- Newest request:
- Desired outcome:

## State to preserve

- Branch and working tree:
- Relevant files and constraints:

## Evidence and completed work

- Finding or change:
  - Evidence:

## Validation

- Passed:
- Not run or stale:

## Risks, assumptions, and blockers

- Item:

## Next action

1. First incomplete action.
```
