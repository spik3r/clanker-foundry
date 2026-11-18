---
id: CF-registry-intake-docs-v1
status: in-progress
owner: Foundry implementation lane, 2026-11-19
repositories: [spik3r/clanker-foundry]
---

# Explain the registry's existing intake rules

## Implementation claim

Prepare a bounded documentation clarification from main
`1d019d5f31c9af0cfe5c66fd966ffb935037f7eb`. The continuing-work request
permits useful independent Foundry maintenance. Live PR lookup found no open
Foundry PRs. Existing unfinished tasks retain their owners; this change does
not reopen or release them. A source inspection found that the registry's
intentional permissive field handling is covered by TestUnicodeRegistryIntake,
but the guide only refers to unspecified existing rules.

Owner: Foundry implementation lane, 2026-11-19. Scope: only
`modeling/foundrygen/README.md`, this task and new immutable Project Memory
source/events. Explain the registry path, its 1 MiB limit, ignored unknown
fields, last-value handling for repeated string fields, and refusal of repeated
generator IDs across entries. Keep strict project-file rules distinct. Do not
change the decoder, tests, generated output, dependencies, prior evidence,
other claims, runtime or held work. No merge or deployment.

This local claim is a proposal until the coordinator accepts it through the
shared source. Prepare the exact candidate for independent coordinator review;
source publication waits for that review. Do not claim integration or public
reservation from a local source snapshot.

## Validation and handoff

Check the wording directly against LoadRegistry and the existing
TestUnicodeRegistryIntake regression, run that focused regression with bounded
Mac execution, repository validation, whitespace and committed Project Memory
validation. Report actual results and unavailable optional tools. Return the
exact candidate commit/tree and patch for independent coordinator review before
publication. Preserve all old lifecycle bytes and the primary checkout.
