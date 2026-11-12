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

- `detailText` is optional: `"clip"` (default) or `"wrap"`. Clip shows each
  detail field as one row, cut at the pane width, with newlines shown as spaces.
  Wrap splits values at newlines and wraps each line to the detail pane width by
  terminal cell width, at spaces when possible and between characters otherwise,
  never inside a grapheme. Scrolling uses the wrapped row count for the current
  width. At most 4,096 rows are shown per record, then `… truncated`. The list
  pane is not affected.
- `emit` is optional: `{"mode": "module"}` (default) or
  `{"mode": "package", "package": "name"}`. Module mode writes a standalone
  module with the fixed path `example.com/generated/recordbrowser`, package
  `recordbrowser` and a demo. Package mode writes only the Go files with
  `package name`, `model.json` and `foundry-manifest.json`: no `go.mod`, `go.sum`
  or `cmd/demo`, so the output can sit inside a host module that supplies the
  dependencies. `package` is required in package mode and refused in module mode.
  It matches `[a-z][a-z0-9]{0,31}` and must not be a Go keyword, a predeclared
  identifier (such as `string`, `len` or `nil`) or `main`. Both members are plain
  choices: the model still cannot name an import path, module path or file.

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
emits fixed relative file names, formatted Go source (package `recordbrowser`, or
the model's `emit.package`), pinned dependencies (module mode), normalized model and a `foundry-manifest.json` integrity
record (generator `tuic` 1.1.0, toolchain constant `go1.27.4`). It does
not execute generated code. Rendering and independent behavior tests remain
separate acceptance steps. Model edits require regeneration, not edits to the
emitted files.

## Go versions

Running the generator needs Go 1.27.0 or later. Building generated output needs
Go 1.25.5, the highest `go` directive among the pinned dependencies (`flexgo`
v0.1.2); the emitted code itself uses nothing newer than Go 1.21. A package-mode
host needs `go 1.25.5` or later in its `go.mod`; `go 1.26.0` is tested. See the
README for details.

## Canonical encoding

`canonicalSHA256` is the SHA-256 of the normalized model as the compiler writes
it to `model.json`, without the final newline. The encoding is Go's
`encoding/json` indented with two spaces (default escaping of `<`, `>` and `&`),
members in the order `formatVersion`, `profileVersion`, `id`, `title`, `fields`,
`identity`, `summary`, `details`, `listLabel`, `detailLabel`, `layout`, then `emit` and `detailText` when present. Each field
is `id` then `label`; `layout` is `narrowWidth`, `narrowHeight`, `wideWidth`,
`wideHeight`, `listWidth`; `emit` is `mode` then `package`. `fields` is sorted by
`id`.

The two optional members were added in generator 1.1.0 without a profile
version change. A default is the same as absence: `"detailText": "clip"` and
`"emit": {"mode": "module"}` are dropped from the canonical encoding, so a model
without the members, or with them set to the defaults, has the same canonical
bytes and `canonicalSHA256` as before. The digest of `examples/browser.json` is
`6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53`, and a test pins
it. Only non-default values (`"wrap"`, package mode) appear in `model.json` and
change the digest. Empty strings are not defaults: `"detailText": ""` and
`"emit": {}` are refused.

JSON whitespace, member order in the source and the order of `fields` do not
change the digest. The order of `details` does. Emitted bytes depend only on the
canonical model and the generator version, not on source bytes, file names or
paths.
