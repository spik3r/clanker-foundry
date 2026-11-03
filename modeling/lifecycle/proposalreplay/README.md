# Replay a reviewed first-model proposal

`modelproposal check|execute` connects the first-design handoff to an actual,
candidate-bound scenario runner. It accepts original model, suite and acceptance
manifest bytes, not a fixed demo model or an arbitrary program. It reuses existing
strict validation, acceptance preflight, qualification export/verification and
modelplay. There is no lifecycle interpreter or request-selected callback.

**Execute only when that local action is authorized.** Resolve the required human
choices and independently review the frozen expectations first. An idea, a request
to prepare review views, a valid report, or this command's presence supplies no
execution permission. The tool does not inspect a human's approval or an intake's
open questions. It never turns those external gates into a machine approval.
The original first-design worked handoff still has its unanswered real-use gate;
this command does not close it.

## Inputs and two separate actions

Build with installed Go **1.27.3** from `modeling/lifecycle`. Put the binary and
request/report files in a new authorized writable directory outside the checkout.
No toolchain or dependency installation is performed:

```sh
set -eu
: "${PROPOSAL_WORK:?set a new authorized private output directory}"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
go build -o "$PROPOSAL_WORK/modelproposal" ./cmd/modelproposal
```

Supply exactly these four required JSON keys:

```json
{"formatVersion":"lifecycle-proposal-request/v1","model":"original model JSON text","suite":"original suite JSON text","manifest":"original manifest JSON text"}
```

These are UTF-8 text strings, not nested objects, paths or base64. Unknown/duplicate
keys, invalid Unicode, missing/null fields and trailing JSON refuse. No request
field selects an executable, source root, source pins, environment or action.
The request is at most 16 MiB; decoded model and suite are each at most 1 MiB and
manifest at most 64 KiB. Existing strict model/suite/manifest limits still apply.

For an already reviewed proposal in `PROPOSAL_DIR`, preserve file bytes when
constructing a request (do not parse and reserialize the inner JSON):

```sh
set -eu
: "${PROPOSAL_DIR:?set the reviewed proposal directory}"
export PROPOSAL_DIR
python3 -I - <<'PY' > "$PROPOSAL_WORK/request.json"
import json, os, pathlib, sys
p = pathlib.Path(os.environ['PROPOSAL_DIR'])
r = {'formatVersion': 'lifecycle-proposal-request/v1'}
for role in ('model', 'suite', 'manifest'):
    r[role] = (p / (role + '.json')).read_bytes().decode('utf-8')
sys.stdout.buffer.write(json.dumps(r, ensure_ascii=False).encode('utf-8'))
PY
"$PROPOSAL_WORK/modelproposal" check < "$PROPOSAL_WORK/request.json" > "$PROPOSAL_WORK/check.json"
```

`check` is pure preflight: it launches no process, reads no runner source and makes
no temporary project. It works without a module working directory. The existing
acceptance checker uses a private panic-only callback as a **non-executing model
context**. Its nested `runnerBinding: current` is not a compiled-runner claim;
`validationContext` names that limitation. Check-only replay remains `not-run`.
A stale declaration retains its complete preflight, suite `not-checked` and no
candidate invocation. A malformed original suite is not repaired.

After separately obtaining execution permission, run the other action from the
trusted Foundry `modeling/lifecycle` directory:

```sh
set -eu
"$PROPOSAL_WORK/modelproposal" execute < "$PROPOSAL_WORK/request.json" > "$PROPOSAL_WORK/replay.json"
```

This action performs the same preflight before any process. It then reads only the
closed current [qualification source inventory](../qualification/README.md),
builds and verifies the existing candidate archive, and writes verified members
to a new private temporary project. It compiles the pinned modelplay with the
candidate's generated package, then invokes that binary on the unchanged suite.
Unlike the default checkout modelplay, this binary really is compiled against the
custom model. No model metadata is substituted to make a default runner accept it.
The generated structural test file is packaged, but this command does not run it;
only the separately frozen literal scenario suite is replayed.

The existing archive step still says suite unvalidated/unexecuted. The earlier
preflight and later actual replay are separate observations; that archive is not
rewritten as a replay pass. No new portable evidence format/verifier is added.

## Receipt and failure meanings

