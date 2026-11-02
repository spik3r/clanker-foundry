# Lifecycle profile catalog

This package describes the bounded custom lifecycle language for model authors.
It does not validate a caller's model, execute guards, solve constraints, generate
code, authenticate a source or review, grant authority, or claim UML/BPMN/SysML
conformance. JSON-v1 is the only existing source adapter; this is not a JSON
Schema or a second format.

`model.ReferenceCapabilities` returns sorted owned scalar copies of the existing
reference registry. `profilecatalog.Describe` returns a fresh owned descriptor;
`profilecatalog.Encode` computes its own descriptor and emits bounded compact
JSON plus LF. The fixed `modelprofile` CLI takes no arguments and does not read
stdin. It requires the pinned Go version and writes only the catalog to stdout.

Operator policies and limits are a reviewed descriptive projection, not a
reflective operator registry or an evaluator. Existing `model.TypeOf` and
`model.Validate` remain the authorities. The four source identities are **expected
reviewed source** pins: the API and CLI do not read those files at runtime.
Test-only hash and size checks detect source drift and require a new reviewed
projection; neither hashes nor finite parity vectors authenticate the binary,
host or review, or prove the operator list exhaustive.

From the module root with actual Go 1.27.3, run offline:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 ./model ./profilecatalog ./cmd/modelprofile
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelprofile
```

The sample lifecycle's owner/reviewer rules are model content, not universal
language policy. Reviewpack's separate visual gate remains incomplete.
