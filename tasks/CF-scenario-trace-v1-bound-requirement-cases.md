---
id: CF-scenario-trace-v1
status: in-progress
owner: Foundry scenario trace contributor
repositories: [spik3r/clanker-foundry]
---

# Link requirements to the actual scenario suite

## Implementation claim

Base: main at `ade8fc8e23755638e162aa0dd897edf313e22d74`. Branch:
`codex/scenario-trace-v1`. The coordinator requested the next bounded part of
the model-driven pipeline, starting with a local claim and independently stated
tests before implementation. This local claim is not a shared reservation.
Publication, integration and production use are not authorized by this task.

Read-only inspection found no open public PR and no competing public task or
branch for this slice. Existing traceability, scenario comparison and revision
tasks have completed local lifecycle frontiers. A separately started external
worker's private work cannot be inspected, so its status is unknown. Recheck
current ownership before implementation or publication, and stop on overlap.

The existing trace links requirements to an oracle catalog, not to the actual
scenario suite. Revision checks preserve both sources without inferring links.
Add an explicit, version-bound mapping from requirement IDs to real scenario
case IDs. Optionally inspect a caller-supplied native scenario report against
the exact suite literals. This is a review aid, not evidence authentication,
proof that a requirement is satisfied, or permission to run code.

Write only new files under `modeling/lifecycle/`: `scenariotrace/check.go`,
`scenariotrace/report.go`, `scenariotrace/report_binding.go`,
`scenariotrace/check_test.go`, `scenariotrace/contract_test.go`,
`scenariotrace/testdata/vectors.json`, `scenariotrace/README.md`,
`cmd/modelscenariotrace/main.go`, `cmd/modelscenariotrace/main_test.go`;
this task and its new immutable Project Memory source/events. Add exact fixture
paths to this claim before creating any further files. Do not edit native
trace/scenario packages, current revision or change-entrypoint work, compiler,
source policy, frozen older fixtures, generated files, visual assets, runtime,
CI, permissions or other tasks. Do not access held review-pack assets.

## Prospective v1 contract

`Check([]byte) ([]byte, error)` accepts a strict wrapper with exactly these fields:
`formatVersion` equal to `lifecycle-scenario-trace-request/v1`, and original UTF-8
texts in `model`, `requirements`, `catalog`, `suite`, `bindings`. `scenarioReport`
is required and is either null (not supplied) or original native report text.
The API performs no file, network, subprocess, provider or reducer calls.

`bindings` is a strict `lifecycle-scenario-trace-bindings/v1` document with
`formatVersion`, `modelDigest`, `requirementsSHA256`, `catalogSHA256`,
`suiteSHA256` and `links`. Each link contains exactly `requirementId` and
`scenarioCaseId`. All digests are lowercase 64-hex. Compare them against the
accepted model's canonical digest and exact original source bytes. A mismatch
refuses the entire result rather than refreshing any identity. These assertions
bind bytes; they do not prove a person reviewed them or establish provenance.

Use native `traceability.Build` and its `Report.JSON` without changing its
meaning. Validate the actual suite with a native validation-only runner bound
to the accepted model digest. Its reducer callback must never run. Preserve
the original native trace report bytes and hash. Check every explicit mapping
against the native requirement IDs and actual suite case IDs. Reject unknown
IDs and duplicate pairs. Many-to-many mappings are allowed; do not infer a
relationship from a shared name, catalog link, transition or assertion.

No mappings is valid incomplete information. Sort mappings by requirement ID
then case ID. Retain suite case and step order where it is meaningful. List
requirements with no scenario mapping and suite cases with no requirement
mapping. Preserve the native trace's gaps and catalog `pending`/`executable`
labels independently. A scenario observation never promotes a catalog label,
changes a trace gap, or supplies a requirement verdict.

