# Walk through a lifecycle revision

This invented example starts with a draft lifecycle whose `prepare` transition
allows a blank actor. The revision requires a nonblank actor. The initial and
revised requirement statements and literal success/refusal suites are fixed in
this directory. A separate read-only agent reviewed the revision and its expected
outcomes before either tutorial suite was run; the labels inside the files are
unverified example declarations, not proof of independent authorship or human
approval. A reader should arrange their own appropriate review of intent, model and
expectations before treating the updated suite as authoritative.

The walkthrough uses the merged intake, authoring schema, compiler, acceptance
authoring and scenario replay. It never updates the literal expectations from a
run. It does not operate Foundry workers or establish real review, authentication,
production policy or approval. Reports can contain supplied task data.

## Prerequisites and an isolated copy

Use Bash, Python 3, Node 24 or later, and the exact Go **1.27.3** toolchain. Start
at the repository root. Installing the pinned Ajv dependency may require network
access if it is not cached; all Go operations below use offline module resolution.
The commands create a disposable module copy, never overwrite this checkout, and
leave the copy for inspection. Keep the resulting `$TUT_DIR` private.

Run the code blocks in order in the same Bash session:

```bash
set -euo pipefail
test "$(go env GOVERSION)" = go1.27.3
export TUT_DIR="$(mktemp -d "${TMPDIR:-/tmp}/foundry-tutorial.XXXXXX")"
mkdir "$TUT_DIR/npm"
cp runtime/package.json runtime/package-lock.json "$TUT_DIR/npm/"
(cd "$TUT_DIR/npm" && npm ci --ignore-scripts --no-audit --no-fund)
export TUT_AJV="$TUT_DIR/npm/node_modules/ajv"
cp -R modeling/lifecycle "$TUT_DIR/lifecycle"
cd "$TUT_DIR/lifecycle"
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off
printf 'Disposable tutorial copy: %s\n' "$TUT_DIR"

# Schema validation checks shapes; the Go model validator remains authoritative.
node <<'NODE'
const fs = require('fs');
const Ajv = require(process.env.TUT_AJV);
const schema = JSON.parse(fs.readFileSync('authoring/lifecycle-v1.schema.json', 'utf8'));
const validate = new Ajv({strict:true, allErrors:true}).compile(schema);
for (const path of ['examples/lifecycle.json', 'tutorial/model-v2.json']) {
  if (!validate(JSON.parse(fs.readFileSync(path, 'utf8')))) {
    console.error(path, validate.errors); process.exit(1);
  }
  console.log(path + ': schema valid');
}
NODE

go build -o "$TUT_DIR/modelc" ./cmd/modelc
go build -o "$TUT_DIR/modelintake" ./cmd/modelintake
go build -o "$TUT_DIR/modelacceptance-v1" ./cmd/modelacceptance
go build -o "$TUT_DIR/modelplay-v1" ./cmd/modelplay
"$TUT_DIR/modelc" validate
"$TUT_DIR/modelc" check

# These functions only assemble requests from the frozen files. They do not
# calculate, rewrite or regenerate any expected scenario result or hash.
intake_request() {
  python3 - "$1" "$2" "$3" "$4" <<'PY'
import json, pathlib, sys
read = lambda p: pathlib.Path(p).read_bytes().decode('utf-8')
d = json.loads(read(sys.argv[1]))
d['observed'] = dict(zip(('requirements','model','scenarioSuite'), map(read, sys.argv[2:])))
print(json.dumps(d, ensure_ascii=True))
PY
}
acceptance_request() {
  python3 - "$1" "$2" "$3" "$4" <<'PY'
import json, pathlib, sys
read = lambda p: pathlib.Path(p).read_bytes().decode('utf-8')
print(json.dumps(dict(formatVersion='lifecycle-acceptance-authoring-request/v1',
                      mode=sys.argv[1], manifest=read(sys.argv[2]),
                      model=read(sys.argv[3]), suite=read(sys.argv[4])), ensure_ascii=True))
PY
}
```

## Initial declaration and actual replay

