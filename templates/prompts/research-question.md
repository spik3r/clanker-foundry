# Prompt template: research a question

Copy this block, fill in the placeholders, and delete any section that does not apply.

```text
Use investigate-codebase. Read-only.

## Question
<the single question to answer, e.g. "How do refunds flow from the API to Stripe?">

## Why it matters
<the decision this answer feeds into — changes what "done" looks like>

## Starting points
- <paths, symbols, docs, or "unknown — find them">

## Boundaries
- Timebox: <e.g. one pass, no more than N files deep>
- Out of scope: <adjacent questions to leave for later>

## Output
- Direct answer first, then the evidence.
- Trace the path with file and line references.
- List what remains unknown and how it could be confirmed.
- State assumptions separately from findings.
```

## Notes

- One question per prompt. Stacked questions get shallow answers.
- Asking for the unknowns explicitly stops the answer from papering over gaps.
