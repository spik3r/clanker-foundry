---
id: CF-diagram-json-unicode-v1
status: in-progress
owner: Codex coordinator with delegated builder
repositories: [spik3r/clanker-foundry]
---

# Refuse malformed Unicode escapes in diagram inputs

## Implementation claim

The owner authorized continued Foundry and related UI library improvements with
cheaper agents, independent review and tested merges. The coordinator selected
this separate diagram-module defect after Foundry PR 183 and Flexgo PR 24 were
merged and their exact trees verified. Base is
`fbfd2b1b38af1340bba5c01cbb3cf95c21061022`. No open Foundry PR or overlapping
claim was found. The completed TUI repair has no claim on this Go module.

An independent temporary probe exercised process, dashboard, sequence and class
models. A lone high or low surrogate in a title was accepted and replaced by
U+FFFD. Its canonical digest and every generated file matched valid replacement
text; the source digest differed. Valid pairs and literal backslashes worked.
The common StrictScan checks raw UTF-8 but not escaped surrogate pairs.

The coordinator owns branch `codex/diagram-json-unicode` and integration.
The delegated builder may edit only:

- `modeling/diagram/contract/strict.go` to invoke a local Unicode guard;
- new `modeling/diagram/contract/unicode.go` and `unicode_test.go` for the guard
  and scanner boundaries;
- new `modeling/diagram/unicode_test.go` for all four profiles and the existing
  dashboard binding/Tesserae routes;
- new `modeling/diagram/cmd/diagramgen/unicode_test.go` for CLI refusals;
- `modeling/diagram/README.md` and `modeling/diagram/REPORT-BINDING.md` for the
  input contract;
- this task's observed result and new immutable Project Memory source/events.

Reject unpaired high/low, reversed and broken surrogate escapes in keys and
values before the profile or binding decoder can silently replace them. Keep
valid pairs, literal/escaped U+FFFD and escaped literal backslashes. Preserve
size, depth, shape and syntax refusals, including existing profile selection.
Unicode-guard errors use fixed text without the offending input. Do not broaden
this into a change to unrelated profile-selection diagnostics.

Valid inputs must keep their prior canonical/source digests, identities and
emitted bytes. No renderer, template, manifest format, version, profile, schema,
dependency, consumer pin, runtime, TUI or JSON-validator change. In particular,
the JSON-validator profile deliberately rejects U+FFFD; its contract is separate.
Keep the helper local to this Go module, with no new cross-module abstraction.

## Validation and handoff

Commit the reviewed public-safe claim, register and independently check its
pinned source, then record a genuine v2 start before implementation edits.
Freeze regressions and prove them red on unchanged production before repair.
Check scanner edge cases, nested model text across every supported profile,
binding payload strings and all affected CLI verbs. Refused input emits no
success identity or transformed document and does not create or modify output.
Compare valid Unicode inputs and complete generated files before and after.

Run focused and full diagram race tests, vet/build and gofmt with Go 1.27.4;
repository validation, validator/installer/setup/lifecycle regressions; relevant
HTML/Solid acceptance checks from the existing workflow; whitespace and full
committed Project Memory validation. Record observed checks and limitations.
A separate reviewer and the coordinator review the exact final commit. Wait for
hosted gates before normal merge; do not tag, release or activate runtime.
