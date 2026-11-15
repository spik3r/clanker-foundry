---
id: CF-candidate-content-budget-v1
status: in-progress
owner: Codex coordinator with delegated builder
repositories: [spik3r/clanker-foundry]
---

# Honor small candidate workspace content budgets

## Implementation claim

The owner requested continued delegated Foundry maintenance, independent review
and tested normal merges. The coordinator reviewed this separate public-safe task
against main `821508e20930066f9308c7609d320f311ffb2a48`. Live PRs are empty;
Project Memory reports 178 tasks, no conflicts and no active run. Use an isolated
worktree and preserve all prior claims and immutable history.

The candidate workspace API accepts maxBytes from 1 through 128 MiB and charges
selected blob bytes to an aggregate snapshot total. It also passes that same
budget to ls-tree. An owned one-byte blob with a 16-byte content budget therefore
fails on unrelated tree metadata; the identical fixture admits at 1024 bytes.
This is an accepted-argument mismatch, not a deployed consumer incident. The
exported API has no ordinary in-repository caller beyond tests and live isolation.

Use the existing candidate default 32 MiB as the metadata capture floor:
Math.max(DEFAULT_TOTAL_BYTES, maxBytes). Default and larger custom metadata
capture limits stay unchanged. Configured per-blob capture and aggregate selected
content charging stay unchanged. Metadata remains bounded; this does not make
arbitrary repositories fit or bypass malformed tree-name refusal.

Exact scope:

- `runtime/src/workspace.mjs`: candidate ls-tree capture limit only.
- `runtime/test/workspace-budget.test.mjs`: new public-API budget regressions.
- `runtime/README.md`: distinguish candidate content and metadata budgets.
- This canonical task and new immutable Project Memory source/events.

Do not import the read-only adapter's 2 MiB policy or reopen its historical budget
claim. No maxFiles policy, new parameter, error-code definition, blob decoder,
path/pin/Unicode rule, schema, dependency, version, provider, store/state/runner,
host isolation, consumer or historical metadata changes are included. This task
does not activate managed orchestration or change namespace/authority/dispatch.

## Acceptance and validation

Commit this reviewed source, register its exact revision, independently compare
canonical bytes and identity, and record a genuine Node 24 v2 start before any
tracked tests or code. Freeze public-API literals and observe old-code failures.

One-byte content at budget 1, four-byte content at 4 and fitting content at 16
must admit despite larger tree metadata. Include empty content, a four-byte astral
character and BOM plus ASCII, checking exact content and original-byte digests.
Unrelated entries must not consume the selected content budget or expose blobs.

Exact aggregate selected content must admit; aggregate overflow must refuse
before verification through the existing WORKSPACE_TOO_LARGE path. A single
oversized blob must still refuse through the existing bounded Git capture path.
Overlapping read/write-base selection must charge a blob once. Preserve defaults,
larger configured bounds, symlink/pin/path/cleanup rules and malformed-name
refusal. Source review must verify the metadata expression remains finite without
adding a huge fixture solely to restate that expression.

Use owned Git objects, fake host counters and command recording as needed. No
provider, container, authority database, candidate execution or dispatch. Install
pinned dependencies with scripts disabled and an owned npm cache. Run focused and
full runtime tests on actual Node 24, pack validation, installer and whitespace
checks. Preserve package manifests/locks and source repository bytes.

The coordinator independently checks the candidate, stages named paths, records
completion from actual staged blobs, commits new evidence and validates the full
committed diff with explicit code/source identities. Two reviewers bind final
files and commit/tree. All hosted gates, including orchestration live isolation,
must pass before normal merge. No live-host claim follows from local fake tests.
