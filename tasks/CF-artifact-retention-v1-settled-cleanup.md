---
id: CF-artifact-retention-v1
status: in-progress
owner: Codex artifact retention contributor
repositories: [spik3r/clanker-foundry]
---

# Release abandoned output capacity after matching settled cleanup

## Implementation claim

Owner: Codex artifact retention contributor. Branch `codex/artifact-retention-v1`
from reviewed main `2844622`. The repository owner requested implementation of the
next selected larger tasks. The lead independently reproduced this defect and
approved this bounded contract. The worker may publish a draft and validated
implementation; the lead owns independent review and merge.

Current public adapter/review claims reserve no further edits, and the live PR
check found no competing runtime lane. This repairs abandoned output retention,
not held general recovery conformance. No held task, source or evidence is adopted
or its limits waived.

Exact write scope:
- `runtime/src/adapters.mjs`
- `runtime/src/runner.mjs`
- New `runtime/test/artifact-retention.test.mjs`
- `runtime/README.md`
- This task and its immutable Project Memory source/events.

Preserve existing tests, store/evaluator, schemas, CLI, CI, provider transports,
workspace host implementation and other recovery claims. Use invented callbacks
and private temporary databases; no real provider, worker activation or deployment.

## Reproduced problem

Run valid text attempts through the real store and runner, inject ENOSPC at
artifact persistence, then owner-cancel each attempt and confirm settled cleanup.
After 16 attempts, all tasks report resources and cleanup released, but the text
adapter retains all 16 output capacity references. The 17th healthy attempt waits
before invoking its provider. The lead independently confirmed this reproduction.
Fake adapter cleanup has the same retained-reference gap.

## Approved retention contract

Keep output while its attempt is active or its outcome remains unknown without
matching settled cleanup. Release a reference after durable persistence, or when
revocation confirms that exact owned attempt has settled and matching cancellation
cleanup discards its abandoned output. Never discard an active/unsettled result.
Bind retained references to attempt ID, worker and fencing token so identical
outputs cannot free another attempt's capacity. Repeated cleanup and release do
not decrement twice.

Keep `releaseArtifact(digest)` compatible as a deterministic single-reference
release. Add an optional full attempt identity object with attempt_id, worker_id
and fencing_token for exact scoped release; the runner supplies its assigned
attempt. Wrong or incomplete identities return false and release nothing.
Full matching settled revocation checks the attempt identity and any supplied
outer fencing token before discarding its own output. Legacy text revocation by
attempt ID alone keeps its stop/abort behavior but does not discard output;
custom callers retain their existing explicit release responsibility.

## Acceptance and validation

Freeze separately authored lifecycle tests before production edits. Observe the
old code fail the new retention expectations while valid handoff, unknown-output
retention and compatibility controls pass. Cover persistence failure followed by
owner cancellation/settled cleanup and successful later invocation, shared output
digests, wrong identities/fences, incomplete scoped release, double cleanup,
active/unsettled revocation and bounded cleanup of waiting attempts.

Run focused and full runtime tests on the actual local Node version (24 or later),
mutation controls, syntax, pack, whitespace and committed-diff Project Memory
validation. Report the actual version and any environment limits; hosted Node 24
checks remain a separate merge gate. No claim of universal host cleanup, automatic
retry or arbitrary callback cancellation follows from this repair.
