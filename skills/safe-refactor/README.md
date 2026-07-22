# safe-refactor

Improves code structure without intentionally changing observable behaviour.
Small reversible steps, a characterisation safety net, continuous validation,
and a clean final diff.

## When to use

- extracting components, simplifying logic, reducing duplication;
- renaming, reorganising modules, replacing internals;
- paying down technical debt while preserving contracts.

Do not use when behaviour must change too — separate the behaviour change from
the refactor (two commits), or use `engineer` for the combined task.

## How it behaves

- States the goal and non-goals, identifies the observable contracts, and
  protects unrelated changes in the working tree before editing.
- Establishes a safety net first: existing tests, plus characterisation tests
  where important behaviour is untested.
- Breaks the work into small transformations — rename, extract, move, inline,
  replace dependency, delete — each runnable, reviewable, and revertible.
- Runs targeted checks after each step and inspects the diff continuously.
- Watches for hidden behaviour changes: exception types, validation order,
  logging, transaction boundaries, evaluation order, defaults, serialisation,
  timing.
- Finishes clean: full validation, dead code and shims removed, no scope creep.

## Output

**Refactor goal** → **Preserved contracts** → **Structural changes** →
**Validation** → **Residual risks**, including any behaviour that could not be
proven equivalent.

## Example invocations

```text
Use safe-refactor to split this service without changing behaviour.
Use safe-refactor to extract the pricing rules from the order controller.
```

## Related

- `review-change` — independent check of the final diff
- `checklists/code-review.md` — self-review before handoff

## Installation

See the repository root `README.md`.
