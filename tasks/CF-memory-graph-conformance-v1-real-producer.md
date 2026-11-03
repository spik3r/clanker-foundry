---
id: CF-memory-graph-conformance-v1
status: in-progress
owner: Astra Project Memory conformance contributor
repositories: [spik3r/clanker-foundry]
---

# Check real Project Memory graph behavior from a declared contract

## Implementation claim

Add a bounded, read-only conformance layer over the existing Project Memory v2
`buildGraph` and `completionHeads` functions. They remain the only implementation
under test. This is model-driven test and coverage preparation for a real tool,
not another task-state evaluator, a replacement writer or production orchestration.

The coordinating owner requested substantial disjoint work and accepted this
scope. Main was freshly verified at
`9a05ecdcf263d3ca1b09f78e3fe29bc3ae753be8`. Current main task records and local v2
frontiers show the original `CF-001` producer run complete and no active overlap.
The two open Foundry PRs, #20 and #51, own different paths. This check is evidence
of the inspected state, not a distributed lock; recheck before publication.

Exact maintained scope:
- `modeling/project-memory-conformance/README.md`
- `modeling/project-memory-conformance/check.mjs`
- `modeling/project-memory-conformance/coverage.json`
- `modeling/project-memory-conformance/testdata/cases.json`
- `modeling/project-memory-conformance/testdata/README.md`
- `tests/project-memory-conformance.test.mjs`
- `.github/workflows/project-memory-conformance.yml`
- this task record and its new immutable Project Memory source/events.

Do not edit the authoritative producer, its existing tests, `FORMAT.md`, the
generic Go lifecycle model/compiler, README reorganization, #20 artifacts,
modelwalk fixtures/cache, #51 trace/runtime paths, historical metadata, packages
or current workflows. The new workflow has read-only repository access, no saved
checkout credentials, no secrets, no deployment and no manual operational action.
No Heisentick task, protected source, prior denial or production gate is changed.

## Contract and independent evidence

Represent the covered rules and concrete graph cases as fixed declarative data.
A separate contributor states literal accept/refuse outcomes, accepted frontiers
and fresh-completion selections from the existing written v2 contract, before
running the producer or new harness. Freeze those original bytes before replay.
The contributor may inspect existing fixtures for data shape, but must not derive
expectations from observed behavior. Record assistant separation honestly; it is
not an independent human approval or an independent semantic proof.

The harness validates its bounded fixture shape, calls the existing producer on
invented in-memory sources/events, and compares observations with frozen literals.
It also replays bounded event-order permutations, checks input preservation and
derives a deterministic rule-to-case coverage inventory. It must not implement
the lifecycle's decisions itself, repair expectations, write real task events,
invoke Git or child processes, or fetch data. Reports distinguish graph refusal,
completion-selection refusal, mismatch, missing coverage and unrun cases.

Cover starts, checkpoints, terminal events, restart, run identity, scope,
parent integrity, competing heads, antichain reconciliation and fresh completion
selection. This custom conformance inventory is not the Go `lifecycle/v1`
semantic profile: that task-state profile does not represent the causal DAG.
No UML, SysML, BPMN, UTP or fUML conformance is claimed. Graph checks do not cover
the Git index, filesystem writer, committed-diff validator or live authority.

## Validation and handoff

Run the frozen corpus against the real producer, focused harness refusals,
repeated/permuted replay and existing Project Memory/workflow regressions.
Demonstrate that disposable producer mutations are caught by behavioral cases,
not merely by byte hashes. Preserve the original expectation bytes and disclose
any contract discrepancy instead of silently changing policy or assertions.
Run repository checks and ordinary full committed-diff Project Memory validation;
obtain separate exact-source review, verify a clean restore and report before
publication. Ordinary draft publication is authorized; no merge, formal approval,
deployment, new runtime authority or denied-action retry follows from this claim.
