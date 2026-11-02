---
id: CF-lifecycle-clean-regeneration
status: in-progress
owner: Codex reproducibility contributor
repositories: [spik3r/clanker-foundry]
---

# Verify clean lifecycle regeneration

## Implementation claim

On `codex/lifecycle-clean-regeneration`, starting from main `9d06274`, add a
validation-only, offline integration test that copies the lifecycle module into a
disposable directory, removes the eight committed generator outputs, regenerates
both compiler and requirements trace outputs with the pinned Go toolchain, and
compares their exact bytes with the originals across two generations. Exercise
frozen acceptance and authoring Go tests against the regenerated copy and verify
that schema, model, catalog, requirements and expected-result files are unchanged.
Use fixed explicit output names; detect drift in each output without refreshing
frozen expectations. The existing schema CI still checks the schema with Ajv.

Scope: new `modeling/lifecycle/reproducibility/` test and README, this record,
and its own immutable Project Memory files. Existing CI runs `go test ./...`;
no production runner, generator change, deployment or activation. Exclude PR 20,
intake, model authoring, held reviewpack/modelintakecheck and tutorial lanes.
Main, open PRs, published branches and local claims were inspected; a local
claim is not a shared lock. Recheck live work before publication.
