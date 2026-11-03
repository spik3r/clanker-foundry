---
id: CF-adapter-conformance-v1
status: complete
owner: Byte adapter-conformance extraction worker
repositories: [spik3r/clanker-foundry]
depends_on: [CF-model-bound-runner-v1]
---

# Extract reusable adapter conformance checks

## Implementation claim

The coordinator cleared this bounded extraction after the completed consumer
integration owner explicitly released its three paths. Current main is
0c1beec1f441aed8943e7b8af007991b5a278020, tree
b0f766491b10d016afe5c20bc71b25cfde927cb2. Open PRs and live task records were
checked; this is a separate branch, codex/adapter-conformance-v1.

Exact frontier under modeling/lifecycle:
- New adapterconformance/README.md, conformance.go and conformance_test.go.
- Coordinated consumerintegration/README.md, consumer_test.go and
  testdata/adapter.go.txt changes only.
Plus this task and its own immutable Project Memory source/events.

Preserve all eight frozen consumer requirements/model/suite/manifest/inventory
files byte-for-byte. Do not change compiler, generated source, IR, runner,
acceptance APIs, CLIs, workflows, runtime, or any held consumer source/pin/fixture.
No private consumer diagnostics or policy access, deployment or worker activation.
This does not adopt or qualify a private consumer or the held reviewpack stack.

## Bounded contract

Extract optional test-only assertions, not another scenario suite or runtime
adapter protocol. The generic API is Adapter[T,C,R] with four typed callbacks:
ToTask(transport.Task) T, FromTask(T) transport.Task,
ToCommand(transport.Command) C and FromResult(R) transport.Result. Transport
types remain those of the existing generated package. Check(adapter) returns
nil or the first conformance error. It has no Transition callback, model binding,
runner construction, report serialization, filesystem, network or process I/O.

T, C and R must be concrete generated-shaped structs. Reflection checks the
complete ordered Task, Command, nested Review and Result field inventory and
types before invoking any callback. Result.Task must point to T. State permits
the generated named string type; other scalar, pointer and slice shapes match
the existing transport. Missing, extra, embedded or unexported fields and wrong
types are refused. Nil callbacks are refused before invocation. Errors are test
diagnostics, not a versioned machine protocol or evidence certificate.

Fixed sentinel cases independently check each conversion direction. An inverse
adapter must never construct an expected value or validate the other direction:
paired field swaps can cancel in a round trip. Literal named expectations must
cover each task/command/review/result field and preserve actual mode, authority,
model digest, success flag and transition identity without repair. Use distinct
nonempty field strings, zero and MaxUint64, and distinct expected/task revision
sentinels so swapping revisions cannot pass. Check nil, empty and populated
lists/owners/reviews/tasks, ordered duplicate refusals, and nested ownership.
Inputs may not be mutated by conversion; source and result mutations must not
cross slice/pointer boundaries. Each probe constructs fresh inputs/expectations.

Callbacks are trusted test code. They can panic, block, allocate, have side effects
or recognize sentinels. Check does not recover panics, sandbox, cancel or prove
arbitrary-input correctness. Passing these finite probes establishes neither
real policy correctness nor authenticated generated-source provenance.

Replace only the duplicated adapter assertions in the existing external-consumer
template with a helper call. Keep its field-only adapter and real generated
Transition calls, both fourteen-step frozen suites, binding/protocol/preflight
checks, and both generated-code behavioral mutants. Preserve installed-toolchain
selection, offline environment, GOENV=off, -buildvcs=false, bounded subprocess
waits and the existing lack of a descendant-cleanup guarantee.

Add disposable adapter mutations for coupled swaps, drops and nested aliases
against both distinct generated models. A mutation counts only if it compiles
and produces the expected conformance failure marker; compile failure does not
count. Restore the original adapter and require passing conformance afterward.
No frozen expectation is derived, rebound or refreshed from actual replay.

## Independent acceptance and checks

An independent worker authors the helper's one-way tests before implementation
logic, with a separate correct field-only adapter and wrong adapters for swaps,
drops, aliases, nil/empty changes, narrowing and refusal-order changes. Freeze
these tests before helper logic. Then run focused and full pinned Go 1.27.3
offline race/vet, existing model/trace drift checks, repository validation,
frozen-file continuity and full committed-diff Project Memory validation. Obtain
independent exact-source review and record actual limits and failures.

Publication, merge and any later consumer adoption remain separately authorized
actions. A local completion never establishes those actions occurred.

