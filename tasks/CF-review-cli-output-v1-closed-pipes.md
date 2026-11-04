---
id: CF-review-cli-output-v1
status: in-progress
owner: Foundry review CLI output repair contributor
repositories: [spik3r/clanker-foundry]
---

# Preserve review command output-failure exits

## Implementation claim

Base: main at `249cf7434a43c4d1b0ab6f04b2f0d7d396b7c2ab`. Branch:
`codex/review-cli-output-v1`. The coordinator approved this bounded local repair
after a current-main reproduction. The live open-PR endpoint showed no open PRs;
unmerged branches and canonical claims were checked. The separate unpublished
pilot-guide claim remains untouched. Historical model-impact, intake and trace
tasks remain unchanged. This claim reserves only the repair below, not their
native APIs or any unfinished original scope. Recheck live work before publishing.

Write only `modeling/lifecycle/cmd/modeldiff/main.go`,
`modeling/lifecycle/cmd/modelintake/main.go`,
`modeling/lifecycle/cmd/modeltraceimpact/main.go`, one new
`modeling/lifecycle/clioutputacceptance/output_test.go`, this task, and its new
immutable Project Memory source/events. No other command, native package, source
policy, fixture, generated artifact, visual asset, runtime, README, workflow or
permission changes. Publication and integration remain with the coordinator.

## Observed defect and required behavior

All three selected commands document exit 3 for a failed or short report write.
On the base commit, a compiled command with valid input and a stdout pipe whose
reader is already closed instead terminates with SIGPIPE, with no diagnostic.
Normal report output succeeds. This is inconsistent process-level failure
reporting, not successful delivery, corrupted report acceptance, a semantic model
error or an authorization bypass.

Ignore SIGPIPE in the three process entrypoints, matching the existing newer
commands. Keep each run function, normal report bytes, input refusal, diagnostic
prefix and existing exit meaning unchanged. A closed pipe must reach the existing
write-failure branch, return exit 3 and emit its fixed output-failure diagnostic.
Do not add input-dependent behavior or change library signal handling.

## Acceptance and handoff

Register this reviewed pinned claim and record a genuine start before adding
tests or production edits. A separate contributor writes the process acceptance
tests before the fix. Build the real binaries with the installed Go toolchain;
check normal output against the existing native API, invalid-input exit 2 with no
stdout, and a real preclosed stdout pipe returning exit 3 with the exact fixed
diagnostic. Bound subprocess waits and preserve existing fixtures. The new tests
must fail on the unmodified base before the production edit and pass afterward.

Run Go 1.27.3 offline focused and full race tests, vet, modelc validation/drift,
repository validation, installer regressions, formatting, whitespace and committed
Project Memory checks. Keep scratch/caches outside the checkout. Record failures
and environment limits accurately. Obtain independent exact-source review and
report the local commit and artifact before any publication. No candidate model,
visual review, worker dispatch, deployment or formal PR approval is included.
