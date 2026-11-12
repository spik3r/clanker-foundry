---
id: CF-record-browser-consumer-gaps-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Close record-browser gaps found by the first consumer review

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is gap-driven
profile work under steps 4 and 5 of the
[generator pipeline plan](../docs/generator-pipeline-plan.md). The owner named
a consumer application; a read-only review of it on 12 November 2026 found
three gaps that block using a generated browser inside an application module.

Scope, all under `modeling/tui/`: an optional `emit` model member whose package
mode emits only an importable Go package with a model-chosen validated package
name; an optional `detailText` member whose wrap mode shows long and multi-line
detail values in full; and documentation of the Go versions the generator and
its output actually require. Generator version becomes 1.1.0. A model without
the new members keeps its canonical digest. Also this task record and new
immutable Project Memory files.

Exclusions: no change to the consumer application, to the profile's fixed key
contract, to the contract text, the harness, the registry or other generators.
Embedding inside a host page shell, typed fields and list search remain open
gaps.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui`, `go test -race`
in `modeling/gencontract` and `modeling/foundrygen`, the repository validator,
whitespace check and committed Project Memory validation.
