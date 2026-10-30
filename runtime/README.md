# Deterministic orchestration pilot

An opt-in local runtime for typed tasks, guarded state transitions and durable
worker dispatch. Existing skills and manual task workflows work without it.
This is a single-host read-only pilot, not an unattended code-writing service.

## Run

Use Node 24 or later and local storage. From the repository root:

```sh
npm ci --prefix runtime --ignore-scripts
npm test --prefix runtime
node runtime/src/cli.mjs demo
```

The demo creates a private temporary database and runs fake explore, plan, build
and independent-review stages. It records check results, simulated owner approvals,
cleanup and replay. It prints the database location; remove that demo directory
when finished. It performs no provider calls or repository writes.

An explicit database path must be absolute, in a real private directory owned by
the current user (mode `0700`). Database files use mode `0600`. Keep SQLite, WAL
and journal files out of Git and off network filesystems. Symlinks and unsafe
existing paths are rejected. On macOS, use the resolved `/private/...` path when
the path would otherwise traverse `/var` or `/tmp` symlinks.

## Authority and permissions

Callers are trusted local programs with access to the private database. The API
accepts an authenticated context from its embedding process; it does not implement
a network authentication service. Supplying a role to the CLI is an operator
action, not proof of remote identity. Do not expose this API to untrusted users.

The pilot admits only `read-only-v1` workflows with empty write scope and
`integration_policy: read_only`. Tasks must state finite retry/rework budgets,
lease and validation durations, ordered stages, acceptance checks and accountable
owner. Admission reserves no work in an external tracker. Current project tasks
and their claims are unchanged until an explicit cutover outside this pilot.

Workers acknowledge and submit results for their current attempt. They cannot
assign workers, expand scope, change task ownership or accept their own stages.
Owner commands control stage acceptance, cancellation and recovery. A separate
registered validator must supply evidence; configured approvers handle required
subject-bound approvals. Reviewer identity must differ from prior builders.

Task acceptance checks run against the last non-reviewer candidate stage (or the
final report for a standalone review). Stage-specific checks validate that stage's
own output. A workflow containing a builder must end with independent review;
passing checks on a review report cannot substitute for checking the candidate.

## Modules and wire contract

- [Contract](contracts/v1.schema.json): versioned JSON Schema with strict types,
  discriminated commands/results, extra-field rejection and bounded collections.
- [Parser and canonicalization](src/contract.mjs): duplicate-key rejection, UTF-8
  string rules and deterministic digests. Object keys sort; array order matters.
- [Evaluator](src/state.mjs): pure transitions with supplied clock/identity inputs.
- [Store](src/store.mjs): SQLite transactions, reservations, immutable specs,
  artifacts/subjects, receipts, history, outbox and replay.
- [Runner](src/runner.mjs): executes typed effect intents through trusted adapters
  and registered checks. It never approves or accepts stages automatically.
- [Adapters](src/adapters.mjs): fake worker, allowlisted workspace reader and
  injected text-only model callback. No bundled live provider client is included.
- [Import/export](src/import-export.mjs): reviewed Project Memory snapshots become
  proposals; generated Markdown remains an inert view.

## Embedding

```js
import { openStore } from './runtime/src/store.mjs';

const store = openStore('/absolute/private-directory/authority.sqlite', {
  worker_ids: ['research-worker', 'review-worker'],
  approver_ids: ['operator'],
  trusted_checks: [
    { check_id: 'artifact-integrity', check_version: '1', runner_id: 'validator' },
  ],
});
try {
  // command is validated against the Command contract.
  // Identity comes from the trusted embedding process, never command JSON.
  const receipt = store.execute(command, { actor_id: 'operator', role: 'owner' });
  console.log(receipt.execution.state);
} finally {
  store.close();
}
```

`store.execute` commits state, command receipt, event, resources and effect intents
in one transaction. Repeating an identical command with the same principal returns
its original receipt. Reusing its ID with different content or identity fails.
One nonterminal execution is allowed per task ID across specification revisions.
Assignment requires satisfied pinned dependencies, registered workers and available
resources. Conflicting reservations survive cancellation until confirmed cleanup.

