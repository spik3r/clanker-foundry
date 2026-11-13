---
id: CF-json-boolean-gaps-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Close recorded gaps in the JSON boolean validator profile

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. Two records name the
same gaps in `json-boolean/v1`: the first consumer review in the
[generator pipeline plan](../docs/generator-pipeline-plan.md) and the gap
report for brief 4 of the
[authoring trial](../docs/trials/2026-11-13-authoring/README.md).

Scope, under `modeling/jsonvalidator/` and the profile registry: optional
string `const`, `enum`, `minLength` and `maxLength`; `additionalProperties`
may be `false`; canonical encoding rules for these with unchanged digests for
models that use none; generator version 1.1.0 with regenerated fixtures; two
new fixtures with literal cases run in Go and TypeScript, one being the
trial's plugin manifest requirement; schema, registry and documents updated;
one loop in the `json-validator-go` workflow that checks the new fixtures. Also
this task record and new immutable Project Memory files.

Exclusions: numbers, regular expressions, unions, cross-field rules, error
messages, defaults and nullable values stay outside the profile. No change to
other generators, the harness or the contract.

Checks: gofmt, `go vet` and `go test -race` in `modeling/jsonvalidator`, the
TypeScript check, `go test -race` in `modeling/gencontract` and
`modeling/foundrygen`, the node schema test, the repository validator,
whitespace check and committed Project Memory validation.
