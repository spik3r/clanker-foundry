# Brief 5 notes: searchable service list

- Profile: `record-browser/v1`. Chosen because the core of the brief is a read-only list of records with a detail pane and fixed keys. The filter and the reload are not part of the profile.
- Covered by `brief5.model.json`: the list of services with name, owner team, on-call contact and status; moving through the list; the detail view; one pane when narrow; the tiny resize message; application-supplied snapshots.
- Gaps (two reports in `brief5.gap.md`): the `/` filter (no search or filter state in the profile) and the `r` reload (no custom key action, no service loading, no background job).
- Assumed: the service name is unique and is the identity. The brief gives no service ID. If names can repeat, the model is refused.
- Assumed: the list shows the name as the label, and the details show all four fields including the name. The brief does not say which field is the list label.
- Assumed: default module output (no `emit` member), since the brief does not say the screen is a package. Layout values are the profile example values, not choices from the brief.
- Cases cover only the browsing and snapshot behaviour. They do not cover filtering or reloading, which the profile cannot express.
- `model.canonicalSHA256` is the literal `UNKNOWN`, which fails the schema pattern and must be replaced after review. Cases were not run.
- The model was not validated: no generator was run.
