---
id: CF-builder-preparation-guide-v1
status: in-progress
owner: Foundry builder preparation guide contributor
repositories: [spik3r/clanker-foundry]
---

# Separate trusted tool and frozen input roots

## Implementation claim

Repair the existing builder-preparation walkthrough from main
2630845592b1b0fe28d7e2b862ce8da608501fe0. The guide uses one root both to
build the preparer and supply its 15 frozen inputs. Historical input revision
882b5002bd4dc70b4be0060e5e5668d1faa8bbdd lacks the build target; current
main has the tool but correctly refuses its changed root instructions against
the unchanged manifest. This is documentation friction, not a production defect.

Exact maintained scope:
- examples/model-builder-go/preparation.md
- tasks/CF-builder-preparation-guide-v1-split-source-roots.md
- New immutable Project Memory source/events for this task

Use separate explicitly authorized TOOL_ROOT and INPUT_ROOT locations. Name the
reviewed tool revision above and the original frozen input revision. Build in
TOOL_ROOT/modeling/lifecycle and pass INPUT_ROOT through existing --root. Update
all related references and the count of absolute host paths. Require an existing
authorized private output parent outside both checkouts. Historical instructions
are frozen exercise data, never current operational authority. Preserve current
instructions, live claims, target-state review, independent review, Q_use, visual
and real-use gates, and every no-dispatch/no-candidate/no-execution boundary.

No production code, API, fixture, test, manifest, other frozen input, AGENTS,
dependency, workflow, runtime/provider/namespace, other task or historical record
changes. No new tool, wrapper, tutorial, automatic checkout, search, fallback,
pin refresh, cleanup or retry. Do not substitute test-only instruction snapshots.
No builder candidate is written, compiled or executed, and no builder is invoked.

Before registration or start, obtain independent coordinator review of this claim
and resolve any live overlap. Commit the reviewed source, verify its exact bytes
and source revision, then register and record a genuine start before guide edits.
Preserve the original guide and command block privately. Historical complete
claims do not confer current ownership; recheck main and open PRs before publishing.

## Prospective checks and handoff

Freeze expectations before editing. In isolated private scratch, run the original
block with explicit current-root paths: expect exit 1, a blocked receipt with the
root instruction mismatch and no preparation workspace. At the named historical
root, expect the original build to fail because cmd/modelbuilderdemo is absent,
before any preparation invocation. Retain both actual failures. These informed
route checks are not cold-operator trials or measured usability evidence.

Run the exact corrected block with only explicit host paths substituted in one
strict Bash session, using reviewed tool source, exact historical input source,
offline native Go 1.27.3 and private cache/temp/output outside both checkouts.
Expect exit 0, prepared-draft, three matched complete reports, 15 byte-identical
input copies, five exact module copies, executionStatus not-run,
reviewStatus not-established and authority none. Check sources remain unchanged
and no candidate or executable exists in the prepared workspace. In separate
private scratch replace only the historical root instruction bytes with current
bytes: expect exit 1, an informative blocked receipt and no workspace creation.
Do not reuse stale outputs or change expectations to force a pass.

Keep inputs.json byte-identical at SHA-256
6a87f4db69a9c43376b405aca0d1252a662a3a4cdc7496329f2b41fac2896032,
all 15 original pins unchanged, and the historical root AGENTS input at SHA-256
4574ee25fe89b6658ae237f3f50648ef1e542c487a2c23d7ab2833696f9455eb.
Run existing focused builderdemo/modelbuilderdemo race tests and vet, repository
validation, documented block shell syntax, whitespace checks and full committed
Project Memory validation. Report unavailable installer/optional checks honestly;
hosted CI remains separate. Retain failures and concise actual author observations
without claiming independent, visual, cold-trial or real-use acceptance.

Publish only the tested complete final branch as a draft PR. Verify remote
head/tree, committed evidence and clean worktree. Fresh independent review and
exact-head hosted CI remain separate gates; this implementation phase does not
mark Ready, merge, deploy, change held PRs or enable any managed runtime.
