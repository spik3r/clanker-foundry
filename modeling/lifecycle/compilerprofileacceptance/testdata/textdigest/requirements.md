# Frozen text and binding requirements

This invented model gives each binding a distinct trigger and guard. Its thirteen
guards exercise all fifteen profile references, including the digest pairs and
three uint64 references. Different task, owner, actor, reviewer and review-task
labels prevent accidental aliasing from passing. Positive and discriminating
negative cases use legal task IDs, review decisions and lowercase digest syntax.
Absent review values fail comparisons; absent is not the empty string.

Task text is exact Unicode: composed and decomposed accents are not normalized.
Nonempty text trims Unicode whitespace, not just ASCII; acceptance nonempty is
length-based but each entry must independently satisfy input validation.
Invalid UTF-8 and blank task IDs are preflight invalid inputs, not guard witnesses.
The nil acceptance and empty acceptance representations hash alike, while nil
and empty owners hash differently. Acceptance order and decimal uint64 precision
are part of the hash. Digest checks authenticate neither actor nor history.

Independent model identity uses Python json/hashlib, removes formatVersion, sorts
semantic collections by ID and transition guard references, sorts object keys,
preserves effects/comparison operands, emits compact UTF-8 without a newline and
escapes U+2028/U+2029. No all/any occurs in this model.
Task hashes independently encode fields in this order: id, state, revision as a
quoted unsigned decimal string, outcome, acceptance, owner. Nil acceptance is [];
owner is null or a string. HTML characters stay literal; U+2028/U+2029 are escaped.
No Foundry helper or generated reducer was called to obtain expectations.

Frozen canonical model SHA-256: `9995b9e76a43b07b7b55ddeda0448e2a02c90b71344669bbdc73244233877866`.

Frozen task digest vectors:
- base: `ffee8398cb44cff6e329036a7b1d24c040b011a3a1450d9b6753d00796fe6a8d`
- reversed: `a1abab37e3a48897cd0b962215ea58089ba10f0d411ec224985db92eb5dcf92f`
- nil_acceptance: `f78253a6905a4674693a14f5c414592797aefbf0d1da7448db43e78edd8f3f5b`
- nil_owner: `8c71f86ce3d14f80a604140c04e171ede3fabd174a76ef64d5d74a12f8651b56`
- empty_owner: `680c5d52646a39ecc62d15e2f4d7bcee5d74e69af5e29656899ddf4231236e04`
- decomposed: `75269c56b20763d4fe4063fe252a9e3e9fbb0218f614cb3c0a9f12ab5cdd23f9`
- max_revision: `261bce177be167221546a750c48f04f940763922696f9a324a7651f39538979d`
