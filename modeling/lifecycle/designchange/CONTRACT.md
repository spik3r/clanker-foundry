# Inert design-change project

This package joins the existing model patch planner and offline qualification
exporter. It emits one deterministic archive containing the accepted request,
original before JSON token, normalized candidate, declared impact report and a
nested candidate project. It never extracts or executes that project. Successful
verification establishes consistency only, not authentication, human review,
approval, freshness or execution. Supplied content may be private.

This is a fresh adoption of the separately completed original design-change source
at f0877271765346ef8f21b8ef0c799b2acc219f93, not adoption or completion of its held
reviewpack ancestors. See the [fresh task](../../../tasks/CF-designchange-adoption-v1-current-project.md)
for original source identities, exact scope and fresh qualification gates. The
three independently authored original fixtures remain byte-for-byte unchanged.
No reviewpack visual work, browser operation or old lifecycle event is imported.

## Current dependency boundary

The current qualification source revision is
`1e88dbebb4ab13873e2a4159dc928dd1f5f73e20`: twelve exact source files and eighteen
payloads plus one manifest, including the additive `model/capabilities.go` file.
All eleven previous source hashes, compiler/codec and fixed GUIDE bytes remain
unchanged. Source names/count/name bounds are obtained from the qualifier's closed
policy API, never from a duplicate private pin table.

The public historical policy `3d6d17188cc2723cfe04649d0f8b4220a37b1afe` has eleven
sources/seventeen payloads plus manifest. Existing archives retain those exact
bytes and revision and remain verifiable, not upgraded or rebound. The held
nine-source revision and unknown/mixed policies remain unsupported. The outer
project always has seven payloads plus manifest. Sharing generator code between
these two policies is justified by its unchanged bytes, not a promise of automatic
future compiler compatibility or authenticated binary provenance.

The shared archive change only adds the `lifecycle-designchange/v1` profile. Its
format, limits and previous-profile behavior stay unchanged. No product process
runner, model evaluator, runtime authority or standards conformance is added.

## API and carrier

Build(request, suite []byte, sources []artifactbundle.File) ([]byte,error)
Verify(raw []byte, expected Expected) error
Execute(raw []byte) ([]byte,error)
Expected{RequestSHA256, BeforeSourceSHA256, BeforeModelDigest, CandidateSHA256, CandidateModelDigest, SuiteSHA256 string} — all six mandatory lowercase64hex; never inferred from archive assertions. Sources are precisely qualification.SourceNames()/existing qualification pins, no additional filesystem access. Concurrent caller mutation unsupported; no mutation or escaping aliases otherwise. No typed patch DTO duplication.

Build passes original request bytes to existing modelpatch.Execute first. Its exact returned JSON bytes become planner-result.json. Decode only that trusted fixed result carrier (after as RawMessage plus before/afterDigest), and extract request.before raw JSON token bytes only after Execute accepted the complete request. Do not duplicate operation semantics or normalize request bytes. Candidate transport is exactly existing jsonformat.Encode of accepted after; check extraction/encoding roundtrip via existing jsonformat/canonical. Then qualification.Build(candidate,suite,sources), followed by qualification.Verify with candidate/suite hashes. This reuses compiler.Generate through qualification; no second direct generator or evaluator. Request errors retain modelpatch's deterministic stage behavior via errors.Is; qualifier/source/toolchain failures map ErrInvalid. No partial archive/error output.

Execute strict carrier has exactly formatVersion="lifecycle-designchange-json/v1", operation="build"|"verify", requestBase64, suiteBase64, sources, archiveBase64, expected. Every key mandatory; selected union payloads required, irrelevant payloads null. sources is nonnull list of exactly12 {name,dataBase64}; names and maximum name length derive from qualification.SourceNames(), not a copied private pin table. Source decoded cap1MiB each/4MiB aggregate, checked before allocation as specified below. Build has expected=null/archiveBase64=null; verify has requestBase64=null/suiteBase64=null/sources=null and full expected. Standard canonical padded base64 only (decode/reencode equality; reject ignored newlines). Empty vs null distinct. Unknown/duplicate/case-variant/missing fields, malformed Unicode, trailing JSON reject. Build stdout is raw ZIP; verify stdout is fixed canonical JSON {formatVersion:"lifecycle-designchange-verification/v1",consistent:true,mode:"example-only",authority:"none",suiteStatus:"unvalidated and unexecuted",suiteBindingStatus:"not inspected"} plus LF. CLI no args, bounded stdin, full buffering, exit0 success/2 rejection or read failure/3 failed or short output. No extraction, paths, commands or networking.

## Exact inventory and reproducibility

Outer archive has exactly7 payloads plus artifactbundle reserved manifest.json:
GUIDE.txt
before.json
candidate.json
change-request.json
design-change.json
planner-result.json
qualification.zip

