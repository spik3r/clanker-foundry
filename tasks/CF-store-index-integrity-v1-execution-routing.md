---
id: CF-store-index-integrity-v1
status: in-progress
owner: Codex store index integrity contributor
repositories: [spik3r/clanker-foundry]
---

# Reject inconsistent indexed execution records before routing

## Implementation claim

Owner: Codex store index integrity contributor. Branch
`codex/store-index-integrity-v1`, reviewed main base
`2e2a190b629a983ba854d59ecc356274ec7e5a4f`. The repository owner requested the
next bounded larger tasks after the prior wave was reviewed and merged. The lead
approved this reproduced defect and exact repair plan. The lead owns independent
review and merge; this worker may publish its branch and PR after validation.

Current public store/review claims reserve no further edits, and the open-PR
check found no competing lane after PR56 merged. The lead approved this new
indexed-execution integrity defect separately from any held general recovery
conformance work. An unknown unpublished scope is not a claim on this defect.
Do not publish, modify or waive the limits of held work.

Exact scope:
- `runtime/src/store.mjs`
- New `runtime/test/store-integrity.test.mjs`
- `runtime/README.md`
- This task and its new immutable Project Memory records.

Exclude evaluator, runner, CLI, schemas, CI, database DDL/version/migrations,
other tables' integrity rules and recovery-conformance task records. Preserve
existing evidence and unrelated work.

## Reproduced problem and approved behavior

In a new invented private SQLite database, admit a task, then change only its
indexed `executions.state` to COMPLETED while its JSON body remains PENDING.
Existing replay accepts this mismatch. A second admission for the same task then
succeeds, leaving two PENDING execution bodies because the indexed active-task
query and partial unique index see the altered column instead of the body.

Before each existing transaction operation, stream an unfiltered execution-row
scan. Strictly parse and validate each TaskExecution body and its specification
digest; compare indexed execution_id, task_id, spec_digest, state and revision
with their body counterparts. This happens before receipt shortcuts and filtered
routing. Validate returned get/list rows; make dependency output guarding and
selection atomic within the existing transaction boundary. Replay also rejects
inconsistent rows. Use fixed EXECUTION_INDEX_CORRUPTED and fixed diagnostic:
Stored execution routing fields disagree with its validated body.

Do not repair, rewrite or delete inconsistent records. Document the linear scan
cost per transaction. This checks duplicated execution routing fields, not all
SQLite tables, missing nonindexed lookups or malicious concurrent database writers.
Opening/schema initialization retains its current behavior; this task adds no
constructor migration or authority activation.

## Acceptance and checks

A separate contributor freezes raw-SQL corruption tests before production edits.
Observe the old implementation fail those expectations, while valid restart and
principal-bound receipt controls pass. Cover every duplicated field, hidden active
rows, strict body failures, specification digest mismatch and rollback without
logical-table changes. Confirm consistent non-routing body drift remains replay's
existing responsibility. Exercise live store reads, dependency output selection,
receipt shortcuts, replay and reopening without effects or workers.

Run focused and full runtime tests under Node 24 or later, independent mutation
controls, pack validation, whitespace and committed-diff Project Memory validation.
Record actual results and limitations. Only lead review and passing hosted checks
may authorize merge; local completion alone does not declare integration.
