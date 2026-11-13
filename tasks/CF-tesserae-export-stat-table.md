---
id: CF-tesserae-export
status: in-progress
owner: Foundry integration worker
repositories: [spik3r/clanker-foundry]
---

# Bounded Tesserae export

## Implementation claim

Implement a local-only stat/table dashboard-to-tesserae/v1 export seam and CLI.
Foundry owns model validation and mapping, not themes or markup. Refuse unsupported
charts, nonnumeric stat values and target-limit overflow. Preserve reading order,
not source grid rows. Test strict input, determinism and numeric/table conversion.
No renderer vendoring, package publication, remote changes, browser previews or
claims of complete dashboard or framework support.
