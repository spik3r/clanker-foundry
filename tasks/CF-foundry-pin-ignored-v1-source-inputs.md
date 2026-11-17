---
id: CF-foundry-pin-ignored-v1
status: in-progress
owner: Codex pinned-source coordinator
repositories: [spik3r/clanker-foundry]
---

# Refuse ignored local inputs in a pinned modeling checkout

## Implementation claim

The owner requested another supervised round of task selection, implementation,
review, publication and merging after checks. Main at `5c37473` and live open
PRs were checked; no open PR or active claim overlaps this change. Earlier pin
maintenance tasks are completed and retained their previous ignored-file policy.

Investigate the clean-source check's exclusion of ignored files under modeling.
An ignored Go source or embedded data file can affect a generator build even
when ordinary Git status is empty. Reproduce this in disposable local fixtures
before changing behavior. If confirmed, deliberately tighten the pinned mode
to refuse ignored modeling content as well as other local changes. Do not
delete files, change ignore rules or waive checks for dependency directories.
Document the trade-off: installed verification dependencies require a separate
clean Foundry source checkout for pinned consumer generation. Keep ignored
content outside modeling and the explicit development-only --unpinned mode
unchanged. This remains a trusted-host check, not a filesystem sandbox.

Scope: modeling/foundrygen/exec.go, a new
modeling/foundrygen/pin_ignored_test.go, modeling/foundrygen/README.md,
this task and new immutable Project Memory source/events. No generator behavior,
consumer code, project manifest format or runtime changes.

Checks: demonstrate the ignored-input regression before repair; prove clean
source still works and ignored content outside modeling remains allowed;
check that refusals preserve local files and source state. Run gofmt, the full
foundrygen race suite and vet, scripts/validate.sh, whitespace checks and full
committed Project Memory validation. Independently review the change and wait
for hosted checks at the exact PR head before merging.
