---
id: CF-custom-merge-rehearsal-v1
status: review
owner: Foundry custom merge rehearsal contributor
repositories: [spik3r/clanker-foundry]
---

# Rehearse supplied concurrent design changes

## Implementation claim

Owner: Foundry custom merge rehearsal contributor. Integration owner: coordinator;
publication and integration require separate permission. This is a locally reviewed
claim, not a published reservation or an authority database. Base:
29ca17fb2d11b27e542556ad0dbb44497126a142, tree
f507162f366b3e7cac8f4e7f149eec56f65d95d5. Local branch:
cf-custom-merge-rehearsal-v1. The coordinator reviewed this contract. Fixture authoring requires a genuine
lifecycle start; production remains pending separate frozen independent acceptance.

The coordinator's public ownership review at 2026-11-04 06:37 UTC found no open
PRs and no custom-merge claim across 76 branch names and main tasks. An external
project orchestration task exists but its current contents are inaccessible. Its
work is unknown, not idle. Recheck public tasks, branches and PRs before production
or fixture edits, and stop on observed overlap. Local evidence does not exclude
unpublished work. The native merger remains owned by Byte model merge contributor;
the fixed example remains owned by Codex concurrent-design rehearsal contributor.
This claim transfers neither ownership. The completed scenario-trace lane and its
unpublished artifacts are outside this claim.

Write only these new repository-relative files:

- modeling/lifecycle/custommergerehearsal/merge.go
- modeling/lifecycle/custommergerehearsal/execute.go
- modeling/lifecycle/custommergerehearsal/merge_test.go
- modeling/lifecycle/custommergerehearsal/execute_test.go
- modeling/lifecycle/custommergerehearsal/contract_test.go
- modeling/lifecycle/custommergerehearsal/testdata/observations.json
- modeling/lifecycle/custommergerehearsal/README.md
- modeling/lifecycle/cmd/modelmergerehearsal/main.go
- modeling/lifecycle/cmd/modelmergerehearsal/main_test.go
- tasks/CF-custom-merge-rehearsal-v1-checked-execution.md
- new immutable source/event files under
  .project-memory/tasks/CF-custom-merge-rehearsal-v1/.

No existing source, fixed fixtures, task/evidence, operator CLI, source inventory,
source policy, runtime, workflows or visual assets may change. No foundry route,
publication, push, PR, merge, deployment, worker activation, held review-pack
access or unrelated platform testing belongs to this lane.

## Proposed strict request

Public APIs: Check(raw []byte) ([]byte, error), Execute(ctx context.Context,
raw []byte, sourceRoot string) ([]byte, error), and ExitCode(raw []byte, err error)
int. The standalone command accepts exactly check or execute, stdin only.
sourceRoot is trusted host configuration; the command uses its working directory.
No request field chooses files, executables, callbacks, paths or source policy.

Required request fields, with no unknown, duplicate, null or missing fields:

- formatVersion: lifecycle-custom-merge-rehearsal-request/v1.
- mergeRequest: original lifecycle-model-merge-json/v1 request text.
- base, left, right, candidate: objects each containing exactly suite and manifest,
  both original independently supplied nonempty UTF-8 source texts. Their JSON
  and native validity is checked only on the merged path.
- expectedCandidateDigest: independently supplied lowercase 64-digit SHA256 digest.

The request carries no second model source or caller-supplied merge report. First
validate the original mergeRequest through modelmerge.Execute. Only after native
validation extract its exact base/left/right raw JSON tokens. Extract the candidate
raw token only from the native merged report. Do not normalize or re-encode these
tokens, suite texts or manifest texts. Construct native proposal request envelopes
with those texts unchanged. Verify each proposal.Check CurrentModelDigest against
the corresponding native merge canonical digest. The supplied expected candidate
digest must equal the native candidate digest; never fill it from observed output.

Separate original proposal request texts were considered but would duplicate model
sources and require an extra caller reconciliation rule. The token-derived request
retains source identities without that ambiguity. It still rehearses all four
independent expectation suites rather than treating a valid merge as behavior proof.

## Preparation and refusals

