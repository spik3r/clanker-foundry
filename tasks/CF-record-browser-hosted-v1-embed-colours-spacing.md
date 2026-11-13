---
id: CF-record-browser-hosted-v1
status: in-progress
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Let a generated record browser live inside a host application's page

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is the first item
under "Next, from these trials" in the
[generator pipeline plan](../docs/generator-pipeline-plan.md): the gap that
kept the first consumer's generated screen out of its main navigation.

Scope, under `modeling/tui/` and the profile registry: an optional `embed`
member whose hosted mode draws only the body, never quits the program and
leaves the title and key hints to the host; read-only `CapturesKeys`,
`KeyHints` and `Title` accessors in both modes; a run-time `WithColors` method
taking validated colour roles; an optional `detailSpacing` member that puts a
blank row between detail fields; `capturesKeys` and `keyHints` expectations in
the cases format; a hosted example with cases; an acceptance test that embeds
a hosted browser in a small shell with its own bars and page keys; two document
fixes the authoring trial asked for. `tuic` becomes 1.3.0 with unchanged
canonical digests for existing models. Also this task record and new immutable
Project Memory files.

Exclusions: no change to the consumer application, the generator contract, the
harness or other generators. The page background stays a fixed colour; list
search and typed fields remain open gaps.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui`, `go test -race`
in `modeling/gencontract` and `modeling/foundrygen`, the node schema tests, the
repository validator, whitespace check and committed Project Memory validation.
