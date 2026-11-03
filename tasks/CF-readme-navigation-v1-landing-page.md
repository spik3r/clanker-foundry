---
id: CF-readme-navigation-v1
status: in-progress
owner: README navigation contributor
repositories: [spik3r/clanker-foundry]
---

# Reorganize the Foundry landing page

## Implementation claim

The owner requested a concise, inviting README introducing installation and the
main areas of Foundry: skills, AGENTS.md, Project Memory, the model-driven
pipeline and orchestration. This local documentation lane starts from reviewed
main `4fe0611d032ea7060e9804b23634d1ec13c481a4`; it is not a published shared
reservation. The merged pilot guide's README link is preserved. Existing claims
and open work were inspected; the historical pilot guide claim described only a
local proposal and its change was merged in #48.

Exact scope: `README.md`, new `docs/getting-started.md`, new `skills/README.md`,
this task record, and new immutable Project Memory source/event files for this
task. Preserve useful install instructions and the skill catalog in focused
guides; check links, commands, optional dependencies and capability limits
against current implementation. Do not alter runtime, models, generated files,
other task records or prior evidence.

## Validation and handoff

Run the repository validator, installer regressions, whitespace and full committed
diff Project Memory validation; independently review the documentation where
available. Report actual outcomes and limits. Normal branch publication and a
draft PR are authorized; merge and deployment are not part of this scope.