Bound outer input at 32 MiB and depth 5. Keep native merge input 4 MiB and result
2 MiB, each extracted model at 1 MiB, each suite at 1 MiB, each manifest at 64 KiB,
each generated proposal request at 16 MiB, each native receipt at 9 MiB. Bound final
report at 128 MiB. Before any executor call, reserve the fixed envelope plus the
maximum escaped size of four native receipts. JSON text produced by native encoders
contains only JSON whitespace as raw control bytes, so string embedding needs at
most twice its byte length. Four 9 MiB receipts, 4 MiB original merge request and
2 MiB merge report require at most 84 MiB together; fixed metadata fits below the
128 MiB ceiling. Reject an impossible pre-execution envelope as a resource error.
An unexpected later encoding or output failure remains a boundary/delivery failure,
never evidence that execution did not occur. Strict native Unicode, JSON, model, manifest and suite validation
still applies. Bound derived request encodings before any executor can run.

Malformed/oversized outer or native input and mismatched identity assertions return
no report, an error and CLI exit 2. Conflicted or invalid-combination are native
semantic results, not malformed input: retain exact native merge bytes/hash, return
the same named outcome and exit 1. They have no candidate model, no proposal checks
and zero executor calls. The four supplied suite/manifest text fields still satisfy
the outer required/type/size rules, but are not claimed natively validated when no
candidate exists. Do not repair, select a branch or invent a candidate.

For a merged result, run all four native proposal.Check preflights. Keep all four
receipts, including stale ones. A native check error refuses the request; stale
bindings return outcome stale and exit 1. Check has no executor capability, filesystem
access, subprocess, network access or scenario execution. Current preflight yields
checked, not human review, approval, authenticated evidence or execution readiness.

## Explicit execution and receipts

Execute repeats preparation from the unchanged request. All four preflights must
be current before the first executor call. Invoke only proposalreplay.Execute,
sequentially base, left, right, candidate, with the same trusted sourceRoot/context.
Public callers cannot inject executors. A private test seam may exercise suppression.

Continue only after a native matched receipt with action execute, compilation
passed, candidateStarted true, exit 0, replay matched, cleanup removed, no failure
stage and no native error, with a live context before the next stage. A genuine
mismatch requires native mismatched/replay mismatched, exit 1, completed compilation,
started candidate, removed cleanup and no error/failure stage. Mismatch stops later
stages. Failure, cleanup failure/uncertainty and cancellation stop later stages.
Check context after every invocation, including the final one; observed cancellation
returns failed even if the last native receipt records completed matched work.
No cancellation or cleanup claim proves rollback or descendant termination.

The report version is lifecycle-custom-merge-rehearsal-report/v1. Fixed fields:
mode example-only, authority none, reviewStatus not-established, action check or
execute, outcome, requestSHA256, mergeRequestSHA256, mergeRequestJSON,
mergeReportSHA256, mergeReportJSON, expectedCandidateDigest. Four named stage
objects base/left/right/candidate contain modelSHA256 (null if no candidate),
suiteSHA256, manifestSHA256, proposalRequestSHA256 (nullable), operation,
receiptSHA256 (nullable), receiptJSON (nullable).

Operation starts as not-attempted for a native non-merged result. For merged results
it is check with the exact encoded native check receipt; only an actual native
executor invocation replaces that slot with its exact encoded native execution
receipt and operation execute. A suppressed stage retains its actual check receipt.
Native receipts preserve source, preflight, candidate-start, replay, cleanup and
failure facts. Do not fabricate an execute receipt for a skipped call or conflate
an executor call with the candidate process starting. The native merge report
retains canonical digests and source hashes; modelSHA256 binds exact raw tokens.
Original request bytes remain untouched. Hashes describe bytes, not provenance.

Outcomes: checked, conflicted, invalid-combination, stale, matched,
base-mismatched, left-mismatched, right-mismatched, candidate-mismatched, failed.
CLI exit 0 is checked/matched; 1 is semantic merge refusal, stale or mismatch;
2 is initial input/resource/identity refusal; 3 is failed or short stdout delivery;
4 is boundary failure. After execution begins, preserve actual partial native
receipts and return report plus error with failed/4. Pre-call cancellation can
return failed/4 with check receipts only. If native receipt/report encoding itself
fails, return a boundary error without fabricated receipt bytes, exit 4. A failure
to deliver a report can follow completed execution. Ignore SIGPIPE to report exit
3 on a closed stdout. Diagnostics are fixed strings without caller payloads.

## Review and independent acceptance gates

