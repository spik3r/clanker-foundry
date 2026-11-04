---
id: CF-check-preparation-eligibility-v1
status: review
owner: Foundry validation eligibility
repositories: [spik3r/clanker-foundry]
---

# Check current eligibility before starting validation work

## Implementation claim

The user authorized the next implementation and review batch. The integration
lead allocated this new bounded defect separately from held general recovery
and completed dispatch-preparation work. No existing claim transfers. Base
`6eb8049`, branch `codex/check-preparation-eligibility-v1`. The isolated checkout
had no staged, unstaged or untracked changes. The lead checked live PRs and found
only the read-only cancellation trace lane, which does not edit these paths.
The admitted-builder milestone lane has no runtime production overlap.

Exact scope:

- `runtime/src/runner.mjs`
- `runtime/test/check-preparation-eligibility.test.mjs` (new)
- `runtime/README.md`
- `tasks/CF-check-preparation-eligibility-v1-final-boundary.md`
- New immutable Project Memory source and event files for this task.

A public API reproduction used the actual store and runner with an invented
validation callback. At the exact validation deadline, draining checks still
invoked that callback. Only the later evidence command rejected the expired
attempt. The integration lead independently reproduced this behavior.

Before each new check callback, prepare an owned input and artifact snapshot,
then re-read the current execution. Require the same VALIDATING attempt identity,
fence and current fence, no pending cleanup, and the same result subject and
specification digest as the snapshot. Check the fresh injected clock strictly
before validation_deadline. Do not await between the final check and check.run.
Changed eligibility keeps the existing UNKNOWN status
`ATTEMPT_CHANGED_DURING_CHECKS`; expiry uses the existing fixed error code
`ATTEMPT_EXPIRED` through normal uncertain-outcome handling.

Keep already recorded checks skipped without starting new work. Preserve late
output enforcement, no automatic resume and separate owner cleanup. This check
does not revoke a callback already running or promise atomic cancellation across
processes. Do not edit state, store, schemas, adapters, providers, CLI, existing
tests, authority configuration or held general recovery records. Use invented
checks and temporary private databases; no real provider, process or container.

The lead reviews this published claim before source registration and genuine
start. Freeze independently authored tests after that start and before production
edits. Preserve frozen tests and all historical evidence. This lane may publish
its draft PR and implementation; the integration lead alone reviews and merges.

## Acceptance and validation plan

Independently freeze actual-store checks for expiry at and after the deadline,
an eligible callback just before the deadline, and a first check durably recorded
before time advances to the deadline ahead of the second check. Cover changed
state, fence, result subject and specification identity during input preparation,
plus existing recorded-check skips, late output refusal and uncertain-outcome
cleanup controls. Count callback starts rather than relying only on final state.

Observe old-code failures before implementation. Run focused and full runtime
checks on the available Node 24-or-later host, repository and installer checks,
whitespace and committed Project Memory validation against the actual PR base.
Use disposable syntax-valid mutation controls to demonstrate the tests detect
removed deadline or eligibility checks. Record only observed results and separate
local results from hosted Node 24 CI. No live runtime activation is included.

## Observed result

The runner now prepares the owned callback input, then refreshes current attempt,
worker, fence, cleanup state, result subject and specification identity and checks
the current validation deadline. There is no await before starting the callback.
Already recorded check skips and evidence enforcement after callback completion
remain unchanged. Production changes are limited to the runner and README.

Independent tests were frozen after the genuine start and before production at
SHA-256 `02c0d49daff7777d90cf22dadc35ad3392c77bf918e2978223b41d1824d35afe`.
The author observed two positive controls passing and seven semantic failures on
old code. The integration lead separately confirmed old-code failure. Frozen
bytes and all existing tests remain unchanged.

Actual Node v26.0.0 passed all nine frozen tests. The cases cover a current owned
input, exact-deadline refusal, cancellation or worker/fence/subject/specification
changes during artifact preparation, the second check reaching its deadline after
first evidence commits, and late-output refusal. Planned extra timing or recorded
skip scenarios are not claimed as separate new observations. Independent source
review confirmed the strict comparison and existing recorded-check skip path.

The full pre-integration runtime suite passed 304/304 tests without failures,
skips or cancellations. An ordinary merge incorporated main
`a00494b771c726f2d51a89132978720ba76273b6` without conflicts or changes to this
lane's runtime bytes. Afterward the combined eligibility and cancellation-trace
checks passed 21/21. Existing trace pins and all merged trace paths remain intact.

Six syntax-valid mutants in a disposable clone caused assertion failures: removed
deadline check (two), attempt check (two), worker check (one), subject check (one),
specification binding (one), or refreshed read (five). Restored source passed all
nine tests. Repository validation, installer/model resolver/wiring regressions and
whitespace checks passed; optional skills-ref is unavailable. Committed Project
Memory validation follows this staged result and completion record.

A separate reviewer found no issue in source `abebd8c` and independently passed
the nine frozen tests. All callbacks and workspace data were invented; no real
validation tools, provider or container was invoked. Local Node 26 results do not
establish hosted Node 24 results. Exact-head hosted checks and final integration
review remain required before the lead merges. This claim reserves only bounded
review corrections and does not acquire held general recovery work.
