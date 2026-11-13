# Gap report: the `kind` value must be one of three strings

## Requirement

> `kind` which must be exactly one of the strings "scanner", "reporter" or "exporter"

Source: brief 4 (plugin manifest check), the list of manifest members.

## Profile considered

- Profile: `json-boolean/v1`
- Registry entry and `cannotExpress` line that applies: "defaults, enums or nullable values"; "unions or recursive references".
- Profile document or schema section read: `modeling/jsonvalidator/PROFILE.md` (opening node shapes; "The model has no defaults, enums, nullability ..."; "Parsed-value semantics"); `modeling/jsonvalidator/authoring/json-boolean-v1.schema.json`.

## Missing construct

The string node is `{"type":"string"}` and checks the type only. There is no member that restricts a string to a set of values, no `const` for strings (only booleans have `const`), and no union of alternatives. The model cannot say "a string equal to one of these three values".

## What was not done as a result

- The model omits `kind` entirely. It is not written as a plain string, because that would accept any string and look like a check.
- Consequently the emitted predicate accepts a manifest with any `kind`, or none, and does not enforce the rule.
- Nothing in the generated code checks the three allowed values.

## Options

- Extend the profile: add a string enum (or string `const`) to the node shapes, the schema, the canonical encoding and the generator.
- Hand-written seam owned by a person: a separate check on `kind` in the application code, next to the generated predicate, with its own tests.
- Leave the feature out: the loader accepts any `kind`, and the three-value rule is not enforced.

Decision owner: unknown

Decision: (blank until the owner decides)
