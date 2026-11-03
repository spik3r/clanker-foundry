---
id: CF-orchestration-trace-dispatch-v1
status: review
owner: Foundry runtime trace contributor
repositories: [spik3r/clanker-foundry]
---

# Trace assignment, acknowledgement and submission

## Implementation claim

Owner: Foundry runtime trace contributor. Base: main at 634427c. Branch:
codex/orchestration-trace-dispatch-v1. The repository owner requested a bounded
team of changes with PR publication and coordinator review before merge. The
coordinator alone integrates this change. The live open-PR check found only the
separate process-view PR20; that lane and its files are excluded.

Write scope: runtime/model-trace-v1/manifest.json, check.mjs, README.md,
views/state.dot, views/sequence.md, testdata/dispatch-oracle.json,
runtime/test/model-trace-v1.test.mjs, this record and its new immutable Project
Memory source/events. Preserve the original five-case oracle bytes. Do not edit
runtime evaluator, contract, store, runner, providers, CI or operational authority.

## Acceptance

Extend the partial review trace with successful ASSIGN, ACKNOWLEDGE and SUBMIT
through PENDING, ASSIGNED, EXECUTING and VALIDATING. Include wrong-worker,
stale-fence, lease-expiry and invalid-artifact-subject refusals. A separate
contributor freezes invented literal expectations from the existing contracts
before trace implementation; record authorship and its limits. Compare actual
pure evaluator events, effect intents and relevant execution fields with those
literals and check input preservation. No real worker dispatch or execution.

Retain explicit partial/unmapped coverage and deterministic views. Detect dangling
references, fixture drift and stale views. Behavioral guard mutations must fail
the literal comparisons independently of source-pin checks.

## Validation and handoff

Run the focused trace checks, full runtime tests with pinned dependencies,
repository validation, installer regressions, whitespace checks and full committed
Project Memory validation. Record observed results and tool limits. Return the
exact revision and PR to the coordinator for independent review; do not merge.

## Review handoff

The trace now covers successful assignment, acknowledgement and successful
submission, plus nine refusal observations. The original five-case oracle is
unchanged. A separate contributor statically authored twelve new full literal
observations before implementation, frozen in d588cc1. This is separate-assistant
authorship rather than independent-person review or observed execution evidence.

Validation: focused trace tests passed 6/6, including four in-memory behavioral
guard mutation controls. The full runtime suite passed 199/199 with Node 26.0.0
and a locked offline dependency install. The fixed trace checker, pack validator,
installer regressions and whitespace checks passed. Optional skills-ref is absent.
No evaluator, runtime contract, dispatcher, provider or workflow was changed.

The first local registration rejected the task filename without a description
suffix. The corrected source claim was committed and published before genuine
registration/start or implementation. The first fence mutation test retained a
second fence check and did not alter behavior; the control now weakens both fence
checks, and the stale-fence fixture detects the resulting behavioral difference.

Next: independent exact-source review, immutable completion/full-diff validation,
and coordinator review of hosted checks before any merge. This task grants no
managed-runtime authority and launches no real worker.
