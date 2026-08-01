---
name: orchestrator
description: Coordinate complex, decomposable work across scoped subagents with isolated context by decomposing, delegating, cross-checking, and synthesizing. Use for large, separable, verification-heavy, or parallelisable tasks such as multi-file changes, broad research, and research-plan-build-review delivery. Do not use for simple, single-step, or tightly coupled tasks where one focused agent is faster and cheaper.
---

# Orchestrator

Coordinate a small team while keeping the lead session focused on decisions, state, and
handoffs. Do not perform the workers' deep work in the orchestrator context.

## Decide whether to orchestrate

Use multiple workers only when the task has independent slices, needs isolated research,
or benefits from a fresh reviewer. Collapse to one agent when the work is small or when
each slice needs the full detail of every other slice.

## Procedure

1. **Decompose.** Split the goal into bounded tasks with clear inputs, outputs, and
   dependencies. Copy [`templates/plan.md`](templates/plan.md) into the task workspace.
2. **Set up coordination files.** Create only the artifacts the task needs: `plan.md`,
   `findings.md`, `handoff.md`, and `review.md`. Use the templates bundled with this skill.
3. **Delegate scoped work.** Give each worker one tight brief, the minimum relevant
   context, and a short return contract. Use the existing roles:
   - `explorer-researcher` for bounded, read-only discovery;
   - `architect-planner` for the design and ordered plan;
   - `builder-developer` for one implementation slice;
   - `reviewer-verifier` for the independent audit.
4. **Schedule by dependency.** Run independent workers in parallel. Sequence work only
   when one output is an input to another. Avoid parallel builders that edit the same files.
5. **Collect distilled artifacts.** Read conclusions, evidence, and changed files—not
   worker transcripts. Update the coordination files when the task state changes.
6. **Review independently.** Give a fresh `reviewer-verifier` the goal, approved plan,
   complete diff, and validation evidence. Ask it to find where the result fails the plan.
7. **Loop with a cap.** Return blocking findings to a builder, then review the revised
   result. Stop after two or three failed loops and ask the user for direction.
8. **Synthesize.** Reconcile the plan, findings, implementation, and review. Report what
   changed, what was verified, what was not verified, and any open issue.

## Rules

- Stay thin: decompose, delegate, sequence, and aggregate.
- Distill findings; do not copy raw tool output into the lead context.
- Give each worker one job and only the context needed for it.
- Coordinate through durable files when the work spans stages or sessions.
- Keep the reviewer separate from the builder and its reasoning.
- Parallelise independent work and sequence dependent work.
- Preserve unrelated user changes and assign clear file ownership to parallel builders.
- Stop orchestrating when its coordination cost exceeds its value.

## Templates

- [`templates/plan.md`](templates/plan.md) — task graph, dependencies, and done criteria.
- [`templates/handoff.md`](templates/handoff.md) — worker brief and return contract.
- [`templates/findings.md`](templates/findings.md) — distilled research output.
- [`templates/review.md`](templates/review.md) — independent review findings and verdict.
