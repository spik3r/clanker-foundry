---
id: CF-revision-check-v1
status: done
owner: Codex revision-check contributor
repositories: [spik3r/clanker-foundry]
---

# Review two supplied design revisions through existing checks

## Integration result

[PR60](https://github.com/spik3r/clanker-foundry/pull/60) merged as
`bd134f661b5b6163493d9c5ead9f7d4e7125b57a` after independent source review
and all five hosted workflows passed on
`4d1da3c52430dcfe1bb180a9ac638f032e520e7c`. The lead tested the ordinary-merge
union with PR57 using focused Go race and runtime trace checks plus committed
evidence validation. Its tree `45a4afad806d597ced19569e28c7357c52eceff0`
matches the final merge exactly. The completed claim reserves no further work;
the checker establishes no review or execution authority.

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


## Implementation and review result

Source `578dbce` adds the bounded composition, CLI, tests and guide. Documentation
`ce164ab` adds an explicit output-directory guard. The independent custom vectors
were frozen first in `82b80fc`, with unchanged file SHA-256
`7987c38be6221b2ae1a44b45cfd0ab81fbe7dcc70dbab7ff3f9ed18fd4caa669`.
The examples retain a nondefault model, Unicode label revision, stale original
suite/manifest, source mismatches and open questions without claiming execution.

Observed validation: exact Go 1.27.3 offline race tests passed for revisioncheck,
cmd/modelrevisioncheck, firstdesignreview, intakeimpact and traceimpact. Whole-module
`go vet ./...`, pack validation and whitespace checks passed. CLI checks covered
complete output, short writes and actual closed stdout exit 3. The guide's build
and command were separately run against the frozen custom request; its candidate
preflight remained stale and replay not-run. Private temporary output was removed.
Optional skills-ref was unavailable and reported skipped.

A fresh source reviewer checked exact `ce164ab`, found no blocking findings and
independently passed the same five-package race checks. Ordinary main merges
`c49d3c6` and `30582b2` add only integrated task closeout records and the current
workflow guide; every claimed implementation/fixture byte remains identical to
reviewed `ce164ab`. No existing native package, source policy, fixture, runtime or
workflow file was edited by this task. Full committed-diff evidence validation is
run after the completion commit. Hosted checks and the lead's final review/merge
remain pending; this record does not declare integration or release its claim.
