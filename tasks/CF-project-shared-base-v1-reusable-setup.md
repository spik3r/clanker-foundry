---
id: CF-project-shared-base-v1
status: in-progress
owner: Astra implementation agent
repositories: [spik3r/clanker-foundry]
---

# Reusable project-local shared base

## Implementation claim

Implement a project-local setup and read-only verification helper for a reviewed
Git submodule at `.agent-baseline`, a short project-owned instruction overlay and
selected portable skill links. Cover setup, clean clones, updates, customization,
client discovery limits and the distinction from personal Stow installation.
Review nearby instructions for accuracy and maintainability. Add safety regression
tests and CI coverage; run the pack validator and a clean consumer trial.

Scope: `scripts/project-setup.sh`, `tests/test-project-setup.sh`,
`templates/project/shared-agents.md`, `templates/project/project-agents.md`,
`templates/project/new-project-checklist.md`, `docs/project-shared-base.md`,
`docs/getting-started.md`, `global-agents/README.md`, `global-agents/AGENTS.md`,
`AGENTS.md`, `README.md`, `.github/workflows/validate.yml`, and this task record.
Preserve all other claims and legacy evidence. No machine-wide installation,
runtime or modeling change, merge, deployment, or protection change is included.

This public-safe local claim records the requested implementation. Publication
requires the coordinator's ownership and authorization check. The implementation
uses an isolated checkout to preserve the active source checkout.

## Acceptance

- Reject wrong pins, invalid skill names, destination conflicts and symlink parents
  before creating links or instruction files; repeat setup without overwrites.
- Verification is read-only and detects missing baseline files or selected links.
- Consumer instructions use only the reusable base, never Foundry's root rules.
- Record observed checks and committed Project Memory evidence for review.
