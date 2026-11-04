---
id: CF-builder-preflight-v1
status: in-progress
owner: Foundry builder preflight contributor
repositories: [spik3r/clanker-foundry]
---

# Check pinned builder inputs without executing them

## Implementation claim

Base: verified public main f16c94d456e89b7ee0680b91d13a44fb79be5556.
Branch: codex/builder-preflight-v1. Add a local Go preflight for the exact
files a coordinator intends to hand to a model-driven builder. It compares
original bytes with a separately pinned manifest, does not discover substitute
inputs, and keeps identity separate from semantic review and authority.

Exact maintained scope:
- modeling/lifecycle/builderpreflight/check.go
- modeling/lifecycle/builderpreflight/files.go
- modeling/lifecycle/builderpreflight/check_test.go
- modeling/lifecycle/builderpreflight/README.md
- modeling/lifecycle/cmd/modelbuilderpreflight/main.go
- modeling/lifecycle/cmd/modelbuilderpreflight/main_test.go
- tasks/CF-builder-preflight-v1-pinned-inputs.md
- new immutable source/event files under this task ID

Do not edit existing workflow documents, templates, frozen Go pilot fixtures,
models, compiler, CLI routes, runtime, CI, visual artifacts or other tasks. Do
not execute a candidate, launch a worker, install tools, publish or merge. No
extra-platform testing is requested. This local proposal cannot acquire a
shared task lock or substitute for a live ownership check.

## Ownership and source check

GitHub main and open PRs were checked on 4 November 2026. The only open PR was
79, head 11cba2cde170727ee7fb71ac4b11c4b315750148, which adds handoff guidance
and a frozen Go pilot. This work does not change or depend on those unmerged
files. Repository task claims were inspected; no existing preflight claim was
found. The coordinator was given this exact additive scope to coordinate the
separate external worker. Recheck live ownership before later publication.

## Prospective v1 contract

`Check(manifest []byte, rootPath, manifestSHA256 string) ([]byte, error)` accepts
an absolute, explicitly authorized input root and a separately supplied lowercase
64-hex SHA-256 of the exact manifest bytes. Validate the entire manifest and its
pin before any filesystem access. A malformed or unpinned manifest produces no
report and no artifact reads. This pin proves byte correlation, not approval.

The strict JSON manifest has exactly `formatVersion` equal to
`model-builder-inputs/v1` and `artifacts`, a nonempty array of at most 64 items.
Each item has exactly `id`, `role`, `path` and `sha256`. IDs are 1-64 ASCII letters,
digits, underscores or hyphens and unique. Roles are one of requirements, model,
interface, acceptance, harness, rubric, packet, instructions or claim. Paths are
unique UTF-8 slash-separated root-relative names of at most 1024 bytes, with no
empty/dot/dot-dot components, backslash, colon or control characters. A path is
not a glob or URL. Digests are lowercase 64-hex. Reject unknown/missing/duplicate
JSON keys, nulls, invalid Unicode, trailing JSON and manifests over 64 KiB.
Artifact roles only label bytes; no role makes its contents trusted or complete.

Open only the explicit root, the named path components for metadata and the
listed regular file leaves. Reject static symlinks at the root leaf or any
artifact component, directories and special files. Use Go's rooted filesystem
API to bound reads; do not enumerate directories or follow substitute names.
Each artifact is bounded to 1 MiB, including growth during reading. Compare
exact bytes, without parsing, normalization or candidate import. Use a trusted,
exclusive local input tree: these checks do not authenticate mounts, hardlinks,
provenance, atomic multi-file snapshots or hostile concurrent modifications.

Produce a deterministic newline-terminated JSON report with formatVersion
`model-builder-input-report/v1`, manifestSHA256, outcome `matched` or `blocked`,
executionStatus `not-run`, reviewStatus `not-established`, authority `none`, and
artifacts in manifest order. Each item retains id, role, path and expectedSHA256,
and has status `matched`, `hash-mismatch`, `missing`, `refused` or `read-failed`,
with observedSHA256 and observedBytes null unless bytes were fully read within
the limit. A missing path is `missing`; symlinks, nonregular leaves/components,
size limits or changed file identity are `refused`; other I/O failures are
`read-failed`. Report all listed outcomes even after one fails. Never refresh
pins, copy files or embed file contents. A matching report grants no permission,
source authenticity, review, independence, semantic correctness or readiness.

The standalone CLI accepts exactly `--root ABSOLUTE-PATH --manifest-sha256 HEX`
in either pair order, with the manifest on stdin. No defaults or repeated flags.
Exit 0 means all listed bytes match, 1 means a blocked report, 2 invalid arguments,
manifest or root, 3 failed/short stdout delivery, 4 input I/O or internal failure.
A rejected input prints no report. Diagnostics do not echo private paths or
file contents. Closed-pipe output must use exit 3, not signal termination.

## Ordered evidence plan

1. Commit this claim, register its pinned task source, and record a genuine start.
2. Have a separate author freeze literal Go tests from this contract before
   implementation. Preserve those tests; discrepancies require explicit review.
3. Implement the bounded package and CLI, then run relevant native Go tests and
   race checks, vet, repository validation and committed Project Memory checks.
4. Seek independent review of exact final source, record observations, and hand
   off the local branch and a selected-ref full-history bundle. No publication.
