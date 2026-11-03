# Try the model-driven lifecycle pilot

This guide shows what to run and what the results mean for the small, isolated
lifecycle example on merged main at
[`81c6647b`](https://github.com/spik3r/clanker-foundry/commit/81c6647b00c56cdb8341a160e4fc06faf4f8e55d).
Later revisions may change commands or outputs. For the motivation and the wider
ambition, see the [pipeline introduction](model-driven-pipeline.md); its own
merged-capabilities table is an earlier, explicitly dated snapshot.

The model is structured data describing finite states, transitions, guards and
effects, not an AI model. Fixed tools validate it and generate a small Go reducer,
structural tests and views. The commands below run **invented, fixed examples**;
they do not ask an AI to design an application or turn a prompt into software.
They compile and execute trusted local code, but do not start Foundry workers,
deploy, integrate with an application or establish real task authority.

## Before running anything

Use a trusted local, single-writer checkout of the reviewed revision. The examples
need Bash, Python 3, **installed Go 1.27.3** on `PATH`, sufficient writable temporary storage
and a writable Go cache outside the checkout. The Go commands disable module and toolchain downloads;
they do not install dependencies. No Node or AI provider is needed for these
examples. Run them only when local execution is authorized: the scripts build and
run Go, including generated code. Temporary copies limit which files are changed;
they are **not an OS security sandbox**, a resource limit or protection against a
hostile concurrent writer. Abrupt termination or cleanup failure can leave files.

From the checkout root, run this setup block first. Later blocks use the same
Bash session; choose the fixed demonstrations or the check-only custom-proposal
path below. This creates a new private directory outside the checkout for reports,
cache and temporary files; it does not overwrite the checkout's generated files. Set `PILOT_PARENT` to
an existing, authorized private directory outside the checkout with sufficient
free space before starting, or the commands use `/tmp`.

```bash
set -euo pipefail
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"
ROOT="$(pwd -P)"
WORK_PARENT="$(cd -- "${PILOT_PARENT:-/tmp}" && pwd -P)"
case "$WORK_PARENT/" in
  "$ROOT/"*) echo "Choose a parent outside the checkout" >&2; exit 1 ;;
esac
umask 077
WORK="$(mktemp -d "$WORK_PARENT/foundry-pilot.XXXXXXXX")"
mkdir "$WORK/tmp" "$WORK/go-cache"
export TMPDIR="$WORK/tmp" GOTMPDIR="$WORK/tmp" GOCACHE="$WORK/go-cache"
printf 'Private results: %s\n' "$WORK"
```

If the version check fails, install or select the reviewed toolchain through your
normal approved process and start again; do not silently substitute another version.
The reports can contain input and task data. Keep them private as appropriate.
Redirected files can contain partial output after a failure: check exit status and
never treat a partial file as a success receipt.

## 1. Run one design change

```bash
cd "$ROOT"
modeling/lifecycle/rehearsal/run.sh > "$WORK/rehearsal.json"
python3 -I - "$WORK/rehearsal.json" <<'PY'
import json, sys
with open(sys.argv[1], encoding='utf-8') as f:
    r = json.load(f)
assert r['mode'] == 'example-only' and r['authority'] == 'none'
assert r['baselineMatched'] and r['baselineSteps'] == 2
assert r['candidateMatched'] and r['candidateSteps'] == 6
assert r['candidateBindingRefusedByBaseline'] and r['baselineBindingRefusedByCandidate']
assert r['modelReviewRequired'] and r['suiteReviewRequired']
assert r['topologyMode'] == 'structural-only'
assert r['archiveConsistent'] and r['archiveSuiteStatus'] == 'unvalidated and unexecuted'
assert r['archiveBindingStatus'] == 'not inspected'
print('Baseline: 2 steps matched; candidate: 6 steps matched; stale bindings refused.')
print('Model and suite review required; archive suite unvalidated/unexecuted.')
PY
```

The script applies a fixed proposed owner guard, reviews declared model/suite
changes and structural topology, checks archive consistency, then generates,
compiles and replays the candidate in a private copy. The two and six steps are
separate baseline and candidate observations. Expected refusals can **match**:
a matched test does not mean every attempted transition succeeded. The old suite
is refused against the new model and vice versa; stale evidence is not a test run.

The archive receipt says its suite is **unvalidated and unexecuted** and its
binding **not inspected**. The later candidate replay is a separate observation;
it does not rewrite that archive receipt. `modelReviewRequired` and
`suiteReviewRequired` are change indicators, not completed approvals. The full
[design-change rehearsal](../modeling/lifecycle/rehearsal/README.md) explains the
frozen expectations, cleanup and failure behavior.

## 2. Try two concurrent design edits

```bash
cd "$ROOT"
modeling/lifecycle/mergerehearsal/run.sh > "$WORK/merge.json"
python3 -I - "$WORK/merge.json" <<'PY'
import json, sys
with open(sys.argv[1], encoding='utf-8') as f:
    r = json.load(f)
assert r['mode'] == 'example-only' and r['authority'] == 'none'
assert r['generationRuns'] == r['replayRuns'] == 6
assert r['cleanup'] == 'completed' and r['inputs'] == 'unchanged'
assert {(m['id'], m['status']) for m in r['merges']} == {
    (kind + '-' + order, kind)
    for kind in ('merged', 'conflicted', 'invalid-combination')
    for order in ('left-right', 'right-left')
} and len(r['merges']) == 6
assert set(r['revisions']) == {'base', 'left', 'right', 'merged'}
assert all(v['cases'] == 7 and v['steps'] == 9 for v in r['revisions'].values())
for m in r['merges']:
    if m['status'] == 'merged':
        assert m['generation'] == m['replay'] == 'observed'
        assert m['candidate']['cases'] == 7 and m['candidate']['steps'] == 9
    else:
        assert m['status'] in ('conflicted', 'invalid-combination')
        assert m['generation'] == m['replay'] == 'not-run'
print('Six generation/replay runs; both merge orders matched 7 cases / 9 steps.')
print('Conflict and invalid-combination examples stopped before generation/replay.')
PY
```

This fixed parcel example merges two actor-rule changes in both orders, compiles
and replays successful candidates, and checks base, left, right and merged
versions. It also checks deliberately removed guards and a wrong expectation;
those controls are detected as mismatches, not silently repaired. A merge command
can return exit 0 for a valid **conflicted** or **invalid-combination** report;
inspect its status rather than treating exit 0 as a merged design. See the
[concurrent-design rehearsal](../modeling/lifecycle/mergerehearsal/README.md) for
its witnesses and exact limits.

## 3. See review preparation versus candidate execution

This example uses the fixed ParcelSeal proposal and its separately authored,
frozen literals. It does not resolve a real person's open questions. First build
the existing command and construct a request from original bytes; the inner JSON
files are embedded as text, not parsed and rewritten. This step creates no
candidate project.

```bash
cd "$ROOT/modeling/lifecycle"
mkdir "$WORK/proposal"
go build -o "$WORK/proposal/modelproposal" ./cmd/modelproposal
python3 -I - <<'PY' > "$WORK/proposal/request.json"
import json, pathlib, sys
p = pathlib.Path('proposalreplay/testdata')
r = {'formatVersion': 'lifecycle-proposal-request/v1'}
for role in ('model', 'suite', 'manifest'):
    r[role] = (p / (role + '-a.json')).read_bytes().decode('utf-8')
sys.stdout.buffer.write(json.dumps(r, ensure_ascii=False).encode('utf-8'))
PY
"$WORK/proposal/modelproposal" check < "$WORK/proposal/request.json" > "$WORK/proposal/check.json"
python3 -I - "$WORK/proposal/check.json" <<'PY'
import json, sys
with open(sys.argv[1], encoding='utf-8') as f:
    r = json.load(f)
assert r['outcome'] == 'checked' and r['validationContext'] == 'non-executing-model-context'
assert r['preflight']['replay'] == 'not-run' and r['candidateStarted'] is False
assert r['compilation'] == r['replay'] == 'not-requested'
print('Preflight checked; candidate compilation and replay not requested.')
PY
```

The nested `runnerBinding: current` in this check is part of a **non-executing
model context**, not evidence of a compiled candidate. Check mode launches no
process and does not execute the suite. A stale declaration stops before execution;
it is not repaired. An exit code alone is not a substitute for reading the report.

**Stop here if execution is not separately authorized.** For this fixed invented
example, an operator who is authorized to execute local generated code can run
this separate command from the lifecycle module directory:

```bash
cd "$ROOT/modeling/lifecycle"
"$WORK/proposal/modelproposal" execute < "$WORK/proposal/request.json" > "$WORK/proposal/execute.json"
python3 -I - "$WORK/proposal/execute.json" <<'PY'
import json, sys
with open(sys.argv[1], encoding='utf-8') as f:
    r = json.load(f)
assert r['outcome'] == 'matched' and r['compilation'] == 'passed'
assert r['replay'] == 'matched' and r['candidateStarted'] is True
assert r['candidateExitCode'] == 0 and r['cleanup'] == 'removed'
assert r['candidateReport']['matched']
assert r['candidateReport']['casesCompleted'] == 4
assert r['candidateReport']['stepsExecuted'] == 6
print('Candidate compiled; 4 cases / 6 steps matched; cleanup completed.')
PY
```

`execute` verifies the closed source/archive inputs, creates a private project,
compiles a runner bound to this candidate and replays the unchanged suite. Its
packaged generated structural tests are **not run by this command**. Read the
complete report, including individual expected and actual results, not just counts.
A mismatch exits 1 with a report; malformed input exits 2 without a receipt; other
host, output or cleanup failures have distinct statuses. See the
[proposal replay contract](../modeling/lifecycle/proposalreplay/README.md) for
all exit codes, limits and cleanup semantics. Failure after candidate start does
not imply rollback. Archive consistency and preflight are not replay results.

## Check your own proposal without running it

This is a separate path from the fixed demonstrations above. You can run the
initial **Before running anything** setup and then skip directly here: sections 1
and 2 execute fixed examples, but are not prerequisites for this check. Use this
path only for a proposal that fits the existing `lifecycle/v1` profile, after
resolving the choices needed for review. AI can help draft it, but the person
responsible must review the meaning and unresolved questions. Have the expected
outcomes independently stated and reviewed against the requirements, not derived
from a generated reducer or an observed run.

Choose an existing, authorized `PROPOSAL_DIR` containing **`model.json`,
`suite.json` and `manifest.json`**. Set and export `PROPOSAL_DIR` to its absolute
path in the same Bash session as the setup. The
[proposal replay contract](../modeling/lifecycle/proposalreplay/README.md) defines
the inputs and limits; the [first-design handoff](../skills/model-driven-design/references/foundry-handoff.md)
explains preparation and review. Keep originals unchanged and handle all request
and report files as potentially sensitive. The following builds an existing tool
and performs **check only**; it neither compiles nor invokes the candidate. It
creates a new `$WORK/custom` directory and refuses to reuse one.

```bash
: "${PROPOSAL_DIR:?set PROPOSAL_DIR to your reviewed proposal directory}"
export PROPOSAL_DIR
cd "$ROOT/modeling/lifecycle"
mkdir "$WORK/custom"
go build -o "$WORK/custom/modelproposal" ./cmd/modelproposal
python3 -I - <<'PYREQUEST' > "$WORK/custom/request.json"
import json, os, pathlib, sys
p = pathlib.Path(os.environ['PROPOSAL_DIR'])
r = {'formatVersion': 'lifecycle-proposal-request/v1'}
for role in ('model', 'suite', 'manifest'):
    r[role] = (p / (role + '.json')).read_bytes().decode('utf-8')
sys.stdout.buffer.write(json.dumps(r, ensure_ascii=False).encode('utf-8'))
PYREQUEST
if "$WORK/custom/modelproposal" check < "$WORK/custom/request.json" > "$WORK/custom/check.json"; then
  CHECK_EXIT=0
else
  CHECK_EXIT=$?
fi
printf 'Check command status: %s; report: %s\n' "$CHECK_EXIT" "$WORK/custom/check.json"
if [ "$CHECK_EXIT" -eq 0 ] || [ "$CHECK_EXIT" -eq 1 ]; then
  python3 -I - "$WORK/custom/check.json" <<'PYREPORT'
import json, sys
try:
    with open(sys.argv[1], encoding='utf-8') as f:
        r = json.load(f)
except (OSError, ValueError):
    raise SystemExit('No complete JSON receipt; do not assume success.')
p = r['preflight']
print('outcome:', r['outcome'], '; context:', r['validationContext'])
print('bindings:', p['modelBinding'], p['suiteBinding'], p['suiteModelBinding'])
print('suite validation:', p['suiteValidation'], '; preflight replay:', p['replay'])
print('source SHA-256:', r['sources'])
PYREPORT
else
  echo 'Do not assume a complete receipt; see the proposal replay exit-code contract.'
fi
```

The short readers display selected fields; they do not authenticate a receipt or
replace review of the full report. A command status can also reflect a shell or
redirection failure before the CLI starts; do not classify it without a complete
receipt. A CLI exit 0 with `outcome: checked` is validation-only: inspect the full
preflight, source identities and unresolved human decisions. Its `replay: not-run`
is not an observed behavioral pass. The nested current runner binding is a
**non-executing model context**, not a compiled runner. Exit 1 indicates stale
preflight: inspect its binding and `not-run` fields, retain the old evidence, and
stop for review rather than silently rebinding or changing expectations. Exit 2
means invalid arguments/request or input read failure and no receipt; exit 3 can
leave a partial receipt; exit 4 indicates a host, tool or cleanup boundary failure.
Do not infer success from an empty or partial output. The linked contract gives
the full meanings; compare exact source hashes with the reviewed original bytes.

### Separately authorized replay

**Stop after check unless the operator has explicit authority to execute this
candidate locally.** Review the requirements, model, independently authored
expectations and any required human decisions first. An AI draft, a valid check,
a matching hash or a command shown here is not that authorization. The command
below uses the request bytes saved above and compiles and runs generated Go on
the trusted host; it is not a sandbox. Verify that those bytes are still the
reviewed proposal. This creates a private project and uses outside-checkout
temporary storage and cache; cleanup is attempted, not guaranteed after abrupt
termination. It does not deploy, integrate or start workers. It refuses to
overwrite an existing report; use a fresh `$WORK` for a new attempt.

```bash
cd "$ROOT/modeling/lifecycle"
if [ "${CHECK_EXIT:-}" != 0 ]; then
  echo "Stop: check did not complete successfully" >&2; exit 1
fi
test ! -e "$WORK/custom/replay.json"
test ! -L "$WORK/custom/replay.json"
if "$WORK/custom/modelproposal" execute < "$WORK/custom/request.json" > "$WORK/custom/replay.json"; then
  REPLAY_EXIT=0
else
  REPLAY_EXIT=$?
fi
printf 'Replay command status: %s; report: %s\n' "$REPLAY_EXIT" "$WORK/custom/replay.json"
if [ "$REPLAY_EXIT" -eq 0 ] || [ "$REPLAY_EXIT" -eq 1 ]; then
  python3 -I - "$WORK/custom/replay.json" <<'PYREPORT'
import json, sys
try:
    with open(sys.argv[1], encoding='utf-8') as f:
        r = json.load(f)
except (OSError, ValueError):
    raise SystemExit('No complete JSON receipt; do not assume success.')
print('outcome:', r['outcome'], '; compilation:', r['compilation'], '; replay:', r['replay'])
print('candidate started:', r['candidateStarted'], '; cleanup:', r['cleanup'])
report = r['candidateReport']
if report is not None:
    print('matched:', report['matched'], '; cases:', report['casesCompleted'],
          '; steps:', report['stepsExecuted'])
PYREPORT
else
  echo 'Do not assume a complete receipt; see the proposal replay exit-code contract.'
fi
```

Counts and outcomes depend on your suite; the fixed example's four cases and six
steps are **not** an expected result for your proposal. Read the entire candidate
report, including expected and actual results and differences, against the
reviewed requirements. Exit 1 can mean stale preflight with no candidate launch
**or** a mismatched replay; the stage fields distinguish them. A matched refusal
is a passing comparison, not a successful transition. A mismatch is evidence to
investigate, not permission to rewrite a literal. For exit 3 or 4, consult the
contract: a partial receipt or a failure after candidate start does not establish
rollback or that nothing ran. The command packages but does not run generated
structural tests. Matching replay is an observed comparison, not approval,
authenticated provenance, visual review, coverage of all behavior or permission
to integrate or deploy.

## Moving beyond the fixed examples

For a new idea, start with the [model-driven design skill](../skills/model-driven-design/README.md)
and its [first-design handoff](../skills/model-driven-design/references/foundry-handoff.md).
It clarifies the brief and profile fit, records unresolved choices, and prepares a
proposal and review material. AI may help draft; it cannot fill missing human
decisions or grant execution permission. Have expectations stated and reviewed
independently from the generated implementation and its outputs. A declared author
label is not proof of independence. Freeze those literals before execution and
investigate a mismatch instead of changing expectations to make a run pass.

The [revision tutorial](../modeling/lifecycle/tutorial/README.md) is a longer
step-by-step walkthrough of stale evidence and a changed model. Its full manual
path also needs Node 24+ and pinned Ajv, whose installation may need network access;
`go test ./tutorial` has different prerequisites described there. The
[portable rehearsal carrier](../modeling/lifecycle/rehearsalhandoff/README.md)
checks supplied report bytes against supplied hashes; it does not inspect those
reports or prove replay occurred.

Generated `model_paths_test.go` checks structural properties derived from the same
model; it is not an independent behavior oracle and some properties may be vacuous.
Separately specified scenario expectations check the observed examples, not all
possible behavior. Hashes establish consistency, not authenticated provenance.
Repeatability does not prove correctness. The SVG and tables are review aids, not
full UML, SysML, BPMN, UTP or fUML implementations or conformance claims; these
commands do not render or visually review a diagram. People must still review
requirements, model meaning, expectations and any actual visualization, and make
separate decisions about execution, integration and deployment. No matching result
or receipt supplies those decisions.

When finished, retain or remove the private directory printed above according to
your local data-handling policy. If a run was interrupted, inspect for remaining
temporary files before cleanup; the commands cannot guarantee descendant cleanup.