If a report is supplied, use the native runner's `DecodeReport`, verify model,
scenario and exact suite-source identities, then bind every ordered report entry
to the original suite's case ID, step ID, input flow, command and expected
literals. Accept only a complete matched report or a valid fail-fast prefix
ending at its first mismatch with correct completed-case counts. Preserve the
entire original report text and its hash. Refuse altered expectations, omitted
or extra entries, reordered cases/steps and incompatible input chaining. Do
not invoke a reducer to check supplied actual results. A fabricated but internally
consistent report remains possible and must remain explicitly unauthenticated.

For each suite case, report one observation status: `not-supplied` when no report
was provided; `matched` only when every step is present and matched; `mismatched`
for the case ending at the valid first mismatch; or `not-observed` for later
cases absent after that mismatch. Preserve step IDs and supplied entry positions.
An expected refusal can match. There is no aggregate requirement pass rate,
readiness, approval, or acceptance decision.

The output version is `lifecycle-scenario-trace-report/v1`. Top-level fields
include `mode: example-only`, `authority: none`, `executionStatus: not-established`,
`reviewStatus: not-established`, `reportStatus` (`not-supplied` or
`inspected-for-consistency`), exact source and normalized identities, native
trace report text/hash, optional original scenario report text/hash, the sorted
explicit mappings, per-requirement scenario references, suite-case observations
and unlinked lists. Finalize the precise field names in the frozen test contract
before implementation. Null report text/hash stays null when absent; empty arrays
remain arrays. Original texts, literals, declarations and input slices stay intact.

Limits: 32 MiB wrapper, 64 MiB encoded output; model/requirements/suite 1 MiB
each, catalog 4 MiB, bindings 512 KiB, report 8 MiB, at most 4096 mapping pairs.
Native stricter limits remain in force. Cap retained native trace JSON at 4 MiB.
Use native strict parsing and explicit version selection for the wrapper and
bindings, including duplicate, unknown, case-variant, missing/null, malformed
Unicode and trailing-data refusal. Only wrapper `scenarioReport` permits null;
embedded sources retain their native validation rules. Explicit route byte/count
bounds return `ErrLimit`; invalid structure, native refusals and report consistency
failures return `ErrInvalid`; mismatched valid bindings-document digest assertions
return `ErrBinding`; unexpected encoding/native failures return `ErrBoundary`.
Errors return no partial API bytes and must not echo input.

The standalone CLI accepts no arguments or paths, reads bounded stdin, buffers
one complete output and makes one stdout write. Exit 0 means report produced,
2 means input/binding/limit refusal, 4 means unexpected boundary failure, 3 means
failed or short delivery, including a real closed stdout pipe. A redirected
partial file is not a valid result. Do not add a route to `foundry` in this lane.

## Independent acceptance before implementation

A separate contributor freezes exact output fields and hand-stated expected
relationships before production code. Use invented custom lifecycle data and
literal suites, not protected files or another system. Native-only fixture
controls may validate input contracts but must not obtain expectations from the
new implementation. Cover missing links, many-to-many mappings, duplicate and
dangling IDs, each stale binding, source/canonical identity differences, pending
catalog labels, no report, matched expected refusal, early mismatch and later
unobserved cases, forged but consistent observations, malformed/altered report
literals, wrong input flow, strict boundaries and deterministic ordering.

Keep frozen cases unchanged during implementation. Add concurrency, no-mutation,
limits, no-reducer-invocation and compiled CLI read/write/closed-pipe checks.
Run the installed exact Go 1.27.3 toolchain offline, focused and full race tests,
vet, model validation/drift, formatting, repository validation, installer
regressions and committed Project Memory validation. Record unavailable checks
honestly. Obtain independent exact-source review before any delivery claim.
No remote publication, formal PR approval, merge, deployment, cross-system
testing, new tool installation or held visual review is part of this scope.

## Frozen report field contract

The independent acceptance harness fixes these exact top-level fields:
`formatVersion`, `mode`, `authority`, `executionStatus`, `reviewStatus`,
`reportStatus`, `inputs`, `traceReport`, `scenarioReport`, `links`,
`requirements`, `cases`, `unlinkedRequirementIds`, `unlinkedScenarioCaseIds`.
The fixed status strings remain those defined above. No extra verdict is added.

