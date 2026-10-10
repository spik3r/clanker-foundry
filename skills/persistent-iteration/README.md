# persistent-iteration

Keeps a bounded task moving until the user-requested outcome or time boundary,
including waiting for results and recovering from failures within the approved scope.

## When to use

- the user explicitly asks to keep working or iterating;
- work has a stated deadline and timezone;
- an in-progress check or external result needs monitoring;
- an ongoing task resumes from a checkpoint.

It does not authorise unrelated improvements, new recurring jobs, or extra publication
and deployment actions.

## How it behaves

Defines observable completion, verifies current state, works in small validated steps,
and records useful handoffs. A coherent commit is a checkpoint rather than automatic
completion. Pending checks stay pending; a real blocker is reported with the next
decision or access needed.

## Output

Meaningful progress updates and a final state: completed outcome or stopping reason,
validation evidence, unresolved work, and relevant active processes.

## Example invocations

```text
Use persistent-iteration to keep fixing the authorised test failures until the suite passes.
Use persistent-iteration to watch this run until it finishes and report its result.
Use persistent-iteration on this feature until 16:00 UTC; preserve the stated scope.
```

## Installation

See the repository root `README.md`. This skill has no dependency on another skill.
