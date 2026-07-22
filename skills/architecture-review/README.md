# architecture-review

Assesses a codebase, subsystem, or proposed design: boundaries, coupling, data
flow, failure handling, security, operability, and whether the design can evolve
safely. Evidence-based; recommends incremental change over idealised rewrites.

## When to use

- periodic health checks on a system or subsystem;
- design reviews before building;
- service-boundary and refactor-planning questions;
- maintainability assessments.

Invoke it deliberately — it is a broad instrument, not part of the daily edit
loop. Read-only unless asked for a plan.

## How it behaves

- Establishes what the architecture must optimise for before evaluating it:
  scale, team, latency, compliance, cost.
- Builds a model from evidence: modules, dependency direction, request and data
  flows, consistency model, trust boundaries, deployment units, observability.
- Reviews six dimensions: responsibility and boundaries, coupling and cohesion,
  data and consistency, reliability and operations, security and privacy,
  evolution and cost.
- Refuses to recommend patterns (microservices, event-driven, CQRS, caches)
  merely because they are common — only when they solve an identified
  constraint.

## Output

Architecture summary → what works → material findings ordered by impact →
recommended target direction → incremental sequence → risks and unresolved
questions. Findings are labelled **Keep**, **Fix now**, **Improve next**,
**Watch**, or **Avoid**, each with evidence, time horizon, migration path, and
trade-offs.

## Example invocations

```text
Use architecture-review on the authentication and authorisation boundaries.
Use architecture-review to assess whether the billing service should be split.
```

## Related

- `templates/architecture-review/architecture-review-template.md` — report structure
- `adr` — record the decisions that come out of a review
- `checklists/security.md` — depth for the security dimension

## Installation

See the repository root `README.md`.
