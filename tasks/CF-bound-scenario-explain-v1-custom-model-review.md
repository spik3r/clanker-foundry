---
id: CF-bound-scenario-explain-v1
status: complete
owner: Byte bound-model explanation worker
repositories: [spik3r/clanker-foundry]
---

# Typed custom-model explanation

## Implementation claim

Owner: Byte bound-model explanation worker. The coordinator authorized this fresh
local implementation after original producer semantic review and explicit consumer
harness owner clearance. Exact corrected contract SHA-256
e92d8bb63abca0481e14b04c2362686e7ee7a6b0cb73d0295caa68746fb9a463.
Scope is exactly the eight modeling paths below plus this new task and its genuine
immutable evidence. Fresh current-main PM inventory has no active/conflicting
heads; branch codex/bound-explanation-v1 is isolated from all held work.
Register this committed claim and record start before maintained implementation
edits. Local source work is authorized; publication is separate and not implied.
Current public base: 107fd91742291b651c54837cc46b6c8b424e4f8a, tree
63c79053a5ba6bdcd2747e1c29b267f7d794a12e. No held profile-policy, reviewpack or
private consumer ancestry is included. Profile publication is explicitly held;
this proposal neither retries nor routes around that action.

## Need and non-goals

Merged Runner and CheckWithRunner execute independently authored custom-model
suites, but scenarioexplain.Explain hardcodes the example model and package-level
report encoder. Reuse its declared refusal-to-guard/requirement projection for an
explicit custom runner, without a second mapper, runner, CLI or report parser.
This supplies reviewer context, not the missing authentication/operational gates.
No callbacks, evaluation, I/O, subprocess, runtime activation, provenance
attestation, review/coverage verdict or UML/SysML claim is added.

## Exact proposed maintained frontier

1. modeling/lifecycle/scenarioexplain/explain.go: narrowly factor the existing
   mapper so baseline and opt-in validation paths reuse its single mapping body.
2. modeling/lifecycle/scenarioexplain/bound.go: additive typed API, result type,
   version constant and private bounded encoding check.
3. modeling/lifecycle/scenarioexplain/bound_test.go: independent typed acceptance,
   refusal/ownership/limit/mutation tests; no maintained subprocess harness.
4. modeling/lifecycle/scenarioexplain/testdata/bound-vectors.json: independently
   authored public-safe expected projections for two existing custom models.
5. modeling/lifecycle/scenarioexplain/CONTRACT.md: additive contract and boundaries.
6. modeling/lifecycle/consumerintegration/consumer_test.go: copy the new frozen
   explanation expectation to its existing disposable external-module fixture
   folder, pin its bytes, and retain actual post-restoration proof execution.
7. modeling/lifecycle/consumerintegration/testdata/adapter.go.txt: check the new
   explanation after actual custom replay for both already generated models,
   with zero additional callback invocations and exact frozen projection checks.
8. modeling/lifecycle/consumerintegration/README.md: describe those finite checks.

Plus one new public-safe canonical task and its genuine immutable PM records.
No edits to scenarios, model, canonical/compiler, generated sources, qualification,
designchange, artifactbundle, default CLI/report decoder, original explanation
vectors/tests, the eight existing consumer fixtures, runtime/workflows, private
consumer pins or source, held reviewpack/dossier/profile work, HT390 or diagnostics.
The eight fixture bytes are frozen from the current main and preserved exactly.

## Typed API and exact output

Add:
  ExplainWithRunner(model.Model, scenarios.Report, scenarios.Runner)
      (BoundExplanation, error)

The concrete Runner is supplied explicitly by trusted Go composition, never
selected from a request or invented from a report identity. Zero runner refuses.
The output is a distinct type with these required JSON fields, in this order:
formatVersion, identitySource, mode, authority, status, modelDigest,
scenarioDigest, sourceDigest, matched, casesCompleted, stepsExecuted, entries.
Constants:
- formatVersion: lifecycle-bound-scenario-explanation/v1
- identitySource: caller-supplied-go-binding
- mode: example-only
- authority: none
- status: caller-supplied-assertions

