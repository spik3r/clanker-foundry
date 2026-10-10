---
id: CF-001
category: tooling
status: in-progress
owner: Foundry implementation
repositories: [spik3r/clanker-foundry]
depends_on: []
credentials: none for local implementation or validation
---

# Reusable project-memory v2 and engineering safeguards

## Purpose

Upgrade the standalone project-memory skill from a shared v1 JSONL writer to
immutable per-task and per-run v2 events with reviewed canonical-source binding.
Keep the skill usable with explicitly configured public-safe task sources and
without network access during normal operation or validation.

Add two focused generic skills for browser/UI verification and bounded persistent
iteration. Improve existing orchestration, review and engineering guidance instead
of adding overlapping skills or project-specific workflows.

## Local implementation claim

Owner: Foundry implementation. The repository owner requested this local producer
upgrade. Public publication is a separate approval; this record does not authorize
publishing, merging, installing globally or changing permissions.

Static scope:
- `skills/project-memory/SKILL.md`, `README.md` and `FORMAT.md`
- `skills/project-memory/scripts/project-memory.mjs`
- `skills/project-memory/scripts/lib/state.mjs` and `lib/git.mjs`
- `tests/project-memory.test.mjs` and `tests/project-memory-workflow.test.mjs`
- `.github/workflows/validate.yml`, only lifecycle tests and validation context
- `skills/ui-verification/SKILL.md` and `README.md`
- `skills/persistent-iteration/SKILL.md` and `README.md`
- `global-agents/AGENTS.md`
- `skills/engineer/SKILL.md`
- `skills/orchestrator/SKILL.md` and `templates/plan.md`
- `skills/review-change/SKILL.md`
- Root `AGENTS.md` and `README.md`
- This task record and its genuine immutable v2 source/event metadata

Preserve unrelated files, existing v1 ledger bytes, workflow permissions and
installer behavior. Do not import confidential task records, source snapshots,
events, account data, logs, credentials or domain-specific workflows. Public-safe
generic material and invented test fixtures are the only input to this candidate.

## Acceptance

- New events use the documented v2 format; old v1/unversioned history remains
  read-only and cannot satisfy fresh v2 completion coverage.
- Explicit source/code identity, canonical task uniqueness, immutable source/event
  files, causal frontiers and actual merge-base/HEAD coverage are preserved.
- Normal use and validation do not fetch code, task records or missing Git objects.
- Unsafe paths, symlinks, malformed metadata, unknown versions, stale coverage and
  conflicting history fail clearly without rewriting history.
- Reusable skills preserve user authorization, local constraints and existing work.
  Instructions and checks are not represented as proof of universal compliance or
  required branch-protection settings.

## Validation and integration

Run the pack validator, installer regressions where GNU Stow is available, all
relevant memory/workflow regressions, JavaScript syntax checks and whitespace
checks. Record actual commands, counts and limitations; do not treat an unrun check
as passed. Review the exact final diff independently for correctness and public
data safety, then record fresh completion against final staged content.

This source record is initially a local bootstrap commit. Verify its bytes and
Git identity before registering it. Public publication, if later approved, must
retain the real source commit and genuine implementation chronology. No automatic
consumer rollout, release, merge, global installation or permission change follows.
