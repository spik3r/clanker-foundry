# model-driven-design

Helps a person and an agent turn an idea into a reviewable first-model proposal,
without skipping the choices that belong to the person or allowing generated
code to define its own expected results.

The default is read-only clarification. An authorized draft can add a brief,
strict model and independently checked expectation proposals in a chosen workspace.
An authorized review-preparation step uses existing tools for validation and draft
views; it does not authorize candidate execution, publication or deployment.

## What changes in the workflow

Instead of jumping from a prose plan to code, the skill makes four things explicit:
profile fit and non-goals; supplied decisions versus unanswered questions; separately
authored expected behavior; and the difference between model validity, byte
consistency, actual execution and human approval. It does not add a model language,
validator, compiler or runner. AI drafting and model correctness are not deterministic.

[SKILL.md](SKILL.md) is the procedure. The
[profiles reference](references/foundry-profiles.md) covers choosing among the
registered Foundry profiles from a brief, and the
[gap report template](assets/gap-report.md) records a requirement no profile can
express. The
[Foundry reference](references/foundry-handoff.md) routes current lifecycle/v1 work
to the existing profile catalog, strict intake/model check, suite comparison,
generator and qualification tools. It requires an existing trusted Foundry checkout;
stowing this skill alone does not install those tools.

## Worked proposal

The [invented brief](examples/brief.md), [model](examples/model.json),
[independent suite](examples/suite.json), [manifest](examples/manifest.json),
[intake declaration](examples/intake.json) and [handoff](examples/handoff.md) form
a small first-design example. It is not a real human approval or application.
The unresolved use/authorization question remains visible even when model and
binding checks pass. The suite is unexecuted by this handoff.

A Go integration test in `modeling/lifecycle/authoringhandoff` checks the frozen
example through existing APIs, including negative controls. It neither runs the
example reducer nor automates human/visual review. Independent forward-tests of the
skill are separate evidence of behavior on those prompts, not a guarantee for all
future requests. The exact fixture identities are in
[inventory.json](examples/inventory.json).

## Example requests

```text
Use model-driven-design to clarify this idea before we choose a model. Do not edit.
Use model-driven-design to draft a lifecycle proposal in the approved worktree;
leave unresolved policy choices visible and arrange independent expectations.
Prepare the model and generated views for review, but do not run the candidate.
```

For installation, use the repository's existing skill-pack instructions. No other
skill is required; target-repository instructions and authority still apply.
