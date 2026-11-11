---
id: CF-runtime-terminal-imports-v1
status: in-progress
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
