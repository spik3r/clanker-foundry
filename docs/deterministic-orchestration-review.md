# Fresh-context orchestration review and corrections

Date: 2026-10-30. Reviewed implementation: `220afbf`, against `origin/main`
`7a03b09`. Follow-up task:
[CF-orchestration-review](../tasks/CF-orchestration-review-followup.md).

A Sol reviewer started without conversation history and inspected the full branch:
runtime code, schemas, tests, CI, task evidence, ADR, architecture plan and pack
integration. It reproduced seven defects in isolated local fixtures. A separate
Luna investigation checked the existing filename-test failure. The lead and a
bounded Luna builder implemented corrections; the original reviewer independently
rechecked every correction.

## Findings addressed

| Severity | Finding at the reviewed implementation | Correction and regression |
|---|---|---|
| High | Parsed multi-resource/dependency specs could throw inside Ajv's prototype-sensitive object-array comparison. | Validate a safe canonical JSON copy while preserving original records and the parser's null-prototype behavior. Test parsed resource/dependency arrays, duplicates, accessors and persisted assignment/replay. |
| High | A fresh fake adapter could report an unknown real invocation as stopped, releasing its reservations. | Require recorded, settled ownership with matching worker, attempt and fencing token. Test a still-pending text invocation, wrong identities, fresh instances and legitimate cleanup. |
| Medium | Owner-confirmed absent launch could not resume after a durable acknowledgement; it was suppressed as stale. | Check adapter configuration before acknowledgement. Persist owner absence confirmation and permit only that exact live acknowledged attempt to resume, without a second acknowledgement. Test absent configuration and post-ack recovery. |
| Medium | Replay applied a 1 MiB external-request limit to its combined database response. | Keep external request limits while using an explicit size for individually validated internal replay records. Test 200 legal tasks whose combined response exceeds 1 MiB. |
| Medium | A legal aggregate could cross 1 MiB through small commands, causing the next command's stored-record/receipt round trip to fail. | Separate internal stored-record parsing from the external wire limit at every store read/round trip. Test a legal 16-stage task growing past 1 MiB in nine commands, reopening it, retrying its receipt and replaying its history. Oversized new commands still fail before mutation. |
| Medium | A stage's 256 required checks plus a distinct task check passed admission but exceeded effect limits later. | Reject an effective check union over 256 at admission; test overflow and the valid overlapping boundary. |
| Medium | Legal long check identities produced default evidence references that violated the reference contract. | Generate bounded digest references and enforce the existing 128-character custom-reference limit. Test maximum-length registered check IDs and versions. |

The existing invalid UTF-8 filename test now writes raw path bytes to a Git index
and commits the resulting tree instead of creating that filename on the host
filesystem. Its replacement-character deletion snapshot and non-UTF-8 rejection
assertion remain. No test is skipped and the helper's production code is unchanged.

The cleanup regression that previously used an unrelated fake adapter now uses
the adapter that actually ran the synthetic attempt. A separate regression asserts
that unowned cleanup remains pending. This removes a false-success assumption
rather than weakening the check.

## Validation and review limits

Independent rechecks passed all seven original reproductions, including eight
positive/negative recovery scenarios. The bounded growing-record regression uses
the existing 16-stage schema limit; no limit was widened to make a test pass.

Node v24.22.2 is used for final checks. The previously failing lifecycle fixture
and both Project Memory suites passed after its correction (143/143, no skips).
The combined final run passed 227/227 tests: 84 runtime and 143 lifecycle/workflow,
with no skips. Pack validation and installer regressions passed. The pinned
dependency audit reports zero known vulnerabilities. The independent reviewer
found no remaining reproduced material defect after rechecking all seven fixes.
Staged whitespace and committed-evidence outcomes are recorded in the follow-up
task's immutable completion sequence.

These checks exercise local SQLite transactions, separate-process races, filesystem
fixtures and injected model callbacks. They do not verify a live vendor call or
an unattended writable worker. Remote CI has not run for this unpublished branch.

## Remaining broader-plan work

- Select and verify a live stateless provider integration in the embedding host;
  credentials stay outside tasks and logs.
- Establish and test host isolation before enabling writable workers. Until then,
  the runtime continues to reject write scope.
- Implement the optional evidence bridge if needed; existing manual Project Memory
  source/start/completion remains authoritative for implementation evidence.
- Obtain an explicit task-namespace cutover decision before any live tracker or
  ownership migration. This review performs none.

These are explicit integration or operator decisions beyond the local read-only
pilot, not completion claims for unbuilt features. Publication, merge and deployment
have not been requested by this follow-up.
