# Use the current model-driven review path

The native contracts below were checked against reviewed main
[`2e2a190b`](https://github.com/spik3r/clanker-foundry/commit/2e2a190b629a983ba854d59ecc356274ec7e5a4f)
on **4 November 2026**. The `foundry` entrypoint in this checkout prepares requests
and exposes those contracts through one binary. This is a selected workflow.
The [pipeline introduction](model-driven-pipeline.md) retains its earlier dated
snapshot. The [pilot testing guide](model-driven-pilot-testing.md) provides longer
fixed-example and custom-proposal exercises.

The supported design language is the finite `lifecycle/v1` profile: states,
transitions, typed guards and ordered effects over fixed task/command data.
Use the [model-driven-design handoff](../skills/model-driven-design/references/foundry-handoff.md)
to assess fit, clarify requirements and keep model drafting separate from literal
expectation authorship. People still resolve questions, review meaning and decide
whether to execute or integrate. These tools do not generate a complete app.

| Step | Tool and contract | What to review |
|---|---|---|
| Prepare a proposal | `foundry prepare` plus the [design skill](../skills/model-driven-design/SKILL.md), [profile catalog](../modeling/lifecycle/profilecatalog/README.md), [acceptance authoring](../modeling/lifecycle/acceptanceauthor/README.md) | Requirements, model, independently stated suite literals, manifest and open questions. Author labels do not authenticate independence. |
| Correlate intake and proposal | [`foundry review`](../modeling/lifecycle/firstdesignreview/README.md) | Full native results, exact model/suite byte comparisons and retained questions. Exit 0 only means a report was produced. This check starts no candidate. |
| Check or execute a proposal | [`foundry proposal check\|execute`](../modeling/lifecycle/proposalreplay/README.md) | Check is non-executing preflight. Separately authorized execute compiles and invokes a candidate-bound runner and preserves its comparisons. |
| Inspect a process view | [`foundry process json\|svg`](../modeling/lifecycle/processview/README.md) | Declared guards, refusal paths, effect order, destination and canonical model identity. A rendered view does not establish execution or approval. |
| Inspect carried observations | [`foundry inspect`](../modeling/lifecycle/rehearsalinspection/README.md) | Native report consistency and binding to carried suite literals. Invented consistent reports can pass; execution remains not established. |

The process view is a review aid, with invalid input and illegal transitions
outside its per-transition lanes. It makes no BPMN conformance claim. The
[runtime trace](../runtime/model-trace-v1/README.md) is a separate partial map of
the existing orchestration evaluator; this design path does not dispatch workers.

## Prepare private outputs and check a first design

Use Bash, Python 3, Git and installed **Go 1.27.3** on a trusted, exclusive local
checkout containing the `foundry` entrypoint; the linked revision identifies the
native contract baseline. Set `WORK_PARENT` to an authorized directory
outside that checkout. The examples compile trusted tools and write only new
private scratch files; the later execute and rehearsal commands also run generated
Go. Local execution requires the operator's authority and is not an OS sandbox.
Keep real proposals and reports private as appropriate.

```bash
set -euo pipefail
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
ROOT="$(git rev-parse --show-toplevel)"
ROOT="$(cd -- "$ROOT" && pwd -P)"
: "${WORK_PARENT:?set an authorized output parent outside the checkout}"
WORK_PARENT="$(cd -- "$WORK_PARENT" && pwd -P)"
case "$WORK_PARENT/" in
  "$ROOT/"*) echo 'Choose an output parent outside the checkout' >&2; exit 1 ;;
esac
umask 077
WORK="$(mktemp -d "$WORK_PARENT/foundry-current.XXXXXXXX")"
mkdir "$WORK/tmp" "$WORK/cache"
export ROOT WORK TMPDIR="$WORK/tmp" GOTMPDIR="$WORK/tmp" GOCACHE="$WORK/cache"
cd "$ROOT/modeling/lifecycle"
go build -o "$WORK/foundry" ./cmd/foundry
"$WORK/foundry" prepare \
  --intake "$ROOT/skills/model-driven-design/examples/intake.json" \
  --requirements "$ROOT/skills/model-driven-design/examples/brief.md" \
  --model "$ROOT/skills/model-driven-design/examples/model.json" \
  --suite "$ROOT/skills/model-driven-design/examples/suite.json" \
  --manifest "$ROOT/skills/model-driven-design/examples/manifest.json" > "$WORK/review-request.json"
python3 -I - "$WORK/review-request.json" "$WORK/proposal-request.json" <<'PY'
import json, pathlib, sys
request = json.loads(pathlib.Path(sys.argv[1]).read_bytes())
pathlib.Path(sys.argv[2]).write_bytes(request['proposalRequest'].encode('utf-8'))
PY
"$WORK/foundry" review < "$WORK/review-request.json" > "$WORK/review.json"
"$WORK/foundry" proposal check < "$WORK/proposal-request.json" > "$WORK/preflight.json"
printf 'Private results: %s\n' "$WORK"
```

Preparation fills null observations in a new intake text and retains every source
byte outside its `observed` value. Existing observations must exactly match the
supplied files or preparation refuses. Brief, model, suite and manifest texts
retain their original UTF-8 bytes, including line endings; declarations,
dependencies and literal expectations are preserved. Source files are never
rewritten. The small Python block extracts the unchanged native proposal request;
it does not rebuild or normalize its artifacts.

All five flags are required once. Preparation reads only their regular UTF-8 files,
refusing symlink leaves and nonregular inputs. Intake is bounded at 4 MiB;
requirements/model/suite at 1 MiB each; manifest at 64 KiB. The new native requests
also retain their limits. Parent path aliases are allowed; the trusted host
assumption does not protect against a hostile concurrent writer. Preparation
checks the intake scaffold and assembles a request. Model meaning, suite validity,
and manifest bindings are checked by the separate native review/preflight routes.

The fictional input retains `Q_use`, its unanswered real-use question. A successful
check does not answer it. For your own proposal, follow the handoff contract and
supply your own five input files. Use `foundry help` for the route list. Native
routes preserve standalone tool diagnostic prefixes and exit meanings; preparation
uses exit 2 for refused input and exit 3 for failed or short request delivery.

Read `review.json` completely: `modelComparison` and `suiteComparison` should
both be `matches` for these unchanged files. Decode `intakeReport.intakeReportJSON`
and `proposalReportJSON` to inspect native statuses and questions. The outer
report says `executionStatus: not-run` and `reviewStatus: not-established`.
`preflight.json` should say `outcome: checked`, `candidateStarted: false`, and
compilation/replay `not-requested`. Its nested runner binding is a non-executing
model context, not a compiled candidate. Source hashes bind bytes, not provenance.

## Render the same model and decide separately about execution

```bash
"$WORK/foundry" process json < "$ROOT/skills/model-driven-design/examples/model.json" > "$WORK/process.json"
"$WORK/foundry" process svg < "$ROOT/skills/model-driven-design/examples/model.json" > "$WORK/process.svg"
```

Inspect the SVG in a trusted native viewer at readable zoom. Check the header,
each complete transition lane, both refusal branches and the guard glossary;
compare them with the JSON and original model. Long views need scrolling or
zooming: a thumbnail alone is not a layout review. Native rendering is separate
from XML/source checks. Preserve the SVG bytes used for that review; neither a
valid SVG nor a visual pass authenticates the design or its execution.

Only after separate authorization to execute the reviewed local candidate, run
the saved request unchanged:

```bash
"$WORK/foundry" proposal execute < "$WORK/proposal-request.json" > "$WORK/replay.json"
```

For this invented suite, the pre-existing nine case literals each have one step.
A matching receipt reports `compilation: passed`, `replay: matched`,
`candidateStarted: true`, `candidateExitCode: 0`, `cleanup: removed`, and a
candidate report with nine completed cases and nine steps. Read every expected
and actual result. Expected refusals can be matching comparisons. The command
packages but does not run generated structural tests. Its local observations do
not close the example's real-use question or authorize integration.

Check exits and report fields together. `modelproposal` exit 1 can mean stale
preflight with no candidate start or a behavioral mismatch after execution. Its
[contract](../modeling/lifecycle/proposalreplay/README.md) defines the distinct
stages and failure codes. A mismatch is a reason to investigate; do not rewrite
literals from observed output. Abrupt interruption or cleanup failure can leave
scratch files, and failure after candidate start does not imply rollback.

## Inspect a separate fixed rehearsal carrier

This is the existing invented revision rehearsal, not the first-design proposal
above. When that fixed local execution is authorized, the established
[rehearsal script](../modeling/lifecycle/rehearsal/README.md) can produce a carrier:

```bash
cd "$ROOT/modeling/lifecycle"
"$ROOT/modeling/lifecycle/rehearsal/run.sh" --handoff > "$WORK/rehearsal.json"
python3 -I - <<'PY'
import base64, json, os, pathlib
p, work = pathlib.Path('.'), pathlib.Path(os.environ['WORK'])
e = json.loads((p / 'rehearsal/testdata/expected.json').read_bytes())
i = json.loads((p / 'rehearsalhandoff/testdata/report-identities.json').read_bytes())
x = e['externalExpected']
refs = dict(requestSHA256=x['RequestSHA256'], beforeSourceSHA256=x['BeforeSourceSHA256'],
            beforeModelDigest=x['BeforeModelDigest'], candidateSHA256=x['CandidateSHA256'],
            candidateModelDigest=x['CandidateModelDigest'], candidateSuiteSHA256=x['SuiteSHA256'],
            baselineSuiteSHA256=e['baselineSuiteSHA256'], baselineReportSHA256=i['baseline']['sha256'],
            candidateReportSHA256=i['candidate']['sha256'])
request = dict(formatVersion='lifecycle-rehearsal-inspection-request/v1', references=refs,
               bundleBase64=base64.b64encode((work / 'rehearsal.json').read_bytes()).decode('ascii'))
(work / 'inspection-request.json').write_bytes(json.dumps(request).encode('utf-8'))
PY
"$WORK/foundry" inspect < "$WORK/inspection-request.json" > "$WORK/inspection.json"
```

The reference identities come from the existing separately authored fixed
fixtures, not from the carrier or the run being inspected. Keep the carrier
bytes unchanged. Those fixture identities apply only to this example; real
carriers require separately obtained, reviewed references.

The fixed receipt reports baseline two cases/two steps and candidate four
cases/six steps, both matched. It still says `executionStatus: not-established`
and `reportStatus: inspected-for-consistency`: the inspector checks supplied
records without invoking their reducer or authenticating their origin. The
legacy carrier verifier continues treating report blobs as opaque hash-matched
bytes. A valid early mismatch is an inspected mismatch, not malformed evidence
and not a passing replay. See the inspector's contract for limits and failures.

Retain original inputs, full reports, actual command results and any visual
review notes for the responsible reviewer. Decide the next engineering or
integration step from those results and the unresolved requirements. Remove
private scratch outputs when your retention policy permits.
