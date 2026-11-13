# Gap report: threshold condition on the approval route

## Requirement

> If approved and the amount is over the finance threshold, finance must also check it; otherwise it goes straight to payment.

Source: brief 2 (expense approval process), the sentence after the manager review.

## Profile considered

- Profile: `process-diagram/v1`
- Registry entry and `cannotExpress` line that applies: "flow conditions or execution semantics"; the profile document also excludes "data objects".
- Profile document or schema section read: `modeling/diagram/PROFILE.md` (Input, Validation, BPMN output); `modeling/diagram/authoring/process-diagram-v1.schema.json`.

## Missing construct

A sequence flow has only `id`, `from`, `to` and an optional display `label`. The label is plain text with no meaning to the generator, the validator or the BPMN output. The model has no member that states a condition on a flow, no data object that carries the claim amount, and no threshold value. The "Over finance threshold?" gateway can therefore be drawn with "yes" and "no" branches, but the rule that selects a branch cannot be stated.

## What was not done as a result

- The model draws the threshold gateway and its two branches. It does not state the threshold rule, so nothing in the model or its output tests which branch a claim takes.
- The threshold amount is not in the brief and is not in the model.
- The "yes"/"no" flow labels are display text. They are not a condition and must not be read as one.

## Options

- Extend the profile: add a condition member on flows (or a data object with a typed comparison) to the input, schema, validator and generator. This would change what the profile means and needs a new profile or version decision.
- Hand-written seam owned by a person: keep the threshold rule in the application code that picks the route, and have the diagram show the route only. The owner would need to keep the diagram and the code in step.
- Leave the feature out: the diagram shows the approval route without the amount rule, and the process document does not state when finance is involved.

Decision owner: unknown

Decision: (blank until the owner decides)
