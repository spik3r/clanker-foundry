---
id: CF-model-scenarios-v1
status: review
owner: Astra model-first lead
repositories: [spik3r/clanker-foundry]
depends_on: [CF-model-lifecycle-v1]
---

# Offline model-scenario execution and replay

## Implementation claim

The owner requested continued Foundry work and approved this isolated local
implementation with independent checks. Branch `codex/model-scenarios-v1` is based
on exact reviewed lifecycle-demo head
`ff661f39c7e176450e175213565217958acab5dd`. PR11 remains unchanged for user review.
This local follow-up depends on PR11; it is not independently mergeable. No
publication, merge, production adoption or previously denied action is retried.

Fresh main and open PR5/PR11 task scopes were inspected. The separate receipt-time
fix owns existing runtime state-engine files; another worker's timeout tests are
in a different repository. No inspected scenario claim overlaps this frontier.

Exact scope under `modeling/lifecycle/`:
- `scenarios/README.md`
- `scenarios/types.go`, `scenarios/types_test.go`
- `scenarios/decode.go`, `scenarios/decode_test.go`
- `scenarios/run.go`, `scenarios/run_test.go`
- `scenarios/testdata/lifecycle-v1.json`
- `cmd/modelplay/main.go`, `cmd/modelplay/main_test.go`
- `scenarioacceptance/oracle_test.go`

Also this task and its new immutable Project Memory source/events only. Do not
edit the existing compiler/IR/generated artifacts, lifecycle example, independent
oracle/catalog, package files, workflow, previous task or evidence. Existing Go
`./...` validation discovers these new packages. The lead owns the implementation;
an independent contributor/reviewer owns new acceptance expectations. Commit and
register this prospective local task source and record a genuine start before
maintained implementation. Local source publication status must remain explicit.

## Purpose and boundary

Make model-designed example behavior inspectable from a small hand-authored
machine-readable scenario, rather than only imperative Go tests. Reuse the existing
pure generated reducer exactly once; do not add another business-rule engine,
generate expectations from actual results, refresh expected values automatically,
or claim arbitrary guard satisfiability or standards conformance.

The JSON adapter is a scenario-fixture format, not a production Task/Command
transport. It is initially bound to lifecycle/v1 and the exact compiled model
digest. Existing structural tests and independent acceptance remain untouched.
Keep the initial fixture small: a happy path, expected refusal, rework/replay and
same-revision content mismatch. Do not bulk-copy the existing case catalog.

## Closed scenario semantics

JSON format `lifecycle-scenarios-json/v1` maps to semantic scenario profile
`lifecycle-scenarios/v1`, with exact lifecycle profile/model identity, an ID, and ordered
cases. Each case has an explicit initial task and ordered steps. Each step has an
ID, a complete command and a complete literal expected result: OK, transition ID,
returned task or null, and ordered refusal IDs. Each observed result must also
match the fixed example-only/authority-none envelope and bound compiled model
digest; no runtime result field is ignored. No executable expressions, macros,
implicit defaults, patch/merge operations, arbitrary file paths or automatic
expected-value derivation. Missing, null-in-wrong-position, unknown or duplicate
fields and trailing/malformed JSON refuse. uint64 values use canonical unsigned
decimal strings without sign, whitespace, fractions or leading zeros. Owner,
review and returned-task null/present distinctions, and nil/empty acceptance,
are preserved exactly.

Validate the entire fixture and its finite structural/resource limits before the
first reducer evaluation. All public typed entrypoints enforce the same bounds.
Fixture bounds permit well-typed values that intentionally exercise reducer
INVALID_INPUT, rather than preemptively enforcing the reducer's business rules.
The suite's model digest must match the compiled constant, not merely another
caller-supplied label. Review digests in the example are independently fixed
vectors; the runner never computes them to make a case pass. Hashes establish
consistency only, not authenticity of actors, reviews, source or executing binary.

Cases reset to their explicit initial task. Steps run in listed order. Compare the
complete result against the explicit expected result before advancing. A matching
successful transition advances to its returned task. A matching refusal retains
the unchanged current task, including INVALID_INPUT whose returned task is null.
An expectation mismatch stops the whole suite immediately; no later step or case
executes. Expected refusal is a passing scenario comparison, not an OK transition.
Malformed fixture or wrong model binding is a rejected run, not an expected step.

Run produces an ordered deterministic report, with scenario/model identity,
observed command/result and expectation comparisons, no wall clock or random ID.
It reports example-only/authority-none and distinguishes observed example
execution from real workflow/owner/reviewer evidence. Inputs and report values
must not alias or mutate caller-owned task, command, review or expectation data.
Canonical scenario-content identity is separate from original source-byte hash;
format and semantic identity are explicit. No private consumer data in examples.

