# Plan — <task>

## Goal and scope

- Outcome: <observable result>
- In scope: <bounded work>
- Out of scope: <adjacent work to leave alone>
- Stop condition: <completed outcome, requested deadline with timezone, or blocker>

## Current coordination

- Authoritative task or claim: <existing record; omit if the project has none>
- Last verified: <time and source revision when available>
- Comparison base: <verified branch or commit>
- Protected changes: <existing staged, unstaged, and untracked work to preserve>
- Integration responsibility: <who may integrate, and authorised actions>

Keep ownership and task lifecycle at the authoritative source. This plan records the
execution graph and links to that source; it is not a second claims ledger.

## Work graph

| # | Task | Worker | Inputs | Output and acceptance checks | Depends on | Writable paths | Execution status |
|---|---|---|---|---|---|---|---|
| 1 | <research> | explorer-researcher | <scope> | findings.md + <evidence> | — | none | <pending> |
| 2 | <plan> | architect-planner | findings.md | plan + <decision> | 1 | <plan artifact only> | <pending> |
| 3 | <build slice> | builder-developer | plan | <change + tests> | 2 | <exact paths> | <pending> |
| 4 | <independent review> | reviewer-verifier | plan + complete diff | review.md + <verdict> | 3 | <review artifact only> | <pending> |

## Scheduling and handoffs

- Parallel: <independent task numbers>
- Sequential: <dependency chain>
- Branch/worktree per writable lane: <verified base and isolated location>
- Overlap or blocker: <owner, dependency, and next action>
- Supersedes: <prior plan and reason, if applicable>

## Evidence and done criteria

- [ ] Every slice produced its stated output and met its acceptance checks.
- [ ] Exact validation outcomes and tested revisions are recorded.
- [ ] Later changes have not invalidated the relied-on checks.
- [ ] Independent review has no blocking findings.
- [ ] Current task/claim state was rechecked before integration.
- [ ] Open risks, blocked work, and unverified areas are reported.
- [ ] Authorised handoffs and task updates are recorded at the existing source.
