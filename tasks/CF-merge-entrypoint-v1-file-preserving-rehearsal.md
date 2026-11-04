---
id: CF-merge-entrypoint-v1
status: in-progress
owner: Foundry merge entrypoint contributor
repositories: [spik3r/clanker-foundry]
---

# Prepare and rehearse concurrent designs through the operator command

## Implementation claim

Local-only lane from main 3afc6b0da037f7a933ca1a1c12a7e7d1c447e3d9,
tree 44c6c9c4286e699f1fceade86fddc0c66235f494, after custom merge rehearsal
PR 76 merged. The public check at 2026-11-04 07:42 UTC found zero open PRs,
77 branches and no merge-entrypoint claim in main tasks or branch names. The
existing change entrypoint and custom merge rehearsal branches are merged.
An external project worker's current work is inaccessible and unknown, not idle.
Recheck public ownership before implementation and stop on observed overlap.
This local claim is not a published reservation or authority to publish, merge,
deploy, activate workers or inspect held artifacts. Separate scenario trace
work and unpublished artifacts remain outside this claim.

Exact maintained paths, plus this task's new immutable Project Memory events:

- tasks/CF-merge-entrypoint-v1-file-preserving-rehearsal.md
- modeling/lifecycle/internal/operatorcli/merge.go
- modeling/lifecycle/internal/operatorcli/merge_prepare.go
- modeling/lifecycle/internal/operatorcli/merge_contract_test.go
- modeling/lifecycle/internal/operatorcli/merge_dispatch_test.go
- modeling/lifecycle/internal/operatorcli/dispatch.go
- modeling/lifecycle/internal/operatorcli/change_dispatch_test.go
- modeling/lifecycle/cmd/foundry/main.go
- modeling/lifecycle/cmd/modelmergerehearsal/main.go
- docs/model-driven-merge-entrypoint.md

Do not edit native merge, custom rehearsal, proposal, replay or model packages;
source policies, inventories, frozen fixtures, runtime, CI, other claims or held
visual artifacts. The one existing change dispatch test may replace its complete
help comparison with preserved old route lines; the new test freezes updated
complete help. No native behavior or expectation may change.

## Intended result

Add explicit foundry merge prepare, check and execute routes. Prepare requires
each of these flags exactly once, in any order, with one separate value:

- --merge-request FILE
- --base-suite FILE and --base-manifest FILE
- --left-suite FILE and --left-manifest FILE
- --right-suite FILE and --right-manifest FILE
- --candidate-suite FILE and --candidate-manifest FILE
- --expected-candidate-digest HEX

The digest is independently supplied lowercase 64-digit hexadecimal. Never derive
or replace it from the model, report, execution or file content. Read nine regular
UTF-8 files using the existing trusted single-writer host admission helper. Preserve
every input byte. The merge request limit is 4 MiB, each suite limit 1 MiB and each
manifest limit 64 KiB, inclusive. Encode the existing native request at no more
than 32 MiB including its newline. Preserve all four stage identities and nested
suite/manifest pairs. Empty and semantically invalid text can assemble and is left
to the explicit native check. Preparation is neither semantic validation, a native
merge, candidate execution, expectation authoring nor human review.

Refuse missing, unknown, repeated, positional and joined-form flags, malformed
digests, missing files, symlink leaves, nonregular leaves, invalid UTF-8 and oversized
files. Preserve the existing helper's admitted parent aliases and stated limits:
it protects against static leaf substitution, not hostile concurrent writers.
Do not read stdin for preparation. Bad arguments refuse before reading files.
Preparation exit 0 means request delivery, 2 means refusal with no stdout and
3 means failed or short output. Diagnostics contain no caller paths or payloads.

Extract the existing standalone native CLI body into operatorcli.RunMerge and use
it from both binaries. Check and execute remain stdin-only and preserve exact native
reports, diagnostics, exit meanings, source-root configuration, partial receipts,
cancellation and delivery failures. Trusted sourceRoot remains the current working
directory, never request data. Only exact merge execute receives execution capability
and cooperative cancellation in foundry main. Help, prepare, check and invalid
forms cannot invoke execution. Keep existing proposal/change signal behavior and
SIGPIPE handling. Help must name all three explicit execution routes honestly.

All native outcomes remain unchanged: conflicted and invalid-combination have no
candidate and no execution; stale preflights and identity refusals cannot execute;
matched execution requires all four stages; a mismatch or failure suppresses later
stages while retaining actual receipts. These are inherited semantics, not new
claims of approval, authentication, rollback or readiness.

## Acceptance and handoff

Obtain independent review of this claim before registering its pinned local source.
Record a genuine v2 start before fixture or production edits. A separate contributor
must author and commit frozen acceptance before production begins. Use the unchanged
independent native ArchivePacket vectors and literal outcomes, not observed output,
for behavior. Freeze exact file preservation, CRLF and Unicode, all nine flag/file
positions, four distinct expectation pairs, digest noninference, arbitrary order,
inclusive file and expanded wrapper bounds, malformed flags, unsafe leaves, short
output, input nonmutation and capability separation. Compare compiled foundry and
standalone check/refusal/execute outputs. Cover conflicts, invalid combinations,
each stale stage, identity refusal, real matched and all four real mismatch paths,
cancelled context, failed delivery and closed stdout without changing native tests.
State any finite coverage limits rather than claiming universal safety.

Document the standalone/native relationship, complete file-based command, independent
expectations and candidate identity, check versus execute distinction, all outcomes,
receipt interpretation, trusted-host scope and limits. Do not add unrelated routes.

Use installed pinned Go 1.27.3 offline and put its binary directory on PATH for
native subprocesses. Run focused and full relevant race tests, vet, modelc drift,
gofmt, repository validation, installer checks, whitespace and committed Project
Memory checks. Record actual passes, failures and unavailable checks. Keep source
frozen during full qualification and obtain independent exact-source review.
No other-system testing, downloads, publication, push, PR, merge or deployment.
