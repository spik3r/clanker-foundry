# Review literal scenario-suite edits

`scenariodiff` shows what a proposed change alters in two supplied scenario suites,
including their separate model bindings, input values, literal expectations and
case/step order. This is **declaration comparison, not execution evidence**. A suite
can contain fabricated or incorrect expectations. No reducer runs, and an unchanged
report establishes neither correctness, compatibility, coverage nor approval.

Use it before accepting edits to scenario expectations alongside model changes.
Keep the original suites for line-by-line review: the report lists changed fields
and identities, not a second copy of literal task/command/expectation payloads.
It does not repair bindings, refresh expectations, infer renames or generate tests.
It does not interpret model edits; use the separate model-review tools for that.

## Typed interface

```go
report, err := scenariodiff.Compare(beforeModel, beforeSuite, afterModel, afterSuite)
```

Each model is normalized and validated by the existing lifecycle/v1 model APIs,
required to fit the existing JSON-v1 adapter, and given its canonical digest.
Each entire suite must validate against that side's actual supplied model identity.
A stale binding rejects the request, even if every literal input is unchanged.
The existing scenario schema deliberately permits invalid business inputs so a
fixture can describe expected refusals. This tool does not check whether expected
transition IDs or refusal strings exist in the model, let alone whether they occur.

The existing scenario API exposes bound validation/decoding/digest methods through
Runner. This package creates only a **private validation adapter** with a panic
sentinel; it never calls Run/Execute or a reducer, and never exports that adapter.
Its derived binding is only input consistency, not a generated-code identity or a
binding suitable for execution. Callers cannot supply callbacks to this package.
Existing model and scenario packages are unchanged.

The report has version `lifecycle-scenario-diff/v1`, profile `lifecycle/v1`,
`mode=example-only`, and `authority=none`. It contains:

- Before/after canonical model and suite digests. `modelChanged` and `suiteChanged`
  compare those identities. `reviewRequired` is their OR, merely a change flag,
  not a review policy or approval result. Model labels/text are part of model
  identity; no claim of behavioral equivalence is made.
- `suiteFields`: changed `id` and/or `modelDigest`, lexically sorted. Version and
  profile are fixed by validation.
- `caseOrder`: complete before/after ordered case IDs and their change flag.
  Adding or removing a case also changes these sequences.
- `cases`: added/removed/modified cases, sorted by ID. Modified fields are `initial`
  (the complete initial Task) and/or `steps` (ordered full Step sequence/content).
  Each entry has complete beforeSteps/afterSteps ID arrays; absent sides are empty.
  Mere case reordering produces no case entries.
- `steps`: added/removed/modified steps, sorted by (case ID, step ID), including
  children of added/removed cases. Pairing is local to a case; moving an ID across
  cases is removal plus addition. Pure step reordering changes its case but does
  not produce an unchanged step entry.

Added/removed entries have empty `fields`. Modified step fields are sorted:
`command.actor`, `command.expectedRevision`, `command.review`, `command.type`,
`expected.ok`, `expected.refusals`, `expected.task`, `expected.transitionId`.
Initial/expected Tasks and command Reviews are atomic, including pointer presence.
Ordered arrays, nil versus empty task acceptance and exact uint64 values matter.
Nil expected refusals are rejected by the existing schema, not normalized away.
All output arrays are non-null and owned. Inputs must remain immutable during a
call; concurrent read-only calls are supported. Errors return no partial report.
Typed reports have null `sources` because no original byte stream was observed.

## Strict JSON and stdin command

`Execute([]byte)` accepts exactly this wrapper, with all fields required/non-null:

```json
{"formatVersion":"lifecycle-scenario-diff-json/v1","beforeModel":{},"beforeSuite":{},"afterModel":{},"afterSuite":{}}
```

Replace placeholders with complete existing JSON-v1 model and scenario-suite
documents. The wrapper is bounded to 5 MiB/depth66; each embedded model/suite keeps
its existing 1 MiB limit and schema. Duplicate, unknown, case-variant and missing
keys, unsupported formats, malformed Unicode and trailing data are rejected.
The complete report is buffered and capped at 1 MiB. Wrapper/output byte-limit
failures return ErrLimit; structural/depth and other invalid input return ErrInvalid.
All errors are fixed messages
without echoed input; no public arbitrary-report encoder is provided.

Wire `sources` hashes exact request bytes and each embedded JSON token's bytes,
excluding whitespace outside that token. These SHA-256 values are separate from
semantic digests: formatting can change source hashes without changing comparison
results. Hashes establish consistency, never origin or provenance. Fixed-order
output has HTML escaping disabled and one final LF. Identical bytes produce
identical output; the operation has no time, random or environment dependencies.

From the module directory with Go 1.27.3 installed, a no-change invented example:

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
python3 -c 'import json; m=json.load(open("examples/lifecycle.json")); s=json.load(open("scenarios/testdata/lifecycle-v1.json")); print(json.dumps({"formatVersion":"lifecycle-scenario-diff-json/v1","beforeModel":m,"beforeSuite":s,"afterModel":m,"afterSuite":s}))' | go run ./cmd/modelscenariodiff
```

For a real review, supply separately authored before/after model and suite files;
do not change expected values or bindings merely to make this command succeed.
The example shell explicitly reads local files; the CLI itself accepts no paths
or arguments and has no file/network/subprocess/provider operations. It reads
bounded stdin and emits one complete report, exit 0 even when changes exist.
Argument/read/validation rejection exits 2; failed/short output, including closed
pipes, exits 3. Buffered computation prevents partial reports on validation error,
but cannot make a failing stream write atomic. IDs and digests may be sensitive;
comparison is not de-identification, and original suites should remain private.
