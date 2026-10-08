# Project Memory

A standalone skill for durable, append-only task state.

Use it before and after every coherent change so interrupted work can resume
from repository evidence rather than chat history. The helper records
append-only `start` and `complete` events in `.project-memory/tasks.jsonl`,
reports the latest task state, and validates that changed files are covered by a
started and completed task.

```bash
node skills/project-memory/scripts/project-memory.mjs status
node skills/project-memory/scripts/project-memory.mjs start --id task-id --summary "Intent" --next "Implement" --files path/to/file
node skills/project-memory/scripts/project-memory.mjs complete --id task-id --summary "Validated result" --next "Merge" --files path/to/file
node skills/project-memory/scripts/project-memory.mjs validate --base origin/main
```

The task's start and completion events belong in the same commit or pull request
as its changes. CI rejects a pull request when any changed non-memory path is
not listed by a task with both events. This applies to docs, configuration, and
workflow changes as well as code.

Never record credentials, tokens, private user data, or raw logs in task memory.
