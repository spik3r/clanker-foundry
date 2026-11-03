---
id: CF-orchestration-trace-completion-v1
status: done
owner: Foundry completion trace contributor
repositories: [spik3r/clanker-foundry]
---

# Trace validation, approval, stage acceptance and cleanup

## Integration result

[PR57](https://github.com/spik3r/clanker-foundry/pull/57) merged as
`a13bc2747a049ba2d94b358669356e1d3ee93c6d` after independent source review
and all five hosted workflows passed on
`b7ce60f562a4201cfa55c0c75505e34ae9175547`. The lead independently passed
the nine focused trace tests, checker and committed evidence validation. The
completed claim reserves no further work. The trace remains a bounded review aid;
its effect intents were not delivered.

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

## Review handoff

The separate contributor statically authored the eighteen new observations and
froze them unchanged in 87466d8 before implementation or replay. The original
five-case and twelve-case files remain byte-identical. The trace checks the full
frozen post-states, events, inert effect intents, refusals and unchanged inputs.
The nominal success inputs connect to the original submitted execution; the
nonfinal stage case remains a separate bounded observation.

Validation: focused trace tests passed 9/9, including the original checks and
four new behavioral mutation controls. The full runtime suite passed 202/202 on
8ed0a74 with Node 26.0.0 and locked offline dependencies. An ordinary merge of
main d040257 produced 18f3a35 and changed only prior task records/events; runtime
bytes stayed identical. Focused tests, trace checking, pack and whitespace checks
passed on the merged source. Installer regressions also passed. Optional
skills-ref is unavailable. No runtime evaluator, policy, schema or CI changed.

A separate reviewer checked exact source 18f3a35, independently passed the focused
checks and found no remaining issue. This is separate-assistant review, not a
human approval or proof of universal evaluator coverage. Final acceptance and
resource release remain distinct; no effect was delivered and no actual check,
approval service or worker was invoked.

Next: genuine immutable completion, full committed evidence validation, then
coordinator review of hosted checks before integration.
