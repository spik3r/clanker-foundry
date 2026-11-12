---
id: CF-generator-contract-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Define generator contract v1 and its conformance harness

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by supervised
subagents and coordinator review of every diff. This is part 1 of step 1 in the
[generator pipeline plan](../docs/generator-pipeline-plan.md), accepted by the
repository owner on 12 November 2026.

Scope: a written contract in `modeling/CONTRACT.md`; a new Go module
`modeling/gencontract/` holding a black-box conformance harness, a small
reference generator and tests that show the harness detects each class of
violation; a registry of conforming generators that starts with the reference
generator only; and a new `generator-contract` CI workflow. Also this task
record and new immutable Project Memory files.

Exclusions: no change to `modeling/lifecycle/`, `modeling/jsonvalidator/`,
`modeling/tui/`, the runtime, skills or existing workflows. Bringing each
existing generator into conformance is a separate task and claim.

Checks: gofmt, `go vet` and `go test -race` in the new module, the repository
validator, whitespace check and committed Project Memory validation.
