# ADR 0002: Managed execution behind explicit authority and host checks

Status: accepted for local implementation; operating authority requires its own pinned decision.

## Context

The read-only runtime separated model output from task decisions, but left provider
transport, repository writes and implementation evidence to embedding code. Adding
writes without those boundaries would turn a typed workflow into an unchecked host
callback.

## Decision

Keep the evaluator pure. Add `isolated_write` and `isolated-write-v1` only behind a
persisted managed namespace, a genuine Project Memory binding and an exact
per-attempt host proof. Preserve read-only legacy operation outside managed
namespaces. The user selected new `CF-AUTO-` tasks only; existing `CF` records and
claims do not transfer.

Persist namespace decisions and reviewed admissions in the same SQLite authority
as executions. Verify their source bytes against actual local Git objects. A
source digest alone does not establish ownership. Reject overlapping prefixes,
stale generations, unreviewed admission and late evidence binding. Draining blocks
new dispatch and retains unknown outcomes and reservations until resolved.

Keep credentials in the provider controller. Direct Responses requests carry no
tools or prior response identity. The signed-in Codex CLI is a supervised text
transport with explicit integration disables, not a hard tool-isolation boundary.
Reject tool events and unknown diagnostics.

Write only to a private candidate workspace in a constrained Docker container.
Mount no repository metadata, home, credentials, database or Docker socket into
the worker. Accept exact typed file operations against pinned before-digests.
Promotion is a separate owner action bound to accepted artifacts and unchanged
host files; it does not stage or commit changes.

The evidence bridge binds a real source and start before dispatch. It calls the
unchanged Project Memory helper after observed validation and staging. A private
journal records completion intent before invoking the helper and reconciles an
exact immutable child after a lost reply. Runtime completion alone supplies no
repository implementation coverage.

## Alternatives and limits

A fully autonomous provider process with broad repository access would require a
larger sandbox and credential boundary. Structured plans keep the writable worker
small, at the cost of limiting edits to bounded text files and trusted operations.

SQLite coordinates one host using one database. Copied databases and independent
manual writers require an operating agreement; a local UUID cannot fence them.
Docker isolation protects the declared mounts and process boundary, not a hostile
host administrator or a compromised Docker daemon. A provider request can produce
different text on repeat runs; only lifecycle decisions and accepted transitions
are deterministic.
