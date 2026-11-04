# Requirement-to-scenario trace

`scenariotrace.Check` links native requirement IDs to actual scenario suite case
IDs through an explicit bindings document. It returns a deterministic JSON report
or an error with no partial bytes. The package performs no I/O, invokes no reducer,
and does not infer links from names, catalog entries, transitions or assertions.

## Inputs and identities

Supply a JSON object with exactly these required fields:

- `formatVersion`: `lifecycle-scenario-trace-request/v1`.
- `model`, `requirements`, `catalog`, `suite`, `bindings`: original UTF-8 JSON
  source texts, encoded as strings, not paths or nested objects.
- `scenarioReport`: null when absent, otherwise the original native
  `lifecycle-scenario-report/v1` text encoded as a string.

The bindings text has exactly `formatVersion`, `modelDigest`,
`requirementsSHA256`, `catalogSHA256`, `suiteSHA256` and `links`. Its version is
`lifecycle-scenario-trace-bindings/v1`. Each link has exactly `requirementId` and
`scenarioCaseId`. Digests are lowercase 64-character hex strings. `modelDigest`
is the accepted native canonical model digest; the other assertions hash exact
source bytes. Any stale assertion rejects the result. Empty links are valid
incomplete information. Duplicate pairs and unknown IDs are rejected.

The wrapper and bindings reject missing fields, unexpected fields, case variants,
duplicate keys, disallowed nulls, malformed Unicode and trailing data. Embedded
sources retain native validation rules, including native nullability and opaque
catalog-owned fields. The original native trace report is retained unchanged;
its catalog labels and gaps are not replaced by scenario observations.

The output version is `lifecycle-scenario-trace-report/v1`. `inputs` distinguishes
raw source SHA256 identities from normalized model, requirement and scenario
digests. `traceReport` and `scenarioReport` retain text and raw SHA256; the latter
has null text and hash when absent. `links` and `requirements` sort by requirement
ID, with case references sorted lexically. `cases` preserves suite case/step
order, with sorted requirement references and zero-based report entry positions
(null for absent entries). Both unlinked ID lists are sorted. Empty lists remain
arrays. Reordering binding declarations changes their raw hash even though the
sorted relationships do not change.

## Supplied observations and review

A validation-only native runner decodes the suite and any supplied report. It
checks native result envelopes and binds each report entry to the original case,
step, command, expected literal and input flow. It accepts a complete match or a
valid fail-fast prefix ending at the first mismatch. Matching refusals keep the
previous input, including refusals whose actual task is null. Altered literals,
wrong identities, omitted or extra entries, reordered steps/cases and wrong
completed-case counts are refused.

Each suite case is `not-supplied`, `matched`, `mismatched` or `not-observed`.
Every suite step remains visible, even if the supplied prefix never reached it.
These are observations about supplied text, not a requirement verdict. A caller
can fabricate an internally consistent report. Nothing here authenticates who
ran code, which code ran, who reviewed it or whether expectations are correct.
The output therefore keeps `mode: example-only`, `authority: none`, and execution
and review status `not-established`. There is no aggregate pass rate or approval.

For manual review, examine the original requirement, model, suite literals and
explicit mapping together. Check unlinked IDs and native trace gaps. Separately
establish the provenance and execution context of any supplied report, and assess
whether its expectations test the requirement. This check does not do that work
or authorize production use.

## Bounds and errors

The route caps the wrapper at 32 MiB and encoded output at 64 MiB, including its
final newline. Model, requirements and suite sources are each capped at 1 MiB;
catalog at 4 MiB; bindings at 512 KiB; supplied report at 8 MiB; retained native
trace JSON at 4 MiB; and mappings at 4096 pairs. Native stricter limits still apply.

- `ErrLimit`: an explicit route byte/count bound was exceeded.
- `ErrInvalid`: malformed input, a native structural/internal-limit refusal,
  dangling/duplicate mapping, or report identity/literal/flow inconsistency.
- `ErrBinding`: valid independent bindings digest assertions disagree with inputs.
- `ErrBoundary`: an unexpected native or encoding failure after validation.

Errors contain fixed diagnostics, never input text. Success is deterministic for
the same original bytes; the API does not mutate its input or share output state.

## Local CLI

From `modeling/lifecycle`, with the repository's required Go toolchain:

```sh
go run ./cmd/modelscenariotrace < request.json > report.json
```

The standalone CLI accepts no arguments or paths. It reads bounded stdin, buffers
one complete result and makes one stdout write. Exit codes are 0 for a produced
report, 2 for an input/binding/limit refusal, 4 for a boundary or stdin read failure,
and 3 for a failed or short delivery, including a closed stdout pipe. Inspect the
exit code: a redirected partial file is not a valid report. The command is not a
new `foundry` route and makes no end-to-end execution guarantee.

The frozen acceptance fixture uses invented ParcelReview data and fabricated
observations, not executed evidence. Tests cover its hand-stated relationships,
native trace preservation, strict refusal paths, concurrency, owned buffers,
output bounds and compiled CLI delivery failures.
