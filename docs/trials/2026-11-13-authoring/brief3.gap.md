# Gap report: passed and failed bars stacked in one bar

## Requirement

> builds per day Monday to Friday as bars, split into passed and failed

Source: brief 3 (build health dashboard), the chart list.

## Profile considered

- Profile: `dashboard/v1`
- Registry entry and `cannotExpress` line that applies: "chart kinds other than bar and line". `modeling/diagram/DASHBOARD.md` ("What it cannot express") names "stacked" as a chart kind the profile does not provide.
- Profile document or schema section read: `modeling/diagram/DASHBOARD.md` (Widgets, "Bar" row; Chart rules, "Bars"; What it cannot express); `modeling/diagram/authoring/dashboard-v1.schema.json`.

## Missing construct

The `bar` widget draws grouped bars only: within each category the series sit side by side. There is no member that stacks the series into one bar per category, so a bar cannot be drawn as a passed segment and a failed segment on top of each other. The profile's numbers and axis rules assume grouped bars and do not describe stacked totals.

## What was not done as a result

- The model draws "Passed" and "Failed" as two grouped bars per day, not as one split bar per day. Read the brief as stacked and the chart is wrong.
- Nothing in the model shows a daily total bar height.
- The brief's wording may mean side-by-side bars. If so, this report does not apply and the model covers the chart. The brief does not settle it.

## Options

- Extend the profile: add a stacked bar kind to the widget table, the schema, the chart rules and the generator, which changes the byte output for the profile.
- Hand-written seam owned by a person: a hand-written chart outside the generated page. This would not be generated and would need its own review.
- Leave the feature out: keep the grouped bars, which show passed and failed per day without the split. Or drop the passed/failed bar and keep the table, which carries the same figures.

Decision owner: unknown

Decision: (blank until the owner decides)
