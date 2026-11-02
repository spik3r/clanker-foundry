# Typed model patch planner

Propose a bounded explicit candidate over a model accepted by existing profile
validation. Supplied content may be private; no provenance, public safety,
authenticated freshness or real review is inferred. The separate reviewpack
rendering denial remains in force and is not addressed.

APIs:

```go
Plan(before model.Model, patch Patch) (Result, error)
Execute(raw []byte) ([]byte, error)
```

The planner normalizes/deep-copies the before model, requires exact
canonical BeforeDigest, applies bounded explicit edits privately, then normalizes
and validates the whole final candidate. No intermediate graph validation prevents
coordinated explicit reference repair. Existing model semantics, canonical identity
and impact.Compare remain authoritative. No reducer, code generation, scenario
execution, expectation refresh, automatic reference repair or evidence rebinding is added.

## Operations and preconditions

Action is add, replace or remove. Kind is model, state, requirement, guard or
transition. Every add/replace carries exactly the matching complete record; remove
has no payload. Non-root record ID equals target ID. Root uses empty target ID,
Replace only, and complete Root{ID,Initial}; version/profile cannot change.
A root identity change is explicit, not inferred. Final global IDs/references
must satisfy lifecycle/v1 including the root ID's uniqueness.

Add requires no record of that kind/ID; replace/remove require an existing record
of that kind/ID. Global uniqueness is deferred until final validation, enabling
explicit cross-kind reuse and temporarily dangling references. Each (Kind,ID)
may appear once only, including root. Even add/remove cancellation conflicts;
use replace for a same-kind/same-ID update. Rename is remove old +add new, with
separate explicit reference edits. Operations remain ordered; semantic sets and
effect order use existing Normalize, with no second behavior engine.

Zero operations and identical replacements are legal no-ops. Empty changes or
unchanged digest do not grant approval. No missing target is silently ignored.
Any failure discards the private candidate and returns zero Result/no bytes.
The caller before model, patch payloads and returned result never share mutable
slices/pointers. Concurrent calls are safe with immutable inputs.

## Validation and deterministic error stages

1. Existing Normalize validates/copies before; existing JSON-v1 Encode must accept
   its representation/size. Upstream failures become ErrInvalid.
2. Structural/resource validation only: known action/kind, exact payload union,
   matching target/record IDs, all UTF-8/scalar bounds, model.Identifier grammar
   for model IDs/reference IDs/triggers, bounded counts. Bound all expression
   trees before copying/encoding, then reuse model.TypeOf and require bool for
   guard roots. Dotted expression field references use TypeOf's finite registry,
   not element-ID grammar. No graph existence, label/text blankness, guard-list
   duplicate policy or effect support/order/duplicate semantics are checked here.
   Bounded unknown effect names and temporarily dangling valid IDs pass shape.
   Those profile rules remain exclusively at stage6. Cyclic/deep Args reject
   within planner budgets.
3. Reject duplicate targets with ErrConflict.
4. Exact canonical before digest must match, otherwise ErrPrecondition.
5. Apply ordered target-existence preconditions, otherwise ErrConflict.
6. Normalize/validate final whole model and require JSON-v1 representability;
   upstream failures, including profile collection bounds, become ErrInvalid.
7. Compute digests and existing impact.Compare(before,after,nil), then bounded
   result output. Planner-owned budgets produce ErrLimit.

Stale plus blank label/text, duplicate guard references, unsupported/reordered
bounded effects or dangling valid-ID references returns ErrPrecondition. Stale
plus invalid ID/trigger grammar or invalid guard typing returns ErrInvalid.
Thus invalid expression typing beats stale digest; valid duplicate targets beat
stale digest; a well-shaped unique stale patch fails before target lookup. Errors
are fixed categories and never include caller text. No partial report is returned.

## Strict JSON carrier

One request has exactly formatVersion lifecycle-model-patch-json/v1,
patchVersion lifecycle-model-patch/v1, profileVersion lifecycle/v1, before (full
existing JSON-v1 model object), beforeDigest, operations (required nonnull array).
Unknown/case-variant/duplicate/missing keys, trailing input, malformed Unicode and
null in any other nonnullable position reject. No implicit defaults or macros.

Every operation has all eight keys in this order: action, kind, id, root, state,
requirement, guard, transition. Irrelevant payloads are explicitly null. Record
fields are complete, required and in this order:
- root: id, initial
- state: id, label, terminal
- requirement: id, text
- guard: id, requirement, expr
- transition: id, from, trigger, to, guards, effects
- effect: op

