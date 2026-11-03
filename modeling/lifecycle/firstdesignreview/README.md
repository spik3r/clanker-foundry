# Review an intake and proposal together

This check-only composition helps a reviewer correlate a first-design intake
with a proposal preflight. It calls the existing `intakemodel.Inspect` and
`proposalreplay.Check` APIs, then compares the intake's observed model and suite
texts with the corresponding texts in the proposal request. It retains both full
native results and the intake's unresolved questions. It does not run a scenario,
compile a candidate, make a decision about readiness, or grant execution or
production authority.

This is an isolated `lifecycle/v1` review path, not an application generator,
sandbox, identity or provenance service. A matching comparison, accepted model,
valid suite or `checked` result does not mean the requirements are right, an
independent author is authenticated, questions are answered, a human has
approved the design, or execution is authorized. Review the full reports and the
requirements. Separate any later candidate execution or deployment decision.
The CLI invokes no subprocesses, performs no network calls, and opens no files;
the caller's shell handles the input and output redirection below.

## Inputs and exact identity

The request is one `lifecycle-first-design-review-request/v1` JSON object with
exactly three required keys:

```json
{"formatVersion":"lifecycle-first-design-review-request/v1","intakeRequest":"complete native intake request JSON text","proposalRequest":"complete native proposal request JSON text"}
```

The last two values are UTF-8 **strings containing complete native requests**,
not paths, base64, or nested JSON objects. The intake request must already have
its observations; see [intake/model](../intakemodel/README.md). The proposal
request contains original model, suite and manifest texts; see
[proposal preflight](../proposalreplay/README.md). Both native contracts and
limits apply. This tool does not assemble observations, amend declarations,
rebind a suite, or calculate expected behavior.

After each native request passes its own strict request checks, the tool compares
the decoded `observed.model` and `observed.scenarioSuite` strings from the intake
with the decoded `model` and `suite` strings in the proposal request. Each result
is independently `matches`, `mismatch`, or `missing` (null intake observation).
Empty text is present, not missing. This is direct UTF-8 byte equality after JSON
string decoding: a newline or Unicode normalization difference is visible;
different JSON escape spellings that decode to the same bytes match. It is not a
comparison of canonical model digests or proof that either artifact is correct.

A valid intake report can still say model `missing` or `rejected`, and its suite
observation remains opaque. A valid stale proposal report can say suite
`not-checked` and replay `not-run`. Even two `matches` comparisons do not override
those native results. Unresolved questions remain in the preserved intake report.
No aggregate `ready`, `approved` or authorization status is produced.

## Check two prepared requests

Use Bash, Git, Python 3 and installed **Go 1.27.3** on a trusted local host. From the
`modeling/lifecycle` directory, set `INTAKE_REQUEST` and `PROPOSAL_REQUEST` to the
absolute paths of your independently reviewed complete native request files.
Set `WORK_PARENT` to an existing authorized writable directory outside the
checkout, with space for the Go cache and private output. Go operations use the
installed toolchain and offline module resolution. The two input files and all
embedded artifact bytes remain unchanged. Never publish private requests or
reports without the appropriate authorization.

```bash
set -euo pipefail
: "${INTAKE_REQUEST:?set the complete native intake request path}"
: "${PROPOSAL_REQUEST:?set the complete native proposal request path}"
: "${WORK_PARENT:?set an authorized output parent outside the checkout}"
export INTAKE_REQUEST PROPOSAL_REQUEST
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
ROOT="$(git rev-parse --show-toplevel)"
ROOT="$(cd -- "$ROOT" && pwd -P)"
WORK_PARENT="$(cd -- "$WORK_PARENT" && pwd -P)"
case "$WORK_PARENT/" in
  "$ROOT/"*) echo "Choose an output parent outside the checkout" >&2; exit 1 ;;
esac
umask 077
WORK="$(mktemp -d "$WORK_PARENT/first-design-review.XXXXXXXX")"
mkdir "$WORK/tmp" "$WORK/cache"
export TMPDIR="$WORK/tmp" GOTMPDIR="$WORK/tmp" GOCACHE="$WORK/cache"
go build -o "$WORK/modelreviewcheck" ./cmd/modelreviewcheck
python3 -I - <<'PY' > "$WORK/request.json"
import json, os, pathlib, sys
r = {'formatVersion': 'lifecycle-first-design-review-request/v1'}
for key, env in [('intakeRequest', 'INTAKE_REQUEST'), ('proposalRequest', 'PROPOSAL_REQUEST')]:
    r[key] = pathlib.Path(os.environ[env]).read_bytes().decode('utf-8')
sys.stdout.buffer.write(json.dumps(r, ensure_ascii=False).encode('utf-8'))
PY
"$WORK/modelreviewcheck" < "$WORK/request.json" > "$WORK/report.json"
printf 'Private report: %s\n' "$WORK/report.json"
```

The program's exit 0 means a report was produced, **not** that comparisons or
native bindings match. Inspect all fields, including the decoded
`intakeReport.intakeReportJSON` and `proposalReportJSON`; those strings are JSON
texts within the outer report. The source request hashes identify the exact decoded
native request strings. `intakeReport` preserves the complete typed result of
`intakemodel.Inspect`; its nested `intakeReportJSON` and hash retain the original
native intake bytes. The separate outer typed result is re-encoded, not asserted
byte-identical to the `modelintakecheck` CLI output. `proposalReportJSON` preserves
the exact native encoded check receipt; `proposalReportSHA256` hashes those decoded
bytes, including the final newline. Hashes are consistency observations, not
signatures or authentication.

The report always says `mode=example-only`, `authority=none`,
`executionStatus=not-run`, `reviewStatus=not-established`. A reviewer must
consider the model and suite comparisons, native binding and model statuses,
preflight suite status, retained questions and original requirements separately.
The proposal's nested runner binding is a validation-only context, not a
compiled candidate. The report gives no decision about whether to execute.

## Limits, failures and evidence

The wrapper is limited to 8 MiB and the encoded report to 32 MiB; nested request
and artifact limits may be narrower. Duplicate, unknown, case-variant, missing or
null required wrapper fields, invalid Unicode, malformed native requests, and
limit excess are refused. No partial report is returned by the in-memory API.
The CLI returns 2 for an invalid request or input/report limit, 4 for a native or
encoding boundary failure, and 3 for a failed or short report write. A short write
can expose a prefix; never treat a partial file as a report. A successful process exit does
not hide a native stale or rejected result.

The [frozen observation cases](testdata/README.md) were stated by a separate
assistant contributor before this composition was implemented. They cover the
invented first-design files, matching and mismatched source bytes, missing and
rejected observations, stale proposal bindings, unresolved questions and strict
input refusals. This is contributor separation, not independent-person review,
authenticated provenance, human approval or proof of arbitrary designs. On the
same trusted toolchain and native platform, repeat checks over the same bytes are
expected to agree; the native proposal receipt includes the host platform. No
visual review, executable reachability, integration or deployment is established.
