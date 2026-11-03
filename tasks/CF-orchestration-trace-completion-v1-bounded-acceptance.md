---
id: CF-orchestration-trace-completion-v1
status: in-progress
owner: Foundry completion trace contributor
repositories: [spik3r/clanker-foundry]
---

# Trace validation, approval, stage acceptance and cleanup

## Implementation claim

Owner: Foundry completion trace contributor; the coordinating lead alone merges.
Base: main at 83007fa3083dc46add5f83e628d2aac30d1a972b. Branch:
codex/orchestration-trace-completion-v1. The owner requested the next bounded
larger tasks after PR20 merged. The coordinator approved this finite slice.
The live open-PR check found separate proposal-report work in PR56; its paths,
authority and other task records are excluded.

Write only runtime/model-trace-v1/check.mjs, manifest.json, README.md,
views/state.dot, views/sequence.md, testdata/completion-oracle.json,
runtime/test/model-trace-v1.test.mjs, this task and its new immutable Project
Memory source/events. Preserve the original five-case and twelve-case oracle
bytes and their checks. Do not change evaluator, runtime policy, contracts,
store, runner, providers, source policies, global instructions or CI.

## Bounded acceptance

Continue the earlier submitted single-review-stage example through successful
RECORD_VALIDATION, RECORD_APPROVAL, ACCEPT_STAGE and CLEANUP_CONFIRMED. Acceptance
reaches COMPLETED while cleanup and resources remain pending; confirmation later
releases them. Compare REVOKE_CAPABILITIES as an inert intent, never deliver it.
Add one nonfinal acceptance observation with a two-stage read-only workflow:
advance to PENDING, retain the accepted subject and pending cleanup, and clear
result, validations and approvals. Do not claim the full second-stage workflow.

The separate contributor freezes five success and thirteen refusal observations
before trace implementation or replay. Refusals cover independent/trusted/bound,
fresh, nonduplicate validation; bound and unexpired approval; checks and valid
approval at acceptance; owner identity; cleanup fencing and pending cleanup.
Compare full execution, exact events/effects or refusal codes/messages, plus
unchanged inputs. Keep all unrelated commands and states explicitly unmapped.

Four in-memory behavioral controls weaken validator independence, required PASS,
revoked-approval filtering and cleanup fencing. They must fail the frozen literal
comparisons independently of source-pin checking. Finite observations and source
links do not establish universal coverage, authenticity or actual operations.

## Validation and handoff

Publish this claim before implementation, record genuine Project Memory start,
and freeze the separately authored corpus before replay or trace changes. Run
focused and full runtime tests with locked dependencies, the fixed trace checker,
pack and installer checks, whitespace and full committed evidence validation.
Obtain independent exact-source review. The coordinator reviews hosted checks
and owns integration. No real validator, approval service or worker is activated.
