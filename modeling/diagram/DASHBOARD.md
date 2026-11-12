# Profile `dashboard/v1`

A dashboard of stat tiles, bar charts, line charts and tables over small
datasets that sit inside the model. `diagramgen` turns the model into one
self-contained HTML file. The model carries meaning and data only: no markup,
coordinates, colours or script. Selected by `"profileVersion": "dashboard/v1"`;
the input format is `diagram-json/v1`, as for [`process-diagram/v1`](PROFILE.md).

Go validation is authoritative. The
[schema](authoring/dashboard-v1.schema.json) is an authoring aid and is looser
in the places its own description lists.

## Input

```json
{
  "formatVersion": "diagram-json/v1",
  "profileVersion": "dashboard/v1",
  "id": "WeeklyOps",
  "title": "Weekly operations",
  "subtitle": "Week 45",
  "datasets": [
    {
      "id": "Orders",
      "columns": [
        {"id": "Day", "label": "Day", "type": "text"},
        {"id": "Placed", "label": "Placed", "type": "number"},
        {"id": "Shipped", "label": "Shipped", "type": "number"}
      ],
      "rows": [["Mon", 120, 96], ["Tue", 135, 130]]
    }
  ],
  "widgets": [
    {"id": "Total", "kind": "stat", "label": "Orders placed", "value": "872", "note": "+4% on last week"},
    {"id": "ByDay", "kind": "bar", "title": "Orders by day", "dataset": "Orders", "category": "Day", "series": ["Placed", "Shipped"]},
    {"id": "Trend", "kind": "line", "title": "Shipped per day", "dataset": "Orders", "category": "Day", "series": ["Shipped"]},
    {"id": "Detail", "kind": "table", "title": "Daily detail", "dataset": "Orders", "columns": ["Day", "Placed", "Shipped"]}
  ],
  "layout": [["Total"], ["ByDay", "Trend"], ["Detail"]]
}
```

Full examples: [ops-dashboard.json](examples/ops-dashboard.json) (the model
above with seven days of data), [sales-dashboard.json](examples/sales-dashboard.json)
(two datasets, negative values, four series, a 12-column table, a long title and
long category labels) and [minimal-dashboard.json](examples/minimal-dashboard.json)
(one stat widget, no datasets).

`subtitle` and `datasets` are optional; every other top-level member is
required. A dataset's `rows` is optional and means no rows when absent.

### Widgets

| Kind | Members besides `id` and `kind` | Shows |
| --- | --- | --- |
| `stat` | `label`, `value`, optional `note` | `label` small, `value` large, `note` muted. All three are plain text; nothing is computed. |
| `bar` | `title`, `dataset`, `category`, `series` | Grouped vertical bars. `category` is a text column; `series` lists 1 to 4 number columns. |
| `line` | `title`, `dataset`, `category`, `series` | One line per series. The category column gives the x order as listed in the rows. |
| `table` | `title`, `dataset`, `columns` | 1 to 12 columns of the dataset, in the listed order, all rows. |

A member that the kind does not use is refused, even when empty.

### Layout

`layout` is 1 to 12 rows of 1 to 4 widget ids. Every widget appears exactly
once. Widgets in a row share the width equally. Under 720 px wide the page
shows one column.

## Strict decoding

Refused, as for the process profile (the shared strict scanner): unknown
members (a case variant such as `Title` is unknown), duplicate JSON keys,
`null`, an object or array where a scalar belongs, trailing data after the
model, invalid UTF-8, nesting deeper than 16 and input larger than 512 KiB. A
cell that is `true` or `false` is refused. `subtitle` and `note`, when present,
must not be empty.

## Validation

- `formatVersion` and `profileVersion` are exactly the values above.
- Every `id` (model, dataset, column, widget) is an ASCII identifier
  `[A-Za-z][A-Za-z0-9_]{0,47}`. Ids are unique among datasets, among widgets and
  among the columns of one dataset. A dataset id and a widget id may be equal.
