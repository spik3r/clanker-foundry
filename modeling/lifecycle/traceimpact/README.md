# Declared trace change review

This offline tool compares two complete snapshots, each containing a lifecycle
model, the existing requirements document and the existing acceptance catalog.
It combines [model change review](../impact/README.md) with the
[requirements trace](../traceability/README.md) to surface requirements whose
own declarations changed or whose declared model links reach changed elements or
transitions. Links and catalog statuses are declarations, not evidence that a
requirement is satisfied, that a case tests it, that a test ran, or that any person
reviewed or approved it. This is not dynamic impact analysis or a standards engine.

A model identity change sets `modelChanged`, irrespective of which requirements
appear in the candidate list. All model-bound reviews need reconsideration when
the model digest changes; an empty candidate list does not establish safety or
unaffected behavior. `reviewRequired` means at least one observed identity changed,
not that a review occurred or an approval exists. Requirements and case IDs are
not guessed across renames.

The catalog reader checks only the fields used by the existing trace. To avoid
missing changes in opaque fields, **any change in catalog bytes**, even whitespace
or unlinked-case content, conservatively flags every requirement with an acceptance
link on either side. It does not claim to identify which case changed or whether
any behavior changed. A catalog's `executable` status is an assertion, not an
execution result. Scenario suites, actual test results and requirements without
links are not inferred from the catalog. Existing gaps are reported for both
snapshots even when they do not become candidates.

## Run

Use Go 1.27.3 and offline module resolution, from `modeling/lifecycle`. The
following compares the existing files with themselves. The Python snippet only
assembles input; it does not change any artifact or compute expected results.

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
outdir="$(mktemp -d "${TMPDIR:-/tmp}/foundry-trace-impact.XXXXXX")"
go build -o "$outdir/modeltraceimpact" ./cmd/modeltraceimpact
python3 - <<'PY' > "$outdir/request.json"
import json, pathlib
read = lambda path: pathlib.Path(path).read_bytes().decode('utf-8')
s = dict(model=read('examples/lifecycle.json'), catalog=read('examples/oracle.json'),
         requirements=read('traceability/requirements.json'))
print(json.dumps(dict(formatVersion='lifecycle-trace-impact-request/v1',
                      before=s, after=s), ensure_ascii=True))
PY
"$outdir/modeltraceimpact" < "$outdir/request.json" > "$outdir/report.json"
printf 'Report: %s\n' "$outdir/report.json"
go test ./traceimpact ./cmd/modeltraceimpact
```

The command accepts no paths or arguments. Its bounded stdin request requires
`formatVersion`, `before` and `after`; each snapshot has exactly `model`, `catalog`
and `requirements`, each containing the literal UTF-8 file contents as a JSON
string. Preserve exact bytes when reading the files. The tool opens no paths and
writes no files; redirection in the example creates the request and report. Input
or error diagnostics may contain private material; the tool emits fixed error
categories but the report includes IDs and declared links, so keep it private.

Both snapshots must pass the existing model and trace validation. A dangling link
is invalid input, not a hidden gap. The implementation accepts at most 1 MiB each
for model and requirements, 4 MiB for the catalog, 512 parsed catalog cases per
snapshot, and 16 MiB for the outer request. A report is capped at 4 MiB and a
comparison at 65,536 reasons. It rejects excess limits rather than truncating.
Exit 0 means a valid report, including a changed one; exit 2 means invalid input,
limits or toolchain/arguments, with no report; exit 3 means an output write failed
or was short, which can leave a partial stream.

## Reading candidates

A candidate's `change` describes its requirement record, not whether its model
behavior changed. `fields` identifies changed requirement text, declared link sets
or derived gap set; an added or removed ID has an empty field list. `beforePresent`
and `afterPresent` distinguish absent requirements from requirements with no gaps.
Reasons are sorted by code, side, target and cause:

- `requirement-added` / `requirement-removed`: a declared ID appeared/disappeared.
- `requirement-changed`: the named requirement field changed.
- `model-element-changed`: the side's direct link names an element changed by the
  existing model-diff comparator; `cause` is its added/removed/modified type.
- `linked-transition-changed`: the side links a transition listed by model-diff's
  direct declared-link projection; `cause` is the existing model-change cause.
- `catalog-source-changed`: the catalog bytes changed and at least one side links
  the requirement to a catalog case. This is deliberately broad.

The before and after graphs are each considered, so removed links and their causes
remain visible. This finite projection does not establish that omitted requirements
or transitions are unaffected. The report is deterministic for identical inputs;
model and requirements semantic-set reordering is normalized, while catalog source
byte changes remain observable. No generated or independently specified expectation
is updated or rebound by this tool.
