---
id: CF-check-completion-revision-v1
status: in-progress
owner: Foundry check completion contributor
repositories: [spik3r/clanker-foundry]
---

# Preserve completed checks across concurrent owner approvals

## Implementation claim

Owner: Foundry check completion contributor. Branch
`fix/runtime-recovery-20261111`, reviewed main base
`cb507d19068161dfb114d9cc983b8363a6ba64ee`. The integration lead authorized one
bounded runtime correctness fix after reviewing the candidate and its scope.
The isolated checkout began clean. The live open-PR check found only unrelated
PR101; existing completed runtime work reserves no competing implementation.
This local claim is not a shared lock or a runtime namespace decision.

Exact scope:

- `runtime/src/runner.mjs`
- New `runtime/test/check-completion-revision.test.mjs`
- This task and its new immutable Project Memory source and event files.

The current check runner captures an execution revision before awaiting its
callback and uses that revision when recording the result. Reproduce whether a
concurrent valid owner approval discards an otherwise current passing result as
REVISION_CONFLICT. Implement only if the real-store reproduction confirms it.

Refresh execution after the callback, require the same validating attempt,
worker, fence, subject and specification with no cleanup, and then use its current
revision for the evidence command. Keep original result identity, strict deadline
enforcement, duplicate-evidence refusal and unknown-outcome handling. Do not
rerun callbacks automatically. Cover concurrent approval, another validation,
cancellation/replacement, and expiry using invented callbacks and private
throwaway stores.

Do not change evaluator semantics, store APIs, schema, adapters, providers, CLI,
TUI/Flexgo, namespace configuration, activation, existing claims or held general
recovery work. No live provider, container or production deployment is included.
The integration lead owns independent review and merge. This lane may publish its
branch and draft PR after local validation, without marking PR101 ready.

## Validation plan

Commit this prospective source, register it and record a genuine start before
implementation edits. Observe old-code regression failures before changing the
runner. Run focused tests and the full runtime suite on Node 24 or later,
repository validation, installer regressions, whitespace checks and committed
Project Memory validation against actual main. Preserve old evidence and record
only observed results. Local checks do not establish hosted CI or integration.
