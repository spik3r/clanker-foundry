---
name: performance-investigation
description: Diagnose latency, throughput, memory, CPU, I/O, database, network, build, or frontend performance problems through measurement. Use when a system is slow, resource-heavy, timing out, regressing, or when optimisation is proposed. Require a baseline, isolate the bottleneck, estimate impact, change one factor, and measure again.
---

# Performance Investigation

Measure first. Optimise the limiting resource, not the most visible code.

## Define the problem

Establish:

- the affected user or workload;
- the metric: latency, throughput, CPU, memory, allocations, I/O, cost, bundle size, render time, build time, or another measure;
- target and current value;
- percentile and time window where relevant;
- environment and dataset;
- whether the issue is a regression or an unmet target.

Reject vague goals such as "make it faster" until a measurable outcome is chosen.

## Build a reproducible baseline

Record:

- exact command, request, scenario, or benchmark;
- versions, configuration, hardware, and environment;
- warm-up and cache state;
- dataset size and concurrency;
- repeated measurements and variance.

Prefer production traces or representative workloads over synthetic microbenchmarks. Use microbenchmarks only for isolated hypotheses.

## Locate the bottleneck

Gather evidence with the tools appropriate to the stack:

- tracing and request breakdowns;
- CPU and allocation profiles;
- database query plans and query counts;
- network timing and payload sizes;
- lock, queue, and thread contention;
- browser performance and render profiles;
- build timing and dependency graphs;
- cache hit rates and eviction behaviour.

Account for coordinated omission, noisy neighbours, debug builds, warm caches, and non-representative local data.

## Form hypotheses

For each likely cause, state:

- evidence;
- predicted effect;
- test that would confirm or reject it;
- expected gain and possible trade-offs.

Test the cheapest high-information hypothesis first.

## Change one factor

Apply the smallest change that tests the hypothesis. Avoid combining unrelated optimisations. Preserve correctness and observability.

Common techniques such as caching, batching, concurrency, denormalisation, indexing, memoisation, and pooling introduce costs. Analyse invalidation, staleness, memory, contention, consistency, and failure behaviour before adopting them.

## Measure again

Run the same workload under comparable conditions. Report:

- before and after;
- absolute and percentage change;
- variance or confidence;
- resource trade-offs;
- whether the target was met;
- remaining bottleneck.

Revert changes that do not produce a meaningful gain unless they have another proven benefit.

## Output

**Problem and target**  
**Baseline**  
**Evidence and bottleneck**  
**Experiment**  
**Results**  
**Trade-offs and next action**

Never claim an optimisation worked from code inspection alone.
