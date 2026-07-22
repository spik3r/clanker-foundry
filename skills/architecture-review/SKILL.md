---
name: architecture-review
description: Assess a codebase, subsystem, or proposed design for boundaries, coupling, cohesion, data flow, failure handling, security, operability, and evolutionary fit. Use for architecture reviews, design reviews, refactor planning, service-boundary questions, and maintainability assessments. Prefer evidence and incremental recommendations over idealised rewrites.
---

# Architecture Review

Evaluate whether the design fits its real requirements and can evolve safely.

## Establish context

Identify:

- business or user outcome;
- system scope and expected scale;
- key quality attributes;
- constraints such as team size, deployment model, latency, compliance, and cost;
- whether this is an existing system, proposed design, or migration.

Do not evaluate architecture without understanding what it must optimise for.

## Build an evidence-based model

Inspect and describe:

- modules, layers, services, and ownership boundaries;
- dependency direction;
- primary request and data flows;
- persistence and consistency model;
- integration points and trust boundaries;
- deployment units and runtime dependencies;
- observability and operational controls;
- tests and change boundaries.

Use diagrams only when they clarify the analysis.

## Review dimensions

### Responsibility and boundaries

- Do components have clear reasons to change?
- Are domain rules separated from transport, persistence, and vendor concerns?
- Are boundaries aligned with ownership and deployment needs?
- Are abstractions real or ceremonial?

### Coupling and cohesion

- Are dependencies explicit and directed?
- Do changes spread across unrelated areas?
- Are shared modules becoming hidden integration points?
- Is duplication cheaper than a premature shared abstraction?

### Data and consistency

- Who owns each source of truth?
- Are transaction boundaries clear?
- How are concurrency, retries, idempotency, and partial failure handled?
- Are migrations and compatibility planned?

### Reliability and operations

- What happens when dependencies fail?
- Are timeouts, retries, backoff, circuit breaking, and recovery proportionate?
- Can operators detect, diagnose, and reverse failures?
- Is deployment or rollback coupled across components?

### Security and privacy

- Where are trust boundaries?
- How are authentication, authorisation, secrets, sensitive data, and audit events handled?
- Does the design minimise exposure and privilege?

### Evolution and cost

- Does the design support likely changes?
- What complexity does it impose now?
- Is scale evidence-based?
- Could a simpler design meet the same needs?

## Classify findings

For each finding, state:

- evidence;
- impact;
- time horizon: current, near-term, or speculative;
- recommended change;
- migration path;
- trade-offs.

Label recommendations as:

- **Keep**: sound and proportionate.
- **Fix now**: active correctness, security, reliability, or delivery problem.
- **Improve next**: likely near-term friction with a practical incremental fix.
- **Watch**: plausible concern without enough evidence to act.
- **Avoid**: proposed complexity with weak justification.

## Output

1. **Architecture summary**
2. **What works**
3. **Material findings**, ordered by impact
4. **Recommended target direction**
5. **Incremental sequence**
6. **Risks and unresolved questions**

Do not recommend microservices, event-driven systems, CQRS, repositories, caching, or other patterns merely because they are common. Recommend patterns only when they solve identified constraints.
