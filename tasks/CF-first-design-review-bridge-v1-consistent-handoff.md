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

## Observed local result

Added a check-only composition and stdin/stdout CLI using only the two existing
in-memory check APIs. It retains the complete typed intake/model report, including
its exact nested native intake JSON and hash, and the exact encoded proposal
check report and hash. It compares the decoded source texts separately; it does
not claim that the outer typed intake report is byte-identical to the separate
intake CLI's encoding. The fixed no-authority envelope has no aggregate approval
or readiness flag. No producer, existing fixture, runtime or workflow was edited.

A separate assistant contributor froze 33 literal observation cases at commit
`20dfd89` before implementation, deriving them from public contracts and existing
invented inputs without running the new bridge or a candidate reducer. Another
contributor wrote the external oracle test. All cases passed, including stale
native declarations despite matching cross-request bytes, missing versus empty
observations, retained `Q_use`, and strict malformed inputs. The test separately
compares the complete native typed intake result and exact encoded proposal
receipt, hashes and input preservation. Contributor separation is not
independent-person review or human approval.

Two compiling mutants in a disposable copy were rejected by the frozen oracle:
forcing mismatched bytes to match failed at the model-whitespace witness, and
omitting the native intake JSON failed at the open-question witness. The
unchanged source and frozen oracle hashes were checked. The literal README
command ran from the module directory with a cold external cache on the invented
first-design requests: model and suite matched, `Q_use` remained visible, and
proposal preflight was checked/not-run with no candidate start. An inside-checkout
output parent was refused before creating work. A separate source reviewer found
no material issue, passed focused Go race tests, and rechecked the final README.

Go 1.27.3 offline serial `go test -race -count=1 -p 1 ./...` passed 65 tested
packages plus one without tests. Whole-module `go vet`, model validation,
generated-artifact drift and trace-view checks passed. The successful full run
used an isolated Git-initialized workspace temp root and workspace cache; it did
not modify tests or disable VCS checks. An initial aggregate failed the existing
clean-regeneration test because a sandbox-projected empty `.git` above its temp
copy caused Go VCS status to fail. A second aggregate under a low-space `/tmp`
root passed that test but failed an existing qualification subprocess with exit 1
and no exposed child diagnostic. Neither failed aggregate is counted as a pass.
Both affected unchanged packages passed focused reruns in the workspace temp root
before the fresh full pass. Repository validation, Markdown links and `gofmt`
passed; optional `skills-ref` was unavailable. No formal approval, real-use
adoption, visual review, deployment or worker activation occurred.
