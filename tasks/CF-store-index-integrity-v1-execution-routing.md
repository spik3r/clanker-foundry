---
id: CF-store-index-integrity-v1
status: review
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


## Implementation and observed result

Production source `20661ac1eec16bb7fbe5e6b57d5516eeed1451a8` adds the streamed
unfiltered transaction guard, strict selected-row checks and atomic dependency
selection. The duplicated revision is retrieved through SQLite text/type
projections before comparison, so wide integer metadata cannot bypass the fixed
diagnostic through a driver conversion error. Database schema/version/startup
behavior, other table integrity rules and held recovery work are unchanged.

The separate contributor froze `runtime/test/store-integrity.test.mjs` in
`24ac434` before production, SHA-256
`2d2cfd73fc188b7521ee9a1769f781af3a5d3d18f5fa85a317a5f205f545eb79`.
On unchanged old code, 21 corruption expectations failed and both positive
restart/large-aggregate controls passed. The new source passes all 23 tests,
including exact fixed diagnostics and unchanged logical-table snapshots on
refusal. The frozen test bytes remain unchanged.

Six compiling mutants were rejected: removing the unfiltered transaction scan
caused 20 failures, task-id comparison 5, specification-content digest 2, strict
body schema validation 1, selected read validation 15 and atomic dependency
checking 2. Restoring the source passed all 23 tests. Experiments used a separate
disposable clone and invented private databases, never an operating authority.

Observed local validation on Node v26.0.0: all 225 runtime tests passed, no skips,
with pinned dependencies installed offline and scripts disabled. The first full
run passed 224/225; the existing offline consumer-pack test tried the protected
default npm cache. An environment-only rerun selected the private populated cache
and passed 225/225 without source changes. Syntax, pack and whitespace checks
passed. Optional skills-ref is unavailable and reported skipped.

A fresh source reviewer and the lead reviewed exact `20661ac` without blocking
findings. The independent reviewer reran all 23 unchanged contract tests and
checked the full diff's whitespace. Linear scan cost and the limits of selected
reads are documented. This detects duplicated execution routing inconsistencies,
not arbitrary database corruption or malicious writers. Hosted Node 24 CI,
committed-diff evidence validation and lead merge remain separate final checks;
local completion does not declare the PR integrated or release its claim.
