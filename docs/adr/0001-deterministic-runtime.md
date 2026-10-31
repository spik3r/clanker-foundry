# Use an opt-in single-host orchestration pilot

Status: accepted for local implementation, 2026-10-30.

## Context

The owner authorized orchestrated implementation of the
[architecture plan](../deterministic-orchestration-plan.md). This pack already
supports manual task workflows. Automatic assignment needs a shared transactional
authority; prompt changes cannot serialize claims or reject stale results.

## Decision

Keep an optional private Node package under `runtime/`. Use Node 24 or later,
built-in SQLite on local durable storage, one versioned JSON Schema, strict Ajv
validation and a pure transition evaluator. Pin dependencies with a lockfile.
The existing skills remain usable without installing this package.

The pilot operates only on explicitly admitted tasks in its local namespace.
It does not import live ownership, publish task changes or migrate another tracker.
All callers are trusted local processes; owner, worker and validator identities
are supplied through the trusted API boundary. No network authentication service
is included. A person with database access is inside that trust boundary.

Support fake workers for recovery tests and a bounded read-only worker adapter.
No adapter receives repository write, publication, merge or deployment authority.
Unknown dispatch outcomes require an owner decision rather than automatic relaunch.
Read-only provider integration is supervised; inherited external tools must not
bypass the capability envelope. Hosts unable to enforce it are unsupported.

Retry counts, rework budgets and lease/delay durations are required finite task
policy fields. There are no hidden infinite defaults. Only registered checks and
explicit subject-bound approval can satisfy configured acceptance gates.

Preserve Project Memory's current committed Markdown source contract. Manual
source registration, genuine starts and fresh completion remain required for
repository implementation; runtime records do not replace that evidence.

## Consequences

This gives a testable local transaction and recovery boundary without installing
a full BPMN engine. It adds a runtime dependency and an operator-owned database.
The pilot has no claim to BPMN/fUML conformance or multi-host correctness.
Unattended writable workers, external tracker cutover and distributed operation
need separate authority/isolation decisions before enablement.

The implementation and its acceptance results live in the bounded
[contracts](../../tasks/CF-orchestration-contracts-implementation.md),
[core](../../tasks/CF-orchestration-core-implementation.md),
[store](../../tasks/CF-orchestration-store-implementation.md) and
[pilot](../../tasks/CF-orchestration-pilot-implementation.md) tasks.
