# Offline candidate qualification project

This example-only package creates a deterministic project containing a candidate
model, its generated reducer and structural tests, the unchanged existing runner,
and a literal suite. The suite is copied without semantic validation. Exporting
or verifying a project does not execute or qualify it, authenticate its source,
or establish review or worker authority.

`Build` requires the exact pinned Go toolchain and accepts candidate and suite
bytes plus the exact nine-file source
inventory pinned by path, size and SHA-256 to checkpoint
`3312a090afac526ddeda392dbd31ed49fa0e04ca`. The candidate goes through the
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