before.json preserves the exact request.before JSON token bytes (excluding surrounding carrier whitespace); change-request.json preserves every request byte. candidate.json uses existing canonical JSON-v1 Encode, not a claim to preserve arbitrary candidate source. qualification.zip is the complete canonical archive from existing qualifier: exactly18 current payloads (17 for the supported historical policy) plus its reserved manifest. Nesting avoids copying its inventory/provenance logic. Inner inventory is the selected twelve current or eleven historical pinned sources plus GUIDE.txt,candidate.json,suite.json,qualification.json,generated/lifecycle.go,generated/model_paths_test.go. The suite occurs only inside qualification.zip, exact original bytes. No HTML/SVG/browser or automatic extraction.

design-change.json exact ordered fields: formatVersion="lifecycle-designchange/v1", mode="example-only", authority="none", requestSHA256,beforeSourceSHA256,beforeModelDigest,candidateSHA256,candidateModelDigest,suiteSHA256,plannerResultSHA256,qualificationSHA256,modelFormatVersion,profileVersion,generatorVersion,toolchainVersion,qualificationSourceRevision,suiteStatus="unvalidated and unexecuted",suiteBindingStatus="not inspected". Version values use existing constants, no timestamps, absolute paths or ambient VCS claims. Fixed GUIDE states consistency only, supplied content may be private, no review/approval/execution/authentication/freshness claim; suite may be malformed and is not silently rebound.

Verifier validates all six external Expected fields as lowercase64hex and raw archive limits; shared archive verification bounds and validates canonical outer ZIP, then require exact seven payload names. Bound inner qualification.zip at8MiB and call artifactbundle.Verify(inner, QualificationProfile) to obtain canonical owned payloads. Select the exact policy only from bounded (4 KiB), strict inner qualification.json.sourceRevision: forty lowercase hex and one of the two compiled revisions. This is routing, not validation. Require exact selected 18-current/17-historical-payload qualifier inventory, no extra nested archives or unexpected domain members (do not recursively interpret arbitrary bytes). Obtain candidate.json/suite.json and exact named sources. Compare all six external expectations to actual request bytes, extracted before token, candidate/suite payload bytes and existing canonical.Digest on decoded before/candidate. Also require outer candidate bytes equal inner candidate bytes. Do not trust declared metadata hashes. Call qualification.Verify(inner, actualCandidateHash, actualSuiteHash) to enforce exact trusted pins/generated artifacts, then rerun the private same-revision Build path over original request/suite/verified sources, using qualification.BuildForRevision, and compare complete outer bytes. Outer metadata never selects or overrides the inner policy. A self-consistent replacement request/suite with rewritten hashes still fails caller's external expectations. Same exact trusted executable/toolchain and source pins are prerequisites, not authenticated by archive hashes.

Limits: existing request3MiB/modelpatch bounds; suite1MiB nonempty; source names/exact count derive from qualification.SourceNames(), currently12. No copied qualifier private pin registry. Generic source caps1MiB each/4MiB aggregate; exact hashes/sizes remain qualification authority. Candidate/before1MiB, planner result3MiB, metadata8KiB, GUIDE4KiB; inner qualification archive8MiB to fit unchanged shared MaxFileBytes. Outer16MiB/shared total32MiB/count64 unchanged. Strict carrier32MiB/depth16 outside opaque base64 inputs. Before decoding base64, require canonical alphabet/padding shape, encoded length <=4*((decodedCap+2)/3), compute decoded length from length and terminal padding, and account entire decoded source total <=4MiB before any bulk decoded allocations. Then decode and reencode-equality check (reject ignored newlines/noncanonical pad bits). No lower source cap inferred from private pins.

Build/Execute common deterministic resource ordering: request/suite/source collection gross sizes first (excess ErrLimit; empty request/suite or missing source entries ErrInvalid); require exact source-name set, duplicates/unknown/invalid names ErrInvalid; per-source and total decoded byte caps ErrLimit; only then modelpatch.Execute request semantics, qualification pin/toolchain checks, assembly. Typed Build checks same resource/name conditions over actual byte lengths. Execute first bounds carrier, checks strict shape/union (missing/null/extra/duplicates ErrInvalid), computes all decoded budgets and source names before allocating selected blobs, then invokes Build; malformed base64 is carrier ErrInvalid. Transport errors necessarily precede opaque request semantics; it never parses or validates suite contents.

ErrLimit includes exporter budgets and every artifactbundle.ErrLimit, including late inner/outer assembly or verification. Other shared archive profile/malformed/inventory errors and all qualifier failures map ErrInvalid, except shared ErrLimit wrapped by qualifier remains ErrLimit via errors.Is. Preserve modelpatch ErrInvalid/ErrLimit/ErrPrecondition/ErrConflict using exporter-owned fixed wrapping without caller text. Error returns nil bytes. Verify ordering is external syntax/raw bounds -> outer canonical/inventory -> bounded inner canonical/inventory -> all actual external identities -> qualification.Verify -> full Build/rebuild byte equality. No-op patch allowed; raw and semantic identities remain distinct.

## Suite boundary decision

