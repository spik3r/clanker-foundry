# JSON boolean validator generator

This isolated Go module generates standalone Go and TypeScript predicates from
the bounded [JSON boolean profile](PROFILE.md). It is separate from the
`lifecycle/v1` module and the root shadow npm package. It has no Go module or npm
runtime dependencies beyond the language runtimes.

The generator takes an explicit model and a new, absent output directory:

```sh
GOTOOLCHAIN=local GOPROXY=off go run ./cmd/jsonvalidator \
  -model testdata/sample/model.json -out /tmp/new-validator-output
```

Use a fresh destination. The command writes `validator.go` and `validator.ts`.
Go exposes package `validator`, `Validate(value any) bool` and `ModelSHA256`.
TypeScript exposes `validate(value: unknown): boolean` and `MODEL_SHA256`.
The hash identifies the exact input model bytes, not a canonical semantic form.
The generated code does not read files, fetch data or require this generator at
runtime. A consumer can check in the generated TypeScript without adding a
browser package dependency.

The predicate operates on already-parsed ordinary JSON values. It does not parse
HTTP responses, map application rows, choose errors or return a first-error
diagnostic. It is a boolean predicate, not a TypeScript type guard claiming that
unvalidated extra fields match an application interface.

## Reproduce and check

Use Go 1.27.3 and Node 24 as recorded in `toolchain.json`. Normal compilation and
checks do not download dependencies. From this module:

```sh
GOTOOLCHAIN=local GOPROXY=off go test -race -count=1 ./...
GOTOOLCHAIN=local GOPROXY=off go vet ./...
node verify/typescript.mjs
output=$(mktemp -d)
GOTOOLCHAIN=local GOPROXY=off go run ./cmd/jsonvalidator \
  -model testdata/sample/model.json -out "$output/generated"
diff -u testdata/sample/validator.go "$output/generated/validator.go"
diff -u testdata/sample/validator.ts "$output/generated/validator.ts"
```

The synthetic sample has 46 literal cases: 12 accepted and 34 refused. Both
languages execute those exact expectations. Compiler tests also execute seven
root-shape models and nine property-presence models with four literals each.
Presence cases include `__proto__`, `constructor` and `toString`. The tests also
cover unsupported models, exact accepted/refused intake boundaries, determinism
and output preservation. Twelve independently reviewed nested-container literals
execute object items, nested arrays and later-element failures in both languages. The sample and expectations were authored
from the written profile before the generic emitter was adapted. An independent review checked every expectation against the profile and found
no semantic mismatch. These finite cases are not a proof over all inputs.

The maintained requirement-change regression follows one synthetic boolean field
from optional to required to required-and-true. Eleven fixed literals per revision
execute in both emitted languages, including the values each change must refuse
and behavior that must stay unchanged. The real CLI output is checked against
that executed source and each model's exact hash. Earlier outputs stay unchanged;
restoring the original model must reproduce both original output files byte for
byte. This producer regression is not consumer acceptance.

The Node check transforms the generated TypeScript with Node's built-in type
transformation. It is an execution check, not `tsc`, a frontend build or browser
acceptance. Its experimental-transform warning is expected on the tested Node
version. Consumers must run their actual compiler and application tests.

Run `bash scripts/validate.sh` from the repository root too. The separate
`json-validator-go` workflow covers this module explicitly; it does not rely on
the lifecycle module's workflow to cover a new directory.

## Output and adoption boundaries

Unsupported model input refuses before creating the output directory. An
existing destination, including a symlink, is refused without overwriting it.
A later filesystem failure can leave a newly created partial directory; output
writes are not transactional. Regenerate into a temporary directory and compare
bytes rather than using generation as an in-place overwrite operation.

A consumer must separately review its model, pin the generator source and
toolchain, preserve generated-file provenance, add regeneration/drift checks and
qualify the actual call site. Synthetic producer checks do not establish any
consumer's accepted behavior or universal Go/TypeScript equivalence.

This merged producer has no published tagged release or consumer activation.
The [producer task](../../tasks/CF-json-boolean-v1-producer-standalone-validators.md)
records its reviewed scope, validation and integration history. Later changes
still need independent exact-source review and observed hosted checks.
