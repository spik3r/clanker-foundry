# Deterministic task execution: architecture and implementation plan

Date: 2026-10-30. Reviewed base: `7a03b095eff59e76b98354673aec5d538ef76233`
(`origin/main`, including task-workflow adoption PR #8). Previous review: `a25e642`.
Planning task: [CF-deterministic-orchestration](../tasks/CF-deterministic-orchestration-plan.md).
Status: proposed design; no runtime, schema implementation or migration is included.

## Recommendation and changes to the previous proposal

Keep the standalone skill pack and introduce deterministic execution as an opt-in
mode for projects that need automatic assignment and recovery. Start with typed
contracts and a pure transition evaluator. Add a single-host transactional runner
only after its authority, host capabilities and operating costs are agreed.

The latest main now includes [task management](../skills/task-management/SKILL.md),
[agent coordination](../skills/agent-coordination/SKILL.md),
[backlog maintenance](../skills/backlog-maintenance/SKILL.md), and the
[adoption guide](task-workflows.md). The earlier statement that these were absent
from main is superseded. They clarify scope, shared claims, transfer acceptance,
integration evidence and tracker choice. They still describe procedures rather
than an executable scheduler.

Revise three parts of the earlier recommendation:

1. Do not replace every project's tracker or require a service to install skills.
   Define one authority per adoption mode before implementing persistence.
2. Do not restructure the existing `tasks/` directory into JSON specs and views by
   default. Its committed Markdown records remain Project Memory inputs. Use an
   explicit import/export boundary for projects choosing managed execution.
3. Treat SQLite and Node as a proposed single-host implementation, not settled
   requirements. The core contracts and transition rules must survive another host
   or provider. A host without enforceable worker isolation cannot promise safe
   unattended writes merely by changing its prompts.

## Facts, assumptions and scope

Facts: this repository is an asset pack with a local evidence CLI. There is no
queue service, launch adapter, task-schema package or transactional assignment
store. Its four agent roles already fit explore, plan, build and independent
review. Model tiers are separate from provider names in
[the model mapping](../agents/README.md).

Planning assumptions: one operator, one shared host, multiple worker processes
and worktrees, durable local storage, and modest task volume. Neither throughput
nor multi-host requirements have been measured. No production incident trace was
provided; the findings below identify enforcement gaps visible in the code.

Goal: make admitted commands, ownership, routing, recovery and completion follow
explicit rules. Given identical model versions and ordered accepted inputs, the
engine must reproduce state and dispatch intents. Agent output, provider failures,
human decisions and elapsed time are external inputs, recorded before evaluation.
This does not promise identical generated code across invocations.

Out of scope: application implementation in this change, mandatory consumer
rollout, arbitrary BPMN execution, distributed consensus, automatic publication,
merge or deployment, and rewriting existing lifecycle history.

## Current behavior and material findings

| Finding | Evidence | Impact and action |
|---|---|---|
| High: scheduling relies on interpretation | [Orchestrator](../skills/orchestrator/SKILL.md), procedure steps 2, 5, 8 and 9; [plan template](../skills/orchestrator/templates/plan.md), work graph | No executable dependency, retry or closeout guards. Define one transition model before adding dispatch. |
| High: claim checks are not atomic reservations | [Coordination](../skills/agent-coordination/SKILL.md), Before editing steps 2-4 explicitly acknowledge races | Two coordinators can see the same available task. Managed mode needs transactional assignment and resource reservation. |
| High: event publication is not task serialization | [CLI](../skills/project-memory/scripts/project-memory.mjs), lines 117-158; [immutable writer](../skills/project-memory/scripts/lib/git.mjs), lines 106-130 | Concurrent callers can publish different children of the same head. Preserve fork detection; do not use this helper as the scheduler lock. |
| High: results lack enforceable attempt identity | [Handoff](../skills/orchestrator/templates/handoff.md) has canonical task, workspace and scope, but no typed attempt/token | Late or duplicate reports have no machine-level acceptance contract. Add fenced attempts and typed results. |
| High: evidence is not executed acceptance | [Completion](../skills/project-memory/scripts/project-memory.mjs), lines 152-154; [event validation](../skills/project-memory/scripts/lib/state.mjs), lines 209-218 | A nonempty validation declaration is not proof of a check. Introduce trusted check execution and revision-bound results. |
| Medium: review and retry contracts differ | [Review template](../skills/orchestrator/templates/review.md), [reviewer profile](../agents/reviewer-verifier.md), [delivery workflow](../workflows/feature-delivery.md) | Verdict vocabularies and loop limits are not uniform. Use typed verdicts and versioned retry policy. |
| High for unattended execution: scope is instructional | [Worker profiles](../agents/README.md) and handoff write scope | An instruction cannot prevent tool use. Enforce capabilities and isolate candidate artifacts before integration. |

Keep Project Memory's strict event fields, causal checks, immutable publication,
single-frontier completion, exact Git blob/mode coverage and offline operation.
Keep the new skills' distinction between ownership, plans, handoffs and evidence.
No existing canonical task should be closed or reassigned from its age or a local
completion event.

## Authority and adoption modes

| Mode | Authority | Engine behavior |
|---|---|---|
| Existing workflow | Existing task files or tracker own lifecycle and claims | Skills operate as today; no runtime required. |
| Shadow evaluation | Existing store remains authoritative | Import reviewed snapshots, evaluate proposals, compare decisions; no dispatch or ownership mutation. |
| Managed execution pilot | Opted-in task aggregates in the coordinator own lifecycle and active assignment | Existing tracker entries/Markdown become labeled projections for those tasks. Edits request commands; they cannot independently change live state. |

Use a recorded cutover per task namespace. Preserve stable task IDs and links;
drain or explicitly transfer existing claims before enabling managed dispatch.
Do not have editable Markdown and a database independently own the same status.
If a tracker must remain authoritative, defer autonomous assignment until its
conditional-update and claim semantics support the required guarantees. Never
pretend a local database reservation also reserves work in an external tracker.

Separate accountable owner from the current execution worker. Assigning a worker
does not transfer human accountability or grant merge authority. Ownership transfer
requires the receiver's recorded acceptance and the authorized actor's command.

## Platform-independent contracts

Use one versioned JSON Schema contract as the wire format and shared fixtures as
the conformance oracle. Runtime types are derived or checked against that contract.
Node fits the current helper; Python/Pydantic remains an alternative. A Python
implementation must use strict validation and reject extra fields. Do not maintain
two independently authored schema definitions.

| Type | Required fields and constraints |
|---|---|
| `TaskSpec` | `schema_version`, stable namespaced `task_id`, immutable `spec_revision` and digest, outcome, repository identities, dependency refs, workflow ref, exact write scope, shared resource refs, acceptance check refs, accountable owner and integration policy |
| `WorkflowDefinition` | Version/digest, stages, legal transitions, guard IDs, role/capability requirements, join policy, deadlines, bounded retry policy and explicit completion gates |
| `TaskExecution` | `execution_id`, task/spec/workflow refs, state, stage, monotonically increasing revision, active attempt ref, blocker/retry details and terminal outcome |
| `ExecutionAttempt` | Unique ID, stage/attempt number, worker identity, input artifact/base revisions, isolated workspace, capability grant, lease deadline and increasing fencing token |
| `Command` | Schema version, command ID, actor identity, command kind, target execution, expected revision and typed payload; authority comes from authenticated transport, not asserted JSON |
| `WorkerResult` | Attempt ID, fencing token, spec digest, discriminator, artifact manifest and structured findings; worker cannot set execution state |
| `ValidationEvidence` | Check/version, trusted runner identity, subject artifact digest, repository revision, attempt/spec/workflow refs, verdict, time and bounded evidence reference |
| `Approval` | Authorized actor, action, execution/spec/workflow refs, exact artifact subject, decision and expiry/revocation rules |
| `TaskEvent` | Event ID, command ID, ordered sequence, prior/new revisions, actor, transition/reason and referenced artifacts; schema and evaluator versions are recorded |
| `TaskProposal` | Origin task/attempt, proposed outcome, scope, dependencies and acceptance; admission must approve before it can execute |

`WorkerResult` is a discriminated union: `SUCCEEDED` requires artifacts,
`CHANGES_REQUIRED` requires findings against a subject, `BLOCKED` requires a
blocker code/resolver, and `FAILED` requires a failure code. A result cannot contain
fields from another variant. Review verdicts use one vocabulary:
`PASS`, `CHANGES_REQUIRED`, `INCONCLUSIVE`; only policy converts them to transitions.

Reject unsupported versions, duplicate JSON keys, unknown properties, implicit
coercions, duplicate IDs, unbounded payloads and noncanonical paths. Validate
dependency existence/cycles, permissions and cross-task conflicts separately;
schema validation alone cannot establish them. Canonical digest rules must specify
encoding, serialization and array ordering in the contract fixtures.

## Execution model and transition guards

Separate execution state from stage: `stage=BUILD` and `state=EXECUTING` are
different facts. Initial workflow: explore -> plan -> build -> independent review
-> required integration gates. A deterministic check runner validates each stage
output. An LLM reviewer adds evidence; its judgment is still an external input.

| From | Input | Guard and effect | To |
|---|---|---|---|
| No execution | Admit | Valid and authorized spec, resolved references, acyclic graph | `PENDING` |
| `PENDING` | Assign | Dependencies accepted; worker eligible; resources acquired in same transaction | `ASSIGNED` |
| `ASSIGNED` | Acknowledge | Current attempt, token and lease | `EXECUTING` |
| `EXECUTING` | Submit output | Current attempt/token; schema valid; artifact manifest admitted | `VALIDATING` |
| `VALIDATING` | Accept stage | Required checks and approval pass for exact subject; another stage exists | `PENDING`, next stage |
| `VALIDATING` | Accept final stage | All required review, acceptance and integration gates satisfied | `COMPLETED` |
| `VALIDATING` | Request changes | Versioned rework policy allows another build; retire prior attempt | `PENDING`, build stage |
| Active state | Retryable error or recorded timeout | Fence attempt; revoke capabilities; schedule recorded delay | `RETRY_WAIT` |
| `RETRY_WAIT` | Recorded retry timer | Budget remains; readiness and resource safety rechecked | `PENDING` |
| Nonterminal | Block | Named blocker/resolver; fence and retire any active attempt, revoke capabilities, retain reservations until cleanup is confirmed | `BLOCKED` |
| `BLOCKED` | Resolve | Authorized resolution; prerequisites rechecked; no old attempt restored | `PENDING`, retained/resolution stage |
| Nonterminal | Permanent error or exhausted execution budget | Record reason; fence attempt and revoke capabilities | `FAILED` |
| Nonterminal | Cancel | Authorized command; fence attempt and revoke capabilities | `CANCELLED` |

Active means `ASSIGNED`, `EXECUTING` or `VALIDATING`. Validation itself has a
deadline and retry policy. Invalid or stale result messages are rejected without
changing state; they do not fail a newer attempt. Every other unspecified edge
is illegal. Terminal executions have no successors; resubmission creates an
explicitly linked execution with a new ID.

Allow at most one nonterminal execution per namespaced task ID, across all spec
revisions. Enforce this in admission and the database. Admitting a new spec does
not start competing work or mutate a running spec. Supersession requires an
authorized cancellation, confirmed effect cleanup, and then a linked execution
bound to the new spec. Retained resource reservations still guard terminal attempts
whose cleanup is incomplete.

Dependencies initially use all-success joins on pinned accepted execution outputs.
Failure or cancellation blocks dependants with a named reason. Parallel stages
are allowed only with disjoint resources and explicit joins. Defer first-winner
and quorum semantics. Record approvals before guarded dispatch; a changed subject
invalidates old approval and validation. Worker success never implies task success.

Retry policy distinguishes transport retries, execution retries and review rework.
Each has a finite configured budget, deadline and reason codes. Exhausted transport
or execution budgets fail the execution; exhausted review rework blocks for an
owner decision. The pilot must choose explicit numeric values before dispatch.
No hidden infinite repair loop or attempt-count-based completion is allowed.

## Transactions, dispatch and effect safety

Use a pure evaluator: current aggregate plus accepted input produces new state,
events and effect intents. The evaluator reads no clock, network, random generator
or environment. Generate IDs and capture time in the input boundary. Record worker
availability and policy changes. Replay reuses recorded effect IDs without sending
effects. A new workflow version applies only through explicit admission/migration.

For the single-host pilot, recommend one shared Node runner and SQLite database on
local durable storage, outside per-worktree Git state. Confirm the runtime/driver
in phase 0; no package is selected here. All worktrees use the same authority.
Do not put the database on a shared network filesystem or copy it into each clone.
Multi-host deployment needs a separate storage/leader decision.

One transaction checks authorization, idempotency and expected revision, evaluates
guards, reserves resources, updates state, appends the event, stores the command
receipt, and inserts outbox messages. Database constraints enforce unique active
assignment, one nonterminal execution per task, and event sequence. Concurrent
writers recheck under the transaction;
losers receive a revision conflict rather than guessing a winner.

Dispatch reads the durable outbox after commit. Delivery is at least once. Duplicate
command IDs with identical canonical payloads return the original receipt; reuse
with different payloads is rejected. Attempt launch deduplicates by attempt ID.
Effects use stable effect IDs; unknown outcomes are reconciled before a retry.
Do not claim exactly-once external side effects from database transactions alone.

Order eligible tasks by configured priority, admission sequence, then stable ID;
select eligible workers using a stable tie-break. Record timer firing as a command
from the trusted scheduler. No local worker clock or directory listing decides
ownership. Human scheduling overrides are authorized, versioned commands.

Leases describe execution liveness, not accountable ownership. Expiry fences late
results and revokes capabilities; it does not silently transfer someone's claim.
Retain conflicting resource reservations until process termination or effective
access revocation is confirmed. If uncertain, block recovery for operator action.
Fencing a database write cannot stop a rogue process editing a shared directory.

Workers get fresh invocation context, a frozen spec, pinned inputs and a capability
envelope. They return artifacts and typed results only to the coordinator. They
cannot call workers, assign tasks, change claims, or promote their own artifacts.
Task decomposition from an LLM yields proposals requiring admission, not executable
control flow. Trusted workflow edges decide successors.

Use isolated attempt workspaces plus enforced filesystem/tool permissions. Reviewers
have no write capabilities. Separate integration from building; recheck the current
base, reservations and approval before promotion. A changed merged tree requires
applicable validation against that resulting tree. Keep publication/deployment
behind their own authorization gates. An adapter unable to enforce the envelope
is eligible only for read-only or supervised operation.

## Task-directory and evidence boundary

Preserve current task paths and IDs. In managed mode, admit a reviewed Markdown
record through a strict importer into an immutable `TaskSpec`; pin its repository,
commit, blob and claim. Reject conflicting IDs across statuses. Do not execute a
file because an agent dropped it into a directory. New files are proposals until
admission, and exports carry a revision and generated-view label.

Read proposals with byte/size limits, duplicate-key rejection, canonical paths,
no symlink traversal, repository allowlists, dependency validation and authenticated
admission. Restrict check references to trusted registered commands; task text is
not executable shell. Resolve resource conflicts for paths and declared shared
interfaces, including different tasks modifying one API. Conservative conflicts
are preferable to unrecorded overlap in the first pilot.

The current Project Memory helper accepts committed Markdown sources. Keep that
contract, old source/events and legacy ledger bytes unchanged. Runtime completion
and Project Memory completion have different meanings. A future bridge exports
reviewed evidence and committed task projections in causal order; it never invents
starts after work or treats a projection as proof of current ownership. The runtime
must not depend on Project Memory performing network access or enforcing leases.
Bridge failure blocks a configured evidence gate, not rewrites history.

Before the bridge exists, pilot tasks using Project Memory follow its existing
manual lifecycle: commit the reviewed source and claim, register and verify the
pin, and record a genuine start before worker edits. Dispatch checks that setup;
integration checks fresh completion coverage. Do not postpone this requirement to
phase 6 or fabricate earlier events after a worker has run.

Maintain per-task traceability from requirement IDs to guard IDs, check IDs,
artifact subjects and outcomes. Expose current stage, blocker, attempt, lease,
retry budget and rejected command reasons. Redact secrets from evidence, bound log
retention, and use secret-store references in capability configuration. Back up
the state store and verify restore/replay before enabling unattended writes.

## Modeling concepts and limits

| Concept | Use in this plan |
|---|---|
| MBSE | Requirement -> model invariant -> transition -> check -> evidence traceability |
| MDA | Platform-independent contracts/evaluator; provider, persistence, clock, Git and tool adapters implement platform-specific behavior |
| BPMN | Explicit sequence, all-success joins, timers, error and approval paths |
| fUML | Typed object identity and explicit execution semantics; no claim of standards conformance |
| UTP | Model test stimuli, verdicts and coverage against transition requirements |
| SysML | Optional context and deployment views if people/infrastructure constraints justify them |
| Essence | Optional progress/health projections from evidence; never a second scheduling authority |

A schema and enum are not a conformant fUML execution engine. A workflow diagram
alone is not an executable BPMN model. Borrow bounded concepts first. Formal
conformance would need a separate supported-semantics specification and tests.
References: [OMG MDA](https://www.omg.org/mda/),
[BPMN](https://www.omg.org/spec/BPMN/2.0.2/About-BPMN),
[fUML](https://www.omg.org/spec/FUML/1.5/About-FUML),
[UTP](https://www.omg.org/spec/UTP2/2.2/About-UTP2).

## Ordered implementation slices

Paths below are proposals, not files created by this planning change. Each slice
needs its own bounded task/claim and genuine evidence lifecycle before edits.

| Phase | Proposed scope and deliverable | Depends on | Observable exit gate |
|---|---|---|---|
| 0: settle boundary | `docs/adr/` decision on runtime home, managed authority, single-host limits, adapter capabilities, numeric budgets and storage/runtime choice | Owner decisions | One adoption mode chosen; no conflicting owner; read-only fallback defined |
| 1: contracts | `schemas/orchestration/`, invented fixtures under `tests/fixtures/orchestration/`, contract guide | 0 | Positive and negative fixtures define types, digest rules, compatibility and every rejection |
| 2: semantics | Proposed `scripts/orchestration/core/` pure evaluator and `tests/orchestration-state.test.mjs` | 1 | Legal/illegal edges covered; replay identical; invariants checked over generated command sequences |
| 3: durable runner | Proposed `scripts/orchestration/store/`, command API, receipts, resource reservations, timers and outbox tests | 2 | Concurrent assignment has one winner; crash/restore tests preserve receipts and dispatch intents |
| 4: fake-worker pilot | Fake provider, capability broker, candidate artifacts, check runner and independent review contract | 3 | Duplicate/stale outputs rejected; unsupported capabilities fail closed; effect-unknown recovery demonstrated |
| 5: real adapter | One host/provider adapter, opt-in execution guide and supervised single-task pilot | 4 | Full explore/plan/build/review trace passes with enforced isolation; cancellation and recovery verified |
| 6: adoption and evidence | Strict importer/exporter, optional Project Memory bridge, instruction/template changes only where required | 5 | One namespace cut over with preserved IDs/history; restore and return-to-manual drill passes |

If phase 0 chooses a separate runtime repository, move proposed implementation
paths there before claiming later phases. Keep only contracts/guidance appropriate
to this pack. Preserve independent installation of existing skills.

Test scenarios: two assignments race; overlapping resource scopes use different
task IDs; crash after commit before dispatch; duplicate launch/result; lease expiry
followed by late output; cancel races completion; unknown external effect outcome;
dependency absent/cyclic/failed; wrong-revision review; artifact mutation after
validation; approval revoked or stale; retry/rework exhaustion; unsupported host
capabilities; direct worker handoff rejected; restore and replay produce the same
state without redispatch; old Markdown/evidence inputs remain usable.

Use a small reference transition model plus generated command sequences, fault
injection and adapter contract fixtures. Generated tests do not replace independent
acceptance checks for guard correctness. Add CI only with the corresponding code;
required branch protection is an owner decision, not a side effect of this plan.

## Alternatives, open decisions and next action

| Option | Benefit | Cost and recommendation |
|---|---|---|
| Strengthen prompts only | Lowest adoption cost | Keeps the observed race/enforcement gaps; retain for manual use, not unattended coordination |
| Small bounded execution engine | Direct fit for current roles and explicit state rules | Requires persistence/recovery ownership; recommended incremental pilot |
| Full BPMN/workflow engine | Existing execution tooling and richer process support | Larger dependency and operating surface; revisit when required features justify it |

Owner decisions before runtime work: one host or several; runtime in this pack or
a separate repository; which store owns opted-in tasks; who may approve and
transfer claims; which host can enforce capabilities; retry/rework budgets; what
integration or deployment evidence each task requires. These do not prevent
reviewing this plan but can change implementation scope.

The previous baseline run reported pack validation passing and 142/143 lifecycle
tests passing, with an `EPERM` creating an invalid UTF-8 filename fixture. That is
historical evidence, not validation of this updated branch. This planning change
does not modify helper/CI code or establish runtime correctness. Current observed
documentation checks are recorded in the planning task and its completion event.

Next action: owner review of phase 0 decisions, then claim only the agreed first
implementation slice. This document neither enables a scheduler nor authorizes
publication, tracker migration, merge or deployment.
