---
id: CF-tui-reload-offsets
status: review
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

## Regression and chosen behavior

The unchanged main compiler and generated acceptance passed first. The new
acceptance was then run against the unchanged state template before the fix. It
failed 16 baseline subcases and eight changed-model subcases. The baseline
single-record reproduction is: size to 24x6, Enter, G, replace with Loading or
Failed, expand to 120x40, shrink to 24x6, then replace with Ready and the same ID.
The detail offset remained one instead of zero. The list reproduction also
retained offset two after expansion where the last six Ready rows in a five-row
viewport require offset one.

The acceptance bytes used for the old-code failures and final passing tests are
identical, SHA-256
`b852731b3818b69ec9f2edd2800928e90739739965bdc10d524fe2e760b0f099`.
These are implementation-authored expectations that do not call state/layout
helpers to calculate the expected results, not a claim of outside review.

The state now remembers only the two last Ready row counts. Supported viewports
clamp offsets during Loading/Failed as well as Ready; ordinary same-size reloads
preserve valid positions. Status-snapshot records cannot replace those counts.
Ready refreshes them, empty Ready clears selection/positions, and losing the
selected ID starts at the first new record. Tiny sizes still freeze offsets.
Selection and focus survive, and a ready list can advance its offset again when
needed to show the selected row. Help remains captive, including at tiny sizes.
Rejected snapshots return the entire original model unchanged.

## Observed checks

Local checks used official Go 1.27.4 linux/amd64, with its archive verified against
the published SHA-256 before execution. Toolchain and caches were task-local.
Node was v24.22.1 and Git was 2.52.0. The reviewed source commit is
`2e16fd62f9a0d33353aba1d96485af577806de16`; registration and a genuine start
preceded test/template edits. Source and final commits are published as one branch
series rather than separate claim-only CI runs.

- Full compiler/CLI and both generated variants passed
  `go test -race -count=1 -p=1 -timeout=5m -v ./...` in `modeling/tui`.
- All 16 new baseline regression subcases also passed in a separately generated
  module. The full suite exercises the same cases in the changed model.
- Short compiler tests, compiler and generated-module vet, generated demo
  compilation, and source/acceptance/generated Go formatting passed.
- `scripts/validate.sh` passed; optional skills-ref was not installed.
- Project Memory/workflow tests: 143 passed, no failures, skips or cancellations.
- Pack-validator path checks: 92 passed. Model-map cases: 47 passed. Project
  setup, model resolver, nine wiring-preflight and sixteen installer-ordering
  cases passed. Full GNU Stow installation acceptance was not run locally because
  GNU Stow is absent; recording-seam tests do not establish that behavior.
- Staged/unstaged whitespace checks passed. Full committed Project Memory
  validation is run after completion against the unchanged reviewed main.

No manual terminal/color review, downstream application trial or local live
scanner was run. Existing hosted CI, including its dependency checks, and fresh
independent review remain required before integration. No Ready/merge or runtime
activation is claimed by this implementation run.

## Security and scope review

A host can supply invalid snapshots or hostile text; this change must not mutate
retained geometry on rejection or expose stale record content. Snapshot validation
and rendering sanitization are unchanged, atomic rejection is tested, and only
bounded integer counts are retained. Fixtures are synthetic. No public API,
authentication, authorization, dependency, secret, file upload, network endpoint,
logging or deployment configuration changes. Password/session, web, TLS, personal
data and backup controls are outside this local state correction. Independent
review must recheck the exact diff before integration.
