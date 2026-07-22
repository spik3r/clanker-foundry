# performance-investigation

Diagnoses latency, throughput, memory, CPU, I/O, database, network, build, or
frontend performance problems through measurement. Baseline first, isolate the
bottleneck, change one factor, measure again.

## When to use

- a system is slow, resource-heavy, timing out, or regressing;
- an optimisation is proposed and needs evidence before and after.

Do not use for general code cleanup with a performance excuse — use
`safe-refactor` when behaviour stays the same and no measurement is involved.

## How it behaves

- Rejects vague goals ("make it faster") until a measurable target exists:
  metric, current value, target, percentile, environment.
- Builds a reproducible baseline: exact command, versions, dataset, warm-up,
  repeated runs, variance.
- Locates the bottleneck with stack-appropriate tools — traces, profiles, query
  plans, contention analysis — and accounts for coordinated omission, noisy
  neighbours, debug builds, and warm caches.
- Tests the cheapest high-information hypothesis first; changes one factor at a
  time.
- Treats caching, batching, concurrency, and friends as costs to analyse, not
  free wins.
- Reverts changes that show no meaningful gain.

## Output

**Problem and target** → **Baseline** → **Evidence and bottleneck** →
**Experiment** → **Results** (before/after, variance, trade-offs) →
**Trade-offs and next action**. Never claims an optimisation worked from code
inspection alone.

## Example invocations

```text
Use performance-investigation to diagnose the slow portfolio endpoint.
Use performance-investigation: p99 latency doubled since the last release. Find out why.
```

## Related

- `workflows/bug-investigation.md` — the same discipline applied to correctness
- `engineer` in repair mode — for applying a proven fix

## Installation

See the repository root `README.md`.
