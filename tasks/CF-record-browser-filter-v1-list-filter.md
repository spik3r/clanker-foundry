---
id: CF-record-browser-filter-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a list filter to the record browser

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by
supervised subagents and coordinator review of the diff. Two records name this
gap: the first consumer review in the
[generator pipeline plan](../docs/generator-pipeline-plan.md) and the gap
report for brief 5 of the
[authoring trial](../docs/trials/2026-11-13-authoring/README.md).

Scope, under `modeling/tui/` and the profile registry: an optional `filter`
model member naming the fields to match; a `/` filter input with fixed keys,
case-insensitive substring matching, selection that follows identity, and a
label showing the filter and match count; `FilterText`, `FilterOpen` and
`Shown` accessors; a `text` step, `/` and `backspace` keys and `filter`,
`filterOpen` and `shown` expectations in the cases format; a filter example
with cases; schemas, registry and documents updated. `tuic` becomes 1.4.0 with
unchanged canonical digests for existing models. Also this task record and new
immutable Project Memory files.

Exclusions: sorting, regular expressions, typed fields, and any change to the
consumer application, the contract, the harness or other generators.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui`, `go test -race`
in `modeling/gencontract` and `modeling/foundrygen`, the node schema tests, the
repository validator, whitespace check and committed Project Memory validation.