`lifecycle-proposal-replay/v1` records original input SHA-256 values, the complete
labeled preflight, source-policy/archive and binary hashes when reached, native
host platform, compilation/replay stages, candidate start/exit state, the complete
candidate report JSON value, cleanup state and an optional failure stage. Nested
JSON whitespace may be compacted by envelope encoding; original input hashes still
refer to the exact decoded input strings. Candidate report entries and their
literal expectations are not reduced to success counts.

| Exit | Meaning |
|---|---|
| 0 | Requested boundary completed: `checked` does not run; `matched` ran and matched the supplied literals, with cleanup observed. Neither approves the design. |
| 1 | Stale preflight (zero processes) or an actual `mismatched` replay. Read the distinct stage fields. |
| 2 | Invalid arguments/request or input read failure; no receipt and no candidate invocation by the command. |
| 3 | Failed/short receipt write; a prefix can escape. Execution may already have occurred. |
| 4 | Host, toolchain, project, compilation, replay-report or cleanup boundary failed. A receipt records reached stages when possible. |

`candidateStarted` records successful process start, not which internal step ran.
A timeout/protocol/output failure after that point is `failed-after-start`, never
`not-run` or rollback. A failed start is separate. A matching candidate result is
retained if later cleanup fails, but outcome and exit are failure. Empty output or
an error alone does not prove zero execution. Diagnostics do not echo proposal or
subprocess text; receipts themselves can contain sensitive task/command data.

The bridge checks the pinned child's report envelope, source/model/suite identities,
exit agreement and ordered case/step prefix. The existing compiled modelplay owns
full result comparison and report consistency. This is not a general verifier for
reports supplied by other programs, nor authentication of execution or source.

## Trusted-host boundary

The source checkout, installed compiler, operating system and generated code are
trusted. This is local orchestration, **not a sandbox**. Fixed Go subprocesses use
180-second timeouts and two-second pipe WaitDelay; output is bounded, with candidate
stdout capped at 8 MiB and the receipt at 9 MiB. The generated executable is bounded
to 64 MiB for hashing. Static symlink/regular-file and source-pin checks reject
accidental substitution, not a hostile concurrent writer. Hashing a binary before
invoking its path is not atomic source authentication against such a writer.

Only the explicit execute path creates a private project. It refuses temporary or
Go-cache destinations inside the source module/nearest Git checkout, resolving
existing symlinks first. The Go cache may be reused outside that checkout; temporary
compiler files and cached build output are additional host writes. `TMPDIR`,
`GOCACHE` (or the host's user cache), `HOME`, `PATH`, `LANG` and `LC_ALL` are local
host settings, never request fields. Other ambient Go/tool overrides are dropped.
Go workspace/configuration mode and CGo are disabled; resolution stays offline,
VCS stamping is explicitly off for the manifest-identified disposable project,
and trimpath removes temporary source paths. Binary identity/repeatability is
scoped to the same trusted toolchain/native platform, not all hosts.

Cleanup is attempted and checked before success. Abrupt termination can leave
files. Context cancellation and WaitDelay do not guarantee descendant termination,
OS resource containment or rollback. This command activates no workers, performs no
network operation itself and grants no publication, integration or deployment
permission. Human meaning review, real identity, visual review and surrounding
application behavior remain separate.

## Independent example and regression evidence

The two [frozen fixture sets](testdata/README.md) were authored by a separate
assistant before bridge logic. They have different non-default policies and full
literal reports, including owner/null distinctions, ordered criteria, exact large
revisions, selected-transition refusals and invalid-input retention. The author's
executor was unavailable; the implementer mechanically materialized the author's
frozen base64 bytes and verified all source/Git hashes. That is not an independent
runtime rerun by the author.

From the module root with the same toolchain:

```sh
go test -race -count=1 ./proposalreplay ./cmd/modelproposal
```

Tests exercise the real command and candidate compilation/replay, check complete
literal reports and repeatability, and compare source bytes/modes before and after.
Compiling disposable reducer mutants must fail the unchanged behavior oracle.
Separate fault-injection tests cover zero-launch preflight, static source/path
refusals, wrong toolchain, process/output failures, cancellation and cleanup.
Those fakes establish failure reporting, not actual candidate behavior. No frozen
expectation is refreshed from a run and no visual inspection is claimed here.
