---
id: CF-readme-core-guide-v1
status: review
owner: Foundry documentation editor
repositories: [spik3r/clanker-foundry]
---

# Clarify core navigation and the model-driven pipeline

## Implementation claim

The documentation editor owns this bounded documentation change on
`docs/core-overview-pipeline-examples`, from reviewed main
`7d248947b6b367950988dcc88274358830719c0f`.

Exact scope: `README.md`, `docs/model-driven-pipeline.md`, this task record and
new immutable Project Memory source/event files for this task. The owner requested
four clearly separated core sections above getting started: pipeline,
orchestration, skills and Project Memory, each with one concise explanation and
one detailed guide link. Reduce competing navigation links.

Explain human and AI responsibilities versus deterministic tools, the actual
JSON model content, how its typed representation produces executable state-machine
code, and a small accurate example. Distinguish the implemented bounded generators
and agent-written route from the future idea-to-product goal. Preserve independent
expectations, review and authorization boundaries. Do not change implementation,
models, generated files, tests, workflows, runtime authority or earlier evidence.

Current main and open pull requests were checked before claiming these paths.
No open documentation PR for this request was found. Existing model-map,
agent-selector and JSON regression work remains separate.

## Verification and handoff

Review this public-safe claim before registration and record a genuine start
before implementation edits. Check documentation against current source, compare
the example with the existing model, and validate/generate its supported shape
with the pinned offline Go toolchain where available. Run pack/link validation,
whitespace checks and full committed Project Memory validation. Arrange an
independent read-only factual and clarity review of the final docs.

Publish a draft pull request when the reviewed checks pass, verify the actual
remote head and report hosted CI separately. Preserve the source commit ancestry.
No merge, release, deployment, consumer activation or runtime dispatch is in scope.

## Result and observed checks

The main README now presents four distinct core sections before setup, with one
sentence and one guide link each. The pipeline guide explains actual human/AI/tool
responsibilities, the complete existing FirstDesignDraft JSON, typed IR and fixed
Go generation, concrete outcomes, the separate JSON predicate and agent-written
routes, and the limits of current demonstrations.

Independent read-only review found no blocking factual or clarity issues.
The embedded JSON exactly matches the existing model as data. Offline Go 1.27.3
accepted it and emitted identical code/tests twice; six explicit accepted/refused
cases and the three generated structural tests passed in a disposable module.
Pack/link validation and whitespace checks passed. Optional skills-ref was not
installed. No product code, model, generated artifact, test or workflow changed;
full runtime and installer regressions were not rerun for this documentation edit.

The reviewed task source was published as
`90bb1233fa055a2c0ad2b3d81c6136bf817a33d8` and checked through Git before this
checkout registered it and recorded its genuine start. An earlier local-only
preparation encountered missing Git CLI push credentials and was preserved without
implementation edits. The GitHub connector publishes this branch; its actual
remote head, committed evidence and hosted checks are verified separately.