The intake file is a declaration with null observations; the command below adds
observed file bytes to a new request. Matching hashes are consistency checks, not
proof of authorship or execution. The acceptance and scenario commands actually
run the existing compiled v1 reducer against the independently stated literals.

```bash
intake_request tutorial/intake-v1.json tutorial/requirements-v1.md examples/lifecycle.json tutorial/suite-v1.json > "$TUT_DIR/intake-v1-request.json"
"$TUT_DIR/modelintake" < "$TUT_DIR/intake-v1-request.json" > "$TUT_DIR/intake-v1-report.json"
acceptance_request replay tutorial/manifest-v1.json examples/lifecycle.json tutorial/suite-v1.json > "$TUT_DIR/accept-v1-request.json"
"$TUT_DIR/modelacceptance-v1" < "$TUT_DIR/accept-v1-request.json" > "$TUT_DIR/accept-v1-report.json"
"$TUT_DIR/modelplay-v1" < tutorial/suite-v1.json > "$TUT_DIR/scenario-v1-report.json"
# The v1 runner must reject the v2 model binding without running a scenario.
if "$TUT_DIR/modelplay-v1" < tutorial/suite-v2.json > "$TUT_DIR/future-suite-report.json"; then
  echo 'Future suite unexpectedly accepted' >&2; exit 1
else
  status=$?; test "$status" -eq 2; test ! -s "$TUT_DIR/future-suite-report.json"
fi
python3 - "$TUT_DIR" <<'PY'
import json, pathlib, sys
p = pathlib.Path(sys.argv[1]); load = lambda n: json.loads((p/n).read_bytes())
def check(ok):
    if not ok: raise SystemExit('tutorial verification failed')
i = load('intake-v1-report.json'); a = load('accept-v1-report.json'); s = load('scenario-v1-report.json')
check(len(i['bindings']) == 3 and all(x['status'] == 'matches-declaration' for x in i['bindings']))
check(len(i['links']) == 3 and all(x['status'] == 'matches-declaration' for x in i['links']))
check(a['replay'] == 'matched' and a['scenarioReport']['stepsExecuted'] == 3)
check(s['matched'] and s['casesCompleted'] == 3 and s['stepsExecuted'] == 3)
print('v1: declared bytes match; three literal scenarios replayed and matched')
PY
```

## Detect stale evidence

The candidate adds only `G_actor` to `T_prepare`. Feed the changed model to the old
intake and acceptance declarations **before** regeneration. The old intake must
show a model digest mismatch. The acceptance checker must return exit 1 with
`modelBinding: stale`, `suiteValidation: not-checked` and `replay: not-run`.
A stale result is neither a failed scenario nor a completed test.

```bash
intake_request tutorial/intake-v1.json tutorial/requirements-v1.md tutorial/model-v2.json tutorial/suite-v1.json > "$TUT_DIR/intake-stale-request.json"
"$TUT_DIR/modelintake" < "$TUT_DIR/intake-stale-request.json" > "$TUT_DIR/intake-stale-report.json"
acceptance_request replay tutorial/manifest-v1.json tutorial/model-v2.json tutorial/suite-v1.json > "$TUT_DIR/accept-stale-request.json"
if "$TUT_DIR/modelacceptance-v1" < "$TUT_DIR/accept-stale-request.json" > "$TUT_DIR/accept-stale-report.json"; then
  echo 'Unexpected current binding' >&2; exit 1
else
  status=$?; test "$status" -eq 1
fi
python3 - "$TUT_DIR" <<'PY'
import json, pathlib, sys
p = pathlib.Path(sys.argv[1]); load = lambda n: json.loads((p/n).read_bytes())
def check(ok):
    if not ok: raise SystemExit('tutorial verification failed')
i = load('intake-stale-report.json'); a = load('accept-stale-report.json')
check(next(x for x in i['bindings'] if x['kind'] == 'model')['status'] == 'digest-mismatch')
check(a['modelBinding'] == 'stale' and a['suiteValidation'] == 'not-checked')
check(a['replay'] == 'not-run' and a['scenarioReport'] is None)
print('v1 evidence is stale; no candidate scenario was run')
PY
```

