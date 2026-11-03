---
id: CF-scenario-explain-adoption-v1
status: completed
owner: Byte current-main scenario explanation adoption worker
repositories: [spik3r/clanker-foundry]
---

# Adopt the existing baseline scenario explanation locally

## Implementation claim

The coordinator authorized fresh adoption after original-lead dependency/history
review and explicit scenario-owner clearance for the two additive decoder files.
Current public main is 713c24577826b893dd4c5527763078e56c3c50d3, tree
964a4bd69b5ae3d57f00952f065a2d6a5166bf86. Active tasks and public boundaries were
checked; no existing claim owns these absent paths. This branch is
codex/scenario-explain-adoption-v1. Claim/start and independent expected-value
freeze precede implementation. Initial adoption and qualification are local;
publication follows a separate coordinator decision and exact-source review.

The original source is completed local dfe14ea69256698aca869e3bdb391c90c3141f60,
unchanged in held snapshot 747f5e5994fd69639bbea75a8c99cf3130fc5575. Preserve its
old task, completed run and held ancestors as history; do not copy them as fresh
adoption evidence or merge their ancestry. Reviewpack/process-view visual holds,
private consumer/exporter pins and timing publication remain separate.

Exact frontier under modeling/lifecycle:
- scenarioexplain/{CONTRACT.md,types.go,explain.go,explain_test.go}.
- scenarioexplain/testdata/vectors.json.
- cmd/modelexplain/{main.go,main_test.go}.
- scenarios/{report_decode.go,report_decode_test.go}, new files only.
- scenarioexplain/integration_test.go, independently authored before adoption.
Plus this task and its own immutable Project Memory source/events only.
No existing scenario type/conversion/Runner/encoder, model/compiler/generated,
fixture, workflow, archive, visualization, private pin or runtime file changes.

## Reuse and compatibility contract

Reuse the held Explain/Execute mapping and strict DecodeReport without a new
interpreter, runner, output protocol or second explanation command. Six non-scenario
production dependencies match the held source. Current scenarios/types.go,
decode.go and run.go changed through already-merged model-bound Runner work;
reused DTO/conversion helpers and default EncodeReport still exist, but the whole
dependency closure is NOT byte-identical. Fresh current-main tests are required.

This version supports the compiled baseline model only. Strict report decoding
rejects BoundReportVersion and identitySource rather than discarding/rebinding
identity. The typed Report has no envelope/origin discriminator: a compatible typed
report can be invented or come from a baseline-bound callback. Explain cannot
identify that origin and must always label output caller-supplied-assertions.
Do not promise all bound-origin typed reports are rejected or claim custom-model,
callback provenance, guard evaluation, dynamic coverage, review or execution.

Pin the original nine files below. Copy/freeze original CONTRACT/types/vectors
first, preserving literal vector SHA-256
206113fe0d07f04ed3989bc63134b01a64defa5994f9d2a195a0584fa6bf255e.
Freeze independently authored new integration expectations before copying mapping,
decoder and command logic. Do not regenerate expectations from Explain/Execute or
new DecodeReport output. Original mapper/decoder and their original tests remain
byte-identical unless a material finding obtains a separately reviewed repair.

One prospective CLI repair is bounded: a real closed stdout probe of the original
command exited by SIGPIPE (-13), despite its documented exit 3 contract. Freeze
an actual closed-pipe subprocess regression before adding signal.Ignore(SIGPIPE)
to main. Preserve no-args/stdin/stdout interface, toolchain gate and fixed private-
safe diagnostics; retain fake-writer tests as additional controls. CONTRACT may
append adoption/compatibility/use guidance. These intentional CLI/doc deltas are
not described as byte-identical import. No product process runner is added.

## Independent integration expectations

Reuse the already frozen four-case/thirteen-step lifecycle suite via existing
scenarios.Execute and EncodeReport, then the new report decoder and explanation.
Pin raw source/fixture identities and independently state expected explanation
entries, refusal classifications and declared guard/requirement/expression fields.
The test executes existing baseline code only to produce actual results; expected
explanation data is authored separately from contract and frozen sources.

Preserve mismatches and nil/empty distinctions. A fabricated compatible report
must remain caller-asserted. Use a same-baseline NewRunner with an instrumented
callback only for EncodeReport: bound wire must be refused without invoking the
callback; its baseline-compatible typed Report may still be explained as caller
assertions. Reject changed candidate model/digest with zero explanation/output.
Never repair or rebind a suite/report to make the integration pass. All existing
example, tutorial, scenario and generated bytes remain unchanged.

## Qualification and limits

Run original frozen vectors and decoder roundtrip/strict controls, new integration
and actual closed-stdout red/green tests. Run actual Go 1.27.3 offline focused and
full race tests, full vet, format, modelc validate/check, modeltrace check, authoring
schema and repository validation. Use only bounded trusted test subprocesses;
there is no cancellation/sandbox/descendant-cleanup guarantee for arbitrary code.
Independently review exact source and complete lifecycle/fixture continuity before
publication. Keep output assertions distinct from authenticated execution evidence.
No deployment, worker activation, browser rendering or UML/SysML conformance claim.

