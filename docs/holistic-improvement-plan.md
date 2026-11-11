# Repository assessment and improvement plan

Reviewed against `origin/main` at `9d665de` on 11 November 2026. This is a
proposal for review, not a new implementation backlog or a claim on downstream
work. Recheck the base, open pull requests and task ownership before starting any
item. The [current direction](current-direction.md) remains the decision gate for
consumer work.

## What the repository has now

| Area | Observed state | Practical limit |
| --- | --- | --- |
| Agent pack | Nineteen skills, four agent profiles, reusable workflows, templates and checklists. Personal Stow installation and a pinned project setup are documented. | A link or instruction file does not prove that a client loaded the intended context. The [getting-started guide](getting-started.md) asks users to confirm that separately. |
| Model tools | The `lifecycle/v1` module offers a bounded task lifecycle, review, replay and revision tools. A separate [JSON boolean generator](../modeling/jsonvalidator/README.md) emits Go and TypeScript predicates. | The lifecycle path uses fictional task data. The JSON generator accepts already parsed values and returns a boolean, with no application mapping or error details. Neither is an integrated consumer feature. |
| Runtime | The optional [orchestration runtime](../runtime/README.md) has contracts, a store, supervised worker paths and tests. | The [rollout guide](deterministic-orchestration-rollout.md) keeps the sample namespace inactive. Source and CI evidence do not establish an operating service. |
| Checks | Six GitHub workflows cover pack validation, Project Memory, the lifecycle module, model authoring, the JSON generator and orchestration. The pack workflow runs installer and project setup regressions; the Go workflows use race tests. | These checks cover repository contracts and examples. They do not measure a consumer's outcome or prove client loading, browser behavior or unattended runtime operation. |
| Change history | Canonical task files and immutable Project Memory events record scoped work. | They are evidence of claims and declared checks, not ownership locks or independent proof that a check ran. |

The repository's strongest asset is the usable skill pack. The optional model and
runtime paths have unusually explicit limits and failure cases. The main gap is
not a missing generic framework feature: no maintained consumer, real task,
decision owner, success threshold or baseline has been selected for the proposed
model-to-code outcome. The [current direction](current-direction.md) already
records this. More tools or reports would make the gap harder to see.

## Recommended order

### 1. Choose one real consumer job and a stop rule

**Recommendation:** the owner picks one maintained repository and one bounded
change that somebody needs. Record who accepts the result, the current workflow's
time and review cost, exact expected and refused examples, and a maximum acceptable
overhead. Choose ordinary engineering, agent-written code from a reviewed model,
or generated supported code based on that job. Keep requirements the supported
profile cannot express explicit.

**Why first:** the current plan names the lack of consumer evidence, and the
fictional builder and lifecycle results cannot answer it. This is an owner
decision, not a coding task to assign by inference.

**Done when:** a reviewed brief names the user, repository, feature, decision
owner, independent expectation author, route, baseline, benefit threshold and
stop rule. If a core need is outside the profile, choose ordinary engineering or
approve a separately scoped extension; do not silently omit it.

### 2. Observe a cold setup and one complete operator path

**Recommendation:** ask a person who did not write the guide to install either
the personal pack or the pinned project setup in a clean environment, confirm
which instructions and skills the client loaded, then follow one public
model-review path if that path fits the chosen job. Record commands, elapsed
time, failures and where author help was needed. Fix observed navigation or setup
faults in small changes.

**Why next:** [getting started](getting-started.md) and the
[project setup guide](project-shared-base.md) describe distinct installation
routes and explicitly leave client loading to an operator check. The
[model workflow](model-driven-current-workflow.md) has several prepared inputs,
reports and separate execution decisions. An independent operator will show
which steps need work; another wrapper has no evidence yet.

**Done when:** a non-author can repeat the selected path from a clean checkout
using the guide alone, with the exact revision, tool versions, interventions and
unavailable checks recorded. Do not label a fictional replay as consumer value.

### 3. Put one maintained feature through a requirement change

**Recommendation:** after the consumer contract is approved, integrate one
feature through the selected route and test it at its real call site. Freeze
expected results independently before candidate execution. Then change one
meaningful requirement, identify stale model, generated code and test evidence,
update the maintained feature, and rerun acceptance and existing regressions.

**Missing functionality to resolve from that trial:** the JSON generator does
not parse transport data, map application types or return diagnostics; the
lifecycle profile does not supply general persistence, authentication, UI or
concurrency. Add only the adapter or profile behavior the chosen feature needs.
For generated output, define a pinned generator revision and a drift check in
the consumer. These are candidate tasks, not blanket requests to expand either
tool.

**Done when:** the real user action works in the maintained repository, a
reviewer can trace model and expected literals to the implementation, and the
requirement revision causes stale evidence to be detected before reuse. The
decision owner accepts or rejects the result against the brief.

### 4. Measure the second use and simplify what costs time

**Recommendation:** have a different operator make a second agreed change.
Compare actual effort, review findings, errors and support interventions with
the baseline from step 1. Remove repeated manual steps only where the record
shows cost or mistakes. Keep independent expected results and execution decisions
visible.

**Done when:** the owner can decide whether the approach met the agreed benefit
and overhead thresholds. If it did not, simplify, narrow or stop the model path
instead of adding another layer of reports.

## Parallel maintenance with bounded scope

These items need separate reviewed claims and should not delay the consumer
decision unless they block it.

1. **Current status and navigation.** Reconcile the dated
   [current direction](current-direction.md), [runtime rollout](deterministic-orchestration-rollout.md)
   and [JSON generator guide](../modeling/jsonvalidator/README.md) with merged
   source and current CI. For example, the generator guide still calls the
   producer a draft after its code reached main. Keep historical evidence dated;
   show implemented, tested, released and activated states separately.
2. **Install and validation feedback.** Use the cold operator record to decide
   whether the personal and project setup paths need a shorter decision guide,
   clearer failure recovery or client-loading checks. Keep the existing
   conflict-preserving behavior. Open PRs #102 and #104 already address agent
   selector and model-map validation; review their outcome before proposing
   overlapping fixes.
3. **Runtime upkeep.** Fix observed correctness regressions and keep its operating
   guide accurate. Defer new dispatch features or activation until a consumer
   dependency and a separate owner decision justify them. Do not treat the
   inactive namespace template as permission to run managed work.
4. **Release and compatibility policy.** If a consumer adopts the JSON generator
   or shadow package, define the pin, supported API surface, compatibility test
   and update process in that consumer's PR. The root npm package is marked
   private and the generator guide states that no producer release exists; a
   repository merge is not a distribution plan.

## Trade-offs and review questions

- A broad generator or runtime expansion could make more examples pass, but it
  adds maintenance before a real need is known. The recommended route is a small
  consumer trial with a pre-agreed stop rule.
- Personal Stow links follow the moving checkout and are easy to update. The
  project submodule pins a reviewed revision but requires explicit updates and
  client-context checks. Retain both routes; choose per user and document the
  observed setup cost.
- Detailed evidence makes failures inspectable, but more reports increase review
  work. Measure whether reviewers can make a better decision from each artifact
  before adding another one.

The owner needs to decide the first consumer job, who accepts it, and the
benefit/overhead thresholds. Each later implementation should get its own claim,
tests and PR. This assessment changes no runtime policy or product behavior.
