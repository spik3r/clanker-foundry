---
id: CF-workspace-tree-names-v1
status: done
owner: Codex coordinator with delegated builder
repositories: [spik3r/clanker-foundry]
---

# Preserve exact Git tree names in candidate workspaces

## Implementation claim

The owner requested continued delegated Foundry maintenance, review and tested
normal merges. The coordinator reviewed this public-safe claim against main
`058c439f0dca4028afc3266958623e1302cae8c2`. Live PRs and issues are empty;
Project Memory reports 177 tasks, no conflicts and no active run. Use a separate
worktree and preserve all unrelated work. The historical workspace byte-fidelity
claim fixed source content/BOM handling and reserves no new implementation.
Its source is retained in main; do not alter that record.

An owned Git-object fixture contains a filename with byte 0xff, while its exact
read allowlist names a valid U+FFFD filename absent from the tree. Replacement
UTF-8 decoding in parseTree turns the former into the latter. Actual candidate
workspace creation invokes the fake verifier and returns the wrong-path blob.
This violates the documented exact-path allowlist without requiring any invalid
host filesystem filename, provider, container, authority database or dispatch.

Exact scope:

- `runtime/src/workspace.mjs`: decode the bounded whole Git tree capture with
  fatal UTF-8 before constructing records. Preserve BOM characters. Map failure
  to fixed WORKSPACE_GIT_ERROR without including raw names or capture bytes.
- `runtime/test/workspace-tree-names.test.mjs`: frozen public-API regressions
  using owned Git plumbing objects, fake hosts and an owned Git command recorder.
- `runtime/README.md`: state the malformed tree-name refusal and valid-name
  fidelity, including whole-capture refusal for an unselected malformed name.
- This canonical task record and new immutable Project Memory source/events.

Whole-capture refusal is the chosen trade-off. It rejects unsupported names even
outside the allowlist, rather than adding a byte-oriented selective matcher.
The current string path API cannot represent arbitrary malformed filename bytes
without loss. Keep valid Unicode, literal U+FFFD, supplementary text, composed
and decomposed spellings, BOM within names, tabs and newlines unchanged. Keep
NUL record framing and the first-tab header separator; do not normalize names.
No new filename/control-character policy or general Git parser rewrite is part
of this task. Existing path, pin, blob, byte-bound, symlink and cleanup rules stay.

No schema, dependency, version, provider, CLI, store/state/runner, host-isolation,
source-content decoder, consumer or historical metadata edits are included.
This source maintenance does not activate managed orchestration, change namespace
or authority, grant promotion, run a provider/container or migrate claims.

## Acceptance and validation

Commit this reviewed source, register its exact revision, independently compare
record bytes and identity, then record a genuine Node 24 v2 start before tracked
tests or code. Freeze literal tests and reproduce failures on unchanged code.

Malformed byte, truncated UTF-8 and encoded-surrogate names must refuse before
blob lookup, candidate directory creation or host verification. Assert fixed
error code, no raw-name disclosure, zero host verify/apply calls and no cat-file
calls after malformed tree metadata. Preserve the owned source repository.

Cover malformed-name-only and a collision with a real U+FFFD filename, with
malformed bytes ordered both before and after the valid entry. An unrelated
malformed entry also refuses. Valid U+FFFD/astral/non-ASCII/composed/decomposed,
BOM, tab and newline names retain exact path, original content and digest.
Use no invalid names in a real host worktree and no live authority or container.

Use actual Node 24, install pinned dependencies with npm ci --prefix runtime
--ignore-scripts, run focused tests and the full runtime suite, pack validation,
installer regressions and whitespace checks. Preserve package manifests/locks.
Record observed results and exact platform limits. Root stages named paths,
records completion from actual staged blobs, commits new immutable evidence and
validates the committed diff with explicit code/source identities. Do not edit
old memory or legacy ledger bytes.

Two independent reviewers bind frozen files and final commit/tree. The coordinator
alone commits, pushes, creates the PR and merges normally after all current hosted
gates pass, including orchestration Node 24 and its live isolation CI fixture.
Local fake-host checks do not establish a live-host result or runtime activation.

## Observed implementation and checks

The candidate reader now decodes the bounded whole tree capture with fatal UTF-8
and preserves BOM characters. A malformed capture returns fixed
`WORKSPACE_GIT_ERROR` before records, blob lookup, candidate directory creation or
host verification. Valid names retain the existing framing and path rules.

The maintained public-API tests first ran against unchanged production code.
An initial positive fixture put composed and decomposed names in one candidate;
this host filesystem aliases those spellings, so it refused the second file.
The final fixture tests each valid spelling separately. The frozen tests also
cover malformed names before and after a real U+FFFD entry in Git's byte order.
After the test indentation review, the final test file was replayed against the
old production source: 17 tests, eight valid-name passes and nine refusal failures.
Its SHA-256 is
`2d6e28f837b3e71ceb9070f56895cbf6702cdb470a9ffcb13b6492e56440e3df`.

Actual Node 24.15.0 passed all 17 focused tests and all 350 runtime tests with no
skips. The first full run passed 349 tests and failed the offline consumer test
because npm could not write to the ambient user cache. A task-owned npm cache,
seeded with the lockfile-pinned dependencies, allowed the full rerun to pass;
no user cache permissions or contents were changed. Pinned npm installation used
`--ignore-scripts`. Repository pack validation and installer regressions passed;
optional skills-ref was unavailable. Whitespace checks passed.

These checks use owned Git objects, a command recorder and invented host callbacks.
They do not run a provider, container, authority database or dispatch and do not
establish a live-host isolation result. Independent final review, committed
Project Memory completion and hosted checks remain pending with the coordinator.

## Coordinator local completion

An independent coordinator run on Node 24.15.0 installed the locked dependencies
with scripts disabled and an owned npm cache. The full runtime suite passed
350 tests with zero failures, skips or cancellations. Pack validation, validator,
installer and project-setup regressions and ShellCheck passed. Runtime source,
regressions, README and original source/start evidence stayed unchanged.

Local implementation is complete. The coordinator will complete the genuine run
from actual staged blobs and validate the committed diff. Exact-head independent
review and all hosted gates, including the live isolation fixture, remain required
before normal merge. No live host, provider, authority database or dispatch was
used for local qualification.
