---
id: CF-dispatch-preparation-eligibility-v1
status: review
owner: Foundry dispatch eligibility
repositories: [spik3r/clanker-foundry]
---

# Check current eligibility after asynchronous preparation

## Implementation claim

The user authorized the next implementation and review batch. The integration
lead allocated this new bounded defect separately from held general recovery
work and historical completed runtime work; no existing claim transfers. Base
`3091153`, branch `codex/dispatch-preparation-eligibility-v1`. The isolated
checkout had no staged, unstaged or untracked changes. The lead checked current
main and found no open PRs before allocating this lane.

Exact scope:

- `runtime/src/runner.mjs`
- `runtime/src/adapters.mjs`
- `runtime/src/isolation.mjs`
- `runtime/test/dispatch-preparation-eligibility.test.mjs` (new)
- `runtime/README.md`
- `tasks/CF-dispatch-preparation-eligibility-v1-final-guard.md`
- New immutable Project Memory source and event files for this task.

A public API reproduction used the actual store and runner with an invented
text callback and delayed workspace reads. Cancelling the task or advancing the
injected clock beyond its lease while reads waited still allowed the callback
to start. The runner refused submission only afterward. Existing adapter abort
checks require revocation delivery and do not refresh recorded eligibility.

Add an optional synchronous `beforeInvoke` assertion to owned adapter attempts.
Omission preserves direct callers. A supplied assertion succeeds only by returning
undefined; a thrown error refuses the boundary. Reject thenables without awaiting
them. The runner supplies a guard that re-reads the same EXECUTING attempt,
fencing token and specification identity, requires no pending cleanup, calls
verifyDispatch and checks the fresh injected clock strictly before the lease
deadline. Use fixed errors for a changed attempt or expired lease. Do not await
between this guard and the actual callback.

Text and isolated adapters check after asynchronous preparation, including queued
capacity, immediately before the provider callback. The fake adapter checks after
its reads and capacity wait before materializing a result. The isolated adapter
also checks immediately before starting workspace.applyPlan after awaiting the
provider. This checks a new work boundary; it does not revoke work already running
or promise atomic cancellation across processes.

Preserve uncertain outcomes, no automatic resume and explicit owner cleanup or
reconciliation. Do not modify state, store, schema, CLI, providers, existing tests,
namespace activation, credentials or general recovery records. Use invented
callbacks and fake workspace hosts only, with no real provider or container.
The lead reviews this published claim before registration and genuine start;
freeze independently authored tests after that start and before production edits.
This lane may publish its draft PR and implementation. The integration lead alone
reviews the integration and merges.

## Acceptance and validation plan

Independently freeze bounded deferred-read and capacity tests using the actual
store and runner. Cover cancellation, expiry, changed attempt or fence, refreshed
dispatch authority and eligible controls. Check the queued callback boundary and
both isolated boundaries with invented callbacks and workspace operations. Keep
optional-guard omission compatible; refuse supplied asynchronous or invalid guard
results. Confirm refused work starts no callback or new workspace operation,
retains existing uncertain-outcome and cleanup rules and does not resume itself.

Observe old-code failures before implementation. Run focused and full runtime
checks on the available Node 24-or-later host, repository and installer validation,
whitespace and committed Project Memory validation against the actual PR base.
Use disposable mutation controls that remove individual eligibility boundaries;
report only syntax-valid mutants detected by assertions. Preserve frozen test
bytes and all historical evidence. Distinguish local results from hosted Node 24
checks and leave merge to the lead after independent review.

## Observed result

The runner now supplies a final synchronous assertion, and the text, fake and
isolated adapters call it at the claimed boundaries. It refreshes attempt, worker,
fence and specification identity, dispatch authority and lease time. The isolated
adapter checks again before starting workspace execution after provider output.
Direct callers may omit the assertion. Invalid guard results are refused without
awaiting them or invoking arbitrary thenables; native Promise rejection is consumed.
Existing UNKNOWN handling and separate owner cleanup remain in place.

The independently authored test file was frozen before production at SHA-256
`1f1643dddf689d445e74de6af5f241e7ee961161bf4a23cdd5f6a3d884a5dd60`.
The author observed three passes and 12 failures on old code; the integration lead
separately confirmed the old-code failure. An initial missing dependency prevented
loading before the author installed pinned dependencies offline; that load failure
is not counted as a behavioral regression. Frozen expectations remain unchanged.

Actual Node v26.0.0 passed all 15 frozen tests and the full 295-test runtime suite,
with no failures, skips or cancellations. Repository validation, installer/model
resolver/wiring regressions and whitespace checks passed. Optional skills-ref is
unavailable. Eight syntax-valid mutants in a disposable clone caused assertion
failures: omitting the runner guard (six), text guard (five), fake guard (four),
isolated invocation guard (two), isolated apply guard (one), lease guard (two),
fresh dispatch guard (two), or synchronous-return restriction (three). Restored
source passed all 15 tests. No frozen test or existing test was edited.

These observations use invented callbacks and fake workspace operations; no real
provider or container was invoked. They do not prove revocation of running work
or atomic cancellation across processes. The integration lead reviewed production
source `7e05bb6` without requesting corrections. Separate independent review and
exact-head hosted Node 24 checks remain required before the lead merges. Local
Node 26 results do not establish hosted CI results. This completed implementation
claim reserves only bounded review corrections, not held general recovery work.
