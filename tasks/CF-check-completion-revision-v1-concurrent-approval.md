---
id: CF-check-completion-revision-v1
status: review
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

## Observed implementation and checks

Published source commit `e951dda5567561e7cb9596df25cee9d545778ed8` was fetched and
its tree compared with the prepared local source before registration and the
genuine start. Git CLI publication lacked credentials; the GitHub connector
published the same tree. The local preparation commit is not published ancestry.

The real-store regression confirmed that owner approval, approval revocation,
and another check's evidence can advance the revision during the callback and
cause a valid result to become UNKNOWN with REVISION_CONFLICT. The final eight
regressions on the old runner pass the unchanged-work positive control and fail
seven expectations: three result-preservation cases and four precise refusal
outcomes. These tests are implementation-authored, not an independent oracle.
Initial fixture corrections fixed an invalid reason identifier and compared
persisted evidence without object-prototype differences; final test bytes were
rerun against the old source before reporting these results.

The runner now reloads execution after the callback and checks the original
attempt, fence, worker, subject and specification before using the latest revision.
The evidence remains bound to the checked input. A concurrent result for the same
check is not replaced: existing duplicate-evidence validation refuses it.
Cancellation and replacement retain UNKNOWN; deadlines remain strict. A later
race between the refreshed read and transactional command still fails closed
through the store's revision check. No callback is automatically retried.

Observed on Node v24.22.1 and Git 2.52.0:

- All eight new regressions and all 31 focused runner/check tests passed.
- Full runtime suite: 315 passed, zero failures, skips or cancellations.
- Project Memory and workflow suites: 143 passed, no skips or cancellations.
- Repository validation passed; optional skills-ref is unavailable.
- Pack validator passed 92 path checks and 47 model-map cases.
- Project setup, model resolver, nine wiring-preflight cases and sixteen
  installer-ordering cases passed.
- Full installer acceptance stopped because GNU Stow is absent. Preflight and
  recording-seam passes do not establish actual GNU Stow installation behavior.
- Syntax and whitespace checks passed. Committed lifecycle validation is run
  after recording completion and committing the exact tested content.

## Security review scope

The threat considered is a late callback attaching old evidence to an execution
whose authority or subject changed, or losing a valid result merely because a
separate owner decision arrived. Tests exercise real durable commands and replay,
with invented callback data and private temporary databases. Current state,
attempt, fence, worker, subject, specification, cleanup and transactional revision
guards remain explicit. Existing validator registration, independent identity,
duplicate evidence, timestamp and deadline checks still execute in the evaluator.

No authentication surface, role permissions, dependency, secret, schema, SQL,
network endpoint, upload, logging or deployment configuration changes. Web,
password, session-cookie, TLS, personal-data and backup controls are outside this
local callback correction. No private data or credentials are added to the diff.
Independent review and exact-head hosted CI remain integration gates owned by the
lead; local completion does not authorize activation or deployment.
