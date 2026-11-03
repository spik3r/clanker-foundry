---
id: CF-intake-impact-v1
status: in-progress
owner: Byte intake-change integration contributor
repositories: [spik3r/clanker-foundry]
---

# Review changes to original intent and artifact assertions

## Implementation claim

Add a bounded offline comparison of two original intake requests on branch
`codex/intake-change-review`, starting from main
`9d06274f48f51a6c6b05d17980fcb095ea00960d`. Main, open PRs and held claims were
inspected; the coordinator confirmed no known overlapping intake-change lane.
Existing modeldiff compares accepted model meaning; traceimpact compares model,
requirements and catalog declarations. Neither compares the upstream intent,
decisions, unanswered questions, artifact assertions or asserted dependencies.
This tool adds that missing review surface, not another model or trace comparator.

New scope only: `modeling/lifecycle/intakeimpact/compare.go`, `README.md`,
`compare_external_test.go`, `integration_test.go`, `testdata/before.json`,
`testdata/after.json`, `testdata/expected.json`; and
`modeling/lifecycle/cmd/modelintakediff/main.go`, `main_test.go`; this task and
its new immutable Project Memory records. Keep all existing intake, trace,
acceptance, compiler, model, generated, workflow and runtime files unchanged.
Do not touch the held intakecheck/reviewpack/export stack, PR20 process view,
the pipeline guide or the separate regeneration-verification lane.

## Contract

`Compare(before, after []byte) (Report,error)` calls existing `intake.Execute`
on each original request. Preserve its exact report bytes, including final LF,
in each receipt's `reportJSON`, with `reportSha256` and native `requestSha256`.
No prebuilt report input is accepted. Intake remains the validation authority.
Invalid or oversized requests refuse with no partial report. Artifacts remain
opaque: no model parser, schema validator, suite execution, inferred case links,
expectation refresh, automatic rebinding, packaging or filesystem/network access.

Report version `lifecycle-intake-impact/v1`, mode `example-only`, authority `none`.
Each receipt also has `contentSha256`: hash a fixed-order JSON object with
`formatVersion: lifecycle-intake-content/v1`, then the native report's `intent`,
`decisions`, `questions`, `bindings`, `links`, HTML escaping disabled and no LF.
Native normalized list order is retained. This identifies the compared content,
not artifact semantics: raw artifact whitespace remains observable through hashes.
`requestSourceChanged` compares exact request hashes; `contentChanged` compares
content identities. Request formatting and decision/question reordering alone
change only request identity, not content. No approval/ready/qualified flag.

`changes` is sorted by kind then ID. Each item has `kind`, `id`, `change`
(`added`, `removed`, `modified`) and sorted `fields`. Intent uses kind/id `intent`
and field `text`. Decisions use `text`, `rationale`; questions use `text`.
Added/removed decisions or questions have empty fields. Moving an ID between
those lists is a removal plus addition; removal does not prove resolution.
Artifact roles always exist; compare nullable `declaredSha256`, `declaredAuthor`,
`observedSha256` and `status`. Dependency IDs are `model->requirements`,
`scenarioSuite->requirements`, `scenarioSuite->model`; compare nullable
`assertedSha256` and `status`. Artifact/dependency changes are `modified`.
No changed values need be repeated outside the exact native receipts.

`beforeGaps` and `afterGaps` retain every unresolved question (`kind:question`,
status `unresolved`), every artifact binding not `matches-declaration` and every
dependency not `matches-declaration`, sorted by kind then ID. An unchanged gap
never disappears just because the change list is empty. Matching assertions do
not prove authorship, independence, artifact validity, test execution or review.
The fixed statuses `artifactValidation:not-run`, `executionStatus:not-run`, and
`reviewStatus:not-established` make this boundary explicit.

CLI input is one strict JSON object with `formatVersion:
lifecycle-intake-impact-request/v1`, `before` and `after`, the latter two literal
UTF-8 original intake request strings. Required exact keys; no arguments or paths.
Outer input capped at 64 MiB, each intake at its existing 4 MiB bound, buffered
report at 16 MiB. CLI requires Go 1.27.3. Exit 0 means report emitted including
changes and gaps; 2 is invalid input/read/toolchain refusal without output; 3 is
failed/short output, possibly partial. Suppress SIGPIPE termination to report 3.
No clock, random IDs, mutable global data or actual authorization is introduced.

## Verification

A separate assistant freezes invented before/after requests and literal expected
changes/gaps before implementation. This is separation between assistants, not
independent-person approval. Test strict transport, no mutation, stable ordering,
exact native receipts, formatting versus content, addition/removal, null/empty,
unchanged gaps, invalid inputs, bounds and failed writes. Integration tests feed
unchanged observed artifacts to merged traceimpact and acceptanceauthor, showing
raw-byte changes versus canonical meaning and stale acceptance without rebinding.
These integration checks do not create new expected behavior from runner output.
Run pinned offline full Go race tests, vet, modelc/modeltrace drift, repository
validation and full committed Project Memory validation. Obtain independent
source review; prepare one draft PR if permitted. No merge, deploy, activation,
credential changes or reroute around a denied operation.

## Implementation evidence

The separate-assistant oracle was frozen in `00a5890` before implementation;
its SHA-256 remains
`08434f2445f8e79d5ba23f5490d4a77b8dc91adbef69fd975063b831a274e244`.
Implementation commit `add7ecb6af75bef2fa659c147fff7409227d761a` passed separate
source review and 633 independently constructed binary probes, including exact
and plus-one request/observation limits, Unicode/control characters, native
receipt and content hashes, no-op gaps, invalid transport and failed output.
All six tested mutants compiled and were killed by the unchanged frozen tests:
dropped unchanged gaps, conflated source/content identities, removed questions
called resolved, omitted author changes, stripped receipt LF and omitted observed
hash changes. This is separate assistant authorship/review, not person-level
independence or authenticated execution evidence.

Observed implementation checks: pinned offline Go 1.27.3 full race tests (29
packages), vet, modelc validate/check, modeltrace validate/check, repository
validation, 143 Project Memory helper regressions and the README commands passed.
The optional skills-ref tool was unavailable. Existing model trace gaps remain
12 requirements, 9 linked, 3 incomplete and one unlinked catalog case. Five
pipeline tests distinguish source bytes from accepted model/requirements meaning
and stale acceptance bindings without changing any declaration or frozen literal.
Current-main integration and final full-diff lifecycle verification are recorded
in the completion event, not inferred from this implementation-only result.
