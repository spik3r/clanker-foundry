# Feature delivery workflow

A repeatable path from a feature request to a reviewed change. Use
[`orchestrator`](../skills/orchestrator/README.md) to coordinate large or separable work;
one agent can follow the same phases for a small change. Stay in each phase until it
produces its output. Building from an unclear plan makes errors expensive.

## Phase 1 — Explore

Output: distilled findings that state what exists, where it lives, and what constrains
the change.

- Use the read-only [`explorer-researcher`](../agents/explorer-researcher.md).
- Inspect the real code, schemas, tests, and documentation rather than relying on memory.
- Record relevant modules, interfaces, data shapes, dependencies, and side effects.
- Mark unknowns as questions. Do not guess.

## Phase 2 — Plan

Output: an approved design and ordered implementation plan.

- Use the read-only [`architect-planner`](../agents/architect-planner.md).
- Turn the findings into small, testable steps with named files and validation.
- State trade-offs, assumptions, risks, compatibility needs, and non-goals.
- Correct the plan before implementation begins.

## Phase 3 — Build

Output: the implementation and evidence from each completed step.

- Use [`builder-developer`](../agents/builder-developer.md), the
  [`engineer`](../skills/engineer/README.md) skill in implementation mode, and the
  [`implement-feature`](../templates/prompts/implement-feature.md) prompt.
- Implement one reviewable slice at a time and validate it before the next slice.
- Follow the target repository's `AGENTS.md` and preserve unrelated changes.
- Run independent slices in parallel only when their file ownership does not overlap.

## Phase 4 — Review

Output: an independent audit, with blocking findings returned to the build phase.

- Use a fresh, read-only [`reviewer-verifier`](../agents/reviewer-verifier.md) and the
  [`review-change`](../skills/review-change/README.md) skill.
- Walk the [`code review checklist`](../checklists/code-review.md).
- Also walk the [`security checklist`](../checklists/security.md) for changes involving
  authentication, authorisation, external input, secrets, or personal data.
- Check the complete diff against the approved plan and validation evidence.
- Return blocking findings to Build. Cap repeated build-review loops, then ask the user.

## Phase 5 — Synthesize and hand off

Output: a short account of the change, validation, open issues, and next action.

- Reconcile the plan, implementation, and independent review.
- State what changed, what was verified, and what was not verified.
- Use the [`pull request template`](../templates/pr/pull_request_template.md) when opening
  a PR.
- Record a lasting design decision with [`adr`](../skills/adr/README.md) when needed.
