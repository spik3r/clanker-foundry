# Prepare and rehearse concurrent design changes

Use `foundry merge` to assemble the existing [custom merge rehearsal request](../modeling/lifecycle/custommergerehearsal/README.md)
from nine supplied files and an independently reviewed candidate model digest.
The same operator binary exposes native `check` and explicitly selected `execute`.
This is the bounded `lifecycle/v1` example, not a general application builder,
approval service, agent dispatcher or live branch merger.

Keep design drafting, independent literal expectations and review separate. The
[design handoff](../skills/model-driven-design/references/foundry-handoff.md)
explains their roles. Preparation preserves source bytes. It does not resolve
conflicts, choose a candidate, derive its expected digest, refresh bindings or
authorize execution. The existing [change entrypoint](model-driven-change-entrypoint.md)
handles one reviewed patch; this entrypoint handles base, left, right and merged
candidate designs through the native three-way merge contract.

## Prepare the supplied request

Use Bash, Git and installed Go **1.27.3** on a trusted, exclusive local checkout.
Choose an authorized output parent outside the checkout. This block builds the
trusted operator tool but does not compile or run a candidate. It uses no network
or dependency installation. Keep inputs and reports private when their content
requires it.

Set `MERGE_INPUT` to the directory holding the nine reviewed files below.
`merge-request.json` is the original native `lifecycle-model-merge-json/v1`
request containing the three model sources, not a previous merge report. Each
stage has its own independently authored suite and matching manifest. Supply
`EXPECTED_CANDIDATE_DIGEST` from a separate review of the intended candidate;
do not fill it automatically from the merge output or a rehearsal result.

```bash
set -euo pipefail
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
ROOT="$(git rev-parse --show-toplevel)"
ROOT="$(cd -- "$ROOT" && pwd -P)"
: "${WORK_PARENT:?set an authorized output parent outside the checkout}"
: "${MERGE_INPUT:?set the reviewed input directory}"
: "${EXPECTED_CANDIDATE_DIGEST:?supply the independently reviewed candidate digest}"
WORK_PARENT="$(cd -- "$WORK_PARENT" && pwd -P)"
MERGE_INPUT="$(cd -- "$MERGE_INPUT" && pwd -P)"
case "$WORK_PARENT/" in
  "$ROOT/"*) echo 'Choose an output parent outside the checkout' >&2; exit 1 ;;
esac
umask 077
MERGE_WORK="$(mktemp -d "$WORK_PARENT/foundry-merge.XXXXXXXX")"
mkdir "$MERGE_WORK/tmp" "$MERGE_WORK/cache"
export TMPDIR="$MERGE_WORK/tmp" GOTMPDIR="$MERGE_WORK/tmp" GOCACHE="$MERGE_WORK/cache"
cd "$ROOT/modeling/lifecycle"
go build -o "$MERGE_WORK/foundry" ./cmd/foundry
"$MERGE_WORK/foundry" merge prepare \
  --merge-request "$MERGE_INPUT/merge-request.json" \
  --base-suite "$MERGE_INPUT/base-suite.json" \
  --base-manifest "$MERGE_INPUT/base-manifest.json" \
  --left-suite "$MERGE_INPUT/left-suite.json" \
  --left-manifest "$MERGE_INPUT/left-manifest.json" \
  --right-suite "$MERGE_INPUT/right-suite.json" \
  --right-manifest "$MERGE_INPUT/right-manifest.json" \
  --candidate-suite "$MERGE_INPUT/candidate-suite.json" \
  --candidate-manifest "$MERGE_INPUT/candidate-manifest.json" \
  --expected-candidate-digest "$EXPECTED_CANDIDATE_DIGEST" \
  > "$MERGE_WORK/request.json"
printf 'Private request: %s\n' "$MERGE_WORK/request.json"
```

All ten flags are required exactly once, in any order, with separate values.
The digest must be 64 lowercase hexadecimal digits. File limits are 4 MiB for
the merge request, 1 MiB for each suite and 64 KiB for each manifest. The encoded
request limit is 32 MiB, including JSON escaping and the trailing newline.
Preparation reads regular UTF-8 files, rejects symlink leaves and nonregular
files, and checks that the opened file matches the inspected leaf. Parent path
aliases are allowed. These checks do not isolate a hostile concurrent writer.