The remaining fields retain existing Explanation meanings and use its unchanged
Entry, GuardLink and Expression shapes. SourceDigest is an owned nullable copy.
Entries and Guards are non-null arrays, Expression.Args always non-null. Preserve
all supplied comparison differences, unknown expected IDs/literals and accepted
null/empty distinctions; never classify expected reasons as actual guards.
Do not relax Runner.EncodeReport: its provider-envelope validation rejects every
nil actual.Refusals slice, including success. A non-nil empty slice can be accepted
for success. The legacy encoder can accept nil actual refusals; retain that old
behavior separately. Freeze the exact bound rejection and legacy preservation
controls. Null/empty preservation applies only to values accepted by the selected
existing validator.
No arbitrary BoundExplanation encoder, new CLI or bound wire decoder is exported.
The API verifies the fixed compact non-HTML-escaping JSON representation plus LF
fits the existing 8 MiB MaxOutputBytes before returning typed success. There is
no promise about a caller's different serializer. Any failure returns the zero
BoundExplanation and existing fixed ErrInvalid, with no partial typed result.
Callers must not mutate the input model/report concurrently during a call. This
pure API owns its returned collections; it does not synchronize caller memory.

## Validation order and shared mapping

1. ModelBinding validates the explicit nonzero Runner and lifecycle/v1 profile.
2. Normalize the complete supplied model through the existing model package;
   reject any guard ID colliding with INVALID_INPUT, ILLEGAL_TRANSITION or
   REVISION_LIMIT. An unused guard already fails Normalize; collision-branch
   tests must instead rename a used guard and its transition reference in a valid
   custom model and bind that model's actual canonical identity.
3. Compute canonical model digest. Require equality with both the runner binding
   and the supplied report digest; never repair, substitute or rebind identities.
4. Call Runner.EncodeReport on the entire report before mapping. This is bounded
   structural/comparison validation only and must not call the callback. Preserve
   all existing ordering, counts, provider-envelope and mismatch requirements.
5. Reuse the existing whole-input structural mapping: known compatible transition
   for successful/guard/overflow assertions; exact fixed-refusal rules; sorted
   unique transition-local actual guards; declared requirement text/expression;
   limits and owned output. No guard evaluation or effect verification occurs.
6. Create the explicit bound envelope and check final encoding/output bound.

Default Explain continues using generated.ModelDigest and package-level
scenarios.EncodeReport, with its exact validation behavior and all existing
output bytes unchanged. Execute, request version, CLI and strict DecodeReport
remain default-only. Bound API may explicitly use a baseline-compatible Runner;
its result still carries the bound version and caller-binding label.

Runner methods ModelBinding/EncodeReport do not invoke the callback; no Run,
Execute, Transition or acceptance API is called by this package. Typed Report
has no envelope/origin field. A fabricated internally consistent report, or a
report produced elsewhere, can pass. The bound output does not prove a bound
wire origin, complete suite coverage, actual replay, effect correctness,
independent authorship, source/binary provenance or real review/approval.
Reports and model text may be private. Output omits task/actor/review fields as
before but IDs, literals, requirement text and expressions can contain that text.

## Independent acceptance before mapping changes

Reuse only public consumerintegration model-v1/model-v2 and their frozen suites,
requirements, manifests and expected inventory. No private HT407/HT408 fixture
copying. An independent assistant writes complete projection literals and
critical negative expectations before implementation; commit that freeze first.
Expected guard/requirement text, expressions, transition links and case/step
projection fields come from the frozen source declarations and authored literals,
never target output, reducer replay or an inverse round trip.

Required tests:
- Both distinct custom model identities succeed with their own literal reports;
  differing submit guard declarations map to prewritten model-specific links.
  The v1 frozen suite has no submit-guard refusal, so add an explicitly fabricated
  but structurally consistent v1 submit refusal to exercise its original nonempty
  expression alongside v2's equality expression; this is not guard truth evidence.
- Complete success, fixed refusals, single/sorted multiple guard refusals,
  preserved mismatches including unknown expected IDs, source nil/present,
  MaxUint64 input, private text escaping and deterministic bytes/owned output.
- Zero Runner, stale/mismatched model-report-binding, invalid report counts or
  comparison fields, malformed provider envelopes, unknown/foreign/unsorted/
  duplicate actual guards, incompatible transitions, mixed fixed/guard reasons.
- A genuinely valid custom model declaring a reserved guard ID reaches the
  reserved-collision refusal (not an earlier stale digest); fake success without
  actual effects is explicitly accepted only as structural caller assertions.
- Panic/counting callbacks prove exactly zero invocations for every accepted and
  rejected explanation input. No error or success may execute the callback.
- Aliasing controls mutate inputs/outputs in both directions, including SourceDigest,
  nested expressions and refusal/difference lists. Preserve accepted empty/null
  cases without coercion. Bound total escaping/output limits reject whole result.
