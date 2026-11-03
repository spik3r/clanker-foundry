# IntegerWitness independent requirements

This invented two-state model intentionally permits mismatched revisions on
`advance`; no baseline freshness, outcome, acceptance, actor or owner policy is
implicit. Four triggers isolate uint64 equality/inequality against required
command and optional review operands. True guards change Open to Closed and
increment exactly once, preserving all other fields. Failed guards preserve the
entire task. At maximum revision, guard refusal takes precedence over overflow;
a true guard yields REVISION_LIMIT without effects. Invalid typed inputs take
precedence over both. Distinct adjacent integers above 2^53 remain distinct, and
values above 2^32 do not narrow. Review absence is not numeric zero: both eq and
ne require present operands. Acceptance nil and blank outcome are valid inputs.

## Frozen independent identity

Canonical model SHA-256: `107331db4daf02f01fd418af40fe39a3a375a46b63a7ac07d3c42321b6711914`.
Canonical byte length: 1795.
Computed with an independent Python standard-library encoder: remove source-only
`formatVersion`; sort ID collections and guard IDs; recursively sort pure all/any
operands using UTF-8 byte-length-prefixed expression keys; preserve comparison
operand and effect order; JSON object keys sorted, compact separators, UTF-8,
no HTML escaping, no final newline; SHA-256 of those bytes. These model strings
are ASCII, so no cross-encoder Unicode escaping ambiguity enters this digest.
No compiler, decoder, canonical helper, generated reducer, scenario interpreter
or actual reducer result supplied expected values. Every expected Result is a
source literal, including its model digest. The fixture author selected each
expected branch independently from these requirements.

Tests: `TestProfileIntegerBoundaries`, `TestProfileIntegerOptionalOperands`.

The shared assertion compares all public Result fields and deep Task content,
checks task/command/review nonmutation, and mutates returned and original owner
and acceptance storage to expose aliases on success and refusal. Nil and empty
acceptance slices are deliberately distinct in expected Tasks.