`putArtifact({digest,content})` admits bounded text only after verifying SHA-256 of
its UTF-8 bytes. `getArtifact(digest)` and `getSubject(digest)` validate persisted
content/manifest hashes. A successful result cannot reference missing bytes.
Subject digests hash the canonical artifact manifest, not the output text itself.
Reviewers receive the actual accepted candidate bytes rather than a digest alone.

`history()` includes accepted commands and their recorded evaluation context.
`replay()` verifies transitions, events, outputs and final state without sending
effects. Restore the entire database consistently; test recovery before using it
for any work that matters. A local event log is not a backup or external claim.

## Workers and validation

Use `drainOutbox({store,adapter,checks,dispatcherId,now,limit})` with a bounded adapter
and a trusted check registry. The runner returns delivered, unknown and skipped
effect summaries. A callback check receives the subject and stored artifact bytes;
task Markdown never supplies executable shell commands.

`createTextModelAdapter` takes an injected `invoke` function, worker ID, explicit
input paths and a bounded reader. The invocation receives frozen task/input data,
including accepted predecessor outputs. The adapter hashes and retains returned
text; the runner persists it before submission. The callback is trusted host code,
not an isolated process. It must use a stateless text API without tools or external
side effects. Model/provider choice and credentials belong in that embedding
configuration, outside task records and logs. A live vendor call has not been
verified by this pilot's fake tests.

Each adapter instance allows 16 active invocations or undelivered outputs. Further
invocations wait before calling the model. The runner persists output bytes and
then calls `releaseArtifact(digest)`; custom callers must do the same. Identical
output digests use reference counts, so one consumer cannot discard another's
pending output.

The workspace reader allowlists exact paths, rejects symlinks, bounds total bytes
and checks the pinned local Git revision. It is not a sandbox for arbitrary code
or a guarantee against an untrusted concurrent filesystem writer. Use immutable
input snapshots or host isolation if other processes can alter its workspace.

## Recovery

Dispatch is at least once at the outbox boundary. Claims fence concurrent
dispatchers. An expired in-flight effect becomes `UNKNOWN`; the runner does not
silently relaunch it. A recorded cancellation or superseding stage suppresses
obsolete launch/check intents. Unknown outcomes for one task must not starve
eligible effects for another task.

Only the accountable owner may call `reconcileEffect` after checking the actual
outcome. `DELIVERED` confirms it occurred; `PENDING` confirms it never occurred.
Neither is a blind retry. Revocation must match the assigned adapter/attempt and
confirm termination. If the host cannot prove that an invocation stopped, keep
the reservation and resolve the uncertainty manually. A fencing token rejects
late messages but cannot stop an external process by itself.

## Task intake and CLI

```sh
node runtime/src/cli.mjs --help
node runtime/src/cli.mjs validate TaskSpec spec.json
node runtime/src/cli.mjs command /absolute/private-directory/authority.sqlite policy.json command.json operator owner
node runtime/src/cli.mjs show /absolute/private-directory/authority.sqlite execution-id
node runtime/src/cli.mjs export /absolute/private-directory/authority.sqlite execution-id
node runtime/src/cli.mjs replay /absolute/private-directory/authority.sqlite
```

`import REVIEW_INPUT_FILE` takes JSON with `source`, `spec`, `reviewedBy`,
`expectedRepository` and `expectedRevision`. `source` is a reviewed Project Memory
source snapshot, not a tracker URL. The importer checks the exact pin, task/code
identity, claim shape and accountable review before emitting `mode: PROPOSAL`.
The returned proposal is not an automatically admitted command. A hash-valid
snapshot does not authenticate the remote source or prove current ownership;
the operator must review those facts. No backlog text is copied into the proposal.

Project Memory remains the repository's separate implementation-evidence process.
Record the genuine source/start before edits and fresh completion after checks.
There is no automatic evidence bridge, task migration, publication or deployment.
See the [architecture plan](../docs/deterministic-orchestration-plan.md) and
[runtime decision](../docs/adr/0001-deterministic-runtime.md) for scope and limits.
