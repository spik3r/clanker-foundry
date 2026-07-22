# Bug investigation workflow

A repeatable procedure for going from "something is wrong" to a verified fix with a
regression test. Use it with the `engineer` skill (repair mode) and the
`templates/prompts/fix-bug.md` prompt. Stay in earlier phases until each has produced
its output — most bad fixes come from jumping to phase 4 with a guess.

## Phase 1 — Pin down the symptom

Output: a written statement of expected vs actual behaviour.

- Write down what should happen and what actually happens, in observable terms.
- Capture the exact error text, stack trace, or wrong output. Not a paraphrase.
- Note when it started: first-seen date, release, or commit. If it follows a deploy, the diff since the last good deploy is the prime suspect list.
- Check scope: one user or all users, one environment or all, one input shape or any.

## Phase 2 — Reproduce

Output: steps that trigger the bug on demand, or a written record that it cannot yet be reproduced.

- Reproduce in the lowest environment possible: unit-level beats local, local beats staging, staging beats production.
- Reduce the input. Remove every step and every field that is not required to trigger the bug.
- If it reproduces intermittently, record the rate and conditions. Do not "fix" an intermittent bug you cannot trigger.
- If it cannot be reproduced: add logging or tracing at the suspected boundary and stop here. Report what instrumentation was added and why.

## Phase 3 — Isolate the cause

Output: one confirmed root cause with evidence, or a ranked list of hypotheses with the evidence for each.

- Form hypotheses before reading code in depth. Write them down; test them in order of likelihood times cost of testing.
- Use evidence, not vibes: logs, traces, a failing assertion, a bisect (`git bisect` for regressions, a binary search over inputs for data-dependent bugs).
- Change one thing at a time. A test that toggles two variables proves nothing.
- Follow the data across boundaries: where does the wrong value first appear? The cause is upstream of that point, the symptom is downstream.
- Distinguish root cause from trigger. "The deploy triggered it" is not a root cause.
- Timebox. If no hypothesis survives, stop and report the evidence gathered rather than guessing a fix.

## Phase 4 — Fix

Output: the smallest change that removes the root cause.

- Fix the root cause, not the symptom. A guard clause that hides the bad value moves the bug; it does not fix it.
- Keep the diff small and separate from any cleanup noticed along the way. Two changes, two commits.
- Respect the boundaries the codebase already has — do not let the fix leak internals across a module line.

## Phase 5 — Verify

Output: evidence the bug is gone and nothing else moved.

- Run the reproduction steps from phase 2. They now produce the expected behaviour.
- Add a regression test that fails on the old code and passes on the new. This is not optional; an unfixed test means the bug can return silently.
- Run the project's relevant test suite. State what was run and the result.
- Check the blast radius: what else calls the changed code? Verify one caller by hand if the suite does not cover it.

## Phase 6 — Close out

Output: a record that lets the next person skip phases 1–3.

- Commit message states the cause, not just the symptom: "Fix race in session cache that dropped refresh tokens", not "Fix login bug".
- If the investigation revealed adjacent problems, file them — do not expand this fix.
- If the bug class is preventable (missing validation, unchecked error, race), say what would catch it earlier: a lint rule, a type, a test pattern, an alert.
- If the investigation changed how the system is understood, update the docs or ADR that held the wrong picture.

## Escalation

Stop and hand back to a human when:

- the fix requires changing behaviour others depend on;
- the cause is in a third-party system and the workaround is non-trivial;
- reproduction needs production access or data you do not have;
- the timebox expired twice with no surviving hypothesis.

State exactly what is known, what was ruled out, and what is needed to continue.
