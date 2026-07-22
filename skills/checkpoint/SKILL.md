---
name: checkpoint
description: Create, refresh, or resume a durable engineering handoff. Use at the end of a long session, before switching models, before handing work to another agent, before stopping for the day, before opening a pull request, when context is becoming large, or when the user asks to checkpoint, compact, resume, continue, or offload context.
---

# Checkpoint

Maintain a small set of durable project notes so another engineer or agent can continue without reading the full conversation.

The goal is continuity, not conversation compression.

## Files

Use the repository's existing handoff convention when one is clearly established. Otherwise use:

```text
checkpoints/
├── current.md
├── decisions.md
└── assumptions.md
```

### `current.md`

The active handoff. Replace it when creating a new checkpoint.

It records:

- the current objective;
- repository and working-tree state;
- completed work;
- relevant evidence;
- validation status;
- the next actions;
- immediate risks and blockers;
- a resume prompt.

### `decisions.md`

A durable decision log. Preserve previous entries.

Append or update this file only when a decision is important enough that a later engineer might otherwise reopen it.

Examples:

- architecture and dependency choices;
- ownership or boundary decisions;
- compatibility decisions;
- security decisions;
- approaches deliberately rejected;
- trade-offs accepted by the team.

Do not copy every implementation choice into this file.

### `assumptions.md`

A durable register of unverified beliefs and open dependencies.

Preserve unresolved entries. Mark entries verified, disproved, or superseded rather than silently deleting useful history.

Examples:

- external API behaviour not yet confirmed;
- expected production configuration;
- data-shape or migration assumptions;
- performance expectations;
- ownership or consumer assumptions;
- questions that could materially change the implementation.

## Modes

Infer the mode from the user's request.

### Create or refresh

Create or replace `checkpoints/current.md`.

Update `decisions.md` and `assumptions.md` only when the session produced durable information that belongs there.

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

Review:

- the newest user request;
- applicable `AGENTS.md`, `CLAUDE.md`, or project instructions;
- `git status --short`;
- branch and comparison target when relevant;
- important changed files;
- user edits that must not be overwritten;
- known active commands or services;
- validation already performed;
- decisions, rejected approaches, blockers, and unresolved assumptions.

Do not re-investigate unrelated code merely to make the checkpoint look complete.

### 2. Separate information by lifespan

Write session state to `current.md`.

Write long-lived choices to `decisions.md`.

Write unverified beliefs and material questions to `assumptions.md`.

Do not duplicate the same detail across all three files. Link between them when useful.

### 3. Replace stale current state

Replace `current.md` instead of appending indefinitely.

Preserve only information that still affects continuation.

Remove:

- completed next steps that no longer matter;
- stale branch or process state;
- obsolete command output;
- narrative history;
- repeated facts already held in durable files.

### 4. Record validation accurately

Record outcomes rather than a transcript of commands.

Good:

- Unit tests: passed
- Type checking: passed
- Integration tests: not run
- Production behaviour: not verified

Include the command only when the exact command is needed for reproduction or continuation.

Never claim validation that did not occur.

### 5. Add a resume prompt

End `current.md` with a short prompt that a new agent can follow directly.

It must tell the next agent to:

1. read the checkpoint files;
2. verify the repository still matches them;
3. preserve listed user changes;
4. start with the first incomplete next step;
5. avoid repeating completed investigations without new evidence;
6. revalidate stale results before relying on them.

## `current.md` template

```markdown
# Project Checkpoint

Updated: YYYY-MM-DD HH:mm TZ

## Current objective

- Newest request:
- Desired outcome:
- Scope:
- Non-goals:

## Repository state

- Workspace:
- Branch:
- Comparison target:
- Important modified files:
- User edits to preserve:
- Running services or commands:

## Completed

- Change or finding:
- Evidence:

## Already investigated

- Investigation:
  - Result:
  - Evidence:
  - Do not repeat unless:

## Validation

- Unit tests:
- Integration tests:
- Type checking:
- Lint:
- Build:
- Manual verification:
- Not yet verified:

## Next steps

1. Immediate next action.
2. Next validation step.
3. Cleanup, documentation, or follow-up.

## Immediate risks and blockers

- Risk or blocker:

## Relevant durable notes

- Decision: see `decisions.md#...`
- Assumption: see `assumptions.md#...`

## Resume prompt

Read `checkpoints/current.md`, then review the linked entries in
`checkpoints/decisions.md` and `checkpoints/assumptions.md`.

Verify the branch, working tree, changed files, and validation state against the
repository before acting. Preserve the listed user edits. Continue from the
first incomplete item under **Next steps**. Do not repeat work under
**Already investigated** unless new evidence justifies it. Re-run any validation
that may be stale.
```

## `decisions.md` template

```markdown
# Decision Log

## DEC-001: Short decision title

- Status: active | superseded | reversed
- Date: YYYY-MM-DD
- Scope:
- Related files or components:

### Decision

State the choice directly.

### Reason

Explain the constraint or evidence that led to the choice.

### Alternatives considered

- Alternative:
  - Why it was not selected:

### Consequences

- Benefit:
- Cost or limitation:

### Revisit when

- Trigger that should cause the team to reconsider this decision.

### References

- File, issue, ADR, test, benchmark, or other evidence:
```

Use stable identifiers such as `DEC-001`. Do not renumber old entries.

For a major architectural choice, prefer an ADR if the repository uses ADRs. The decision log may link to the ADR rather than duplicate it.

## `assumptions.md` template

```markdown
# Assumption Register

## ASM-001: Short assumption title

- Status: unverified | verified | disproved | superseded
- Added: YYYY-MM-DD
- Owner:
- Affects:
- Risk if wrong: low | medium | high

### Assumption

State what is believed to be true.

### Basis

State why it currently seems plausible.

### Verification

- Evidence needed:
- How to verify:
- Target date or trigger:

### Resolution

Complete when verified, disproved, or superseded.

### References

- Related file, issue, decision, test, or external dependency:
```

Use stable identifiers such as `ASM-001`. Do not silently delete resolved assumptions.

## Quality check

Before finishing, confirm:

- Could a new engineer identify the first next action?
- Are decisions recorded with reasons?
- Are rejected approaches captured only when repeating them would waste time?
- Are facts, assumptions, and unknowns clearly separated?
- Does validation reflect what actually ran?
- Are user changes protected?
- Is stale context removed?
- Is the resume prompt usable without the original conversation?

## Style

Use short factual bullets.

Prefer file paths, symbols, issue references, test names, metrics, and concrete outcomes over narrative.

Avoid filler, emoji, self-congratulation, and achievement language.

After writing or resuming, tell the user which checkpoint files were read or changed.
