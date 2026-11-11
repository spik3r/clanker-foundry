---
id: CF-runtime-cli-input-file-refusal-v1
status: proposed
owner: Codex runtime CLI maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Refuse special CLI input files before a blocking open

## Implementation claim

This claim is for the trusted local operator CLI JSON input path only. On
`origin/main` at `6fbc4a6`, `readWire` in `runtime/src/cli.mjs` opens a path with
`O_RDONLY | O_NOFOLLOW` before `fstatSync` checks that the descriptor is a regular
file. Opening a FIFO with no writer can wait before that check. The shared
`spawnSync` helper in `runtime/test/cli.test.mjs` has no timeout. This finding is
from source inspection; no FIFO or CLI reproduction was run. It does not claim a
remote exploit.

The intended implementation scope is exactly:

- `runtime/src/cli.mjs` — refuse nonregular JSON input paths promptly without
  weakening symlink refusal; preserve descriptor-based validation and cleanup,
  strict UTF-8 and JSON handling, the 1 MiB boundary and redacted diagnostics.
- `runtime/test/cli.test.mjs` — bound CLI subprocess time and cover prompt FIFO
  and other special-file refusal, a valid regular file, symlink behavior, size
  boundary and existing invalid-input redaction.
- This canonical task and new immutable Project Memory source/event files.

Do not change database path handling, worker dispatch, namespace policy,
runtime artifact or isolation cleanup, installer preflight or Go model intake.
Do not run an unbounded FIFO probe. Use an isolated worktree and inspect current
claims and pull requests again before implementation. Review this published claim
with the coordinator before registering its source, recording a genuine start or
editing implementation files. Any scope expansion needs coordinator review.

## Validation gate

After claim review and implementation, use Node 24 or later and the repository's
pinned runtime dependencies. Run bounded focused CLI tests, the runtime suite and
`scripts/validate.sh`, then whitespace and full committed Project Memory checks.
If any test or process restriction remains, record it as unrun rather than
claiming a pass. Publish a PR for independent review only after observed checks.
