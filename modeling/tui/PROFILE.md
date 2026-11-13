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
  the wide list width. Below the narrow minimum, in width or in height, the
  browser asks for a resize. The two-pane layout needs BOTH the wide width and
  the wide height: at a size that reaches only one of them (for example the wide
  width but not the wide height) the browser shows one focused pane, not two.
  The two-pane layout shows both panes with a one-cell gutter. The sizes are
  those of the area the browser is given: the terminal for a standalone browser,
  the body area for a hosted one.

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

- `embed` is optional: `"standalone"` (default) or `"hosted"`. A standalone
  browser owns the terminal: it draws a title line, a status line and a key-hint
  footer, enables the alternate screen and quits on `q` and `ctrl+c`. A hosted
  browser is one page body inside a host application's page. It never returns a
  quit command (`q` and `ctrl+c` are ignored outside help; inside help only the
  documented close keys act, as before), sets no alternate screen and draws only
  the body, in exactly the width and height of the last `tea.WindowSizeMsg`: the
  list and detail panes with their labels, the loading, failed or empty message,
  the tiny-size message without a quit hint, and the help overlay without its
  quit line. All layout rules apply to that body size, so a body of height `H`
  is one label row and `H-1` content rows. The host draws the title and footer
  using `Title()` and `KeyHints()`, and must forward every key while
  `CapturesKeys()` is true. See the README, "Inside an application's page". Both
  embed modes emit `CapturesKeys`, `KeyHints`, `Title`, `Colors` and `WithColors`.
  `embed` is independent of `emit`: a hosted model can be package mode or module
  mode. In module mode the demo is a minimal host.
- `detailSpacing` is optional: `"compact"` (default) or `"spaced"`. Spaced puts
  one blank row between consecutive detail fields, not before the first and not
  after the last. The blank rows are rows of the detail pane: the counts used for
  scrolling, paging, Home/End and clamping include them, in clip and in wrap text
  mode, and in wrap mode they count toward the 4,096-row cap.

Colours are application presentation, not design meaning, so the model has no
colour members. A host sets them at run time with `WithColors` (README).

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
The [JSON Schema](authoring/record-browser-v1.schema.json) is an authoring aid
for drafting a model; it lists the rules it cannot express.

Compilation normalizes field declarations by ID but retains detail order. It
emits fixed relative file names, formatted Go source (package `recordbrowser`, or
the model's `emit.package`), pinned dependencies (module mode), normalized model and a `foundry-manifest.json` integrity
record (generator `tuic` 1.3.0, toolchain constant `go1.27.4`). It does
not execute generated code. Rendering and independent behavior tests remain
separate acceptance steps. Model edits require regeneration, not edits to the
emitted files.

Expectations for a model, such as "after these keys the browser shows X", are
written as data in [expectation cases](CASES.md) and checked with `tuicases`.

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
`identity`, `summary`, `details`, `listLabel`, `detailLabel`, `layout`, then `emit`, `detailText`, `embed` and `detailSpacing` when present. Each field
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

`embed` and `detailSpacing` were added in generator 1.3.0, again without a
profile version change, and follow the same rule: `"embed": "standalone"` and
`"detailSpacing": "compact"` are dropped from the canonical encoding, and only
`"hosted"` and `"spaced"` appear, as `"embed": "hosted"` and
`"detailSpacing": "spaced"`, after `detailText`, in that order. Empty strings are
refused here too. The canonical digests of `examples/browser.json` and
`examples/browser-package.json` did not change; a test pins both
(`6a8370de...ac53` and `43e31a41805dd76946df369f36d53c50df3521a937b330a54d9a3cac6a122414`).
Emitted Go files differ from 1.2.0 output in the generator version in their
header and in the added accessors and colour support, so a manifest changes
even when the canonical digest does not.

JSON whitespace, member order in the source and the order of `fields` do not
change the digest. The order of `details` does. Emitted bytes depend only on the
canonical model and the generator version, not on source bytes, file names or
paths.
