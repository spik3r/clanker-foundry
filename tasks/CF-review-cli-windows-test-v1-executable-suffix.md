---
id: CF-review-cli-windows-test-v1
status: review
owner: Foundry CLI test portability contributor
repositories: [spik3r/clanker-foundry]
---

# Use an executable suffix in review CLI process tests

## Implementation claim

Base: current main `82b62e61ac6038b4b1f4bfd39e19153d2fc2b74b`.
Branch: `codex/review-cli-windows-test-v1`.

PR [#70](https://github.com/spik3r/clanker-foundry/pull/70) merged the closed-pipe
repair and process acceptance tests before a reviewed test portability correction
could be published. Current main contains that original test unchanged. The
reviewed correction in `f96ebf14dde7518a152559b3d6e814d08f1d8256` directly follows
the published original head and adds an `.exe` suffix to each temporary binary.
Apply only that test hunk on current main, with a new task and genuine lifecycle
events. Do not import or rewrite the original task's evidence.

The coordinator assigned this follow-up after checking the merged PR. The only
open PR at preparation was #71, for a separate custom-change rehearsal. Existing
claims and source paths were inspected; this claim covers only the portability
correction and its evidence.

Exact scope: `modeling/lifecycle/clioutputacceptance/output_test.go`, this task
record, and new immutable Project Memory source/event files for this task.
Keep test assertions, subprocess bounds, process inputs and production code
unchanged. Publish a small draft follow-up; do not merge or force-push.

## Acceptance and validation

Commit this claim, register its pinned source, and record a genuine start before
editing the test. The final test must match the independently reviewed test blob
from `f96ebf14dde7518a152559b3d6e814d08f1d8256` exactly.

Run Go 1.27.3 offline race tests for the process acceptance package and the three
affected command packages, focused vet, formatting, repository validation,
whitespace and committed Project Memory validation. Cross-compile the acceptance
test and the three commands for Windows. Keep caches and binaries outside the
checkout. Native Windows execution is unavailable; compilation does not establish
Windows runtime behavior. State that limit in the PR and report actual results.

## Observed validation

The corrected test matches the reviewed test blob from `f96ebf14` byte for byte.
Go 1.27.3 offline race tests passed for the acceptance package and all three
command packages on macOS, including all nine compiled-process cases. Focused
vet, formatting, repository validation and whitespace checks passed. The
acceptance test and all three commands cross-compiled for Windows amd64 as PE32+
executables. Native Windows execution remains unverified. The optional
`skills-ref` validator is unavailable locally. No production source was changed.
