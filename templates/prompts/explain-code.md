# Prompt template: explain unfamiliar code

Copy this block, fill in the placeholders, and delete any section that does not apply.

```text
Use investigate-codebase. Read-only.

## Context
- Repo / area: <paths, modules, or symbols>
- Entry points / how to run it: <commands and main files>
- What I know: <one or two lines>
- Why I am asking: <about to change it / review it / debug a problem>

## Task
Explain how this code works, grounded in the actual repository rather than how similar
code usually works.

- Walk the main flow end to end and cite `file:line` at each key step.
- Identify the entry points, types, functions, and boundaries that matter.
- Call out external dependencies, state changes, side effects, and non-obvious behaviour.

## Constraints
- If a claim is not supported by the inspected files, label it as an assumption or unknown.
- Do not change files or turn the explanation into an implementation plan.

## Output
1. Summary — two or three sentences describing what the code does and its role.
2. Flow — a numbered walkthrough with `file:line` citations. Add a Mermaid sequence or
   flow diagram only when it makes the path easier to understand.
3. Key parts — the types, functions, and boundaries that matter and why.
4. Risks, gotchas, and unknowns — edge cases, side effects, and unproved points.
5. How to confirm — a test, trace, query, or command that would verify the explanation.
```

## Notes

- Give the agent a narrow area or question. Broad requests produce shallow explanations.
- Include the reason for the explanation so the agent can focus on relevant behaviour.
