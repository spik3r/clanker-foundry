---
id: CF-operator-entrypoint-v1
status: in-progress
owner: Foundry operator entrypoint contributor
repositories: [spik3r/clanker-foundry]
---

# Prepare native review requests through one operator command

## Implementation claim

Owner: Foundry operator entrypoint contributor. Base: main at
2e2a190b629a983ba854d59ecc356274ec7e5a4f, after PR56 merged. Branch:
codex/operator-entrypoint-v1. The repository owner requested the next bounded
operator improvement; the coordinator approved this contract and alone merges.
The live open-PR check found no competing PR. Public branch and task searches
found no operator-entrypoint lane; those snapshots are not a shared lock.

Write only new files under modeling/lifecycle/internal/operatorcli/ and
modeling/lifecycle/cmd/foundry/, the five existing main.go adapters under
cmd/modelreviewcheck, cmd/modelproposal, cmd/modelprocess,
cmd/modelrehearsalinspect and cmd/modelrevisioncheck, plus
docs/model-driven-current-workflow.md, this task and its new immutable Project
Memory source/events. Preserve existing run signatures, tests and native signal
setup. Do not edit native packages, source policies, runtime, CI, root README,
global instructions, other claims or PR56's implementation paths.

## Preparation contract

`foundry prepare` takes exactly five required flags, each once: --intake,
--requirements, --model, --suite and --manifest. Read only those named regular
UTF-8 files. Refuse symlink leaves and nonregular files before opening; check the
opened file matches the inspected leaf. Ordinary parent path aliases are allowed.
Intake is at most 4 MiB; requirements/model/suite at most 1 MiB each; manifest
at most 64 KiB, matching native limits. Files are never rewritten.

Validate the original intake with its native API. Fill null observations in a
new intake request; existing nonnull observations must match the supplied exact
artifact bytes or preparation refuses. Replace only the observed JSON value,
using decoder offsets after native validation. Every source byte outside that
value stays unchanged. Preserve questions, decisions, declarations, dependencies,
manifest and literal expectations. Never refresh a binding or repair a mismatch.

Emit the existing first-design request, containing the new intake request text
and an existing proposal request containing the original model/suite/manifest
texts. Bound new intake/proposal/outer request by native 4/16/8 MiB limits. No
new report format or readiness state. Preparation performs no model semantics,
review or execution. Exit 0 means complete request delivery; 2 means invalid
arguments/input/conflict/bounds; 3 means failed or short output. Refusals produce
no stdout. The trusted single-writer host assumption is explicit; this is not
hostile-writer isolation or an authenticated source service.

## Native command compatibility

One binary exposes review, proposal check|execute, process json|svg, inspect and
revisions. Extract existing native CLI runners once and call them from old
adapters and the new command. Preserve exact native output, diagnostics and exit
semantics, including proposal mismatch/stale/boundary statuses, cancellation and
SIGPIPE handling. Only explicit proposal execute may invoke a candidate.

## Acceptance and handoff

A separate contributor freezes preparation fixtures and native compatibility
expectations before production changes. Test artifact UTF-8/CRLF preservation,
null/equal/conflicting observations, unchanged outside-observed intake bytes,
malformed/duplicate input, file/encoding/boundary refusals and short delivery.
Compare compiled old/new native entrypoints on good, stale, malformed and real
closed-stdout cases. Exercise explicit matched/mismatched/cancelled execution
separately. Internal capability stubs must show help, unknown commands,
preparation and check routes never invoke execution; no environment-controlled
execution hooks or plugin registry are added.

Run Go 1.27.3 full race tests, vet, modelc check, pack validation, installer and
whitespace checks. Verify the shorter guide commands on a custom fixture and
record actual outcomes. Obtain independent exact-source review, record genuine
completion and validate committed evidence. The coordinator checks hosted CI
and owns integration. No managed dispatch or permission change follows.
