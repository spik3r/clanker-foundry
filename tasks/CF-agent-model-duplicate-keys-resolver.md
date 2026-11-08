---
id: CF-agent-model-duplicate-keys
category: tooling
status: done
owner: Foundry implementation
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Reject ambiguous model mappings in the agent model resolver

## Implementation claim

Owner: Foundry implementation. The repository owner asked for a fix to ambiguous
model selection in `scripts/agent-model.sh`, published as a draft PR only. Branch
`claude/agent-model-ambiguous-selection-661d1d` from main
`d5e29983bd7e0b49fe05135f6c68fb367a0589c6`.

Problem: with a mapping file holding two `claude.fast` entries the resolver exits 0
and prints the first value. Reversing the lines changes the printed model.

Scope, exact paths:
- `scripts/agent-model.sh`
- `tests/test-agent-model.sh` (new, isolated resolver test)
- `tests/test-install.sh` (one line to run the new test from the CI entrypoint)
- `agents/README.md` (mapping format documentation)
- this task record and its new immutable Project Memory source/event files

Exclusions: `scripts/validate.sh` belongs to another task. No edits to
`agents/model-map.conf`, defaults, workflows, installed skills, credentials or any
local override. Disjoint from draft PR12 (`modeling/lifecycle/`), PR5 (Project
Memory enforcement), active model-tooling work and the held runtime receipt-time
fix. No merge, approval, deploy, auto-merge or agent launch.

## Acceptance and validation

- Mapping files are parsed as data, never sourced or evaluated as shell.
- Duplicate keys in either file fail, even when values match.
- A local entry still overrides the shared entry; a local file that lacks a key
  still falls back to the shared file.
- An empty or malformed local file fails instead of falling back.
- Configuration is validated before any model is written to stdout. Errors go to
  stderr with a nonzero exit and do not print mapping values.
- Isolated tests use invented fixtures in a temporary copy, never the real HOME,
  and never call a model provider.
- Run the new test, installer tests where available, shell syntax, ShellCheck if
  available, pack validation, whitespace checks and committed Project Memory
  validation. Independently review the exact diff before publishing.

## NUL-byte review follow-up

The owner authorized Foundry implementation to take over the outstanding PR 13
fix after review of head `877119947b5fe93e9b2816065229146287edae5e`.
BSD awk can truncate a NUL-containing value or hide a NUL-prefixed duplicate or
malformed record. Reject NUL bytes before awk reads either mapping file, including
bytes in comments and at EOF. Preserve the existing author changes, defaults,
override/fallback behavior and every earlier immutable lifecycle record.

Use invented shared/local fixtures and raw output-file checks. Verify the exact
candidate independently on Linux and native macOS/BSD awk before treating the
portability finding as resolved. Keep the existing PR draft; no approval, merge
or deployment is part of this follow-up.

## Coordinator closeout claim

Owner: Foundry coordinator, for the integration record only. The original
implementation owner and claims above remain historical provenance. Base:
`8b9d9617a24d0666aba7231068d74e0d0c6d6b55`. Intended local branch:
`local/resolver-integration-closeout-20261108`. Recheck the current canonical task,
main, open work and the final PR13 evidence before using this bounded scope.

Record the already merged duplicate/malformed/NUL mapping outcome and close only
that completed implementation scope. Exact writable paths are this task record
and new immutable source/event files for `CF-agent-model-duplicate-keys` only.
Do not edit resolver code, tests, mappings, defaults, other tasks, prior source
snapshots, prior events or legacy ledger bytes. No CRLF profile implementation,
new resolver claim, provider/client execution, installation or settings change
belongs to this closeout. No receiving implementation owner is assigned here.

Independently review this claim before committing its source. Register that exact
reviewed source and make a genuine new start under this claim before changing
status or adding the closeout result. Preserve every previous immutable record.
Verify PR13 integration, current source identity and absence of a visible
competing lane; record facts rather than infer completion from age. Run the pack
validator, whitespace checks and committed lifecycle validation for the final
metadata-only change. Qualify and publish any closeout PR only under its separate
coordinator assignment; this claim does not itself authorize a remote action.
## Integration closeout

PR13 merged on 1 November 2026 as
`f81e07366ba8ddbf174763368307654f6b282dfc`, an ancestor of reviewed main
`8b9d9617a24d0666aba7231068d74e0d0c6d6b55`. All five paths fingerprinted by the final
implementation completion retain their accepted blobs and modes at that main.
At that reviewed main, before this metadata-closeout run, the task's sole
implementation frontier was completion
`719112e5-38f2-4a20-9a8e-b952ad4f1329`. The final PR head was
`7af5d65f8b9f66cad5c5e59ab85f1ea61f59560c`. The PR records independent Linux and
native macOS/BSD-awk review plus successful exact-head hosted checks after the NUL
correction. These are retained implementation results, not new executions by the
metadata closeout.

The coordinator records the integrated duplicate/malformed/NUL mapping outcome
as done and releases only that completed implementation scope. Historical
implementation text and every prior immutable source/event remain intact. The
separate profile-Markdown CRLF finding is not implemented or claimed by this
closeout. Any follow-on needs its own accepted bounded claim and genuine start.

Integration evidence: [PR13](https://github.com/spik3r/clanker-foundry/pull/13).
