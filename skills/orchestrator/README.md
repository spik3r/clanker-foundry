# Orchestrator

Coordinates a large task across scoped subagents while keeping each worker's context
small. It connects the explorer-researcher, architect-planner, builder-developer, and
reviewer-verifier roles into a deliberate pipeline.

## Use it for

- work with independent slices that can run in parallel;
- research followed by planning, implementation, and review;
- broad changes where a fresh reviewer reduces confirmation bias;
- tasks that need durable handoffs across stages or sessions.

## Do not use it for

- a single small change;
- tightly coupled work where every step needs the same full context;
- tasks where coordination would cost more than doing the work directly.

## Behaviour

The lead session decomposes and schedules work, collects short artifacts, and resolves
dependencies. Workers get isolated briefs. A reviewer with fresh context checks the
result against the approved plan. Blocking findings return to a builder for a capped
build-review loop.

The bundled templates provide shapes for the plan, worker handoff, findings, and review.
The skill remains usable without any other skill being installed.
