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

## Accepted local implementation scope

Independent coordinator review accepted this bounded documentation task and the
original proposal commit `cf764ce0b38fa1e6c105f4315d4393dc4471d1d3`.
The accepted precision requirement is: for repeated known string fields, a later
string value replaces an earlier value; null does not overwrite strings and
invalid types may refuse. Unknown fields undergo Unicode preflight before they
are ignored. This accepts an explanation of current semantics, not a decoder
policy change. Preserve the proposal text above as provenance.

A fresh ordinary fetch confirmed main remains
`1d019d5f31c9af0cfe5c66fd966ffb935037f7eb`, and a live PR query returned none.
The coordinator authorized this local claim, pinning and genuine start before
applying the README clarification. Public source reservation and integration
remain pending final independent review and ordinary publication.

The focused Go regression did not execute: sandbox cache access was denied
before test setup completed. Do not retry Go/cache access, process listing or
an alternate route. Documentation qualification uses direct source/regression
inspection, repository documentation validation, whitespace and lifecycle
validation. Unrun tests remain disclosed. No new dependency or runtime suite.

## Local candidate result

The README now states the registry location and 1 MiB bound, Unicode preflight
for unknown fields, later-string replacement for repeated known string fields,
null's lack of effect on strings, possible invalid-type refusal, and unique
generator IDs across entries. No decoder or test behavior changed.

Source inspection checked LoadRegistry, its preflight/read bound and duplicate
ID checks, TestUnicodeRegistryIntake's intentional repeated-field behavior and
unknown-field Unicode refusals, and the installed Go encoding/json source
comments on duplicate members, type errors and null semantics.

The repository documentation validator passed; optional skills-ref was not
installed and was skipped. Whitespace checks passed. The previously attempted
focused Go regression did not execute because the sandbox denied build-cache
access during setup. It was not retried, and no alternate cache, process lookup,
dependency install or runtime test was used. Other Go tests and hosted CI were
not run for this documentation-only candidate. Fresh committed Project Memory
validation is recorded separately after the implementation commit.

The local claim was coordinator-accepted and pinned before a genuine v2 start;
the README was unchanged at that start. Proposal and accepted-source commits
remain in ancestry. All earlier lifecycle files and unrelated claims remain
unchanged. Candidate implementation is ready for independent exact-commit
review, with source publication, PR creation and integration still pending.