- `inputs` contains `modelSHA256`, `modelDigest`, `requirementsSHA256`,
  `requirementsDigest`, `catalogSHA256`, `suiteSHA256`, `scenarioDigest`,
  `bindingsSHA256`. SHA256 fields bind original UTF-8 texts; digest fields are
  the native normalized identities. Whitespace can change source hashes without
  changing normalized identities.
- `traceReport` contains `text` and `sha256`, preserving native `Report.JSON()`.
- `scenarioReport` contains `text` and `sha256`, both null when absent. Otherwise
  they preserve the exact supplied report text and its raw hash.
- `links` contains objects with `requirementId` and `scenarioCaseId`, sorted by
  requirement ID then case ID.
- `requirements` contains every native requirement, sorted by ID, as
  `{requirementId, scenarioCaseIds}`. Case references are lexically sorted.
- `cases` contains every actual suite case in original suite order, as
  `{scenarioCaseId, requirementIds, status, steps}`. Requirement IDs are sorted.
  Every original step is retained in original order as `{stepId, entryPosition}`.
  Positions are zero-based in the supplied report, or null for absent entries.
- Both unlinked lists contain lexically sorted IDs. Empty lists remain arrays.

Same original input gives byte-identical output. Reordered binding declarations
produce the same sorted links/references, but their source hash still changes.
Native trace labels and gaps are preserved in its retained text, not promoted
by scenario observations.

The strict wrapper/bindings rules do not redefine the embedded native formats.
Only the wrapper's `scenarioReport` field permits null; native sources retain
native nullability (for example owner, review, expected task and opaque catalog
sections). The catalog reader checks its consumed fields, not all oracle-owned
payload. Native-compatible extra catalog fields and null unused sections remain
accepted. Every report identity, literal, flow or structural inconsistency is `ErrInvalid`,
whether rejected by native decoding or by this route. `ErrBinding` applies only
to valid digest assertions in the independent bindings document that do not
match current inputs. Unknown mapping IDs and duplicate pairs are `ErrInvalid`.
`ErrLimit` applies to this route's explicit byte/count caps, including retained
native trace JSON; native internal limits or structural refusals are `ErrInvalid`.
The accepted native report format is `lifecycle-scenario-report/v1`, and actual
result envelopes must satisfy bound runner protocol checks. Reports with actual
mode, authority or model digest mismatches are not accepted merely because a
standalone native encoder can represent those mismatch diagnostics.

The invented ParcelReview fixture deliberately differs from the generated
example model. Its first case contains a matched expected refusal followed by a
success, proving refusal retains the prior input even when actual task is null.
A second-case mismatch leaves that completed first case matched and the third
case unobserved. Its pending catalog declaration stays pending in every report.
All supplied observations were fabricated without running a reducer, so even
an internally consistent complete match must remain unauthenticated.

### Acceptance fixture correction before qualification

The first frozen vectors file (SHA256
`c4f4a5391a57f1d2f9205530edf132e4ac29ee82eb8427bd013fab63c842cca3`)
contained a construction defect in the fabricated `late` report. Reused Python
success dictionaries retained shared task references through deep copying.
Changing the second-case actual task outcome also changed the earlier matched
entry's expected and actual outcomes. Native report consistency accepted that
pair, but it did not match the original suite's earlier expected literal.

Independent inspection of the original suite and report found only this earlier
expected-literal discrepancy; matched and early report literals were intact.
The prospective correction restores exactly `late.entries[1].expected.task.outcome`
and `late.entries[1].actual.task.outcome` from `Fabricated mismatch` to the
original suite literal `Seal café <&>`. The second-case actual mismatch, first
case completion, counts, mappings and intended observation statuses are unchanged.

Native-only controls now compare all supplied entries with the original ordered
suite's case IDs, step IDs, commands, expectations and input chain, and check
complete-match or first-mismatch-prefix counts. The strengthened control failed
on the old literal before this correction. A named rejection test reconstructs
the original two altered fields: that report must remain native-valid but be
refused by the bridge as `ErrInvalid`. This changes no production acceptance
policy and does not derive expected behavior from implementation output.
