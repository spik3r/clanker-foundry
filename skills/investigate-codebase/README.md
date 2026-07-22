# investigate-codebase

Explains an unfamiliar repository or subsystem without changing files. Traces how
behaviour actually flows — entry points, call sites, data, tests, ownership
boundaries — and reports evidence-backed findings.

## When to use

- before implementation, debugging, planning, architecture review, or code review
  in code you do not know;
- to answer "how does X work?" with file-and-line evidence;
- to map the constraints a change must preserve.

Do not use when the answer is already known and the task is to change code — that
is `engineer`.

## How it behaves

- Turns the request into a concrete investigation question with boundaries and
  non-goals.
- Confirms conclusions from code and runtime configuration, not from file names
  or comments.
- Traces the full path: entry point, validation, business logic, persistence,
  external services, error handling, observability, and the tests that exercise
  it — happy and failure paths.
- Records contracts, invariants, and security boundaries a future change must
  preserve, and flags where evidence is incomplete or contradictory.

## Output

**Summary** (direct answer) → **Execution path** (flow with files and symbols) →
**Key constraints** → **Tests and validation** (coverage and gaps) →
**Unknowns and assumptions** (kept separate) → **Recommended next step** (one
proportionate action).

## Example invocations

```text
Use investigate-codebase to trace how refunds flow from the API to Stripe.
Use investigate-codebase: what owns user session state, and what breaks if it changes?
```

## Related

- `templates/prompts/research-question.md` — prompt starter
- `architecture-review` — the follow-on when the question is about design quality, not behaviour

## Installation

See the repository root `README.md`.
