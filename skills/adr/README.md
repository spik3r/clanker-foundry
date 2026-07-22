# adr

Creates or updates an Architecture Decision Record: a durable, concise record of
a meaningful technical choice — context, options, decision, trade-offs,
consequences, and what should trigger a revisit. Writes documentation only.

## When to use

Use an ADR when the choice:

- changes architecture, boundaries, data ownership, security, deployment, or
  core technology;
- has real alternatives and trade-offs;
- will affect future engineers;
- is hard or costly to reverse;
- resolves repeated debate.

Do not use for routine implementation details, style choices already covered by
standards, or temporary experiments.

## How it behaves

- Gathers evidence before writing: decision owner, constraints, options
  seriously considered, reversibility. Uses placeholders for unknown facts —
  never invents dates, metrics, or consensus.
- Uses the repository's existing ADR template and numbering when present.
- Keeps facts, assumptions, and opinions distinct; explains rejected options
  fairly; records trade-offs instead of declaring a universal winner.
- Does not rewrite accepted history: updates arrive as notes, superseded marks,
  or a new ADR.

## Output

An `ADR-NNN` document: status, date, owners → context → decision drivers →
considered options → decision → consequences (positive, negative, follow-on) →
validation and review triggers.

## Example invocations

```text
Use adr to record our decision to use managed identity.
Use adr to supersede ADR-003 with the move to queue-based reconciliation.
```

## Related

- `templates/adr/adr-template.md` — the default structure the skill falls back to
- `architecture-review` — reviews often produce the decisions worth recording
- `checkpoint` — its `decisions.md` log links to ADRs rather than duplicating them

## Installation

See the repository root `README.md`.
