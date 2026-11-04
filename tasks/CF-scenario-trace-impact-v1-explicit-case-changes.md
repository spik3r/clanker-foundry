---
id: CF-scenario-trace-impact-v1
status: in-progress
owner: Scenario trace impact contributor
repositories: [spik3r/clanker-foundry]
---

# Review changed scenario links beside requirement declarations

## Implementation claim

The coordinator assigned this independent Go slice after checking it against the
separate builder input-manifest preflight. Base main is
`01b18785e738d9cc656ea1fb622f8c9aad1b220a`, after PR 79 merged. Local branch:
`codex/scenario-trace-impact-v1`. Public inspection found PR 80, a separate custom
assertion-inventory claim, and no competing scenario-trace-impact implementation
or claim. Related public trace branches were merged. Unpublished worker work is
not presumed idle; this claim grants no shared lock or publication authority.

Write only:

- modeling/lifecycle/scenariotraceimpact/compare.go
- modeling/lifecycle/scenariotraceimpact/project.go
- modeling/lifecycle/scenariotraceimpact/types.go
- modeling/lifecycle/scenariotraceimpact/compare_test.go
- modeling/lifecycle/scenariotraceimpact/contract_test.go
- modeling/lifecycle/scenariotraceimpact/README.md
- modeling/lifecycle/cmd/modelscenariotraceimpact/main.go
- modeling/lifecycle/cmd/modelscenariotraceimpact/main_test.go
- this task and its new immutable Project Memory source/event files.

Existing native packages, fixtures, operator dispatch, builder manifest work,
pilot evidence, runtime, workflows, policies, CI and protected artifacts remain
outside the write scope. Do not publish, merge remote PRs, install tools or expand
platform testing. Keep the result local for independent review. The coordinator
subsequently authorized one normal local merge of current main for integration
qualification, preserving all reviewed commits and avoiding repeated base updates.
That integration imports existing main history without authoring changes to its
unrelated files.

## Contribution to the workflow

Current trace impact compares model/requirement/catalog declarations; scenario
diff compares suite literals; scenario trace binds explicit requirement IDs to
actual case IDs. This slice joins those established observations across revisions.
A reviewer can see removed or added requirement-to-case links and requirements
linked to changed case literals, alongside native declaration changes and gaps.
It does not infer behavioral impact, requirement satisfaction, execution,
independence, review approval or readiness.

## Frozen prospective contract

Public `Check(raw []byte) ([]byte, error)` accepts strict JSON with exactly
`formatVersion: lifecycle-scenario-trace-impact-request/v1`, `before`, and `after`.
The latter two are original complete scenario-trace request texts as strings.
All fields are required and nonnull. Reject unknown/duplicate/case-variant keys,
malformed Unicode, trailing input and unsupported versions. Wrapper limit: 64 MiB;
each original snapshot retains the native 32 MiB request limit. There is no file,
network, subprocess, callback, reducer or provider capability in the package.

First call native `scenariotrace.Check` on each original request. Native stale
bindings and malformed supplied observations refuse the comparison; never repair
or replace sources. After successful validation, extract original embedded texts
and compose native `scenariodiff.Execute` and `traceimpact.Execute` inputs. Preserve
model and suite JSON value tokens when composing the scenario comparison: remove
only surrounding JSON whitespace, preserving all bytes inside each value token.
Do not marshal RawMessage values through a compactor. Native scenario-diff source
hashes identify those tokens; scenario-trace source hashes identify the complete
decoded source texts, including surrounding whitespace. Preserve decoded source
text bytes when composing the native trace-impact request. Check native
model/suite/requirement/catalog identities across the composed reports.

Retain all four exact native output texts and SHA-256 values: before trace, after
trace, scenario diff and declared trace impact. Each retained record has exactly
`requestSHA256`, `reportJSON`, `reportSHA256`. Native limits remain in force.
Before/after request hashes identify the original scenario-trace request texts;
the comparison request hashes identify the exact constructed native requests.
Cross-report identities agree within each side; before and after may differ.
Each trace output is additionally capped at 8 MiB and total encoded output at
128 MiB. Output is buffered, HTML escaping disabled and terminated with one LF.
Preflight both constructed native comparison wrappers against their native input
byte caps and return `ErrLimit` on excess. Other native admission restrictions,
including trace-impact's 512 catalog cases, remain input refusals. Two individually
valid trace snapshots therefore need not be admissible to every comparison.
Identical input produces identical output. Inputs are not mutated; concurrent
read-only calls and independent returned buffers are supported.

