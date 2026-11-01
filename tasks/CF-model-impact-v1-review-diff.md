---
id: CF-model-impact-v1
status: review
owner: Astra model-first lead
repositories: [spik3r/clanker-foundry]
depends_on: [CF-model-lifecycle-v1, CF-model-scenarios-v1]
---

# Read-only model change review

## Implementation claim

The owner requested continued model-driven Foundry development. This claim is
for a bounded LOCAL ONLY follow-up on exact scenario-runner commit
`33fc04e7b45cd9d7f696c3a935378f2a49990da4`, branch `codex/model-impact-v1`.
PR11 remains unchanged for user review. This stacked change is not independently
mergeable and does not authorize publication, merge, activation, or retry of any
previously denied publication. Only public-safe invented models are used.

Fresh open PR5/PR11 and active worker scopes were inspected. The runtime
receipt-time correction and unrelated consumer work remain separate. This claim
does not transfer their ownership or alter their gates.

Exact frontier under `modeling/lifecycle/`:
- `impact/README.md`
- `impact/types.go`, `impact/compare.go`, `impact/compare_test.go`
- `impact/decode.go`, `impact/decode_test.go`
- `cmd/modeldiff/main.go`, `cmd/modeldiff/main_test.go`
- `impactacceptance/oracle_test.go`

Also this task and its new immutable Project Memory source/events. No existing
model, compiler, reducer, views, scenario code, original oracle, package, workflow
or previous task/evidence changes. An independent Codex contributor owns only
the new acceptance file. Review and commit this prospective claim, register its
pinned source, and record a genuine start before implementation edits.

## Purpose and exclusions

Help a reviewer answer what changed in an accepted design and which declared
links should be inspected. Reuse existing format-neutral IR normalization and
canonical identity; do not add a second interpreter, regenerate code, run a
candidate model, infer equivalence or satisfiability, or rebind scenario/review
digests. This is a semantic-ID model review diff, not behavioral impact analysis.
No UML/BPMN/SysML/fUML conformance or real review/authentication claim follows.

## Typed comparison contract

`impact.Compare(before, after model.Model, bindings []Binding) (Report, error)`
normalizes and validates both complete models before producing any report.
Bindings are optional caller assertions `Binding{ID, ModelDigest}`: at most 128,
unique valid model-profile IDs, lowercase 64-hex digest. Nil typed bindings means
an empty list; binding order is a set and normalizes by ID. The tool does not read
scenario cases, bypass their existing compiled-model gate, or claim coverage.

Report includes exact before/after canonical model digests, a changed flag and
review-required flag equal to digest inequality. Editorial labels/requirement
text remain part of existing identity and therefore invalidate prior binding.
The flags do not approve anything when false. Report is example-only, authority
none, with explicit format/profile identity, no time or random IDs.

Changes identify kind (`model`, `state`, `requirement`, `guard`, `transition`),
semantic ID (empty only for the model root), type (`added`, `removed`, `modified`)
and lexically sorted changed field names. Added/removed elements use an empty
field list, not invented values. Modified fields are: model id/initial; state
label/terminal; requirement text; guard requirement/expr; transition from/trigger/
to/guards/effects. Identity renames are removals and additions, never guessed.
Changes sort lexically by kind then ID. Compare full normalized values; expression
operand/set reorder is ignored only where existing Normalize permits it. Effect
order remains semantic and unsupported effects still refuse.

Declared-link impact is computed separately for before and after graphs, sorted
before then after and by transition ID. A transition is listed if it changed,
an endpoint state changed, a referenced guard changed, or a referenced guard's
requirement changed. Model id/initial changes list every transition. Each entry
has every applicable sorted unique cause key `kind:ID` (root `model:`), not only
the first matching cause. Use both graphs so removed
links are not lost. This finite direct-link projection is not exhaustive dynamic
impact, reachability, execution, coverage, safety or equivalence analysis.

For each binding report exactly `matches-after`, `matches-before-only`, or
`matches-neither`. Equal digests yield matches-after, not an invalidated binding.
These comparisons apply only to supplied asserted metadata. No expectation or
review record is rewritten and no scenario is executed.

