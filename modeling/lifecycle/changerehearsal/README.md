# Rehearse a supplied design change

This bridge checks a native model patch against a baseline proposal, derives the
candidate model, and rehearses separately supplied before/after expectations.
It uses the existing [patch planner](../modelpatch/README.md) and
[proposal boundary](../proposalreplay/README.md) without changing either contract.
It supports invented local examples; it grants no execution or integration authority.

## Input and identities

`Check([]byte)` and `Execute(context.Context, []byte, sourceRoot)` take the same
strict JSON request. The trusted host supplies `sourceRoot`, never request data.
All six fields are required:

- `formatVersion`: `lifecycle-change-rehearsal-request/v1`.
- `patchRequest`: the original native patch-request JSON text.
- `baselineProposalRequest`: the original native proposal-request JSON text.
- `candidateSuite`: independently reviewed candidate suite text.
- `candidateManifest`: the manifest bound to that suite and candidate model.
- `expectedCandidateDigest`: a separately reviewed lowercase SHA256 model digest.

The patch planner supplies the exact candidate `after` JSON token. Its before
canonical digest must equal the baseline model's canonical digest; its after
digest must equal the supplied candidate assertion. The bridge does not rewrite
sources, refresh manifests, copy expectations, or infer human approval. Whitespace
may differ between canonically equal baseline models. Source hashes still bind
the supplied original bytes. `candidateModelSHA256` binds the exact derived token,
which differs from the canonical model digest.

The outer request limit is 32 MiB; the native patch request limit is 3 MiB,
baseline proposal 16 MiB, candidate suite 1 MiB, and candidate manifest 64 KiB.
Native parsing, depth and semantic limits still apply. Output is bounded at 64 MiB.
Unknown fields, duplicate keys, malformed UTF8/JSON and identity conflicts refuse
before any execution. No request field selects a path, process or source root.

## Check, then separately authorize execution

From a trusted, exclusive checkout's `modeling/lifecycle` directory, choose a
new private absolute directory **outside** the checkout for all artifacts. Supply
its name yourself; `mkdir` must refuse if it already exists. Place the reviewed
request at `$CHANGE_WORK/request.json` before checking it. Use the trusted local
Go 1.27.3 toolchain; keep its actual binary directory on PATH for native children.

```bash
set -eu
export CHANGE_WORK=/absolute/path/to/new-private-change-work
mkdir -m 700 "$CHANGE_WORK"
test "$(GOTOOLCHAIN=local go env GOVERSION)" = go1.27.3
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go build \
  -o "$CHANGE_WORK/modelchangerehearsal" ./cmd/modelchangerehearsal
# Place the independently reviewed request at "$CHANGE_WORK/request.json".
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off "$CHANGE_WORK/modelchangerehearsal" check \
  < "$CHANGE_WORK/request.json" > "$CHANGE_WORK/checked.json"
# Review the original inputs and full patch/preflight reports first.
# Only when this local rehearsal is authorized:
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off "$CHANGE_WORK/modelchangerehearsal" execute \
  < "$CHANGE_WORK/request.json" > "$CHANGE_WORK/execution.json"
```

`check` invokes no compiler, child process or scenario replay. `execute` recomputes
the patch and both native preflights from the same unchanged request. Both must be
current before it invokes either native executor. A stale result executes neither.
The baseline runs first. Only a matched baseline with successful native cleanup
and a live context permits the candidate run. Baseline mismatch, boundary failure,
cleanup uncertainty or cancellation suppresses the candidate. The source-root,
Go toolchain, source-pin and host restrictions are the native proposal contract's.

The report uses `lifecycle-change-rehearsal-report/v1`, `mode: example-only`,
`authority: none` and `reviewStatus: not-established`. It retains the exact native
patch report and each side's full receipt text with SHA256, plus original request,
patch request, proposal request and candidate-token identities. Each side starts
with `operation: check`; it changes to `execute` only after that native executor
is called. An execute operation does not prove the native candidate started:
inspect its `candidateStarted`, failure stage, replay and cleanup fields. Native
partial receipts remain present after boundary errors. These hashes establish
byte relationships, not provenance or authenticated execution evidence.

Outcomes are `checked`, `stale`, `matched`, `baseline-mismatched`,
`candidate-mismatched` or `failed`. The CLI exits 0 for checked/matched, 1 for stale
or either mismatch, 2 for an initial input/identity refusal, 3 for failed report
delivery, and 4 for a boundary failure. Boundary failure can include a report on
stdout. A delivery failure may follow completed execution. Cancellation and
cleanup records do not prove process rollback or reversal of external effects.
Retain full stdout, stderr and exit status rather than treating file presence as
success. SIGPIPE is ignored so a closed stdout produces exit 3 in either mode.

The frozen invented ParcelSeal example changes an actor guard from nonempty to
exactly `Courier`: Mira succeeds before and is refused afterward, while Courier
succeeds on both sides. Independent literals were committed before this bridge.
Tests also retain actual baseline/candidate mismatches, suppressed stages, native
partial errors, cancellation, concurrency and actual binary delivery behavior.
These finite examples do not establish coverage of every design or host failure.
