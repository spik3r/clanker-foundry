---
id: CF-research-page-adapter-path
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Move the research-page adapter import beside the output directory

## Implementation claim

Owner: Claude pipeline coordinator. This corrects CF-research-page-v1,
found during the heisentick trial: the generated component imported
`./researchPageParts`, so the host adapter had to sit inside the generated
directory, where `diagramgen check` reports it as drift.

Scope: the import becomes `../researchPageParts` in
`modeling/diagram/researchpage/`, with its tests, the regenerated
`evidence-atlas-page` example, `RESEARCHPAGE.md`, the `emits` text in
`modeling/profiles.json`, this task record and new immutable Project Memory
files.

Exclusions: other profiles and the heisentick application.
