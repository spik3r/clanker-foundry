# Managed execution rollout

## Scope and authority

The selected scope is new `CF-AUTO-` tasks only. Existing `CF` records and claims
stay under their current workflow. Nothing in this change publishes tasks, starts
a background service, stages user work, commits candidates or transfers old claims.

Use one private local authority database across all worktrees for this repository.
Do not place it inside a candidate workspace or copy it to create another writer.
The database UUID binds decisions to that database; it cannot fence an independent
manual tracker or copied database. Operators and agents must use the same authority.

The [namespace template](../runtime/config/namespace-cf-auto.json) is deliberately
inactive. Its reserved sample UUID does not match a real database. Configure an
accountable owner and actual database identity, review and commit that exact JSON
in the source repository, then record its actual revision/path/blob pin. Keep the
source commit available locally. Source checks deny transports and do not fetch.

```sh
node runtime/src/cli.mjs authority-id /absolute/private-state/authority.sqlite
node runtime/src/cli.mjs authority-record /absolute/private-state/authority.sqlite shadow-request.json operator
node runtime/src/cli.mjs authority-status /absolute/private-state/authority.sqlite cf-auto
```

`shadow-request.json` contains `decision`, `source` (repository, revision, path,
blob) and `source_root`. The committed JSON must equal `decision`. First record
`SHADOW` with expected generation zero. Commit a separate reviewed decision naming
`MANAGED`, the current generation and explicit prior-authority drain confirmation;
record its exact pin. Do not claim that publishing a config template activates it.

Register each exact reviewed specification and actual canonical task source with
`store.registerAdmission({namespace_id,generation,spec,source,source_root,reviewed_by},auth)`
before `ADMIT`. `authority-admission DATABASE REQUEST_FILE ACTOR` exposes that API
for a trusted local operator. The source is a genuine registered Project Memory
snapshot verified against its pinned Git task records, not arbitrary JSON text.

Modes are `SHADOW → MANAGED → DRAINING → MANUAL`. Draining prevents admissions,
assignments and fresh launches, while existing results, validation, cancellation
and cleanup can finish. Rollback to manual requires terminal executions, confirmed
cleanup, released reservations and no unresolved outbox effects. Re-entry requires
a new pinned decision and generation. Unknown outcomes require owner inspection.

## Provider and host boundary

`createResponsesProvider({model,apiKey})` uses the fixed official Responses endpoint,
no tools, no stored response and no prior response ID. Configure credentials in the
controller environment. The live API check requires an API key; the existing Codex
login is a separate authentication path.

`createCodexCliProvider({cwd,model?})` uses the installed signed-in CLI in a private
empty directory with user configuration, rules and integrations disabled. It
accepts completed assistant text and the exact diagnostic saying Code Mode fails
closed because its host is disabled. Other error/tool/unknown events fail. This
is supervised text transport; do not treat integration flags as an OS sandbox.

```sh
node runtime/test/provider-live.mjs --provider codex
node runtime/test/provider-live.mjs --provider responses --model YOUR_MODEL
node runtime/test/isolation-live.mjs
```

The CLI check uses an invented prompt and prints only result metadata. Direct API
and CLI tests are distinct; a passing injected-fetch unit test is not a live API
result. Provider output remains nondeterministic text; it cannot choose transitions,
worker identity, scope, checks or successor agents.

Writable candidates use a digest-pinned Node image and a static plan applier.
Docker runs without network, capabilities, privilege escalation or a writable root,
under an unprivileged user with process and resource bounds. It mounts only a
private snapshot of explicitly reviewed files. No home, metadata, database,
credentials, sibling worktrees or Docker socket enter the worker. The Docker daemon
and controller are trusted host components. Exact allowlists still require content
review: an innocently named input file can contain confidential text.

`createCandidateWorkspace` snapshots committed Git blobs. The model returns a
strict `WorkspacePlan`; exact scope and before-digests constrain writes and deletes.
The host produces changed-file artifacts. An owner separately promotes the accepted
candidate after checking the target is still at the pinned base and its bytes have
not changed. Promotion does not stage, commit, publish or run repository hooks.
Reservations remain held until actual host termination is confirmed.

## Evidence and embedding

Create a genuine canonical task/source/start using Project Memory before edits.
The bridge does not create these facts or acquire ownership. It binds one execution,
specification, source, claim, run, start, frontier and exact file set before assignment.

