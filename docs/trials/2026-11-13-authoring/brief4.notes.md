# Brief 4 notes: plugin manifest check

- Profile: `json-boolean/v1`. Chosen because the brief is a boolean check over parsed JSON, which is the profile's purpose. Only part of it fits.
- Covered by `brief4.model.json`: `name` (required string), `version` (required string), `enabled` (required boolean), `tags` (optional array of strings).
- Gaps (two gap reports in `brief4.gap.md`): the three-value `kind` rule (no enums or string constants), and the rule that no other properties are allowed (the profile states that forbidding extra properties is unsupported).
- The model is a partial check and must not be used as the manifest check on its own. It accepts a manifest with any or no `kind` and with any extra properties.
- Assumed: `name` and `version` may be empty strings, since the profile treats empty strings as valid for a string node and the brief does not forbid them.
- Assumed: an absent `tags` is valid and a present `null` is not, which follows the profile's optional-field rule.
- Unclear: the brief says "a function". The profile emits a predicate in Go and TypeScript; I did not decide the function's name or package.
- `json-boolean/v1` has no `cases` member in the registry, so no cases file was written.
- The model was not validated: `jsonvalidator validate` was not run.
