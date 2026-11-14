---
id: CF-foundry-identity-strict-v1
status: in-progress
owner: Codex identity parser contributor
repositories: [spik3r/clanker-foundry]
---

# Enforce the generator identity line contract

## Implementation claim

Owner: Codex identity parser contributor, delegated by the coordinating session.
The owner authorized continued bounded tasks, subagent work, publication and
reviewed merges. This contributor will publish a PR for coordinator review.
Base: main at `365caf3`, after PR #140. Open PR #139 changes TUI hint IDs and
does not overlap this claim.

The generator contract requires exactly four string members, in a fixed order,
with no extra whitespace and one final newline. `foundrygen` currently decodes
only the first JSON value, accepts duplicate members, and accepts a missing or
null profile. Enforce the existing identity wire contract before accepting a
successful generator response. Add regression cases for trailing data, duplicate
members, missing/null/wrong-type members, whitespace and member order, while
keeping valid responses accepted.

Scope: `modeling/foundrygen/exec.go`, `modeling/foundrygen/run_test.go`,
`modeling/foundrygen/README.md`, this task and new immutable Project Memory
source/events. No contract changes, generator changes, profile semantics,
consumer changes or runtime work.

Checks: observe new regressions fail before the fix; gofmt, go vet and full
race-enabled tests in modeling/foundrygen; repository validator; whitespace
checks and committed Project Memory validation. The coordinator owns final
review and merge; this contributor must not merge the PR.
