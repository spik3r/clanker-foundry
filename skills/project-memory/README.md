# Project Memory v2

A dependency-free Node/Git helper for repository-local implementation continuity.
Read the [workflow](SKILL.md) and [format contract](FORMAT.md) before use.

New work uses immutable source snapshots and one JSON file per task/run event.
Explicit parents define order. Conflicting frontiers require reconciliation and a
new completion. There is no shared append file, mutable latest pointer, ownership
lock or runtime download.

Completion records bind declared checks to exact staged Git blobs and modes.
Committed validation rejects changed history, stale coverage, incomplete runs,
identity mismatches and malformed metadata. It handles literal UTF-8 paths,
renames, deletions, mode changes, symlinks, refactors and gitlinks. It checks
content and declarations; it cannot prove that tests ran or that the task still
has the same owner.

Register a reviewed canonical task from a pinned local Git checkout. The source
may be this repository or a separate task repository. Code and source identities
are explicit or checked against a recognized local GitHub origin; validation
always requires the source identity. There is no hardcoded consumer repository.
Only public-safe task bytes may enter a public repository.

Use Node 24 and Git 2.43 or later. Every helper Git process denies all transports
and replacement objects. Missing objects fail without fetching. Newer Git also
suppresses lazy-fetch subprocesses. No package installation, credentials or
repository write token is needed for validation.

Run the tests:

`node --test tests/project-memory.test.mjs tests/project-memory-workflow.test.mjs`

Foundry's `project-memory` CI job runs those tests on pushes, pull requests and
manual runs. Pull requests also validate the full committed diff using the real
PR base SHA and explicit code/source repository identities. The source here is a
reviewed task committed in Foundry itself. Consumers with another canonical task
repository must configure its identity in their own workflow.

The v1/unversioned JSONL ledger stays byte-preserved and read-only. It cannot
satisfy new v2 coverage. No historical events are invented or migrated. Upgrades
are reviewed local changes, never runtime synchronization. Required-status rules
remain the repository owner's separate decision.
