# Scenario assertion inventory

This offline, example-only tool inventories **caller-supplied assertions**, not
actual test coverage. It groups a scenario report's matched assertions against
all declared transitions and their guard associations. A valid report can be
fabricated. Neither a matching comparison, a digest, an empty list, nor this
inventory proves that a reducer ran, a guard was evaluated, a requirement was
satisfied, a reviewer approved anything, or the tested behavior is complete.
There is no authority, scheduling, dynamic instrumentation, guard evaluation or
standards-conformance claim.

## Boundary and output

`Analyze(model.Model, scenarios.Report) (Inventory, error)` first uses the
existing `scenarioexplain.Explain` boundary: the model is normalized, its
digest must match the compiled default model and report, and the entire report
and each actual result must pass existing consistency and declaration checks.
The analyzer does not select or invoke a callback or reducer. This first version
is deliberately tied to the built-in model and existing report format; it is not
the separate custom-model-bound scenario runner. A digest binds declarations
for consistency, not provenance or binary identity. An error returns no partial
inventory. Concurrent mutation of inputs is unsupported.

Each transition appears once in sorted model ID order, including those without
eligible assertions. Each guard association appears in sorted guard ID order.
References retain report entry order and contain case and step IDs. Arrays are
always present, even when empty. For **matched** entries only:

- `successes` groups assertions that claim the transition succeeded. This does
  not independently establish that every guard was evaluated or satisfied.
- `refusals` under each transition/guard pair groups only explicitly named
  guard reasons. Other guards are not inferred to have passed or failed. A
  multi-guard refusal is referenced under each listed guard.
- `revisionLimit` groups transition-specific fixed refusal assertions.
- `invalidInput` and `illegalTransition` group fixed refusals without a claimed
  transition. These are supplied classifications, not diagnoses.

A mismatched entry is listed only in `excluded` and never contributes to these
buckets, even if it carries a known actual transition or refusal. Report
execution is fail-fast, so a mismatch may also mean later steps were omitted.
`inputMatched`, `casesCompleted`, `stepsExecuted` and digest fields preserve the
supplied report's internally consistent claims. Empty lists mean only that the
supplied eligible entries did not assert that category; they do not show that
behavior is unreachable or untested elsewhere. No aggregate percentage is
calculated. Output owns its nested slices and digest pointer.

## JSON and command

`Execute([]byte)` accepts one strict wrapper, all fields required and non-null:

```json
{"formatVersion":"lifecycle-scenario-assertion-inventory-request/v1","model":{},"report":{}}
```

Replace the placeholders with a valid existing JSON-v1 model and scenario report
object. Strict decoding rejects duplicate, missing, unknown and case-variant
keys, trailing values, bad Unicode, invalid reports and incompatible models.
Input is capped at 10 MiB and depth 66, with the existing 1 MiB model and 8 MiB
report bounds. Complete output is buffered and capped at 4 MiB. No input content
is echoed in diagnostics. The returned error does not authenticate input or
explain why a supplied reducer reported a refusal.

From the module directory, with Go 1.27.3 installed, an invented local example is:

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
# The following files may contain sensitive data; choose a private directory.
go run ./cmd/modelplay < scenarios/testdata/lifecycle-v1.json > /tmp/scenario-report.json
python3 -c 'import json; print(json.dumps({"formatVersion":"lifecycle-scenario-assertion-inventory-request/v1","model":json.load(open("examples/lifecycle.json")),"report":json.load(open("/tmp/scenario-report.json"))}, ensure_ascii=False))' | go run ./cmd/modelcoverage
```

The command accepts no arguments or paths: it reads bounded stdin and emits one
complete JSON inventory on stdout. It returns 0 on success, 2 on argument,
toolchain, input or validation failure, and 3 on failed/short output. Buffering
does not make a stream write atomic: a failing writer can expose a prefix. The
program itself does no file, network, subprocess, provider, or worker operations.
The example shell commands explicitly perform local file operations. IDs and
digests may be sensitive; keep reports and inventory private and remove temporary
files when appropriate.

## Validation

The fixture expectations in `coverage_test.go` were hand-transcribed from the
existing separately authored scenario fixture before this mapping was written.
The new inventory expectations were not independently authored. Tests also
exercise fabricated yet consistent reports, fixed refusals, mismatch exclusion,
invalid mappings, strict input, deterministic bytes and output ownership. These
checks validate this bounded mapping, not the correctness or completeness of
underlying scenario expectations. Run with the pinned local toolchain:

```sh
go test -race -count=1 ./scenariocoverage ./cmd/modelcoverage
```
