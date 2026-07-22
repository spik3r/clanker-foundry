# checkpoint

Preserves engineering continuity across long sessions, model switches, agent
handoffs, and interrupted work. Creates a curated project handoff rather than
summarising the full conversation.

## What it manages

Inside the target repository:

```text
checkpoints/
├── current.md
├── decisions.md
└── assumptions.md
```

- `current.md` — the active, replaceable handoff: objective, repository state,
  completed work, investigations not to repeat, validation status, next steps,
  risks, and a ready-to-use resume prompt.
- `decisions.md` — a durable log of meaningful decisions and rejected
  alternatives. Not a substitute for an ADR when the project uses ADRs.
- `assumptions.md` — a durable register of unverified beliefs, kept visible
  until verified, disproved, or superseded.

## Why three files

The information has different lifespans:

| File | Lifespan | Update pattern |
|---|---|---|
| `current.md` | One active task or handoff | Replace often |
| `decisions.md` | Months or years | Append selectively |
| `assumptions.md` | Until verified or disproved | Update status over time |

Keeping them separate stops the active checkpoint from growing forever while
preserving what future agents need.

## Modes

**Create or refresh** — after a long session or before a handoff:

```text
Use checkpoint to capture the current repository state and next steps.
```

**Resume** — at the start of a new session:

```text
Use checkpoint in resume mode, verify it against the repository, and continue
from the first incomplete next step.
```

The repository is the source of truth; the skill checks the checkpoint for stale
branch, file, and validation information before relying on it.

**Close** — when the task is finished or abandoned:

```text
Close the current checkpoint, resolve any affected assumptions, and retain any
durable decisions.
```

## The resume prompt

Every `current.md` ends with a self-contained restart instruction telling the
next agent to read the checkpoint files, verify them against the working tree,
preserve user edits, continue from the first unfinished step, avoid repeating
completed investigation, and re-run stale validation. The checkpoint works even
when the next agent has none of the previous conversation.

## Token impact

The skill does not shrink an active conversation. It reduces future token use by
avoiding re-reading prior conversation, repeating repository discovery,
reopening settled decisions, repeating failed investigations, and reconstructing
validation status. Invoke at meaningful boundaries, not after every small edit.

## Recommended invocation points

- before switching models
- before handing work to another agent
- at the end of a long session
- before stopping for the day
- before opening a PR
- after reaching a major decision
- when the working context has become hard to navigate

## Suggested `.gitignore` policy for target repos

- Commit everything when checkpoints hold team-useful, non-sensitive knowledge
  (shared long-running projects, complex migrations).
- Balanced: ignore `checkpoints/current.md` only; commit the durable files.
- Ignore `checkpoints/` entirely when files may contain local paths, temporary
  state, or sensitive material.

Never store secrets, credentials, tokens, private keys, or sensitive production
data in checkpoint files.

## Relationship to other documents

- `AGENTS.md`: standing rules and repository conventions
- `checkpoints/current.md`: current task state
- `checkpoints/decisions.md`: concise durable choices
- `checkpoints/assumptions.md`: unresolved beliefs
- ADRs: formal architectural decisions (see the `adr` skill)
- Issues and PRs: work tracking and review history

Checkpoints link to these sources rather than copying them.

## Installation

See the repository root `README.md`.
