# Go-first model-driven lifecycle example

This is an isolated design-to-code demo, not a scheduler or production policy.
An agent can draft the [model](examples/lifecycle.json); a reviewer accepts its
meaning before the Go toolchain derives code and views. AI drafting is not claimed
to be deterministic. Accepted model + pinned generator/toolchain is reproducible.

## Run

Use **Go 1.27.3**, from this module directory, with a writable local Go cache:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelc validate
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelc generate
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelc check
go test ./...
go test -race ./...
go test -bench=. -benchmem ./compiler
```

There are no third-party module dependencies. Normal tool operation does not
install, download, fetch or execute model-provided code. `validate` and `check`
never write outputs. Commands accept no arbitrary input/output paths; the source
is `examples/lifecycle.json` and the five outputs are under `generated/`.

## Architecture and versions

`jsonformat` decodes/encodes JSON-v1. `model` owns format-independent typed IR,
normalization and semantic checks. `canonical` identifies that normalized IR.
`compiler` emits a pure Go package and structural tests. `views` projects the same
IR into an SVG, decision table and trace. A future format adapter must map to this
same semantic profile without silently dropping unsupported constructs; no second
parser or plugin system is included.

- Input format: `lifecycle-json/v1`.
- Semantic profile: `lifecycle/v1`.
- Generator: `lifecycle-go/v1`; toolchain: `go1.27.3`.
- The original source SHA-256 identifies exact input bytes. The canonical model
  SHA-256 identifies normalized accepted IR, including requirements and labels.
  Even editorial changes to those reviewed fields invalidate model review.
- Source whitespace, object-key order, ID-collection order and pure `all`/`any`
  operand order do not change normalized meaning. Effect order and acceptance
  sequence are preserved. Canonical bytes use sorted JSON map keys and normalized
  IR arrays, valid UTF-8, no HTML escaping and no trailing newline.
- Trace records model/build identities and hashes of four generated artifacts,
  excluding itself. Semantic-ID paths are not fabricated source line numbers.
  Neither hashes nor identity labels prove authentication or executed review.

Existing Foundry Node orchestration and Project Memory stay unchanged. Node is
not a dependency of this Go demo's semantics or generated reducer. This is not a
runtime migration, consumer integration or a parallel TS/Go implementation.

## Example model, not universal policy

```text
Draft --prepare--> Ready --start--> InProgress --requestReview--> InReview
                                      ^                         |     |
                                      +----requestChanges-------+     +--approve--> Done