The decoded request strings preserve every original file byte, including CRLF,
Unicode and whitespace. Preparation never rewrites files or reads stdin. Even
empty or non-JSON text can assemble: native checking remains a separate action.
Exit 0 means complete request delivery, exit 2 means argument/input/limit refusal,
and exit 3 means failed or short output. A failed delivery can leave a partial
output file. Diagnostics omit caller paths and payloads. Retain statuses as well
as files.

## Check without running a candidate

```bash
if "$MERGE_WORK/foundry" merge check < "$MERGE_WORK/request.json" \
    > "$MERGE_WORK/check.json"; then
  CHECK_STATUS=0
else
  CHECK_STATUS=$?
fi
printf 'Check exit status: %s\n' "$CHECK_STATUS"
```

Read the complete native merge report and all available stage receipts. A
`conflicted` or `invalid-combination` result has no candidate and performs no
proposal preflights or execution. Its expectation texts satisfy outer field
rules but are not claimed natively validated. Do not invent or select a candidate
to make the result pass.

For a merged design, the native check verifies the expected candidate identity
and all four independent proposal preflights. `checked` means those checks
completed with current bindings, not human design review or execution approval.
Each stage has `operation: check`. No compiler, subprocess or scenario execution
runs. A `stale` result retains all four check receipts and is not an executed
scenario mismatch. Reports retain `mode: example-only`, `authority: none` and
`reviewStatus: not-established`.

## Execute only after a separate authorization decision

Review the requirements, original files, independently supplied candidate digest,
native merge report, all preflight receipts and outstanding questions. Only when
the responsible operator separately authorizes this local rehearsal, run the
saved request unchanged from the trusted lifecycle module directory:

```bash
cd "$ROOT/modeling/lifecycle"
if "$MERGE_WORK/foundry" merge execute < "$MERGE_WORK/request.json" \
    > "$MERGE_WORK/execution.json"; then
  EXECUTION_STATUS=0
else
  EXECUTION_STATUS=$?
fi
printf 'Execution exit status: %s\n' "$EXECUTION_STATUS"
```

The native bridge repeats the merge and every preflight. All four must be current
before the first execution. It runs base, left, right, then candidate. Only a
matching native execution receipt, completed compilation, confirmed cleanup and
a live context permit the next stage. A mismatch, failure or cancellation stops
later stages, which retain their actual check receipts. The operator command
preserves native report bytes and status meanings:

- 0: checked without execution, or all four supplied suites matched after execution.
- 1: conflicted, invalid-combination, stale, base-mismatched, left-mismatched,
  right-mismatched or candidate-mismatched. Inspect the actual stages.
- 2: invalid arguments, input, limits or initial identity preconditions.
- 3: failed or short report delivery. Execution may have occurred; a partial
  output file is not a complete report.
- 4: failed outcome or execution/host boundary failure. Inspect native receipts when
  a report was delivered; absence of a report does not prove no execution occurred.

A stage's `operation: execute` means its executor was invoked, not necessarily
that a candidate process started. Inspect its native compilation, `candidateStarted`,
replay, candidate exit code, failure stage and cleanup fields. Earlier successful
stages remain recorded when a later stage fails. Expected refusals can be matching
cases. Hashes bind bytes, not provenance. A rehearsal match is not authenticated
review or authority to publish, integrate or deploy. Never update supplied literal
expectations from observed results to turn a mismatch into a pass.

The local compiler, generated code, source checkout and host remain trusted;
this is not a sandbox. Source root is the command's current working directory,
never request data. Native inventory and toolchain restrictions remain unchanged.
SIGINT/SIGTERM cancellation is cooperative on exact `merge execute`; it does not
prove rollback or descendant termination. Closed stdout preserves the native
delivery-failure status. Abrupt interruption can leave private scratch files.
Retain original inputs, reports and statuses for review, then remove scratch
outputs when the applicable retention policy permits.

The standalone `modelmergerehearsal check|execute` remains available with the
same native contract and shared CLI runner. Run `foundry help` to discover this
route alongside first-design, proposal, change and revision commands. The
[current workflow guide](model-driven-current-workflow.md) covers those earlier
steps; none silently selects merge execution.
