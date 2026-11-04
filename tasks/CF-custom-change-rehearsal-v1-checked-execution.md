---
id: CF-custom-change-rehearsal-v1
status: review
owner: Foundry custom change rehearsal contributor
repositories: [spik3r/clanker-foundry]
---

# Rehearse a supplied model change against independent literals

## Implementation claim

Owner: Foundry custom change rehearsal contributor; the coordinator alone merges.
Base: main at 249cf7434a43c4d1b0ab6f04b2f0d7d396b7c2ab. Branch:
codex/custom-change-rehearsal-v1. The owner requested another bounded batch and
the coordinator approved this custom execution milestone. Live open PRs 69 and
70 own provider reply handling and three separate review CLIs; those paths and
claims are excluded. Existing fixed rehearsal and native source policies remain
under their current ownership. This public claim is not an authority database.

Write only new modeling/lifecycle/changerehearsal/change.go, execute.go,
change_test.go, execute_test.go, contract_test.go, testdata/observations.json and
README.md; new cmd/modelchangerehearsal/main.go and main_test.go;
docs/model-driven-current-workflow.md; this task and its new immutable Project
Memory source/events. Do not edit native packages, qualified source inventories,
old tests/fixtures, existing command files, foundry help/routes, runtime, CI,
global permissions, held reviewpack gates or other tasks.

## Request and identity contract

The standalone CLI accepts exactly check or execute. Both consume the same
unchanged strict lifecycle-change-rehearsal-request/v1 JSON with patchRequest
text, baselineProposalRequest text, candidateSuite text, candidateManifest text
and expectedCandidateDigest (mandatory lowercase SHA-256). Public APIs are
Check(raw []byte) ([]byte,error) and Execute(ctx,raw,sourceRoot) ([]byte,error).
SourceRoot is trusted host configuration, never request data.

Use existing modelpatch.Execute to derive the candidate after token and patch
identities. The baseline proposal's actual canonical model digest must equal the
patch before digest. The caller's separately reviewed expected candidate digest
must equal the patch after digest. These are identity preconditions, not approval.
Keep the native candidate token and original suite/manifest bytes; do not refresh
bindings, duplicate patch semantics, repair expectations or accept a supplied
planner report as the candidate authority.

Check invokes only the native patch planner and both proposal.Check preflights.
It receives no execution capability. Stale preflight stays an observed refusal;
no candidate runs. Execute recomputes the same preparation and requires both
preflights current before any executor call. Run baseline first, then candidate
only after baseline matched, native cleanup succeeded and context is live.
Baseline mismatch, failure, cleanup uncertainty or cancellation suppresses the
candidate. Cancellation before a stage makes no new executor call and never
fabricates a native execution receipt.

## Receipt and failure contract

The fixed lifecycle-change-rehearsal-report/v1 receipt has mode example-only,
authority none, reviewStatus not-established, action, outcome, requestSHA256,
patchRequestSHA256, expectedCandidateDigest, candidateModelSHA256,
patchReportSHA256/JSON, and baseline/candidate objects. Each side records its
operation (check or execute), proposalRequestSHA256 and receiptSHA256/JSON.
The full native patch report contains candidate content and canonical identities;
native proposal receipts retain their source hashes, actual CandidateStarted,
replay, cleanup and failure facts. A suppressed side retains its actual check
receipt and operation. No general execution-status or approval envelope is added.

Initial malformed, resource or identity-precondition refusal returns nil receipt
and exit 2. Check-current or both matching runs exit 0; stale preflight or mismatch
exits 1. After an executor attempt, preserve actual partial native receipts and
fixed failed outcome on boundary errors; return report bytes plus the error and
CLI exit 4. Cancellation can retain preflight receipts with failed outcome even
before an attempt. Failed/short report delivery exits 3 and may follow execution.
Diagnostics are fixed and omit payloads; receipt fields remain authoritative.

Bound the request at 32 MiB and receipt at 64 MiB. Preserve native patch 3 MiB,
baseline proposal 16 MiB, candidate suite 1 MiB and manifest 64 KiB limits. Keep
one native receipt slot per side, replacing check with actual execute receipt
when attempted, rather than duplicating a complete check envelope. No arbitrary
program selection, network policy, sandbox, descendant cleanup guarantee or
qualification source update is introduced.

## Acceptance and handoff

Commit and publish this claim, register a reviewed source and record genuine
start before any new tests or production edits. The coordinator reviews this
claim. A separate contributor freezes a real behavioral change with independently
stated before/after suites, manifests and candidate digest before implementation.
Never copy expected results from observed execution.

Check exact native bytes/hashes and original input preservation. Private executor
controls prove zero calls for check/stale/identity refusal and candidate suppression
for baseline mismatch/failure/cleanup/cancellation, preserving honest partial data.
Run separately authorized invented local matched and candidate-mismatch revisions,
strict/boundary controls and actual compiled-binary delivery checks. Use existing
native execution only; public APIs never accept a caller-selected executor.

Run Go 1.27.3 full race, vet, modelc drift, gofmt, pack, installer, whitespace and
committed Project Memory validation. Obtain independent exact-source review and
hand the PR to the coordinator for hosted checks and merge. No fixture match,
source hash or local run establishes human review or production authority.


## Implementation result

Source a5fcd1771ec028a1ff1557634dfe37d0f4f0609a adds the new package and standalone
command, with the current workflow section. Ordinary merge 87a2a44 incorporates
main 82b62e6 without changing this task's Go source or independent observations.
The independent native behavioral controls and fixed input bytes were committed
in 859f790 before production; their hashes remain unchanged. The coordinator and
a separate reviewer reviewed the exact final source with no remaining findings.

Go 1.27.3 full `go test -race -count=1 ./...` passed: 75 tested packages and two
packages with no tests. Focused checks exercised actual native matched runs and
both mismatch stages, as well as actual compiled-binary stdout and closed-pipe
behavior for check and execute. An initial full run correctly caught source
changes made while checkout-immutability tests were active. The complete rerun
passed after source was frozen; no test or native policy was weakened.

Four separate disposable-clone mutations were detected: removing the candidate
digest precondition, baseline mismatch stop, cleanup prerequisite or context
gate. These controls used the private executor seam without native execution.
Actual standalone invented examples separately produced checked, matched,
baseline-mismatched and candidate-mismatched reports with exit codes 0, 0, 1 and
1. Both matching runs reported candidate started and cleanup removed; a baseline
mismatch retained the candidate's actual check receipt without a child claim.

Go vet, modelc generated-artifact check, gofmt and whitespace checks passed.
Pack validation and installer regressions passed; skills-ref was unavailable
and the validator reported that optional check skipped. Full committed Project
Memory validation follows the completion commit. Existing native source,
qualified inventories, old fixtures and foundry routes remain unchanged.

The examples and injected failure controls are finite observations. They do not
establish human approval, authenticated provenance, descendant rollback or all
possible host failures. The next step is the coordinator's final hosted checks
and merge; this contributor has not merged the PR.
