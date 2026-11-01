---
id: CF-agent-model-duplicate-keys
category: tooling
status: in-progress
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
