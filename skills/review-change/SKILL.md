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
