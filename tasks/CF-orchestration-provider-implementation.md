---
id: CF-orchestration-provider
status: review
owner: Foundry managed execution
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Add and verify live stateless provider adapters

## Implementation claim

The user requested continuing orchestration and finishing the outstanding provider,
writable-host, evidence-bridge and namespace-cutover work. Owner/integration lead:
Foundry managed execution, branch `codex/deterministic-orchestration-plan`, base
`aa89562`. The checkout was clean and the connected open-PR search returned no
open PRs. This is a bounded local claim, not a published shared reservation.

Exact implementation scope:
- `runtime/src/providers.mjs`
- `runtime/test/providers.test.mjs`
- `runtime/test/provider-live.mjs`
- This task record and new immutable Project Memory source/event files.

Builders use disjoint paths; the lead owns lifecycle records, integration and
local commits. Preserve existing tasks, active claims and all immutable evidence.
Live verification uses only invented public-safe payloads and owned temporary
workspaces. Credentials remain in the existing provider login or host environment;
never copy them into repository files, logs or model inputs. No publication,
merge, external deployment or silent transfer of existing CF claims is included.

## Acceptance and dependencies

Implement the corresponding remaining architecture boundary with strict validation,
revision/identity checks, deterministic failure/recovery and regression tests.
Reuse the current single-host runtime and offline Project Memory helper.
Managed writes require verified isolation and exact authorized scope; provider
and coordinator credentials stay outside the writable host. Evidence starts
must precede edits; completion requires actual staged fingerprints and checks.
Namespace activation needs a recorded reviewed source and drained prior authority;
new namespaces preserve historical CF records. Missing external requirements fail
closed and are reported rather than counted as successful live validation.

## Validation and next action

Implemented the assigned boundary and reviewed the integrated change. Node
v24.22.2 passed 154 runtime tests and 143 existing lifecycle/workflow tests, with
no skips. Pack, installer and whitespace checks passed. The signed-in Codex text
smoke and actual Docker probe/write/promotion smoke passed using invented inputs.
The direct Responses transport has no live verification because no API key is
configured. Fresh-context review found no remaining material defect after fixes.

Next: owner review of the local branch. The new CF-AUTO namespace has an inactive
configuration template; operating activation requires its actual shared database
identity and a separately pinned owner decision. No existing CF claim transfers,
publication, merge or deployment occurred. See the rollout guide for the operating
steps and limits. Completion records capture staged fingerprints; the committed
full-diff evidence gate is run after the local commit.