## Original source identity manifest

Pins identify local source bytes, not authentication or execution evidence.
Only the prospectively named CLI main/test and CONTRACT may intentionally differ.

| Path under modeling/lifecycle | Git blob | SHA-256 | Bytes |
|---|---|---|---:|
| scenarioexplain/CONTRACT.md | a90c2e0c48217139e605a9055c64b195181a26c6 | ae99e44e9fdba4a39f6f4a418b586b50b79e5ff0a99e67482652ffeb839e07aa | 9483 |
| scenarioexplain/types.go | 9ccb48196b81d8f4cb2c816a6dc627334317d0dd | fbd9af915c46b1e1b7e5cf58d8d9bb539be5c7d46d29132f543f76d8346cee22 | 2850 |
| scenarioexplain/explain.go | 550b3af6bdf0ed0b0e3d43a03c19fa0d556322c9 | bc411e8870b1cad4e679ea0378c3ba62225f48bd82d8b88069ef7333b8757477 | 6750 |
| scenarioexplain/explain_test.go | 04ecf24aa910d41ba1a022b5706f6c7ec73864ff | e7a83beb0550d77bbc26d1cc07b4e31855df599f1eaac78d33b552a83eef3a20 | 8275 |
| scenarioexplain/testdata/vectors.json | 9da35e8cd26d6e80ce7751a2b9196787d0b49139 | 206113fe0d07f04ed3989bc63134b01a64defa5994f9d2a195a0584fa6bf255e | 142572 |
| cmd/modelexplain/main.go | 10d8a8fc7f6e55490c8d250f24a94a04e9436cae | 605fe58c2fe0222e8d4dc0f47a44040a8f7caa9e881393d0d69928b2afcf7b09 | 969 |
| cmd/modelexplain/main_test.go | 674b87aa41d992d91d5ce5d236aa573837fe6ce2 | b89c0460095d874d5b8a94d81190390cd6aaf50e5451b0a20014f375c1d72526 | 2341 |
| scenarios/report_decode.go | 2886a78eaba56118c7c44899ad850431ffad155b | f14816a19faacf707e23b01d5d16b26be8e8c03a9f5f9af801029b89fb2ceddf | 3626 |
| scenarios/report_decode_test.go | f8121f03bdf0253596a10e42ac88e673e7d78e31 | 36ad86ba5ad3aa31e2af82be8dd55124fe35eb0c88e6f14712b13a16e4ab4013 | 1587 |


## Local completion and independent exact-source review

Reviewed source is 5427d0d40953ca9cbc3874c62cb4bf628ad8ae70, tree
af7acceaf2049db42eb8a5ba4a07ec60acde2a48. The original lead found no material
source, ownership, authority or compatibility defect. The independent integration
author separately verified its frozen expectations and appended contract claims.
The original six mapper/decoder/type/test/vector files remain byte-identical to
dfe14ea; only the prospectively named CLI main/test and contract intentionally
differ. Integration SHA-256 remains
884d4b7dfa0125b8e62315ec1b84fecbe61513c9bb9a9763ffe21446966f7962,
frozen at 1eb30ca before logic adoption at fc54c2a. No existing current-main path
was modified. Public main was rechecked at 713c2457 before completion.

Both author and original lead independently passed actual Go 1.27.3 offline full
race tests (40 tested packages), focused tests including real closed stdout, full
vet, formatting, modelc validate/check, modeltrace check, authoring schema 2/2 and
repository validation. Schema runs used an existing installed dependency closure
with a byte-matched lockfile, not a new install. Optional skills-ref is absent.
The integration author independently reran the three affected packages under race.

The CLI's real closed-stdout regression failed original main at fc54c2a with
SIGPIPE and passed the reviewed correction with exit 3 and fixed diagnostic.
Both author and lead executed the maintained modelplay-to-wrapper-to-modelexplain
recipe: the existing four-case/thirteen-step suite produced declared refusal links
and caller-supplied-assertions status. No explanation output was used to refresh
an expected value, and the receipt is not authenticated execution evidence.

The author also ran four isolated compilable mutants: accepting the wrong bound
report version, promoting status to execution-verified, dropping guard links and
collapsing nil refusals to empty. Unchanged frozen integration assertions rejected
all four for behavioral reasons, not compilation failure. These are author-run
controls, not a claim that the reviewer independently ran them.

Final full committed-diff Project Memory validation and completion blob/mode
continuity follow this metadata-only completion. A complete-history bundle will
preserve this fresh adoption history without held ancestors, old tasks or old
events. Publication remains a separate coordinator decision; no reviewpack,
process-view, private-consumer, timing or browser restriction is waived.
