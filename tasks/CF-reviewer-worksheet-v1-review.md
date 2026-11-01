---
id: CF-reviewer-worksheet-v1
status: in-progress
owner: Codex reviewer-workflow contributor
repositories: [spik3r/clanker-foundry]
---

# Human review worksheet for intake and trace

## Implementation claim

On `codex/reviewer-evidence-linking`, based on main `9d06274`, add a reusable
human-completed worksheet and a worked invented example connecting an intake
decision and unresolved question to explicit requirements and acceptance cases.
The worksheet records exact source-byte hashes separately from canonical model
and normalized requirements identities, declared versus observed bindings,
trace gaps, unrun checks and explicit unresolved questions. It does not turn
correspondence into evidence of satisfaction or approval. Use the merged intake
and traceability APIs as the only validators and make no new renderer, bundle,
intake/model bridge, comparison, execution runner or identity rebinding.

Exact scope: new `modeling/lifecycle/reviewer/README.md`, `worksheet.md`,
`example-intake.json`, `example.md`, `worksheet_test.go`, this record and its
own immutable Project Memory evidence. Commit independently specified example
and literal expectations before writing the test. Test the stated identities,
existing intake statuses and linked trace cases including pending and unlinked
gaps; demonstrate that source-byte edits and changed links are detected without
rewriting examples. Document distinct identities and checks not performed.

Inspected current main, open PRs 5, 20, 22 and 23, visible branch and task
claims, and held reviewpack scope. Excludes reviewpack/modelintakecheck, PR 20,
the guide, Astra intake-change comparison, and other current worker lanes. This
local claim is not a shared lock; recheck before publishing. No deployment,
activation, formal approval or browser visual inspection.