```

Done is terminal. All accepted transitions require current expected revision,
nonblank outcome and nonempty acceptance. Ready may have no owner. Start explicitly
models simulated owner assignment to its named actor only if owner is absent;
otherwise actor must match. RequestReview requires the named owner acting.
Approve/requestChanges require a distinct reviewer matching the actor, the correct
`approved`/`changesRequested` decision, and task ID, current InReview revision,
current task-content digest and compiled model digest matching the review.
Rework/re-review advances revision and invalidates old approval.

Every effect is declared: optional `assign_owner_if_absent`, then `set_state`
(using the transition's sole `To` target), then exactly one `increment_revision`.
There is no trigger-name policy hidden in the reducer. All effects occur on a new
copy only after guards and overflow checks pass. No real owner claim, worker
assignment, actual approval or execution authority follows from this example.

Guards have stable requirement links and use only allowlisted typed references,
string literals, `eq`, `ne`, `nonempty`, `absent`, `all` and `any`.

| Guard | Meaning |
|---|---|
| G_revision | task revision equals command expected revision |
| G_outcome / G_acceptance | outcome / criteria supplied |
| G_actor / G_ownerCompatible | Start actor nonblank; owner absent or matches actor |
| G_owner / G_ownerActor | owner nonblank / equals acting owner |
| G_reviewer / G_independent | reviewer nonblank / differs from owner |
| G_reviewTask / G_reviewRevision | exact task ID / current revision |
| G_reviewModel / G_reviewContent | compiled model / current content digest |
| G_reviewActor | reviewer equals command actor |
| G_approved / G_changesRequested | exact modeled review decision |

Missing operands make both equality and inequality false. `absent` is permitted
only for optional references and means nil, not an empty present string.
`nonempty` strings use Go `strings.TrimSpace` (Unicode whitespace); lists use
length. Guard lists are conjunctions; all failed guard IDs are returned sorted.
`all`/`any` contain 1–16 pure boolean operands; empty/duplicate operands refuse.
Equality compares strings or uint64 references, never coerces types or compares
lists. Structural reachability does not prove satisfiability or eventual completion.

## Typed runtime API and invalid input

The generated package exports `Transition(Task, Command) Result`, `TaskDigest`
and `ModelDigest`. It accepts typed Go values, **not raw JSON requests**. An omitted
Go uint64 field is its real zero value; there is no claim to detect omitted struct
fields or unknown JSON fields at that boundary. Strict duplicate/unknown/missing
JSON checks apply to the source-model adapter.

Task revisions are uint64. Owner is `*string`: nil is unassigned; a present empty
or whitespace owner is not coerced to nil. Outcome/actor/reviewer can be blank
well-typed values and then fail relevant guards. Acceptance is ordered strings;
each supplied item must be nonblank, valid UTF-8 and bounded. Duplicates are not
silently forbidden. Nil acceptance and an empty list both fail its nonempty guard.

Input bounds: task ID is nonblank and at most 256 UTF-8 bytes; owner/actor/reviewer
and review task ID at most 256; outcome 4096; at most 32 criteria of at most 1024
bytes each. State/trigger must be known. Supplied review decisions must be the two
supported values; digests must be lowercase 64-hex. Invalid UTF-8 refuses.

Error precedence:
1. Invalid typed input: `INVALID_INPUT`, `Task=nil`, no mutation.
2. Known trigger unavailable from current state: `ILLEGAL_TRANSITION`, unchanged
   deep task copy.
3. Failed modeled guards: all IDs sorted, unchanged deep task copy.
4. Otherwise revision MaxUint64: `REVISION_LIMIT`, unchanged deep copy.
5. Otherwise apply ordered effects; revision increases exactly once.

All results say `example-only` and `authority=none`. Actor/reviewer strings and
review decisions are supplied assertions, not authenticated identities/evidence.

`TaskDigest` covers exactly a fixed-order JSON object: `id`, `state`, `revision`,
`outcome`, `acceptance`, `owner`. Revision is a decimal string, nil acceptance is
`[]`, nil owner is `null`; strings are preserved without Unicode normalization.
Encoding uses Go JSON with HTML escaping off, JSON escapes for U+2028/U+2029, and
no trailing newline. No review field is included, so no self-reference exists.
Binding detects changed content at the same revision, not forged approval.

## Refusals, generation and tests

JSON-v1 input is capped at 1 MiB / depth 64, with duplicate-key, unknown/case-folded
field, missing/null required field, trailing data and invalid UTF-8/surrogate
refusal. IR limits include 16 states, 64 requirements/guards, 32 transitions,
512 expression nodes/depth 16 and finite typed refs. Programmatic IR gets the same
semantic checks. A semantically valid IR too large for JSON-v1 encoding is refused
by that format encoder, not returned as unreadable output.

Generated files: [Go reducer](generated/lifecycle.go), [state view](generated/decision.svg),
[decision table](generated/decision-table.json), [structural tests](generated/model_paths_test.go)
and [trace](generated/trace.json). SVG is inert and generated from the IR. It is
UML-style, not a claim of UML, BPMN, SysML, UTP or fUML conformance. Standards
selection/mapping and executable-standard semantics need separate review.

Derived edge properties can be vacuous when guards refuse. They are not a witness
solver or an independent correctness oracle. The independently authored
[oracle](acceptance/oracle_test.go) covers actual accepted/rejected example paths;
mutation tests must show it rejects always-refuse and removed guards. The fixture
[case inventory](examples/oracle.json) may list additional planned cases; only
executed Go assertions count as passes. A separate intentional model-change test
predicts changed behavior and all artifacts without rewriting the frozen oracle.

Use an exclusive trusted worktree/single writer. `os.Root` limits path traversal,
but this CLI is **not** the separately reviewed immutable invocation boundary and
makes no security claim against a hostile concurrent host writer. Static symlinks,
nonregular files, colliding/partial/modified outputs and stale owned files refuse.
Unrelated files are preserved. Generate updates only a complete matching prior
manifest or an empty owned frontier; corrupted/partial outputs require explicit
manual resolution. Files are staged, with trace installed last. An interruption
may leave a detectable partial update and is never reported as success.

Independent exact-source review, local full validation and the genuine
[task/lifecycle](../../tasks/CF-model-lifecycle-v1-vertical-slice.md) remain required
before final handoff. No production activation, migration or deployment occurs.
