---
id: CF-merge-entrypoint-v1
status: review
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

## Local implementation result

Implementation source is 0ee9f24a2c0de0b020cfc8926096e829926cf35b. A separate
exact-source reviewer found no defects, verified the former standalone runner
body is unchanged except its name, and confirmed unchanged native sources,
policies and fixtures. The reviewer also passed 40 randomized all-nine-file
preparations with shuffled flags, Unicode/control bytes and input nonmutation,
and refused FIFO leaves in all nine positions. A Unix-socket probe was blocked
at socket creation by host permissions; it was not retried and is not a pass.

Independent acceptance froze before production in
e608a52bb8b5e3f24ac37dd785a4c3fa3787adc7. Its two files remain unchanged:

- merge_contract_test.go SHA256:
  7a7d7c604f1e7cdf453e4aa1a2e31155d44274be4429e6efc64be982b56f56d7.
- merge_dispatch_test.go SHA256:
  fa392293ad875023e11092ccfe190b8b5e449b57bdca4a57382d2856af7e6a92.

An initial exact-wire test rejected HTML escaping by the preexisting generic
preparation helper. Only the new merge preparer changed to native non-HTML-escaped
encoding; acceptance literals and existing helpers stayed unchanged.

Final offline Go 1.27.3 validation passed focused race tests, all 77 tested module
packages under go test -race -count=1 ./... (two packages have no tests), go vet,
modelc validate/check with no generated drift, gofmt and whitespace checks. The
documented prepare/check/execute shell sequence also passed with the unchanged
independent ArchivePacket inputs and separately frozen candidate digest.

The first full race run failed only the existing clean-copy reproducibility
test because the host exposed an empty ancestor .git directory. The isolated
test scratch parent was initialized as a normal local Git repository; focused
reproducibility and the full race suite then passed without source changes,
test changes, disabled VCS stamping or relaxed checks.

Repository pack validation and agent-model resolver regressions passed. Installer
regressions could not complete because GNU Stow is unavailable. Optional skills-ref
validation was unavailable. Neither tool was installed or reported as passed.

The public recheck at 2026-11-04 08:00 UTC found main still at the recorded base
and no open PRs. The external worker remains unknown. Local implementation is
complete and awaits separately authorized publication/integration. No source
inventory, native policy, runtime, frozen fixture or unrelated lane changed.
Completion evidence describes local work only; committed Project Memory validation
follows this result/evidence commit. Rehearsal results do not establish hostile
writer isolation, rollback, human review or authority to execute elsewhere.
