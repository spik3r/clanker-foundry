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

## Observed local result

Extended the existing guide with a standalone non-executing check of an
operator-selected directory and a separate, explicitly authorized replay block.
The result readers display variable outcomes and counts without asserting the
fixed example's results. They do not authenticate a receipt or replace review.
The original guide's setup now permits the reader to skip the executing fixed
examples when choosing the check-only path.

The literal setup and new blocks were run with the two untouched, separately
frozen invented proposal sets copied into external private directories. Check
returned `checked` and `not-run` for each; after separate local authorization,
actual candidate replay matched 4 cases / 6 steps for A and 4 / 7 for B. A stale
suite returned exit 1 with no candidate start and the documented path skipped
replay. A disposable deliberately incorrect expectation with a matching test
manifest passed preflight but produced a real mismatch at the first step; no
frozen input was modified or expectation repaired. An invalid model returned
exit 2 and no receipt; that test also used a directory containing spaces.

Separate shell controls verified a missing or unsuccessful prior check, an
existing output, and a dangling output symlink stop before candidate invocation;
an empty JSON receipt is not reported as a success. A read-only reviewer ran the
literal final blocks independently on fixture B, checked all seven Bash blocks
with `bash -n`, reviewed the exact guide bytes and found no blocker. This is not
formal human approval or proof of arbitrary-model correctness. Repository
validation and relative links passed; optional `skills-ref` was unavailable.
No production integration, visual review, deployment or worker activation occurred.