The suite remains wholly opaque and bounded, matching qualification.Build. A claimed candidate binding is NOT inspected or asserted; a mismatched header, malformed later scenario or non-JSON suite remains archived with unvalidated/unexecuted + not-inspected status. This is deliberate, testable and honest. The positive literal suite is independently known candidate-bound from the earlier frozen fixture, not validated by this exporter. Do not call baseline-bound scenarios.Validate, infer validity from a header, rewrite expectations or rebind.

## Preserved independent acceptance

`testdata/request.json`, `suite.json` and `expected.json` were frozen before the
original exporter logic and stay unchanged. They describe one T_prepare guard
change, the exact raw before-token and normalized candidate identities, declared
before/after transition links and no invented bindings. The differently formatted
original candidate is a distinct raw identity, despite equal semantic digest.
Current-adoption tests separately freeze the current twelve-source/eighteen-payload and historical eleven-source/seventeen-payload
closures, complete metadata meanings and stale/missing/tampered-member negatives.

The CLI retains exit 0 on success, exit 2 for read/input rejection and exit 3 for
failed or short output, including real closed stdout. Diagnostics are fixed and
do not include caller text. No suite content is interpreted to produce a pass.

## Byte-preserving example commands

Use actual Go 1.27.3 from the `modeling/lifecycle` module root on a trusted local
single-writer host. These commands package the already frozen example fixtures;
replace them only with separately reviewed inputs and independently held expected
identities. Do not derive expected identities from the archive being checked.
The Python carrier preserves exact bytes, including CRLF; it does not parse or
rewrite the request or suite. No automatic extraction or replay follows.

```sh
export GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off
python3 - <<'PYBUILD' | go run -buildvcs=false ./cmd/modeldesignchange > /tmp/design-change.zip
import base64, json, pathlib, sys
root = pathlib.Path('.')
names = ['go.mod', 'model/ir.go', 'model/capabilities.go', 'canonical/canonical.go',
         'strictjson/decode.go', 'compiler/generate.go', 'scenarios/types.go',
         'scenarios/decode.go', 'scenarios/run.go', 'scenarios/runner.go',
         'scenarios/report_decode.go', 'cmd/modelplay/main.go']
def encoded(path):
    return base64.b64encode((root / path).read_bytes()).decode('ascii')
request = {'formatVersion': 'lifecycle-designchange-json/v1',
           'operation': 'build',
           'requestBase64': encoded('designchange/testdata/request.json'),
           'suiteBase64': encoded('designchange/testdata/suite.json'),
           'sources': [{'name': n, 'dataBase64': encoded(n)} for n in names],
           'archiveBase64': None, 'expected': None}
sys.stdout.buffer.write(json.dumps(request, ensure_ascii=True).encode('ascii'))
PYBUILD
python3 - <<'PYVERIFY' | go run -buildvcs=false ./cmd/modeldesignchange
import base64, json, pathlib, sys
# This fixture is an independent pre-export reference, not archive metadata.
reference = json.loads(pathlib.Path('designchange/testdata/expected.json').read_bytes())
expected = {k[0].lower() + k[1:]: v for k, v in reference['externalExpected'].items()}
request = {'formatVersion': 'lifecycle-designchange-json/v1',
           'operation': 'verify', 'requestBase64': None, 'suiteBase64': None,
           'sources': None,
           'archiveBase64': base64.b64encode(pathlib.Path('/tmp/design-change.zip').read_bytes()).decode('ascii'),
           'expected': expected}
sys.stdout.buffer.write(json.dumps(request, ensure_ascii=True).encode('ascii'))
PYVERIFY
```

The consistency receipt always retains `authority: none`,
`suiteStatus: unvalidated and unexecuted` and `suiteBindingStatus: not inspected`.
Any later explicit use of the nested project's offline GUIDE is a separate
operation and must retain its own observed result. A passed trusted repository
test does not mark a later caller's suite executed. The repository's existing
qualification test builds a fixed exported candidate, checks full independent
reports, catches two compiled behavioral mutants and restores/replays the original;
that bounded test harness is not an arbitrary-process sandbox or a guarantee of
descendant cleanup.

Run focused checks with the same Go environment:

```sh
go test -race -count=1 ./artifactbundle ./designchange ./cmd/modeldesignchange ./qualification
```


## Historical verification is not currentness

Build and wire-build select only the current twelve-source policy; an old
hard-coded eleven-file source carrier rejects rather than being silently upgraded.
Wire Verify and typed Verify accept unchanged archives under either recognized
policy, preserving all six external identities and complete canonical same-policy
reconstruction. Unknown/empty/held-nine-source revisions, mixed inventories,
metadata substitutions and outer/inner revision disagreements reject. All limits
and source budgets apply to the selected exact policy before acceptance.

A successful unchanged consistency receipt does not mean the archive is current.
A consumer needing currentness must first complete native Verify, then inspect the
verified nested qualification source revision and compare it with the current
qualification.SourceRevision. Before verification, that declaration proves nothing.
The suite remains unvalidated/unexecuted with binding not inspected in both policies.
Existing public eleven-source ZIP fixture bytes and original request/suite/semantic
expectations are preserved; no implicit migration, authentication or review occurs.
