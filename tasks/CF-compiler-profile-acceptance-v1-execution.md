---
id: CF-compiler-profile-acceptance-v1
status: review
owner: Byte compiler profile acceptance
repositories: [spik3r/clanker-foundry]
---

# Independently exercise the lifecycle compiler profile

## Implementation claim

The coordinator assigned this bounded local lane on 2026-11-02. Main was checked
at `55cc43870ad8017ecb334536a269e6ac9c49afee`, tree
`3bab0a7b8dc5944e23c3708352ceb890c7e2102d`; the only open PRs were #20 and #5.
Public tasks and branches and the available held local claims had no competing
compiler-profile execution corpus. Existing merged claim statuses are not changed.
Branch `codex/compiler-profile-acceptance-v1` uses an isolated worktree from that
exact main. This local claim is not a shared reservation or published task.
Publication, merge, deployment and remote changes remain outside this authority.

Byte owns integration, the harness and closeout. Independently delegated fixture
authors own invented requirements, models and literal oracle tests before the
harness is implemented or any generated reducer is run against the new fixtures.
A separate reviewer checks the frozen inputs and exact implementation source.
Commit this prospective task, register it and record a genuine start before edits.

Exact maintained paths under `modeling/lifecycle/compilerprofileacceptance/`:
- `README.md`
- `conformance_test.go`
- `testdata/requirements.md`
- `testdata/inventory.json`
- `testdata/integers/requirements.md`
- `testdata/integers/model.json`
- `testdata/integers/oracle_test.go.txt`
- `testdata/presence/requirements.md`
- `testdata/presence/model.json`
- `testdata/presence/oracle_test.go.txt`
- `testdata/dispatch/requirements.md`
- `testdata/dispatch/model.json`
- `testdata/dispatch/oracle_test.go.txt`
- `testdata/textdigest/requirements.md`
- `testdata/textdigest/model.json`
- `testdata/textdigest/oracle_test.go.txt`

Also this task and its own new immutable Project Memory source/events only.
No existing compiler, IR, generated artifact, acceptance fixture, consumer adapter,
runtime, workflow, toolchain, dependency, standard projection, visual artifact,
other task or historical evidence edits. No access to the denied #20 artifact.
If a production defect is found, report a minimal reproduction and seek a separate
repair lane rather than changing the compiler here.

## Contract

Add a test-only independent executable corpus, not an evaluator or runtime API.
Freeze invented model sources, their written requirements, external-package Go
oracle tests and hash inventory before harness implementation. The oracle asserts
complete literal public Results and Tasks, including identity, mode, authority,
nil/empty distinctions, refusal order and ownership. Expected canonical model and
task digests must come from independently encoded frozen bytes, never the reducer
or compiler output under test. Document how the independent hashes were computed.

The harness calls existing strict decoding and generation APIs, then compiles the
actual generated package with the frozen oracle in a temporary standard-library
only module. It must require the exact declared test inventory to pass without
skips. Generated structural tests are additional checks, not the independent oracle.
Use the installed Go 1.27.3 binary, local toolchain, disabled module network,
disabled ambient workspace/config/flags and VCS stamping, and bounded subprocess
time/output. This trusted repository test harness is not an arbitrary-code sandbox.

Cover uint64 eq/ne beyond float precision, optional operands, absent versus empty,
all/any composition, custom dispatch and multiple terminal states, self-loops,
assignment on a non-start trigger, validation/guard/overflow precedence, Unicode,
all frozen reference bindings, exact TaskDigest bytes, nonmutation and owned copies.
Use valid inputs to reach predicate witnesses. Input validation is distinct from
guard rejection; tautologies over valid inputs do not require impossible witnesses.
Check exact declared effect order structurally: commuting effects cannot generally
be distinguished by final state. Do not overstate runtime evidence as an order proof.

## Verification and next action

Before implementing, separately review and commit frozen requirements/models/tests
and inventory. After implementation, run each positive oracle and inspect exact
test results. Apply disposable generated-code mutants, require them to compile and
then fail a relevant named behavioral assertion, restore original bytes and rerun.
Equivalent mutants, compilation failures, timeouts or unrelated failures are not
behavioral kills. Record actual witness coverage and remaining gaps rather than
claim exhaustive proof or standards conformance.

Run focused and full Go 1.27.3 race tests, full vet, modelc validate/check, modeltrace
check, formatting, repository validation, Project Memory regression tests and fresh
committed-diff validation against the verified base. Inspect the full changed-path
frontier and byte-preservation of existing code. Obtain independent exact-source
review. Complete local evidence with observed results, then hand the unpushed branch
and next integration action to the coordinator. Hosted CI and publication are not
implied by local completion.

## Local implementation and independent review

The prospective claim was committed at `af70a9854fd059d1b8ff30ad392292035371d824`.
Its reviewed source was registered and the genuine start recorded before fixture
edits. Two separate authors froze four invented model families and independent
public-API oracles; a separate reviewer recomputed all four canonical model hashes,
eight task-digest vectors, thirteen input-file hashes and the exact test inventory
without executing a new reducer. Freeze commit
`4d34bdf` preceded the harness and all new-model execution. Inventory SHA-256 is
`e96013d6d9d5d2bf603d3dd473075de47a5d89d4a3de633dc5b483ff767e663c`.

Implementation was reviewed at `1e84bc89c51ada598c72f858215b0fb00bbd0306`, tree
`0ba79e0756101092cf2dc90e69947df0f1b6e9a3`. The independent reviewer found no
blocking issue and separately ran the focused Go 1.27.3 race test. This is
independent assistant review, not authenticated human approval. No compiler defect
was observed, and no frozen expected value was refreshed after execution.

Observed results on that source:
- All 154 independent leaf cases across eleven top-level tests pass, with exact
  run/pass inventories and no skips. Four model identities and all fifteen profile
  references are exercised through public APIs.
- Fifteen disposable generated-code mutants compile, then fail their specified
  named semantic assertions. Exact restoration passes every frozen case.
- All 31 transition effect sequences match frozen declarations. A commuting-effect
  reorder compiles and passes behavior but fails the separate structural check.
- Full offline Go 1.27.3 race suite passes: 49 packages, 48 with tests and modelc
  with no test files. Full vet, modelc validate/check, modeltrace check, formatting
  and whitespace checks pass. Existing base paths and frozen fixture bytes remain
  unchanged.
- Repository validation passes; optional skills-ref is not installed. All 143
  Project Memory regression tests pass. Fresh committed-diff coverage is checked
  against the verified base before handoff.

The default home Go cache was not writable; tests used a separate authorized cache.
The first full race run encountered the existing clean-regeneration test's host
VCS-stamping failure under an external temporary directory. The full suite then
passed with temporary files under this valid isolated Git worktree. No existing
test, VCS policy or host security setting was changed. Temporary work was removed.

Coverage remains finite: nonempty slice/owner ownership probes do not prove every
possible alias shape, and final Task behavior does not prove effect order. The
harness is trusted repository test code, not a sandbox for arbitrary caller input.
No existing production file, dependency, workflow, visual artifact or other claim
changed. Main was rechecked at the same base before closeout. This branch remains
local and unpushed; publication and hosted CI require the coordinator's next action
and applicable authority. No integration, deployment or standards conformance is
claimed. Held #20, #5 and UML owner clearance remain separate.
