---
id: CF-orchestration-missing-inputs
status: in-progress
owner: Foundry input fixes
repositories: [spik3r/clanker-foundry]
---

# Reject missing immutable task inputs

## Implementation claim

The user authorized fixing the fresh-review missing dependency finding. This local
follow-up starts at `55022ea` on `codex/deterministic-orchestration-plan`. The branch,
staged diff, working diff and untracked paths were clean. Existing review claims
remain unchanged. The open-PR check identifies a separate Project Memory CI lane;
this fix does not change that helper or CI. Exact scope:

- `runtime/src/runner.mjs`
- `runtime/test/store.test.mjs`
- `runtime/README.md`
- `tasks/CF-orchestration-missing-inputs-implementation.md`
- New immutable source and event files for this task.

Require every dependency, accepted candidate and explicit subject manifest before
claiming or acknowledging dispatch. Treat only the exact full pinned repository
revision as a Git base reference, including SHA-256 Git IDs. Reject unsupported
Git revisions and missing subjects instead of silently dropping them. Keep dispatch
fairness, bounded inputs, fencing, replay and cleanup behavior intact. Do not
activate namespaces, rewrite other claims, publish, merge or deploy.

## Validation and next action

Add actual SQLite regressions for missing dependency manifests after reopening,
no invocation or acknowledgement, later-work fairness and successful recovery
after restoring the manifest. Cover explicit missing inputs and pinned SHA-1 and
SHA-256 Git bases. Run Node 24 runtime tests, pack/whitespace checks and the
committed evidence gate. Next: register this pinned source and record a genuine
start before implementation.
