---
id: CF-json-enum-schema-v1
status: in-progress
owner: Codex authoring maintenance coordinator
repositories: [spik3r/clanker-foundry]
---

# Reject duplicate string enum values in the authoring schema

## Implementation claim

The owner authorized continued bounded maintenance with independent review and
merging after validation. Main at `7b5a2d6`, live claims and open PRs were
reviewed. Other active work changes output-root checking, class layout and
public authoring guidance; none owns the two implementation files below.

The json-boolean authoring schema accepts duplicate string enum values that
the authoritative compiler refuses. Its description and agreement fixture
incorrectly call uniqueness inexpressible. A finite local Ajv probe accepted
`["a", "b", "a"]` under the existing schema and rejected it after adding
`uniqueItems: true`, while still accepting distinct values.

Add the standard uniqueness constraint to the enum array and correct its
schema descriptions. Change the existing duplicate fixture from an admitted
schema limitation to a mutual refusal. Cover empty and escaped-equivalent
duplicates, a nested enum, and distinct values that differ in case or Unicode
normalization. Preserve the profile's exact string comparison and current
compiler validation; no generator version or emitted bytes change.

Scope: `modeling/jsonvalidator/authoring/json-boolean-v1.schema.json`,
`modeling/jsonvalidator/authoring/schema.test.cjs`, this task and new immutable
Project Memory files. Boolean-constant wording and command guidance belong to
the separate authoring-guidance task.

Checks: demonstrate the stricter agreement fixtures fail before the schema
fix; run the full JSON schema agreement suite against the actual generator,
its Go model refusal tests, repository validation, whitespace and committed
Project Memory checks. Independently review the final commit and await all
hosted checks before merging.