- `title`, `subtitle`, `label`, `value`, `note` and every widget `title` are 1
  to 120 UTF-8 bytes of plain text: not blank, no control characters, no
  U+2028, U+2029, U+FFFE or U+FFFF. Column labels follow the same rule.
- 0 to 8 datasets, 1 to 12 columns each, 0 to 200 rows each. Every row has
  exactly one cell per column. A `text` cell is a JSON string of 0 to 120 bytes
  with the same character rules (it may be empty or blank). A `number` cell is a
  JSON number, see below.
- 1 to 24 widgets.
- Every reference resolves with the right type: `dataset` names a dataset,
  `category` a text column of it, each `series` entry a number column, each
  table `columns` entry a column. `series` and `columns` hold no duplicates.
- A `bar` or `line` widget's dataset has 1 to 60 rows.
- Layout rules as above.

## Numbers

A number cell is read as decimal text, never through a float. It must be a
valid JSON number written without an exponent. Its canonical text is then:

- trailing zeros after the decimal point are dropped, and the point with them
  when nothing is left (`1.50` becomes `1.5`, `100.0` becomes `100`);
- `-0`, `-0.0` and similar become `0`;
- no leading zeros, no `+`.

The canonical text must have at most 6 fractional digits, at most 15
significant digits (digits after leading zeros, counting integer digits and the
remaining fractional digits) and a magnitude below 1e12. So `1.5000000` is
accepted (it is `1.5`), `0.0000001` is refused, and `123456789012.3456` is
refused (16 significant digits).

The canonical encoding, every emitted file and every chart value use the
canonical text, so `1.50` and `1.5` give the same digest and the same bytes.
Chart arithmetic uses the value as an integer count of millionths.

## Canonical encoding

Input to the model digest (`canonicalSHA256`): compact JSON, no trailing
newline, no HTML escaping, members in this fixed order:

- model: `formatVersion`, `profileVersion`, `id`, `title`, `subtitle` (only when
  present), `datasets`, `widgets`, `layout`. `datasets` is `[]` when there are
  none.
- dataset: `id`, `columns`, `rows`. Column: `id`, `label`, `type`.
- widget: `id`, `kind`, then those of `label`, `value`, `note`, `title`,
  `dataset`, `category`, `series`, `columns` that the kind uses.
- text cells are JSON strings; number cells are canonical decimal text.

Datasets are sorted by `id` and widgets by `id` (byte order). Column order, row
order, `series` order, table `columns` order and `layout` are significant and
are kept. JSON whitespace, member order in the source, dataset order, widget
order and number spelling do not change the digest. Any text, id, type, cell,
series, column or layout change does.

## Output files

| File | Content |
| --- | --- |
| `dashboard.html` | the dashboard |
| `model.json` | the canonical model, two-space indented, trailing newline |
| `foundry-manifest.json` | integrity record: generator `diagramgen` 1.1.0, toolchain constant `go1.27.4` |

No emitted file is `.go` or `.ts`, so the generated-file header rule has
nothing to apply to.

### dashboard.html

One inert HTML5 file. It has no `<script>`, no event-handler attribute, no
`<link>`, `src`, `href`, `<base>`, `<iframe>`, `<form>` or `<meta http-equiv>`,
no `url(...)`, no `@import` and no web font. Nothing in it names a URL. It holds
one inline `<style>` element. All model text is HTML-escaped. Output depends only on the canonical model
and the generator version.

Structure: `<header>` with an `<h1>` title and an optional subtitle paragraph;
then `<main>` with one `<section class="row cols-N">` per layout row, holding
one `<article data-id data-kind>` per widget in layout order.

Look: system font stack; neutral page background, white cards with a 1 px
border and 8 px radius, 16 px gaps; CSS grid with equal columns per row,
collapsing to one column under 720 px; light and dark themes through
`prefers-color-scheme`, using custom properties defined once on `:root`. Four
series colours, assigned by position (series 1 is always colour 1):
blue, orange, green and purple, with separate values for the dark theme.
Colour is never the only cue: a chart with more than one series has a legend
with text labels, and line series also differ in dash pattern (solid, dashed,
dotted, dash-dot).

