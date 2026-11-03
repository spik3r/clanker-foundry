# Rehearse concurrent designs through compiled replay

This fixed, invented parcel example connects **two design edits → actual three-way
merge → generation → compiled candidate replay**. It uses the existing merger,
compiler and runner unchanged. It accepts no caller models, paths or commands.
No AI/provider call, automatic conflict resolution or expectation repair occurs.

From the repository root, with Bash, Python 3 and installed Go **1.27.3** on PATH:

```sh
modeling/lifecycle/mergerehearsal/run.sh
```

Use writable local temporary storage and a writable Go cache. `TMPDIR`, `GOTMPDIR`
and `GOCACHE` may select those locations. The script checks the exact installed
toolchain, starts Python in isolated mode, clears ambient Go workspace/config/flags,
disables toolchain/module downloads and runs fixed subprocesses with 180-second
deadlines. It creates a private temporary module copy, changes only that copy, and
emits one deterministic JSON receipt only after cleanup succeeds. It neither
renders nor visually reviews any generated view. An abrupt kill or cleanup failure
can leave temporary files; a failed stdout write can expose a prefix of a receipt.

This runs trusted repository code, the installed compiler and generated Go on a
trusted single-writer host. It is not an OS sandbox, resource containment system
or guarantee of descendant termination. Subprocess cancellation does not establish
that every descendant stopped. Reports and real replacement inputs could contain
private data; this fixed public example contains invented data only.

## Why both branches matter

The [requirements](testdata/requirements.md) describe Raw → Packed → Dispatched.
The base permits any nonblank actor at each step. Left changes only the used
`G_packActor` record to require literal `packer`; right changes only the used
`G_dispatchActor` record to require literal `dispatcher`. The merged model must
retain both edits. Model labels are assertions, not authenticated actors.

| Witness | Base | Left | Right | Merged |
| --- | --- | --- | --- | --- |
| Guest packs at Raw | success | G_packActor | success | G_packActor |
| Guest dispatches at Packed | success | success | G_dispatchActor | G_dispatchActor |

These are separate transition witnesses. No combined-only same-input witness is
claimed. Each suite also checks the complete happy path, terminal refusal, stale
revision, whitespace actor, revision overflow and unknown-trigger invalid input.
Each has seven cases and nine steps, including complete retained/null tasks.

## Observed stages and refusals

1. Check the frozen oracle and all twelve original source-file hashes. An
   independent assistant derived all expected values from the requirements and
   public serialization contracts before workflow implementation. It did not use
   merger, generator or runner output as its oracle. A separate reviewer checked
   those literal documents and identities. This is separation between assistants,
   not independent-person verification, authenticated provenance or human approval.
2. Build and invoke the actual `cmd/modelmerge` CLI with six exact frozen request
   strings: both orientations of success, atomic conflict and invalid combination.
   Compare the **whole report**, including complete candidate/conflicts and source
   identities, to the independent literals. The competing pack rule conflicts;
   two distinct added transition IDs collide at one state/trigger after combination.
3. Inspect status despite exit 0. Conflicted and invalid-combination reports must
   expose no candidate and stop before generation/replay. No partial design is
   repaired or selected. Tests instrument this control flow and reject corrupted
   full reports that retain the expected status.
4. For each successful orientation, install its actual returned candidate only in
   the private copy, run native generation/check and generated structural tests,
   build unchanged `modelplay` against that reducer, and compare the complete
   actual replay report with the frozen merged report. Separately generate and
   replay base, left, right and merged fixtures to observe both branch witnesses.
   All four reports preserve raw suite identity separately from semantic identity.
5. Show stale suite refusal by the actual candidate-specific CLI. A separate
   callback-spy regression checks all twelve cross-model suite/binding pairs:
   each refuses with zero callbacks and unchanged source bytes. Empty CLI output
   alone would not establish zero invocation.
6. Remove each changed guard's refusal append in disposable generated Go. Each
   mutant must compile, return exit 1 and match its independently frozen complete
   mismatch report at the named guest witness. Restore the original reducer and
   replay successfully. A separately frozen wrong outcome expectation must remain
   a fail-fast mismatch at `happy_pack`; neither original suite is modified.
7. Check original checkout bytes/modes and remove the temporary directory before
   publishing the receipt. It records six generation/replay runs, both successful
   orientations, four model versions, full-report/generated-Go hashes, refusal
   statuses and three mismatch controls. Hashes are consistency observations,
   not executing-source authentication or permission to continue.

The [frozen oracle](testdata/expected.json) SHA-256 is
`8ae4be53b0ee172073e6e36d9c21f3dbb0e6a39b50e8729d23a85e2669acac4e`.
Its source and behavior expectations are never refreshed by the workflow. If an
observation differs, investigate the discrepancy and obtain renewed independent
review before any substantive expectation correction.

## Regression checks

From `modeling/lifecycle` using that same toolchain and writable storage:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 ./mergerehearsal
```

The regression runs the real documented command twice and compares complete
receipts, checkout bytes/modes and temporary cleanup. Poisoned Python startup and
ambient Go flags test environment isolation. Host-side test doubles cover
short-circuiting, corrupted reports, subprocess launch/exit/timeout failures and a
cleanup error after actual removal that must publish no success output. These
controls are test code, not environment-selected product execution modes.

The receipt always says `example-only`, `authority=none`. A successful structural
merge, matched literals or cleanup does not approve requirements, authenticate
actors, acquire task ownership or authorize execution elsewhere. No deployment,
worker activation, source integration, standards conformance or visual review is
established. The existing producer task and evidence, original fixtures, portable
handoff and proposal-execution lanes remain separate. Exact validation and review
evidence belongs in the [task](../../../tasks/CF-merge-rehearsal-v1-concurrent-design-demo.md).
