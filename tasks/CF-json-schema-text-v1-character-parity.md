---
id: CF-json-schema-text-v1
status: in-progress
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
