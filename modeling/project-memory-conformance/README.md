# Project Memory graph conformance

This connects a declared rule/case inventory to the real Project Memory v2
producer. It checks `buildGraph` and `completionHeads` from the existing
[state module](../../skills/project-memory/scripts/lib/state.mjs). There is no
second lifecycle evaluator, task writer or runtime authority. It is a bounded
model-driven testing step, not adoption of generated code into Project Memory.

The [frozen corpus](testdata/cases.json) contains 50 concrete invented graphs and
literal expectations under 14 declared rules. A separate assistant contributor
wrote those expectations from the [format contract](../../skills/project-memory/FORMAT.md)
and [workflow](../../skills/project-memory/SKILL.md) before execution. Source
inspection exposed implementation details while constructing valid fixtures:
this is separate pre-execution authorship, not blind or human-independent review.
See [provenance and limits](testdata/README.md).

## Run

From a trusted exclusive checkout with Node 24 or later:

```sh
node --test tests/project-memory-conformance.test.mjs
node modeling/project-memory-conformance/check.mjs check
```

No dependency install, Git call, child process, network access or real task-event
write is performed by the checker. It imports trusted local JavaScript and reads
fixed repository files. This is not a sandbox or an immutable-host guarantee.
Use the ordinary repository instructions for preparing the checkout and for
running other checks.

`check` returns a deterministic JSON report. Exit 0 means all literal graph and
completion comparisons, input-preservation checks and the generated coverage
file match. Exit 1 means a behavioral, input-preservation or coverage mismatch;
read `failedCases`, `coverageMatches` and the individual observations. Exit 2
means fixture/host/usage failure, without a successful report. No exit status
grants ownership, publication, integration, approval or execution authority.

The current corpus produces **50 case results and 598 event-order runs**.
Twenty-five graph cases are expected to refuse. An expected refusal is a passing
comparison, not a successfully accepted graph. An accepted graph can still have
unresolved competing heads: those cases expect completion selection to refuse.
An empty selected-completion list can be correct; do not equate it with coverage.

Each result reports its original-order observation, number of failed orderings
and first mismatch, if any. Expected results never choose which producer call
runs. Every accepted graph is passed to the actual completion selector. The
checker compares exact normalized head IDs and selected completion IDs, not just
aggregate counts. It checks the supplied objects, source map and freshness set
remain unchanged. Inputs are copied separately for each ordering.

## What the inventory models

The declared rules cover initial starts, checkpoints, terminal events and
restart, run identity, scope, parent integrity, competing heads, antichain
reconciliation, timestamp irrelevance and fresh completion selection. The corpus
is data, not executable instructions. No string in a fixture is run as code.

This custom conformance inventory does **not** use the Go `lifecycle/v1` semantic
profile, which describes a different task-state model and cannot represent this
causal DAG. No UML, SysML, BPMN, UTP or fUML conformance is claimed. The existing
Node producer is the only implementation whose behavior is observed here.

[coverage.json](coverage.json) is generated from rule references and case IDs.
It shows declared test coverage and planned ordering counts, never test success
or proof that a requirement is correct. Unreferenced rules, unknown references,
duplicate identities and malformed case wrappers fail fixture validation.
Cases include full source/event objects, not a second program that computes
expected validity or chooses winning heads.

The current graphs have at most five events, so all their event orders are
replayed. For future graphs of six through sixteen events, the checker deliberately
uses only original, reverse and rotated orders. It will label their actual run
count, not claim exhaustive graph exploration. Reordering adds determinism
evidence, not new independent semantic cases.

## Updating the declared coverage

This explicit command writes only the fixed generated coverage file. It does
not replay behavior or alter the frozen corpus:

```sh
node modeling/project-memory-conformance/check.mjs coverage
```

Normal checks never rewrite that file. A stale file makes `check` fail. New
expectations need separately reviewed requirements and authorship before replay;
do not regenerate them from observed results. The test pins the original corpus
SHA-256. Preserve it and document any future independently justified replacement
instead of silently changing an assertion to make a producer pass.

## Bounds and evidence limits

The checked-in fixture file is at most 8 MiB of UTF-8 JSON, with a bounded
validated wrapper: at most 128 rules/cases, 8 source objects and 16 events per
case, depth 32 and 250,000 traversed values. Native JSON decoding is used for
this trusted repository fixture, not advertised as a hostile-input or
duplicate-key-rejecting protocol. Producer event validation remains separate;
intentionally invalid event fields reach the producer unchanged.

The producer uses ordinary `Error` for refusal, without typed refusal codes.
Those messages are retained as observations but not matched as contract text.
Non-`Error` throws and subclasses such as `TypeError` are reported as unexpected
errors and cannot match an expected refusal. A plain thrown `Error` still does
not prove which internal rule caused it. Some deliberately invalid corpus cases
violate more than one condition; the corpus notes identify this limitation.

The report hashes current corpus, source and written-contract bytes for review.
Source hashes are observations, not a separate pass/fail criterion, provenance,
loaded-code authentication or race-free invocation proof. This lets behavioral
mutations fail on actual mismatches instead of merely failing a hash pin.

This suite does not exercise Git fingerprint matching, the committed-diff
validator, filesystem writes, registration authenticity, operational scheduling
or all possible graphs. Existing [Project Memory tests](../../tests/project-memory.test.mjs)
and [workflow tests](../../tests/project-memory-workflow.test.mjs) remain required.
The separate read-only CI workflow runs this corpus but does not install a
required-status rule or authorize merging. No existing producer policy changes.
