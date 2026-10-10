---
name: project-memory
description: Use when starting, resuming, handing off or completing repository changes. Link a reviewed canonical task claim to immutable implementation events and validate fresh committed evidence. Read-only lookups do not need events; task records remain authoritative for ownership.
---

# Project Memory

Keep repository-local implementation history: intent, exact paths, declared
checks, results and the next action. A canonical task owns identity, ownership
and claims. This helper does not acquire a lock, authorize work or change a claim.
Read the real task and check overlapping work before editing.

A pinned snapshot records the task reviewed at that revision. Its hash checks
stored bytes, not external truth or current ownership. Independently compare its
repository, revision, task and claim with the actual canonical source.

Never import confidential records into a public repository. Do not record
credentials, private conversations, personal account data or raw logs. Import
only the reviewed task, not a whole backlog. Summaries and validation strings are
inert declarations; the helper cannot prove that a command ran.

## Prerequisites and source contract

Use Node 24 and Git 2.43 or later. The helper uses only local files and Git
objects. It never fetches tasks, code or updates. Every Git process denies all
transports and replacement objects. Missing objects fail; newer Git also
suppresses lazy-fetch subprocesses. See [the format contract](FORMAT.md).

A source can be a separate task repository or the code repository itself. It
must contain a committed, public-safe Markdown record at
`tasks/TASK-ID-description.md`, with simple frontmatter and a nonempty claim:

```markdown
---
id: CF-101
status: in-progress
owner: Implementation owner
repositories: [example/application]
---

# Implement the reviewed change

## Implementation claim

Owner, exact scope, intended result and any exclusions agreed in the task.
```

Task IDs have an uppercase project prefix and a bounded alphanumeric slug, such
as `CF-101` or `CF-project-memory-evidence`. The filename and frontmatter ID must
agree. Duplicate canonical IDs across all statuses, including closed tasks,
block registration. Explicit noncanonical redirects must name an existing exact
canonical record. A prose mention of another task is not its identity.

## Register and start

Examples assume this skill is installed at `skills/project-memory`. For a global
installation, use the installed script's full path inside the target repository.
All `--files` values identify exact repository-root-relative files.

1. Read the current canonical task, coordinate a bounded claim and commit that
   claim in its source repository before implementation.
2. Register the reviewed local source revision:
   `node skills/project-memory/scripts/project-memory.mjs register --id CF-101 --source-repo ../task-source --ref REVIEWED-COMMIT --record tasks/CF-101-description.md --source-repository example/tasks`
3. Independently verify the printed revision, source snapshot and claim. A
   declared repository identity is not proof of remote authenticity.
4. Inspect existing work:
   `node skills/project-memory/scripts/project-memory.mjs status`
5. Record a genuine start before editing:
   `node skills/project-memory/scripts/project-memory.mjs start --id CF-101 --repository example/application --files path/to/file --summary "Intended change" --next "Implement and validate"`

`--source-repo` is a local Git checkout path. `--source-repository` is its
`owner/repository` identity. The source identity may be inferred from a recognized
GitHub `origin`; an explicit mismatch is rejected. The code identity similarly
uses `--repository` or its recognized `origin`. There is no private or global
repository default. For validation the source identity is always explicit.
`--backlog` and `--backlog-repository` remain aliases; do not pass both aliases.

Start selects the sole registered source and claim. If several exist, choose
`--source PINNED-COMMIT` and `--claim "Exact claim heading"`. No latest-source
file is authoritative. The branch comes from Git; detached starts require
`--branch IMPLEMENTATION-BRANCH`. Add `--pr VERIFIED-PR-URL` when known.
Before a PR exists, that field is null.

## Resume, complete and commit

An active task uses `resume` or `checkpoint`, with `--id`, `--summary` and `--next`.
To change scope, include `--files` with the full new scope. Task/source/claim,
code repository and origin branch cannot change during a run. A PR link may be
added once; it cannot then be swapped.

1. Run relevant checks against the final code, tests and documentation.
2. Stage only the task's explicit changed paths, including deletions. Preserve
   other workers' staged changes.
3. Record completion:
   `node skills/project-memory/scripts/project-memory.mjs complete --id CF-101 --summary "Observed result" --validation "Commands and observed results" --next "Review PR"`
4. Stage the new source/event files printed by the helper and commit them with
   the implementation. Do not edit existing metadata files.
5. Check the full committed diff:
   `node skills/project-memory/scripts/project-memory.mjs validate --base origin/main --repository example/application --source-repository example/tasks`

Completion fingerprints staged Git blobs and modes. Unstaged or untracked
intended files are refused. Deleted paths have null fingerprints. Include both
sides of renames, the old file and every changed child for file-to-directory
changes, and removed children plus the new file for the reverse. Directories do
not cover children. Valid UTF-8 paths stay literal; unsupported bytes fail closed.
Gitlink changes remain visible even when Git is configured to ignore submodules.

If content changes after completion, start a new run, rerun checks, stage final
content and complete again. Fresh coverage is based on the actual merge-base and
committed HEAD, not a timestamp or an older completion of the same path.

`abandon --id CF-101 --summary "Why work stopped" --next "Owner action"` closes
local work without coverage or changing the canonical claim. Starting after
completion or abandonment creates a new run linked to the previous head.

## Conflicts and interruptions

Each event is immutable under
`.project-memory/tasks/TASK-ID/RUN-UUID/EVENT-UUID.json`. Each source snapshot is
immutable under `.project-memory/tasks/TASK-ID/sources/COMMIT.json`. UUIDs and
causal links are generated automatically. There is no shared mutable index,
latest file, append file or ownership lock.

Status reports every task frontier. Multiple heads are unresolved, including a
completion plus an active or abandoned head. Review the work and actual claim,
then explicitly reconcile:

`node skills/project-memory/scripts/project-memory.mjs reconcile --id CF-101 --repository example/application --files current/path --summary "How the competing work was reconciled" --next "Revalidate final content"`

Reconciliation starts a new run linked to every current head. It preserves all
history and supplies no coverage until completion. Its scope may omit reverted
or dropped paths; old fingerprints are never unioned into new coverage. Missing
parents, cycles, cross-task links and malformed records fail rather than choosing
a timestamp winner.

`status --stale-days 7` flags old active work for review without abandoning it or
releasing its claim. One ordinary active task does not lock unrelated work.

## Validation, CI and adoption

Validation preserves all merge-base source/event files and legacy ledger bytes
and modes. Only a task's single frontier can supply coverage, through a newly
added completion that matches HEAD. Unknown files under `.project-memory/` are
ordinary project files and still require coverage.

Pass the actual PR base and code/source identities in CI. `--branch` reports the
integration context; historical event branches remain origin provenance after
renames or merges. Never suppress failures, guess a fallback base, weaken other
checks or change protection settings as part of adoption. A workflow becomes a
merge requirement only when repository rules require its status.

Existing `.project-memory/tasks.jsonl`, including v1 and unversioned rows, remains
byte-preserved read-only history. Status labels it as providing no v2 coverage.
Do not append, rewrite, delete, backfill or silently migrate it. Re-register a
reviewed task and make a genuine new v2 start for new implementation work.

For a new source in the code repository, commit the reviewed task record first,
register that pinned local commit, and record a real start before implementation.
A separately prepared local helper can bootstrap adoption. Retain the source
commit and chronology; do not backdate events or present a local proposal as an
already published task. This bootstrap itself does not authorize publication.
