# JSON boolean profile v1

This bounded profile is named `json-boolean/v1`. It is distinct from Foundry's
`lifecycle/v1` language. The top-level model has exactly `profile` and `root`.
Each node has one of these shapes:

- String: `{"type":"string"}`, with optional extras (see "String rules").
- Boolean: `{"type":"boolean"}`, with an optional boolean `const`.
- Array: `{"type":"array","items":NODE}`.
- Object: `{"type":"object","additionalProperties":true|false,"properties":[...]}`.
  Each property has exactly `name`, boolean `required`, and `schema`.

Model property names must be unique within an object and nonempty ASCII
identifiers, including `_` and `$`, at most 128 bytes long. This restriction
applies to authored model fields, not extra input keys. With
`additionalProperties` true, extra input properties are permitted and are not
traversed. With false they are refused (see "Closed objects").

The model has no defaults, nullability, numbers, regular-expression constraints,
unions, coercion, recursive references, cross-field rules or diagnostic
constructs. Enums exist for strings only; boolean and string nodes support
constants. Unknown fields and unsupported versions are refusals. The
[JSON Schema](authoring/json-boolean-v1.schema.json) is an authoring aid for
drafting a model; Go validation is authoritative and the schema lists the rules
it cannot express.

## String rules

All are optional. A string node with none of them checks only that the value is
a string.

- `const`: a string. The value must equal it.
- `enum`: 1 to 32 unique strings. The value must equal one of them.
- `minLength`, `maxLength`: integers 0 to 65535, written as plain integers (no
  fraction, exponent, sign or leading zero). If both are present, `minLength`
  must not exceed `maxLength`. They may be used together.

The authoring schema sees parsed numeric values, so it accepts spellings such
as `1.0`, `1e0` and `-0` that the compiler refuses. Run generator validation
on the original model text to check the plain unsigned integer rule.

`const` and `enum` cannot be used together or with a length bound. Every `const`
and `enum` string is at most 256 bytes of UTF-8, contains no control character
(C0, DEL or C1) and does not contain U+FFFD. U+FFFD is refused because an
unpaired surrogate escape in a model decodes to it; refusing it avoids a model
that silently means something else. Duplicate enum values are an error.

Equality for `const` and `enum` is exact: the same code units in TypeScript, the
same bytes in Go, with no Unicode normalization and no case folding.
`"caf\u00e9"` and `"cafe\u0301"` are different strings. Model strings contain no
unpaired surrogate and parsed JSON strings in Go are valid UTF-8, so the two
languages agree on equality for every input from `JSON.parse` or `encoding/json`.

Length is counted in Unicode code points. A surrogate pair counts as one. An
unpaired surrogate counts as one. Go and TypeScript agree on this for every
parsed input: Go's JSON decoder turns an unpaired surrogate escape into U+FFFD,
which is one code point, while `JSON.parse` keeps the lone code unit, which
iteration by code point also counts as one. Combining marks count separately.
The Go predicate ranges over runes; the TypeScript predicate iterates the string
by code point. The one input excluded from the cross-language cases is a Go
string that is not valid UTF-8: a JSON decoder cannot produce it, Go would count
each invalid byte as one, and TypeScript has no such value. A lone surrogate
never equals a `const` or `enum` value in either language.

## Closed objects

With `"additionalProperties":false` the predicate also refuses an input object
that has any key that is not a declared property name. The keys examined are the
object's own enumerable string keys: `Object.keys` in TypeScript, every map key
in Go. Declared names are compared by value, so `__proto__`, `constructor` and
`toString` as own keys are refused unless declared, and inherited names are never
counted. Symbol keys, non-enumerable keys and non-JSON objects are outside the
contract. An empty closed object accepts only `{}`. Each object node has its own
`additionalProperties`; closing one does not close its children.

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
satisfy the item schema. Unknown properties of an open object are not inspected or mutated.

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
- String: `{"type":"string"}`, then at most one of `,"const":"x"` or
  `,"enum":["a","b"]`, or the length members `,"minLength":N` and `,"maxLength":M`
  (in that order, each only if present). Strings are written with only `"` and
  `\` escaped; all other characters appear as UTF-8. Enum values are sorted in
  byte order. A one-value `enum` and the same `const` are different canonical
  forms; they are not normalised into each other.
- Boolean: `{"type":"boolean"}`, or `{"type":"boolean","const":true}` (or
  `false`) when a constant is declared.
- Array: `{"type":"array","items":NODE}`.
- Object: `{"type":"object","additionalProperties":true,"properties":[...]}`
  (`false` for a closed object), with each property written as `{"name":"x","required":true,"schema":NODE}`.
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
array `items`, `additionalProperties`, the presence and value of every string
member, and the nesting of nodes. Absent members stay absent: a model that uses
none of the string rules and only open objects keeps the digest it had under
generator 1.0.0, and its emitted predicate behaves the same. The generator
version in headers and manifests is now 1.1.0. Duplicate keys, unknown members and
other refused input never reach the digest.

`sourceSHA256` in the identity line hashes the exact model bytes and does
change with whitespace or ordering. The emitted `ModelSHA256` and
`MODEL_SHA256` constants carry the canonical digest, not the source digest,
because emitted bytes may not depend on source bytes.

## Not expressible

Numbers and numeric ranges, regular expressions or formats, unions, nullable
values, defaults, cross-field rules, recursive references, error messages and
first-error results, and enums or constants of anything but strings.
