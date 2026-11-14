---
id: CF-json-schema-text-v1
status: completed
owner: Codex schema text contributor
repositories: [spik3r/clanker-foundry]
---

# Reject unsupported model string characters in the authoring schema

## Implementation claim

Owner: Codex schema text contributor, delegated by the coordinating agent
under the owner's continued maintenance request. Main at `ccad85b`, direction
and claims were reviewed. Enum uniqueness PR #146 is merged. Other active work
owns generator output paths and class labels, with no overlap here.

The JSON boolean authoring schema admits control characters, U+FFFD and lone
surrogates in string constants and enum values, though Go refuses them. Its
description incorrectly lists this as a rule schemas cannot express. Finite
independent probes matched Go with a standard not/pattern constraint, including
paired and lone surrogates in Unicode and non-Unicode regular expression modes.
The 256-byte UTF-8 limit remains a separate gap: schema maxLength counts code
points, so 100 euro signs pass it while Go refuses their 300 bytes.

Scope: `modeling/jsonvalidator/authoring/json-boolean-v1.schema.json`,
`modeling/jsonvalidator/authoring/schema.test.cjs`, this task and new immutable
Project Memory files. Share one model-string definition between string const
and enum values. Reject C0, DEL/C1, U+FFFD and unpaired surrogates without
rejecting valid astral characters, U+2028/U+2029, empty strings, case differences
or normalization-distinct strings. Keep maxLength 256 and state its remaining
byte-count gap. Add agreement fixtures, preserve raw surrogate spellings and
check both regex modes. No Go, version, generated-output or policy changes.

Checks: show new schema refusal regressions fail before the fix; full schema
agreement suite with actual Go validation; existing Go model-refusal race tests;
repository validation, whitespace and committed Project Memory validation.
Commit, push and open a PR for independent review and hosted checks; merge
belongs to the coordinating agent.

## Integration result

The coordinating agent reviewed `22d7ff9e14e8a3eadf2062903724e12b1670bc6c`;
all ten hosted workflows passed on that head. [PR #149](https://github.com/spik3r/clanker-foundry/pull/149)
merged on 14 November 2026 as `74e36ac4ebbb071bd63497dd64cd37eb05eee505`.
This bounded implementation claim is complete and reserves no further work.

### Local toolchain clarification

The JSON validator module requires and auto-selects Go 1.27.3. Earlier local
validation notes that name Go 1.27.0 report the shell launcher, not the module
toolchain. Module-local version checks confirm Go 1.27.3; the historical
`go.mod` has the same requirement. The passed checks stand, and the immutable
events remain unchanged.
