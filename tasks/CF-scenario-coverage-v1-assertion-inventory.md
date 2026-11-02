---
id: CF-scenario-coverage-v1
status: in-progress
owner: Codex scenario assertion inventory contributor
repositories: [spik3r/clanker-foundry]
depends_on: [scenario explanation and scenario report APIs at Foundry main]
---

# Inventory scenario assertions against model declarations

## Prospective implementation claim

The coordinator assigned this bounded local lane after checking Foundry main
`55cc43870ad8017ecb334536a269e6ac9c49afee` and the permitted open PR list.
Only #5 and #20 were open; neither is in this frontier. A coverage branch search
found none. Existing compiler-profile acceptance, UML/process views, #20's
restricted visual artifact, and #5 are excluded. This local claim is not a
shared reservation; recheck current claims before any separately authorized
publication and stop for overlap.

The exact maintained scope is new files only:

- `modeling/lifecycle/scenariocoverage/coverage.go`
- `modeling/lifecycle/scenariocoverage/coverage_test.go`
- `modeling/lifecycle/scenariocoverage/README.md`
- `modeling/lifecycle/cmd/modelcoverage/main.go`
- `modeling/lifecycle/cmd/modelcoverage/main_test.go`
- This task and this task's new immutable Project Memory files.

Register and start genuine Project Memory after committing this source, before
implementation. Freeze reviewable literal expectations before implementing the
mapping; identify their authorship honestly. Perform exact-source independent
review and all applicable local tests before handoff. No remote push, PR,
merge, deployment, worker activation or network-dependent execution is
included. Do not inspect or bypass any denied artifacts or access controls.

## Bounded outcome

Provide a pure deterministic Go inventory of supplied scenario report assertions
against the exact normalized model. Compose existing scenario explanation and
report checks, not a second reducer, guard evaluator, or comparison engine. List
all declared transitions and their guard associations in stable order, including
those with no corresponding assertion. Group only matched report entries into
successful-transition, explicit guard-refusal and revision-limit references;
separately list matched invalid-input and illegal-transition references. Preserve
mismatch entries as excluded references, never treat an expected or mismatched
result as coverage. Guard-refusal references indicate only explicit listed reasons;
do not infer other guards passed. Every reference identifies case and step. A
zero count means no eligible supplied assertion, not impossible behavior. No
percentages, inferred satisfaction, security, complete test coverage or proof of
execution. Digest and matched status are supplied consistency claims, not
provenance. Preserve order deterministically and do not mutate or alias input.

A bounded stdin-only command accepts one strict versioned wrapper carrying an
existing JSON-v1 model and existing report. Reject malformed, oversized,
unknown, duplicate and missing fields, wrong profiles, mismatched binding and
inconsistent or unmappable results before output. Buffer complete output, use
fixed private diagnostics, and treat failed or short writes as failure. No file,
network, process, provider, image, or worker access by the implementation.
Execution evidence or callback authentication remains outside scope. Assertions
and IDs may be sensitive and callers must keep output private.

## Acceptance

Hand-authored literals cover success, explicit multi-guard refusal, fixed runtime
reasons, unmatched entries, absent transition/guard pairs, deterministic ordering,
forged-but-consistent report limits, and invalid model/report combinations.
Tests check whole-input rejection, no aliases or input mutation, strict wrapper
and write failures. Run Go 1.27.3 focused and full race tests, vet, `modelc
check`, repository validation, whitespace and committed-diff Project Memory
validation. Report unavailable checks, independent-review status and hosted CI
separately; do not claim tests establish behavioral or requirements coverage.
