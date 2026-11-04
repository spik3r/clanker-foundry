# Rehearse caller-supplied concurrent designs

This standalone bridge uses the native [merger](../modelmerge/README.md) and
[proposal executor](../proposalreplay/README.md) without changing their contracts.
It derives models from one native merge request and rehearses four separately
supplied expectation suites. It grants no approval or execution authority.

## Request and byte identities

Check and Execute accept strict lifecycle-custom-merge-rehearsal-request/v1 JSON:

- formatVersion: the version above.
- mergeRequest: original native lifecycle-model-merge-json/v1 request text.
- base, left, right, candidate: objects containing exactly suite and manifest,
  each an original nonempty UTF-8 source text string, not a path.
- expectedCandidateDigest: a separately reviewed lowercase SHA256 model digest.

The native merger validates mergeRequest before this bridge extracts the exact
base, left and right JSON tokens. Only the native result supplies the candidate
token. The bridge wraps these original tokens and unchanged suite/manifest texts
in native proposal requests. Each native preflight's model identity must match
its native merge digest. The caller's candidate assertion must match the native
candidate digest; never copy an observed digest into an expectation to repair a
failure. Byte hashes and canonical model digests are different identities.

No field selects a process, callback, source root or filesystem path. Check has
no executor capability, reads no filesystem and launches no process or replay.
Native merge conflicts and invalid combinations retain the complete merge report
and have no candidate or proposal preflights. Nonempty suite/manifest strings on
those paths satisfy outer framing and size checks only, not native JSON validity.

Outer input is bounded at 32 MiB, merge input at 4 MiB and native merge output at
2 MiB. Each model and suite is at most 1 MiB; each manifest is at most 64 KiB.
Generated proposal requests are bounded at 16 MiB and native receipts at 9 MiB.
Final reports are bounded at 128 MiB. Before execution, preparation reserves the
maximum four-receipt envelope so normal native maxima cannot overflow it later.
Native strict parsing, Unicode, depth and semantic limits still apply.

## Check, then explicitly authorize execution

Use installed Go 1.27.3, offline, from a trusted exclusive checkout. Keep its actual
binary directory on PATH because native children resolve go there. Choose a new
private absolute output directory outside the checkout, and preserve the reviewed
request's original texts when creating request.json. No toolchain is downloaded.

```sh
set -eu
: "${MERGE_WORK:?set a new private absolute output directory}"
mkdir -m 700 "$MERGE_WORK"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
# Run from modeling/lifecycle; place the reviewed request outside the checkout.
go build -o "$MERGE_WORK/modelmergerehearsal" ./cmd/modelmergerehearsal
"$MERGE_WORK/modelmergerehearsal" check < "$MERGE_WORK/request.json" > "$MERGE_WORK/check.json"
# Inspect original inputs and native reports, then separately authorize execution.
"$MERGE_WORK/modelmergerehearsal" execute < "$MERGE_WORK/request.json" > "$MERGE_WORK/execute.json"
```

The Go APIs are Check(raw []byte) ([]byte,error) and Execute(ctx context.Context,
raw []byte, sourceRoot string) ([]byte,error). sourceRoot is trusted host configuration.
The command uses its working directory. Public APIs accept no executor callback.
Keep caller input immutable during either call. No foundry route is added.

Execute recomputes the merge and all four preflights. Every preflight must be
current before any executor call. Run order is base, left, right, candidate through
native proposalreplay.Execute only. A matched receipt with completed compilation,
actual candidate start, matching replay/exit code, removed cleanup and no failure
permits the next stage only while context remains live. Mismatch, error, uncertain
cleanup or cancellation suppresses later stages. Context is checked after the last
stage too: cancellation can return failed while retaining a completed native result.
The native trusted-host, source selection, toolchain and cleanup limits still apply.

## Reports and failure meanings

Reports use lifecycle-custom-merge-rehearsal-report/v1 with mode example-only,
authority none and reviewStatus not-established. They retain original merge request
and native merge report text plus SHA256, original outer request SHA256, the candidate
assertion, and named base/left/right/candidate stages. Each stage records original
model/suite/manifest hashes, generated proposal request hash, operation, and exact
encoded native receipt text/hash.

On a non-merged result, operation is not-attempted; proposal and receipt fields are
null, as is the nonexistent candidate model hash. On a merged result each stage
starts with its actual check receipt. Only calling that stage's native executor
changes operation to execute and replaces the receipt. A suppressed stage retains
its check receipt. An execute operation is not proof the native candidate started:
inspect candidateStarted, compilation, replay, failureStage and cleanup inside it.

CLI exits:

- 0: checked or matched.
- 1: conflicted, invalid-combination, stale, or base/left/right/candidate-mismatched.
- 2: initial malformed, resource-limit or identity refusal; no report.
- 3: short/failed stdout delivery, possibly after execution.
- 4: native boundary failure, normally with a failed report and honest partial receipts.

If an unexpected native receipt cannot be encoded, there may be no report despite
execution having occurred. Keep stdout, stderr and exit status. SIGPIPE is ignored
so a closed stdout returns 3. Cancellation or removed cleanup does not prove rollback,
external-effect reversal or descendant termination. Hashes authenticate neither host
nor intent; a matching rehearsal does not establish human approval or readiness.

## Independent invented acceptance

ArchivePacket has independent actor and outcome guards. Base requires nonempty
values, left requires actor Clerk, right requires outcome indexed, and the merged
model requires both. Four literal combinations witness both edits and their joint
refusals. A separate contributor authored model/suite/manifest bytes and candidate
identity before production. Original frozen behavior remains unchanged. Two native
receipt-label assertions were independently corrected after initial source edits but
before bridge tests; they now distinguish top-level not-requested from nested not-run.

Acceptance covers both orientations, native conflicts/invalid combinations, all
four stale/mismatch stages, exact source/receipt identities and partial results.
Synthetic executor controls prove sequence and suppression only; actual native
runs provide the behavioral observations. Compiled-command tests cover output and
closed pipes. These finite cases are not a proof for all designs or host failures.
