# Custom model assertion inventory

This offline tool inventories supplied **matched assertions** for an exact custom
model and scenario suite. The older [inventory](../scenariocoverage/README.md) remains
bound to the built-in model. This new package accepts other native-valid models
without modifying that interface or any native execution policy.

An inventory is not behavioral coverage. Empty buckets mean the supplied matched
entries did not assert that category. They do not show unreachable behavior,
untested behavior elsewhere, correct expectations, satisfied requirements or
human approval. A fabricated but internally consistent report can be accepted.
The tool invokes no reducer, compiler, process, network or filesystem operation.

## Request and validation

`Check([]byte) ([]byte,error)` accepts exactly these required non-null fields:

- `formatVersion`: `lifecycle-bound-assertion-inventory-request/v1`.
- `model`: original native JSON model text, encoded as a string.
- `suite`: original native scenario-suite JSON text, encoded as a string.
- `report`: original native scenario-report JSON text, encoded as a string.

Native model decoding and canonical identity determine a validation-only runner.
Its callback panics if invoked; this package never invokes it. Native suite and
report decoders check structural/result envelopes. `ExplainWithRunner` validates
the entire report, including actual transition/guard declarations at a mismatch,
before any mapping. The report must bind the exact suite byte SHA256 in a nonnull
`sourceDigest`, and its scenario and model identities must match those inputs.

A narrow additional pass binds ordered entries to the suite's exact case/step
IDs, commands, expected literals and input flow. Only a complete match or a valid
fail-fast prefix ending at its first mismatch is accepted. Matching refusals
retain input, including an actual null task. Changed literals, reordered/omitted/
extra entries, invalid flow and wrong completed-case counts refuse the whole
inventory. The original expectation bytes are never refreshed from observations.
An undeclared expected transition in a valid mismatch remains an expectation,
not a declared actual link; native behavior is preserved and the entry is excluded.

Unknown, missing, duplicate and case-variant keys, malformed Unicode, null sources
and trailing JSON refuse. Keep input bytes immutable during the call. Calls return
owned output and are deterministic for unchanged inputs.

## Output

The format is `lifecycle-bound-assertion-inventory/v1`, with `mode: example-only`,
`authority: none`, `status: caller-supplied-assertions`, and execution/review status
`not-established`. `inputs` records exact model/suite/report SHA256 identities,
canonical `modelDigest` and semantic `scenarioDigest`. Formatting can change a
source SHA256 without changing its normalized identity. Hashes do not authenticate
where a report came from, which code ran or who reviewed its expectations.

`transitions` lists every normalized transition in sorted ID order. Each lists
sorted associated `guards`, even if no entry names them. Only matched entries
contribute `successes`, explicit guard `refusals` or `revisionLimit` references.
Guard refusal references say nothing about the other guards. Matched fixed
`INVALID_INPUT` and `ILLEGAL_TRANSITION` assertions have separate lists. A
mismatched entry appears only in `excluded`. References preserve report order
and carry `caseId`, `stepId` and zero-based `entryPosition`. Empty collections
remain arrays. `inputMatched`, `casesCompleted` and `stepsExecuted` preserve the
supplied report's validated observations; there is no aggregate coverage score.

Keep original inputs beside the inventory to review actual literals and any
unobserved steps after a mismatch. Assess expectation quality and report
provenance separately. This tool establishes neither execution nor approval.

## Limits and command

Input is capped at 16 MiB, with model and suite each at 1 MiB and report at 8 MiB.
Output is buffered and capped at 8 MiB. Native stricter depth/count/semantic bounds
still apply. Explicit byte/output bounds return `ErrLimit`; native decoding,
structural/link/literal/flow refusals return `ErrInvalid`. After native acceptance,
a source/scenario assertion mismatch returns `ErrBinding`. Native model-binding
refusals remain `ErrInvalid`. Unexpected native/encoding failures are `ErrBoundary`.
Errors return no partial bytes and fixed diagnostics without source payloads.

Build the standalone command from a trusted exclusive `modeling/lifecycle`
checkout. Select a new private absolute work directory outside the checkout and
place your reviewed original-text request there. Use installed Go 1.27.3 offline:

```sh
set -eu
: "${INVENTORY_WORK:?set a new private absolute output directory}"
mkdir -m 700 "$INVENTORY_WORK"
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
test "$(go env GOVERSION)" = go1.27.3
go build -o "$INVENTORY_WORK/modelboundinventory" ./cmd/modelboundinventory
# Place the unchanged reviewed request at "$INVENTORY_WORK/request.json".
"$INVENTORY_WORK/modelboundinventory" < "$INVENTORY_WORK/request.json" > "$INVENTORY_WORK/inventory.json"
```

The command accepts no arguments or paths. Exit 0 means a complete inventory,
including a valid mismatched prefix; it does not mean passing replay. Exit 2 is
an input/binding/limit refusal, 3 a short/failed stdout delivery, and 4 an unexpected
boundary or stdin read failure. SIGPIPE is ignored so a closed stdout exits 3.
Delivery can expose a partial prefix: inspect command status before using output.
IDs and hashes may still be private. Retain and dispose of inputs/results under
your own policy. This command does not add a foundry route or execute a candidate.

Three independent fabricated observation fixtures were committed before the
implementation. They include a custom model, explicit multi-guard refusals, fixed
refusals, an unused transition and two mismatch prefixes. They are consistency
examples, not executed evidence. Tests retain those literals and add strict,
binding/flow, concurrency, owned-buffer, output-bound and actual CLI controls.
