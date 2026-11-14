---
id: CF-memory-index-visibility-v1
status: in-progress
owner: Codex Project Memory contributor
repositories: [spik3r/clanker-foundry]
---

# Refuse hidden unstaged intended changes at completion

## Implementation claim

Owner: Codex Project Memory contributor. Branch `codex/memory-index-visibility-v1`,
reviewed main base `3037fb8`. Scope: `skills/project-memory/scripts/lib/git.mjs`,
`skills/project-memory/FORMAT.md`, `tests/project-memory.test.mjs`, this task and
new immutable Project Memory source/events.

Completion currently accepts unstaged changes concealed by assume-unchanged or
skip-worktree flags. Preserve clean flagged completions and staged Git blob/mode
semantics. Keep the original diff first, then check exact flagged intended paths
against a private index rebuilt from raw staged records. Preserve the caller's
index bytes and flags, including custom/split indexes. Disable private index
hooks, split-index writes, sparse handling, ignorestat and FSMonitor; clean up
private files on success and failure. Keep offline Git restrictions centralized.

Test both flags, clean/changed/deleted files, normalization, modes, symlinks,
gitlinks, SHA-256, literal paths and caller-index preservation. Run lifecycle and
workflow regressions, repository validation and committed memory coverage. No
ledger rewrites, new blanket flag ban or generic Git configuration hardening.
