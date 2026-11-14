---
id: CF-runtime-workspace-budget-v1
status: completed
owner: Codex read-only adapter maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Honor the configured read-only workspace byte budget

## Implementation claim

Owner: Codex read-only adapter maintenance contributor, delegated by the
coordinating session under the owner's maintenance window. Base: main at
`b3faf9c`. Runtime direction and rollout boundaries were reviewed; this is
correctness maintenance of the existing read-only adapter. Other current work
owns modeling output handling and diagrams, with no overlap here.

A finite committed Git fixture reproduces a limit mismatch. With maxBytes set
to 3 MiB, a 2 MiB file reads successfully, but a 2 MiB + 1 byte file fails with
ERR_CHILD_PROCESS_STDIO_MAXBUFFER. The file read honors maxBytes; its pinned-byte
comparison still captures git show with a fixed 2 MiB buffer.

Give the blob comparison a finite capture budget derived from maxBytes. Retain
the existing small default capture cap for metadata commands and do not lower
old blob capture behavior for smaller reader limits. Keep maxBytes admission,
per-file and aggregate limits, pinned-byte matching, symlink checks, UTF-8
handling and existing refusal behavior. No unrelated limit changes or unbounded
capture. Use only invented owned Git fixtures; no live database, provider,
namespace, dispatch, migration or recovery work is authorized by this task.

Scope: `runtime/src/adapters.mjs`, `runtime/test/adapters.test.mjs`, a short
clarification in `runtime/README.md`, this task and new immutable Project Memory
files. No schema, state, store, provider or isolation changes.

Checks: prove the old mismatch, then test committed files at exactly 2 MiB,
2 MiB + 1, the configured maximum and one byte above it. Preserve revision and
dirty-byte refusal controls without exposing file content in diagnostics.
Use Node 24 or later, pinned npm ci with ignored scripts, focused and full
runtime tests, repository validation, whitespace and committed Project Memory
validation. Publish a PR for independent final review; this contributor must
not merge or activate runtime work.

## Integration result

The coordinating agent confirmed independent review of
`d611f4f8fec9b0918218dc7f5ca44396337f3107`; all ten hosted workflows passed on that exact head.
[PR #155](https://github.com/spik3r/clanker-foundry/pull/155) merged on 14 November 2026 as
`e7f342533068227e54865e21ad7d6bb29f56536f`. Git confirms the reviewed head is the merge's second parent.
This bounded implementation claim is complete and reserves no further work.
