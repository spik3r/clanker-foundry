# Gap report

Use one report per requirement that no registered profile can express. Fill in
only what is known. Leave a field marked `unknown` rather than guessing. This
report is the outcome for that requirement; it is not followed by code or by a
prose workaround.

## Requirement

Quote it exactly from the brief, and say where it came from.

> (quoted requirement)

Source: (brief section, message or document)

## Profile considered

- Profile: (for example `record-browser/v1`)
- Registry entry and `cannotExpress` line that applies, if any:
- Profile document or schema section read:

## Missing construct

State precisely what the profile cannot say, in terms of the model: the member,
type, rule or behaviour that does not exist. Do not describe the whole feature.

(missing construct)

## What was not done as a result

List what the proposal leaves out or leaves unchecked because of this gap. For
example: the requirement is absent from the model, the generated code does not do
it, nothing was written to cover it.

- (item)

## Options

Present these for the decision owner. Do not pick one.

- Extend the profile: (what would change in the profile, schema and generator)
- Hand-written seam owned by a person: (where the seam would sit and who would own
  and review it)
- Leave the feature out: (what the product then lacks)

## Decision owner

Name (or `unknown`):

Decision: (blank until the owner decides)
