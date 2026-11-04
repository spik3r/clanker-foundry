# Use the current model-driven review path

The four Bash blocks for the fixed first-design and rehearsal examples below
were exercised on reviewed main
[`ffdce31f`](https://github.com/spik3r/clanker-foundry/commit/ffdce31feb3fd5e8c6c44e9487533a696a0cd5c7)
on **5 November 2026**. That observation did not include native visual review or
the later optional revision and custom-change routes; those sections describe
their linked contracts and require separate inputs. The original native-contract
baseline was
[`2e2a190b`](https://github.com/spik3r/clanker-foundry/commit/2e2a190b629a983ba854d59ecc356274ec7e5a4f),
which predates the `foundry` entrypoint. Use a current checkout containing that
entrypoint to prepare requests and use the selected workflow through one binary.
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
checkout containing the `foundry` entrypoint. Set `WORK_PARENT` to an authorized
directory outside that checkout. The examples compile trusted tools and write only new
private scratch files; the later execute and rehearsal commands also run generated
Go. Local execution requires the operator's authority and is not an OS sandbox.
Keep real proposals and reports private as appropriate.

Run the four fixed-example Bash blocks below in order in the same Bash session.
Keep its shell options, exported variables and working directory between blocks.

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

## Review changed scenario literals between two revisions

For two prepared revisions, the [revision-check contract](../modeling/lifecycle/revisioncheck/README.md)
accepts before/after first-design request texts, trace requirements and catalogs.
Send `lifecycle-revision-check-request/v2` to `foundry revisions` to retain its four
native review reports and add the existing native scenario-suite comparison.
The request fields are the same as v1; only the explicit format version changes.
The standalone `modelrevisioncheck` remains strict v1.

Inspect `scenarioDiff.status` before reading its report. `checked` includes full
native `reportJSON` text and `reportSHA256`; the native report identifies changed
expected fields, inputs and ordering. Retain the original suites to inspect the
actual literals. A stale manifest can coexist with a checked suite comparison.
`not-checked` has null report/hash and a native input/limit refusal reason; it
must not be treated as unchanged expectations. Stale suite bindings are never
repaired by this route. Questions, preflight statuses and other source comparisons
remain visible independently, with no overall approval or readiness decision.
Use the separate model-diff contract for complete model edits, and keep any later
execution or integration decision separate from these declared comparisons.

## Review full model changes beside those literals

Send the same prepared snapshot fields with explicit
`lifecycle-revision-check-request/v3` to `foundry revisions` to include the complete
native model diff as well as all v2 observations. Decode `modelDiffReportJSON` and
check its `modelDiffReportSHA256` against the exact report text. The native report
lists changed model fields and direct transition links in both models. Retain the
original model texts to review the actual changed values.

A stale suite can leave the scenario comparison `not-checked` while the model
diff still shows the changed design. Neither result repairs the other. Source
hashes and canonical model digests answer different questions; a formatting-only
source change can preserve canonical identity, while labels and requirements
are part of the reviewed model. The native `reviewRequired` flag is not a human
approval, a behavioral-equivalence claim or permission to execute. V1 and v2
remain supported; the standalone `modelrevisioncheck` remains strict v1.

## Rehearse a reviewed design change against separate expectations

For a custom change, the standalone [change rehearsal bridge](../modeling/lifecycle/changerehearsal/README.md)
accepts the original native patch request and baseline proposal request, a separately
reviewed candidate suite/manifest, and an explicit expected candidate model digest.
It derives the candidate from the native patch report instead of asking a script
to copy or rebind that model. It preserves the supplied expectation bytes.

Build `./cmd/modelchangerehearsal` from the trusted `modeling/lifecycle` checkout.
Run `modelchangerehearsal check < request.json` and review its full patch and both
native preflight receipts. This command does not compile or replay. When the local
rehearsal is separately authorized, send the **same unchanged request** to
`modelchangerehearsal execute`. Both preflights must be current. The baseline runs
first; only a matched baseline with successful cleanup permits the candidate.
A mismatch, failure or cancellation keeps the not-attempted side's check receipt.
Inspect the full native receipts and actual command exit status; a side's
`operation: execute` means the native executor was called, not that its child ran
or matched. The report supplies no approval or integration decision. This separate
command does not add a `foundry` route or replace revision v2's literal comparison.