The report has exactly these top-level fields:

- `formatVersion: lifecycle-scenario-trace-impact-report/v1`
- `mode: example-only`, `authority: none`
- `executionStatus: not-established`, `reviewStatus: not-established`
- `requestSHA256`: exact outer request bytes
- `before`, `after`, `scenarioDiff`, `traceImpact`: retained native records above
- `requestSourceChanged`: original before/after request texts differ bytewise
- `sourceChanges`: lexically sorted changed embedded source names among `bindings`,
  `catalog`, `model`, `requirements`, `scenarioReport`, `suite`. Optional report
  absence is distinct from presence; source formatting changes remain visible.
- `reviewRequired`: equal to `requestSourceChanged`, an identity-change flag,
  never evidence that review happened. Outer-wrapper formatting alone changes its
  own request hash but does not change this flag.
- `candidates`: only requirements with at least one reason, sorted by exact ID.

Each candidate contains exactly `requirementId`, `beforePresent`, `afterPresent`,
`beforeCaseIds`, `afterCaseIds`, `reasons`. Case lists are sorted, including empty
lists for absent requirements. Every reason contains exactly `code`, `side`,
`scenarioCaseId`. Deduplicate and sort by that tuple. Reasons are:

- `native-trace-impact`, side `both`, empty case ID: the requirement is a native
  declared-trace-impact candidate; its full reasons remain in the retained report.
- `scenario-link-added`, side `after`, or `scenario-link-removed`, side `before`:
  the explicit pair exists only on that side, including added/removed requirements.
- `scenario-case-added`, `scenario-case-removed`, `scenario-case-modified`:
  project native changed-case IDs through each side's actual links, with that
  side (`before` or `after`) and case ID. Native fields and steps remain in the
  complete scenario diff rather than being reinterpreted.
- `scenario-case-order-changed`, side `both`, empty case ID: native global case
  order differs and the requirement has a link on either side. This conservative
  rule also applies when adding/removing a case changes the global sequence.
- `scenario-suite-fields-changed`, side `both`, empty case ID: native suite fields
  (`id` or `modelDigest`) differ and the requirement has any link on either side.
- `scenario-observation-changed`, side `both`, case ID: the case exists on both
  sides and its validated observation status or ordered `(stepId, entryPosition)`
  list differs. Project through the union of that requirement's explicit links.
  This compares supplied observations, not actual execution. Added/removed cases
  already have declaration reasons; they do not receive this observation reason.
  A changed ordered step list can trigger this reason even when neither side has
  a supplied report: the retained observation layout still contains every step
  with a null entry position. It is not evidence of newly observed execution.
  Requirement references are excluded from this observation comparison. Different
  supplied actual payloads with identical validated statuses/positions are visible
  in the retained report/source identities, not as this narrow observation reason.

All arrays are nonnull. No ID rename is inferred, and no aggregate pass rate or
coverage-completeness claim is made. A source-byte change may yield an empty
candidate list; retained reports and `reviewRequired` keep that visible. Even a
fully consistent supplied report can be fabricated. Unchanged or omitted
requirements are not declared behaviorally unaffected.

Fixed errors, with nil output: `ErrInvalid` for structural/native input refusals
(including stale bindings), `ErrLimit` for explicit or native comparison output
limits, `ErrBoundary` for unexpected native encoding or cross-report inconsistency.
No input is echoed in diagnostics. The standalone command accepts no arguments,
reads bounded stdin and makes one complete stdout write. Exit 0 means a report
was produced, 2 means arguments/read/input/limit refusal, 3 means failed/short
delivery, and 4 means boundary failure. Ignore SIGPIPE to report a closed stdout.

Preserve native `scenariotrace.ErrLimit`, `scenariodiff.ErrLimit`, and
`traceimpact.ErrReportLimit` as `ErrLimit`. Native stale/input refusals become
`ErrInvalid`; `scenariotrace.ErrBoundary`, malformed native output or inconsistent
composed native identities become `ErrBoundary`.

## Acceptance and validation

Freeze independent hand-stated tests before production. The unchanged invented
ParcelReview native scenario-trace vectors may supply valid inputs; transformations
and expected link/change outcomes must be specified independently of this new
implementation. Native-only controls may check fixture validity, not derive the
new comparison expectations. Preserve the original vectors and native tests.

