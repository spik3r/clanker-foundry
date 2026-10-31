---
id: CF-orchestration-astra-review
status: review
owner: Astra PR review
repositories: [spik3r/clanker-foundry]
---

# Repair cancellation and never-started container cleanup

## Implementation claim

Astra owns this bounded review follow-up for PR 9 on
`codex/deterministic-orchestration-plan`, reviewed head `61f7b27757b8` and
main `7a03b095eff5`. The current PR has no assignee or requested reviewer;
its prior implementation evidence frontiers are complete and no active
implementation claim overlaps these files. The repository owner authorized
review, fixing unowned findings, publication and merge after verification.
This local claim does not transfer any runtime namespace or existing task.

Exact scope:
- `runtime/src/adapters.mjs`
- `runtime/test/adapters.test.mjs`
- `runtime/src/isolation.mjs`
- `runtime/test/isolation.test.mjs`
- `tasks/CF-orchestration-astra-review-fixes.md`
- New immutable Project Memory source and event files for this task.

Reject cancellation immediately before an injected text callback is invoked,
including the queued-microtask boundary. Preserve honest cleanup while an
already-running callback has not settled. Recognize an inspected, owned Docker
container that never started as nonrunning and removable; do not treat an
uninspected or running container as stopped. Add deterministic regressions for
both failures and independently recheck the final diff.

No provider calls, namespace activation, operating database changes, deployment,
credential changes, claim migration, history rewriting or branch-protection
changes are included. The lead owns staging, lifecycle evidence and commits.

## Validation and next action

Before publication: run targeted regressions, the complete Node 24 runtime and
Project Memory suites, the pack validator, whitespace checks and the committed
full-diff evidence gate. Run installer and actual isolation checks where the host
supports them; otherwise distinguish host limits from final-head CI results.
Recheck the remote head before pushing and use normal merge controls only after
reviewing the exact final revision.

## Result and observed checks

Added an abort guard inside the generic text adapter's queued provider callback.
The regression now cancels in that exact window with zero provider invocations;
the preexisting active-call test waits for a real invocation before revocation.
Inspected owned Docker containers in `created` state are now safely removable
after start failure and during cleanup, but can never count as a successful plan.
Running or ownership-mismatched containers remain unconfirmed.

Node v24.22.1 passed all 311 runtime and Project Memory/workflow tests, with zero
skips. Pack validation, whitespace checks and dependency audit passed (zero known
vulnerabilities). The CLI demo completed all four fake stages and replay. The
optional skills-ref check is unavailable. Local installer testing stopped at the
missing GNU Stow prerequisite; actual Docker and live provider checks were not
rerun on this host. Original-head CI passed installer and actual Docker isolation;
new exact-head CI is checked separately after publication.

Independent review passed 17 adapter/isolation tests, exercised 32 queued
cancellations followed by all 16 output-capacity slots, and checked that a client
exit code of zero cannot accept a never-started container. New regressions fail
against the original source. No remaining material finding was reproduced.

Implementation editing is complete and this claim reserves no further changes.
Next: record genuine completion, commit and validate the full committed diff,
publish only if the remote head still matches the reviewed parent, then verify
exact-head checks and normally merge PR 9. No operating namespace, existing task
claim, live provider, deployment or branch protection is changed.