Every expression node is an explicit IR DTO: kind, ref, value, op, args. All fields
are mandatory; args is a nonnull array and unused scalar fields are empty strings.
This maps to model.Expr and existing TypeOf/Normalize. It is intentionally distinct
from the existing model JSON adapter's short-form expressions, whose standalone
parser is not exported. No model parser or operator semantics are copied here.
Typed nil leaf Args encodes as[]. Record lists are nonnull in the wire carrier.
The output contains reportVersion, profileVersion, mode, authority,
beforeDigest, afterDigest, after (complete existing JSON-v1 model) and review
(existing impact.Report, typed source hashes null, empty bindings). Execute
computes that result; no arbitrary-report encoder claims caller data is observed.

## Resource contract

Actual constants are in types.go. At most64 operations; labels256 bytes;
requirement/Expr literal4096; ref128; expression kind/op16; IDs48 ASCII bytes;
UTF-8 strings; guard references64; ordered effects3, each effect-op scalar64
bytes. Empty/blank labels and requirement text, duplicate guard IDs and unsupported
or reordered bounded effects are final-profile failures, not shape failures. Patch expressions together
have at most512 nodes, depth16, args16. Existing final model bounds are unchanged.
Intermediate collections can temporarily exceed final counts within64 edits.

Both typed and wire paths enforce a1MiB compact patch DTO budget. Its fixed fields
are patchVersion, profileVersion, beforeDigest, operations; include every nullable
payload slot and explicit leaf empty args. Exclude outer formatVersion and before.
Use compact Go JSON UTF-8, no final LF, no HTML escaping or Unicode normalization.
Before and after must pass existing1MiB JSON-v1 Encode bounds. Whole request and
result each at most3MiB; JSON request depth64. Planner budgets fail before recursive
copy/encoding, mutation or output, as applicable.

The CLI accepts no args or paths: bounded stdin, fully buffered stdout,
exit0 success, exit2 rejection/read failure, exit3 failed/short write. It writes no
files, extracts nothing, and runs no subprocess/network/install action.

## Independent acceptance and qualification

Six handwritten before/patch/after fixtures with literal changes/links and fixed
canonical digests preceded patch logic. Their original10726-byte oracle prefix
retains SHA-2565007eb7d96cdea215484eca05213950e617cbfa6582d6bad858ad2fa1890f68b.
Twenty-one mixed-error expectations were appended and checkpointed before logic;
labels/text blankness, guard-list/effect semantics and graph rules stay final-only.
No expected values are refreshed from Plan or Execute output.

Tests distinguish fixture self-checks from real planner/wire execution. They cover
coordinated edits, root/cross-kind updates, rollback, immutable owned results,
concurrent deterministic replay, typed cyclic/deep AST budgets, scalar/list/byte
limits, exact required/null/Unicode DTOs, effect ordering and CLI I/O failures.
Selected valid isolated mutants remove the stale check, leak a partial candidate,
or omit an edit; unchanged behavioral tests must reject each. No candidate model
behavior or scenario execution occurs.

Run actual Go1.27.3 offline from the module directory:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race ./...
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go vet ./...
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelc check
```

Use `go run ./cmd/modelpatch` with one strict request on stdin. A successful JSON
result is an inspectable proposed model and declared-link review, not permission
to regenerate code, rewrite evidence, execute a model or publish anything.
Final exact-source lead review and genuine completion evidence remain separate
project gates. A source checkpoint is not completion or publication authority.


## Current-main integration

This is an adoption of the existing local planner, not a new modeling engine or
an adoption of the held reviewpack/exporter stack. The task records the original
nine-file source identities. The planner, strict decoder and original oracle stay
byte-for-byte unchanged. The CLI now ignores SIGPIPE so a real closed stdout pipe
returns exit 3 and the fixed output-error diagnostic, as promised above. Its test
launches only the trusted test binary with a deadline and finite WaitDelay; this
is not a general process runner or a descendant-cleanup guarantee.

The independently authored `integration_test.go` spells a complete `T_prepare`
replacement against the frozen baseline digest. It adds the existing `G_actor`
guard reference and compares the whole normalized candidate to the existing
[tutorial candidate](../tutorial/model-v2.json), not to output from the planner.
The strict JSON request is independently spelled as well. Run it from the module:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 ./modelpatch ./cmd/modelpatch
```

Canonical model identity is distinct from source-byte identity: normalized output
has the tutorial candidate's canonical digest, not its raw file hash. The tests
keep both old and candidate manifests/suites unchanged. Existing
[acceptance authoring](../acceptanceauthor/README.md) check mode reports the old
model binding stale; the candidate manifest can match the proposed model while
the compiled baseline runner remains stale. Neither check rebinds evidence or
claims scenario validation/replay. Reapplying the original patch fails its old
precondition, with no candidate or output bytes.

For actual generated-code execution, use the existing
[tutorial](../tutorial/README.md) and its separate frozen expectations:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 ./tutorial
```

That test already builds and replays its candidate in a disposable copy. The
planner integration proves candidate identity and stale-evidence preservation;
it does not add a second orchestration command, execute caller-provided code or
certify intent, arbitrary adapters, provenance, UML or SysML conformance.
