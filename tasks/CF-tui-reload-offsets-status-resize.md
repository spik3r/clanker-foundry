---
id: CF-tui-reload-offsets
status: in-progress
owner: Foundry TUI maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Clamp retained browser offsets during status displays

## Implementation claim

Owner: Foundry TUI maintenance contributor. Branch
`fix/tui-reload-offsets-20261111`, reviewed main base
`6fbc4a6b3accd558188dd2444e47e29d21894211`. The integration lead reviewed this
bounded follow-on after PR107 and PR108 merged. The completed CF-flexgo-tui-v1
owner is reconciled in its canonical record; previous history remains immutable.
The current open-PR check found only unrelated draft PR101. This local claim is
not a runtime namespace decision or shared lock.

Exact scope:

- `modeling/tui/templates/state.go.tmpl`
- `modeling/tui/acceptance/browser_test.go.txt`
- `modeling/tui/README.md`
- This task and its new immutable Project Memory source/event files.

First reproduce through generated acceptance whether Loading or Failed between
supported viewport expansion and shrinking resurrects an old scroll offset.
Compare that result with the documented clamping contract before choosing a fix.
Preserve ordinary same-size reload positions rather than resetting every status
transition. If confirmed, retain only the last Ready list/detail extents so that
supported resizes can clamp offsets while a status is shown, without retaining
old record content. Tiny viewports must continue to freeze offsets.

Preserve selected identity, focus, selected-row visibility on Ready, atomic
snapshot validation, captive help, empty-ready behavior and changed-ID resets.
Cover both statuses, repeated/interrupted transitions, partial list clamping,
expansion and shrinking, same-size retention, rejected snapshots, and both
baseline and changed generated models. Do not derive expected values from the
state/layout implementation. Document the precise status/scroll behavior.

Do not change Flexgo core or the held prototype, other modeling profiles,
runtime code, providers, scanner configuration, production, namespace activation,
deployment or PR101 readiness. No new features or component extraction. Publish
a draft after local validation; independent review and hosted CI are separate
gates, and this worker must not mark ready or merge.

## Validation plan

Commit this reviewed public-safe source, register it and record a genuine start
before implementation edits. Record the generated old-code failure, then run the
new focused regressions and full compiler/generated acceptance with race checks
on patched Go 1.27.4. Run compiler/generated vet, formatting, pack validation,
whitespace checks and committed Project Memory validation against actual main.
Use isolated synthetic fixtures and task-local build caches. Report unrun manual
terminal visual checks and any blocked checks precisely. Preserve all existing
source/event and legacy-ledger bytes.
