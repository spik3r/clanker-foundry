# Prepare and rehearse a reviewed design change

Use `foundry change` to assemble the existing [change rehearsal request](../modeling/lifecycle/changerehearsal/README.md)
from four supplied files and a separately reviewed candidate model digest. The
same binary exposes native `check` and explicitly selected `execute`. This is the
bounded `lifecycle/v1` example, not a general application builder or an approval
service. It does not dispatch agents or implement BPMN, SysML, UTP or fUML.

Keep requirements, model drafting and independent literal expectation review
separate. The [model-driven design handoff](../skills/model-driven-design/references/foundry-handoff.md)
explains that distinction. Preparation preserves the inputs; it does not resolve
questions, validate their meaning, refresh bindings or authorize execution.

## Prepare the unchanged native request

Use Bash, Git and installed Go **1.27.3** on a trusted, exclusive local checkout.
Choose an authorized output parent outside the checkout. The following block
builds the trusted operator tool but does not compile or run a candidate. It
uses no network or dependency installation. Keep supplied inputs and reports
private when their content requires it.

Set `CHANGE_INPUT` to the directory containing the four reviewed files listed
below. `baseline-proposal.json` is a native proposal request, not a previous
receipt. `patch-request.json` is the original native patch request, not a planner
report. Supply `EXPECTED_CANDIDATE_DIGEST` from the separate review of the intended
candidate; do not fill it automatically from a new planner result.

```bash
set -euo pipefail
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
ROOT="$(git rev-parse --show-toplevel)"
ROOT="$(cd -- "$ROOT" && pwd -P)"
: "${WORK_PARENT:?set an authorized output parent outside the checkout}"
: "${CHANGE_INPUT:?set the reviewed input directory}"
: "${EXPECTED_CANDIDATE_DIGEST:?supply the separately reviewed candidate digest}"
WORK_PARENT="$(cd -- "$WORK_PARENT" && pwd -P)"
CHANGE_INPUT="$(cd -- "$CHANGE_INPUT" && pwd -P)"
case "$WORK_PARENT/" in
  "$ROOT/"*) echo 'Choose an output parent outside the checkout' >&2; exit 1 ;;
esac
umask 077
CHANGE_WORK="$(mktemp -d "$WORK_PARENT/foundry-change.XXXXXXXX")"
mkdir "$CHANGE_WORK/tmp" "$CHANGE_WORK/cache"
export TMPDIR="$CHANGE_WORK/tmp" GOTMPDIR="$CHANGE_WORK/tmp" GOCACHE="$CHANGE_WORK/cache"
cd "$ROOT/modeling/lifecycle"
go build -o "$CHANGE_WORK/foundry" ./cmd/foundry
"$CHANGE_WORK/foundry" change prepare \
  --patch-request "$CHANGE_INPUT/patch-request.json" \
  --baseline-proposal "$CHANGE_INPUT/baseline-proposal.json" \
  --candidate-suite "$CHANGE_INPUT/candidate-suite.json" \
  --candidate-manifest "$CHANGE_INPUT/candidate-manifest.json" \
  --expected-candidate-digest "$EXPECTED_CANDIDATE_DIGEST" \
  > "$CHANGE_WORK/request.json"
printf 'Private request: %s\n' "$CHANGE_WORK/request.json"
```

All five flags are required exactly once, in any order. The digest must contain
64 lowercase hexadecimal digits. The file limits are 3 MiB for the patch request,
16 MiB for the baseline proposal, 1 MiB for the candidate suite and 64 KiB for its
manifest. The encoded request is limited to 32 MiB, including JSON escaping and
its trailing newline. Preparation reads regular UTF-8 files, rejects symlink
leaves and nonregular files, and checks that the opened file matches the inspected
leaf. Parent path aliases are allowed. These checks do not isolate a hostile
concurrent writer.

The four decoded request strings retain the exact input bytes, including CRLF,
Unicode and whitespace. Files are never rewritten. Even non-JSON artifact text
can be assembled: the separate native check rejects malformed or inconsistent
artifacts. Exit 0 means a complete request was delivered, exit 2 means arguments,
inputs or bounds were refused, and exit 3 means output delivery failed or was
short. A failed delivery can leave a partial output file. Diagnostics do not echo
private paths or payloads. Retain the status as well as the file.

## Check without executing a candidate

```bash
if "$CHANGE_WORK/foundry" change check < "$CHANGE_WORK/request.json" \
    > "$CHANGE_WORK/check.json"; then
  CHECK_STATUS=0
else
  CHECK_STATUS=$?
fi
printf 'Check exit status: %s\n' "$CHECK_STATUS"
```

Read the complete native patch report and both proposal preflight receipts.
`checked` means identity and preflight checks completed, not design approval.
The report retains `mode: example-only`, `authority: none` and
`reviewStatus: not-established`. Both sides have `operation: check`; no compiler,
child process or candidate replay is invoked. A stale result is not a failed
executed scenario, and an input refusal is not permission to repair expectations.

## Execute only after a separate authorization decision

Review the requirements, original inputs, expected candidate digest, full check
report and outstanding questions. Only when the responsible operator has
separately authorized this local rehearsal, run the saved request unchanged from
the trusted lifecycle module directory:

```bash
cd "$ROOT/modeling/lifecycle"
if "$CHANGE_WORK/foundry" change execute < "$CHANGE_WORK/request.json" \
    > "$CHANGE_WORK/execution.json"; then
  EXECUTION_STATUS=0
else
  EXECUTION_STATUS=$?
fi
printf 'Execution exit status: %s\n' "$EXECUTION_STATUS"
```

The native bridge recomputes the patch and both preflights. Both must be current
before any execution. It runs the baseline first; only a matching baseline,
confirmed cleanup and live context permit the candidate. A baseline mismatch,
failure or cancellation leaves the candidate's original check receipt. The
operator command preserves native report bytes and exit meanings, including
partial receipts after boundary failures:

- 0: checked without execution, or both supplied suites matched after execution.
- 1: stale preflight, baseline mismatch or candidate mismatch. Inspect the stages.
- 2: invalid arguments, input, limits or initial identity preconditions.
- 3: failed or short report delivery. Execution may already have occurred.
- 4: execution or host boundary failure. Read the retained native receipts.

A side's `operation: execute` says its executor was called. Inspect its native
`candidateStarted`, compilation, replay, candidate exit code, failure stage and
cleanup fields before describing what actually ran. Expected refusals can be
matching cases. A match does not authenticate review or authorize integration.
Never refresh the supplied literals from observed results to turn a mismatch
into a pass.

The local compiler, generated code, source checkout and host remain trusted;
this is not a sandbox. The source root is the command's current working directory,
not request data. Native source inventory and toolchain restrictions still apply.
SIGINT/SIGTERM cancellation is cooperative on explicit change execution; it does
not prove descendant termination or rollback. Closed stdout produces the native
output-failure status. Abrupt interruption can leave private scratch files.
Retain original inputs, reports and command statuses for review, then remove
scratch outputs when the applicable retention policy permits.

The standalone `modelchangerehearsal check|execute` remains available with the
same native contract. The [current workflow guide](model-driven-current-workflow.md)
covers first-design preparation and revision comparison, which remain separate
from this change execution decision.
