---
id: CF-model-lifecycle-v1
status: in-progress
owner: Astra model-first lead
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Go-first model-driven lifecycle demo

## Implementation claim

The owner approved this isolated model-driven demo and coordinated independent
implementation, oracle and view/reproducibility workers. Astra leads one integration
branch `codex/model-lifecycle-v1` from main
`b4e5e75511bc271ed43ba2ca850ba2b9c230f0f2`. This is a manual producer claim,
not managed dispatch, consumer adoption or production lifecycle policy.

Main task records and open PR5 were inspected. The separate validation-evidence
correction owns existing state-engine files; this task touches none. The new module
and new CI path have no inspected competing claim. Recheck before publication.

Exact frontier under `modeling/lifecycle/`:
- `go.mod`, `README.md`, `toolchain.json`
- `model/ir.go`, `model/ir_test.go`
- `strictjson/decode.go`, `strictjson/decode_test.go`
- `jsonformat/json.go`, `jsonformat/json_test.go`
- `canonical/canonical.go`, `canonical/canonical_test.go`
- `compiler/generate.go`, `compiler/generate_test.go`
- `cmd/modelc/main.go`
- `examples/lifecycle.json`, `examples/oracle.json`
- `generated/lifecycle.go`, `generated/model_paths_test.go`
- `generated/decision.svg`, `generated/decision-table.json`, `generated/trace.json`
- `views/render.go`, `views/trace.go`, `views/views_test.go`
- `acceptance/oracle_test.go`

Also `.github/workflows/model-lifecycle-go.yml`, this task record and its new
immutable Project Memory source/event files only. No existing runtime, package,
lock, helper, workflow, claim or evidence is edited.

The independent oracle worker owns `examples/oracle.json` and
`acceptance/oracle_test.go`; it derives expectations from the written requirements,
not the compiler. The view worker owns the three `views/` files. Astra owns all
other source and coordinated generation/integration. Contributors use isolated
workspaces and transfer reviewed exact files; no competing branch updates. The
lead records the genuine prospective source/start and final integrated evidence.
No contributor begins maintained edits before that start is confirmed.

## Architecture

Use a new stdlib-only Go module, not a Node compiler generating Go and not paired
TS/Go business rules. JSON-v1 is only the first source format. Its strict parser
and serializer map to typed semantic IR; normalization, validation, canonical
meaning and generation depend on that IR, not JSON traversal. No speculative
second parser/plugin system. Future format adapters must preserve the reviewed
semantic profile or reject unsupported/lossy constructs.

Separate input-format version, semantic-profile version and generator version.
Normalize unordered ID collections and pure conjunction/disjunction operands;
preserve effect order and acceptance sequence. Include reviewed requirements and
labels in the canonical model digest: editorial changes deliberately invalidate
review too. Original source-byte identity is separate from canonical model identity.

The Go toolchain is pinned; normal validation/generation/check has no downloads,
installation or network access. Existing Node runtime and Project Memory remain
unchanged repository components, not dependencies of the demo's business logic.
This task is not a wholesale migration.

## Example behavioral assumptions

States: Draft, Ready, InProgress, InReview, Done. Transitions:
prepare Draft to Ready; start Ready to InProgress; requestReview InProgress to
InReview; approve InReview to Done; requestChanges InReview to InProgress.
Done is terminal. The model, not event-name special cases in the generator,
declares the guards and ordered effects.

Every successful transition requires current expected revision, nonblank outcome
and nonempty acceptance criteria. Ready may be unassigned. Start names a nonblank
actor, simulates assigning it only when owner is absent, otherwise requires actor
to match owner. RequestReview requires the named owner acting.

Approve/requestChanges require a nonblank reviewer distinct from owner, actor
equal to reviewer, exact decision approved/changesRequested, and review bound to
task ID, current InReview revision, canonical current-task content digest and
accepted canonical model digest. These supplied identities/decisions and hashes
provide consistency in an invented example, not authentication, actual review,
worker authority or operating permission.

Content digest covers exactly current task ID, state, revision, outcome, ordered
acceptance list and optional owner, excluding review to avoid self-reference.
Canonical bytes are a fixed-field-order JSON object with decimal-string uint64
revision, valid UTF-8, no Unicode normalization, no HTML escaping or trailing
newline; absent owner is null and nil acceptance normalizes to an empty array.
The compiled model digest is not accepted from an untrusted assertion.

Success applies ordered declared effects atomically and increments revision
exactly once. Refusal leaves input and returned task content unchanged; overflow
at MaxUint64 refuses before effects. Review replay after rework/re-review and
same-revision content changes fail. Output is example-only/authority-none.

## Closed semantic profile

Stable globally unique ASCII element IDs; finite typed field references; string
literals; same-type string/integer equality and inequality; nonempty/absent checks;
bounded nonempty all/any expressions. Missing operands fail eq and ne.
Root guards must be boolean. Guard IDs on a transition are an unordered conjunction;
failed-guard IDs are sorted, independent of source order.

Ordered effects are assign_owner_if_absent, set_state, increment_revision.
set_state uses the transition's sole To field, not a second target.
Each transition has exactly set_state then increment_revision, optionally preceded
by assignment. Terminal states have no outgoing transitions. No executable model
text, arbitrary fields/imports/effects or hidden first-match priority.

Strict input rejects duplicate/unknown fields, trailing JSON, invalid UTF-8 or
surrogates, unknown format/profile/operator/version, ambiguous state/trigger pairs,
duplicate IDs, unresolved/mistyped references and excessive bounds. Programmatic
IR entrypoints enforce the same semantic limits. Invalid input is distinct from a
valid request whose guards refuse. Empty/missing values and exact types are explicit.

## Derived artifacts and independent acceptance

Accepted model produces one pure Go transition package, UML-style inert SVG state
view, decision table, model-derived transition tests and trace manifest. Views
show guards and ordered effects. Trace binds stable semantic ID locations and
requirements to four fixed generated code/test/view/table digests; no self-hash.
This provisional profile is not UML/BPMN/SysML/UTP/fUML conformance.

Independent frozen oracle precedes compiler implementation. Test happy path,
assignment/matching owner, missing prerequisites, illegal/terminal transitions,
wrong actor/reviewer/decision, task/revision/model/content mismatches, rework replay,
overflow, invalid inputs and deep non-mutation. Model-derived tests are additional
coverage, not the sole correctness oracle. Kill deliberate code/guard mutations.

An intentional model change must cause independently predicted behavior and
view/table/code/trace changes. Repeated generation with the same accepted model
and tools is byte-identical; harmless source formatting/order changes preserve
canonical meaning. Check-only does not write, detects missing/modified/stale owned
outputs, rejects symlinks/collisions and preserves unrelated files.

## Verification and handoff

Run Go tests, race tests, independent oracle, generated transition tests, mutation
and regeneration checks, syntax/format, pack validation, existing runtime/lifecycle
regressions, links/whitespace and genuine full-diff Project Memory validation.
Record measured benchmarks without unsupported performance claims. New CI is
validation only, read-only repository permission, no credentials or deployment.
Local complete validation may proceed while hosted CI is unavailable; no required
gate is silently waived.

Commit/register this prospective source and start before implementation; preserve
real chronology and existing evidence. Review exact final integrated source
independently and aim for one draft PR, not many competing integration PRs.
Report denied actions exactly and stop them; no permission expansion, worker
activation, private consumer source/data, registry publication, deployment or
production replacement is part of this task.
