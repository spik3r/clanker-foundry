---
id: CF-revision-scenario-review-v2
status: in-progress
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
