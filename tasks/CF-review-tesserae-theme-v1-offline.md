---
id: CF-review-tesserae-theme-v1
status: in-progress
owner: Byte
repositories: [spik3r/clanker-foundry]
---

# Use Tesserae colour tokens in offline Foundry review reports

## Implementation claim

Apply the existing Tesserae Slate light/dark tokens to the single
`foundryreview` HTML renderer. Embed a revision-pinned, attributed copy;
preserve offline viewing, evidence semantics, CLI inputs, and inert diagrams.
Default to the viewer's colour scheme and retain light print mode. No new
renderer, dependency, service, external asset, or deployment.
