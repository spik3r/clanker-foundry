---
id: CF-deterministic-orchestration
category: architecture
status: review
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
implementation by this task. Existing claims and records remain unchanged.

## Local planning result

Fast-forwarded the clean worktree from `a25e642` to origin/main `7a03b09`, then
created the plan against the merged task workflows. The plan keeps standalone
skills and existing Markdown sources, defines opt-in authority cutover, typed
contracts and guarded execution, and gives seven bounded implementation phases.
Only this planning task, the design document and new immutable evidence changed.

Independent read-only design review found two ambiguities: competing nonterminal
executions for one task, and active attempts retaining access after blocking.
The plan now forbids the former and fences/retires blocked attempts while retaining
resources until cleanup is confirmed. Follow-up review found both resolved.
The lead also checked the merged task/coordination workflows and migration rules.
These are document reviews, not tests of an implemented runtime.

Observed checks: `bash scripts/validate.sh` passed, including relative Markdown
links; `git diff --check` passed. Optional `skills-ref` is not installed.
The host runs Node `v26.0.0`; the apparent `node@24` installation resolves to that
same binary, so the skill's Node 24 prerequisite has not been met locally. Helper
commands ran on Node 26; no Node 24 compatibility result is claimed. No helper or
CI code changed; lifecycle and installer suites were not rerun for this
documentation-only change. Staged whitespace and committed-diff checks run during
final recording.

Planning edits are finished; the claim reserves no runtime implementation work.
Next: owner review of phase 0 choices, especially runtime location, host topology,
task authority, capability enforcement and retry budgets. Publication, task-store
migration, merge and deployment have not been performed.
