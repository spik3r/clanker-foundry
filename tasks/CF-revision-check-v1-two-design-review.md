---
id: CF-revision-check-v1
status: in-progress
owner: Codex revision-check contributor
repositories: [spik3r/clanker-foundry]
---

# Review two supplied design revisions through existing checks

## Implementation claim

Owner: Codex revision-check contributor. Branch `codex/revision-check-v1`,
reviewed base `83007fa`. The repository owner requested another wave of bounded
larger tasks after review and integration of the previous work. The lead approved
this contract before publication. This worker may publish a PR; the lead owns
independent review and merge. Other current claims, including PR56 proposal replay
report binding, remain outside this work.

Exact scope: new `modeling/lifecycle/revisioncheck/check.go`, `check_test.go`,
`oracle_test.go`, `README.md` and `testdata/observations.json`; new
`modeling/lifecycle/cmd/modelrevisioncheck/main.go` and `main_test.go`; this task
and its immutable Project Memory records. No changes to existing native packages,
frozen fixtures, source policies, firstdesignreview, proposalreplay, scenarios,
process views, fixed revisionreview walkthrough, runtime or CI.

## Contract and result

Accept strict `lifecycle-revision-check-request/v1` JSON with before/after
snapshots. Each snapshot contains three required text fields: a complete native
`firstDesignRequest`, trace `requirements`, and trace `catalog`. Keep original
native request and artifact bytes. Run firstdesignreview.Check on each side,
intakeimpact.Execute on the original intake requests and traceimpact.Execute on
the proposal models plus supplied trace documents. Preserve all four full native
report texts and SHA-256 observations. Compare each intake's observed requirements
with the supplied trace document by exact decoded UTF-8 source bytes; report
matches, mismatch or missing without treating equality as semantic agreement.

Preserve native open questions, model/suite comparisons and stale preflight
statuses. Produce no aggregate readiness or approval decision. Report example-only,
authority none, execution not-run and review not-established. No reducer, scenario,
process, provider, network or caller-selected callback runs. Native failure rejects
the wrapper with a fixed diagnostic; it does not reflect request bytes.

Bound the wrapper at 32 MiB and output at 64 MiB. Use actual native per-input
limits and preserve their narrower refusal behavior; a valid native input can
still exceed this composition's combined limit. CLI exit 0 means a report, 2 means
invalid input or a limit, 4 means a native/encoding boundary failure, and 3 means
failed or short output, including real closed stdout.

## Acceptance and checks

A separate contributor freezes custom two-revision fixtures and literal expected
relationships before production implementation or replay. Check nondefault models,
unchanged/changed source, missing/mismatched requirements observations, open
questions, stale manifests, exact native report bytes/hashes, strict wrapper and
native refusals, limits, zero partial output, complete CLI output and closed stdout.
Run exact Go 1.27.3 focused race tests, vet, pack validation, whitespace and full
committed-diff Project Memory validation. Record observed results and limitations;
local implementation completion does not declare integration.