Coordinator review must freeze this contract before fixtures or production.
Commit the reviewed local claim, register that exact public-safe local source and
record a genuine v2 start before new fixture or implementation files. No publication
is implied. A separate contributor authors invented model requirements, independent
four-stage literal expectations, candidate intent/digest and native-only controls
before production implementation. Commit and hash those frozen bytes first.

Acceptance must witness both distinct edits, both merge orientations, conflicts and
invalid combinations with zero executor calls, candidate identity assertion refusal,
each stale preflight, all four real mismatch stages, cancellation and cleanup
suppression, honest partial receipts, exact native bytes/hashes, source/input
preservation, strict missing/null/unknown/duplicate/Unicode/JSON and size boundaries,
concurrent deterministic Check, and actual compiled-command usage and delivery.
Neither observed execution nor production code may author the frozen expectations.
A later independent exact-source reviewer verifies contract and untouched literals.

Use installed pinned Go 1.27.3 offline. Keep its binary directory on
PATH for native subprocesses. Do not fetch compilers or dependencies. Run focused
checks, full relevant Go race tests, vet, modelc drift, gofmt, repository pack and
installer checks, whitespace and committed Project Memory validation. Record actual
passes, failures and unavailable optional checks without weakening native policies.

Four-stage execution is recommended because only candidate execution could miss an
incorrect branch expectation or a discarded independent edit. A smaller check-only
four-proposal bridge is a viable separate milestone if execution scope is rejected;
it provides no behavioral evidence and must not be described as a rehearsal pass.

## Local implementation result

Local implementation source is cdcd29fdc3657968f33e7c86f3b5cf4278aaf07c.
An independent source reviewer inspected that exact revision and reported no
remaining findings. The reviewer independently ran actual native positive and all
four mismatch paths plus framing, limits and source-byte probes, and confirmed
unchanged hashes for all nine reviewed source, support-test, documentation and
acceptance files. The implementation remains byte-identical to that reviewed source.

The separate acceptance contributor froze invented ArchivePacket behavior before
production in ff665b79ca3723a825d874e0fa8e763a2f338223. Initial source edits had
begun when coordinator review found two assertions using the nested preflight
replay label for the top-level native receipt. The independent contributor corrected
only those assertions in 670096f4c39b14abdc926e0c3dfa4c83da39c73f, before any
bridge test ran. No production accommodation or behavioral expectation change was
made. The corrected contract SHA256 is
0cfc28d393be36df85eecb6873b36647e83fac9f7c8f248eea6005704a5a434e.
The original observations SHA256 remains
0c1c99a345875a292d7add9e9b032d246266c01a62555f68ea84f7f7b467e93b.

Go 1.27.3 offline validation passed on frozen source:

- Focused race tests for custommergerehearsal and its compiled command passed.
- Full go test -race -count=1 ./... passed 77 tested packages; two packages had
  no test files. This includes actual four-stage matches, both orientations and
  each mismatch stage with later stages suppressed.
- Go vet, modelc validate/check, gofmt and whitespace checks passed. Generated
  artifacts matched and the drift check wrote nothing.
- Repository pack validation and agent-model resolver regressions passed.
- Installer regression could not complete because GNU Stow is unavailable on this
  host. Optional skills-ref validation was also unavailable. Neither was installed,
  fetched, silently skipped as passed, or qualified by a previous lane's results.

Synthetic cancellation, cleanup, malformed-receipt and partial-result controls
establish sequence and honest retention only. Actual native/compiled-command runs
separately establish the finite behavioral and report-delivery observations.
No existing native sources, source policies, inventories, fixed fixtures, routing,
runtime, workflows or unrelated task/evidence were changed.

The coordinator's public recheck at 2026-11-04 06:58 UTC found main still at
29ca17fb2d11b27e542556ad0dbb44497126a142, zero open PRs, and no competing public
custom merge lane. The external orchestration task's current work remains unknown.
This local claim does not clear that uncertainty or acquire its ownership.

Local implementation is complete and awaiting separately authorized integration.
No push, PR, merge, publication, deployment or worker activation occurred. Completion
evidence records local work only; full committed Project Memory validation follows
the result/evidence commit. The next action belongs to the coordinator: retain this
local result or obtain any needed publication/integration authority. Rehearsal
matches remain neither human approval nor authenticated execution evidence.
