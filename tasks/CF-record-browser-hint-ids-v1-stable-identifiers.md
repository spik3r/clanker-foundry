---
id: CF-record-browser-hint-ids-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Give record-browser key hints stable identifiers

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. The
[generator pipeline plan](../docs/generator-pipeline-plan.md) records the gap
under "Further consumer changes": a host had to translate a key hint by
matching its English text, which breaks silently if the wording changes.

Scope, under `modeling/tui/`: an `ID` field on the generated `KeyHint` type,
drawn from a fixed vocabulary kept in one place and documented in the README
and profile; tests that the observed, coded and documented vocabularies agree;
an optional `id` member on `keyHints` expectations in the cases format, with
schema, document and example updates; `tuic` 1.5.0 with unchanged canonical
digests. Existing hint text does not change. Also this task record and new
immutable Project Memory files.

Exclusions: no change to the consumer application, the contract, the harness
or other generators.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui`, `go test -race`
in `modeling/gencontract` and `modeling/foundrygen`, the node schema tests, the
repository validator, whitespace check and committed Project Memory validation.
