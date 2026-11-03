---
id: CF-pilot-custom-proposal-guide-v1
status: in-progress
owner: Lifecycle pilot testing guide
repositories: [spik3r/clanker-foundry]
---

# Show how to check a user-supplied lifecycle proposal

## Implementation claim

The coordinating owner requested a bounded extension to the merged testing guide
for a user-supplied model, scenario suite and acceptance manifest. Current main is
`09f1b5a0b46b1b641f738baea6670b1b33f55bdb`. The original guide's documented
owner is this same guide lane; its Project Memory frontier is complete, while its
canonical task remains in-progress. A fresh current-main task and relevant open-PR
check found no competing guide change. This claim is a local proposal, not a lock
or a claim about another contributor's authority.

Exact scope: `docs/model-driven-pilot-testing.md`, this task record and new
immutable Project Memory source and event files for this task. Preserve existing
code, generated artifacts, fixtures, unrelated tasks and evidence. The held
process-view artifact and interactive modelwalk work are out of scope.

Document a byte-preserving path for an operator-selected proposal directory,
with non-executing preflight as the default. Describe how to interpret variable
results and failure exits without assuming the fixed example's counts or treating
matching results as approval. Separate any candidate replay into a later explicit,
authorized operator action. Keep the trusted-host, independent-expectation,
privacy, human-review and no-deployment boundaries clear. Link to the existing
proposal replay contract instead of duplicating or changing its implementation.

## Validation and handoff

Run the new documented snippets on the two independently frozen invented proposal
sets in external directories, and exercise stale and mismatched inputs without
repairing expectations. Record which commands actually run and distinguish
preflight from execution. Obtain independent read-only review, run repository and
full committed-diff Project Memory validation and report the exact result. Normal
draft PR publication is authorized; merge is a separate user-authorized step.
No formal approval, deployment or worker activation is implied.
