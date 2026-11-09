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
The TypeScript return type remains `boolean`, not a type predicate. The model
source hash changes if any model byte changes, including whitespace or property
order. This version does not claim semantic normalization across source files.
