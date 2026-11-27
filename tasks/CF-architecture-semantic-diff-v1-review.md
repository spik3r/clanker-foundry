---
id: CF-architecture-semantic-diff-v1
status: in-progress
owner: Byte
repositories: [spik3r/clanker-foundry]
---

# Review changes to a logical architecture model

## Implementation claim

Byte owns a read-only semantic comparison for two validated versions of the
same architecture-diagram/v1 model. Report changes by stable node and link ID
and distinguish relative order of existing nodes from semantic edits. Scope is
the architecture comparator, diagramgen CLI, tests, and its documentation.
Do not modify generator output, add dependencies, or expand other profiles.
