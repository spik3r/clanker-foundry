---
id: CF-change-entrypoint-v1
status: review
owner: Foundry change entrypoint contributor
repositories: [spik3r/clanker-foundry]
---

# Prepare and rehearse reviewed changes through the operator command

## Implementation claim

Local-only lane from reviewed main ade8fc8e23755638e162aa0dd897edf313e22d74.
Live open-PR search was empty; the custom-change-rehearsal, provider-change-closeout
and provider-reply-boundary branches were verified ancestors of main. The separate
revision-check v3 contributor owns revisioncheck and revision_versioned.go. This
claim is not a lock or authority to publish, merge, deploy or activate workers.

Exact maintained paths, plus this task's new immutable Project Memory events:

- tasks/CF-change-entrypoint-v1-file-preserving-rehearsal.md
- modeling/lifecycle/internal/operatorcli/change.go
- modeling/lifecycle/internal/operatorcli/change_prepare.go
- modeling/lifecycle/internal/operatorcli/change_contract_test.go
- modeling/lifecycle/internal/operatorcli/change_dispatch_test.go
- modeling/lifecycle/internal/operatorcli/dispatch.go
- modeling/lifecycle/internal/operatorcli/contract_test.go
- modeling/lifecycle/cmd/foundry/main.go
- modeling/lifecycle/cmd/modelchangerehearsal/main.go
- docs/model-driven-change-entrypoint.md

Do not edit native model, patch, rehearsal or proposal packages; source policies;
existing frozen fixtures; runtime; CI; other claims; or the held visual artifact.
The only old contract-test edit is replacing the old complete-help comparison
with preserved original route lines plus separately asserted updated help text.
No behavioral expectation or native input/output fixture changes are allowed.

## Intended result

Add explicit `foundry change prepare|check|execute` routes. Prepare takes exactly
--patch-request FILE, --baseline-proposal FILE, --candidate-suite FILE,
--candidate-manifest FILE and --expected-candidate-digest HEX, each once. The
reviewed lowercase 64-digit digest is supplied, never derived automatically.
Read four regular UTF-8 files with existing trusted-host admission, preserving
all bytes and limits (3 MiB, 16 MiB, 1 MiB, 64 KiB). Encode the existing strict
native request, bounded at 32 MiB. Preparation assembles only, never executes,
refreshes expectations, validates a semantic model or implies human review.
Exit 0 means delivered request; exit 2 is input/argument/limit refusal with no
stdout, and exit 3 means failed or short delivery. Diagnostics omit private paths
and payloads. An invalid/stale native artifact can be assembled for explicit
native checking; preparation is not preflight approval.

Check and execute use one extracted native CLI runner also used by the original
standalone adapter. Preserve native report bytes, diagnostics, exit semantics,
partial receipts and context cancellation. The source root remains the trusted
current working directory, never request data. Only explicit execute invokes
native execution; help, prepare, check and invalid forms receive no execution
capability. Top-level help must describe both explicit execution routes honestly.
Only the change execute route adds cooperative cancellation to the foundry main;
other native signal behavior stays unchanged. Keep SIGPIPE output-failure exits.

## Acceptance and handoff

Register this reviewed pinned source and record a real start before tests or
production edits. A separate contributor freezes acceptance before production.
Test exact file-byte preservation including CRLF and Unicode, arbitrary flag
order, malformed flags/digest, static symlink/nonregular/invalid UTF-8/oversized
inputs, encoded wrapper bounds, short output and input nonmutation. Use existing
independent native rehearsal vectors unchanged for current/stale/mismatch and
boundary classification. Compare the actual compiled standalone and foundry
check outputs and closed-pipe exits, and explicitly exercise invented local
matched/mismatched execution. No unrelated system tests are in scope.

Run focused and full Go 1.27.3 offline race tests, vet, modelc drift, repository
validator, installer regressions, whitespace and committed Project Memory checks.
Keep source frozen during full tests. Obtain separate exact-source review. Report
what passed, failed or was not run; local matches do not authorize integration.

## Local implementation result

The command now prepares the existing native request from exact file bytes and
an explicitly supplied candidate digest, then exposes separate check and execute
routes. The standalone adapter uses the same extracted native CLI body without
changing its behavior. Help names both execution routes. Only exact change
execute adds cooperative cancellation to the operator binary. Native source
packages, source inventories and frozen fixtures remain unchanged.

A separate contributor froze acceptance in 71b8e6d before production changes,
after independently checking the existing literal native outcomes. The tests
cover exact file preservation, all file and wrapper bounds, refusal and output
failures, execution-capability separation, native report parity, actual compiled
matched/mismatched execution and closed stdout. Existing fixture SHA256 values
remain 8b70e7b9658d124f3ce3f5987b7d20e73c305964801aa6b3989e9dc59cd4edc3
for operator observations and
d9b30467427539c02d3355724b0d9ac13c47f46360996b4dfcc660d1a96403e3 for
change rehearsal observations.

Go 1.27.3 offline full race tests passed: 75 tested packages, with two command
packages compiling without local test files. Focused operator and standalone
race tests, vet, gofmt, model validation, generated-artifact and trace drift,
repository validator and whitespace checks passed. All three guide command
blocks ran on the frozen invented example: preparation and check returned 0,
execution returned 0 with both sides matched, candidateStarted true and cleanup
removed. The receipt bound the unchanged prepared request hash.

A separate source reviewer found no blocking issue and independently passed the
non-executing preparation, routing, help, delivery and precancelled-context race
tests. Actual OS SIGINT/SIGTERM delivery was not separately exercised; the exact
signal wiring was reviewed. The installer regression stopped because GNU Stow
is absent; no installation was attempted. Optional skills-ref is also absent.
These limits are not a full installer or hosted-CI pass.

This remains a local review handoff. No push, PR, formal approval, merge,
deployment, worker activation or unrelated-system test was performed. Request
assembly and matching examples confer no human approval or integration authority.
