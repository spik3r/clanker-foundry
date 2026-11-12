---
id: CF-generator-contract-lifecycle-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a contract-conforming lifecycle generator command

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is part 4 of step 1
in the [generator pipeline plan](../docs/generator-pipeline-plan.md).

Scope: a new command `modeling/lifecycle/cmd/lifecyclegen` with the contract's
`validate`, `generate` and `check` commands over an explicit model path and
output directory, its tests and a README beside it. Output is the existing
compiler and view bytes with the contract header line added to the two Go files,
plus the contract manifest and no `trace.json`. `lifecyclegen` joins
`modeling/gencontract/generators.json`. Also this task record and new immutable
Project Memory files.

The plan named `modelc` for this part. The coordinator chose a separate command
instead: `modelc validate` with no argument is a working legacy command and a
usage error under the contract, and one binary cannot be both without depending
on the current directory.

Exclusions: no change to `modelc`, to `modeling/lifecycle/generated/`, to the
lifecycle README, which is hash-pinned by an existing test, to the compiler,
canonical, format, model or view packages, or to any other lifecycle command.

Checks: gofmt, `go vet` and the full `go test` in `modeling/lifecycle`,
`go test -race` in `modeling/gencontract`, the repository validator, whitespace
check and committed Project Memory validation.
