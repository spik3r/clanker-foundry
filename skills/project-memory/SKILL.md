---
name: project-memory
description: Use when starting, changing, resuming, or completing a Heisentick task. Capture append-only state before and after work and satisfy the CI lifecycle gate.
---

# Project Memory

Use this skill for every task that changes a Heisentick repository: code, research artifacts, documentation, infrastructure, or a new application.

The central backlog remains the source of truth for ownership and claims. This skill records the implementation-level state needed to resume safely.

## Lifecycle

1. Read current state before planning:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs status
   ```
2. Before editing, append a start event:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs start \
     --id HT-202-project-memory --summary "Describe intended change" \
     --next "Implement and validate" --files path/to/file
   ```
3. Make the change and run relevant validation.
4. Before committing, append completion evidence:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs complete \
     --id HT-202-project-memory --summary "Describe change and validation" \
     --next "Open or merge PR" --files path/to/file
   ```
5. Validate the PR diff against the ledger:
   ```bash
   node skills/project-memory/scripts/project-memory.mjs validate --base origin/main
   ```

Events are append-only in `.project-memory/tasks.jsonl`. Do not record credentials, access tokens, private account data, or raw logs.

## Resuming after interruption

Run `status`, identify the latest incomplete or most recent task event, read its `next` field, inspect Git, and continue from durable evidence—not old conversation context.

## CI contract

The PR policy job runs the validator. It fails when any changed non-memory path lacks a matching task ID with both `start` and `complete` events, and the final event must list that path. This applies to documentation and workflow changes too.
