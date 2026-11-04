---
id: CF-operator-entrypoint-v1
status: done
owner: Foundry operator entrypoint contributor
repositories: [spik3r/clanker-foundry]
---

# Prepare native review requests through one operator command

## Integration result

[PR63](https://github.com/spik3r/clanker-foundry/pull/63) merged as
`54f19e1e0dc0026c571fcf28735b7890c1a74b17` after coordinator and independent
source review. All five hosted workflows passed on
`3ac7b00fff928448d67b6899b1a0fd504943172c`, including the full Go workflow.
The local race, native compatibility, guide and custom-fixture results below
remain the observed validation record. The completed claim reserves no further
work. Preparation still grants no execution or integration authority.

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

## Review handoff

The `foundry` binary now prepares the existing first-design request and exposes
all five native command families through shared CLI runners. Old adapters keep
their run signatures, test-facing revision classifier and signal setup. The new
main ignores SIGPIPE on every route and installs cancellation handling only for
proposal commands, matching their existing contract. Preparation leaves source
files untouched and changes only the new intake's observed value.

A separate contributor froze preparation bytes and native compatibility checks
in a304226 before production. Their initial test harness referenced a private
native report type and did not compile; independent correction 1bd3f97 uses a
local test DTO without changing assertions or expected data. The original data
SHA-256 remains 8b70e7b9658d124f3ce3f5987b7d20e73c305964801aa6b3989e9dc59cd4edc3.
A builder-owned wrapper-bound test initially used input below the bound; its
corrected fixture now measures both nested requests below their bounds and the
outer request above 8 MiB before checking refusal. No product fix or frozen
expectation change was made for that test mistake.

Exact source 60953dc passed Go 1.27.3 full offline race tests: 72 tested packages
passed and two command packages compiled without local test files. The operator
suite also builds and runs the new binary and all five standalone binaries,
compares complete native outputs and checks actual closed stdout, fixed
refusals and FIFO admission. Focused operator and five legacy CLI race suites,
vet, modelc drift, gofmt, pack validation, installer regressions and whitespace
checks passed. Optional skills-ref is unavailable. Main's store-only change was
merged normally and changed no task-covered Go or guide bytes.

All four revised guide Bash blocks ran unchanged in private scratch. Review
retained unresolved questions and not-run status; separately invoked replay
matched nine cases/nine steps with cleanup removed. The separate fixed carrier
inspection reported baseline two cases/two steps and candidate four cases/six
steps, with execution not established. No new visual inspection is claimed.

A separate custom ParcelSeal preparation/review flow retained original UTF-8 and
CRLF inputs and separately replayed four cases/six steps with cleanup removed.
All five input hashes stayed unchanged. Its first execution attempt refused at
the toolchain gate before candidate start; selecting the installed Go 1.27.3
binary on PATH allowed the unchanged request to pass. This was an environment
correction, not a native policy change or an expectation update.

The coordinator and independent reviewer approved exact source 60953dc; the
reviewer independently passed the operator and five legacy race suites. Native
APIs, old tests, generated artifacts, source policies and runtime implementation
are unchanged by this task. Candidate execution stays explicit; request assembly
or matching output establishes no human approval or integration authority.

Next: genuine completion, committed evidence validation, coordinator review of
hosted checks and integration of PR63.
