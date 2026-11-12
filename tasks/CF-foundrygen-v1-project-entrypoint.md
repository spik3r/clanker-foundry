---
id: CF-foundrygen-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add the foundrygen project entrypoint and drift check

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is step 3 of the
[generator pipeline plan](../docs/generator-pipeline-plan.md).

Scope: a new Go module `modeling/foundrygen/` with the `foundrygen` command. It
reads a consuming repository's `foundry.json`, checks that the Foundry checkout
is at the pinned revision, builds the listed generators and runs their contract
commands. `check` reports drift across all outputs; `generate` regenerates
outputs and refuses to replace a directory it did not generate. Also a new
`generator-pipeline` CI workflow, its entry in the workflow inventory test, this
task record and new immutable Project Memory files.

Exclusions: no change to any generator, the conformance harness, the contract
text or any consuming repository.

Checks: gofmt, `go vet` and `go test -race` in the new module, the workflow
inventory test, the repository validator, whitespace check and committed
Project Memory validation.
