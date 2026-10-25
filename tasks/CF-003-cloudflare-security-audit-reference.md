---
id: CF-003
category: docs
status: in-progress
owner: Foundry implementation
repositories: [spik3r/clanker-foundry]
depends_on: [CF-001]
---

# Document the Cloudflare security-audit skill as an external reference

## Implementation claim

Owner: Foundry implementation. The repository owner asked for the public
`cloudflare/security-audit-skill` to be listed as an external reference skill with
installation, usage and tips.

Scope: docs/external-skills.md, one new guide under docs/, the README external
skills pointer, this task record, and its immutable v2 source/event metadata.
Reference only: do not vendor or copy upstream files, do not add a skill under
skills/, and do not change validators, workflows or installers. No merge, release
or consumer rollout.

## Acceptance and validation

- Install commands, modes and requirements match the upstream README and SKILL.md
  as read on the date of writing; state that upstream may change.
- Tips distinguish upstream facts from this repo's suggestions.
- Run scripts/validate.sh and the existing memory regressions.
