---
id: CF-rehearsal-inspection-v1
status: review
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


## Review result

Implementation source `0be22ca33dc989eec0b83222f588da84d225809b` adds a pure
report inspector, a no-path stdin CLI, an owned verified-material accessor and a
runner-bound native report decoder. Existing source-policy files, carrier bytes
and legacy verifier behavior are unchanged. Separate-contributor observation
vectors were frozen in `51f55da` before implementation; their SHA-256 is
`fb6433dfe6f2fa1c2e45f8390df9d081bcdab3b007fc9f95abdd9bf066f0efb8`.

Observed checks on that source: Go 1.27.3 offline `go test -race -count=1 ./...`
passed all lifecycle packages with `TMPDIR=/private/tmp`; `go vet ./...` passed;
`go run ./cmd/modelc check` reported matching generated artifacts without writes.
Pack validation, installer regressions and staged/unstaged whitespace checks
passed. Optional `skills-ref` is unavailable and was reported skipped. The CLI
regression checked an actual closed stdout pipe and exit 3, alongside complete
request/receipt output. The separate contributor's literal vectors and updated
external report hashes exercised semantic refusal after legacy verification.

A fresh reviewer and the lead each reviewed exact source `0be22ca`, reported no
blocking findings and independently reran the four affected packages with the
race detector. Their reviews cover source behavior and checks, not authenticated
execution of arbitrary supplied records. The inspector always retains execution
not established and authority none. Hosted CI and merge remain lead actions;
local implementation completion does not declare the PR integrated.
