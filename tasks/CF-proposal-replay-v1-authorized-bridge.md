---
id: CF-proposal-replay-v1
status: review
owner: Byte proposal replay contributor
repositories: [spik3r/clanker-foundry]
---

# Connect a first-model proposal to its actual candidate replay

## Implementation claim

The coordinator assigned this bounded local lane on 2026-11-02 after reviewing
actual main `28739a21b126dfc0a793622ce46593ec7559d379`, tree
`10dc22dbfcd2dbe09d41814a01a2696886dc0cc7`. The first-design skill ends at review
preparation; the separate rehearsal executes only a fixed design change. A custom
first proposal still requires manual recompilation to reach its own reducer.
Current remote main/open PRs and all 53 local Project Memory frontiers were
checked; only excluded PR20 was open and all frontiers were complete. The separate
rehearsal contributor was informed; its retained review-bundle work is distinct.
This is a prospective local claim, not an atomic shared reservation.

Branch `codex/proposal-replay-v1` starts at that canonical main. Byte owns only:
- `modeling/lifecycle/proposalreplay/README.md`
- `modeling/lifecycle/proposalreplay/replay.go`
- `modeling/lifecycle/proposalreplay/host.go`
- `modeling/lifecycle/proposalreplay/replay_test.go`
- `modeling/lifecycle/proposalreplay/host_test.go`
- `modeling/lifecycle/proposalreplay/oracle_test.go`
- `modeling/lifecycle/proposalreplay/testdata/model-a.json`
- `modeling/lifecycle/proposalreplay/testdata/suite-a.json`
- `modeling/lifecycle/proposalreplay/testdata/manifest-a.json`
- `modeling/lifecycle/proposalreplay/testdata/model-b.json`
- `modeling/lifecycle/proposalreplay/testdata/suite-b.json`
- `modeling/lifecycle/proposalreplay/testdata/manifest-b.json`
- `modeling/lifecycle/proposalreplay/testdata/expected.json`
- `modeling/lifecycle/proposalreplay/testdata/README.md`
- `modeling/lifecycle/cmd/modelproposal/main.go`
- `modeling/lifecycle/cmd/modelproposal/main_test.go`
- `skills/model-driven-design/references/foundry-handoff.md` (later authorized
  execution route only; do not broaden review-preparation authority)
- this task and its own new immutable Project Memory files.

An independent fixture author owns testdata and oracle_test.go and freezes its
literal expectations before orchestration implementation. Do not edit those files
without the author's review. Exclude every existing core/compiler/runner/codec,
qualification source policy, fixture, rehearsal, skill other than the named narrow
reference, workflow, dependency, runtime, UML work, denied PR20 artifact and PR5.
Do not access denied artifacts. No deployment or private consumer adoption.
Publication and integration remain separate coordinator operations.

## Contract

Provide one Go-first local bridge, `modelproposal check|execute`, wired into the
existing first-design reference. The command consumes a bounded strict JSON object
with required `formatVersion: lifecycle-proposal-request/v1`, and `model`, `suite`,
`manifest` strings containing original UTF-8 file bytes. No input path, executable,
source pin table, callback, environment variable, mode or arbitrary program can be
selected through the request. An idea or review-preparation request is not authority
to execute. `execute` is an explicit operator action after the needed decisions,
independent expectation review and execution permission; the program does not
establish that those human conditions are true.

Use existing strict model/canonical and acceptance/scenario APIs for preflight.
A private panic-only scenario callback provides a model-matched validation context
only, never executable composition or authenticated runner identity. Check mode
must create no temporary project, read no pinned runner source and invoke no process.
Stale bindings preserve their complete check-only report and invoke no process in
either mode. Malformed input is rejected without a replay report. Never repair
manifest/suite/model bytes or derive expectations from actual behavior.

For execute after successful preflight, read only the fixed current qualification
source inventory from the trusted module root, with bounded regular-file/static
symlink checks. Use existing qualification Build/Verify and artifactbundle APIs.
Only verified, closed-inventory members enter a new private temporary project;
never mutate or regenerate the caller's checkout. The archive remains an internal
consistency step, not a new portable evidence format or proof of execution.

Compile the existing pinned modelplay against the generated candidate using
installed exact Go1.27.3, offline resolution, disabled ambient Go workspace/config,
no VCS stamping and no CGo. No toolchain or dependency installation/download.
The actual binary is candidate-bound by its compiled generated package, not request
metadata or the default checkout reducer. Invoke it on the unchanged suite. Preserve
its full emitted report and mismatch outcome, separately from preflight and archive
consistency. The bridge never interprets lifecycle guards itself. An actual replay
is not model correctness, provenance authentication, human review or approval.

Return a bounded deterministic JSON receipt with original input hashes, semantic
model/scenario identities, the separately labeled validation-only preflight, actual
source-policy/archive and compiled-binary identities when reached, compilation and
replay stage states, subprocess exit/report, and cleanup state. No timestamps,
temporary paths or subprocess diagnostics enter successful receipts. Output must
distinguish no invocation from a started invocation that timed out/failed; missing
output does not mean not-run. Compilation/replay/cleanup failure cannot report
success. A matching replay remains recorded as an observation if later cleanup
fails, with overall failure. Failed/short stdout can expose a prefix and is an
error, never a successful delivery.

