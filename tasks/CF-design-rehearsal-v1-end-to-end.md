---
id: CF-design-rehearsal-v1
status: in-progress
owner: Codex end-to-end design rehearsal contributor
repositories: [spik3r/clanker-foundry]
depends_on: [existing lifecycle model, patch, scenario diff, topology, compiler, qualification and runner]
---

# Rehearse a model change from declared intent to compiled replay

## Prospective implementation claim

This local-only lane was assigned after current main `014eaa2163929a3fb4db4993bb38d59ba0b2aa69`
(tree `81464dcd2d4f867a4cc4a3eebe14a0fb36351822`) and visible ownership were
checked. Existing `CF-lifecycle-tutorial-v1` reserves its own `tutorial/` paths;
this separate fixed demonstration does not edit them. #20 and its denied visual
artifact, #5, UML/projection, and all other active claims are excluded. A local
claim is not a shared lock; recheck current ownership before any separately
authorized publication. No push, PR, merge, deployment or worker activation is
authorized by this local task.

Exact new maintained paths only, plus this task's own immutable Project Memory
source and events:

- `tasks/CF-design-rehearsal-v1-end-to-end.md`
- `modeling/lifecycle/rehearsal/README.md`
- `modeling/lifecycle/rehearsal/run.sh`
- `modeling/lifecycle/rehearsal/rehearsal_test.go`
- `modeling/lifecycle/rehearsal/testdata/baseline-suite.json`
- `modeling/lifecycle/rehearsal/testdata/expected.json`

After committing this public-safe source, register it and record a genuine start
with six separate paths before any implementation. A separately assigned
contributor authors and freezes the literal baseline suite and expected review/
replay observations before this workflow is implemented or run. Preserve the
existing independently authored `designchange/testdata` request, candidate suite
and expected identities unchanged; never derive expected results from a reducer
or rewrite/rebind the suites. Label separate agent review honestly, not as a
human approval or independent-person attestation.

## Outcome and boundaries

Provide a reproducible, fixed-fixture local rehearsal that actually joins the
existing pieces in an isolated temporary module: patch planning and model impact,
scenario-suite declaration comparison, structural topology, a reproducible
qualification archive and consistency verification, compiler generation, and a
candidate-bound executable scenario replay. Demonstrate a separately frozen
baseline suite matches the baseline reducer, stale candidate binding refuses
before baseline execution, and the baseline binding refuses the rebuilt candidate
runner. The separately reviewed candidate suite must execute against the rebuilt
candidate and match its literal outcomes, including the intended `T_prepare`
owner-guard refusal and positive case. Check fixed expected model and suite
identities, key review report fields, topology, archive status and actual replay
against independently frozen literals; no automatic expectation refresh or model
binding substitution. A stale refusal is not a failed executed scenario.

The script is explicitly a local **demo orchestration** on a trusted host, not a
production service, sandbox, scheduler, authenticated review or worker. It may
run fixed repository tools/Go subprocesses and use a private disposable directory;
it never runs arbitrary user-provided commands or downloads dependencies. Go
module/network resolution is disabled and the exact toolchain is required. The
archive verifies consistency only and reports its suite remains unvalidated and
unexecuted at that packaging boundary; separate replay is a distinct observed
step. Generated source is executable code, and local compiler/host are trusted;
no sandbox or child-process containment claim. Supplied inputs and output can be
private. No automatic approval, source mutation, diagram rendering, standards
conformance, Git push, deployment or activation. AI-drafted model/requirements
remain proposals requiring human review appropriate to their process.

`run.sh` operates only on the fixed example fixtures, uses a new private temporary
directory, has explicit cleanup behavior, and never overwrites the checkout.
It fails closed on stale or mismatched evidence, short/failed commands or
unexpected report fields. The integration test invokes the actual documented
script and independently asserts externally frozen outputs and source checkout
nonmutation; a failure or missing tool is reported, not silently skipped except
where an optional prerequisite is explicitly documented. No unbounded or
untrusted subprocess execution is introduced.

## Validation and handoff

Freeze independent literals before implementation and obtain separate exact-source
review. Run the real rehearsal, focused tests and full Go 1.27.3 race suite, vet,
modelc/modeltrace checks, repository validator, whitespace and committed-diff
Project Memory validation. Demonstrate compiling mutation/negative controls for
at least stale binding and changed candidate behavior. Report separately what was
observed, what remains an inert assertion, optional prerequisites and hosted CI.
