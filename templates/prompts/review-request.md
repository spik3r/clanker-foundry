# Prompt template: request a review

Copy this block, fill in the placeholders, and delete any section that does not apply.

```text
Use review-change. Do not edit files.

## Scope
Review <this branch against main | commit <sha> | the working tree | PR <link>>.

## Intent of the change
<one or two sentences so the reviewer can judge whether the diff does what it claims>

## Focus areas
- <e.g. correctness of the retry logic, migration safety, API compatibility>
- Standard checks: `checklists/code-review.md`
- Security-sensitive code touched: <yes — use `checklists/security.md` | no>

## Context the diff does not show
- <decisions made elsewhere, deadlines, known follow-up work>

## Output
Report findings ordered by severity. For each: file and line, what is wrong, why it
matters, and a suggested fix. Separate facts from opinions. Say explicitly if no
issues were found in a focus area.
```

## Notes

- A review without the intent of the change produces style nitpicks instead of substance.
- "Do not edit files" keeps the review independent from the fix.
