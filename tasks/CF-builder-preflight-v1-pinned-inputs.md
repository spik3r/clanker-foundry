---
id: CF-builder-preflight-v1
status: review
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
- modeling/lifecycle/builderpreflight/root_test.go
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

## Review-discovered root boundary

Independent review of the first implementation found that lexical cleaning of an
absolute root can select a different directory across an ancestor symlink and
`..`. It can also erase missing or non-directory components. The original 21
independent acceptance entry points passed but did not cover these root names.
Their frozen bytes remain unchanged. The first local run was stopped without
completion coverage; the expanded scope adds only root_test.go before fixing
this issue. A separate author will freeze root traversal regressions against
the intended contract, verify failure on the unfixed implementation, and retain
that chronology. The root must preserve the host's supplied path resolution;
trailing separators may be trimmed only for checking a final symlink leaf.

## Local completion observations

The initial two independently authored test files were frozen in
b83e68040a8477f9950a29c96d78d12955de8229 before production source was written.
Their SHA-256 values are 64368110b7f94378ea7e519f28d63a06f767f6c241ee82d51a942994249abdc1
(check_test.go) and e66891574802731ee6789aaa811f77017d89cb002c8bf23279b8f4ca44cbd2b3
(main_test.go). They remain unchanged. The separate root regression file was
frozen with SHA-256 db832b915d331f45bef67a891f4f8f629018e6e3f662166501f8412f9c6cf04d;
its four observed failures are retained in the red baseline before correction.
The corrected source preserves host root traversal and passes those regressions.
The original reviewer repeated the same probes and found no remaining source
blocker. That is a bounded source review, not a real-use approval.

Native installed Go 1.27.3 offline checks passed: the selected race suite covering
all remaining tests in the two new packages and strictjson, a separate FIFO-only
race run, focused vet, and CLI build. There are 24 new test entry points. The
Unix-socket subcase is not qualified: the host denied fixture creation. It was
retained as a recorded skip in the initial run and omitted from corrected runs;
no permission workaround was attempted. Static size limits, access-time read
observations, permission errors and the real closed-output process were exercised.
Concurrent growth, identity replacement, partial artifact-read failure and a
complete system-call audit remain untested beyond source review.

The built corrected CLI matched byte-identical copies of the existing fictional
brief and model (exit 0), then reported a newline-only model edit as a mismatch
(exit 1) while retaining the same manifest pin and unchanged brief match. No
candidate was compiled or executed. All 143 Project Memory lifecycle/workflow
regressions, repository validation, formatting and whitespace checks passed.
Installer regression stopped at its missing GNU Stow prerequisite. Optional
skills-ref is unavailable. No other-platform or complete native-module suite
qualification is claimed.

A fresh public source check found main at
4a52cf19a026e8f442d13031135a3e9e10bee8fa, including the merged Go pilot (79) and
workspace byte-fidelity repair (81). The sole open PR was 80, the custom assertion
inventory. Their full changed-path lists are disjoint from this task. This local
branch remains based on f16c94d456e89b7ee0680b91d13a44fb79be5556; no merged-main
union run or publication is claimed. Integration still needs a fresh ownership
and current-base check. Completion evidence covers this local proposal only.
