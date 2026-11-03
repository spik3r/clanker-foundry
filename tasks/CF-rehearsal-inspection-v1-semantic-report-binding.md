---
id: CF-rehearsal-inspection-v1
status: in-progress
owner: Codex rehearsal report inspection contributor
repositories: [spik3r/clanker-foundry]
---

# Inspect carried rehearsal reports against their literal suites

## Implementation claim

Owner: Codex rehearsal report inspection contributor, branch
`codex/rehearsal-inspection-v1`, based on reviewed main `634427c`.
The repository owner requested selection and implementation of larger follow-up
tasks through subagent PRs, with lead review and merge before new assignments.
Integration owner: the lead Codex agent. This worker may publish its branch and
PR; only the lead may merge after independent review and passing checks.

Exact scope: new files under `modeling/lifecycle/rehearsalinspection/` and
`modeling/lifecycle/cmd/modelrehearsalinspect/`, additive
`modeling/lifecycle/scenarios/report_inspection.go` and
`modeling/lifecycle/scenarios/report_inspection_test.go`, additive
`modeling/lifecycle/rehearsalhandoff/material.go` and
`modeling/lifecycle/rehearsalhandoff/material_test.go`, this task record and its
immutable Project Memory source/events. Preserve all existing qualification
source-policy bytes, legacy carrier and receipt formats, generated artifacts,
runtime, CI and the separate process-view PR20.

## Outcome and acceptance

Add a pure, bounded inspection over an existing verified rehearsal carrier and
separately supplied references. Validate each carried report's model, original
suite bytes, scenario digest, ordered commands and literal expectations against
its verified suite; inspect existing native report consistency and input flow.
Retain valid early mismatch observations without calling them execution proof.
Malformed, substituted or inconsistent reports must refuse, including when a
caller updates their external byte hashes. The legacy verifier keeps treating
report bytes as opaque, as its current contract states.

No process, provider, network, filesystem path or reducer callback is invoked by
the inspector. Receipts say example-only, authority none and execution not
established. No inspection grants approval or authenticates execution. Source
pins and supplied references remain consistency checks rather than provenance.

A separate contributor will freeze literal observation vectors before production
implementation. Run exact Go 1.27.3 race tests and vet, pack validation, immutable
committed-diff evidence validation and whitespace checks. Test valid matches,
valid early mismatch, wrong identities, changed commands/expectations/inputs,
truncated and reordered entries, malformed input, limits and output failures.
