# Optional adapter checks for consumer tests

`adapterconformance.Check` extracts the field and ownership checks from the
[external consumer proof](../consumerintegration/README.md) into one reusable
test helper. It is not a new scenario suite, runner, production adapter protocol
or admission check. It never calls a reducer or chooses a model.

Use it from a consumer's `_test.go` file with that consumer's actual four
field-only conversion functions:

```go
func TestAdapterConformance(t *testing.T) {
    err := adapterconformance.Check(adapterconformance.Adapter[
        custom.Task, custom.Command, custom.Result,
    ]{
        ToTask: toTask, FromTask: fromTask,
        ToCommand: toCommand, FromResult: fromResult,
    })
    if err != nil {
        t.Fatal(err)
    }
}
```

Here `custom` is the consumer's generated package. The transport side uses
Foundry's existing `generated.Task`, `Command` and `Result` types. The helper
does not supply the conversions: consumers still own their field-only mapping
and independently authored scenario expectations. The source file is an ordinary
importable Go package so external tests can use it; test-only use is a convention,
not a sandbox or a build constraint.

## What the finite probes check

Before any callback, the helper checks the complete ordered field inventory of
both type families, including nested Review and Result.Task. Missing, extra,
embedded, unexported or wrongly typed fields fail. State may be a named string;
other scalar, pointer and slice shapes must match the generated transport shape.
Result.Task must point to the supplied custom Task type. Struct tags do not affect
these in-memory conversions. All four callbacks must be non-nil.

Each direction is then compared directly against fresh named literal fields.
An inverse callback never constructs the expected value: an ID/Outcome swap in
both directions can cancel in a round trip and must still fail here. Probes cover:

- Every Task, Command, Review and Result field, including the actual result mode,
  authority, model digest, success flag and transition ID, without identity repair.
- Zero and maximum uint64 values; different command/task review revisions.
- Nil, empty and populated lists and owners, absent/present reviews and tasks,
  and both success flags. Invented sentinel strings need not be valid model data:
  the adapter must copy them, not apply business rules or rewrite them.
- Refusal order and duplicate entries; nil and empty are distinct.
- Unchanged callback input values **and storage**: pre-call snapshots cover
  pointer/slice identities, length/capacity and the contents of spare capacity.
  Equal-value replacement of caller-visible nested storage still fails.
  Detached task/review pointers, nested owner/list storage and result refusals.
  Empty lists have spare capacity to expose aliases;
  overlap checks include shifted empty subslices, not only equal starting pointers.
  Fresh probe storage is used for each call.

`Check` returns nil on success or the first error. Diagnostics identify the failed
direction or type shape but are not a versioned serialization contract. Callback
panics propagate. Callbacks remain trusted local test code: they may block,
allocate, have side effects or recognize these finite probes. The helper does not
sandbox or cancel them and cannot prove arbitrary-input correctness, honest
model binding, source provenance, approval or deployment. Do not use a nil error
as an authority or readiness certificate.

## Independent expectations and real consumer coverage

The [helper tests](conformance_test.go) were authored by a separate assistant
before implementation. They use their own explicitly named custom structs and
copying adapter, corrupt or drop each named field, and check swaps, narrowing,
nil/empty changes, ordered duplicate refusals, nested aliases and preflight.
That separation is not independent-person review or proof of real requirements.
Independent review later exposed spare-capacity input mutations that value-only
comparison missed. An independently authored addendum was frozen before the
storage-snapshot correction; the original pre-logic cases remain unchanged.

The existing consumer proof adopts this helper for **both** independently generated
models. It still runs both frozen fourteen-step scenario suites and its original
generated-code behavioral mutants. Seven disposable adapter mutations per model
exercise a coupled field swap, command-field drop and five alias forms. A compile
failure does not count: the conformance test must run and emit its failure marker.
After mutation, each restored adapter must actually pass conformance again.
No frozen model, suite, manifest, requirements or expected-inventory bytes change.

From `modeling/lifecycle` with installed Go 1.27.3:

```sh
GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off \
  go test -race -count=1 ./adapterconformance ./consumerintegration
```

The external-module harness retains its existing offline settings and finite
subprocess waits. It does not guarantee descendant termination or callback
cancellation. This extraction does not qualify or modify any private consumer,
compiler, runner, CLI, workflow or held review artifact.
