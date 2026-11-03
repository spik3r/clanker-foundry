---
id: CF-first-design-review-bridge-v1
status: in-progress
owner: First-design review bridge contributor
repositories: [spik3r/clanker-foundry]
---

# Connect intake and proposal preflight for review

## Implementation claim

The coordinating owner requested a small integration between the existing
first-design intake/model check and proposal check. Fresh main is
`4fe0611d032ea7060e9804b23634d1ec13c481a4`. The existing intake, first-design
handoff and proposal replay tasks retain their owners and paths; this proposal
does not transfer or modify those claims. Targeted task/path and open-PR checks
found no competing claim for the new paths below. This source is a local claim,
not a shared lock. Recheck ownership and main before publication.

Exact scope:
- `modeling/lifecycle/firstdesignreview/README.md`
- `modeling/lifecycle/firstdesignreview/review.go`
- `modeling/lifecycle/firstdesignreview/review_test.go`
- `modeling/lifecycle/firstdesignreview/oracle_test.go`
- `modeling/lifecycle/firstdesignreview/testdata/oracle.json`
- `modeling/lifecycle/firstdesignreview/testdata/README.md`
- `modeling/lifecycle/cmd/modelreviewcheck/main.go`
- `modeling/lifecycle/cmd/modelreviewcheck/main_test.go`
- this task and its new immutable Project Memory source/events only.

Before implementation, a separate contributor independently states and freezes
literal cases and expected observations in the two testdata files, using the
public contracts and existing independently authored invented inputs, without
running the new bridge or deriving expectations from its results. That separation
is between assistant contributors, not an independent human approval. Do not
modify the upstream inputs, producers, their tasks/evidence, existing guide or
skills, compiler, runner, coverage, generated files, workflows or runtime. The
held process-view artifact and modelwalk lane are excluded.

## Contract

Accept one bounded strict JSON wrapper `lifecycle-first-design-review-request/v1`
with exactly `formatVersion`, `intakeRequest` and `proposalRequest`; the latter
are original UTF-8 JSON request texts. Invoke existing in-memory
`intakemodel.Inspect` and `proposalreplay.Check`, not any execute route. Only
after each native request validates, compare the decoded intake-observed model
and suite directly against the proposal request's decoded corresponding text.
For each comparison report `matches`, `mismatch` or `missing`; retain both full
native reports, all original unresolved questions, and their distinct
validation/binding states. Invalid input returns no partial success report. Limit
input and output explicitly, fail closed on malformed/duplicate/unknown keys,
invalid Unicode, limit excess and short output, and keep input bytes untouched.

This is a review composition, not a combined ready/approved/execute decision.
Use `example-only`, `authority=none`, `executionStatus=not-run` and
`reviewStatus=not-established`; do not hide a mismatch or unresolved question,
repair declarations, infer authorship, run generated code, render diagrams,
activate workers, integrate, deploy, access the network or invoke child processes.
An identity match does not prove human approval, correctness or provenance.
This tool runs trusted local code; it is not a sandbox.

## Validation and handoff

Exercise separately frozen positive, stale, missing and cross-paired model/suite
cases; malformed input, size and output errors; complete native-report
preservation; repeatability; source preservation and no candidate invocation.
Use independently reasoned literal outcomes, not generated results as an oracle.
Run pinned-toolchain focused and relevant full-module checks, repository and
committed-diff Project Memory validation, obtain a separate exact-source review,
and report limits. Any public draft PR or merge requires actual user authority;
no formal approval, production use, deployment or worker activation is implied.
