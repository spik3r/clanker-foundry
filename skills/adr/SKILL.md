---
name: adr
description: Create or update an Architecture Decision Record for a meaningful technical choice. Use when a team must record context, options, a decision, trade-offs, consequences, rollout, or reversal criteria. Keep ADRs concise, factual, durable, and separate the decision from implementation detail.
---

# Architecture Decision Record

Write a durable record of a decision, not a design essay or meeting transcript.

## Decide whether an ADR is warranted

Use an ADR when the choice:

- changes architecture, boundaries, data ownership, security, deployment, or core technology;
- has meaningful alternatives and trade-offs;
- will affect future engineers;
- is hard or costly to reverse;
- resolves repeated debate.

Do not create an ADR for routine implementation details, style choices already covered by standards, or temporary experiments unless the experiment itself changes architecture.

## Gather evidence

Identify:

- decision owner and status;
- date;
- problem and forces;
- constraints and assumptions;
- options seriously considered;
- evidence, prototypes, incidents, benchmarks, or requirements;
- dependencies and affected systems;
- reversibility and review triggers.

Do not invent consensus, alternatives, metrics, or dates. Use placeholders when required facts are unavailable.

## ADR structure

Use the repository's existing ADR template and numbering if present. Otherwise use:

```markdown
# ADR-NNN: Decision title

- Status: Proposed | Accepted | Superseded | Deprecated
- Date: YYYY-MM-DD
- Decision owners: [names or team]
- Supersedes: [ADR, if any]
- Superseded by: [ADR, if any]

## Context

What problem requires a decision? Include constraints and forces, not a long history.

## Decision drivers

- Concrete requirement or constraint
- Concrete requirement or constraint

## Considered options

### Option 1: Name

Description, advantages, disadvantages, and material risks.

### Option 2: Name

Description, advantages, disadvantages, and material risks.

## Decision

State the selected option and the key reason in direct terms.

## Consequences

### Positive

- Expected benefit

### Negative

- Accepted cost or limitation

### Neutral or follow-on

- Migration, operational, documentation, or ownership work

## Validation and review

How the team will know the decision works, and what event or date should trigger review.
```

## Writing rules

- Keep facts, assumptions, and opinions distinct.
- Explain rejected options fairly.
- Record trade-offs rather than claiming one option is universally best.
- Avoid implementation detail that will become stale unless it is essential to the decision.
- Link supporting material instead of copying it.
- Use plain words and established technical terms.
- Do not use achievement language.

## Updating an ADR

Do not rewrite accepted history to make it look cleaner. Add a note, mark it superseded, or create a new ADR when the decision changes materially.
