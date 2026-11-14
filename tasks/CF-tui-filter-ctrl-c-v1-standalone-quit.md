---
id: CF-tui-filter-ctrl-c-v1
status: in-progress
owner: Codex TUI maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Quit the standalone browser with Ctrl+C during filter input

## Implementation claim

The user approved immediate standalone Ctrl+C quitting during filter input.
Main at `5cfb052`, live claims and open PRs were inspected; no PR is open.
Preserve captive help, hosted non-quitting capture, q as text and all other keys.

Limit implementation to modeling/tui/templates/browser.go.tmpl,
modeling/tui/templates/view.go.tmpl, modeling/tui/templates/keys.go.tmpl,
modeling/tui/hints.go, modeling/tui/acceptance/filter_test.go.txt,
modeling/tui/generated_test.go, modeling/tui/examples/browser-filter.cases.json,
modeling/tui/contract.go, modeling/tui/contract_test.go,
modeling/tui/hosted_test.go, modeling/tui/options_test.go,
modeling/tui/README.md and modeling/tui/PROFILE.md, plus this record and new
immutable Project Memory evidence. Add a standalone exception before filter-key
processing. Retain the stable quit hint ID, with ctrl+c as its key while input
is open; hosted input hints remain enter and esc. Synchronize help and docs.
Bump tuic to 1.5.2 for changed emitted bytes; canonical model digests stay fixed.
There are no tracked generated TUI packages to regenerate.

Prove the changed behavior fails before implementation. Cover empty text,
no matches, loading/failed replacements, empty replacement and tiny resize
while input remains open; hosted variants never return quit. Keep captive-help
and q/text controls. Run generated acceptance, full TUI vet/race, authoring,
generator contract and Foundry entrypoint checks, repository validator,
whitespace and fresh committed Project Memory coverage. Publish for independent
exact-head review and hosted checks; the coordinator owns merge.
