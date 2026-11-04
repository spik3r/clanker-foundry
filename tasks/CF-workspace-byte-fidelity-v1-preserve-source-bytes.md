---
id: CF-workspace-byte-fidelity-v1
status: review
owner: Foundry workspace byte fidelity
repositories: [spik3r/clanker-foundry]
---

# Preserve UTF-8 source bytes in workspace reads

## Implementation claim

The user authorized the next implementation and review batch. The integration
lead allocated this new bounded defect separately from historical completed
isolation work and held general recovery work. This allocation does not transfer
those claims. Base `f16c94d456e89b7ee0680b91d13a44fb79be5556`, branch
`codex/workspace-byte-fidelity-v1`. The isolated checkout had no staged, unstaged
or untracked changes. The live open-PR check found only PR 79, whose documentation
and example paths do not overlap this lane.

Exact scope:

- `runtime/src/adapters.mjs`
- `runtime/src/workspace.mjs`
- `runtime/test/workspace-byte-fidelity.test.mjs` (new)
- `runtime/README.md`
- `tasks/CF-workspace-byte-fidelity-v1-preserve-source-bytes.md`
- New immutable Project Memory source and event files for this task.

Both public workspace readers decode valid UTF-8 with a decoder that removes an
initial BOM, while reporting a digest of the original bytes. A temporary committed
BOM source reproduced inconsistent returned content and digest in both readers.
The text adapter consequently refused its own reader's output with
`INVALID_INPUT_DIGEST`, without invoking its callback.

Preserve the initial BOM in decoded source content so re-encoding the returned
string reproduces the hashed bytes. Keep fatal UTF-8 decoding, existing malformed
input failure types, byte bounds, pinned revision checks and path checks. Original
bytes, including the BOM, count toward limits. Do not normalize source text.

The lead reviews this published claim before source registration and genuine
start. Freeze independently authored literal tests after that start and before
production edits. Preserve existing tests and historical evidence. No schema,
store, state, runner, provider, CLI, host-isolation or recovery changes are included.
Use invented callbacks and a fake candidate host only; no real provider, container,
namespace activation, credentials or deployment is authorized. This lane may
publish its draft PR and implementation; the integration lead alone merges.

## Acceptance and validation plan

Check literal source bytes and digests for BOM-only and BOM-prefixed text, plain
text, multibyte text, CRLF and literal U+FFFD. Prove the read-only reader feeds the
text callback without digest refusal and that candidate reads preserve the same
content through an invented isolated callback. Cover bounds that include BOM
bytes and preserve malformed-UTF-8 refusal. Existing path, dirty-workspace and
cancellation tests remain unchanged.

Observe the frozen tests fail on the old code before implementation. Run the
focused tests and full runtime suite on the available Node 24-or-later host,
repository validation and whitespace checks. Independently remove either decoder
fix in disposable copies to show that its respective tests detect the regression.
Validate committed Project Memory evidence against the actual PR base. Record
only checks that ran and distinguish local Node results from hosted Node 24 CI.

## Observed result

Both readers now retain the BOM while decoding valid UTF-8. Production changes
are limited to the two decoder options and the runtime README explanation.
The frozen test file remains unchanged at SHA-256
`ac01688046674409fdd6360e17d5d1f6f41803f3f842f9b4302d643d5c3efd82`.
The independent author committed it after the genuine start and before production
edits. On the old code, 21 tests reported 13 passes and eight failures. The author
corrected an oversized-blob harness case before freezing: two individually fitting
blobs now test the candidate's combined bound without hitting Git's per-command
buffer limit. No frozen expectations changed after implementation began.

Actual Node v26.0.0 passed all 21 frozen tests and all 280 runtime tests, with no
failures, skips or cancellations. In a disposable clone, removing only the
read-only decoder fix caused five assertion failures; removing only the candidate
decoder fix caused three. Both mutants passed syntax checks. Restoring both fixes
passed all 21 tests. Repository validation, installer regressions and whitespace
checks passed; optional skills-ref is unavailable. These tests used invented
callbacks and a fake candidate host, with no real provider or container invocation.

An ordinary merge incorporated main `01b18785e738d9cc656ea1fb622f8c9aad1b220a`
without conflicts or changes to this lane's runtime bytes. Local Node results do
not establish hosted Node 24 results. Independent review and exact-head hosted
checks remain required before the integration lead merges. This implementation
claim reserves no further changes beyond review corrections for this bounded task.
