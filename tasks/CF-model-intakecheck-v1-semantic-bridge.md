---
id: CF-model-intakecheck-v1
status: in-progress
owner: Byte intake-model integration worker
repositories: [spik3r/clanker-foundry]
depends_on: [CF-model-intake-v1, CF-model-authoring-v1]
---

# Connect design intake to the accepted model profile

## Implementation claim

The coordinator delegated a standalone integration improvement on current main
without importing the held model-authoring/export/reviewpack stack. The exact
base is c49462d3d14f13647927e96a78048510bede379f, tree
0583fece5616848a4144bfbfa2ce2eab4e492778. This new branch is
codex/model-intakecheck-v1. Main task records and open PRs were inspected.
The coordinator confirmed the concurrent acceptance author owns acceptanceauthor/
and cmd/modelacceptance/ (PR18); those and all scenario/acceptance paths remain
reserved. PR15 traceability and held source-closure/dossier/reviewpack work remain
separate. This task does not adopt or complete them.

Exact new frontier under modeling/lifecycle:
- intakemodel/types.go
- intakemodel/bridge.go
- intakemodel/bridge_test.go
- intakemodel/README.md
- intakemodel/testdata/expected.json
- cmd/modelintakecheck/main.go
- cmd/modelintakecheck/main_test.go
Plus this task record and new immutable own Project Memory source/events only.
No edits to intake, model, jsonformat, canonical, authoring schema, compiler,
generated artifacts, scenarios, acceptance, existing CLIs, workflows or runtime.

## Bounded contract

Reuse intake.Execute on the original complete request bytes, without recreating
its validation or changing its semantics. After successful native validation,
extract only observed.model from that accepted request. The other artifacts stay
opaque. Preserve the exact native report bytes as intakeReportJSON with their
SHA-256, including unanswered questions and mismatched declarations. The outer
JSON carrier escapes this string; hashing applies to its decoded UTF-8 bytes.

Use existing jsonformat.Decode for strict model parsing, semantic validation and
normalization. Never use a schema pass or intake hash match as model acceptance.
Use canonical.Digest on accepted IR and jsonformat.Encode for the normalized
model artifact. Distinguish three exact identities: observed source bytes,
canonical model meaning, and normalized encoded model bytes. Do not overwrite or
silently replace intake observations, declarations or asserted dependencies.

Inspect(raw []byte) (*Report,error) and Execute(raw []byte) ([]byte,error) are pure
Go APIs. They use no filesystem, network, clocks, processes or generated-code
execution. No arbitrary path is accepted. Inputs must not be concurrently mutated;
results own their bytes. Actual Go 1.27.3 is required; no toolchain download.

Report shape:
- formatVersion=lifecycle-intake-model-report/v1, mode=example-only, authority=none.
- intakeReportSHA256 and intakeReportJSON: the exact native report artifact.
- schema: path=authoring/lifecycle-v1.schema.json,
  id=https://clanker-foundry.invalid/schemas/lifecycle-json-v1.schema.json,
  sourceSHA256=dc1f75f1ce24b71cb034c0b8b3af62a70d0bfa297e7bbc7b5c5337eb7336a5b7,
  validation=not-run. This is a pinned authoring reference, not a network endpoint,
  schema execution, semantic profile selection or a current-filesystem observation.
- model: status=missing|rejected|accepted, sourceSHA256, canonicalModelSHA256,
  normalized, problem. The four latter fields are nullable as described below.
- executionStatus=not-run, acceptanceStatus=not-evaluated,
  reviewStatus=not-established. No aggregate ready/qualified/approved flag.

Missing observation: status missing and all four model fields null. Rejected model:
actual source hash, canonical/normalized null, problem=model-validation-failed.
Accepted model: actual source and canonical hashes, normalized object containing
sourceSHA256 and sourceJSON (the exact normalized encoded UTF-8 bytes), problem
null. An accepted model may still be undeclared, mismatched, lack requirements or
have unanswered questions in the native intake report. It is not reviewed intent,
correct behavior, independently accepted code, or an execution authorization.

Input remains the existing intake JSON v1 shape, not a new nested request format.
Native intake strictness, null/key discipline and 4 MiB source / 1 MiB observation
bounds apply. Output is capped at 16 MiB after JSON escaping; errors return no
report. Schema and model format/profile identifiers are not used to fetch data.

Order: input gross limit, exact toolchain, native intake validation, accepted
request extraction, optional model decode, canonical identity/format encoding,
report size. Invalid intake is ErrInvalid; oversized raw input/output ErrLimit;
wrong toolchain ErrToolchain; an unexpected accepted-model canonical/encoding
failure ErrArtifact (no report, not a claim the decoded model is invalid).
An ordinary malformed or semantically invalid observed model is a valid rejected
report, not a fatal intake error. Diagnostics are fixed, without caller contents.

CLI modelintakecheck has no arguments; bounded stdin and one complete buffered
JSON stdout report. Exit 0 means report emitted, including missing/rejected model;
exit 2 input/toolchain/artifact/read refusal with empty stdout; exit 3 failed or
short write, whose partial bytes must not be accepted as a successful report.
Default JSON HTML/control escaping is retained. Reports include normalized model
contents and native intent/author labels, so they can be private.

## Independent acceptance and verification

An independent worker writes literal vectors before bridge logic, using only the
already-merged intake and model components. Freeze matching/mismatching/undeclared
valid models, invalid semantics despite matched raw hash, missing observation,
malformed model text, opaque requirements/suite data and unanswered questions.
No expected value may be refreshed from the new bridge output. Add adversarial
strict-transport, model-boundary, byte-preservation (including CRLF), whitespace
and normalization identity, escaping, limits, no-mutation, concurrency and CLI
failure checks. Verify the schema reference hash/ID against the unchanged tracked
asset so later schema edits require deliberate bridge-reference review.

Run pinned offline Go full race/vet/modelc checks, relevant schema contract tests,
repository validation and full committed-diff Project Memory. Obtain independent
exact-source review. Publication, merge, deployment and worker activation remain
separate authorized actions. No held ancestor is included or gate bypassed.
