---
id: CF-dispatch-preparation-eligibility-v1
status: in-progress
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