All returned collections own their copies; caller models/bindings are not mutated
or aliased. Pure concurrent use is safe with immutable inputs. Typed Report is
ordinary data, not authenticated evidence. There is no public encoder that claims
arbitrary caller-modified reports are validated observations.

## JSON adapter and CLI

`impact.Execute(raw []byte) ([]byte, error)` strictly reads a request with exactly
`formatVersion: lifecycle-model-diff-json/v1`, `before` and `after` as complete
existing JSON-v1 model documents, and explicit nonnull `bindings` array of exact
`id`, `modelDigest` objects. Reuse existing strict JSON and model adapters; reject
unknown/case-variant/duplicate/missing/null/trailing/invalid Unicode input. Both
embedded models retain the existing 1 MiB/depth64 format limits. Whole request
is capped at 3 MiB/depth66, including at most128 bindings. Typed models retain
existing semantic bounds, without forcing them through a JSON roundtrip.

Execute hashes exact original request bytes and exact embedded before/after JSON
value bytes separately from canonical identities. Typed Compare reports these
three observed-source digests as null. Different embedded source bytes with equal
canonical digests are formatting differences, not changed semantics. Outer JSON
whitespace affects request identity but not embedded or canonical identities.

Execute recomputes comparison and buffers one deterministic fixed-order JSON
report, capped at 1 MiB, before returning any bytes. It never encodes a separately
supplied Report. Report contains IDs/field names/cause links and digests, not full
model text; IDs may still be private. Errors use fixed categories without copying
input data. Source identity is consistency, not authenticity or binary provenance.

`modeldiff` accepts no arguments or file paths: bounded stdin to buffered stdout,
exact Go1.27.3, no writes/network/install/hooks. Exit0 is a valid comparison even
when changed or bindings mismatch; exit2 is rejected input/toolchain/report limit
with empty stdout; exit3 is failed/short output write and never success. Never
treat a valid report as approval to regenerate, execute or publish a candidate.

## Acceptance and completion

Independent expected changes, causes and binding statuses freeze before comparison
logic is implemented; a genuinely started type/DTO checkpoint may precede that
freeze. Expectations are written from this contract, not generated comparison
output. Cover no-op/normalization, requirement/editorial/expression/
effect/transition/initial/identity changes, removed and added links on both sides,
binding matches/staleness/unrelated, ordering and aliases. Add strict malformed
input/resource tests, deterministic bytes, input-read/output errors and no writes.
Kill deliberate omission mutants for requirement propagation and removed-side
links; unchanged expectations must fail rather than silently refresh.

Run full Go race tests/vet and existing modelc drift check with offline modules,
pack/whitespace validation, independent exact-source review and genuine completion
evidence. Record passed, failed and unavailable checks honestly. Preserve previous
branches and evidence; hand off local results and a verified private durable
artifact. Publication remains separately gated on parent review/coordination.

## Local review handoff

The additive typed comparison and stdin-only CLI are implemented locally. They
reuse existing normalized model identity, report field changes and all applicable
declared-link causes on both sides, and compare asserted binding digests without
reading, rebinding or executing scenarios. Existing model, compiler, reducer,
views, runner and frozen original oracles are unchanged.

Independent acceptance literals froze before comparison implementation in a
14,600-byte source prefix with SHA-256
`4497a7a693a62ac9f3d431f76660026310c12aa3274a3ae0d8bb3fd2bf753f85`.
Supplemental independent tests preserve that prefix byte-for-byte; complete test
SHA-256 is
`d2f80bc49dff04ecec935db963160f9488f78986e790083d68a453b192c31d60`.
The unchanged independent assertions pass against the real comparison. Scratch
mutants omitting requirement propagation or the entire before-side link list are
killed by behavioral assertion failures, not compiler/setup errors.

The README example edits only an in-memory requirement and reports the changed
requirement plus five linked transitions in each graph; no source model is
rewritten. Tests cover exact wire goldens, separate source/semantic identities,
strict input/resource limits, maximum valid graph/link counts, aliases, concurrent
immutable comparisons, deterministic process output and read/write failures.

Final exact-source checks and independent review are recorded with completion
evidence. Review status is LOCAL ONLY and remains dependent on PR11 and the local
scenario-runner branch. No push, new PR, merge, activation or denied-operation
retry occurred. A matching digest is not design approval or scenario compatibility.
