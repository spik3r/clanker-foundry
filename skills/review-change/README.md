# review-change

An independent reviewer for pull requests, branches, commits, staged changes, or
the working tree. Read-only by default: it will not edit, reformat, commit, or
push.

## When to use

- pre-merge review of your own work;
- a second opinion on someone else's PR;
- merge-readiness, regression, security, or test-quality checks.

Do not use when you want the reviewer to also fix what it finds — review first
with this skill, then hand fixes to `engineer` so the review stays independent.

## How it behaves

- Identifies the real comparison base from repository evidence rather than
  assuming `main`.
- Reads the full diff before judging individual files; reviews resulting
  behaviour, not syntax.
- Hunts introduced problems in priority order: correctness and data loss first,
  then security, regressions, concurrency, error handling, tests, performance,
  maintainability, documentation.
- Validates each finding before reporting — traces the trigger, checks whether
  existing code already handles it, and separates confirmed defects from
  questions.
- Does not manufacture findings to fill categories, and skips style nits that
  formatters already enforce.

## Output

Findings first, ordered by severity — **Blocker**, **High**, **Medium**,
**Low**, **Question**, **Suggestion** — each with file and line, triggering
scenario, impact, smallest fix, and confidence. Then **Validation performed**
and a **Merge recommendation**: `merge`, `merge after fixes`, or `do not merge`,
with one sentence of reasoning. If nothing material is found, it says so
directly.

## Example invocations

```text
Use review-change to review this branch against main. Do not edit files.
Use review-change on the staged diff, focus on the migration safety.
```

## Related

- `templates/prompts/review-request.md` — prompt starter
- `checklists/code-review.md`, `checklists/security.md` — the standards it checks against

## Installation

See the repository root `README.md`.
