# Tasks, ownership and handoffs

These workflows can be added to an existing project. No app scaffold, separate
backlog repository or team of agents is required. Installing the skill pack makes
the procedures available; choose project conventions before making tracking mandatory.

## Start small

1. Keep your existing task tracker. With none, use `tasks/ID-description.md` in the
   code repository and adapt the [task template](../skills/task-management/templates/task.md).
2. Put the task-store location, lifecycle and claim convention in project `AGENTS.md`.
   Use a project prefix and stable IDs for file-backed tasks. Reserve IDs through the
   shared store before parallel work depends on them.
3. Choose one handoff convention: a replaceable task-local context file for short
   work, or checkpoint files for longer-lived decisions and assumptions.
4. Track a real change from scope through validation and review. Add machinery only
   when the current process leaves a demonstrated gap.

Tiny fixes and read-only questions usually need no new task or checkpoint. Project
rules may require them; make that choice explicit rather than inheriting it by accident.

## Choose a task store

| Store | Useful when | Cost |
|---|---|---|
| Files in the code repo | One repo; tasks reviewed with code | Claims on private branches are not shared reservations |
| Existing issue tracker | People already use issues and assignments | Agents need access; offline notes can become stale |
| Separate backlog repo | Several code repos share outcomes and dependencies | Separate changes and access to coordinate |

Use one authority for task identity, lifecycle and ownership. Plans explain how;
ideas capture possibilities. Neither needs another copy of task status.

## Use the right record

| Record | Owns | Update when |
|---|---|---|
| Canonical task | Scope, owner, claim, dependencies, acceptance and closeout | Work is claimed, blocked, reviewed, transferred or accepted |
| Plan or worker brief | Ordered steps and bounded instructions | An approach or dependency changes |
| Context/checkpoint | Resume state, decisions, assumptions and next action | Pausing, handing off or reaching a useful milestone |
| Optional Project Memory | Pinned intent and local implementation evidence | Its installed lifecycle requires an event |

Handoffs link the task; they do not reserve work. On resume, verify both current
repository state and the live claim. A single root `context.md` or `current.md` suits
one active effort. Concurrent efforts should use separate task directories, such as
`checkpoints/APP-001/current.md`, so they cannot overwrite each other's notes.

## One change through the workflow

For “reject expired invite links”, define the observable rejection, relevant files,
expiry boundaries and validation before claiming the task. Check for an existing
task or PR. Claim the bounded scope through the shared store, then implement and test.

If stopping for the day, link the task from a handoff and record the branch, changed
files, actual checks and first incomplete action. The next owner verifies that state
and agrees the transfer before treating the claim as theirs.

When ready, record the checked revision and results; move to review. Close only
after acceptance and any required integration checks. If rollout is a required gate,
merging the code leaves the task awaiting that gate.

## Skills to use

- [Task management](../skills/task-management/README.md): scope, update and close tasks.
- [Agent coordination](../skills/agent-coordination/README.md): check overlap, claim
  bounded work, transfer ownership and integrate changes.
- [Backlog maintenance](../skills/backlog-maintenance/README.md): reconcile the queue
  against evidence; review mode does not edit records.
- [Context offload](../skills/context-offload/README.md): one compact resume file.
- [Checkpoint](../skills/checkpoint/README.md): durable decisions and assumptions.

Each skill works alone. The [feature delivery workflow](../workflows/feature-delivery.md)
connects the steps when implementation spans several phases.

## Optional Project Memory

[Project Memory](../skills/project-memory/README.md) adds local evidence; it does not
replace the task store, acquire ownership or prove a test ran. This pack uses it for
its own changes, but adopting these task workflows does not require enabling it in
another project.

The current helper supports committed Markdown sources under `tasks/ID-description.md`
with its documented frontmatter and nonempty claim. It does not consume arbitrary
issue URLs. If you use an issue tracker, continue with that tracker and handoffs
unless you deliberately choose a supported canonical source arrangement. Do not
create a second live ownership ledger just to enable evidence tooling.

For file-backed adoption, read the [source and lifecycle contract](../skills/project-memory/SKILL.md)
before setup. Commit the reviewed claim, register its pinned revision and record a
genuine start before implementation. Copy only public-safe records into public repos.
CI enforcement and required status checks are separate adoption choices.

## Project instruction example

Adapt this text to the project's real paths and rules; it is not enabled by installing
the pack. For an issue tracker, replace file-specific details with the tracker URL
and its shared assignment mechanism.

```markdown
## Tasks and handoffs

Canonical tasks: tasks/ID-description.md, prefix APP.
States: ready, in-progress, blocked, review, done, cancelled.
Track work that needs coordination or continuation; tiny solo fixes need no new task.
Check existing tasks and open PRs before starting competing work.
Publish a bounded claim through our shared task-review process before parallel edits;
a private branch is a proposal, not a reservation.
Keep owner and lifecycle in the task. Link it from PRs and checkpoints.
Handoffs: checkpoints/ID/current.md; verify repository and live claim before resuming.
Close tasks with validation and integration evidence. Preserve completed IDs and links.
Follow repository permissions for commits, publication, merge and deployment.
```
