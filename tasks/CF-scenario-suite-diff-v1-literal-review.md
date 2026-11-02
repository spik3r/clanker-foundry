---
id: CF-scenario-suite-diff-v1
status: review
owner: Byte scenario suite comparison contributor
repositories: [spik3r/clanker-foundry]
---

# Review changes to literal model-bound scenario suites

## Implementation claim

The coordinator assigned this local lane on 2026-11-02. Independently checked
main `fcda4be584ebbbdf84239a5b849b1e6b8fba2b9b`, tree
`bb13598e1434f7646f9f3dd5afee63608c025a04`, live branches, task records and
Project Memory frontiers. Model merge (#39) and assertion inventory (#40) are
already integrated. The inventory owner confirmed that suite-edit comparison is
not its scope. No competing suite-comparison claim was found. Structural graph
analysis is another contributor's lane. Existing task statuses are not rewritten
or interpreted as transferring ownership. This branch is
`codex/scenario-suite-diff-v1` from exact canonical main, not synthetic ancestry.
This local claim is not a shared reservation or authorization to publish.

Byte owns integration and the following exact new frontier under
`modeling/lifecycle/`:
- `scenariodiff/types.go`
- `scenariodiff/compare.go`
- `scenariodiff/decode.go`
- `scenariodiff/compare_test.go`
- `scenariodiff/decode_test.go`
- `scenariodiff/oracle_test.go`
- `scenariodiff/README.md`
- `cmd/modelscenariodiff/main.go`
- `cmd/modelscenariodiff/main_test.go`

Also this task and its own new immutable Project Memory files only. An independent
fixture author owns `oracle_test.go`, freezing literal expected comparisons before
comparison logic exists. Commit/register this prospective task and record a genuine
start before editing maintained implementation, tests or documentation.

Excluded: existing core/model/compiler/reducer/runner/fixture files, scenario
inventory, graph analysis, semantic diagnostics, UML projection, #20 artifact,
#5, all other task/evidence, dependencies, workflows, runtime, deployment and
remote publication. Do not access denied artifacts. No ownership is released.

## Contract

`Compare(beforeModel model.Model, before scenarios.Suite, afterModel model.Model,
after scenarios.Suite) (Report, error)` compares declarations, not executions.
Validate each explicit model with existing Normalize and JSON-v1 encoding and
compute its existing canonical digest. Validate each entire literal suite against
that side's derived model identity using existing bounded scenario validation.
The private validation-only runner uses a panic sentinel; no Run/Execute/reducer
or caller callback is invoked. The derived binding establishes input consistency,
not a compiled-code identity, executable binding or provenance. Do not export it.
Invalid models, suites or stale bindings return zero Report and fixed ErrInvalid,
without caller-controlled text. Typed inputs must be JSON-v1 representable.

Do not evaluate expected outcomes, require expected transition/refusal IDs to exist
in the model, repair inputs, refresh expectations or infer compatibility, coverage,
correctness, approval, ancestry or authority. Deliberately invalid business inputs
remain legal fixtures when the existing scenario schema accepts them.

Report version is `lifecycle-scenario-diff/v1`, profile `lifecycle/v1`, mode
`example-only`, authority `none`. Include before/after model and suite digests.
`modelChanged` and `suiteChanged` compare those canonical identities;
`reviewRequired` is their OR, a change flag rather than approval policy.
`suiteFields` lists changed root fields (`id`, `modelDigest`) lexically.
`caseOrder` always contains complete before/after ordered case ID arrays and
`changed`, which compares only those arrays. Suite case/step order is semantic.

`cases` contains added/removed/modified cases sorted by case ID. Added/removed
cases have empty `fields`. Modified cases list `initial` for any full initial
Task difference, and `steps` for any ordered Step sequence/content difference.
Each case entry includes complete beforeSteps/afterSteps ordered ID arrays;
an absent side is empty. Unchanged cases are omitted. Mere case reordering does
not produce case entries. An ID rename is remove plus add, never inferred.

`steps` contains added/removed/modified steps sorted by (case ID, step ID),
including every child of an added/removed case. IDs are paired within their case;
a step moved between cases is removed plus added. Pure step reordering changes
its case but produces no unchanged step entry. Added/removed entries have empty
`fields`. Modified fields are lexically sorted, at these atomic boundaries:
`command.actor`, `command.expectedRevision`, `command.review`, `command.type`,
`expected.ok`, `expected.refusals`, `expected.task`, `expected.transitionId`.
Full initial/expected Task and command Review values are atomic. Pointer presence,
nil versus empty task acceptance, ordered lists and all uint64 values matter.
The existing schema rejects nil expected refusals. No literal payload is copied
into the report: reviewers retain the separately identified input suites.

All output arrays are non-null, own their storage and preserve only specified
orders. Callers' immutable inputs are untouched; concurrent input mutation is
unsupported. Errors contain no partial output. Typed Sources is null.

## Wire and command

`Execute([]byte) ([]byte, error)` accepts exactly formatVersion
`lifecycle-scenario-diff-json/v1`, beforeModel, beforeSuite, afterModel, afterSuite.
All fields are required/non-null. Reuse strictjson, jsonformat and bound scenario
Decode; reject unknown/duplicate/case-variant/missing keys, malformed Unicode,
trailing data and unsupported formats. Wrapper limit 5 MiB/depth66; each embedded
model and suite retains existing 1 MiB limits. Report limit 1 MiB. Wrapper byte/output byte
limits return fixed ErrLimit; structural/depth and other invalid input return
fixed ErrInvalid.

Wire output recomputes the comparison and adds SHA-256 Sources for exact request
and each embedded JSON token, separate from semantic digests. Whitespace can alter
source hashes without altering semantic change classification. Hashes establish
consistency, not origin. Fixed-order JSON, HTML escaping disabled, one final LF;
identical inputs yield byte-identical output. No arbitrary-report encoder.

`modelscenariodiff` accepts bounded stdin only, no arguments/paths. Buffered stdout
contains one complete report on successful computation (0 even when changed).
Argument/read/validation failure exits 2; failed/short output exits 3, including
real closed pipes. Diagnostics never echo input. No file/network/subprocess or
execution operations; buffering does not guarantee atomic writes to a stream.

## Verification plan

Freeze independently authored literal expected reports before comparison logic.
Test no-op, root/model-only changes, all atomic fields, additions/removals, order,
move-as-remove/add, maximum integers, nil/empty task fields, binding rejection and
normalization. Check output ownership, deterministic concurrent read-only use,
strict input/size/depth/Unicode boundaries, source identities and real CLI failures.
Use independent behavioral mutants and exact-source review. Run focused and whole
module offline Go 1.27.3 tests/race/vet, generated model check, repository validator,
Project Memory regressions and committed-diff lifecycle validation. Record actual
results, not inferred success. Remain local pending separate publication authority.


## Observed local result

Implementation/source commit `4a3277d822719c898b412446b0f04454b47b165a` was
independently reviewed with no code blocker. The reviewer noted an ambiguous
limit phrase above; this task now distinguishes byte limits from structural/depth
invalidity, matching the implemented checks and existing refusal tests. No logic
or expectations changed for that clarification.

Independent oracle commit `0892adc32224eb1209a6702ad32edbda15fa61d9` predates
Compare and freezes 28 complete literal reports with independently Python-computed
model/suite digests. File SHA-256 remains
`bbfe5de8d87dc1a577ebee738f846c1a0a28be70c56d66e0872ae7c5485c0600`.
Its input self-check ran with a panic-only Compare stub and never invoked comparison.
The unchanged oracle passed against implementation. An independent disposable wire
probe passed all 28 reports and all five exact source hashes, excluding external
token whitespace. Fifteen independently authored compiling behavioral mutants were
killed: eight field omissions, case order, initial task, two root fields, sort order,
added/removed-case children and null empty arrays. These are behavioral checks, not
claims that expectations describe actual reducer results.

Observed checks on unchanged implementation bytes, offline Go 1.27.3:
- `GOMAXPROCS=2 go test -p 1 -race -count=1 ./...`: all 54 tested packages passed.
- Independent bounded focused race for scenariodiff and its CLI passed.
- `go vet -p 1 ./...`, `go run ./cmd/modelc validate` and `go run ./cmd/modelc check`
  passed; generated artifacts match, no writes.
- Ten-second bounded fuzz pass completed 172,834 inputs without failure.
- Repository validator passed; optional skills-ref was unavailable, not run.
- Project Memory regression suites: 143 passed, zero failures/skips.
- README no-change CLI example passed; gofmt and whitespace checks passed.

An initial local test invocation used the module directory with root-relative
creation paths and therefore failed to compile missing Compare. No file was written
by that mistaken creation command. The corrected path and all actual final checks
above passed; that early build failure is not recorded as a pass.

Current canonical main was rechecked at the same `fcda4be`/`bb13598e` identities;
no integration delta existed at that observation. No existing source, fixtures,
runner, compiler, model artifacts, other tasks or legacy history were edited.
Final committed lifecycle validation and exact metadata/documentation delta review
follow this local completion record. No remote publication or deployment occurred.