## Identity, limits and expected shapes

Scenario-content identity is SHA-256 over the complete typed suite in fixed-order
compact JSON: scenarioVersion, profileVersion, modelDigest, id, cases. Each case
contains id, initial, steps; each step contains id, command, expected, with the
complete fixture field order documented alongside the DTO. Ordered cases/steps,
literal expectations and null-versus-empty distinctions are included. No Unicode
normalization or HTML escaping is performed; Go JSON's Unicode/control escaping
is deterministic. JSON formatVersion, original source-byte hash and report are
excluded. Typed Run has no source bytes and reports a null source hash; Execute
on valid raw JSON additionally reports SHA-256 of those exact original bytes.

Limits: source and encoded suite at most 1 MiB; JSON depth at most 64; 1-16 cases;
1-32 steps per case, at most 128 total; globally unique ASCII suite/case/step IDs
at most 48 characters; valid UTF-8 scalar strings at most 8192 bytes; acceptance
lists at most 64 entries; refusal lists at most 64 strings of at most 256 bytes.
These deliberately exceed selected runtime business bounds to permit INVALID_INPUT
scenarios. The fully buffered report is capped at 8 MiB, with incremental bounded
entry accounting; overflow returns no report and executes no further steps.

Expected refusals must be an explicit nonnull list. Expected OK requires a nonnull
returned task, a nonblank transition ID and an empty refusal list. Expected not-OK
requires at least one refusal. These are fixture structural invariants, not guard
logic or a proof that an expectation is correct. Malformed expected shapes anywhere
in the suite reject before the first reducer call. Semantic mismatches remain
observable comparison failures rather than silently repaired data.

The initial independent fixture and fixed digest vectors freeze before execution
or comparison implementation. A genuinely started type/DTO checkpoint may precede
that freeze to give the independent author an exact interface. No generator or
reducer output may supply the frozen expected results.

## CLI and acceptance

`modelplay` reads one bounded fixture from stdin, emits bounded JSON to stdout and
uses a nonzero exit for malformed input or mismatch. It has no file-output option,
network/install action, arbitrary executable hook or worker authority. Normal use
is offline with pinned Go 1.27.3. No Node migration or scheduler activation.

Independent handwritten expectations and digest vectors precede implementation.
Test valid sequence/assignment/review/rework, expected refusal versus mismatch,
wrong model/profile, exact revision/content binding, malformed later cases before
execution, duplicate/unknown/null/number bounds, aliases, no mutation, deterministic
report bytes and CLI stdout/exit behavior. Kill always-refuse and selected
model-guard bypass mutants without changing expected data. The existing independent
oracle remains a separate gate, not replaced by scenario-generated expectations.

Run full Go tests/race/vet, modelc drift check, pack/whitespace and genuine lifecycle
validation; review the exact final local source independently. Preserve PR11's
branch and user-review state. Handoff local progress and limitations; subsequent
publication requires coordinated sequencing and appropriate authority.


## Local implementation handoff

The additive runner and stdin-only CLI are implemented locally. The independently
frozen fixture contains 4 cases and 13 steps; its unchanged SHA-256 is
`ef199aa93c9e70887f3da8e72aa4a79dcaf130c4cec72d27b240d6780a3e4b26`.
The independent acceptance file SHA-256 is
`894a54216a771a04a356d9379f9aeb7764e0919d664d8340719e8ddb8f56d968`.
The fixture freeze preceded execution/comparison implementation; a genuine
local type/DTO checkpoint preceded it, as permitted above. All 13 authored steps
match without changing their expected data, including expected refusals.

The report encoder also rejects internally contradictory comparison flags,
ordered differences, case/step counts, group limits and within-case state
continuity. It does not authenticate execution or prove that a caller supplied
all expected cases. Original source-byte and semantic-scenario digests remain
separate; typed Run never invents a source hash.

Boundary tests cover whole-suite zero-call preflight, first-mismatch stop,
null/empty representations, aliases, concurrent immutable replay, report budgets,
separate-process reproducibility, fixed diagnostics and short/failed writes.
Scratch mutations of the actual generated reducer (always-refuse and selected
actor/content guard bypasses) must fail the frozen independent acceptance
behaviorally. The original demo compiler/oracle remain unchanged separate gates.

Final observed checks and exact-source independent review are recorded with the
completion evidence. Review status is LOCAL ONLY: this branch and follow-up task
have not been pushed or opened as another PR. Publication/merge sequencing remains
dependent on PR11 review and appropriate further authority. No existing runtime,
production transport, worker, scheduler or denied publication action is changed.
