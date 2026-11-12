# Modeling

This directory holds Foundry's model-driven generators. An author, human or AI,
writes a declarative model in JSON. A reviewed generator turns the model into
code or documents. The direction is in the
[generator pipeline plan](../docs/generator-pipeline-plan.md).

## Profiles

A profile is one bounded modeling language with one generator. It fixes which
models are valid and what the generator emits. A requirement the profile cannot
express is a gap to report, not something to work around in generated output.

[profiles.json](profiles.json) is the registry of profiles. For each profile it
gives a one-sentence summary, the generator, the profile document, the authoring
schema, example models, what is emitted and what the profile cannot express.
Read the registry first, then the profile document, then draft a model against
the schema. Go validation in the generator is authoritative; a schema pass is
only a first check.

| Profile | Generator | Summary | Profile document | Schema |
| --- | --- | --- | --- | --- |
| `record-browser/v1` | `tuic` | Read-only list and detail browser for a terminal, as a Go package. | [PROFILE.md](tui/PROFILE.md) | [schema](tui/authoring/record-browser-v1.schema.json) |
| `json-boolean/v1` | `jsonvalidator` | Boolean predicate over parsed JSON, emitted as Go and TypeScript functions. | [PROFILE.md](jsonvalidator/PROFILE.md) | [schema](jsonvalidator/authoring/json-boolean-v1.schema.json) |
| `lifecycle/v1` | `lifecyclegen` | Finite task lifecycle with guards, transitions and ordered effects, as a Go reducer with views. | [README](lifecycle/README.md) | [schema](lifecycle/authoring/lifecycle-v1.schema.json) |
| `process-diagram/v1` | `diagramgen` | Business-process diagram of lanes, events, tasks, gateways and flows, as SVG and BPMN-style XML. | [PROFILE.md](diagram/PROFILE.md) | [schema](diagram/authoring/process-diagram-v1.schema.json) |

The harness's own `reference/v1` generator exists to test the contract and is
not listed.

## Contract and harness

All generators follow the [generator contract](CONTRACT.md): `validate`,
`generate` and `check` commands, an identity line, a manifest and drift
detection. `validate MODEL` is the command to check a draft model; it exits 0
for an accepted model and 2 for a refused one.

The [conformance harness](gencontract/README.md) builds each generator listed in
[generators.json](gencontract/generators.json) and checks it against the
contract. A test there also keeps `profiles.json` consistent with that registry
and with the files it names.

## Using generators in a project

[foundrygen](foundrygen/README.md) is the project entrypoint. A project lists
its models and outputs in `foundry.json`; `foundrygen generate` rebuilds the
outputs and `foundrygen check` fails when any output has drifted from its model.

## Adding a profile

Add a generator to [generators.json](gencontract/generators.json), then add one
entry to [profiles.json](profiles.json) with a profile document, an authoring
schema and at least one example. List only limits that the profile document or
validator states. Give the schema a test that compares its verdicts
with the generator's `validate` exit code, as the tests in `tui/authoring`,
`jsonvalidator/authoring` and `diagram/authoring` do.
