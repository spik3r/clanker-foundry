# Upstream intake change review

This tool compares two original [intake requests](../intake/README.md). It shows
changes to intent, decisions and rationales, unanswered questions, artifact
assertions, observed source identities and asserted dependencies. A reviewer can
see an intent change even when the model and tests have not changed, or an
unchanged unresolved question even when there are no differences.

It complements [model change review](../impact/README.md) and
[declared trace change review](../traceimpact/README.md). Those tools inspect
accepted model meaning and requirement/catalog links. This one treats artifacts
as opaque bytes. It does not validate a model, inspect a suite, run tests, infer
impact, establish approval or prepare an artifact pack. It does not use the held
intake/model bridge or review-pack stack.

## Run

From `modeling/lifecycle`, use Go 1.27.3 with offline modules. The following uses
the invented fixtures and creates only temporary input/output files:

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
outdir="$(mktemp -d "${TMPDIR:-/tmp}/foundry-intake-diff.XXXXXX")"
go build -o "$outdir/modelintakediff" ./cmd/modelintakediff
python3 - <<'PY' > "$outdir/request.json"
import json, pathlib
read = lambda p: pathlib.Path(p).read_bytes().decode('utf-8')
print(json.dumps({
    'formatVersion': 'lifecycle-intake-impact-request/v1',
    'before': read('intakeimpact/testdata/before.json'),
    'after': read('intakeimpact/testdata/after.json'),
}, ensure_ascii=True))
PY
"$outdir/modelintakediff" < "$outdir/request.json" > "$outdir/report.json"
printf 'Report: %s\n' "$outdir/report.json"
go test ./intakeimpact ./cmd/modelintakediff
```

The command accepts no arguments or paths. Its strict input has exactly
`formatVersion`, `before` and `after`; the two latter fields are strings containing
the exact UTF-8 original intake request bytes, including final newlines, not
prebuilt reports or embedded objects. Each original request is validated by the
existing intake tool. Invalid either-side input produces no comparison.

## Read the report

Each `before`/`after` receipt preserves the **exact native intake report** as
`reportJSON`, including its final LF. `reportSha256` hashes that decoded string,
not the outer JSON carrier's escape characters. `requestSha256` is the native
hash of the original intake request. The receipts retain intent, author labels,
questions, artifact and dependency statuses, but not observed artifact contents.
They may contain private information; keep requests and reports appropriately
private. Native input errors are reduced to fixed diagnostic categories.

`contentSha256` hashes the compared content, not artifact semantics. Its bytes are
a fixed-order JSON object with `formatVersion: lifecycle-intake-content/v1`, then
`intent`, `decisions`, `questions`, `bindings`, `links` from the native normalized
report. Encoding disables HTML escaping and omits the final LF. Decision/question
IDs are sorted by intake; binding and dependency order is intake's fixed order.
Content identity excludes the request hash. Thus:

- Request formatting, object-key order and decision/question list order can change
  `requestSourceChanged` without changing `contentChanged`.
- Any change to observed artifact bytes changes their raw hash, even if their
  semantic model or requirements identity remains the same elsewhere.
- Declared author changes are visible assertions, not authenticated identity changes.
- Missing, null, empty and mismatched observations remain distinct native statuses.

`changes` is sorted by kind and ID. Intent has field `text`; decisions have
`text`/`rationale`; questions have `text`. Added/removed decisions and questions
have empty fields. A question removed from the list is **removed**, not proven
resolved. Moving an ID between lists is a removal and an addition.

Artifact roles always exist and compare `declaredSha256`, `declaredAuthor`,
`observedSha256`, `status`. Dependencies use IDs `model->requirements`,
`scenarioSuite->requirements`, `scenarioSuite->model`, and compare
`assertedSha256`/`status`. These records are `modified`, with sorted changed fields.
The native receipts contain the before/after values.

`beforeGaps` and `afterGaps` always carry every unresolved question, artifact
binding not `matches-declaration`, and dependency not `matches-declaration`, sorted
by kind and ID. Unchanged gaps do not vanish from a no-op comparison. Conversely,
a new matching hash does not prove validity, acceptance or review. No aggregate
ready, qualified, safe or approved flag exists. The report always states
`artifactValidation: not-run`, `executionStatus: not-run` and
`reviewStatus: not-established`, with mode `example-only` and authority `none`.

## Continue the pipeline without rebinding

Use the original observed artifacts with the existing domain tools, retaining the
original declarations and frozen expectations. The integration tests demonstrate:

| Change | Intake comparison | Existing trace/model comparison | Existing acceptance authoring |
|---|---|---|---|
| Intent only | Changed intent, question still unresolved | No artifact identity change | Original frozen examples still match; this does not validate the new intent |
| Model formatting | Raw model mismatch | Canonical model unchanged | Original model binding remains current |
| Requirements formatting | Raw requirements mismatch | Normalized requirements unchanged | Unchanged suite/model bindings remain current; no requirements check is implied |
| Suite formatting | Raw suite mismatch | Trace does not consume the scenario suite | Exact suite binding stale; validation/replay not run |
| Model label edit | Raw model mismatch | Canonical model changed, review required | Original model/runner bindings stale; replay not run |

This package does not select a downstream action or execute those checks itself.
Its tests explicitly invoke the already merged tools. A current acceptance binding
and matched supplied literals cannot prove that a changed intent is satisfied.
No declared digest or frozen expected result is refreshed to make a check pass.

## Limits and evidence

Outer input is at most 64 MiB; each intake retains its 4 MiB request and 1 MiB
observation limits. Reports are buffered and capped at 16 MiB. Invalid input or
an excess limit refuses rather than truncating. Exit 0 means a report was emitted,
including changed or incomplete reports; exit 2 means input, read, toolchain or
report refusal with empty stdout; exit 3 means a failed or short output write,
which may leave partial bytes. SIGPIPE is ignored so a closed pipe returns 3.

`Compare(before, after []byte)` and `Execute(raw []byte)` are pure APIs. Callers
must not mutate input concurrently. Returned data owns its contents; there are no
paths, network calls, clocks, global caches or artifact writes. The CLI checks the
pinned Go version; the APIs do not install or select a toolchain.

A separate assistant authored the invented requests, literal expected changes and
gaps, and external tests before the comparator was implemented. This separates
expectations from implementation within the session, not independence between
people or an independent oracle for intake itself. Tests compare preserved native
receipts to the existing intake implementation. Pipeline checks consume already
frozen acceptance literals unchanged; they do not create new acceptance evidence
for real user requirements. See the [task claim](../../../tasks/CF-intake-impact-v1-review.md).