Create the store with trusted closures, then initialize the bridge and workspace:

```js
let bridge;
let workspace;
const store = openStore(databasePath, {
  worker_ids: ['builder', 'reviewer'],
  approver_ids: ['operator'],
  trusted_checks,
  evidence_guard: (binding, execution) => bridge.guard({binding, execution}),
  isolation_guard: (execution, attempt) => workspace.guard({
    verified: true,
    execution_id: execution.execution_id,
    spec_digest: execution.spec_digest,
    attempt_id: attempt.attempt_id,
    workspace: attempt.workspace,
    repository_revision: execution.spec.repository.revision,
    write_scope: execution.spec.write_scope,
  }),
});
```

These are trusted embedding functions, never task-provided callbacks. Initializing
without fresh guards fails closed for managed dispatch. `store.verifyDispatch`
rechecks before acknowledgement and immediately before invocation. The dispatch
claim marks work as potentially in flight; a drain cannot undo a process already
launched after an earlier check. Cancel and reconcile it through the normal path.

`createProjectMemoryBridge({store,codeRoot,sourceRoot,journalPath})` uses a separate
private SQLite journal outside workspaces. `prepare(input,auth)` binds the genuine
start. `complete({execution_id,summary,validation,next})` requires accepted runtime
completion, actual final staged blobs and explicit observed checks. It records
intent before calling the unchanged helper. A lost reply is resolved through the
sole exact immutable completion child; absence or ambiguity blocks blind retry.
`verify({execution_id,base})` checks the supplied actual base and fresh committed
coverage for this exact execution. Another task's completion cannot substitute.

The bridge cannot prove a declared validation command ran. The trusted validator
must record the observed result. It also cannot make SQLite and Git one atomic
transaction; its journal makes uncertain outcomes visible and recoverable.

## Verification record

The signed-in Codex CLI completed the invented text smoke locally. The output was
24 bytes with SHA-256 `105306f4e23c2f4981bcf24a1cb48dad3daefa1dcc491c946d038299a964f3bd`.
No API key is configured, so the direct Responses transport has no live API proof.

Docker Desktop reports server version 27.5.1. The installed test image is
`node@sha256:8a32c8cf649a49915a66e81e7a609223c708d0adabf288ee3f4077d765542f85`.
The actual Docker smoke passed: the fixed capability probe confirmed UID 1000,
read-only root, disabled networking and absent host/socket access. Two typed file
operations produced one canonical plan artifact, followed by explicit owner
promotion in the invented temporary repository. The host inspected termination
before removing its containers.
Remote CI has not run for this unpublished branch. No persistent operating
namespace has been activated by a test fixture or the inactive configuration template.

Final local validation passed under Node v24.22.2: 154 runtime tests and 143
Project Memory/workflow regressions, with no skips. Pack and installer validation
passed. A fresh-context reviewer rechecked the fixes and found no remaining
material defect. The optional `skills-ref` tool is not installed; the pack validator
reported that optional check as skipped. Committed evidence validation is recorded
in the immutable completion records and checked after committing.

## Follow-up: admitted limits and dispatch preflight

A subsequent fresh review reproduced two gaps despite the earlier passing tests:
65 dependency outputs exceeded the runner's 64-subject limit after acknowledgement,
and a legal long rework workflow exhausted the 128-entry attempt history before
using its configured budget.

The follow-up supports the full admitted subject envelope through both runner and
text adapter and checks immutable inputs before claiming or acknowledging a launch.
The 1 MiB content bound remains. A deterministic input rejection leaves a pending
effect with an explicit skipped-dispatch error, rather than an unknown launch.

History capacity now covers the schema's maximum automatic attempt envelope of
356. A regression exercises all 16 stages, 20 reworks and 20 execution retries,
then replays the complete transcript. Extra manual recovery beyond this capacity
has an explicit rejection. No old event, source or task claim is rewritten.

Follow-up validation under Node v24.22.2 passed all 160 runtime tests, with no
skips. Pack validation passed (optional skills-ref unavailable). Independent review
found and rechecked a fairness regression: preflight rejections must not consume
the dispatch limit and starve later work. Its actual SQLite regression now proves
later launch and cleanup proceed with a limit of one. The final review reported no
remaining material findings. Live provider and Docker checks were not repeated for
these coordinator-only changes.
