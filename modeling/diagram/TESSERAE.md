# Tesserae export seam

`diagramgen tesserae MODEL` validates a `dashboard/v1` model and writes a
`tesserae/v1` presentation model to stdout. It does not render HTML, fetch or
install Tesserae, write files, or alter the existing dashboard generator output.

This is an experimental consumer seam against Tesserae PR 1, not a released
package integration. The consumer owns themes, markup and final rendering limits.
The external renderer must still validate the emitted model and enforce its
1 MiB HTML output limit. Source size acceptance does not guarantee rendered size.

Supported mapping:

- Numeric stat values become metrics. Values must be unformatted decimal text
  accepted by the dashboard fixed-point number contract; `42%`, `unknown`, units,
  exponent notation and missing-state inference are refused, not guessed.
- Table selected columns and rows retain order. Number cells become canonical
  display strings, avoiding numeric round trips. At most eight selected columns
  are accepted, even though Foundry can describe twelve.
- Widget reading order follows layout, not the normalized widget array. Grid-row
  groupings are **not retained**: Tesserae owns responsive layout.
- Charts and other unsupported widgets cause whole-export refusal. Nothing is
  silently omitted. Empty tables retain `rows: []`.

Output is deterministic compact JSON with one newline, bounded to 512 KiB.
It carries plain text, not trusted HTML; consumers must escape it. Numeric metric
JSON remains canonical decimal, but Tesserae's JavaScript renderer subsequently
uses Number and its display formatting. This is not arbitrary-precision rendering.
Theme/density/mode use Tesserae defaults; there is no duplicated theme catalogue
in Foundry. React, Solid and Jekyll integrations are not implemented here.

Example (using an already reviewed local Tesserae checkout):

```sh
diagramgen tesserae numeric-table-dashboard.json > presentation.json
# Render presentation.json with Tesserae's renderPage, explicitly selecting
# and verifying the consumer revision in your own build. No auto-download occurs.
```

The existing dashboard bar/line examples are deliberately not portable through
this seam. The fixture tests define the accepted narrow contract.
