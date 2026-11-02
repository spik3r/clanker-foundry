# From design intake to a validated model artifact

`modelintakecheck` connects the existing [intake record](../intake/README.md)
to the existing strict [model adapter](../jsonformat/json.go). It preserves the
native intake report and, when the observed model is valid, emits a normalized
model artifact with separate source and semantic identities.

This closes a specific gap: matching a declared file hash does not make that
file a valid model. Conversely, a valid model can still have mismatched identity
declarations or unresolved design questions. Both facts remain visible; no
aggregate ready, qualified or approved status is invented.

## Run

From the module root with installed Go **1.27.3**:

```sh
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
go build -o /tmp/modelintakecheck ./cmd/modelintakecheck
/tmp/modelintakecheck < intake-request.json > intake-model-report.json
go test -race -count=1 ./intakemodel ./cmd/modelintakecheck
```

The input is exactly the existing `lifecycle-intake-json/v1` request, not a second
intake schema. All native intake validation runs first. The command accepts no
arguments or paths and performs no filesystem operations beyond stdin/stdout.
The caller's shell owns any redirection. No dependencies or toolchains are
downloaded by the command.

Requirements and scenario-suite observations remain **opaque text**. This command
does not author acceptance cases, inspect suite headers, detect stale suites,
replay scenarios, generate code, create archives or control workers. The held
review-pack/export/dossier stack is not a dependency.

## Model check and artifact identities

The report contains:

- `intakeReportJSON` and `intakeReportSHA256`: the exact native intake report
  bytes as a JSON string and their SHA-256, including the final newline. Decisions,
  questions, author labels, declaration mismatches and dependency assertions are
  unchanged. Decode this string to inspect the native report; hashing the outer
  escaped representation would identify different bytes.
- `model.status`: `missing`, `rejected` or `accepted` for the **observed model**.
- `model.sourceSHA256`: the exact observed UTF-8 bytes, when supplied.
- `model.canonicalModelSHA256`: accepted normalized semantic IR, when valid.
- `model.normalized`: `sourceJSON` from the existing JSON format encoder and its
  `sourceSHA256`, when valid. This artifact may differ from the original source
  bytes. It does not replace any original observation or declaration.
- `schema`: the local authoring-schema path, identifier and pinned source hash,
  with `validation=not-run`.

A missing observation yields null identities/artifact/problem. An empty string is
present, not missing: it is rejected. A rejected model retains its observed source
hash, null canonical/artifact fields and fixed `problem=model-validation-failed`.
A valid model yields the canonical and normalized identities with a null problem.

`accepted` means the existing Go adapter accepted this observed model's syntax and
finite semantic profile. It does not mean the intended system is correct, its
declared source matches, its author is authenticated, or tests or reviews passed.
Read the preserved intake report alongside the model check.

The [authoring schema](../authoring/README.md) is an editor aid. The bridge's
schema reference identifies an exact tracked revision, not an observed current
filesystem file or a fetched URL. Its `.invalid` URL is an identifier only.
No JSON Schema evaluator runs here. A Go test checks the pin against the tracked
asset, so future schema changes require deliberate reference review. Node/Ajv
remain test-only dependencies of the existing schema contract, not this Go tool.

Every report retains `mode=example-only`, `authority=none`,
`executionStatus=not-run`, `acceptanceStatus=not-evaluated`, and
`reviewStatus=not-established`. There is no UML/SysML conformance claim.

## Byte-preserving preparation and extraction

Supply observations from exact UTF-8 artifact bytes. Do not trim whitespace,
translate CRLF to LF or append a newline before hashing. In Python, use
`Path(...).read_bytes().decode("utf-8")`, not `read_text()` with universal newline
translation. Do not automatically replace an established declaration merely to
make the observed hash match.

For a public example, this creates a new intake declaration from the same source
bytes. The resulting match is deliberately self-declared, **not** independent
provenance or review evidence:

```sh
python3 - <<'PY' > /tmp/intake-example.json
import hashlib, json, pathlib, sys
source = pathlib.Path("examples/lifecycle.json").read_bytes()
request = {
    "formatVersion": "lifecycle-intake-json/v1",
    "intent": "Inspect the public lifecycle example, without approving it.",
    "decisions": [],
    "questions": [{"id": "Q_review", "text": "Does this model express the intended behavior?"}],
    "artifacts": {
        "requirements": None,
        "model": {"sha256": hashlib.sha256(source).hexdigest(), "author": "unverified example label"},
        "scenarioSuite": None
    },
    "dependencies": {"modelRequirements": None, "suiteRequirements": None, "suiteModel": None},
    "observed": {"requirements": None, "model": source.decode("utf-8"), "scenarioSuite": None}
}
sys.stdout.buffer.write(json.dumps(request, ensure_ascii=False).encode("utf-8"))
PY
/tmp/modelintakecheck < /tmp/intake-example.json > /tmp/intake-model-report.json
```

To extract the normalized artifact without changing its bytes:

```sh
python3 - <<'PY'
import hashlib, json, pathlib
report = json.loads(pathlib.Path("/tmp/intake-model-report.json").read_bytes())
assert report["model"]["status"] == "accepted"
artifact = report["model"]["normalized"]
source = artifact["sourceJSON"].encode("utf-8")
assert hashlib.sha256(source).hexdigest() == artifact["sourceSHA256"]
pathlib.Path("/tmp/normalized-lifecycle.json").write_bytes(source)
PY
```

Extraction does not update the original intake or grant review authority. Keep the
original and normalized hashes distinct in any downstream record. These Python
snippets are optional preparation examples, not dependencies of the Go API.

## API, bounds and failures

`Inspect(raw []byte) (*Report,error)` returns an owned typed result.
`Execute(raw []byte) ([]byte,error)` returns one deterministic JSON report plus LF.
Neither performs I/O. Inputs must not be concurrently mutated during a call.
Both return nil on errors, and apply the same final encoded-report bound.

Native intake limits remain 4 MiB for the request and 1 MiB per observation.
Existing model/format bounds apply; output is limited to 16 MiB after escaping.
Default JSON HTML/control escaping preserves strings as data. Reports include
intent, author labels and normalized model contents and can therefore be private.

Check order is gross input size, pinned toolchain, native intake validation,
model extraction and validation, canonical/normalized artifact creation, output
bound. Invalid intake returns `ErrInvalid`, gross size/output limits `ErrLimit`,
wrong toolchain `ErrToolchain`. Unexpected canonical/encoding failure after model
acceptance returns `ErrArtifact` with no report, rather than mislabeling a decoded
model as invalid. Ordinary malformed/semantic-invalid model observations produce
`rejected` reports. Diagnostics contain fixed text, not upstream input details.

For the built command: exit 0 means a complete report, **including missing or
rejected models**; exit 2 means input/read/toolchain/artifact refusal with empty
stdout; exit 3 means failed/short output. Partial transport is never a successful
report. `go run` adds its own wrapper exit behavior; build the binary when callers
need these exact exit codes.

## Independent checks

[Six frozen vectors](testdata/expected.json) were authored independently before
bridge logic using the already-merged components. They lock native report bytes,
LF/CRLF source identities, canonical/normalized identities, missing and rejected
models, mismatched/undeclared valid models, unanswered questions and opaque peer
artifacts. No expected value is refreshed from this bridge's output. Additional
tests cover strict intake/model boundaries, the merged authoring corpus, schema
pin drift, byte ownership, deterministic concurrent calls, escaping, limits and
CLI failure behavior. See the [task contract](../../../tasks/CF-model-intakecheck-v1-semantic-bridge.md).
