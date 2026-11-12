---
id: CF-generator-contract-jsonvalidator-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Bring the JSON boolean validator generator into generator contract v1

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is part 3 of step 1
in the [generator pipeline plan](../docs/generator-pipeline-plan.md).

Scope: `modeling/jsonvalidator/` gains the contract's `validate`, `generate`
and `check` commands, the contract manifest, the contract header on both
emitted files and a documented canonical encoding in which property order is
not significant. The emitted model hash constants carry the canonical digest.
The old `-model` and `-out` flag form is removed. Committed sample output is
regenerated. `jsonvalidator` joins `modeling/gencontract/generators.json`. The
`json-validator-go` workflow lines that invoke the generator change to the new
form. Also this task record and new immutable Project Memory files.

Exclusions: no change to which JSON values the generated predicates accept or
refuse, to any literal case expectation, to the conformance harness code, the
contract text or any other generator.

Checks: gofmt, `go vet` and `go test -race` in `modeling/jsonvalidator` and
`modeling/gencontract`, the Node TypeScript check, the repository validator,
whitespace check and committed Project Memory validation.
