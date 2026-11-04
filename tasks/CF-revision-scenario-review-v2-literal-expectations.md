---
id: CF-revision-scenario-review-v2
status: review
owner: Foundry revision scenario review contributor
repositories: [spik3r/clanker-foundry]
---

# Show literal scenario changes beside revision review

## Implementation claim

Owner: Foundry revision scenario review contributor; the coordinator alone merges.
Base: main at 2844622c8e0eac5d00d6f715e6cee21b1f65595c. Branch:
codex/revision-scenario-review-v2. The owner requested the next bounded batch and
the coordinator approved this fifth-report scope. Live open-PR and task/branch
checks found no competing revision-v2 claim. Local snapshots do not reserve work.

Write only new modeling/lifecycle/revisioncheck/v2.go, v2_test.go,
v2_contract_test.go, testdata/v2-observations.json; revisioncheck/README.md;
new internal/operatorcli/revision_versioned.go and revision_versioned_test.go;
internal/operatorcli/dispatch.go; docs/model-driven-current-workflow.md; this task
and its new immutable Project Memory source/events. Preserve existing
revisioncheck/check.go, all v1 source/fixture bytes, operatorcli/native.go,
RunRevision, standalone modelrevisioncheck, help, native scenario/model logic,
source policies, runtime, CI, permissions and other tasks.

## Explicit v2 contract

CheckV2 accepts the same before/after snapshot shape under
lifecycle-revision-check-request/v2. It preserves all four original native report
texts and SHA-256 observations, then adds a native scenariodiff comparison of the
actual proposal model/suite pairs. Stale manifest declarations alone do not block
that comparison. Never repair suite bindings or expected values.

The output lifecycle-revision-check-report/v2 retains the existing fields and
adds scenarioDiff: status checked or not-checked, nullable reason, reportSHA256
and reportJSON. Checked retains exact native encoded report bytes and hash,
with null reason. Native invalid-input or limit refusal retains the four reports
and uses not-checked with native-input-refused or native-limit-refused and null
report/hash. Do not guess whether a native refusal means stale or malformed.
Unexpected native failures reject the composition as a boundary failure.
Existing wrapper/native per-input and combined output bounds still apply.

Use original model/suite JSON texts in the fixed native comparison wrapper only
after existing native first-design acceptance has checked full strict JSON
framing. Preserve whitespace inside the embedded tokens. Native source hashes
exclude whitespace outside a JSON token; original full suite hashes remain in
preflight. The scenario report identifies changed fields, identities and order,
not full literal payloads. Reviewers retain the original suites. Complete model
edit review remains the separate modeldiff tool; no sixth report is added.

The foundry revisions route selects v1/v2 only by explicit formatVersion through
a new adapter. Existing v1 output/diagnostics and the legacy strict-v1 CLI remain
unchanged. No new flags, callback, readiness/approval decision, scenario execution,
worker, filesystem operation or network call enters the check composition.

## Acceptance and handoff

Publish this claim, register the reviewed source and record a genuine start before
new tests or production edits. A separate contributor freezes literal-change,
order, formatting, stale-manifest, stale-suite and malformed controls before
implementation. Verify full native byte/hash parity, unchanged v1 source and
outputs, refusal classification, strict input, bounds and no execution. Use a
private classifier test for native limit refusal rather than claiming an
unreachable oversized input passed v1 validation.

Run Go 1.27.3 full race tests, vet, modelc drift, focused native/operator checks,
pack, installer, gofmt, whitespace and committed Project Memory validation.
Review the exact source independently and hand the PR to the coordinator for
hosted checks and integration. Record actual limitations; comparison is not
correctness, authenticated review, execution evidence or authorization.

## Review handoff

The explicit v2 composition retains all four v1 native report texts/hashes and
adds the original scenario comparison report or an explicit native refusal.
The foundry-only adapter selects the declared request version. Existing Check,
check.go, RunRevision/native.go, standalone CLI, help and old expected data remain
byte-identical. No model/suite binding or expected literal is refreshed.

A separate contributor authored twelve static vectors and froze them in cb87aeb
before production. The harness compiled and passed native fixture controls using
existing APIs before CheckV2 existed; its future assertions take a callback, which
the new builder test supplies only after the freeze. The data SHA-256 remains
274b0329ecbd039905d4042d38b5ce0e245e8cf8eed8798a4ace870026703387.
This is separate-assistant authorship, not human independence or executed-test
provenance inferred from a hash.

Go 1.27.3 offline full race passed on exact source 1cd8f7f: 72 tested packages
passed and two command packages compiled without local test files. Focused
revision/operator/legacy CLI tests passed, including original v1 source pins and
native output parity, concurrent read-only calls, version/strict-input controls,
actual compiled v2 output/closed stdout and legacy v2 refusal. Native limit and
unexpected-error classification were tested privately rather than claiming a
corpus input reached an unavailable native-limit path.

Vet, modelc drift, gofmt, pack validation, installer regressions and whitespace
checks passed. Optional skills-ref is absent. An actual foundry revisions check
of the frozen expected-task change reported modelChanged false, suiteChanged
true, expected.task as the changed field and execution not-run. It performed no
scenario replay, candidate build or visual review.

The coordinator and separate reviewer approved exact source 1cd8f7f. Both also
passed the focused three-package race checks independently. All frozen data and
source pins remain unchanged. Comparison flags are declarations for review,
not approval, correctness or execution evidence.

Next: genuine immutable completion and committed evidence validation, then the
coordinator checks hosted CI and integrates PR65.
