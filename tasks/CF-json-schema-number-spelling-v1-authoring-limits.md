---
id: CF-json-schema-number-spelling-v1
status: in-progress
owner: Codex schema maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# State the JSON authoring schema's number spelling limit

## Implementation claim

Owner: Codex schema maintenance contributor, delegated by the coordinating
agent under the owner's continued maintenance request. Main at `c87356a`, the
direction plan and claims were reviewed. Open PRs #142 and #143 do not overlap.
A separate queued JSON validator symlink fix owns command code, excluded here.

The JSON boolean profile requires minLength and maxLength to be plain unsigned
integer source text. JSON.parse turns 1e0 and 1.0 into 1 and preserves -0 as a
numeric zero, so the authoring schema accepts those spellings while Go rejects
them. The schema lists its limits but omits number spelling. Current agreement
fixtures use JSON.stringify for length bounds and erase the spelling difference.
Independent probes confirmed schema acceptance and Go exit 2 for all three.

Scope: `modeling/jsonvalidator/authoring/json-boolean-v1.schema.json`,
`modeling/jsonvalidator/authoring/schema.test.cjs`,
`modeling/jsonvalidator/PROFILE.md`, this task and new immutable Project Memory
files. Explain that the schema checks parsed numeric values and cannot enforce
source spelling. Add raw-text looser agreement fixtures for minLength and
maxLength with decimal, exponent and negative-zero spelling. Do not change Go
validation, schema constraints, generated code, generator versions or policy.

Checks: full JSON validator authoring schema suite, relevant Go model refusal
regressions, repository validation, whitespace and committed Project Memory
validation. Commit, push and open a PR for independent coordinator review and
hosted checks; merge belongs to the coordinator.
