---
id: CF-task-workflows
category: workflow
status: in-progress
owner: Foundry task workflow implementation
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Add reusable task, backlog and ownership workflows

## Implementation claim

Owner: Foundry task workflow implementation, branch codex/task-coordination-workflows.
The user approved adding generic task tracking, coordination and handoff guidance
for people adopting the pack in their projects. Reuse existing Project Memory and
handoff conventions; do not import another project's tasks or operating rules.

Write scope: skills/task-management/**, skills/agent-coordination/**,
skills/backlog-maintenance/**, docs/task-workflows.md, README.md,
templates/project/project-agents.md, templates/project/new-project-checklist.md,
skills/context-offload/SKILL.md, skills/checkpoint/SKILL.md,
skills/orchestrator/templates/handoff.md, workflows/feature-delivery.md,
and this task record. Immutable Project Memory metadata accompanies the work.
No lifecycle helper, CI, global instructions, external tracker or deployment edits.

## Acceptance

- A person can adopt task tracking in an existing project without scaffolding an app.
- Each skill works alone and follows the project's existing task store.
- Task identity, ownership, handoffs, validation and closeout have clear homes.
- Concurrent claims are verified at the shared source; local notes do not lock work.
- Project Memory stays an optional evidence layer for consumers, with its source
  format and limits stated when relevant.

## Validation

Run pack validation, installer regressions, Project Memory committed-diff validation,
and whitespace checks. Review a solo file-backed task, an external issue tracker,
an overlapping claim, and an interrupted handoff against the written procedures.

## Integration

Deliver a local reviewed change. Publication and merge require their own authority.
Record actual results and unresolved work here.
