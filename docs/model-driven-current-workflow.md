# Use the current model-driven review path

This guide describes selected tools on reviewed main
[`83007fa`](https://github.com/spik3r/clanker-foundry/commit/83007fa3083dc46add5f83e628d2aac30d1a972b),
on **4 November 2026**. It is a practical entry point, not an exhaustive inventory.
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
| Prepare a proposal | [Design skill](../skills/model-driven-design/SKILL.md), [profile catalog](../modeling/lifecycle/profilecatalog/README.md), [acceptance authoring](../modeling/lifecycle/acceptanceauthor/README.md) | Requirements, model, independently stated suite literals, manifest and open questions. Author labels do not authenticate independence. |
| Correlate intake and proposal | [`modelreviewcheck`](../modeling/lifecycle/firstdesignreview/README.md) | Full native results, exact model/suite byte comparisons and retained questions. Exit 0 only means a report was produced. This check starts no candidate. |
| Check or execute a proposal | [`modelproposal check\|execute`](../modeling/lifecycle/proposalreplay/README.md) | Check is non-executing preflight. Separately authorized execute compiles and invokes a candidate-bound runner and preserves its comparisons. |
| Inspect a process view | [`modelprocess json\|svg`](../modeling/lifecycle/processview/README.md) | Declared guards, refusal paths, effect order, destination and canonical model identity. A rendered view does not establish execution or approval. |
| Inspect carried observations | [`modelrehearsalinspect`](../modeling/lifecycle/rehearsalinspection/README.md) | Native report consistency and binding to carried suite literals. Invented consistent reports can pass; execution remains not established. |

The process view is a review aid, with invalid input and illegal transitions
outside its per-transition lanes. It makes no BPMN conformance claim. The
[runtime trace](../runtime/model-trace-v1/README.md) is a separate partial map of
the existing orchestration evaluator; this design path does not dispatch workers.

## Prepare private outputs and check a first design

Use Bash, Python 3, Git and installed **Go 1.27.3** on a trusted, exclusive local
checkout of the reviewed revision. Set `WORK_PARENT` to an authorized directory
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
for tool in modelreviewcheck modelproposal modelprocess modelrehearsalinspect; do
  go build -o "$WORK/$tool" "./cmd/$tool"
done
python3 -I - <<'PY'
import json, os, pathlib
root, work = pathlib.Path(os.environ['ROOT']), pathlib.Path(os.environ['WORK'])
p = root / 'skills/model-driven-design/examples'
text = lambda name: (p / name).read_bytes().decode('utf-8')
intake = json.loads(text('intake.json'))
intake['observed'] = dict(requirements=text('brief.md'), model=text('model.json'),
                          scenarioSuite=text('suite.json'))
proposal = dict(formatVersion='lifecycle-proposal-request/v1',
                model=text('model.json'), suite=text('suite.json'), manifest=text('manifest.json'))
encode = lambda value: json.dumps(value, ensure_ascii=False).encode('utf-8')
intake_bytes, proposal_bytes = encode(intake), encode(proposal)
(work / 'intake-request.json').write_bytes(intake_bytes)
(work / 'proposal-request.json').write_bytes(proposal_bytes)
wrapper = dict(formatVersion='lifecycle-first-design-review-request/v1',
               intakeRequest=intake_bytes.decode('utf-8'),
               proposalRequest=proposal_bytes.decode('utf-8'))
(work / 'review-request.json').write_bytes(encode(wrapper))
PY
"$WORK/modelreviewcheck" < "$WORK/review-request.json" > "$WORK/review.json"
"$WORK/modelproposal" check < "$WORK/proposal-request.json" > "$WORK/preflight.json"
printf 'Private results: %s\n' "$WORK"
```

This constructs a new intake request with observations. Embedded brief, model,
suite and manifest texts retain their original UTF-8 bytes, including line endings;
they are not parsed and re-encoded. The fictional first-design input retains
`Q_use`, its unanswered real-use question. A successful check does not answer it.
For your own proposal, follow the handoff contract rather than importing this
example's policy decisions.

Read `review.json` completely: `modelComparison` and `suiteComparison` should
both be `matches` for these unchanged files. Decode `intakeReport.intakeReportJSON`
and `proposalReportJSON` to inspect native statuses and questions. The outer
report says `executionStatus: not-run` and `reviewStatus: not-established`.
`preflight.json` should say `outcome: checked`, `candidateStarted: false`, and
compilation/replay `not-requested`. Its nested runner binding is a non-executing
model context, not a compiled candidate. Source hashes bind bytes, not provenance.

## Render the same model and decide separately about execution

```bash
"$WORK/modelprocess" json < "$ROOT/skills/model-driven-design/examples/model.json" > "$WORK/process.json"
"$WORK/modelprocess" svg < "$ROOT/skills/model-driven-design/examples/model.json" > "$WORK/process.svg"
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
"$WORK/modelproposal" execute < "$WORK/proposal-request.json" > "$WORK/replay.json"
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
"$WORK/modelrehearsalinspect" < "$WORK/inspection-request.json" > "$WORK/inspection.json"
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
