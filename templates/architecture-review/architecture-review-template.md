# Architecture review: <system or boundary under review>

- Date: <YYYY-MM-DD>
- Reviewers: <names>
- Scope: <what is in and out of scope>
- Trigger: <why this review is happening now>

## Summary

<Three to five sentences: the state of the system, the largest risk, the single most
valuable action.>

## Context

- Purpose of the system: <what it does and for whom>
- Load and scale: <requests, data volume, growth — facts, not projections dressed as facts>
- Constraints: <team size, deadlines, compliance, budget>
- Relevant history: <incidents, past decisions, links to ADRs>

## Findings

Rate each finding: **high** (act now), **medium** (plan), **low** (note).

### Boundaries and coupling

- <finding: where responsibilities blur, where modules know too much about each other>

### Data

- <finding: ownership, consistency, migration risk, PII handling>

### Reliability

- <finding: failure modes, retries, timeouts, single points of failure, observability>

### Security

- <finding: trust boundaries, authn/authz, secrets, exposure. Run `checklists/security.md` for depth>

### Evolution

- <finding: what the next year of change will be hard to do — new consumers, scale, team split>

## Risks accepted

<What the team already knows and has chosen to live with. Distinguish these from
findings — accepted risks need a revisit trigger, not a fix.>

## Recommendations

| # | Action | Addresses | Effort | Priority |
|---|--------|-----------|--------|----------|
| 1 | <specific, scoped action> | <finding> | <S/M/L> | <high/med/low> |

Recommend one path where options exist and say why. Do not list everything that could
ever be improved.

## Open questions

<What could not be answered from the code and docs, and who can answer it.>

---

## How to use this template

- Pair with the `architecture-review` skill; the skill drives the process, this file shapes the output.
- Every finding needs evidence: a path, a trace, an incident, a measure. Mark anything unverified as an assumption.
- Prefer five findings with evidence over twenty observations.
- Delete sections that genuinely do not apply; do not fill them with boilerplate.