`stat`: label, value and note as three paragraphs.

`table`: a real `<table>` with `<thead>` and `<tbody>` inside a horizontally
scrolling, keyboard-focusable container. Number columns are right-aligned with
tabular numerals. Numbers print in canonical text, without thousands
separators.

`bar` and `line`: an inline `<svg role="img" aria-label="TITLE, bar chart">`
with a fixed `viewBox="0 0 640 320"`, shown at the width of its card up to 720
px. Every data mark (bar or point) has a `<title>` reading
`SERIES LABEL, CATEGORY: VALUE`, so values and full category text are available
on hover and to assistive technology.

### Chart rules

Y axis. The axis is built from the range of every plotted value.

- Bars: the range always includes 0.
- Lines: if all values are at least 0 and the smallest is at most half the
  largest, the range starts at 0. If all values are at most 0 and the largest
  is at most half as far from 0 as the smallest, the range ends at 0. Otherwise the range is the data range, and the axis starts
  at the nice value at or below the smallest value, so it may not include 0. A
  flat series spans 0 to its value.
- Ticks are multiples of a step of 1, 2 or 5 times a power of ten. The step is
  the smallest one for which the range, widened outward to multiples of the
  step, has at most 5 intervals. If that gives fewer than 3 intervals (a range
  such as -5 to 5 with step 5), whole steps are added, on the side with less
  slack, until there are 3. So every chart has 3 to 5 intervals: 4 to 6
  gridlines, each with a tick label. The first tick is at or below the lowest
  value and the last at or above the highest.
- A zero baseline is drawn when 0 lies on the axis. Negative bars hang below it.

X axis. Categories are laid out left to right in row order, each with an equal
slot. Labels come from the category column. Every n-th label is shown, where n
is the smallest whole number that gives each shown label at least 64 px
of its own; each is cut to the number of characters that fit, assuming 8 px per
character, at most 24, and ends with `…` when cut. The count is of Unicode code
points. Text width is estimated, not measured, so a very wide font can still
overlap neighbours.

Bars. Within a slot the bars of the series sit side by side, centered, with
20% of the slot left empty. A bar is at most 60 px wide. Bar count is rows
times series, zero-height bars included.

Lines. One polyline per series through the category centres, with a round
marker for each point.

### Byte-stable coordinates

No floating-point value takes part in chart geometry. Values are `int64`
millionths. Positions are `int64` hundredths of a pixel, formed with integer
division. The value-to-y mapping multiplies and divides with `math/big` and
rounds half up. No `math.Pow`, no `math` function and no `strconv` float
formatting is used. Coordinates are printed from integers with exactly two
decimals. Tick values and tooltips print canonical decimal text. The same
model therefore gives the same bytes on every platform.

## What it cannot express

- live or loaded data: datasets are embedded in the model, at most 200 rows
- computed values or aggregates: a stat value is text someone wrote; no sums,
  averages, percentages or formulas
- interactivity: no filters, sorting, tabs, tooltips beyond the plain `<title>`
  text, or links
- custom colours, fonts, sizes or layout beyond rows of 1 to 4 equal columns
- chart kinds beyond bar and line (no pie, area, stacked or scatter)
- more than 4 series in a chart, a second y axis, a secondary scale, axis
  titles or axis ranges you choose
- dates or time scales: a date is a text category, in the row order you give
- number formatting such as thousands separators, units or percentages; put
  the unit in a label

## Limits of what tests show

Tests parse the output, check that nothing active or external is emitted,
check chart geometry against the data and compare generated examples byte for
byte. They do not show that a dashboard reads well. A person should open the
page, in light and dark themes and at phone width, before relying on it. On
narrow screens the chart text scales down with the chart.