Cover identical snapshots, pure formatting, reordered bindings, added/removed
links and requirements, changed case initial/expected fields, global case order,
suite identity, model binding changes, absent/matched/early-mismatch observations,
unlinked changed cases, native report byte preservation, fixed strict refusals,
stale sources, deterministic ordering and no mutation. Add owned-buffer and race
checks, bounded input, and compiled CLI output/closed-pipe checks on this Mac.

Use installed Go 1.27.3 offline. Run focused package/CLI/native tests, vet,
generated-artifact validation/drift, repository validator, installer and committed
Project Memory checks. Review exact final source independently. Keep private
inputs out of Git. Record observed results and limitations without treating the
new report as authority, execution evidence or proof of general correctness.

## Observed local result

Implemented the pure Go comparison package and stdin-only command in the claimed
paths. They retain four native reports, compare explicit requirement/case links,
project native case edits through each side's actual links, and separately flag
changes to supplied observation layouts. Source formatting remains visible even
when it produces no candidate. Execution and approval remain not established.

The task source was pinned at `5c7ed4d3baa2f6931f11bc6b5219706178475b92`;
the genuine start was committed at `c45dc0d`. An independent contributor authored
29 positive vectors plus strict refusal, limit, retention and buffer checks before
production existed. All native-only setup controls passed. Acceptance was frozen
at `201eda9f4131dc56853009acc24d0603235ced75`, with SHA-256
`c4fd6ab174cf619b24540a06bda70e1e456fc4f74d9cdc2f826d8811f6f92788`, and
remained byte-identical through implementation and review.

Observed on the local Mac with installed Go 1.27.3, offline module settings:

- `go test -race -count=1 ./scenariotraceimpact ./cmd/modelscenariotraceimpact ./scenariotrace ./scenariodiff ./traceimpact` passed all five packages.
- Compiled CLI cases passed for complete output, a closed stdout pipe and stdin
  read failure; in-process checks passed for bounded reads, single writes and
  short/failed delivery.
- Native-valid input controls confirmed refusal at 513 catalog cases, an over-8
  MiB retained trace report and an over-16 MiB constructed trace-impact request.
- Focused `go vet` on the same five packages passed. New Go files are gofmt-clean.
- `go run ./cmd/modelc validate` and `go run ./cmd/modelc check` passed; generated
  artifacts matched without writes.
- `bash scripts/validate.sh` and `bash tests/test-install.sh` passed. Optional
  `skills-ref` was not installed, so its separate format check was skipped.
- Whitespace checks passed; existing native packages and fixture bytes remained
  unchanged.

A fresh reviewer inspected all eight scoped package/command files and the relevant
native contracts, finding no concrete blockers. Internal native-error mapping and
cross-report boundary refusals were inspected without direct negative injection
tests; ordinary public inputs cannot readily produce malformed native output.
No aggregate full-module run or additional platform testing was performed.

During this work main advanced to
`5bd2c310ecc97015b07b5233bac78e5d1ef5381f` through the custom assertion inventory
and workspace source-byte changes. Their changed paths do not overlap this task
or its native dependencies; no open PR remained at the last metadata check. This
branch keeps genuine ancestry from `01b18785e738d9cc656ea1fb622f8c9aad1b220a`.
At this first completion it had not been pushed, published, merged or run through
hosted CI. The next step was coordinator review of this local branch and a fresh
base/ownership check before any separately authorized publication.

## Local integration qualification claim

Qualify reviewed implementation `46350a2c0a9a7dd8c4eb769537a289c220a5f5f2` by
one ordinary local merge of main snapshot
`5bd2c310ecc97015b07b5233bac78e5d1ef5381f`. An SSH fetch confirmed this base.
The upstream changes are disjoint from the task paths and native dependencies;
repository instructions are unchanged. Do not rebase, amend reviewed commits or
publish. Preserve the frozen acceptance bytes and original native fixtures.

Rerun focused race tests and vet for the comparison, command and native packages,
model validation/drift, repository and installer checks, and committed lifecycle
validation on the combined branch. Review the exact integration diff and confirm
reviewed implementation blobs remain unchanged. Record the tested base, merge,
final head and limitations; do not chase subsequent main movement.
