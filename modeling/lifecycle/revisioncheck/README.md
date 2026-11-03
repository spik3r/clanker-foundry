# Check two design revisions together

`modelrevisioncheck` composes the existing check-only first-design, intake-change
and declared trace-change tools for two caller-supplied revisions. The
[fixed revision walkthrough](../revisionreview/README.md) remains a useful example;
this command accepts a different model and its original input bytes.

Each side supplies one complete native
[first-design request](../firstdesignreview/README.md), one requirements trace
document and its acceptance catalog. The tool preserves four complete native
encoded reports and their byte hashes: before and after first-design checks,
intake changes and declared trace changes. Existing reports keep their own meaning,
including unresolved questions, model/suite byte comparisons and stale proposal
preflight results. A malformed native input rejects the complete composition.

Each side also compares its intake's observed requirements with the supplied trace
document by exact decoded UTF-8 bytes: `matches`, `mismatch` or `missing`. A newline,
CRLF or Unicode normalization difference is visible. This is a source comparison,
not a comparison of meanings or normalized requirement digests. Requirements can
still be checked as trace declarations when this source comparison mismatches.

A trace catalog is independent of the scenario suite. The tool never invents
case-to-requirement links, repairs stale manifests, rewrites expected results or
runs scenarios. Its top-level statuses are always `executionStatus: not-run` and
`reviewStatus: not-established`, with `mode: example-only` and `authority: none`.
There is no overall ready, approved or authorization result. Review the native
reports and original requirements separately before deciding any further action.

## Request and output

The pure API is `Check(request []byte) ([]byte, error)`. Keep input slices immutable
during calls. It performs no filesystem, subprocess, provider or network I/O and
invokes no reducer or scenario callback. Native validation-only contexts retain
their original labels; they are not evidence of a compiled candidate or execution.

The wrapper has exactly these required fields:

```json
{"formatVersion":"lifecycle-revision-check-request/v1","before":{"firstDesignRequest":"complete native first-design request JSON text","requirements":"requirements trace JSON text","catalog":"catalog JSON text"},"after":{"firstDesignRequest":"complete native first-design request JSON text","requirements":"requirements trace JSON text","catalog":"catalog JSON text"}}
```

These are UTF-8 strings containing complete original JSON documents, never paths,
base64 or parsed nested documents. The illustration is not a runnable request.
Prepare the native first-design request with its original intake and proposal
texts using the existing contract; keep those artifact bytes unchanged. The
supplied proposal model is the model used for that side's trace comparison. An
intake observing different model bytes retains its native mismatch result.

With installed Go 1.27.3, build from `modeling/lifecycle` into an authorized private
directory outside the checkout. Set `REVISION_WORK` to that directory first:

```sh
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
go build -o "$REVISION_WORK/modelrevisioncheck" ./cmd/modelrevisioncheck
"$REVISION_WORK/modelrevisioncheck" < "$REVISION_WORK/request.json" > "$REVISION_WORK/report.json"
```

The CLI accepts no arguments. The caller handles file reads and redirection.
Reports can contain original task content; keep private requests and reports out
of public logs and fixtures. Hashes identify supplied bytes, not authenticated
origin, human review or execution.

The output is `lifecycle-revision-check-report/v1`. The `before` and `after` fields
contain `firstDesignRequestSHA256`, `requirementsSHA256`, `catalogSHA256`,
`requirementsComparison`, `firstDesignReportSHA256` and `firstDesignReportJSON`.
The four report JSON fields are strings: decode them to inspect their native
contents. Their hashes cover the exact decoded UTF-8 report bytes, including final
newlines. The other two pairs are `intakeImpactReportSHA256`/`intakeImpactReportJSON`
and `traceImpactReportSHA256`/`traceImpactReportJSON`. No native report is parsed
and rewritten into an abbreviated decision.

## Limits and failures

The wrapper is capped at 32 MiB and the final encoded report at 64 MiB. Each
first-design request retains its 8 MiB limit; trace requirements are limited to
1 MiB and catalogs to 4 MiB. Native intake/model/proposal limits remain in force.
Assembled native comparison requests also retain their existing limits. A pair
of individually valid native inputs can exceed a combined or output limit; the
composition refuses without returning a partial report. These limits bound this
review tool, not arbitrary execution or host resources.

Unknown, duplicate, missing, null and case-variant fields, malformed Unicode,
trailing JSON and malformed native requests reject. The CLI uses fixed diagnostics
that do not repeat input. Exit 0 means a complete report, including ordinary
mismatches and stale preflight. Exit 2 means input or limit refusal, exit 4 means a
native/encoding boundary failure, and exit 3 means a failed or short report write,
including an actual closed stdout pipe. A redirected file may hold a prefix after
failure; check exit status before using it.

## Checks

The [separate-contributor observations](testdata/observations.json) were frozen
before implementation. They use a nondefault ParcelSeal model and a Unicode
state-label revision while preserving the original literal suite and manifest.
They retain open questions and stale bindings, exercise missing and byte-mismatched
requirements observations, model/suite mismatches and catalog whitespace, and
reject malformed native inputs and dangling links. This is assistant contributor
separation, not independent-person authorship, human approval or execution proof.

Tests compare all four native report byte streams directly, check their hashes,
repeat concurrent calls without input mutation, cover unchanged designs, strict
and limit failures, and exercise complete CLI output and a real closed pipe.

```sh
go test -race -count=1 ./revisioncheck ./cmd/modelrevisioncheck
```

The tool changes no native source policy, carrier format, runtime task, worker
claim, deployment or production setting.
