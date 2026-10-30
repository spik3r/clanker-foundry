---
id: CF-orchestration-missing-inputs
status: review
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

Implemented mandatory dependency, candidate and explicit subject resolution.
Actual SQLite regressions cover missing manifests across reopen, no launch or
acknowledgement, later-work fairness, restored-manifest recovery and replay.
Mandatory subjects retain precedence when their digest equals the pinned Git base.
Explicit unknown subjects, other Git revisions and abbreviated pinned values reject;
full SHA-1 and SHA-256 pinned bases remain valid. A missing completed dependency
record also rejects. Corruption tests alter only owned temporary databases.

Node v24.22.2 passed all 163 runtime tests, without skips. Pack and whitespace
validation passed; optional skills-ref is unavailable. Independent review passed
five focused tests and found no remaining material findings. Next: record staged
completion fingerprints, commit locally and validate
the committed full diff against origin/main. No live provider, Docker, namespace
activation, publication, merge or deployment is included.