- Legacy frozen vectors, default API, CLI and strict decoder bytes unchanged;
  default Explain still refuses custom models.
- Existing external consumer harness actually builds both generated packages and
  executes both fourteen-step suites; explanations match the new frozen projection
  with no additional callback calls. Existing adapter/reducer mutation kills and
  restoration remain intact, and final restored run executes the explanation
  check for both models rather than merely compiling it.
- New mapping mutants (e.g. skipped binding, dropped guard, rebound output or
  erased mismatch) compile but fail named expected assertions. A build failure
  is not a mutation kill. Restore and run normal tests afterward.

Full current-main race/vet/format/model+trace drift/schema/repository checks,
original-fixture continuity, fresh PM and independent exact-source review are
required before completion. Source pins cannot change: the new package code is
not in the qualifier's current source closure, and existing pinned files stay
byte-identical. Publication requires its own permitted coordinator action; this
contract cannot waive any denied publication, browser, diagnostic or private gate.


## Observed local qualification

Exact implementation ea22fb7896e97b774db09b10bc4b37d8ef3035ba, tree
0f5c7f6b7ec3aca4334434e5efa3d13a6e7a887a, was independently reviewed with no
material finding. The common mapper keeps baseline validation/output behavior;
the explicit bound wrapper never invokes its callback and checks the complete
new envelope before returning success. No interpreter, CLI, wire decoder or
source-policy change was added. The later e5e8a267 checkpoint changes evidence
only, not maintained source. Fresh read-only fetch still identifies public
main 107fd917 as the unchanged ancestor; no held history was imported.

Independent full projection literals were frozen before mapper changes at
283b1c845d1a8f85337338b09822f5c7134e841b. The vector SHA-256 remains
c5e011aba6dd14f5eaf2de0016de10dd1359196decaf0eb8d6edf4463fc147f9;
the typed acceptance file remains
34c7ff22c8921adbaa20544c76d42b63f025e44b6ec00a4417a570c1baf01916.
The original producer independently checked all twenty-eight public suite-step
projections and the separately fabricated v1 guard assertion before mapping.
Original default vectors, all eight prior consumer fixtures, the pinned source
closure and every excluded path remain unchanged. Fixtures were not refreshed
from explanation or replay results. The type/claim/start/freeze sequence is
preserved in the fresh local history.

The author and original producer independently passed actual Go 1.27.3 offline
full-module race tests (45 tested packages), focused verbose race, vet,
formatting, modelc validate/check, modeltrace check, lock-matched installed
schema tests (2/2) and repository validation. Optional skills-ref is unavailable.
The independent oracle author also reviewed exact source and passed focused
race/vet. Actual external execution built both separate generated models and
replayed both fourteen-step suites; full bound explanations matched the frozen
projections with zero extra callbacks. Existing two reducer mutations and nine
adapter mutations per model failed as intended; final restored tests actually
replayed and explained both models. The precise independently encoded boundary
control has an 8,388,608-byte legacy projection and an 8,388,701-byte bound envelope;
the valid model/native report reach the final bound-output refusal. Pure typed
cases remain explicitly fabricated assertions, not observed replay claims.

The original producer separately ran the unchanged default CLI with a preserved
prior request and confirmed byte-identical explanation output. Bound nil actual
refusals reject, while the existing default behavior remains separately tested.
All success/refusal paths use zero-callback controls; stale identities, reserved
used-guard collisions, invalid report comparisons, ownership, private literals,
null/empty differences and whole-output limits retain their stated boundaries.

Six additional author-only source-overlay mutants compiled and failed named
assertions: ignored native validation, dropped guard projection, erased mismatch,
omitted final envelope bound, rebound output identity and callback invocation.
Restored focused race passed and maintained source bytes did not change. The
initial scratch mutation runner expected panic text for the callback mutant;
the test instead reported its deferred callback-count failure (one call). The
original log is retained; only the scratch runner's marker was corrected, not a
test or expected value. These mutation observations are not attributed to either
independent reviewer.

Local technical completion does not publish this branch or change any held
publication, browser, diagnostics, visual or private operational gate. Consumer
adoption and dependency repinning remain separately owned. Output still describes
caller assertions and declared links, not authenticated source/execution, human
review, behavioral coverage or operational authority. Final committed lifecycle
and metadata-only continuity are checked separately before any permitted
coordinator-controlled publication.
