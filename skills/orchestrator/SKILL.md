---
name: orchestrator
description: Coordinate complex, decomposable work across scoped subagents with isolated context by decomposing, delegating, cross-checking, and synthesizing. Use for large, separable, verification-heavy, or parallelisable tasks and for reconciling their active plans and ownership. Do not use for simple, single-step, or tightly coupled tasks where one focused agent is faster and cheaper.
---

# Orchestrator

Coordinate a small team while keeping the lead session focused on decisions, state, and
handoffs. Do not perform the workers' deep work in the orchestrator context.

## Decide whether to orchestrate

Use multiple workers only when the task has independent slices, needs isolated research,
or benefits from a fresh reviewer. Collapse to one agent when the work is small or when
each slice needs the full detail of every other slice.

## Procedure

1. **Establish the current lane.** Read project instructions, inspect the working tree,
   and check the project's existing task records, active claims, and open pull requests.
   Reuse the matching task rather than creating a duplicate. Verify current ownership
   at its authoritative source; a checkpoint, imported snapshot, or local event log does
   not claim or transfer ownership. If ownership is unclear, resolve it before editing.
2. **Decompose.** Split the goal into bounded tasks with clear inputs, outputs,
   dependencies, and acceptance checks. Reuse an existing active plan; otherwise use
   [`templates/plan.md`](templates/plan.md) in the task workspace.
3. **Use one coordination source.** Create only the artifacts the task needs:
   `plan.md`, `findings.md`, `handoff.md`, and `review.md`. Link authoritative task
   records rather than creating another ownership or lifecycle ledger. Keep local
   execution notes separate from canonical task state.
4. **Delegate scoped work.** Give each worker one tight brief: outcome, inputs,
   exclusions, exact writable paths, base and branch when relevant, allowed actions,
   acceptance checks, dependencies, and the person responsible for integration.
   Use the existing roles:
   - `explorer-researcher` for bounded, read-only discovery;
   - `architect-planner` for the design and ordered plan;
   - `builder-developer` for one implementation slice;
   - `reviewer-verifier` for the independent audit.
5. **Schedule by dependency.** Parallelise independent work. Sequence work when an
   output is another worker's input. Do not assign overlapping write sets without an
   explicit handoff. In a dirty shared checkout, prefer isolated worktrees; do not use
   Git's shared stash. Review the existing branch or PR instead of reimplementing it.
6. **Collect distilled artifacts.** Read conclusions, evidence, and changed paths
   rather than transcripts. Require exact validation results and limitations, the
   revision checked, unresolved dependencies, and the next action.
7. **Review independently.** Give a fresh reviewer the goal, plan, current ownership
   record, complete diff, and validation evidence. Ask where the result fails its
   acceptance checks. Recheck the current base and other active work before integration.
8. **Resolve findings.** Return blocking defects to the responsible builder and review
   the revised result. If attempts repeat the same failure without new evidence, change
   the diagnostic approach or report the specific blocker. Do not invent unrelated
   work, silently broaden scope, or treat an arbitrary loop count as completion.
9. **Close the loop.** Reconcile the plan, implementation, checks, and review. Where
   authorised, update the existing task with evidence and a clear owner handoff or
   next dependency. Release a claim only when its owner has finished or handed it off.
   Report changed, verified, unverified, and blocked work. Publishing or merging still
   requires the applicable authorisation.

## Keep the plan current

- Distinguish active work, blocked work, completed work, superseded plans, and reference
  material. Record why a plan changed and which current record replaces it.
- Mark work complete from implementation and validation evidence, not from an old
  checkbox, merged PR title, or worker assertion alone.
- Before renaming, moving, or removing a plan, check references in code, tests, docs,
  and task records. Preserve stable identifiers, history, and required compatibility
  links. Do not delete or reorganise an external backlog without authorisation.
- Add a task only for a concrete problem with an observable outcome, dependencies,
  acceptance checks, and a next action. Avoid speculative queue growth.
- Reconcile stale next steps after each meaningful result. Keep evidence where it
  already belongs and link it rather than copying it into every coordination file.

## Rules

- Stay thin: decompose, delegate, sequence, and aggregate.
- Give each worker one job and only the context needed for it.
- Keep the reviewer separate from the builder and its reasoning.
- Preserve unrelated changes, including staged files and edits within shared paths.
- Do not treat a claim, plan, checkpoint, or worker message as additional permission.
- Stop orchestrating when its coordination cost exceeds its value.

## Templates

- [`templates/plan.md`](templates/plan.md) — task graph, dependencies, and done criteria.
- [`templates/handoff.md`](templates/handoff.md) — worker brief and return contract.
- [`templates/findings.md`](templates/findings.md) — distilled research output.
- [`templates/review.md`](templates/review.md) — independent review findings and verdict.
