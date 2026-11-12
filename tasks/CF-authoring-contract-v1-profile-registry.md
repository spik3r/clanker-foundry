---
id: CF-authoring-contract-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add the profile registry and authoring schemas

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This is step 2 of the
[generator pipeline plan](../docs/generator-pipeline-plan.md).

Scope: `modeling/profiles.json`, a registry of the four profiles with their
generator, profile document, schema, examples and stated limits, and a
`modeling/README.md` index. Authoring JSON Schemas for `record-browser/v1` and
`json-boolean/v1`. Schema tests for those two and for the existing
process-diagram schema, including a comparison of each schema's verdict with
its generator's `validate` result, and the workflow changes to run them. A
tightened label pattern in the process-diagram schema where it was looser than
the generator. A registry consistency test in `modeling/gencontract/`. One
sentence in two profile documents linking the schema. In
`skills/model-driven-design`: a profiles reference, a gap-report template and
minimal edits to `SKILL.md` and `README.md` so the skill covers any registered
profile. Also this task record and new immutable Project Memory files.

Exclusions: no generator behaviour change, no harness change, no edit to
hash-pinned skill examples or to anything under `modeling/lifecycle/`.
Expectation literals as data for the record browser remain open from step 2.

Checks: the node schema tests, `go test -race` in `modeling/gencontract`, the
workflow inventory test, the repository validator, whitespace check and
committed Project Memory validation.
