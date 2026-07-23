# context-offload

Creates a compact, replaceable `context.md` handoff for one active task. It preserves
the verified repository state, evidence, validation, risks, and first next action so a
new agent can resume without reconstructing the conversation.

## When to use

Use before a model or agent switch, after a long session, before pausing, or when the
user asks to compact or resume context. It writes documentation only.

Use `checkpoint` instead when the repository needs durable decision and assumption
records as well as the current task state.

## Output

Reports the created or read `context.md`, verified repository evidence, validation
state, risks, and the first next action.

## Installation

See the repository root `README.md`.
