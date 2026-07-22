---
name: safe-refactor
description: Improve code structure without intentionally changing observable behaviour. Use for extracting components, simplifying logic, reducing duplication, renaming, reorganising modules, replacing internals, or paying down technical debt while preserving contracts. Require behavioural characterisation, small steps, continuous validation, and a clean final diff.
---

# Safe Refactor

Change structure, not behaviour.

## Confirm the contract

Before editing:

1. State the refactor goal and non-goals.
2. Identify observable behaviour and public contracts.
3. Find callers, consumers, tests, schemas, configuration, and operational dependencies.
4. Record known edge cases.
5. Confirm the working tree so unrelated changes are preserved.

If behaviour must change, separate that work from the refactor where practical.

## Establish a safety net

Use existing tests to characterise behaviour. Add focused characterisation tests when important behaviour is untested or unclear.

A useful safety net covers:

- normal behaviour;
- boundary and failure cases;
- public interfaces;
- side effects and ordering where relevant;
- serialisation, persistence, or compatibility contracts.

Do not lock in accidental implementation details unless they are externally observable or relied upon.

## Plan reversible steps

Break the work into small transformations where each step:

- has one structural purpose;
- keeps the code runnable;
- can be reviewed and reverted;
- avoids unrelated formatting churn;
- has a clear validation command.

Examples include rename, extract, move, inline, replace dependency, then delete obsolete code. Avoid combining all steps into a rewrite.

## Refactor incrementally

After each meaningful step:

1. run targeted tests or type checks;
2. inspect the diff;
3. confirm behaviour remains stable;
4. remove transitional duplication only after the replacement is proven.

Prefer existing abstractions and conventions. Do not create a generic framework to remove a small amount of duplication.

## Watch for hidden behaviour changes

Check:

- exception types and messages;
- validation order;
- logging and metrics;
- transaction boundaries;
- concurrency and evaluation order;
- defaults and null handling;
- timing and retry behaviour;
- API, schema, serialisation, and file-format compatibility;
- performance-sensitive paths.

## Finish cleanly

- Run the relevant full validation.
- Remove dead code, compatibility shims that are no longer needed, temporary flags, and debug output.
- Review the final diff for scope creep.
- Update documentation only where structure or extension points changed.
- State any behaviour that could not be proven equivalent.

## Output

**Refactor goal**  
**Preserved contracts**  
**Structural changes**  
**Validation**  
**Residual risks**

Do not describe the work as behaviour-preserving when validation is missing.
