# Project Memory

A standalone skill for durable, append-only task state.

Use it before and after any coherent change so interrupted work can resume from repository evidence rather than chat history. The helper records `start` and `complete` events in `.project-memory/tasks.jsonl`, reports the latest state, and can validate that changed files are covered by completed task evidence.

```bash
node skills/project-memory/scripts/project-memory.mjs status
node skills/project-memory/scripts/project-memory.mjs start --id task-id --summary "Intent" --next "Implement" --files path/to/file
node skills/project-memory/scripts/project-memory.mjs complete --id task-id --summary "Validated result" --next "Merge" --files path/to/file
node skills/project-memory/scripts/project-memory.mjs validate --base origin/main
```

Never record credentials, tokens, private user data, or raw logs in task memory. Repositories adopting the skill should wire `validate` into their required CI policy job.
