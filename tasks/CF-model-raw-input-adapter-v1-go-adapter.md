---
id: CF-model-raw-input-adapter-v1
status: in-progress
owner: Raw-input adapter worker
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Offline Go adapter for raw Task and Command JSON

## Implementation claim

A cloud agent session owns this isolated demo on branch
`claude/exciting-einstein-h5r52a`, based on main
`d5e29983bd7e0b49fe05135f6c68fb367a0589c6`. It is a manual producer claim,
not managed dispatch, a production API or a second lifecycle engine.

Goal: accept raw Task and Command JSON bytes, validate their structure with the
existing `strictjson` decoder, then call the existing generated reducer
`generated.Transition` unchanged. Make the pending oracle cases `special-03` and
`special-04` executable in a new acceptance package, without editing
`examples/oracle.json` or `acceptance/oracle_test.go`.

Inspected before claiming: main at the base above, and open PR 12 (scenario
replay and model review diff). PR 12 is a separate interface and touches
`impact/`, `scenarios/`, `cmd/modelplay`, `cmd/modeldiff`, `impactacceptance/`
and `scenarioacceptance/`. This task touches none of them. Other lanes named in
the request (model patch, profile catalog, exporter, visual review pack) have no
claim, branch or PR visible in this repository; recheck before publication. If a
claim appears on the paths below, stop and report the overlap.

Exact write scope under `modeling/lifecycle/`, all new files:
- `rawinput/adapter.go`, `rawinput/adapter_test.go`, `rawinput/README.md`
- `rawinputacceptance/oracle_test.go`, `rawinputacceptance/expected.json`

Also this task record and new immutable Project Memory source/event files only.
No edit to the model, compiler, generated reducer, `strictjson`, existing oracle,
workflows, runtime or any existing claim. No HTTP endpoint, no CLI, no network.

## Acceptance

- Reject malformed or trailing JSON, duplicate or unknown keys, wrong or missing
  fields, invalid UTF-8, invalid surrogate escapes, and invalid or overflowing
  revisions before the reducer is called.
- Documented null versus empty semantics, input-size and nesting limits.
- Structurally valid requests call the reducer and return its result unchanged,
  including guard precedence. The adapter adds no lifecycle rules.
- Independent expected outcomes are written and frozen before the adapter exists.
- Invented data only. Results stay `example-only`, `authority=none`.
- No merge, approval, deployment, worker activation or auto-merge.
