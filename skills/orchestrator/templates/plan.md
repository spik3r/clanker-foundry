# Plan — <task>

## Goal

<one sentence describing the observable outcome>

## Work graph

| # | Task | Worker | Inputs | Output | Depends on | File ownership |
|---|---|---|---|---|---|---|
| 1 | <research> | explorer-researcher | <scope> | findings.md | — | read-only |
| 2 | <plan> | architect-planner | findings.md | approved plan | 1 | read-only |
| 3 | <build slice> | builder-developer | approved plan | <change> | 2 | <paths> |
| 4 | <independent review> | reviewer-verifier | plan + diff | review.md | 3 | read-only |

## Scheduling

- Parallel: <independent task numbers>
- Sequential: <dependency chain>

## Done criteria

- [ ] Every task produced its stated output.
- [ ] Validation evidence is recorded.
- [ ] Independent review has no blocking findings.
- [ ] Open risks and unverified areas are reported.
