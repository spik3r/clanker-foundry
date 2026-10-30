---
id: CF-deterministic-orchestration
category: architecture
status: in-progress
owner: Foundry orchestration planning
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Plan deterministic task execution after task-workflow adoption

## Implementation claim

Owner: Foundry orchestration planning, branch
`codex/deterministic-orchestration-plan`. The user requested updating the planning
worktree from latest origin/main, reviewing the previous recommendation, and
creating a plan. The reviewed base is `7a03b09`; the task-workflow adoption
changes are now included. This is a local planning claim, not a published shared
reservation. The checkout was clean and the connected open-PR search returned
no open pull requests before the claim was prepared.

Exact write scope:
- `docs/deterministic-orchestration-plan.md`
- `tasks/CF-deterministic-orchestration-plan.md`
- New immutable Project Memory source and event files for this task.

Produce an evidence-based architecture and phased implementation plan. Preserve
the existing task-store authority, standalone skills, Markdown source contract,
and offline evidence checks. No application code, schema implementation, existing
skill edits, CI changes, task migration, runtime deployment, publication or merge.
Local commits follow the repository's required source/evidence workflow.

## Acceptance

- Review the changes from the previous base `a25e642` to `7a03b09`.
- Separate current facts, recommendations, assumptions and unresolved decisions.
- Define typed contracts, transition guards, authority, concurrency, recovery,
  worker isolation, validation and task-directory admission.
- Give bounded implementation phases with proposed paths and observable exit gates.
- State how adoption preserves existing workflows and Project Memory compatibility.

## Validation

Run `bash scripts/validate.sh`, `git diff --check`, and
`git diff --cached --check`. Review the design against concurrent assignment,
late results, crashes, duplicate delivery, approval, validation and migration
scenarios. Run Project Memory committed-diff validation after local commits.
No lifecycle helper or CI change is planned, so their regression suite need not
be repeated for this documentation-only change.

## Integration and next action

Integration owner: repository owner. No runtime has been selected or approved for
implementation by this task. Next: write and review the plan, record observed
checks, and deliver the local planning branch for review. Existing claims and
records remain unchanged.
