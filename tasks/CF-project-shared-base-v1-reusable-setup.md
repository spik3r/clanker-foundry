---
id: CF-project-shared-base-v1
status: review
owner: Astra implementation agent
repositories: [spik3r/clanker-foundry]
---

# Reusable project-local shared base

## Implementation claim

Implement a project-local setup and read-only verification helper for a reviewed
Git submodule at `.agent-baseline`, a short project-owned instruction overlay and
selected portable skill links. Cover setup, clean clones, updates, customization,
client discovery limits and the distinction from personal Stow installation.
Review nearby instructions for accuracy and maintainability. Add safety regression
tests and CI coverage; run the pack validator and a clean consumer trial.

Scope: `scripts/project-setup.sh`, `tests/test-project-setup.sh`,
`tests/project-memory-workflow.test.mjs` (expected additional CI step),
`templates/project/shared-agents.md`, `templates/project/project-agents.md`,
`templates/project/new-project-checklist.md`, `docs/project-shared-base.md`,
`docs/getting-started.md`, `global-agents/README.md`, `global-agents/AGENTS.md`,
`AGENTS.md`, `README.md`, `.github/workflows/validate.yml`, and this task record.
Preserve all other claims and legacy evidence. No machine-wide installation,
runtime or modeling change, merge, deployment, or protection change is included.

This public-safe local claim records the requested implementation. Publication
requires the coordinator's ownership and authorization check. The implementation
uses an isolated checkout to preserve the active source checkout.

## Acceptance

- Reject wrong pins, invalid skill names, destination conflicts and symlink parents
  before creating links or instruction files; repeat setup without overwrites.
- Verification is read-only and detects missing baseline files or selected links.
- Consumer instructions use only the reusable base, never Foundry's root rules.
- Record observed checks and committed Project Memory evidence for review.

## Implementation result

Added the project-local helper, short overlay, selection manifest convention,
relative skill links and explicit client imports. All destinations are preflighted;
verification checks staged submodule metadata and the gitlink without modifying
the project. Setup preserves existing instructions and custom skills, and rejects
stale links after selection removal. The reuse guide covers initialization,
updates, rollback, custom assets, client discovery and personal-install differences.
Nearby instruction fixes clarify precedence, remove stale Claude import advice and
repair a starter link that broke after copying it into a consumer.

Observed local checks on macOS: pack validator passed (optional `skills-ref` absent),
ShellCheck passed for both new scripts, project setup regression tests passed with
directory spaces, failure-state snapshots, custom instructions, explicit skill
removal/readdition, clean-clone initialization and staged pin updates. Existing
installer/model-resolver regressions passed. The initial combined Project Memory
run had one expected-workflow mismatch after the CI step was added; its exact
assertion was extended without removing existing checks, and the workflow tests
then passed 10/10. The full rerun result is recorded in the completion event.

Live client loading and hosted Linux CI were not exercised locally. The helper
checks file wiring, not instruction semantics or client behavior. No global install,
merge or deployment was performed.

## Deferred review finding

The existing global `wire-global-agents.sh` writes an unescaped absolute Claude
import. A canonical path containing spaces conflicts with Claude's documented
import parsing. This code/documentation mismatch was not exercised in Claude and
is outside this project-local change. The global installer is unchanged. The new
consumer imports use relative paths without spaces; consumer checkout directories
containing spaces are covered by the shell regression tests.

## Frozen-input compatibility correction

PR #93 at `9a15d87982118726d76365bc06253be7501d90e4` passed the hosted
validation, Project Memory conformance, model authoring and orchestration workflows.
The Go workflow failed because builder-demo tests read the live instruction files
against frozen input hashes. A local Go 1.27.3 race-enabled run reproduced failures
in `builderdemo` and `cmd/modelbuilderdemo` before the correction.

The reviewed correction restores `AGENTS.md` and `global-agents/AGENTS.md` exactly
from baseline `c99d96d2ee8515cf6b847f3a009b85b95dd5c558`. All 15 original input
hashes now match. Client discovery and precedence guidance stays in the shared-base
guide and consumer template. The frozen manifests, modeling code, acceptance tests
and historical evidence remain unchanged. A genuine new Project Memory run records
this additive correction without rewriting the original commits; final validation
results are recorded below and in its completion event.

The full Go 1.27.3 race suite passed with a physical temporary directory
(`TMPDIR=/private/tmp/foundry-shared-base-go-tmp`), followed by `go vet ./...`,
`modelc validate` and `modelc check`; modeling files remained unchanged. The first
full local run used macOS's aliased `/var` temporary directory and failed unchanged
path-comparison tests in `builderdemo` and `proposalreplay`, comparing `/var` with
`/private/var`. Targeted builder tests and the full suite passed on the physical
path without changing test code or assertions. This host condition remains visible
in the validation record rather than being reported as an unconditional local pass.

The pack validator, new-script ShellCheck, project setup safety/clean-clone/update
regressions, existing installer/model-resolver regressions and 143/143 Project
Memory/workflow tests passed again. Optional `skills-ref` remains unavailable.
All 15 frozen input hashes match, including the two restored instruction files.

## Reviewed-pin and local-hook correction

The guide's original setup block continued after checkout of an unavailable commit
failed, then ran setup from the default branch and returned success. A regression
executing the actual published block reproduced this before the correction. The
setup and update blocks now run in Bash subshells that stop on failure, require a
full lowercase 40-character commit ID, verify the commit object before checkout,
and compare HEAD with that exact ID before staging or running the helper.

The regressions extract the blocks directly from the shared-base guide and substitute only
the fixture URL and reviewed pin. They cover unavailable and abbreviated pins,
setup at a reviewed commit different from the default tip, a fetched update, and
preserving the previous HEAD and staged gitlink after failed updates. A conflicting
local edit also blocks checkout without running the checker or discarding the edit.

Separately, a local Git `core.fsmonitor` command was observed running during the
helper's status query despite its offline and optional-lock flags. This requires
an already configured local command; the observation does not establish a remote
exploit. Helper Git calls now override `core.fsmonitor=false` for the invocation.
A fixture first proves its local hook runs under ordinary Git status, then verifies
setup and read-only checking suppress it while preserving configuration and Git
metadata. No global settings, frozen inputs or historical evidence were changed.

A genuine new Project Memory run records this continuation. Final observed checks
are recorded in its completion event.
