---
id: CF-generator-contract-tuic-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Bring the record-browser generator into generator contract v1

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is part 2 of step 1
in the [generator pipeline plan](../docs/generator-pipeline-plan.md).

Scope: `modeling/tui/` gains the contract's `validate`, `generate` and `check`
commands, the contract manifest in place of `trace.json`, the contract header
on emitted Go files and a documented canonical encoding. The old positional
`tuic MODEL OUT` form is removed. `tuic` joins
`modeling/gencontract/generators.json`. The `tui-model-go` workflow line that
invokes the generator changes to the new form. Also this task record and new
immutable Project Memory files.

Exclusions: no change to the record-browser profile's meaning, the generated
browser's behaviour, the conformance harness code, the contract text or any
other generator.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui` and
`modeling/gencontract`, the repository validator, whitespace check and committed
Project Memory validation.
