---
id: CF-tui-navigation-affordances
status: in-progress
owner: Foundry responsive browser contributor
repositories: [spik3r/clanker-foundry]
---

# Keep generated master/detail navigation visible

## Implementation claim

Owner: Foundry responsive browser contributor. The integration coordinator reviewed
this bounded scope after inspecting current main `4a4580469df3ce11c2521bd9375580317b1ba3a0`
and the non-overlapping generator-contract PR118. Work on branch
`codex/responsive-record-browser` in an isolated checkout. This is a source claim,
not managed runtime dispatch or a shared authority lock.

Exact paths: `modeling/tui/templates/view.go.tmpl`,
`modeling/tui/acceptance/browser_test.go.txt`,
`modeling/tui/RESPONSIVE-INTEGRATION.md`, this task, and its new immutable
Project Memory source/event files.

Expose Enter/open or Esc/back plus help/quit at the minimum supported viewport.
Show additional hints only when they fit; keep captive help close/scroll hints
visible. Cover baseline and changed models, both layout dimensions, repeated
navigation/resizes, empty/loading/failed states and unchanged rendering state.
Explain the bounded adaptation of Flexgo's public responsive reference.
Preserve schema, package API, imports, snapshot safety, key bindings and existing
scroll/clamping semantics. Do not edit PR118's contract, compiler or README paths,
Flexgo core, private consumer code, runtime authority, or deployment configuration.

## Validation plan

Reproduce missing narrow open/back hints before changing the template, then run
compiler determinism/model tests and both generated acceptance variants with race
checks, Go vet, pack validation and committed lifecycle validation. Use patched
Go 1.27.4. Obtain independent source review before one draft PR and require fresh
exact-head CI before merge. State unrun interactive TTY checks explicitly.

## Observed validation

Both unchanged-template generated variants failed the new minimum-width hint
assertion: the list showed only `tab pane ? help q quit`, without Enter/open.
After the bounded footer change, full compiler/CLI and both generated acceptance
variants passed race checks on Go 1.27.4. Go vet, whitespace and pack validation
passed. The optional skills-ref checker was not installed. No interactive TTY or
private downstream consumer test was run.

The initial genuine start event accidentally encoded the file list as one
comma-joined path. The source claim above already listed the exact scope. An
immutable checkpoint records that correction and discloses that tests and the
guide had been drafted before it; template implementation followed the correction.
No history or chronology was rewritten.

See the [consumer integration guide](../modeling/tui/RESPONSIVE-INTEGRATION.md).
Independent review and exact-head hosted checks are still required for merge.
