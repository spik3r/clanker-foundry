# JSON boolean profile v1

This bounded profile is named `json-boolean/v1`. It is distinct from Foundry's
`lifecycle/v1` language. The top-level model has exactly `profile` and `root`.
Each node has one of these shapes:

- String: `{"type":"string"}`.
- Boolean: `{"type":"boolean"}`, with an optional boolean `const`.
- Array: `{"type":"array","items":NODE}`.
- Object: `{"type":"object","additionalProperties":true,"properties":[...]}`.
  Each property has exactly `name`, boolean `required`, and `schema`.

Model property names must be unique within an object and nonempty ASCII
identifiers, including `_` and `$`, at most 128 bytes long. This restriction
applies to authored model fields, not extra input keys. Extra input properties
are always permitted and are not traversed. Requests to forbid extra properties
are unsupported.

The model has no defaults, enums, nullability, numeric or regular-expression
constraints, unions, coercion, recursive references, cross-field rules or
diagnostic constructs. Unknown fields and unsupported versions are refusals.

## Model intake

The compiler accepts at most 1 MiB of UTF-8 JSON, structural JSON depth 64,
schema depth 32, 512 schema nodes and 128 declared properties per object.
Duplicate JSON keys, trailing values, missing fields and malformed types refuse.
No model string is executable code. These are compiler intake limits, not new
limits on the values accepted by the emitted predicate.

## Parsed-value semantics

Both outputs expose a boolean predicate for already-parsed ordinary JSON.
Go uses `map[string]any`, `[]any`, strings, booleans, nil and decoder-produced
numbers. Use `json.Decoder.UseNumber` when ignored numeric lexemes must survive
without a float64 overflow failure in parsing. TypeScript accepts the result of
`JSON.parse` as `unknown`. Parsing policy stays with the caller.

Null does not satisfy object, array, string or boolean nodes. Required fields must
be present as own properties and satisfy their schema. Optional fields may be
absent; present null still has to satisfy the schema and therefore fails. Empty
strings and empty arrays are valid for their shapes. A boolean constant compares
exactly; neither numbers nor strings count as booleans. All array elements must
satisfy the item schema. Unknown properties are not inspected or mutated.

Declared names such as `__proto__`, `constructor` and `toString` follow ordinary
own-property presence rules; inherited JavaScript values do not supply a required
field. Non-JSON values, including getters, proxies, custom prototypes, cycles,
sparse arrays, functions and explicit `undefined`, are outside the contract.
Alternative typed Go maps and slices are also outside it.

The predicate gives no transport or error policy and no first-error result.
The TypeScript return type remains `boolean`, not a type predicate. The generator follows the [Foundry generator contract v1](../CONTRACT.md).

## Canonical model and digest

`canonicalSHA256` is the SHA-256 of this canonical encoding of the parsed model:
compact JSON with no insignificant whitespace, no trailing newline, and members
in a fixed order.

- The top level is `{"profile":"json-boolean/v1","root":NODE}`.
- String: `{"type":"string"}`.
- Boolean: `{"type":"boolean"}`, or `{"type":"boolean","const":true}` (or
  `false`) when a constant is declared.
- Array: `{"type":"array","items":NODE}`.
- Object: `{"type":"object","additionalProperties":true,"properties":[...]}`,
  with each property written as `{"name":"x","required":true,"schema":NODE}`.
  Properties are sorted by name in byte order.

Not significant, so they do not change the digest or the emitted bytes: JSON
whitespace, the order of members within the top-level object and within every
node and property object, `\u` escapes in strings, and the order of properties
within an object. Property order is not significant because the predicate is a
conjunction of independent per-property checks; no property's result depends on
another's, so the result is the same in any order. The emitter writes
properties, and so numbers nodes, in canonical order, so models with the same
canonical digest produce identical files.

Significant: the profile, every node type, every property name, `required`, the
presence and value of `const` (an absent `const` differs from `true` or `false`),
array `items`, and the nesting of nodes. Duplicate keys, unknown members and
other refused input never reach the digest.

`sourceSHA256` in the identity line hashes the exact model bytes and does
change with whitespace or ordering. The emitted `ModelSHA256` and
`MODEL_SHA256` constants carry the canonical digest, not the source digest,
because emitted bytes may not depend on source bytes.
