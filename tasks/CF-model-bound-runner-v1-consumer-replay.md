---
id: CF-model-bound-runner-v1
status: review
owner: Byte model-bound consumer integration
repositories: [spik3r/clanker-foundry]
depends_on: [CF-model-scenarios-v1, CF-acceptance-authoring-v1]
---

# Replay a separately generated model through the shared pipeline

## Implementation claim

The coordinator approved this bounded producer-side integration after model/scenario
and acceptance-authoring owner coordination. Branch `codex/model-bound-runner-v1`
starts from main `db052febb9e5c18fd0d0b3701e016e265d3980fc`, tree
`33470a500682954091e0e8dc17c6affbe1a9a24c`. Current task records and the actual open
PR inventory (#20 and #5) were inspected. The existing owners cleared this explicit
opt-in frontier; historical claim statuses are not silently rewritten. This claim
is initially committed locally, not represented as already published remotely.

The integration owner owns maintained code and closeout; a separate contributor
owns frozen invented custom-model expectations before implementation. A separate
reviewer checks exact source and compatibility. Commit/register this reviewed
prospective record and genuinely start before any maintained implementation.

Exact paths under `modeling/lifecycle/`:
- `scenarios/types.go`, `scenarios/decode.go`, `scenarios/run.go`
- `scenarios/run_test.go`, `scenarios/README.md`
- `scenarios/runner.go`, `scenarios/runner_test.go`
- `scenarios/testdata/legacy-parity.json`
- `acceptanceauthor/author.go`, `acceptanceauthor/README.md`
- `acceptanceauthor/bound.go`, `acceptanceauthor/runner_test.go`
- `consumerintegration/README.md`, `consumerintegration/consumer_test.go`
- `consumerintegration/testdata/requirements.md`
- `consumerintegration/testdata/model-v1.json`
- `consumerintegration/testdata/model-v2.json`
- `consumerintegration/testdata/suite-v1.json`
- `consumerintegration/testdata/suite-v2.json`
- `consumerintegration/testdata/manifest-v1.json`
- `consumerintegration/testdata/manifest-v2.json`
- `consumerintegration/testdata/expected.json`
- `consumerintegration/testdata/adapter.go.txt`

Also this task and its own new immutable Project Memory source/events only.
No compiler, existing generated artifact, IR, schema, existing CLI, tutorial,
revision walkthrough, workflow, runtime, dependency/package, other task or evidence
changes. No consumer repository adoption, canonical consumer claim, deployment,
worker activation or denied-action retry. The held reviewpack/diagnostics and
qualifier/catalog copies remain separate and require future explicit requalification;
their pins must not be refreshed here. The held consumer shadow adapter and its
immutable-invocation and producer-correctness gates remain untouched.

## Contract and compatibility

Expose a value Runner with private snapshotted profile/model identity and a typed
transition callback. Only lifecycle/v1 is supported. Reject invalid construction
and every zero-runner operation, never default silently. The caller composes it
in reviewed Go code using a generated reducer's constants; no request-selected
provider, dynamic loading, registry or mutable setter is introduced. A supplied
identity is a declaration, not authenticated compiled-code provenance.

Reuse existing strict codecs, canonical identities, bounds, whole-suite validation,
owned copies and literal comparisons. Scope every encode/decode/digest/report check
to the same runner. Invalid final steps and stale identities invoke zero callbacks.
No expectation refresh or business-rule interpreter is added. Existing package APIs,
private runWith policy, CLI behavior, errors and report bytes remain default-bound
and unchanged. Compiler output and all existing generated files remain byte-identical.

For the opt-in path, validate/bound a returned result before cloning it. Wrong
model/mode/authority is a distinct provider protocol failure after invocation, not
an expected mismatch or not-run. Profile is constructor metadata; Result has no
profile field. Ordinary literal mismatches stop before another callback. No returned
identity or result is overwritten. Callbacks are trusted application code: they may
panic, block, allocate or perform I/O. Panics propagate. This API neither sandboxes
nor cancels callbacks, bounds their allocation, rolls effects back, nor promises
purity/determinism of arbitrary closures.

Add an opt-in acceptance BoundReport and version with runnerModelDigest and explicit
caller-supplied-go-binding identity source, never compiledModelDigest. Bind manifest,
canonical model, exact suite bytes, suite model and runner before full validation or
replay. All-current check-only input fully validates the suite and case IDs and
computes its scenario digest, reporting valid/not-run. Stale bindings remain
not-checked/not-run. Both paths invoke zero callbacks.
Propagate provider failure distinctly; it may occur after execution began. Existing
Check/Execute and their CLI retain the original default report contract.

## Independent proof and validation

Freeze separately authored invented requirements, two custom model revisions,
literal suites/manifests and expected identities/results before implementation.
No private consumer data is copied into this public repository. Baseline byte-parity
snapshots are compatibility evidence, not an independent business oracle.

An external temporary consumer module imports Foundry rather than copying its source
or replacing its generated package. The unchanged compiler emits a separate reducer.
A field-only adapter preserves every task, command, review and result field, uint64,
null/empty distinction and refusal order; all semantics come from that generated
Transition. Cover custom success/refusal/review, A/B identity isolation, stale and
malformed preflight zero calls, protocol faults, fail-fast comparisons, ownership,
size bounds, panic propagation, source versus canonical identity and adapter shape.
This proves an offline lifecycle-profile seam, not real task admission or approval.

Run legacy byte/error parity, focused and full Go 1.27.3 race/vet, compiler/trace
drift, authoring schema, repository validation, Project Memory regressions and fresh
committed-diff validation. Check behavioral mutants against frozen expectations.
Record actual results and any limitations; prepare one draft with full history for
independent review, not activation or deployment.

## Implementation and independent review

Implemented at `238310af37cfda0d2c599ad992f2f01790f55b05`, tree
`419c6602ecf0c9e2e407987cce9884728325f58f`. A separate model/scenario reviewer
read that exact source and found no remaining blocking issue. Its independent
same-thread closure probe confirms that callback mutation cannot rewrite later
frozen expectations or the observed pre-invocation source digest. The opt-in
runner snapshots the complete validated suite before invoking application code.
This is independent assistant review, not authenticated human approval.

The separate fixture author froze all eight custom inputs at `8490187`; the
expected inventory remains SHA-256
`119a37c0298dea6e122c3a4f62b2cb1bddbac24e6fc91feadbc2723360e9550f`.
Both distinct generated models match seven cases and fourteen steps, including
review/rework and the intentional changed-actor rule. Two disposable generated
code mutants compile and fail the frozen oracle behaviorally; restoration passes.
Four additional scratch runner mutants are behaviorally killed: missing literal
snapshot, source hashing after callbacks, missing provider protocol checking and
silent suite rebinding. No frozen expected result was refreshed.

Legacy byte-parity inventory was frozen before implementation at `2635a04`, SHA-256
`58194fa32fc0d929248892533ca21e8188c6546def68a393cab630f4f9385fa8`.
The final source passes those hashes and 947 separate baseline-versus-final CLI
comparisons (596 modelplay, 351 modelacceptance), including exact exits/stdout/stderr.
Existing compiler, generated files, IR, schema, CLI source and runtime are unchanged.

Actual Go 1.27.3 offline full race tests pass: 36 packages total, 35 with tests and
modelc with no test files. Full vet, modelc validate/check, modeltrace check,
authoring schema (two tests), repository validation, 143 Project Memory regressions,
formatting and whitespace pass. Independent reviewer full-module race/vet and
external-module replay also pass on the exact implementation source.

The host's default `/tmp/.git` is not a valid repository and caused the pre-existing
clean-regeneration test's VCS-stamped build to fail. Final full runs use an ignored
task-local temporary directory instead; no host/security setting or unrelated test
was changed. The new external-consumer harness itself disables VCS stamping and
ambient Go configuration. A fresh offline npm install lacked a cached tarball;
schema tests used already available Ajv 8.20.0 dependencies whose lockfile matched
exactly, not a claimed fresh install. Optional skills-ref was unavailable.

Local implementation is complete and remains for draft publication and hosted CI.
This producer seam is not consumer adoption, trusted immutable invocation, real
review/approval, deployment or worker activation. Held consumer correctness and
invocation gates and copied-source qualification pins remain unresolved here.
