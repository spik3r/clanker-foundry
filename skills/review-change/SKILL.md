---
name: review-change
description: Review a pull request, branch, commit, staged changes, or working-tree diff without modifying code. Use for independent code review, merge-readiness checks, regression analysis, test review, security review, and maintainability feedback. Report actionable findings by severity and give an evidence-based merge recommendation.
---

# Review Change

Act as an independent reviewer. Default to read-only.

## Scope and comparison target

1. Read repository instructions.
2. Determine what is being reviewed: PR, branch, commit range, staged diff, or working tree.
3. Identify the correct base branch or comparison point from repository context. Do not assume `main` if evidence indicates otherwise.
4. Read the change description, issue, plan, or commit messages when available.
5. Inspect the complete diff before focusing on individual files.

Do not edit files, reformat code, commit, push, or amend history unless the user explicitly asks after the review.

## Worker and pull-request safeguards

For a PR, branch, or worker change:

- Record the current base and head revisions. Inspect the complete effective diff
  against the intended base, not only the PR summary or an old patch. Check for stale
  stacked commits, already-landed fixes, duplicate work, and unrelated file changes.
- Where the project uses claims, verify the live task record and compare the diff with
  the claimed write set and acceptance checks. A copied record is evidence of past
  context, not proof of current ownership or permission.
- Inspect deleted tests, weakened assertions, skipped checks, and changed test discovery.
  Green CI does not excuse removing the coverage that would expose a regression.
- Verify generated artifacts according to current repository policy. Do not prescribe
  committing ignored build output, manually resolving generated files, or rebuilding
  a branch without authorisation. Report the smallest source-level fix needed.
- Check required status results for the exact current head, including pending, failed,
  skipped, and unavailable checks. Separate local results from remote CI. A prior
  commit's green checks or a successful wrapper does not prove this revision passed.
- Re-review affected behaviour after a rebase, conflict resolution, or later edits.
  If the work is superseded, report that finding rather than merging an empty duplicate.

A merge recommendation is not permission to merge, rewrite a branch, delete it, or
bypass branch protection. Stay read-only and report blockers; make changes only under
separate authorisation.

## Understand intent

State the intended behaviour in one or two sentences.

Then inspect enough surrounding code to understand:

- affected execution paths;
- data contracts and invariants;
- callers and downstream consumers;
- existing tests and conventions;
- deployment or migration context.

Review the resulting behaviour, not only the syntax of the diff.

## Review priorities

Look for concrete, introduced problems in this order:

1. correctness and data loss;
2. security, privacy, and authorisation;
3. regressions and compatibility;
4. concurrency, transactions, retries, and idempotency;
5. error handling and failure recovery;
6. missing or misleading tests;
7. performance or resource risks;
8. maintainability and unnecessary complexity;
9. accessibility and user experience where relevant;
10. documentation, migration, and operational gaps.

Do not manufacture findings to fill categories. Avoid subjective style comments already enforced by formatters or linters.

## Validate findings

For each potential issue:

- trace the affected path;
- identify the input or state that triggers it;
- check whether existing code handles it elsewhere;
- run a focused test or static check when useful;
- distinguish a confirmed defect from a question or suggestion.

A valid finding must explain why the changed code causes or fails to prevent the issue.

## Review tests

Check whether tests:

- cover the changed behaviour;
- assert meaningful outcomes;
- include failure and boundary cases;
- would fail before the change;
- avoid excessive mocking of the behaviour under test;
- remain deterministic and isolated.

Do not demand tests that add little confidence.

## Output contract

List findings first, ordered by severity:

- **Blocker**: unsafe to merge; likely data loss, security breach, outage, or fundamentally broken behaviour.
- **High**: material correctness or regression risk that should be fixed before merge.
- **Medium**: real defect or important coverage gap with narrower impact.
- **Low**: small but actionable problem.
- **Question**: missing context that could change the verdict.
- **Suggestion**: optional improvement; clearly mark it as non-blocking.

Each finding must include:

- a concise title;
- file and line or symbol;
- the triggering scenario;
- the impact;
- the smallest reasonable fix or direction;
- confidence: high, medium, or low.

Then provide:

**Validation performed**  
Commands or analysis used.

**Merge recommendation**  
`merge`, `merge after fixes`, or `do not merge`, with one sentence of reasoning.

If no material issues are found, say so directly and note any limits in the review. Do not produce praise or filler.
