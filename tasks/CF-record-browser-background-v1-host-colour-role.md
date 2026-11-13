---
id: CF-record-browser-background-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add a background colour role to the record browser

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent. The subagent returned no usable report, so the
coordinator read the diff and ran every check itself. This follows the hosted
mode change: its review found that the page background was not among the host
colour roles, so a host with a different background would see a patch of
another colour behind a hosted browser.

Scope, under `modeling/tui/`: a `Background` role in the generated `Colors`
type, used wherever the templates paint the page background, with the gutter
following it when `Border` is unset; tests; the two documents that list the
roles; `tuic` 1.3.1 with unchanged canonical digests. Also this task record and
new immutable Project Memory files.

Exclusions: no model member, no change to default output for a host that sets
no colour, nothing outside `modeling/tui/`.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui`, `go test -race`
in `modeling/gencontract` and `modeling/foundrygen`, the repository validator,
whitespace check and committed Project Memory validation.
