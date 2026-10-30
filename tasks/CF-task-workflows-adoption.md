---
id: CF-task-workflows
category: workflow
status: review
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

## Local result

Added three standalone skills, a task template and an adoption guide. Linked task
references and live ownership checks into handoffs, delivery and optional project
instructions. Existing trackers remain authoritative; Project Memory remains an
optional consumer choice with a documented Markdown source requirement.

Validation: `bash scripts/validate.sh` passed, including relative links and skill
conventions. `bash tests/test-install.sh` passed. `git diff --check` passed.
Reviewed the written procedures for solo file-backed work, an external issue
tracker, competing claims and interrupted task handoffs. This was a manual review,
not a test with an independent agent or live external tracker.

The optional `skills-ref` tool is absent. The skill-creator `quick_validate.py`
could not run because the available Python lacks PyYAML. No helper, installer or
CI logic changed. Project Memory committed-diff validation runs after commit.

The first local start event used a comma-separated file list, which the helper
treated as one path. A subsequent resume records the exact paths and preserves
that mistake in the immutable history.

Implementation editing is finished. Next: owner review of the local branch and
decision on publication. The claim reserves no additional implementation work;
no shared reservation or external tracker update has been published.
