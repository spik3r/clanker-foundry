# Project Memory format contract

This document describes the locally validated v2 contract. Field names and file
paths are interfaces; validation fails on unknown fields and unsupported versions.
There is no runtime network request or automatically downloaded schema.

## Task and repository identity

Task IDs use `[A-Z][A-Z0-9]{0,15}-[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*`, with total length
3–80. Examples include `CF-101` and `CF-project-memory-evidence`. IDs cannot contain
slashes, traversal, whitespace, dots or empty hyphen-separated segments.

Repositories use an `owner/repository` identity with alphanumeric first characters
and only ASCII alphanumerics, underscores, periods and hyphens thereafter. The
helper recognizes HTTPS GitHub and `git@github.com:` origin URLs. Without such an
origin, callers must declare the identity explicitly. This is a consistency
check, not authentication of a remote or an authorization decision.

A canonical source record is a regular UTF-8 Markdown blob at
`tasks/ID-description.md`. It has simple single-line frontmatter, an exact `id`
and an inline `repositories: [owner/repository]` list of code repositories.
A claim is a nonempty Markdown heading of level 2–6 containing `claim` or
`subclaim`. Fenced examples, indented code and HTML comments do not create claims.
The claim digest includes exact bytes through the next heading of equal or lower
level. Review its actual text; a syntactically present claim is not authorization.

Registration reads all Markdown records under `tasks/` at a pinned commit and
refuses duplicate selected canonical IDs across every status. A different valid
ID may share a filename prefix only when its own ID agrees with its own path.
Explicit noncanonical redirects require an existing exact canonical destination;
ordinary prose references do not claim identity. Unsupported declared ID syntax
anywhere in the scanned records fails closed; inline comments on ID values are
not supported by this intentionally small frontmatter grammar.

## Immutable source snapshots

Path: `.project-memory/tasks/ID/sources/COMMIT.json`.

A source object has `sourceVersion: 1`, `repository`, `revision`, `id`, `path`,
`blob`, `record` and `identityEvidence`. The revision and blob are full lowercase
40- or 64-character Git object IDs. The exact task text must hash to `blob`.
`identityEvidence` contains the one canonical record and any verified explicit
redirects, with their exact paths, IDs, blobs and source records where needed.

Sources preserve the reviewed import. They are not live task state. Offline
validation checks internal consistency; independent review must verify the
source repository and pinned revision. Repeating an identical registration is
safe. An existing differing snapshot cannot be overwritten.

## Events and runs

Path: `.project-memory/tasks/ID/RUN-UUID/EVENT-UUID.json`.

Events have `formatVersion: 2`, v4 `eventId` and `runId`, `kind`, `parents`,
`sequence`, `task`, `repository`, `branch`, `pullRequest`, ISO `at`, `summary`,
`next` and nonempty unique `files`. A completion additionally has a nonempty
`validation` declaration and an exact `snapshot` object keyed by its files.
Other event kinds cannot contain those completion fields.

The `task` binds source `repository`, `id`, `path`, `revision`, `blob` and a
`claim` containing its exact heading and SHA-256 digest. Code repository membership
must agree with the pinned task record. A PR is null or a GitHub pull URL for the
code repository. Branches record where implementation happened, even after merge.

- `start` begins a new run with sequence 0. A task with no history has no parents.
  A later run references the terminal head it supersedes.
- `checkpoint` and `resume` record a checkpoint in the same run, with one active
  parent and sequence one greater than that parent. They may change scope.
- `complete` or `abandon` terminates a run through one active parent. Its scope
  must be within that parent's scope.
- `reconcile` creates a start referencing every competing head. It preserves all
  previous events and requires a new completion before providing coverage.

Every run has exactly one start. All parent IDs must exist. Cycles, cross-task
edges, changed run identity, successors to terminal events and a single-parent
restart of active work are refused. Multi-parent starts require an antichain:
no parent may be an ancestor of another parent. Timestamps are human context and stale-age
hints, never the ordering rule. Every task must have one frontier for committed
validation, even when competing heads have identical fingerprints.

## File evidence and history preservation

File paths are exact repository-root-relative UTF-8 paths. Traversal outside the
repository, backslashes, drive prefixes, empty components and recognized metadata
self-references are refused. Directories do not recursively cover files.

Fingerprints are `{blob, mode}` from the staged Git index, with modes `100644`,
`100755`, `120000` or `160000`. A deletion has both values null. The helper refuses
unstaged intended changes and unstaged new files, including ignored files.
Completion includes no raw file contents.

Validation resolves explicit base and HEAD using their actual merge base, reads
only committed evidence, and preserves every base metadata path byte-for-byte
and mode-for-mode. A changed non-metadata path needs a newly added final completion
whose fingerprint equals HEAD. Historical completions cannot cover new edits.
Only each task's single frontier contributes; old snapshots are never unioned.
Unrecognized files under `.project-memory/` remain ordinary project paths.

Legacy `.project-memory/tasks.jsonl` is read-only data, whether its rows were
unversioned or version 1. Existing bytes, including line endings and a missing
final newline, remain unchanged. Status labels it as legacy and no v2 coverage.
Adding a new legacy ledger or modifying an existing one fails validation.

Metadata must be regular non-executable files with real directory ancestors.
Symlinks, unsupported UTF-8 and invalid JSON fail closed. Writes publish an
exclusive hard link from a complete temporary file; another writer cannot replace
an existing event. Temporary files and ordinary project files are not trusted as
metadata. Independent tasks do not share an append lock.

## Scope and limitations

All Git invocations deny transports, lazy fetching and replacement-object
substitution. Git 2.43 can still attempt a lazy-fetch subprocess, but its transport
is denied; newer versions also suppress that subprocess. The helper never runs
strings stored in task records, summaries or validation declarations.

This is local evidence, not proof of wall-clock editing order, test execution,
security, current ownership or remote authenticity. Review and ordinary tests
still apply. CI configuration does not create required-status rules. An initial
local source commit and genuine start may bootstrap adoption; they must remain
reviewable and must not be described as already published before publication.
