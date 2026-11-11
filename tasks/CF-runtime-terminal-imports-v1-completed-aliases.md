---
id: CF-runtime-terminal-imports-v1
status: review
owner: Foundry importer maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Reject completed task aliases during proposal import

## Implementation claim

Owner: Foundry importer maintenance contributor. The integration lead assigned
this bounded correction on `fix/runtime-terminal-task-imports-20261111`, from
verified main `106fcd58c59a8545e207934d2e24595a63d54984`. Historical pilot and
review-follow-up claims explicitly reserve no further implementation work.
The live open-PR check found held PR101 and another owner's PR110 for CLI
special-file refusal. Neither owns the importer paths below. Do not modify,
review or integrate either PR under this claim. Recheck main and ownership before
publication. This local source is not a published reservation or runtime authority.

Exact scope:

- `runtime/src/import-export.mjs`
- `runtime/test/import-export.test.mjs`
- `runtime/README.md`, only the task-import status and authority explanation
- This task and its new immutable Project Memory source/event files

Reject the existing exact terminal aliases `complete` and `completed` with
`SOURCE_NOT_ACTIONABLE`, alongside `done`, `closed` and `cancelled`. Preserve
`in-progress` and `review` proposal imports, identity and provenance checks,
exactly-one-claim requirements, caller nonmutation and proposal-only output.
This fixes false actionable-looking shadow proposals, not an admission bypass:
managed admission already independently requires current `in-progress` status.

Do not introduce a new status taxonomy or normalize unknown/case-variant
statuses. Do not rewrite existing tasks, claims or history; change authority,
schemas, dependencies, workflows or runtime activation; invoke providers; or
touch the held Flexgo prototype or Rhizome process. Publish a qualified draft
only. Fresh independent patch review and full hosted CI remain separate gates;
the integration lead owns any Ready transition or merge.

## Frozen expectations and validation plan

On unchanged main, the existing import/export tests passed 3/3 on Node 24.22.1.
A read-only synthetic reproduction using the existing source/spec fixture
produced proposals for `complete` and `completed`. The intended literal outcomes
are refusal with `SOURCE_NOT_ACTIONABLE` for `done`, `closed`, `cancelled`,
`complete` and `completed`, and `mode: PROPOSAL` for `in-progress` and `review`.
The other five outcomes already matched. These are implementation-authored
expectations, not a claim of independent test authorship.

Review this narrow source separately, commit it, register its exact local
revision and record a genuine start before implementation or regression edits.
Freeze focused tests before changing production code and observe the two alias
failures against the old importer. Keep those test bytes unchanged thereafter.
Run focused import/export, CLI, authority and consumer-shadow tests, the full
Node 24 runtime suite, repository validation, whitespace checks and full committed
Project Memory validation against current main. Use invented fixtures and a
task-local npm cache. Record only observed checks and preserve the real local
source/start chronology when publishing the completed series.

## Observed result

The maintained guard adds only `complete` and `completed` to the existing exact
terminal-status refusal. The intake guide lists all five refused statuses and
keeps the separate managed-admission requirement explicit. No authority code or
existing task/history changed.

After separate claim review, source object
`49d8c9dfd32587e0a037956ee7436f13867bb75f` was created through the GitHub connector,
fetched and verified against the reviewed tree before registration and the real
start. No branch ref was published at that stage. Local source preparation
`00018f0988ff0f3e42f9a6fd13e606d9f81a16e7` is not remote ancestry.

The frozen importer tests have SHA-256
`ad904a96f854024600d0e0148d3dfe4c58ca7c3760d08527a4435de6b4aa6a59`.
They first failed only `complete` and `completed` against the old importer
(7 passed, 2 failed). Those exact test bytes passed after the guard change.
The unchanged full runtime baseline passed 315/315 tests.

Observed final checks on Node 24.22.1:

- Focused importer, authority, CLI and consumer-shadow tests passed 35/35.
- Full runtime tests passed 321/321, with no failures, skips or cancellations.
- Project Memory and workflow regressions passed 143/143.
- Repository validator, JavaScript syntax and whitespace checks passed.
- Validator regressions passed 47 model-map and 92 path cases.
- Project shared-base safety, clean-clone and update regressions passed.
- Installer prechecks passed nine wiring-target and 16 ordering cases, but the
  installer suite then stopped because GNU Stow is unavailable. Full installer
  acceptance is not claimed. Optional `skills-ref` is also unavailable.

No live provider, Docker isolation or runtime activation was run. Hosted checks
remain separate from these local results. Final content is complete for review;
this claim reserves no further implementation edits. Next: commit fresh immutable
completion, validate the full committed diff, publish the completed draft series,
and obtain fresh independent review and exact-head CI before any integration.