## Review the expectation update, then compile and replay

Before proceeding, review [requirements-v2.md](requirements-v2.md), the
`T_prepare` guard change in [model-v2.json](model-v2.json), and every literal in
[suite-v2.json](suite-v2.json) independently of generated code. In particular,
blank actor must refuse with `G_actor` and the Draft at revision 0 unchanged;
named actor still reaches Ready at revision 1; an unknown command remains
`INVALID_INPUT`. The v2 manifest and intake declarations pin the reviewed files.
This tutorial does not collect or assert a human approval. If your process requires
one, stop until it is recorded through that process.

After that review, generate only inside the disposable copy and build new tools
bound to the candidate reducer. The new runner refuses the old suite's model
binding; it must not silently rebind it. Then run the fixed v2 declarations and
expectations:

```bash
cp tutorial/model-v2.json examples/lifecycle.json
"$TUT_DIR/modelc" validate
"$TUT_DIR/modelc" generate
"$TUT_DIR/modelc" check
go build -o "$TUT_DIR/modelacceptance-v2" ./cmd/modelacceptance
go build -o "$TUT_DIR/modelplay-v2" ./cmd/modelplay
if "$TUT_DIR/modelplay-v2" < tutorial/suite-v1.json > "$TUT_DIR/old-suite-report.json"; then
  echo 'Old suite unexpectedly accepted' >&2; exit 1
else
  status=$?; test "$status" -eq 2; test ! -s "$TUT_DIR/old-suite-report.json"
fi
intake_request tutorial/intake-v2.json tutorial/requirements-v2.md tutorial/model-v2.json tutorial/suite-v2.json > "$TUT_DIR/intake-v2-request.json"
"$TUT_DIR/modelintake" < "$TUT_DIR/intake-v2-request.json" > "$TUT_DIR/intake-v2-report.json"
acceptance_request replay tutorial/manifest-v2.json tutorial/model-v2.json tutorial/suite-v2.json > "$TUT_DIR/accept-v2-request.json"
"$TUT_DIR/modelacceptance-v2" < "$TUT_DIR/accept-v2-request.json" > "$TUT_DIR/accept-v2-report.json"
"$TUT_DIR/modelplay-v2" < tutorial/suite-v2.json > "$TUT_DIR/scenario-v2-report.json"
python3 - "$TUT_DIR" <<'PY'
import json, pathlib, sys
p = pathlib.Path(sys.argv[1]); load = lambda n: json.loads((p/n).read_bytes())
def check(ok):
    if not ok: raise SystemExit('tutorial verification failed')
i = load('intake-v2-report.json'); a = load('accept-v2-report.json'); s = load('scenario-v2-report.json')
check(len(i['bindings']) == 3 and all(x['status'] == 'matches-declaration' for x in i['bindings']))
check(len(i['links']) == 3 and all(x['status'] == 'matches-declaration' for x in i['links']))
check(a['replay'] == 'matched' and s['matched'] and s['stepsExecuted'] == 3)
blank, named, unknown = (x['actual'] for x in s['entries'])
check(not blank['ok'] and blank['transitionId'] == 'T_prepare' and blank['refusals'] == ['G_actor'])
check(blank['task']['state'] == 'Draft' and blank['task']['revision'] == '0')
check(named['ok'] and named['task']['state'] == 'Ready' and named['task']['revision'] == '1')
check(not unknown['ok'] and unknown['task'] is None and unknown['refusals'] == ['INVALID_INPUT'])
print('v2: old binding refused; three separately specified scenarios replayed and matched')
PY
```

`go test ./tutorial` independently checks this sequence against fixed expected
results in another disposable module copy; it requires Go 1.27.3 and does not
require Node. A report-snippet regression also runs when Python 3 is available.
The test exercises real compiler and CLI binaries, not a new
product runner. It cannot replace a person's review of intended behavior, and a
passing test does not confer approval. After inspection, remove the disposable
copy using your normal temporary-file cleanup process.
