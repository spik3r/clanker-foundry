# Record browser profile v1

This profile describes one read-only list/detail browser, not a general UI tree.
Its input format is `tui-json/v1`, with profile `record-browser/v1`.

## Semantic input

- `id` names the reviewed model. It is an exported ASCII identifier of at most
  48 characters.
- `fields` declares 1 to 16 named string fields and their display labels. Field
  IDs use the same identifier rule and must be unique. No expression or Go type
  syntax is accepted.
- `identity` names the stable, nonempty record identity field. `summary` names
  the list label field. Both reference declared fields.
- `details` lists 1 to 16 unique declared field IDs in display order.
- `title`, `listLabel` and `detailLabel` are plain text. All labels are 1 to 120
  UTF-8 bytes and cannot contain control characters.
- `layout` defines explicit narrow and wide minimum widths and heights, plus
  the wide list width. Below the narrow minimum the browser asks for a resize.
  Below either wide minimum it shows only the focused pane. The wide variant
  shows both panes with a one-cell gutter.

The narrow minimum cannot be smaller than 24 by 6 or larger than 240 by 80.
Wide minima must be at least the narrow minima and at most 400 by 120. Each wide
pane gets at least eight columns. The exact defaults in the example are design
choices, not a claim that every terminal or application should use them.

## Fixed behavior

The profile fixes keyboard behavior, selection by stable ID, viewport clamping,
list/detail focus, data replacement and captive help. Models cannot bind shell
commands, define arbitrary actions, choose import paths, create widgets, access
files, invoke a service or change runtime authority. There are no editable form
fields or background jobs. Application code supplies typed records and adapts its
own loading and error states.

Unrecognized fields, including case variants, are rejected. Duplicate JSON keys,
null values, unsupported versions, invalid UTF-8, trailing data, overly deep JSON
and input larger than 32 KiB are rejected. Go validation is authoritative.

Compilation normalizes field declarations by ID but retains detail order. It
emits a fixed package name and fixed relative file names, formatted Go source,
pinned dependencies, normalized model and a SHA-256 artifact manifest. It does
not execute generated code. Rendering and independent behavior tests remain
separate acceptance steps. Model edits require regeneration, not edits to the
emitted files.
