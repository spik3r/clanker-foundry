---
id: CF-store-crash-regressions-v1
status: in-progress
owner: Foundry store crash regression contributor
repositories: [spik3r/clanker-foundry]
---

# Preserve store process-crash regressions

## Implementation claim

Add maintained tests for existing transactional persistence contracts at reviewed
main d1bb80cc2dad0d9e66f4fcaeb880192e450580af. Independent private probes
passed; no production defect was reproduced. This work preserves missing maintained
qualification, not a bug fix or runtime feature.

Exact maintained scope:
- runtime/test/store-crash.test.mjs
- runtime/test/helpers/store-crash-child.mjs
- tasks/CF-store-crash-regressions-v1-process-boundaries.md
- New immutable Project Memory source/start/completion records for this task

Do not edit production source, SQL semantics, interfaces, dependencies, package
scripts, workflows, existing tests or fixtures, production documentation, other
tasks or old evidence. Existing test discovery must find the new test file.
Test/harness execution must not invoke a provider, adapter or candidate, activate
a managed namespace, access a user authority database or use the external network.
Authorized GitHub source, branch and draft publication remains in scope as stated
below. No held PR or other repository action is included. No production failpoint
is added. Previously completed runner callback, cancellation, deadline
and claim-recovery tests are outside scope.

Obtain separate coordinator review of these exact source bytes before committing
the claim. Verify the genuine pinned source commit, register it and record a real
start before test/helper edits. Recheck current main, ownership and open PRs.
Publish only one complete tested branch and a draft PR. Independent exact-head
review and hosted CI remain required; this phase does not mark Ready or merge.

## Frozen expected outcomes

1. Kill an owned child after all assignment writes and immediately before the
   real transaction COMMIT. A pre-opened reader sees only the prior committed
   state. With no competing retry, reopen and replay match the exact prior
   logical-table snapshot. A later identical command produces one assignment.
2. Open an independent competing store before the writer pauses. Release that
   process to the identical command and observe its transaction-entry marker
   while the writer still holds the lock. Kill the writer, then require the
   competitor to return one complete assignment and stable duplicate receipt.
   This proves ordered attempted transaction entry and recovery under an owned
   writer lock, not a measured native SQLite blocked interval or timeout limit.
3. Kill after actual assignment COMMIT returns but before execute returns a
   response. The committed receipt and rows are visible before death. Two
   independently opened processes released to identical retries return that
   original receipt; changed payload under the same command ID is refused with
   COMMAND_ID_REUSED. Only one assignment event and effect intent remain.
4. Kill after actual acknowledgment COMMIT but before its response. DELIVERED
   and the original synthetic receipt survive reopen and concurrent identical
   acknowledgments. Wrong claim token yields STALE_EFFECT_CLAIM; changed receipt
   yields EFFECT_OUTCOME_CONFLICT. Replay and reclaim retain delivered status
   with no pending/redeliverable work.

Before assignment the six logical-table counts must be executions=1, events=1,
attempts=0, reservations=0, outbox=0, receipts=1. After assignment they must be
1,2,1,1,1,2 respectively. Verify complete normalized rows/snapshots where needed,
replayed executions against persisted executions, the single original effect,
and semantic receipt equality across JSON and null-prototype representations.
No provider is invoked by any synthetic receipt.

## Harness and validation

Use readable test-local helpers and bounded invented databases under private
resolved temporary directories. Track all opened connections and child handles;
close connections, terminate only live owned children, await checked exits and
remove temporary files on success or failure. Capture bounded diagnostics and
fail clearly on unexpected exits, messages or missed failpoints. Use explicit
ready/go/transaction-boundary markers, finite parent and child deadlines, and no
short sleep as evidence that a competing process is blocked. Polling for an
explicit marker is permitted; elapsed time alone is not synchronization.

Install child-local DatabaseSync wrappers only after store initialization. Match
the transaction BEGIN/COMMIT boundaries, call original methods with their original
arguments and receiver, and pause immediately before or after the real COMMIT.
Before-COMMIT observation uses the transaction's own writes to confirm the full
expected assignment write set. A missed boundary must fail, never pass silently.
The existing SQL/evaluator behavior is otherwise unchanged. Only the exact known
spawned handle may receive SIGKILL. Tests target the supported local POSIX setup;
report actual platform qualification and do not add broad silent skips.

Preserve the original independent proof unchanged. Run the new suite at least
three times on Node 24, and run two independent suite instances concurrently with
separate temporary directories. Run existing store/integrity suites, full Node 24
runtime tests with pinned dependencies and a private npm cache, repository
validation, syntax, whitespace and full committed Project Memory validation.
Record actual durations and deadlines without claiming performance or platforms
not run. Report optional unavailable tools honestly.

In separate disposable copies, retain and distinguish intentional negative
controls: bypass the COMMIT hook (expect missing-boundary failure), commit before
the pre-COMMIT pause (expect prior-snapshot failure), return a changed duplicate
receipt (expect original-receipt mismatch), and omit acknowledgment persistence
(expect delivered-state failure). Mutate only disposable copies, never production
source in the maintained checkout; preserve test expectations. Restore a pristine
copy and require a passing run. If these controls expose a harness weakness, fix
only the approved harness and rerun its validation before completion.

SIGKILL qualifies process-crash behavior only, not host power loss, disk corruption,
busy-timeout exhaustion, multi-host storage or exactly-once provider execution.
This change does not establish rollout, consumer acceptance or measured benefit.
