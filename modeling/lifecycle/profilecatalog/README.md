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


## Fresh adoption and export compatibility

The original catalog and TypeOf parity vectors remain unchanged. The catalog
complements the [JSON authoring schema](../authoring/README.md): it describes
reference types, operator policies and limits, while Go acceptance still decides
whether the original source bytes form a valid model. Discovery is not validation.

The additive `model/capabilities.go` file is included in the current qualification
source policy at `1e88dbebb4ab13873e2a4159dc928dd1f5f73e20`: twelve exact source
files and eighteen payloads plus manifest. The existing eleven-source policy at
`3d6d17188cc2723cfe04649d0f8b4220a37b1afe` remains supported for byte-exact archive
verification and explicit historical typed builds. No old archive acquires the
new source or is relabelled current. See [qualification](../qualification/README.md)
and [design-change export](../designchange/CONTRACT.md) for the two closed policies.
The catalog's four expected language-source identities still describe unchanged
bytes; they do not read or authenticate source files at runtime.

The CLI's exit-3 output-failure policy includes actual closed stdout. It still
reads no stdin, accepts no arguments and emits the same independently frozen
catalog bytes. No caller model, replay, approval or execution evidence is produced.
