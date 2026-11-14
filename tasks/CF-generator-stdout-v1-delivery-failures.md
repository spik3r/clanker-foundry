---
id: CF-generator-stdout-v1
status: in-progress
owner: Codex generator stdout contributor
repositories: [spik3r/clanker-foundry]
---

# Report generator stdout delivery failures

## Implementation claim

Owner: Codex generator stdout contributor, delegated by the coordinating agent
under the owner's continued maintenance request. Main at `74e36ac`, direction
and current claims were reviewed. Active Foundry path-alias work is excluded;
its stdout handling is a separately queued follow-up after that work merges.

Finite probes of all five registered generators found that closed stdout pipes
terminate them with SIGPIPE. Four also return exit 0 on a failed writer;
lifecyclegen reports writer errors but accepts a short write with no error.
All five accept such short writes. The generator contract declares I/O failure
exit 1. Existing lifecycle commands show the intended process-entrypoint pattern.

Scope: `modeling/CONTRACT.md`, `modeling/gencontract/README.md`,
`modeling/gencontract/conformance.go`, `modeling/gencontract/conformance_test.go`,
`modeling/gencontract/cmd/refgen/main.go`,
`modeling/gencontract/cmd/refgen/stdout_test.go`,
`modeling/tui/cmd/tuic/main.go`, `modeling/tui/cmd/tuic/stdout_test.go`,
`modeling/jsonvalidator/cmd/jsonvalidator/main.go`,
`modeling/jsonvalidator/cmd/jsonvalidator/stdout_test.go`,
`modeling/lifecycle/cmd/lifecyclegen/main.go`,
`modeling/lifecycle/cmd/lifecyclegen/stdout_test.go`,
`modeling/diagram/cmd/diagramgen/main.go`,
`modeling/diagram/cmd/diagramgen/stdout_test.go`, this task and new immutable
Project Memory files. Check stdout byte counts and errors, return exit 1 with a
fixed diagnostic, and ignore SIGPIPE only in each process main. Add failed,
short and full-count/error writer tests plus real preclosed-pipe conformance and
a reference fault proving the harness detects false success. Keep normal bytes,
refusal/drift checks and generator versions unchanged. Clarify only the necessary
contract exception for bytes already delivered before an output failure.

Checks: prove package and real-process regressions fail before the repair; all
affected module vet/race suites, JSON checks with pinned Node 24 and lifecycle
scratch under canonical TMPDIR; all five generators' conformance; repository,
format, whitespace and committed Project Memory validation. Finish edits before
full tests. Commit, push and open a PR for independent coordinator review and
hosted checks; no merge, output regeneration or Foundry entrypoint edits.
