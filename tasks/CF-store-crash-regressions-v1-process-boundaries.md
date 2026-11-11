---
id: CF-store-crash-regressions-v1
status: review
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


## Observed result

The coordinator separately reviewed the source before GitHub source commit
85700166571a17ddea80073ea9d6ef38f3f9ede5 was created, fetched and verified
against the approved bytes. Registration and a genuine start preceded maintained
test/helper edits. Main remained d1bb80cc2dad0d9e66f4fcaeb880192e450580af;
114 prior task frontiers were single completed records, and the coordinator
confirmed no runtime overlap. The three existing open PRs were untouched.

Four tests now preserve the frozen outcomes. Children open before their explicit
go signal, report transaction-entry and actual COMMIT boundaries, and never
invoke a provider. Parent reads verify the complete prior or committed row set.
Before-COMMIT transaction-local rows match the later successful retry exactly.
The contention marker proves attempted entry while the writer owns the lock;
it does not measure residence inside SQLite's native busy handler. Receipt
comparisons normalize JSON representation without dropping fields.

On Linux x64 with Node 24.22.1, all four tests passed three consecutive runs:
3.429s, 2.704s and 2.729s. Two separate suite processes also passed concurrently,
4/4 each, in 2.928s and 2.963s. The original first run passed 4/4 before diagnostic
and cleanup tightening. Existing store/integrity tests passed 61/61 in 67.639s.
The full runtime suite passed 325/325 in 73.456s with no skips. Pinned dependencies
were installed with scripts disabled and offline mode using a private npm cache.
These timings are observations, not performance guarantees. Wait-for-marker,
paused-child, child-watchdog and per-test limits are 10s, 20s, 25s and 40s.

Four disposable negative controls failed at their intended checks: a bypassed
COMMIT hook reported the missing requested boundary; premature commit broke the
prior-row snapshot assertion; altered duplicate receipt content differed from
the committed original; omitted acknowledgment persistence retained CLAIMED
instead of DELIVERED. A pristine disposable copy then passed 4/4 in 2.994s.
Initial disposable copies omitted a transitive Project Memory import and failed
at module loading. Those setup failures are retained separately and are not
negative-control evidence; the corrected copies produced the results above.
No production file in the maintained checkout was changed. Original independent
proof script/expectation hashes still match, and no crash-fixture directories
remained after the runs, including the deliberately failed controls.

Repository validation, both new files' Node syntax and whitespace checks passed.
Optional skills-ref is unavailable; GNU Stow installer acceptance and other OS
qualification were not run locally. Full committed Project Memory validation and
remote head/tree verification follow completion. Fresh independent exact-head
review and hosted CI remain separate gates. This draft does not claim a source
bug, provider execution, power-loss durability, real-use acceptance or activation.
