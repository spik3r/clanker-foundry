# Reuse a report template with fresh data

A recurring report should not need new HTML, CSS or layout instructions every
time. `diagramgen bind TEMPLATE DATA` takes a reviewed `dashboard/v1` model as
a template and a separate `dashboard-data/v1` payload. It outputs a complete
canonical dashboard model to stdout. The existing generator then renders it.

The human or AI supplies changing rows and stat text. Layout, column types,
labels and chart choices stay in the reviewed template. The renderer owns the
shared stat, chart and table components and light/dark responsive theme. No new
framework, SolidJS dependency or styling language is introduced.

## Try two reports

From `modeling/diagram`, with Go 1.27.4, choose a new scratch directory:

```sh
WORK="$(mktemp -d)"
go build -o "$WORK/diagramgen" ./cmd/diagramgen
"$WORK/diagramgen" bind examples/ops-dashboard.json examples/ops-report-data.json > "$WORK/report-one.json"
"$WORK/diagramgen" bind examples/ops-dashboard.json examples/ops-report-data-two.json > "$WORK/report-two.json"
"$WORK/diagramgen" generate "$WORK/report-one.json" "$WORK/one"
"$WORK/diagramgen" generate "$WORK/report-two.json" "$WORK/two"
"$WORK/diagramgen" check "$WORK/report-one.json" "$WORK/one"
"$WORK/diagramgen" check "$WORK/report-two.json" "$WORK/two"
```

Open `one/dashboard.html` and `two/dashboard.html`. Both fictional reports use
the same components, columns, chart definitions and theme; only rows and stat
values/notes differ. Check readability at phone and desktop widths and in light
and dark themes. Automated structural checks do not establish visual usability.

## Payload

```json
{
  "dataVersion": "dashboard-data/v1",
  "templateSHA256": "<SHA-256 of exact template file bytes>",
  "datasets": [{"id": "Orders", "rows": [["Mon", 120, 96], ["Tue", 135, 130]]}],
  "stats": [{"id": "Total", "value": "255", "note": "Two days"}]
}
```

Use the actual stat ids in your template; the examples are complete. The pin
is 64 lowercase hexadecimal digits. It binds data to reviewed bytes, including
whitespace, not to a filename or an unverifiable statement of approval. Compute
it when reviewing the template, for example with `shasum -a 256 TEMPLATE`.
A changed template requires a deliberately updated payload pin; the tool never
accepts a stale pin or rewrites it for you.

Both `datasets` and `stats` arrays are required, including `[]` when the template
has none. Every template dataset and stat must be supplied exactly once. Missing,
extra, duplicate and unknown ids are refused. Rows replace, not append. Every
stat value replaces the template value; omitting `note` removes any template
note. This prevents last week's data silently leaking into a partial update.
An explicit empty note is refused under the existing dashboard rules.

The fixed template is itself a valid dashboard model, with fictional sample
rows and stat values. Binding does not alter its title, subtitle, labels, column
types, chart choices or layout. A changing reporting-period caption can be a
stat value or note; changing a fixed title/subtitle means reviewing a new template.
There are no arbitrary paths, field overrides, expressions, formulas or computed
aggregates. Stat text is supplied, not recalculated from rows, so the author must
check its consistency with the data.

Input uses the dashboard's existing strict JSON scanner: no duplicate or unknown
keys, nulls, trailing data, invalid UTF-8 or deep structures. Each input and the
materialized model are bounded by 512 KiB, including output indentation and its
final newline. A compact payload within the input limit may still be refused if
its formatted materialized model exceeds that limit. The complete materialized model must
pass the existing `dashboard/v1` validator, including column types, decimal
precision, text limits and chart row limits. Decimal tokens never pass through
floating point. An empty table is allowed; an empty chart is refused.

The command reads regular leaf files only, rejecting symlinks, directories and
special files. Ancestor aliases and a trusted non-concurrently-mutated directory
are still assumed; this is not a filesystem sandbox. It never writes files.
Your shell redirection writes stdout and may truncate its destination before a
refusal, so write to a fresh scratch file, never over your template or input.
Successful stdout is one complete model followed by a newline. Refusal is exit
2 with no model on stdout. Filesystem read errors are also refusals, as with
existing generator input. No HTML is emitted until `generate` is called.

## Shared UI boundary

`dashboardbind.Bind` is a pure composition layer, not a second renderer. It calls
`dashboard.Decode` on the template and final model. All report instances then use
`diagram.Build` and the existing `dashboard` component/theme implementation.
No raw model or payload text becomes markup. This keeps escaping, chart geometry,
accessibility structure and responsive CSS in one place, with existing renderer
regressions unchanged. A future extracted library should preserve this boundary
and be driven by a second maintained consumer, rather than duplicating components
now. There is no claim of Flexgo or SolidJS output from this static HTML path.

The payload pin is checked at binding time. The emitted model carries the bound
data, not the template relationship. Keep the reviewed template and payload as
source alongside the model. Existing `foundrygen` can generate/check that model,
but does not automatically rerun binding or detect a stale template/payload pair.
A future project binding contract must address that dependency explicitly.

AI still uses tokens to produce data or narrative. This removes the need to have
it emit repeated HTML/layout, but actual token savings need measurement with
real colleague report formats. No savings percentage or human acceptance is
established by these fictional examples.
