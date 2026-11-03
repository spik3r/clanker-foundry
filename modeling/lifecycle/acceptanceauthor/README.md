# Frozen scenario acceptance authoring

Use this offline workflow to record an author's intended behaviors alongside an
existing, hand-authored [scenario suite](testdata/suite.json). The
[manifest](testdata/manifest.json) names the case intent, declared author, exact
suite source SHA-256 and semantic model digest. The suite itself holds the literal
expected transition results; this tool never derives, repairs, or refreshes them
from a reducer or a previous run. The committed [expected outcomes](testdata/expected.json)
were written before the checker, with the success and refusal examples.

To author a new suite, write the expected results from the intended behavior before
running the reducer. Follow the [scenario fixture guide](../scenarios/README.md),
including its exact task, command, refusal and model-binding rules. Have the author
record the suite's exact raw SHA-256, the semantic digest of its intended model,
and a short intent for each case in a `lifecycle-acceptance-authoring/v1` manifest.
The `cases` IDs must match the suite's case IDs exactly. The author label is a
claim, not proof of independence. Obtain an independent human review when needed.
A digest does not authenticate a file or establish that the expectations are right.

Supply the original manifest, current model and suite as literal UTF-8 text inside
one JSON request with `formatVersion: lifecycle-acceptance-authoring-request/v1`
and `mode: check` or `mode: replay`. Each of `manifest`, `model`, and `suite` is a
JSON string containing the complete original file bytes, not an object, path, or
base64. This example constructs a request from the local sample files without
changing them (run from `modeling/lifecycle`):

```sh
python3 -c 'import json,pathlib,sys; p=pathlib.Path; print(json.dumps({"formatVersion":"lifecycle-acceptance-authoring-request/v1","mode":"check","manifest":p("acceptanceauthor/testdata/manifest.json").read_bytes().decode("utf-8"),"model":p("examples/lifecycle.json").read_bytes().decode("utf-8"),"suite":p("acceptanceauthor/testdata/suite.json").read_bytes().decode("utf-8")},ensure_ascii=True))' | GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelacceptance
```

The source hash covers the exact decoded suite bytes, including whitespace and the
final newline; a format-only suite edit is stale. The current model is decoded and
validated by the existing Go model adapter and identified by its canonical digest;
a format-only model edit does not change its identity. The suite model header must
match the manifest's model digest, and the supplied model must match the compiled
scenario reducer's model digest. A changed model or suite is not automatically
rebound: review the intended behavior and explicitly author a new manifest and, if
needed, new literal expectations. This example's runner supports only its compiled
model; a different model is reported stale at the runner binding even if its other
identities are current.

The report gives four separate `current`/`stale` bindings: manifest versus current
model, declared versus observed source bytes, suite header versus manifest model,
and current model versus compiled runner. Only when all four are current does the
checker fully decode and validate the suite with `scenarios.Decode`, verify the case
IDs and calculate the existing scenario digest. When any binding is stale, the
suite is `not-checked` and replay is `not-run`: only its strict JSON syntax and
model-digest header have been examined, so no claim of complete suite validity is
made. In `check` mode replay is also `not-run` even for a valid suite.

In `replay` mode, after validation, the checker calls the existing
`scenarios.Execute` and embeds its unmodified report. `matched` means that run
matched the supplied literals; `mismatched` includes an expected refusal with the
wrong result. An expected refusal that matches counts as a matched comparison,
not a successful state transition. A report, claimed author, or matching hash is
not proof of independence, review, approval, authenticated execution, safety or
production authority. Reports can contain the suite's task and command data; keep
requests and reports private when they contain private material.

The CLI uses bounded stdin and writes a buffered report to stdout. Exit 0 means
bindings and validation are current and, if replay was requested, its comparisons
matched. In check mode it does not mean a test ran. Exit 1 means stale bindings or
a replay mismatch, with a complete report; exit 2 means malformed input or a read
failure, with no report; exit 3 means a failed or short write, which can leave
partial output. The request is at most 8 MiB, the manifest 64 KiB, and the model
and suite each retain their existing 1 MiB limits. The manifest permits 1–16 unique
model-profile case IDs, nonblank case intents of at most 4096 bytes, an author of
at most 256 bytes and a purpose of at most 4096 bytes. All keys are required;
unknown or duplicate keys, invalid Unicode, invalid hashes and trailing JSON fail.

Run `go test ./acceptanceauthor ./cmd/modelacceptance` with Go 1.27.3 and offline
module resolution. The fixture author also wrote the checker and tests; the fixed
literals prevent deriving those outcomes from the implementation, but do not
constitute independent-person verification. This is a bounded lifecycle example,
not a general UML or SysML acceptance engine.

## Explicit generated-model runner (Go API only)

`CheckWithRunner(manifestBytes, modelBytes, suiteBytes, mode, runner)` accepts a
[scenario Runner](../scenarios/README.md) constructed explicitly in trusted Go
code. Supply the separately generated package's profile/model constants and a
field-only adapter to its `Transition` function. The callback owns all reducer
semantics; no model interpreter, request-selected provider or dynamic loading is
added. A zero runner refuses rather than selecting the example reducer.

The runner binding is a **caller declaration**. A digest and callback do not prove
that the callback was compiled from that model. Callback code may panic, block,
allocate or perform I/O; this API does not sandbox it, cancel it, bound the callback's
allocations or roll its effects back. Panics propagate. Independent review of the
composition and generated source remains separate from these consistency checks.

The original manifest, model and suite undergo the same input validation and
binding checks as the legacy API, with the explicit runner used consistently for
suite validation, digest, replay and report encoding. Current `check` inputs fully
validate the suite and exact case IDs and calculate its scenario digest, but invoke
no callback and report `valid`/`not-run`. Stale bindings retain `not-checked`/
`not-run` with no callback. Neither path derives or refreshes expectations.

The opt-in return type is `BoundReport`, version
`lifecycle-bound-acceptance-report/v1`. It uses `runnerModelDigest` and
`identitySource: caller-supplied-go-binding`, **not** `compiledModelDigest`.
Nested replay reports use `lifecycle-bound-scenario-report/v1` with the same
identity-source label. These identities establish neither binary provenance nor
execution authentication, authorship, approval or deployment authority.

A callback returning an invalid provider envelope or result violates the runner
protocol. `CheckWithRunner` preserves `scenarios.ErrProviderProtocol` and returns
a zero report, rather than calling it a frozen-expectation mismatch. Invocation
may already have happened: this error does not mean `not-run` or rollback.
Ordinary valid result mismatches still produce `mismatched` and stop before the
next callback. Invalid input returns no report; all suite steps and cases are
validated before the first invocation. Aggregate report overflow may occur after
multiple callbacks and returns `ErrInvalid` without a report. An arbitrary error
or absent report therefore does not prove zero execution; only the documented
preflight-refusal, check-only and stale paths guarantee zero calls.

Existing `Check`, `Execute`, `Report` and `modelacceptance` remain bound to the
built-in example and retain their existing JSON bytes and error behavior. The CLI
cannot select this new runner. Frozen legacy hashes test that boundary; the
[consumer integration](../consumerintegration/README.md) exercises a separately
generated reducer without copying or replacing the Foundry module.
