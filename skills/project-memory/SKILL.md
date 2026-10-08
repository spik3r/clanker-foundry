---
name: project-memory
description: Use when starting, changing, resuming, or completing any Clanker Foundry task. Capture append-only task state before and after changes; CI rejects pull requests without the required evidence.
---

# Project Memory

Use this skill for every task that changes a repository: code, research
artifacts, documentation, infrastructure, configuration, or a new application.

The central backlog remains the source of truth for ownership and claims. This skill records the implementation-level state needed to resume safely.

## Lifecycle

1. Read current state before planning or resuming work:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs status
   ```
2. Before editing, append a start event with every path the task is expected to
   change:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs start \
     --id HT-202-project-memory --summary "Describe intended change" \
     --next "Implement and validate" --files path/to/file
   ```
3. Make the change and run relevant validation.
4. Before committing, append completion evidence using the same task ID and all
   paths changed by the task:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs complete \
     --id HT-202-project-memory --summary "Describe change and validation" \
     --next "Open or merge PR" --files path/to/file
   ```
5. Commit the ledger with the task changes. The CI job validates the PR diff
   against the ledger; run the same check locally when the base is available:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs validate --base origin/main
   ```

Events are append-only in `.project-memory/tasks.jsonl`. Each event records a
task ID, timestamp, summary, next action, and affected paths. Do not record
credentials, access tokens, private account data, or raw logs.

## Resuming after interruption

Run `status`, identify the latest incomplete or most recent task event, read its
`next` field, inspect Git, and continue from durable evidence—not old
conversation context. A completed task's `next` field should say what follow-up
or merge action remains; an in-progress task's `next` field is the immediate
resume action.

## CI contract

The required `project-memory` CI job runs the validator. It fails when any
changed non-memory path lacks a matching task ID with both `start` and
`complete` events, or when a completion was recorded before its start. The
completion event must list that path. This applies to documentation and workflow
changes too. The job itself is protected by the repository's required-status
rule before it can block merging.
