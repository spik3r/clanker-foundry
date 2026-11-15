---
id: CF-filter-input-allocation-v1
status: completed
owner: Codex filter-input allocation contributor
repositories: [spik3r/clanker-foundry]
---

# Bound transient allocation while appending filter input

## Implementation claim

Owner: Codex filter-input allocation contributor, delegated by the coordinator
under the repository owner's continued Foundry maintenance request. Reviewed
main: `85be773db2f3530425b767dca710a4cd40872d8f`. Current PRs and active Project
Memory frontiers are empty. Earlier filter, Unicode and offset claims are
completed implementation runs; leave their records unchanged.

Owned, freshly generated standalone and hosted browsers retain the documented
64-code-point query limit, but a 1 MiB printable paste or key-text message
allocates about 4 MiB before keeping that prefix. A control prefix followed
by printable text also causes input-sized allocation. An independent generated
browser measurement confirmed the scaling. This is a measured resource issue,
not a violation of the retained-query limit or a consumer-value result.

Stream the incoming UTF-8 through the existing sanitizer into the remaining
query space. Preserve valid U+FFFD, invalid-byte removal, the existing control
and separator rules, code-point counting and the already-full fast path. Keep
other format characters' current behavior. A rejected prefix must still be
scanned to find later valid text; stop after the available slots fill. Add no
paste-size policy, model member, schema, profile or cases-language extension.
Remove the obsolete whole-string helper if it has no remaining caller.

Publish tuic 1.5.3 because emitted bytes change. Canonical model digests and
profile identity stay unchanged. No generated consumer code or pin is edited.

Scope: `modeling/tui/templates/state.go.tmpl`,
`modeling/tui/acceptance/filter_test.go.txt`, `modeling/tui/generated_test.go`,
`modeling/tui/contract.go`, `modeling/tui/contract_test.go`,
`modeling/tui/hosted_test.go`, `modeling/tui/options_test.go`,
`modeling/tui/README.md`, `modeling/tui/PROFILE.md`, this task and new immutable
Project Memory source/event files. No other generator, runtime, installer,
workflow, dependency or historical evidence changes.

## Validation

Freeze the resource expectations and demonstrate their failure on unchanged
generated code before the repair. For an already-open, single-record filter
at 80 by 12, allocate each input outside measurement and require less than
64 KiB allocated per Update for a 1 MiB printable paste, a 1 MiB key-text
message and a 1 MiB control prefix followed by 64 printable code points. Check
both embed modes, exact query text and nil commands. Do not gate elapsed time.
This bounds the tested sanitizer work, not input-string allocation, record
matching, global memory or the scan time of a rejected prefix.

Test empty, full and partly filled queries; mixed valid/invalid UTF-8; valid
replacement characters beside invalid bytes; controls and line separators;
supplementary and combining characters; backspace and repeated events. Preserve
matching, selection, help, quit and hosted behavior. Run full TUI race tests and
vet, Node 24 authoring tests, generator-contract and foundrygen race tests and
vet, repository validation, whitespace checks and fresh committed Project Memory
validation. Verify canonical digests, deterministic output and updated manifests.
Publish for independent exact-head review and hosted checks; the coordinator
alone owns merge.

## Observed implementation result

The append step now decodes into a fixed 64-rune buffer, skips exactly the prior
invalid-byte/control/separator set, and stops when the remaining query space is
filled. The obsolete whole-string cleaning helper is removed. Tuic is 1.5.3;
model digests, profile identity, dependencies and consumer pins are unchanged.

Before the repair, the original acceptance bytes failed all three allocation
conditions in both generated embed modes. Printable 1 MiB inputs allocated about
4,195,800 bytes per Update; a rejected prefix allocated about 1,058,600 bytes.
The compatibility literals passed. Original acceptance SHA-256:
`20a0192055a2e5a8f8d276dd195499d855b4d21320b77e022a497f79633247cf`.

Review then found that a detached benchmark failure alone would not fail its
parent test. The test now checks semantics through the parent before measurement,
propagates a callback failure flag, and requires a positive measured iteration
count. The budget and expected strings did not change. Final acceptance SHA-256:
`eccfafb3d546543b97c8a32d2833cf7d6460cc0f5821e7a8e25935a0d79bd9d8`.
This is stronger observation, not a new sanitizer policy or unchanged-test-hash
claim.

Observed with cached offline Go 1.27.4:

- Full TUI module race tests passed (40.909s for the main package).
- Generator-contract and foundrygen module race tests passed (13.657s and 54.079s).
- All three modules' `go vet ./...` checks passed.
- After the final test-only observation changes, generated standalone and hosted
  filter acceptance passed again with race checks (14.345s).
- Node 24.15.0 TUI authoring tests passed: 8 tests, no skips, cached AJV 8.20.0.

The full module passes preceded the final test-only observation changes; the
last generated filter run checks those final bytes. Repository validation and
whitespace results are recorded separately before completion evidence.

Pinned source registration and the genuine start used Node 24.15.0 before the
new tests or implementation. This bounded implementation claim is complete;
independent final review, committed evidence validation and PR publication remain
next. No merge, human usability result, private consumer trial or runtime
activation is claimed.
