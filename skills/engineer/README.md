# engineer

The general-purpose workhorse: plans, implements, and validates software changes
end to end. Understands the repository first, chooses the smallest maintainable
solution, edits code, runs checks, and reports evidence.

## When to use

- feature work and bug fixes where code will change;
- autonomous tasks that need investigation, implementation, and validation in one pass;
- validating an existing change without redesigning it;
- repairing a failed implementation or check.

Do not use for review-only requests — use `review-change` so nothing gets edited.

## Modes

The skill infers the narrowest mode that satisfies the request; naming one in the
prompt pins it:

| Mode | What it does | Edits? |
|---|---|---|
| investigate | Understand code, return findings | No |
| plan | Investigate, propose an implementation plan | No, unless asked |
| implement | Investigate, plan, edit, validate, report | Yes |
| validate | Inspect existing changes, run checks | No |
| repair | Diagnose a failure, apply the smallest fix, validate | Yes |

## How it behaves

- Establishes scope before editing: reads project instructions, inspects the
  working tree, protects unrelated user changes.
- Fixes causes, not symptoms; reproduces bugs before editing where practical.
- Keeps diffs small; no unrelated refactors, speculative abstractions, or new
  dependencies.
- Validates with repository evidence — targeted tests first, broader checks when
  warranted — and never claims a check passed that did not run.

## Output

Three short sections: **Changed**, **Validation**, **Remaining** (risks,
assumptions, unverified items). Detail only where it affects review or the next
action.

## Example invocations

```text
Use engineer in implementation mode to add rate limiting to the login endpoint.
Use engineer in validation mode on my current changes.
Use engineer in repair mode: the integration tests fail on CI but pass locally.
```

## Related

- `templates/prompts/implement-feature.md`, `templates/prompts/fix-bug.md` — prompt starters
- `workflows/bug-investigation.md` — pairs with repair mode
- `checklists/code-review.md` — self-review before handing off

## Installation

See the repository root `README.md`.
