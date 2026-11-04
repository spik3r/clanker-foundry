---
id: CF-revision-model-review-v3
status: review
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
`modeling/lifecycle/internal/operatorcli/revision_versioned.go`, the single
unknown-version literal in `revision_versioned_test.go`, and new
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
The existing route test uses v3 as its unknown-version example; advance only that
literal to v99. Its v1/v2 assertions and frozen fixtures remain unchanged.

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

## Local implementation and validation

Independent assertions were frozen in `2044b8d` before production edits: eighteen
vectors cover ten accepted inputs and eight malformed/refused inputs. The native
check-only controls passed before implementation. The frozen corpus SHA-256 is
`e0c892b4c65a295eba8ffb68257399317f2352a770b7b531fdb9391446df9098`;
the contract harness SHA-256 is
`8d5e996ad81f6f427f7f4259d2c395fc10796fced758ae896cda37f1b08b19b5`.
Both files remain unchanged. Expected model edits and direct-link causes were
stated before native observations, not copied from a new v3 report.

V3 now preserves every v2 observation and retains the native model-diff bytes and
hash. State, requirement and guard edits stay visible while their stale suites
remain `not-checked`; model spacing and surrounding whitespace retain their
separate source/canonical identities. The version adapter adds only explicit v3
selection. The old unknown-version test marker changed from v3 to v99; its v1/v2
behavior assertions and all old frozen files remain unchanged.

Go 1.27.3 offline full race tests passed: 75 tested packages and two no-test
command builds. Full vet, model validation, unchanged generated-artifact check,
gofmt and repository validation passed. Focused tests also passed all frozen
cases and actual compiled-command complete output, closed stdout exit 3 and
legacy-v1 refusal. A separate exact-source reviewer found no blocking issue and
independently passed focused race tests, vet, formatting and whitespace checks.
No implementation source changed after that review or during aggregate checks.

Installer regression reached its GNU Stow prerequisite and stopped because Stow
is unavailable; it did not pass. Optional skills-ref was unavailable and skipped.
No tool installation, additional platform testing, protected visual review,
remote publication, formal PR approval, merge or deployment was performed.
Committed Project Memory validation follows the completion commit. Publication,
exact-remote-head checks and any integration decision remain with the coordinator.
