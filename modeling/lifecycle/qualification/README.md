# Offline candidate qualification project

This example-only package creates a deterministic project containing a candidate
model, generated reducer and structural tests, a pinned runner source snapshot,
and a literal suite. The suite is copied without semantic validation. Exporting
or verifying a project does not execute or qualify it, authenticate its source,
or establish review or worker authority.

## Two closed source policies

| Exact source revision | Source files | Payloads, excluding manifest | Build selection |
|---|---:|---:|---|
| `3d6d17188cc2723cfe04649d0f8b4220a37b1afe` | 11 | 17 | Explicit historical policy |
| `1e88dbebb4ab13873e2a4159dc928dd1f5f73e20` | 12 | 18 | Current default or explicit policy |

`SourceRevision` identifies the current default. Its actual committed checkpoint
adds `model/capabilities.go` to the complete model package; all eleven existing
path/size/SHA-256 pins stay unchanged. The complete scenarios package still
includes `types.go`, `decode.go`, `run.go`, `runner.go` and `report_decode.go`.
`SourceNames()` returns a fresh list for the current policy.
`SourceNamesForRevision(revision)` returns a fresh list and error for one of the
two exact supported revisions. Unknown, empty, malformed or held nine-source
revision `3312a090afac526ddeda392dbd31ed49fa0e04ca` is rejected.

`Build(candidate, suite, sources)` builds only the current twelve-source project.
`BuildForRevision(revision, candidate, suite, sources)` explicitly selects one of
the two compiled policies. Both require actual Go 1.27.3 and exact selected
source names, sizes and hashes; callers cannot supply their own pin table. The
candidate goes through the existing JSON-v1 decoder, normalized IR, canonical
digest and Go generator. Suite bytes are bounded but not decoded, changed or
rebound. The archive profile remains `lifecycle-qualification/v1`.

The shared compiler, codec, fixed GUIDE and eleven source bytes are unchanged
between these two policies. This is why the historical policy can reproduce
existing archive bytes; it is not a promise of automatic compatibility after
future compiler changes. Source pins identify expected bytes, not an authenticated
executing binary, compiler, callback, host or human review.

## Verify without upgrading old archives

`Verify(archive, expectedCandidateSHA256, expectedSuiteSHA256)` supports both
recognized policies. It bounds and strictly decodes the existing identity metadata
(maximum 4 KiB) to select an exact forty-lowercase-hex source revision. That selection
is routing, not validation. The verifier still checks the selected complete domain
inventory and pins, external candidate/suite hashes, regenerated code/tests/GUIDE,
and complete canonical rebuilt archive equality. Unknown/mixed/relabelled policies
fail; there is no fallback based on file count or whichever hashes happen to match.

Existing eleven-source archives remain byte-for-byte verifiable. They are not
rewritten as current, migrated or silently rebound. A successful verification means
consistency under that archive's recognized source policy, **not current-source
freshness**. A caller needing the current policy must first verify successfully,
then inspect that archive's `qualification.json.sourceRevision` and compare it to
`SourceRevision`. Reading this declaration before verification proves nothing.
Externally expected input hashes must come from a separate reference; supplying
replacement hashes with a replacement archive cannot authenticate its origin.

The pure APIs have no filesystem, subprocess or network activity. Callers must
keep input byte slices immutable during calls. The archive layer never extracts
payloads, and no verifier launches the contained project.

## CLI and explicit execution boundary

From the module root, `modelqualify export` reads a JSON object on stdin with
base64 `candidate` and `suite`, reads only the current fixed source inventory and
writes an archive to stdout. Export always uses the current twelve-file policy;
old hard-coded eleven-source lists are not accepted as current inputs. The typed
historical builder is available only through its explicit revision selection.
`modelqualify verify` accepts base64 `archive` plus the two independent expected
hash strings and verifies either supported policy. Its unchanged receipt remains
`verified; suite unvalidated and unexecuted`; it makes no currentness claim.

Use a trusted local single-writer host. Static symlink, regular-file and bounded
read checks limit accidental path substitution, but cannot prevent a malicious
concurrent writer from racing filesystem operations. Do not put private inputs in
public source, logs or fixtures. Exit 2 remains the policy for rejection and failed
or short output, including real closed stdout; no output error is a verified project.

The unchanged exported GUIDE gives explicit offline Go 1.27.3 commands. Those
commands disable VCS stamping with `-buildvcs=false`, `GOWORK=off`, `GOENV=off` and
empty `GOFLAGS`. Project identity is the manifest and separate expected hashes,
not ambient VCS metadata. The candidate-compiled default runner checks binding and
expectations only when separately invoked. The suite can be malformed or bound to
another model and still be exported/verified as unchanged opaque bytes.

## Regression and adoption scope

The original candidate/suite and full expected replay/mismatch reports remain
unchanged. Readable eleven-source expectations are retained as historical policy
evidence. Independently frozen current/legacy controls check the exact old public
archive, both source policies, wrong/mixed revisions, metadata and inventory
substitution, ownership, limits and unchanged opaque suites. The existing trusted
execution test builds the current exported module, runs structural tests, checks
full four-case/six-step literal reports, catches two compiling reducer mutants,
then restores/rebuilds/replays the original. That bounded harness is not a product
runner, arbitrary-process sandbox or descendant-cleanup guarantee.

Profile discovery is a descriptive authoring aid; see the
[profile catalog](../profilecatalog/README.md). This adoption does not complete held
reviewpack visual work, adopt held nine-source projects, change private pins,
activate workers, or claim UML/SysML conformance.

With actual Go 1.27.3 from the module root:

```sh
GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off go test -race -count=1 ./artifactbundle ./qualification ./cmd/modelqualify ./profilecatalog
```
