---
name: checkpoint
description: Maintain a durable, multi-file engineering handoff (current state, a decision log, and an assumption register) so another agent or engineer can continue across sessions, model switches, or ownership changes. Use at the end of a long session, before switching models or agents, before stopping for the day, before opening a pull request, or when the user asks to checkpoint, resume, or close out durable project state. For a single ephemeral one-task handoff with no durable decision or assumption history, use context-offload instead.
---

# Checkpoint

Maintain a small set of durable project notes so another engineer or agent can continue without reading the full conversation.

The goal is continuity, not conversation compression. For a compact single-file handoff of one active task with no durable history, use the `context-offload` skill instead.

## Files

Use the repository's existing handoff convention when one is clearly established. Otherwise use:

```text
checkpoints/
├── current.md
├── decisions.md
└── assumptions.md
```

Templates for all three live alongside this skill in [`templates/`](templates/):
[`current.md`](templates/current.md), [`decisions.md`](templates/decisions.md),
and [`assumptions.md`](templates/assumptions.md). Copy the relevant template into the
target repository's `checkpoints/` directory and fill it in.

### `current.md`

The active handoff; see [`templates/current.md`](templates/current.md). Replace it when creating a new checkpoint. It records the current objective, repository and working-tree state, completed work, relevant evidence, validation status, next actions, immediate risks and blockers, and a resume prompt.

### `decisions.md`

A durable decision log; see [`templates/decisions.md`](templates/decisions.md). Preserve previous entries. Append or update only when a decision is important enough that a later engineer might otherwise reopen it: architecture and dependency choices, ownership or boundary decisions, compatibility and security decisions, approaches deliberately rejected, and trade-offs accepted by the team. Do not copy every implementation choice into this file. Use stable identifiers such as `DEC-001` and do not renumber old entries. For a major architectural choice, prefer an ADR if the repository uses ADRs and link to it rather than duplicating it.

### `assumptions.md`

A durable register of unverified beliefs and open dependencies; see [`templates/assumptions.md`](templates/assumptions.md). Examples: external API behaviour not yet confirmed, expected production configuration, data-shape or migration assumptions, performance expectations, ownership or consumer assumptions, and questions that could materially change the implementation. Use stable identifiers such as `ASM-001`. Preserve unresolved entries; mark entries verified, disproved, or superseded rather than silently deleting useful history.

## Modes

Infer the mode from the user's request.

### Create or refresh

Create or replace `checkpoints/current.md`. Update `decisions.md` and `assumptions.md` only when the session produced durable information that belongs there. Follow the workflow below.

### Resume

Use when the user asks to resume, continue, restore context, or pick up previous work.

1. Read `checkpoints/current.md`.
2. Read relevant unresolved items in `decisions.md` and `assumptions.md`.
3. Verify that the repository still matches the recorded branch, files, and validation state.
4. Inspect `git status --short` and relevant recent changes.
5. Identify stale or contradictory checkpoint content.
6. Summarise the current objective, first next action, and any mismatch.
7. Continue only within the user's requested scope.

Do not trust the checkpoint over the repository. The repository is the source of truth for current code state.

### Close

Use when work is finished, merged, abandoned, or replaced.

- Record any final durable decision.
- Resolve or update affected assumptions.
- Mark `current.md` as closed or replace it with a short completed-state checkpoint.
- Do not claim completion unless validation supports it.

## Create or refresh workflow

### 1. Gather only what matters

Review the newest user request; applicable `AGENTS.md`, `CLAUDE.md`, or project instructions; `git status --short`; branch and comparison target when relevant; important changed files; user edits that must not be overwritten; known active commands or services; validation already performed; and decisions, rejected approaches, blockers, and unresolved assumptions.

Do not re-investigate unrelated code merely to make the checkpoint look complete.

### 2. Separate information by lifespan

Write session state to `current.md`, long-lived choices to `decisions.md`, and unverified beliefs and material questions to `assumptions.md`. Do not duplicate the same detail across files; link between them when useful.

### 3. Replace stale current state

Replace `current.md` instead of appending indefinitely. Preserve only information that still affects continuation. Remove completed next steps that no longer matter, stale branch or process state, obsolete command output, narrative history, and facts already held in the durable files.

### 4. Record validation accurately

Record outcomes rather than a transcript of commands, for example:

- Unit tests: passed
- Type checking: passed
- Integration tests: not run
- Production behaviour: not verified

Include a command only when the exact command is needed for reproduction or continuation. Never claim validation that did not occur.

### 5. Add a resume prompt

End `current.md` with a short prompt (see the template) that a new agent can follow directly. It must tell the next agent to read the checkpoint files, verify the repository still matches them, preserve listed user changes, start with the first incomplete next step, avoid repeating completed investigations without new evidence, and revalidate stale results before relying on them.

## Quality check

Before finishing, confirm the checkpoint stands on its own: a new engineer could identify the first next action; decisions carry reasons and rejected approaches are captured only where repeating them would waste time; facts, assumptions, and unknowns are clearly separated; validation reflects what actually ran; user changes are protected; stale context is removed; and the resume prompt is usable without the original conversation.

## Style

Use short factual bullets. Prefer file paths, symbols, issue references, test names, metrics, and concrete outcomes over narrative. Avoid filler, emoji, self-congratulation, and achievement language.

After writing or resuming, tell the user which checkpoint files were read or changed.
