# Offline candidate qualification project

This example-only package creates a deterministic project containing a candidate
model, its generated reducer and structural tests, the unchanged existing runner,
and a literal suite. The suite is copied without semantic validation. Exporting
or verifying a project does not execute or qualify it, authenticate its source,
or establish review or worker authority.

`Build` requires the exact pinned Go toolchain and accepts candidate and suite
bytes plus the exact eleven-file source
inventory pinned by path, size and SHA-256 to checkpoint
`3d6d17188cc2723cfe04649d0f8b4220a37b1afe`. The candidate goes through the
existing JSON-v1 decoder, normalized IR, canonical digest and Go generator. The
suite is bounded but is not decoded, changed or rebound. The artifactbundle
profile is `lifecycle-qualification/v1`.

`Verify` also requires Go 1.27.3 for reconstruction. It takes an archive and
independently supplied expected SHA-256 values for
the candidate and suite. It validates the archive, pinned sources and complete
domain inventory, regenerates the candidate project, and compares the complete
archive bytes. The expected hashes must come from a trusted external reference:
rebuilding an archive and replacing both expected hashes cannot establish the
origin of its contents. Verification is a consistency check, not authentication.

From the module root, `modelqualify export` reads a JSON object on stdin with
`candidate` and `suite` as base64-encoded byte strings and writes the archive to
stdout. It reads only the fixed source inventory from the current module root.
`modelqualify verify` reads a JSON object containing base64 `archive` and string
`candidateSHA256` and `suiteSHA256`, and prints a verification status. It never
launches Go, a shell, a container, a worker or a network operation. The exported
`GUIDE.txt` gives explicit offline commands for a human operator using Go 1.27.3.
The unchanged candidate-compiled runner checks the suite's model binding and
expectations when run.

Use a trusted local single-writer host. Static symlink, regular-file and bounded
read checks limit accidental path substitution, but do not prevent a malicious
concurrent writer from racing filesystem operations. Do not put private inputs in
public source, logs or committed fixtures.

Exported project commands disable Go VCS stamping with `-buildvcs=false` and
set `GOWORK=off`, `GOENV=off` and an empty `GOFLAGS`. Project identity is the
verified manifest and separately supplied hashes, not ambient VCS metadata.

Callers must keep candidate, suite, and source byte slices immutable during
`Build`, and archive bytes immutable during `Verify`. This is an API resource
and aliasing condition, not a claim about hostile hosts.


## Current-source adoption

This is a fresh adoption of the original two-profile pure archive subset and
qualification exporter, not adoption or completion of the held reviewpack or its
visual work. The archive implementation and original independent ZIP oracle stay
unchanged. The inert review profile name grants no review or rendering authority;
the later design-change profile is not included in this adoption.

The copied source closure is explicitly reconciled to the public revision above.
It includes the complete current non-test scenarios package: `types.go`,
`decode.go`, `run.go`, `runner.go` and `report_decode.go`. Three old scenario pins
changed and the last two files are new to the inventory; the other six source
pins stay unchanged. The exact eleven path/size/hash pins are in qualification.go
and independently checked by adoption_test.go. The archive contains seventeen
payloads plus one manifest. The longest permitted source name is updated for
`scenarios/report_decode.go`; names are still restricted to the exact inventory.

Old nine-source archives intentionally fail this verifier's revision and
inventory policy. There is no silent pin refresh, migration, suite rebinding or
claim of compatibility with those archives. A source revision identifies expected
bytes, not the provenance of an executing compiler, binary or callback. The
exported default runner still compiles against the candidate generated constants;
including the complete Runner package does not authenticate caller bindings.

The existing qualification test retains the original candidate and literal suite
exactly. Its separately observed trusted execution builds the exported module,
runs generated structural tests and compares the full four-case/six-step report
to independently frozen literals. Wrong binding, wrong expectations and malformed
input remain failures. Two real generated-code mutants must compile and fail
behaviorally; the original code is then restored, rebuilt and replayed. This test
uses bounded waits/output and a fixed offline environment, not a sandbox or a
guarantee of arbitrary-process termination. It is not a new product runner.

Every export still copies its suite opaque and every successful verify still
prints `verified; suite unvalidated and unexecuted`. Running these fixed tests
does not mark a later caller's suite executed. Even a bounded malformed suite can
be exported and verified as unchanged bytes; the compiled runner rejects it when
explicitly invoked. Existing finite-model and trusted single-writer assumptions
remain; no runtime, worker, deployment or standards-conformance claim is added.

The CLI preserves exit 2 for rejected requests and failed/short writes. SIGPIPE
is ignored so real closed stdout also follows that existing error policy rather
than terminating by signal. No output error is treated as a verified project.

Run with actual Go 1.27.3 from the module root:

```sh
GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off go test -race -count=1 ./artifactbundle ./qualification ./cmd/modelqualify
```
