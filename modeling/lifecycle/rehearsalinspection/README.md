# Inspect carried rehearsal reports

The existing [rehearsal carrier](../rehearsalhandoff/README.md) deliberately treats
report bytes as opaque hash-matched records. This separate inspector checks those
reports against the carried literal suites. It preserves the legacy carrier,
verifier, receipt and closed qualification source policies.

`Inspect(carrier, references)` first runs the existing complete carrier verifier.
It then decodes each native scenario report under its verified model identity,
checks native report structure and comparison fields, and compares its scenario
digest, original suite-byte hash, case and step order, inputs, commands and literal
expectations with the corresponding suite. Every successful report must cover the
whole suite. A valid mismatch must stop at its first mismatching step and retain
the exact preceding suite prefix. Refusals retain the prior input; successes feed
the supplied actual task into the following step.

**Inspection does not establish execution.** Self-consistent invented actual
results can pass these checks. The inspector does not recompute reducer behavior,
invoke a callback, authenticate the reports or decide that the expectations are
right. A valid early mismatch is returned as `matched: false`, with its original
partial counts. Native malformed comparison fields, wrong identities, changed
literals and missing or reordered steps are rejected even if a caller updates the
report-byte references. Neither matching hashes nor an inspection receipt grant
approval or permission to execute.

## API and command

The pure APIs use bytes, never filesystem paths:

- `Inspect(carrier []byte, references rehearsalhandoff.References) (Receipt, error)`
- `InspectRequest(request []byte) ([]byte, error)`

Inputs must remain immutable during a call. The additive
`rehearsalhandoff.InspectMaterial` accessor returns owned verified payload bytes;
its reports remain opaque until separately inspected. The additive
`scenarios.Runner.DecodeReport` validates native `lifecycle-scenario-report/v1`
reports using the runner's declared binding without calling its reducer. It does
not accept the separate bound-report envelope. Existing source-policy files and
legacy default decoding are unchanged.

From the lifecycle module, with installed Go 1.27.3, build the command into an
authorized output directory outside the checkout:

```sh
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
go build -o "$INSPECTION_WORK/modelrehearsalinspect" ./cmd/modelrehearsalinspect
"$INSPECTION_WORK/modelrehearsalinspect" < "$INSPECTION_WORK/request.json" > "$INSPECTION_WORK/receipt.json"
```

Set `INSPECTION_WORK` to your private output directory first. The caller supplies
one strict JSON request with these three required fields:

```json
{"formatVersion":"lifecycle-rehearsal-inspection-request/v1","references":{"requestSHA256":"...","beforeSourceSHA256":"...","beforeModelDigest":"...","candidateSHA256":"...","candidateModelDigest":"...","candidateSuiteSHA256":"...","baselineSuiteSHA256":"...","baselineReportSHA256":"...","candidateReportSHA256":"..."},"bundleBase64":"..."}
```

The ellipses stand for the separately obtained full lowercase SHA-256 references
and canonical padded standard base64 carrier bytes; this illustration is not a
valid request. Obtain references independently of the carrier under review.
Request assembly and shell redirection are caller actions. The CLI opens no
caller path and makes no network or provider call. Reports and references may
contain private information; keep real inputs out of public fixtures and logs.

A receipt says `formatVersion: lifecycle-rehearsal-inspection/v1`,
`mode: example-only`, `authority: none`, `executionStatus: not-established`, and
`reportStatus: inspected-for-consistency`. Its `baseline` and `candidate` fields
retain model and scenario digests, exact original suite-byte SHA-256, match status,
completed cases and inspected report step counts. These fields describe supplied
records, not observed execution or authenticated provenance.

## Bounds, failures and checks

Requests are capped at 36 MiB, carrier bytes at 24 MiB and decoded carrier payloads
at 16 MiB; existing nested limits still apply, including 2 MiB per report. Encoded
and decoded base64 lengths are checked before bulk decoding. Duplicate, unknown,
missing, null and case-variant fields, invalid Unicode, noncanonical base64 and
trailing input reject. Pure APIs return no partial receipt. The CLI returns 0 for
an inspected receipt, 2 for input or consistency rejection and 3 for a failed or
short output write, including a real closed stdout pipe. A redirected file can
still hold a prefix after failure; check the exit status before using it.

The [frozen observations](testdata/observations.json) were authored by a separate
assistant contributor before inspector implementation. They include predicted
baseline and candidate records, wrong source/model/scenario identities, literal
and input substitutions, truncation, reordering, false comparison claims, opaque
and duplicate-key reports, and a valid early mismatch. This contributor separation
is not independent-person review or proof of execution.

```sh
go test -race -count=1 ./rehearsalinspection ./scenarios ./rehearsalhandoff ./cmd/modelrehearsalinspect
```

The CLI regression builds the trusted inspector and checks both a complete receipt
and an actual closed stdout pipe; it does not compile or run a candidate reducer.
Legacy carrier tests still cover opaque records. No runtime task, worker dispatch,
production integration, deployment or visual-review claim follows.
