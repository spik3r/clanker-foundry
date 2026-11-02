# PresenceWitness independent requirements

This invented two-state model isolates optional strings, list presence and
boolean composition. Absent strings compare neither equal nor unequal; two absent
operands do not compare equal. Present empty ownership is not absent. Raw string
inequality preserves whitespace, while nonempty uses Unicode whitespace trimming.
Nil and empty acceptance lists both fail nonempty but remain distinct Task values;
a whitespace-only criterion is invalid input rather than a guard result.
All sixteen conjunction operands have a separate false witness; all sixteen
disjunction operands have a separate sole true witness. No expression is a
duplicate and the conjunction is satisfiable. Nested all/any must preserve the
stated conjunction: owner absent OR actor matching owner, AND a nonempty list.
No claim about short-circuit execution order is made for these pure expressions.
Success changes only state and revision; refusals preserve all fields and return
owned storage. These fixtures do not assign owners.

## Frozen independent identity

Canonical model SHA-256: `c93f00645c2e8fcf763d65f294df8f742d024bb519fc0851a9e7320df04253e3`.
Canonical byte length: 5967.
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

Tests: `TestProfileOptionalStrings`, `TestProfileListPresence`, `TestProfileBooleanComposition`.

The shared assertion compares all public Result fields and deep Task content,
checks task/command/review nonmutation, and mutates returned and original owner
and acceptance storage to expose aliases on success and refusal. Nil and empty
acceptance slices are deliberately distinct in expected Tasks.
