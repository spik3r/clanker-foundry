# Portable rehearsal records

This fixed-demo handoff preserves the original change request, two literal
suites, the existing verified design-change archive, two report byte blobs and
the candidate's generated Go reducer and structural tests in one deterministic,
bounded JSON carrier. The [rehearsal](../rehearsal/README.md) can optionally
produce it after executing its fixed invented scenario flow. It carries no SVG,
visual artifact, arbitrary filesystem path or private consumer data. The carrier
can still contain sensitive task content: treat it and its hashes as private when
appropriate.

**A verified carrier is not proof that replay occurred.** References are
separately supplied caller assertions, not signatures or authenticated provenance.
The verifier checks the two report blobs **only against caller-supplied byte
hashes**; it does not parse or semantically validate those reports. A caller can
supply fabricated bytes and matching hashes, and verification can still pass.
The receipt deliberately says `caller-supplied-records`, `authority: none`, and
`reportBytesStatus: hash-matched-only-not-inspected`. It does not establish
execution, code provenance, identity, approval, safety or operational readiness.
A reviewer must obtain and assess reference identities independently of the
carrier. Matching hashes are consistency checks, not proof of origin.

## Build and verify boundary

`Build(Material, References) ([]byte, error)` and
`Verify([]byte, References) (Receipt, error)` are pure in-memory operations; they
perform no filesystem, subprocess, network, provider or worker actions. Inputs
must not be concurrently mutated. `Build` returns owned canonical bytes and
`Verify` returns a value-only receipt. Both reject any inconsistency without a
partial result. The profile is deliberately narrower than the general input limits
of the underlying tools; it is not a general-purpose evidence archive or runner.

References contain nine separate lowercase SHA-256 strings: request bytes,
before-source bytes, before-model digest, candidate bytes, candidate-model digest,
candidate-suite bytes, baseline-suite bytes, and baseline/candidate **report
bytes**. No reference is taken from carrier metadata. The verifier:

1. Bounds and strictly decodes the complete canonical JSON carrier and canonical
   padded base64. All keys are mandatory; duplicate, unknown, case-variant,
   missing, null, empty, trailing, malformed or noncanonical input rejects.
2. Compares request, both suite bytes and both opaque report blobs to separately
   supplied references. Report content is not inspected.
3. Delegates nested archive reconstruction and the six external design-change
   identities to existing `designchange.Verify`; verifies the outer and nested
   canonical archives with `artifactbundle.Verify`. It requires the carried
   request and candidate suite to match the verified archive, and the carried
   generated Go and structural-test bytes to match verified nested members.
4. Delegates the two suites' model-binding preflight to existing `scenariodiff`
   against the verified before and candidate models. A stale binding rejects even
   if a caller changes its suite-byte reference. This independent declaration
   check does not change the nested archive's own `unvalidated and unexecuted`
   suite status or `not inspected` binding status and does not run a reducer.

The carrier is compact UTF-8 JSON with a final LF, fixed field order and canonical
padded standard base64. A semantically equivalent but differently formatted
carrier is rejected. Limits are 24 MiB encoded carrier, 36 MiB CLI request and
16 MiB total decoded payload: request/suites/generated files at most 1 MiB each,
archive at most 8 MiB and each report at most 2 MiB. Encoded-length and decoded
budgets are checked before bulk base64 decoding. Invalid data returns `ErrInvalid`;
this profile's resource limits return `ErrLimit`. No report contents are treated
as an expectation or repaired.

## Produce and inspect the invented example

From the repository root on a trusted local host, with Bash, Python 3 and Go
1.27.3 installed and a private writable temporary directory:

```sh
umask 077
HANDOFF_DIR="$(cd -- "${TMPDIR:-/tmp}" && pwd -P)"
HANDOFF="$(mktemp "$HANDOFF_DIR/foundry-handoff.XXXXXXXX")"
if ! modeling/lifecycle/rehearsal/run.sh --handoff > "$HANDOFF"; then
  rm -f -- "$HANDOFF"; echo 'handoff failed; partial output discarded' >&2; exit 1
fi
```

The default no-argument rehearsal output remains byte-compatible. `--handoff`
runs the same fixed baseline and candidate checks and writes the carrier **only
after successful temporary-directory cleanup**. It is local demo orchestration,
not a sandbox; the trusted compiler and generated Go execute in the private copy.
A failed or short write can still leave a prefix in a redirected file, so check
its exit status and verify it before use.

For the **invented frozen example only**, construct references separately from
its two existing independently prepared fixture files and send a verification
request to the no-path stdin CLI:

```sh
cd modeling/lifecycle
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off
python3 -I - "$HANDOFF" <<'PY' | go run -buildvcs=false ./cmd/modelrehearsalhandoff verify
import base64, json, pathlib, sys
p = pathlib.Path('.')
e = json.loads((p/'rehearsal/testdata/expected.json').read_bytes())
i = json.loads((p/'rehearsalhandoff/testdata/report-identities.json').read_bytes())
x = e['externalExpected']
refs = {'requestSHA256':x['RequestSHA256'], 'beforeSourceSHA256':x['BeforeSourceSHA256'],
        'beforeModelDigest':x['BeforeModelDigest'], 'candidateSHA256':x['CandidateSHA256'],
        'candidateModelDigest':x['CandidateModelDigest'], 'candidateSuiteSHA256':x['SuiteSHA256'],
        'baselineSuiteSHA256':e['baselineSuiteSHA256'], 'baselineReportSHA256':i['baseline']['sha256'],
        'candidateReportSHA256':i['candidate']['sha256']}
req = {'formatVersion':'lifecycle-rehearsal-handoff-verify/v1', 'references':refs,
       'bundleBase64':base64.b64encode(pathlib.Path(sys.argv[1]).read_bytes()).decode('ascii')}
sys.stdout.buffer.write(json.dumps(req, separators=(',', ':')).encode())
PY
```

The CLI accepts exactly `pack` or `verify` and bounded stdin; it never opens a
caller path. Its exit codes are 0 for a complete result, 2 for input/toolchain
rejection, and 3 for failed or short output. Fixed diagnostics omit payloads.
The example shell and Python carrier explicitly read local files; that is not a
capability of the Go verifier. The predicted complete report identities were
separately reasoned from frozen suite literals and report format before handoff
implementation; this is separate-contributor evidence, not independent-person
attestation or proof that a stored report was executed. The historical #43
summary hash is a separate observed-output compatibility pin, not a predicted
behavior oracle. Remove the private carrier when it is no longer needed.

## Tests

From `modeling/lifecycle`, run:

```sh
go test -race -count=1 ./rehearsalhandoff ./cmd/modelrehearsalhandoff ./rehearsal
```

This exercises canonical round trips,
source/reference/archive/code/report tampering, stale suite bindings with updated
byte hashes, malformed and bounded input, opaque-report limitations, repeated
handoff bytes and default-output compatibility. The retained rehearsal test also
checks actual candidate replay and cleanup failure. These are bounded regression
checks, not proof of a general caller's execution or security.
