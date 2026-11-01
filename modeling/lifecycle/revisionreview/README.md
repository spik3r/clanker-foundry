# Review a lifecycle revision with the merged tools

This invented review follows the [existing tutorial](../tutorial/README.md):
revision 1 permits a blank actor on `T_prepare`; revision 2 requires the existing
`G_actor` guard. The tutorial's model, suites, manifests and literal expected
behavior remain unchanged. This walkthrough adds separately authored requirements
trace documents and intake declarations for those documents. They are **not** the
tutorial's Markdown requirements or its original intake declarations.

The [two requirements snapshots](testdata/requirements-v1.json) each declare a
link from `UR-101` to `T_prepare`; [revision 2](testdata/requirements-v2.json)
also links `G_actor`. Neither links an acceptance case. The trace's
`lifecycle-semantic-oracle-v2` catalog is separate from the tutorial scenario
suites; do not transfer case IDs or infer a case-to-requirement link. Both trace
snapshots therefore retain `no-acceptance-link`. The intake question `Q_case_link`
remains open across the revision.

Use the reports together, preserving their different meanings:

1. Compare the original intakes. `modelintakediff` shows intent, decision and
   artifact-assertion changes and the unchanged open question. Its artifact
   validation and execution statuses remain `not-run`.
2. Check observed candidate bytes with `modelintakecheck`. A model can be
   `accepted` by the strict parser while its old declared hash is mismatched.
   It preserves the native intake receipt, original and canonical identities,
   and the schema reference with `validation: not-run`.
3. Compare the two model/requirements/catalog snapshots with
   `modeltraceimpact`. Expect a changed model and requirement, an `UR-101`
   candidate linked to `T_prepare`, and the missing acceptance link on both
   sides. This compares declarations, not behavioral coverage or test results.
4. Check the unchanged authoring manifests and literal scenario suites. With the
   baseline runner, the old suite is stale for the candidate; the candidate's
   own suite and model match each other but the runner binding is stale. Both
   report `suiteValidation: not-checked` and `replay: not-run`. The baseline
   suite is valid in check mode but also `not-run`. A reviewer must separately
   review the changed intent, model and literal expectations; this walkthrough
   does not record a review or approve or authorize anything.

## Run the review locally

From `modeling/lifecycle`, use Bash, Python 3 and Go **1.27.3**. All inputs are
invented. This creates only a temporary directory of binaries, assembled requests
and reports; it does not alter source files, declarations, the baseline reducer,
manifests or literal expected results. Reports include observations and may be
sensitive for real inputs. The tools are offline; redirection belongs to the shell.

```bash
set -euo pipefail
test "$(go env GOVERSION)" = go1.27.3
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off GOFLAGS=
export REV_DIR="$(mktemp -d "${TMPDIR:-/tmp}/foundry-revision-review.XXXXXX")"
for name in modelintakediff modelintakecheck modeltraceimpact modelacceptance; do
  go build -o "$REV_DIR/$name" "./cmd/$name"
done

# Assemble temporary requests by copying exact UTF-8 bytes. Existing declarations
# and expected results are not computed, changed, or rebound by this preparation.
python3 <<'PY'
import json, os, pathlib
out = pathlib.Path(os.environ['REV_DIR'])
read = lambda p: pathlib.Path(p).read_bytes().decode('utf-8')
write = lambda name, value: (out/name).write_bytes(json.dumps(value, ensure_ascii=True).encode('utf-8'))
model1, model2 = read('examples/lifecycle.json'), read('tutorial/model-v2.json')
req1, req2 = (read('revisionreview/testdata/requirements-v%d.json' % i) for i in (1, 2))
suite1, suite2 = (read('tutorial/suite-v%d.json' % i) for i in (1, 2))
manifest1, manifest2 = (read('tutorial/manifest-v%d.json' % i) for i in (1, 2))
intake1, intake2 = (json.loads(read('revisionreview/testdata/intake-v%d.json' % i)) for i in (1, 2))
def observed(decl, req, model, suite):
    result = dict(decl)
    result['observed'] = {'requirements': req, 'model': model, 'scenarioSuite': suite}
    return result
before = observed(intake1, req1, model1, suite1)
stale = observed(intake1, req2, model2, suite2)
after = observed(intake2, req2, model2, suite2)
for name, data in [('before.json', before), ('stale.json', stale), ('after.json', after)]:
    write(name, data)
write('intake-diff-request.json', {'formatVersion':'lifecycle-intake-impact-request/v1',
      'before':read(str(out/'before.json')), 'after':read(str(out/'after.json'))})
catalog = read('examples/oracle.json')
write('trace-diff-request.json', {'formatVersion':'lifecycle-trace-impact-request/v1',
      'before':{'model':model1, 'requirements':req1, 'catalog':catalog},
      'after':{'model':model2, 'requirements':req2, 'catalog':catalog}})
for name, manifest, model, suite in [('baseline',manifest1,model1,suite1),
                                      ('old-candidate',manifest1,model2,suite1),
                                      ('new-candidate',manifest2,model2,suite2)]:
    write(name+'-request.json', {'formatVersion':'lifecycle-acceptance-authoring-request/v1',
          'mode':'check', 'manifest':manifest, 'model':model, 'suite':suite})
PY

"$REV_DIR/modelintakediff" < "$REV_DIR/intake-diff-request.json" > "$REV_DIR/intake-diff.json"
"$REV_DIR/modelintakecheck" < "$REV_DIR/stale.json" > "$REV_DIR/stale-model.json"
"$REV_DIR/modelintakecheck" < "$REV_DIR/after.json" > "$REV_DIR/candidate-model.json"
"$REV_DIR/modeltraceimpact" < "$REV_DIR/trace-diff-request.json" > "$REV_DIR/trace-diff.json"
"$REV_DIR/modelacceptance" < "$REV_DIR/baseline-request.json" > "$REV_DIR/baseline-acceptance.json"
for name in old-candidate new-candidate; do
  if "$REV_DIR/modelacceptance" < "$REV_DIR/$name-request.json" > "$REV_DIR/$name-acceptance.json"; then
    echo "Unexpected current binding: $name" >&2; exit 1
  else
    code=$?
    test "$code" -eq 1
  fi
done
printf 'Inspect the separate reports in %s\n' "$REV_DIR"
go test -count=1 -v ./revisionreview
```

Inspect `intake-diff.json` for the retained question and changes;
`stale-model.json` for an accepted candidate model and its native receipt's
`digest-mismatch`; `candidate-model.json` for the separate current declaration;
`trace-diff.json` for the candidate and its gaps; and all three acceptance
reports for their separate bindings. Decode `intakeReportJSON` before reading the
native receipt; its SHA-256 covers the decoded exact bytes, including its final
newline. Do not compare the exact source-byte hashes with the canonical model or
normalized requirements digests as though they were the same identity.

[`expected.json`](testdata/expected.json) contains hand-stated expectations frozen
before the integration test. The same contributor wrote both: that chronology
is independent of tool output, not proof of independent-person authorship or
review. The test checks the declared relationships and refusals using existing
APIs, without running a scenario. It also shows that an extra newline changes the
suite's exact identity and leaves its validation and replay unrun. It does not
prove intent, coverage, authentication or correctness. To actually build and run
the candidate-specific example, follow the separate tutorial after appropriate
review; this walkthrough neither bypasses that review nor supplies an execution
record. No deployment or worker activation follows.
