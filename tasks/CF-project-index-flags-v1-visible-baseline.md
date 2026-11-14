---
id: CF-project-index-flags-v1
status: completed
owner: Codex shared-baseline maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Refuse hidden tracked baseline files during project setup

## Implementation claim

Owner: Codex shared-baseline maintenance contributor, delegated by the coordinator
under the owner's continued maintenance and reviewed-merge request. Base:
origin/main at `b3faf9c`. The coordinator assigned these paths exclusively; the
separate Foundry pin follow-up owns no files in this scope.

In isolated local submodule fixtures, assume-unchanged and skip-worktree flags
hid edits to global-agents/AGENTS.md from the baseline status check. Setup and
--check both returned zero while importing bytes that differed from the staged
revision. The documented setup requires a clean pinned baseline.

Before trusting Git status or creating wiring, refuse baseline index entries
marked assume-unchanged or skip-worktree, including clean flagged files. Do not
clear flags, refresh the index or write probe files. Read Git output with NUL
record boundaries and propagate Git failures. Keep ordinary full checkouts and
project-owned overlays working. State the conservative full-checkout requirement
and its effect on sparse checkouts; do not claim protection from a hostile host.

Scope: `scripts/project-setup.sh`, `tests/test-project-setup.sh`,
`docs/project-shared-base.md`, this task and new immutable Project Memory files.
No global installation, real-home changes, Foundry entrypoint, generator,
consumer or runtime changes. Fixtures use owned scratch repositories and public
local source only.

Checks: prove new cases fail before the repair; exercise both flags on clean and
changed files in setup and --check, preserve bytes/index flags on refusal, test
NUL-safe paths and failed Git queries, and retain ordinary setup/overlay controls.
Run Bash syntax, full project-setup and installer regressions, repository
validation, whitespace and committed Project Memory validation. Commit, push and
open a PR for independent review; the coordinator owns merge.

## Integration result

The coordinating agent confirmed independent review of
`11cd8d663718c0f88065ad17bee0109be941737f`; all ten hosted workflows passed on that exact head.
[PR #156](https://github.com/spik3r/clanker-foundry/pull/156) merged on 14 November 2026 as
`fa8ca14616ad9fa812edadf9cdbcb294782f0a45`. Git confirms the reviewed head is the merge's second parent.
This bounded implementation claim is complete and reserves no further work.
