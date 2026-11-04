---
id: CF-revision-model-review-v3
status: in-progress
owner: Foundry revision model review contributor
repositories: [spik3r/clanker-foundry]
---

# Retain full model changes beside revision evidence

## Implementation claim

Base: main at `ade8fc8e23755638e162aa0dd897edf313e22d74`. Branch:
`codex/revision-model-review-v3`. The coordinator requested continued planned
model-driven pipeline implementation and approved this bounded local scope.
Current main, live open PRs and branch/task ownership were inspected. The v2
revision task is done and reserves no further work. The parallel change-command
adoption lane owns dispatch, native execution adapters and the main entrypoint;
this task does not edit those files. Local claims are not shared reservations.
Recheck current ownership before publication. Publication and integration remain
with the coordinator.

Write only new `modeling/lifecycle/revisioncheck/v3.go`, `v3_test.go`,
`v3_contract_test.go`, `testdata/v3-observations.json`;
`modeling/lifecycle/revisioncheck/README.md`;
`modeling/lifecycle/internal/operatorcli/revision_versioned.go` and new
`revision_v3_test.go`; `docs/model-driven-current-workflow.md`; this task and its
new immutable Project Memory source/events. Preserve v1/v2 implementation and
fixtures, native model/scenario/trace packages, standalone modelrevisioncheck,
other command routes/help, generated artifacts, visual assets, source policies,
runtime, CI, permissions and other tasks. No access to held visual-review assets.

## Explicit v3 contract

`CheckV3` accepts the existing before/after snapshot shape under the explicit
`lifecycle-revision-check-request/v3` version. It preserves all five native
reports and their hashes from v2, then adds `modelDiffReportJSON` and
`modelDiffReportSHA256` with the exact native model-diff encoded bytes and hash.
The output version is `lifecycle-revision-check-report/v3`. All other report
fields retain v2 meaning. No overall readiness, acceptance or approval decision
is added. Execution remains `not-run` and review remains `not-established`.

After existing native checks accept the original inputs, invoke `impact.Execute`
with the exact original before/after proposal model texts and an empty bindings
array. Do not compact the model tokens, infer scenario assertions, synthesize
expected behavior or repair stale evidence. Native model-diff source hashes refer
to embedded JSON values; original full-text hashes remain in proposal reports.
Canonical-equal models may have different source bytes. Model differences remain
observable even when scenario comparison is not-checked for stale suite binding.

A native model-diff invalid input, output limit or unexpected boundary failure
rejects the whole v3 composition, with no partial report, mapped to the existing
invalid/limit/boundary categories. Preserve the wrapper's existing 32 MiB request
and 64 MiB output bounds and all narrower native bounds. `foundry revisions`
selects v3 only by explicit formatVersion in its existing version adapter.
Existing v1/v2 routes and the strict-v1 standalone command stay unchanged.

## Acceptance and handoff

Commit this reviewed public-safe claim, register its pinned source and record a
genuine start before tests or implementation edits. A separate contributor freezes
hand-stated model-change assertions and native byte-parity checks before production
code. Cover unchanged and formatting-only models, edited state/requirement/guard
content and both graph links, stale manifests, stale suites, input refusal, exact
native byte/hash retention and all prior report preservation. Check strict wrapper
and version refusal, limits, deterministic concurrent calls and no input mutation.
Exercise the actual foundry version route and output failure without any scenario
execution, native change executor, network call or platform-expansion work.

Run the existing trusted Go 1.27.3 toolchain offline: focused and full race tests,
vet, model validation/drift, formatting, repository validator, installer regressions,
whitespace and committed Project Memory checks. Report actual skipped or blocked
checks. Obtain independent exact-source review. This report composition neither
executes a candidate nor authenticates requirements, review or evidence provenance.
