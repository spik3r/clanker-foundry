# Compiler profile acceptance

This test-only package exercises actual Go reducers generated from four invented
`lifecycle/v1` models. It complements the profile catalog's type checks, baseline
acceptance oracle and consumer integration. It changes no production package or
existing fixture and adds no evaluator, runtime API or standard-conformance claim.

Two independent fixture authors wrote the models, requirements and public-API
oracle tests. A separate reviewer checked their behavior, inventory and independent
Python JSON/SHA-256 calculations. The [frozen inventory](testdata/inventory.json)
was committed at `4d34bdf` before harness implementation or new-model execution.
Its SHA-256 is `e96013d6d9d5d2bf603d3dd473075de47a5d89d4a3de633dc5b483ff767e663c`.
This records independent assistant authorship/review, not authenticated human
approval or a proof of correctness. No actual output supplied an expected value.

## Scope

| Family | Independent cases | Main distinctions |
| --- | ---: | --- |
| integers | 25 | uint64 eq/ne beyond 2^53 and 2^32, optional review revision, zero and MaxUint64, guard/overflow precedence |
| presence | 72 | absent/empty strings, optional eq/ne, Unicode trimming, list length, sixteen-operand all/any and nesting |
| dispatch | 11 | custom state/trigger pairs, self-loop, two terminals, non-start owner assignment, refusal ordering and atomicity |
| textdigest | 46 | all fifteen reference bindings, exact Unicode, invalid-input distinction, eight independent digest vectors |

The 154 leaf cases belong to eleven independent top-level tests. Expectations
compare complete Results and Tasks, including literal model identities, mode,
authority, transition IDs, nil/empty values and refusal order. Nonempty slices and
owner pointers have storage-mutation checks. This finite set does not cover every
possible alias pattern, such as empty slices with spare capacity, or exhaust all
models, strings, operator combinations and input values. See the
[written requirements](testdata/requirements.md) and each family's requirements.

## Harness and controls

The harness checks all frozen input hashes before decoding/generation. The unchanged
compiler emits a package in a disposable standard-library-only Go module. Exact
oracle bytes and compiler-derived structural tests are copied beside it; derived
tests are additional checks, never the independent behavior oracle. JSON test
events must show every frozen test/subtest running and passing exactly once, no
skips and no extra profile tests. Separate negative controls test that event gate.

Fifteen controls mutate only disposable generated code: swapped uint64 comparison,
float/uint32 narrowing, missing optional uint/string presence, early overflow,
missing Unicode trim, all/any swaps, assignment hardcoded to `start`, reversed
refusals, shallow Task copying, a misbound reference and numeric digest revision.
Each mutant must compile and then fail its specified named oracle with an ordinary
test exit and semantic mismatch marker. Syntax/setup errors, timeouts, missing
tests and unrelated failures do not count. All positive cases rerun after exact
generated-byte restoration. The frozen files are never edited by tests.

An independent structural check compares every generated transition's statement
sequence with the frozen ordered effects and destination. A swapped-effect control
must compile, fail that structural check, and still pass the behavioral oracle:
those effects commute. This makes the distinction explicit rather than claiming
that final Task values prove execution order. These checks intentionally depend on
the current simple generated-code shape; a legitimate lowering change needs review,
not a silent expected-value refresh. No vendor or UML tooling is involved.

## Run

From `modeling/lifecycle`, using actual Go 1.27.3:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 -v ./compilerprofileacceptance
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 ./...
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go vet ./...
```

The harness invokes the installed toolchain by absolute path, disables ambient
Go workspace/config/flags, downloads and VCS stamping, and pins target OS/architecture
to the running toolchain. Each subprocess has a three-minute limit and an 8 MiB
diagnostic cap. Existing cache/temp locations must be writable; set `GOCACHE` and
`TMPDIR` to authorized locations when necessary. The fixtures are reviewed trusted
repository code: this runner is not a sandbox or loader for caller-provided code.
Existing CI's full module test command picks up this package without workflow edits.
