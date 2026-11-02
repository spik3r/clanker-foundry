---
id: CF-rehearsal-handoff-v1
status: in-progress
owner: Codex design-change handoff contributor
repositories: [spik3r/clanker-foundry]
depends_on: [CF-design-rehearsal-v1 merged, existing designchange and scenariodiff]
---

# Carry a fixed design-change rehearsal for offline review

## Prospective implementation claim

The coordinator cleared this Foundry-only lane after checking actual main
`28739a21b126dfc0a793622ce46593ec7559d379`, tree
`10dc22dbfcd2dbe09d41814a01a2696886dc0cc7`, visible PRs and assigned lanes.
The prior `CF-design-rehearsal-v1` was implemented by this contributor and merged
as #43; this explicitly scoped follow-on extends only its three maintained
workflow/test/documentation paths. The separate first-model-to-runner bridge
owner confirms portable evidence and verification are outside that lane. This
claim is not a shared lock; recheck current main, active claims and open PRs
before publication. HT-409 and its consumer, #20 and its denied visual artifact,
other runtime or production paths, live services and deployment are excluded.

Exact maintained scope, plus this task's own new immutable Project Memory files:

- `tasks/CF-rehearsal-handoff-v1-portable-evidence.md`
- `modeling/lifecycle/rehearsal/run.sh`
- `modeling/lifecycle/rehearsal/rehearsal_test.go`
- `modeling/lifecycle/rehearsal/README.md`
- `modeling/lifecycle/rehearsalhandoff/handoff.go`
- `modeling/lifecycle/rehearsalhandoff/handoff_test.go`
- `modeling/lifecycle/rehearsalhandoff/README.md`
- `modeling/lifecycle/rehearsalhandoff/testdata/report-identities.json`
- `modeling/lifecycle/cmd/modelrehearsalhandoff/main.go`
- `modeling/lifecycle/cmd/modelrehearsalhandoff/main_test.go`

Commit and register this public-safe claim, then record a genuine start with the
individual exact paths before implementation. A separate contributor freezes
predicted complete report-byte identities and literal negative-control contract
from the already frozen suites and documented report wire order before any
handoff implementation or replay; never derive them from generated output. A
separate exact-source reviewer challenges bounds, trust boundaries and controls.
Separate agents do not attest independent-person or human review.

## Fixed-fixture handoff boundary

Add an opt-in `--handoff` mode to the existing fixed-fixture demo. No-argument
stdout bytes and behavior must remain identical to #43. Opt-in mode runs the
same actual baseline and candidate replay, then builds one deterministic bounded
JSON carrier and publishes it only after successful private-directory cleanup.
It carries exact original patch, baseline and candidate suites, verified
existing design-change archive, baseline/candidate report bytes, and generated
Go reducer/structural-test bytes. It carries no generated SVG, diagram, private
consumer input or filesystem path. A caller can redirect stdout to a private
file; a short write may leave a prefix and is not success. No arbitrary model,
path or executable interface, live network, worker, approval or deployment.

The separate pure Go Build/Verify boundary accepts a strict canonical carrier
and a **separately supplied** references object: exact source/model identities,
baseline suite identity and independently frozen baseline/candidate report byte
identities. References are caller assertions and must never be silently taken
from the carrier as authority. Check canonical lowercase SHA-256 syntax, strict
JSON/known fields, canonical padded base64, finite sizes and total budgets before
bulk decode. Use existing `designchange.Verify` and `artifactbundle.Verify`
for the nested archive and qualification inventory; match carried request,
candidate suite, generated Go and tests against verified members. Reuse existing
`scenariodiff` validation to require baseline/candidate suites bind to the
actual verified before/candidate models; reject stale suites even if a caller
updates their byte-hash reference. The two report blobs are matched only to the
separate report hashes and are **not parsed, semantically validated or proven to
have executed** by this verifier. Return an explicit `caller-supplied-records`,
`reportBytesStatus: hash-matched-only-not-inspected`, `authority: none` receipt;
no provenance, human review, liveness, code execution or safety attestation.
Hashes and report content may be sensitive. The actual demo independently
compares its reports to previously frozen literal scenarios before packing.

Provide a stdin-only Go CLI with explicit `pack` and `verify` operations, no
filesystem paths or arbitrary process execution. Build/Verify must reject invalid,
unknown, missing, duplicate, case-variant, noncanonical, oversized, stale or
tampered data without partial results; deterministic canonical bytes and input/
output ownership are required. Existing designchange/compiler/scenario semantics,
archive profiles, dependencies, generated source and other claims stay intact.
No second evaluator or duplicated guard/transition logic. Limits are a narrow
fixed-demo profile, not a promise to accept all otherwise-valid 8 MiB reports.

## Acceptance and handoff

Independently freeze complete report identities and negative controls before
implementation; include malformed JSON/base64, metadata substitutions, stale
suite with updated byte reference, changed archive/code, changed report and
changed external reference, zero/empty, size/depth limits and nonmutation. Prove
normal mode remains byte-for-byte identical, two handoff runs return identical
carrier bytes, no stdout on failure before successful cleanup, and restored
candidate replay remains real. Include compiling/behavioral negative controls,
not compile failures counted as detection. Run Go 1.27.3 focused/full race, vet,
modelc/modeltrace, repository validator, whitespace and committed Project Memory
validation; report optional tools and hosted CI separately. Obtain separate
exact-source review and preserve genuine chronology. A successful verifier
receipt is consistency only; publication and merge follow existing authority,
not from this task or the receipt.
