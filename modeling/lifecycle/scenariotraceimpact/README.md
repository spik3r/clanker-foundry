# Review requirement-to-scenario changes

`scenariotraceimpact.Check` compares two complete original
[scenario-trace requests](../scenariotrace/README.md). It joins explicit
requirement-to-case links with [scenario declaration changes](../scenariodiff/README.md)
and [declared trace impact](../traceimpact/README.md). It performs no I/O and
executes no reducer, callback, command or provider.

A changed expected result in case `seal` produces candidates for requirements
explicitly linked to that case on each side. Removing a link is reported even if
the case is unchanged. Changing an unlinked case does not invent a requirement
link. The full native reports remain available for inspecting declaration fields,
steps, model links and trace gaps.

## Input and command

The strict wrapper has exactly three required, nonnull fields:

```json
{
  "formatVersion": "lifecycle-scenario-trace-impact-request/v1",
  "before": "<original complete scenario-trace request JSON text>",
  "after": "<original complete scenario-trace request JSON text>"
}
```

The placeholders above stand for JSON strings containing the original request
bytes. For example, from the module directory, two existing requests can be
wrapped without rewriting their contents:

```sh
jq -n --rawfile before before-trace-request.json \
  --rawfile after after-trace-request.json \
  '{formatVersion:"lifecycle-scenario-trace-impact-request/v1",before:$before,after:$after}' \
  > trace-impact-request.json
go run ./cmd/modelscenariotraceimpact < trace-impact-request.json
```

The compiled command accepts stdin only, reads at most 64 MiB plus one byte, and
buffers its report before one stdout write. Exit 0 means a report was produced;
2 means argument, read, input or limit refusal; 3 means failed or short output
delivery; 4 means a boundary failure. Diagnostics are fixed and omit source data.
`go run` may wrap a nonzero program exit; use a compiled binary when consuming
these exit codes.

Both snapshots must pass native trace validation, including exact source bindings
and consistency checks for any supplied scenario report. The comparison rejects
stale sources and never repairs bindings. Duplicate, unknown, case-variant or
missing wrapper fields, malformed Unicode, nulls and trailing data are refused.

## Report semantics

The `lifecycle-scenario-trace-impact-report/v1` report always says
`mode: example-only`, `authority: none`, `executionStatus: not-established`, and
`reviewStatus: not-established`. Its candidates describe supplied changes; they
do not establish behavioral impact, requirement satisfaction or approval.

Each candidate names a requirement, its presence and sorted case IDs on each
side, and deduplicated reasons sorted by `(code, side, scenarioCaseId)`:

| Reason | Meaning |
| --- | --- |
| `native-trace-impact` | The native declaration comparison named the requirement. Its precise reasons are retained. |
| `scenario-link-added` / `scenario-link-removed` | An explicit requirement/case pair exists only on that side. |
| `scenario-case-added` / `scenario-case-removed` / `scenario-case-modified` | A native changed-case ID is explicitly linked on the named side. |
| `scenario-case-order-changed` | Global case order changed; conservatively includes every requirement linked on either side. Case additions/removals can also change this order. |
| `scenario-suite-fields-changed` | Suite `id` or `modelDigest` changed, and the requirement has a link on either side. |
| `scenario-observation-changed` | A case present on both sides has a changed validated status or ordered `(stepId, entryPosition)` list. The requirement links to it on either side. |

Observation changes describe supplied report structure. A step rename or reorder
can change that structure even when both reports are absent: each step still has
a null entry position. Link changes alone do not change an observation. Different
actual payloads with identical statuses and positions remain visible in retained
reports and source hashes without this narrow reason. Even a fully consistent
supplied report can be fabricated.

`requestSourceChanged` and `reviewRequired` mean only that the original two
request texts differ bytewise. `sourceChanges` names changed decoded source texts
among `bindings`, `catalog`, `model`, `requirements`, `scenarioReport`, and `suite`.
Formatting alone may set these flags while candidates remain empty. Outer wrapper
formatting alone changes only its `requestSHA256`. No absent candidate is a claim
that behavior is unaffected, and no coverage or pass-rate total is calculated.

`before`, `after`, `scenarioDiff`, and `traceImpact` each retain exact native
`reportJSON`, `reportSHA256`, and `requestSHA256`. Snapshot request hashes identify
the original texts; comparison request hashes identify the constructed native
wrappers. Scenario-diff model/suite source hashes preserve all bytes inside each
JSON value token, trimming only surrounding JSON whitespace. Trace source hashes
identify full decoded texts, including that whitespace. Native model, suite,
requirement and catalog identities must agree within each side.

## Limits and verification

The outer request limit is 64 MiB, each native snapshot is limited to 32 MiB, each
retained trace report to 8 MiB, and total encoded output to 128 MiB. Native source,
comparison-input and report limits also apply. Native trace impact admits at most
512 catalog cases. Valid individual trace snapshots can therefore be refused as a
comparison. `ErrInvalid`, `ErrLimit` and `ErrBoundary` return nil output.

Independent acceptance cases were committed before production and use the
unchanged invented ParcelReview vectors with hand-stated expected changes. Local
tests also cover concurrent read-only calls, independent returned buffers,
comparison admission limits, complete CLI delivery and a closed output pipe.
These checks do not establish general correctness or execution of supplied cases.