Bound stdin, per-file reads, captured output and subprocess duration. Use a trusted
single-writer host: static file checks are not hostile-writer resistance; the
compiler and generated code are trusted. Context cancellation/WaitDelay and removal
attempts are not OS sandboxing, descendant-kill, callback rollback or abrupt-kill
cleanup guarantees. Report cleanup failure; do not hide it behind a pass. No network,
worker activation, authority grant, arbitrary-process interface or deployment.

## Acceptance checks

Freeze two non-default model/suite/manifest sets and complete expected candidate
reports independently before implementation. Check model/source/suite identities
without executing a reducer first. Then exercise the actual documented command on
both custom models, preserving input and checkout bytes. Compare complete actual
reports with the frozen literals. Exercise stale/malformed inputs with a process
spy proving zero launches, a changed expectation, and compiling reducer mutations
caught by behavioral assertions rather than compile failure. Cover bounded input,
strict JSON, source tampering/static symlinks, wrong toolchain, process/output
failure, cancellation, cleanup and broken output without success claims.

Run focused and whole-module Go1.27.3 offline race/vet/modelc checks, repository
validation and committed-diff lifecycle validation. Obtain independent exact-source
and invocation-boundary review. Record failures/skips honestly; installation is not
part of this Go bridge and existing installer evidence is separate. Keep all prior
fixture bytes and old lifecycle files untouched.

## Local result and evidence

The bridge and later-authorized first-design reference route are implemented.
Source review is pinned to `48848bb51d6b1e780950b46944776ebb98567d8b`, tree
`ce295704592b7306f7b23945ebac4dc2b612d37f`, on the canonical main above. The
subsequent completion changes only this task and its append-only event. No shared
compiler, runner, source policy, baseline fixture, dependency or workflow changed.

### Independently frozen expectations

- A separate author selected ParcelSeal and CuratorRelease policies and their
  eight cases / thirteen complete literal steps from published baseline contracts.
  It froze eight exact payloads and hashes in the review transcript before bridge
  logic existed. Its native executor was unavailable. The implementer mechanically
  decoded those payloads, verified every byte count, SHA-256 and Git blob identity,
  and committed only those files in `0f9d553bcd4fd48dc4a76e30897064887d0e1ea6`.
  Transcript freeze precedes implementation; later materialization is not claimed
  as an earlier Git commit or as an independent runtime rerun by that author.
- The complete oracle remains SHA-256
  `cdc873a15122f8ccd1bd020cf7d193f101d88e5360e3a873a1bec0317cf57804`.
  Strict decode/digest and acceptance check-only self-checks passed without invoking
  a reducer. The optional `oracle_test.go` was not authored; the checkpoint removed
  that absent path from the planned scope. All eight fixture files remain unchanged.
- Actual candidate A and B report bytes match the independently frozen full reports:
  `d4a22fddf07c543137912226728d3e1e2b3124769ea0fb223d02647695055095` and
  `7707b41b23fb536ab5124a7bb058f676ff2028ce05b5a8dbabc2c328626db3ee`.
  Both actual CLI executions compiled, started, exited zero and observed cleanup.
  Repeating A produced an identical receipt on the same host/toolchain. Source
  bytes/modes and original input hashes were unchanged.
- Two disposable generated reducers compiled but failed the unchanged behavior
  oracle: removing A owner assignment first differed on owner Mira versus null at
  step A_assign; ignoring B's curator guard first omitted G_curator from refusals
  at B_three_guards. Both exited one with cleanup, not compilation failure. The
  independent author predicted these exact witnesses before seeing their receipts.
  A separate negative suite changes A's first expected revision from 8 to 9 and
  updates only its own manifest hash; replay retains expected 9 versus actual 8,
  rather than repairing the expectation. Frozen files were not changed.

### Checks and independent review

- Installed Go 1.27.3, offline bounded serial
  `go test -p 1 -race -count=1 ./...`: **60 tested packages passed**.
  Whole-module `go vet -p 1 ./...`, `modelc validate`, `modelc check`,
  `modeltrace check`, focused normal tests, gofmt and whitespace checks passed.
- Repository validation passed. Its optional `skills-ref` executable was absent
  and skipped. Installation was not rerun: no local GNU Stow installation or
  installer pass is claimed by this change; previous installer evidence is separate.
- A different native-capable reviewer independently reviewed the exact source and
  reran focused race/vet and both actual CLI proposals with no findings. It verified
  all frozen hashes, complete reports, unchanged source/input bytes, fixed version
  and build commands, sanitized offline environment, the exact 18-member private
  project, 0700 directory permissions and removal. Poisoned-go/non-module/missing-temp
  controls confirmed check and stale execute do not launch a process. It separately
  verified full mutant receipts. The whole-module aggregate remains implementer
  evidence, not a claim that this reviewer reran it.
- The fixture author separately reviewed the comparator excerpts: hard-coded
  oracle/provenance hashes, exact raw report hashes and full JSON equality consume
  its literal expectations without refreshing them from execution. That static
  assessment is distinct from the other reviewer's native execution evidence.

An early real output-limit regression failed because embedding `bytes.Buffer`
promoted `ReadFrom`, allowing `io.Copy` to bypass the bounded `Write`. The retained
test now passes after changing to a named buffer. An initial modelc check used an
unsupported flag; the documented subcommand retry passed. Neither failed attempt
is reported as a pass. No expectation was changed to accommodate a runtime result.

This completes the local bridge, not human design approval, visual review, the
worked first-design example's open use decision, publication, integration or
deployment. Current main and exact lifecycle coverage must be checked again for
any later publication/integration operation. No denied artifact was accessed.
