---
id: CF-foundry-identity-strict-v1
status: completed
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

## Integration result

The coordinating agent reviewed `772496c2553a388d4c3a5edc432942d099314521`;
all ten hosted workflows passed on that head. [PR #141](https://github.com/spik3r/clanker-foundry/pull/141)
merged on 14 November 2026 as `6529cd27ed3d4f532d6e24c097e897bf7ba1e42c`.
This bounded implementation claim is complete and reserves no further work.
