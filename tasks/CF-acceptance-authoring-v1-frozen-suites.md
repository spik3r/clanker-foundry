---
id: CF-acceptance-authoring-v1
status: in-progress
owner: Codex acceptance-authoring contributor
repositories: [spik3r/clanker-foundry]
---

# Frozen independent scenario authoring

## Implementation claim

On `codex/acceptance-authoring-v1` from main, add a small offline authoring
manifest and checker for literal scenario suites. Record a declared author,
purpose and per-case intended behavior, pin the exact suite source hash and the
semantic model digest, compare both against independently supplied current bytes,
and explicitly detect changed suite/model bindings. Never manufacture or refresh
expected results from observed output. Optional replay delegates exclusively to
the existing scenario runner and only after all identities and the compiled runner
binding match. Unknown or stale artifacts cannot yield a validated or replayed
claim. Author labels and hashes do not establish authorship, independence,
authentication, execution provenance, review, approval or authority.

Exact scope: `modeling/lifecycle/acceptanceauthor/`,
`modeling/lifecycle/cmd/modelacceptance/`, this task and its own immutable Project
Memory records only. No changes to PR15 traceability, intake or other worker's
handoff, scenario runner, generated code, held modelpatch, workflowacceptance,
reviewpack, qualification or exporter. Held workflowacceptance is an integration
test over frozen candidate artifacts; it is not an authoring manifest or stale
binding check. Inspect live claims again before publishing. A local record is not
a shared lock.

## Acceptance

After genuine start, freeze independently specified literal behavior and expected
outcomes before implementing the checker. Demonstrate a successful transition and
an expected refusal, an intentionally wrong literal expectation, altered suite
bytes, altered model meaning, invalid input, no stale replay, deterministic output,
CLI error handling and unchanged source data. The existing scenario runner owns
execution and result comparisons. State precisely when stale or incompatible input
has not been fully validated. Run pinned Go race/vet/drift and repository checks,
record completion, and prepare one draft for independent review if allowed. No
merge, deployment, worker activation or retry of a denied action.
