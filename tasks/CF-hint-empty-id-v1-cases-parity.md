---
id: CF-hint-empty-id-v1
status: completed
owner: Codex hint validation contributor
repositories: [spik3r/clanker-foundry]
---

# Reject explicit empty key hint IDs

## Implementation claim

Owner: Codex hint validation contributor, delegated by the coordinating agent.
The owner requested continued subagent work and reviewed PR merges. This claim
fixes the existing P2 review finding on PR #139 at
`d4071d3f8d28d54872e40ce63c09884681371767`. The original hint ID task is complete;
this follow-up changes no existing immutable Project Memory evidence.

Scope: `modeling/tui/cases.go`, `modeling/tui/cases_test.go`,
`modeling/tui/authoring/cases.schema.test.cjs`, this task record and new immutable
Project Memory files. Preserve the distinction between absent and explicit
empty hint IDs during JSON decoding. Reject empty IDs with key, action, or both;
keep absent-ID legacy key/action pairs accepted. Add Go decoder regressions and
schema parity cases. Keep public types, generated code and schema unchanged.

Checks: prove the regression fails before the fix; focused Go decoder tests;
full TUI race tests and vet; Node authoring tests; generator contract and
foundrygen race tests; repository validation, whitespace checks and committed
Project Memory validation. Push a fast-forward update to PR #139 only after
rechecking its live head. The coordinating agent reviews and merges.

## Integration result

The coordinating agent reviewed `b5e55cb8af631518b6e297611fce51b4c9a26234`;
all ten hosted workflows passed on that head. [PR #139](https://github.com/spik3r/clanker-foundry/pull/139)
merged on 14 November 2026 as `c87356af78b86ee3c944cbf91b81604c18e21782`.
This bounded implementation claim is complete and reserves no further work.
