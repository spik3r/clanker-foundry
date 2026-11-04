# Check a builder's pinned inputs

This Go preflight checks the exact bytes a coordinator intends to give a builder.
It reads a separately pinned manifest and only its named files beneath an
explicit input root. Missing or changed inputs remain visible instead of being
replaced by a search result or silently rebound to new hashes. It does not
compile, import or execute a candidate, invoke Git, generate an implementation,
launch an agent or write files.

This is an additive standalone tool. It does not change the existing
[model-driven review path](../../../docs/model-driven-current-workflow.md) or
install itself with the skill pack. It works with an agent-written Go handoff,
but its opaque byte checks do not depend on the language of an artifact.

## Prepare the input boundary

The coordinator must separately authorize an exact input root and the listed
reads, including metadata for their parent components. A manifest cannot grant
that authority. Review the manifest against the actual handoff before giving
its exact SHA-256 through the trusted receiving instructions. Preserve this pin
separately from the manifest. Hashing whatever manifest arrived and accepting
that same value would not detect replacement of the manifest itself.

Keep original source identities, semantic review, oracle independence, live
ownership, unresolved decisions and permission to edit or execute in the handoff.
This tool establishes none of them. A role such as `model` or `acceptance` is an
inert label, not a completeness check or a reason to trust the contents.

The strict UTF-8 manifest has this shape. This one-file example declares the
literal bytes `abc`; it is not a complete design packet:

```json
{
  "formatVersion": "model-builder-inputs/v1",
  "artifacts": [
    {
      "id": "brief",
      "role": "requirements",
      "path": "source.txt",
      "sha256": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
    }
  ]
}
```

Every field is required. Unknown fields, duplicate JSON keys, nulls, invalid
Unicode and trailing data are refused. The complete manifest is limited to
64 KiB and 1-64 artifacts. Each artifact has:

- A unique `id`, 1-64 ASCII letters, digits, underscores or hyphens.
- A `role`: requirements, model, interface, acceptance, harness, rubric, packet,
  instructions or claim. There is no required mix of roles.
- A unique `path`, 1-1024 UTF-8 bytes, relative to the supplied root with `/`
  separators. Empty, `.` and `..` components, backslashes, colons and control
  characters are refused. Names containing `*`, `?`, `[` or `]` are literal
  filenames; no glob expansion occurs.
- A lowercase 64-hex `sha256` of the original file bytes. Source normalization,
  JSON key order, Unicode normalization and line endings are not ignored.

Artifacts may contain arbitrary bytes, including empty and non-UTF-8 content.
The preflight never parses their meaning. Each file is limited to 1 MiB, with a
bounded extra byte read to detect growth past the limit. With at most 64 files,
there is no unbounded scan of the root or a shared workspace.

## Build and check

Use the installed toolchain required by [the native module](../README.md), from
a trusted checkout. Name authorized output/cache/temp locations outside the
checkout as usual. Building this tool does not execute a candidate. No download
or new dependency is required.

```bash
cd /exact/trusted/clanker-foundry/modeling/lifecycle
GOTOOLCHAIN=local GOWORK=off GOENV=off GOPROXY=off GOSUMDB=off \
  go build -o /exact/authorized/scratch/modelbuilderpreflight ./cmd/modelbuilderpreflight
/exact/authorized/scratch/modelbuilderpreflight \
  --root /exact/authorized/input-root \
  --manifest-sha256 "$REVIEWED_MANIFEST_SHA256" \
  < /exact/reviewed/input-manifest.json \
  > /exact/authorized/scratch/input-report.json
```

The shell opens the named manifest and report; those paths need their own read
and write authority. The program accepts only the two required flag/value pairs,
in either order, and the manifest on stdin. There is no implicit current root,
manifest discovery, candidate flag or execution mode. Check the process exit and
the complete report together; in a shell with `set -e`, capture exit 1 explicitly
when a blocked report is an expected outcome.

The library API is `Check(manifest []byte, rootPath, manifestSHA256 string)
([]byte, error)`. It validates the complete manifest and its separate pin before
any filesystem access. `ErrInvalid` rejects the request or root without a report.
No report means no artifact-level pass or failure has been established.

## Read the report without promoting it to approval

Output is deterministic JSON with one final newline and artifacts in manifest
order. The envelope contains:

- `formatVersion: model-builder-input-report/v1` and the exact `manifestSHA256`.
- `outcome: matched` only when every listed file matches; otherwise `blocked`.
- `executionStatus: not-run`, `reviewStatus: not-established`, `authority: none`.

Each item retains `id`, `role`, `path` and `expectedSHA256`. Its `status` is:

- `matched`: the bounded, fully read bytes have the declared digest.
- `hash-mismatch`: bounded bytes were read, but their digest differs.
- `missing`: a named component or leaf was absent.
- `refused`: a static symlink, non-directory parent, nonregular leaf, size limit
  or detected identity/change condition prevented a byte observation.
- `read-failed`: another filesystem read/metadata error prevented observation.

Only `matched` and `hash-mismatch` include `observedSHA256` and `observedBytes`.
The other statuses carry null observations, never a hash of partial bytes. Every
listed artifact is attempted even if an earlier one is blocked. File content,
absolute root paths and raw operating-system errors are not included.

Exit 0 means all listed bytes match; 1 delivers a blocked report; 2 rejects the
arguments, manifest or root; 3 means failed or short stdout delivery; 4 means
stdin I/O or an internal failure. Closed stdout uses exit 3 rather than SIGPIPE
termination. A partial or undelivered report is not a completed observation.

A matching report does **not** establish the source revision, semantic model
identity, authenticity, independent authorship, current ownership, completeness,
review, approval, permission or readiness. It says nothing about unlisted files.
Carry unresolved decisions unchanged. If a pin is stale, investigate which input
changed; do not refresh the expected digest just to turn the result green.

## Filesystem and observation limits

The root must be absolute and its final component a real directory. Its ancestor
aliases are allowed. Static symlinks in artifact components, including links to
other files inside the root, are refused. Go's rooted filesystem API bounds named
reads beneath the opened root; this is not a sandbox or an authority mechanism.
Directory listings, globbing and fallback searches are not performed.

Use a trusted, exclusive input tree. Metadata checks detect ordinary identity,
size and timestamp changes, but do not provide an atomic multi-file snapshot or
protection from hostile concurrent writers. Hardlinks, bind mounts and filesystem
provenance are not authenticated. Some filesystems update access times on reads;
no explicit writes are made. A report can become stale immediately after a check.
Recheck exact inputs at the authorized builder boundary and separately inspect
any later candidate before deciding whether to execute it.

The independent tests were frozen before implementation. Run native checks with
`go test -race -count=1 ./builderpreflight ./cmd/modelbuilderpreflight` and
`go vet ./builderpreflight ./cmd/modelbuilderpreflight`. The tests exercise static
filesystem boundaries and a real closed output pipe. Access-time probes skip
when the filesystem cannot support the observation. They do not supply a complete
system-call audit, concurrent-writer proof or cross-platform qualification.
