# Offline lifecycle scenarios

Run hand-authored machine-readable examples against the generated lifecycle
reducer. A passing comparison can contain an expected refusal: `matched` and the
transition's `ok` are different facts.

This is a local follow-up to the [lifecycle demo](../README.md), based on its exact
reviewed commit `ff661f39c7e176450e175213565217958acab5dd`. It does not change PR11,
its compiler, model, generated artifacts or independent oracle. The follow-up
still depends on PR11's review. See the [task](../../../tasks/CF-model-scenarios-v1-offline-runner.md).

## Run it

With actual Go **1.27.3**, from `modeling/lifecycle`:

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
go run ./cmd/modelplay < scenarios/testdata/lifecycle-v1.json
go test -race ./scenarios ./scenarioacceptance ./cmd/modelplay
go test -race ./...
go vet ./...
go run ./cmd/modelc check
```

`modelplay` has no arguments or file-output option. It reads bounded stdin and
writes one complete buffered JSON report to stdout. The caller may redirect that
stream. No network request, package installation, worker start or scheduler action
is part of normal execution. `go run` itself uses the installed compiler/cache;
the demo does not download a toolchain automatically.

The [frozen fixture](testdata/lifecycle-v1.json) has four cases and 13 steps:

- Prepare, assign the named owner, submit and approve.
- Expected blank-actor refusal, expected invalid command, then a valid start.
- Request changes, resubmit, refuse stale review and full-command replay, then
  accept a freshly bound approval.
- Refuse a review whose content digest predates a same-revision criteria change.

Its literal expectations and review digests were independently authored and frozen
before the runner/comparison implementation. The runner never fills or refreshes
those expectations from its own results. This is a small new executable example,
not another copy of the original case catalog.

### Exit codes

| Code | Meaning | stdout |
| --- | --- | --- |
| 0 | All scenario comparisons matched | Complete report |
| 1 | First expectation mismatch; remaining steps/cases did not execute | Complete report through mismatch |
| 2 | Arguments, toolchain, input, binding or report bound refused | Empty |
| 3 | Output write failed or was short | May be incomplete; never treat as a successful report |

Diagnostics are fixed categories, not raw input or upstream I/O error details.
Reports intentionally contain supplied task, review and expectation data. Keep
stdout private when using private fixtures; do not publish it as real review or
ownership evidence.

## Typed profile and fixture adapter

`Suite`, `Case`, `Step` and `Expected` in [types.go](types.go) are typed data.
[decode.go](decode.go) holds separate JSON DTOs. The first wire format is
`lifecycle-scenarios-json/v1`; scenario semantics are `lifecycle-scenarios/v1`.
The reducer profile must be `lifecycle/v1`, and these package-level APIs require
the suite's model digest to equal the built-in compiled reducer's constant.
They remain unchanged; the explicit Go-composition API below is separate.

```go
Decode([]byte) (Suite, error)
Encode(Suite) ([]byte, error)
Validate(Suite) error
Digest(Suite) (string, error)
Run(Suite) (Report, error)
Execute([]byte) (Report, error)
EncodeReport(Report) ([]byte, error)
```

### Explicit model-bound Go composition

An external application can use the same fixture pipeline with its separately
generated lifecycle/v1 reducer without copying Foundry or replacing its example
package. See the [external consumer proof](../consumerintegration/README.md).
The existing compiler and all default generated artifacts remain unchanged.

```go
NewRunner(Binding, func(lifecycle.Task, lifecycle.Command) lifecycle.Result) (Runner, error)
Runner.ModelBinding() (Binding, error)
```

`Binding` contains `ProfileVersion` and `ModelDigest`. Construct it in reviewed Go
code from the custom generated package's constants and provide a field-only typed
adapter to that package's `Transition`. The existing generated Task/Command/Review/
Result types are only the transport profile at this boundary; the adapter must
preserve all fields, nil/empty distinctions, uint64 revisions and refusal order.
It must not rewrite actual model identity or derive expected results. This is not
a production command transport or an interpreter for another semantic profile.

The value `Runner` snapshots private configuration and exposes the same
`Validate`, `Decode`, `Encode`, `Digest`, `Run`, `Execute` and `EncodeReport`
signatures as the package-level functions. Every method uses that same binding;
the zero runner and nil callback refuse, never silently select the default.
There is no JSON-selected callback, registry, dynamic loader or identity setter.
Do not construct a binding by copying a request's expected model digest.

A supplied binding is a **declaration**, not proof that the callback is the named
generated reducer. The callback is trusted application code. It can perform I/O,
mutate a closure, allocate, block or panic; panics propagate. Runner neither
sandboxes nor cancels it and cannot bound the callback's allocation or roll back
its effects. Callers must not concurrently mutate input or callback-owned returned
data. Deterministic replay requires a deterministic callback and stable inputs.

The complete suite is checked and copied before the first call; raw source identity
is recorded before invocation. A callback retaining the caller's original suite or
raw bytes therefore cannot rewrite later comparisons or their source identity.
A stale or malformed suite
invokes no callback. After each call, the opt-in path bounds the returned structure
**before cloning it**, and checks actual model digest, mode and authority without
repair. Invalid envelopes, shapes or result bounds return `ErrProviderProtocol`
and no report; execution has begun and no subsequent step is called. This error
does not mean not-run or rollback. Profile is checked at construction because a
Result has no profile field. Structurally valid frozen-expectation mismatches still
produce an ordinary fail-fast mismatch report. Legacy `Run` retains its existing
envelope-mismatch behavior rather than adopting this new error policy.

Aggregate report overflow can also return `ErrReportLimit` after multiple calls.
An error or absent report alone never proves zero execution. Only the documented
preflight-refusal paths and acceptance check-only/stale paths promise zero calls.

`Runner.EncodeReport` uses `lifecycle-bound-scenario-report/v1` with
`identitySource: caller-supplied-go-binding`. It rechecks the configured identity,
actual envelopes, comparisons, continuity, counts and output bounds. The older
package encoder and CLI keep their original format and byte behavior. Neither
report authenticates code or proves execution; callers can fabricate typed data.
Canonical suite identity and original source-byte identity remain unchanged.

Held copied-source consumers must explicitly requalify their complete closure,
including the new `runner.go` dependency. This change does not refresh those pins,
adopt a real consumer repository, or establish task/reviewer/operating authority.

The complete fixture fields are shown in the small JSON file. Every key is
required, even a nullable key. No unknown/case-variant keys, duplicate decoded
keys, trailing data, invalid UTF-8 or unpaired surrogate escapes are accepted.
Unsigned 64-bit revisions are canonical decimal **strings**: `0` or a nonzero
ASCII digit followed by ASCII digits, at most `18446744073709551615`. Numeric JSON,
signs, whitespace, leading zeros, fractions, exponents and overflow refuse.

`owner`, `review` and expected `task` distinguish null from present values.
Acceptance distinguishes null from an empty list. This differs from the existing
**task-content** digest, which deliberately treats nil acceptance as empty; equal
content digests do not establish equal scenario representations. Each case starts
from its own initial task. Typed zero values are actual values, not evidence that
a field was supplied in JSON.

This is a **scenario fixture adapter**, not a production Task/Command transport.
It permits well-typed values that intentionally exercise the reducer's
`INVALID_INPUT`; it does not duplicate the reducer's business-input rules.
Expected OK structurally requires a nonnull task, nonblank transition ID and an
explicit empty refusal list. Expected not-OK requires a nonempty refusal list.
An incorrect but structurally valid expectation is a comparison failure, not
something to repair automatically.

## Execution and reports

The entire suite is validated before the first evaluation, including its last
case/step, expected shapes and model binding. Cases and steps execute in supplied
order. Every actual result field is compared with the expected result: fixed
example-only/authority-none/model envelope plus literal OK, transition ID, full
task and refusal list. Refusal order and null/empty distinctions are exact.

Only a matching successful transition advances current state. A matching refusal
retains current state, including `INVALID_INPUT` with a null returned task.
The first mismatch stops the whole suite. Reports own their copies; mutating a
report cannot alter the supplied suite, another entry or a subsequent run.
Concurrent evaluation assumes callers do not concurrently mutate input data.

The JSON report contains ordered entries, original inputs/commands, expected and
observed results, comparisons and counts. It has no timestamp or random ID.
`EncodeReport` rechecks internal comparison flags/differences, state continuity and counts against
ordered case groups before emitting bytes. That check does **not** authenticate
execution, prove that a caller supplied every case, or bind a binary to a trusted
source checkout. Supplied actor/reviewer labels and hashes remain consistency
checks only. There is no real scheduling, approval or ownership authority.

## Two distinct scenario identities

`Digest` hashes the complete typed suite, including literal expectations and
ordered case/step arrays, in fixed-order compact UTF-8 JSON. The semantic field
order is:

- Root: `scenarioVersion`, `profileVersion`, `modelDigest`, `id`, `cases`.
- Case: `id`, `initial`, `steps`.
- Step: `id`, `command`, `expected`.
- Task: `id`, `state`, `revision`, `outcome`, `acceptance`, `owner`.
- Command: `type`, `actor`, `expectedRevision`, `review`.
- Review: `taskId`, `taskRevision`, `modelDigest`, `taskDigest`, `reviewer`, `decision`.
- Expected: `ok`, `transitionId`, `task`, `refusals`.

These names/order are a **semantic-profile v1 contract**, not incidental DTO
layout. A future format adapter must not silently change them. Encoding uses no
HTML escaping, no Unicode normalization and no trailing newline in hashed bytes;
Go JSON's control/U+2028/U+2029 escapes remain deterministic. Null/empty and exact
string bytes remain distinct. Format version, original source-byte hash and report
are excluded. `Encode` adds the JSON format label and one final newline.

`Run` has typed data only, so its original `sourceDigest` is null. `Execute` also
hashes the exact original JSON bytes. A formatting-only edit changes this source
hash but not the scenario digest. Neither hash proves authenticity. Independent
fixed vectors lock down canonical fields and the frozen fixture's digest.

## Bounds and failure behavior

| Resource | Limit |
| --- | --- |
| Source and encoded suite | 1 MiB |
| JSON depth | 64 |
| Cases | 1–16 |
| Steps per case / whole suite | 1–32 / 128 |
| Globally unique suite/case/step IDs | ASCII identifier, 48 characters |
| UTF-8 scalar | 8192 bytes |
| Acceptance / refusal lists | 64 entries each |
| Each refusal string | 256 bytes |
| Encoded report | 8 MiB |

Typed entrypoints enforce the same profile and representation bounds. Output
accounting checks each bounded entry before retaining it; overflow stops further
evaluation and returns no report. Encoding buffers the final value before stdout;
an actual write failure still produces a nonzero exit and must not be accepted.

Tests include strict decoding, whole-suite call-count spies, stop-on-mismatch,
revision/content/rework cases, null/empty distinctions, deep nonmutation,
concurrent pure replay, deterministic separate-process output, error/short writes,
report limits and altered-result/always-refuse mutations. Scratch copies of the
actual generated reducer are also mutated; the frozen independent acceptance
must fail behaviorally rather than merely fail to compile. The original demo's
independent oracle and model-derived tests remain separate checks.

No UML, BPMN, SysML, UTP or fUML conformance is claimed. This sidecar is an explicit
small example toward executable model-driven acceptance, not a standards engine.
