# Prompt template: implement a feature

Copy this block, fill in the placeholders, and delete any section that does not apply.

```text
Use engineer in implementation mode.

## Goal
<one or two sentences: what the feature does and who it is for>

## Context
- Repo / service: <name>
- Relevant code: <paths, symbols, or links>
- Ticket / spec: <link>
- Related work: <links or "none">

## Requirements
- <observable behaviour, e.g. "POST /refunds returns 409 when the order is already refunded">
- <observable behaviour>

## Constraints
- <versions, compatibility, performance budgets, security requirements>
- Out of scope: <what must not change>

## Acceptance criteria
- [ ] <testable statement>
- [ ] <testable statement>

## Done means
- Relevant tests pass: <command>
- <lint / typecheck / build command, if any>
```

## Notes

- State behaviour, not implementation, unless the implementation is fixed by an outside decision.
- One prompt, one feature. Split work that needs more than one reviewable change.
- If a requirement is a guess, mark it as an assumption so the agent records it instead of silently inventing the rest.
