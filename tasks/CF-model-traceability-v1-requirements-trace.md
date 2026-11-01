---
id: CF-model-traceability-v1
status: in-progress
owner: Requirements-traceability worker
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Requirements-to-model traceability for the Go lifecycle demo

## Implementation claim

A cloud agent session owns this bounded slice on branch
`claude/pensive-bohr-acjpnp`, based on main
`f808cb457da21bcf2adef3f63b727aa442e5cc39` (merge of PR 14). It is a manual
producer claim, not managed dispatch, a production API or a second lifecycle
authority.

Goal: a strict, versioned JSON file of human-stated outcome requirements with
stable IDs, each explicitly linking to existing model elements (states,
transitions, guards, model requirements, by their existing IDs) and to existing
independent acceptance cases (the `caseId` catalog in `examples/oracle.json`).
Deterministic Go validation separates invalid input (structure, duplicate IDs,
dangling or mistyped references, duplicate or ambiguous links) from incomplete
coverage (no model link, no acceptance link, only pending acceptance). A
generated JSON report, Markdown table and inert SVG view show requirements,
links and untraced items. Links are reviewer declarations; the report does not
claim that a link proves a requirement correct or satisfied.

Inspected before claiming: main at the base above, task records, open PR 12
(scenario replay and model diff: `impact/`, `scenarios/`, `cmd/modelplay`,
`cmd/modeldiff`, `impactacceptance/`, `scenarioacceptance/`), open PR 13 (agent
model mapping: `scripts/agent-model.sh`, `agents/`), open PR 5 (Project Memory
CI) and all remote branches. No traceability claim, branch or PR exists. The
model patch, scenario explainer, exporter, profile catalog and review-pack lanes
named in the request have no claim, branch or PR visible here; recheck before
publication. If a claim appears on the paths below, stop and report the overlap.

Exact write scope under `modeling/lifecycle/`, all new files:
- `traceability/trace.go`, `traceability/render.go`, `traceability/trace_test.go`
- `traceability/README.md`, `traceability/requirements.json`
- `traceability/view/trace.json`, `traceability/view/trace.md`, `traceability/view/trace.svg`
- `traceability/testdata/` negative and reordered fixtures
- `traceabilityacceptance/acceptance_test.go`, `traceabilityacceptance/expected.json`
- `cmd/modeltrace/main.go`, `cmd/modeltrace/main_test.go`

Also this task record and new immutable Project Memory source/event files only.
No edit to the model, compiler, generated reducer or views, `strictjson`,
`jsonformat`, the existing oracle, raw-input adapter, workflows, runtime or any
existing claim. No network, no HTTP endpoint.

## Acceptance

- Reject malformed or trailing JSON, unknown or duplicate keys, nulls, wrong
  version, wrong model or catalog identity, invalid or duplicate requirement
  IDs, blank or duplicate statements, unknown link kinds, dangling model or
  acceptance references, kind mismatches and repeated links. Report all issues,
  sorted, before any report is produced.
- Valid input with gaps produces a report; gaps are listed per requirement and
  never reported as invalid input.
- Report bytes are identical across repeated runs and across reordering of
  requirements and links in the input.
- Expected outcomes for the example and negative fixtures are written and
  committed before the implementation.
- Existing model, generated outputs and oracle tests pass unchanged.
- Invented data only. Output says `example-only`, `authority=none`. No SysML or
  UTP conformance claim. No merge, approval, deployment or auto-merge.