## Implementation and observed review

The optional helper and three-file adoption were independently source-reviewed
at a0b257f410ba5459194f170e1067bda955f2cb26, tree
e452d677dfd7ab992b5ba04a93489e42a01a1455, with no remaining source finding.
The original independent tests were frozen at 5680ea1 before helper logic; their
bytes remain an unchanged prefix (SHA-256
f1fe69034862e09d20c1149d219453f11086e21d06327a8c4c1d37cb94b70764).

Independent review exposed a real value-only comparison gap: an adapter could
mutate spare slice capacity or replace caller-visible nested storage with equal
values. A separately authored addendum was frozen at bfa503c before correction,
showing seven failing storage regressions and four passing ordinary-mutation
controls. Pre-call snapshots now compare pointer/slice identity, length, capacity
and full capacity contents as well as independent literal values. The combined
test file is SHA-256
4920d95ee7774764db519dc89aeb62f855683e7ee237f0393e59cd5c141936bc.
The reviewer independently confirmed the correction and a shifted-empty alias
probe. No original expected value was refreshed from adapter or reducer output.

Observed checks on the corrected source:
- Actual offline Go 1.27.3 full-module race tests passed (36 tested packages),
  as did full vet and modelc/modeltrace validate/check. A separate fresh-cache,
  isolated valid Git clone also passed all of these checks.
- Both external generated models still execute their seven-case/fourteen-step
  frozen suites. All fourteen adapter mutations and the two original generated
  reducer mutations were behaviorally rejected; restored adapters passed actual
  conformance calls for both models.
- All eight original consumer fixtures remain byte-identical to 8490187. The
  expected inventory remains SHA-256
  119a37c0298dea6e122c3a4f62b2cb1bddbac24e6fc91feadbc2723360e9550f.
- Repository validation, formatting/whitespace, 143 lifecycle-helper regressions
  and two authoring-schema checks passed. Optional skills-ref was unavailable.
  Schema checks reused available dependencies with an identical lockfile, not a
  claimed fresh install.

Initial full runs hit the existing clean-regeneration test's ambient invalid
ancestor Git metadata in scratch locations. Using an ignored repository-local
temporary directory allowed that unchanged test to pass. No host security policy,
existing test, compiler, workflow or process-lifecycle contract was weakened.
Existing traceability gaps remain visible and are not resolved by this helper.

This completion records local extraction, tests and independent assistant review,
not arbitrary adapter correctness, human approval, source authentication, private
consumer adoption, publication, merge, deployment or worker activation.


## PR review repair

Independent review of published draft PR 29 at 5641393 found that comparing
ownership only along matching field paths missed cross-field aliases. The first
reproduction reused empty input Task.Acceptance as output Refusals; twelve probe
calls reached that one pattern. Five independently authored regressions cover
both slice directions and Owner pointers into another field's spare capacity.
They were reproduced before repair; none changes a frozen scenario expectation.

The existing owner/coordinator authorized a new lifecycle run within this task's
same seven-file frontier. Compare every reachable input mutable-storage region
against every output region, regardless of field path or pointer/slice kind,
including pointer referent sizes and full slice capacity. Keep original objects
live during address comparison. Preserve prior completion/history, rerun both
model proofs and independent exact-source review, and do not merge this draft.

The repair passed independent exact-source review at
9967e06e7ec471a72f57b41c41f26396875e799b, tree
0c9ad5b543929936066dad7550aa8e54ce45e395. The five new regression cases
were frozen at c445b3f before correction; the complete test file remains SHA-256
48b6548d8fdf566662e62a9d7fc841a9e1ed69119fb0f14d37b306377a11935b,
with all previous test bytes unchanged as a prefix. All-region comparison now
rejects the original false negative on its first reached alias. An independent
second reviewer also checked shifted empty subslices, final spare-slot pointers,
forced GC and adjacent fresh storage without false positives.

Corrected-source full offline Go race tests (36 tested packages), vet,
modelc/modeltrace validate/check, repository validation, formatting and schema
checks passed. The exact-source reviewer independently reran the full Go checks.
The external proof now kills eighteen adapter mutants, including cross-field
lists and pointer-into-list aliases for both generated models, plus the original
two reducer mutants. Restored adapters and both frozen suites pass. All eight
consumer fixtures remain unchanged. Previous lifecycle events and completions
are preserved; this new completion records the reviewed repair only. Updated
draft publication, hosted checks and any merge remain separate coordinator gates.
