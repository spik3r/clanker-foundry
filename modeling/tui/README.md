# Generate a responsive Flexgo record browser

This bounded Foundry compiler turns a reviewed JSON presentation model into an
importable Go package and a runnable synthetic demo. It uses Flexgo v0.1.2 and
Bubble Tea v2. No private application source or data is included.

The model chooses string fields, the identity and summary fields, ordered detail
fields, labels and explicit wide/narrow breakpoints. The compiler supplies one
fixed read-only interaction contract. It does not generate an arbitrary app,
network client, database, editable form or general widget tree.

## Try the supplied model

Running the generator needs Go 1.27.0 or later (`go 1.27.0` in this module's
`go.mod`; CI uses Go 1.27.4). Building generated output needs only Go 1.25.5, see
[Go versions](#go-versions). Use a current patched Go release. From this directory:

```sh
output=$(mktemp -d)
go run ./cmd/tuic validate examples/browser.json
go run ./cmd/tuic generate examples/browser.json "$output/catalog"
go run ./cmd/tuic check examples/browser.json "$output/catalog"
cd "$output/catalog"
go run -mod=readonly ./cmd/demo
```

The demo contains 40 synthetic records. It cannot fetch data or run jobs. The first
build may download the exact dependencies and checksums emitted in `go.mod` and
`go.sum`. The compiler itself uses only the standard library and does not install
or execute generated code.

Use Tab or Shift+Tab to switch panes, Enter to show details and Esc to return to
the list. j/k or arrows move through the focused pane. PgUp/PgDown move a visible
page; Home/End go to the ends. `?` opens help. While help is open, only its scroll
and close keys work, including at tiny sizes. q or Ctrl+C quits outside help.

The sample shows both panes at 80x12 or larger, one focused pane at 24x6 or larger,
and a clipped resize message below that. Both dimensions matter. Selection and
focus survive resize; offsets clamp to the new viewport and do not reappear after
later shrinking. A focused selection stays visible. Text is clipped by terminal
cell width, not wrapped, unless the model sets `detailText` to `"wrap"` (see
below). ANSI and control characters from data are sanitized.

## Wrapped detail text

By default (`"detailText": "clip"`) each detail field is one row. A value is cut
at the pane width with `…`, and a newline in it is shown as a space. A 2000
character value shows only what fits in one row.

With `"detailText": "wrap"` the detail pane shows the whole value. Each value is
split at newlines, and each line is wrapped to the pane width by terminal cell
width. Breaks fall at spaces when the row has one; a word longer than the pane
is broken between characters, never inside a grapheme or a wide character. The
label stays in the first row (`Label: value`); there is no indent on later rows.
Sanitizing is unchanged: ANSI sequences are removed and other control characters
become spaces. Newlines are the only control character that starts a new row.

Scrolling, PgUp/PgDown, Home/End and offset clamping use the wrapped row count at
the current width. A resize re-wraps and clamps. At most 4,096 rows are shown for
one record, then a final row `… truncated`. The list pane is unchanged: the
summary stays one clipped row.

## Change the requirement, then regenerate

Copy [the model](examples/browser.json), change `layout.wideWidth` to 100,
`layout.listWidth` to 32, and `details` to `["Status", "Name"]`. Compile it into a
new output directory. At 80 columns the changed browser now shows one pane, and
the details now show Status before Name. No generated Go file needs hand editing.
The acceptance test runs this exact change against independently stated expected
behavior, not an expectation generated from the model.

## Application integration

To use a browser inside an application, set `"emit": {"mode": "package",
"package": "name"}` in the model, as in [the package example](examples/browser-package.json).
The generator then writes only the package's `.go` files (with `package name`),
`model.json` and `foundry-manifest.json`. It writes no `go.mod`, `go.sum` or
`cmd/demo`. Generate into a directory inside the application's module, such as
`internal/mcpbrowser`, and import it by that module's own path. Two browsers with
different package names can live in one module. The model cannot choose import
paths, module paths or file names.

The host module must provide the dependencies, because the package has no
`go.mod` of its own. Its `go.mod` must require compatible versions of
`charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `github.com/spik3r/flexgo`
and `github.com/charmbracelet/x/ansi`, plus their indirect requirements. The
versions the generator was tested with are in `templates/go.mod.tmpl` (bubbletea
v2.0.6, lipgloss v2.0.3, flexgo v0.1.2, x/ansi v0.11.7); newer compatible versions
may work but are not tested here. The host's `go` directive must be at least 1.25.5.
The acceptance tests build package-mode output in a host module with `go 1.26.0`.

[`foundrygen`](../foundrygen/README.md) can regenerate and check such a package
from its model, so the committed files do not drift. Without `emit`, or with
`"emit": {"mode": "module"}`, output is the standalone module
`example.com/generated/recordbrowser` with package `recordbrowser` and the demo.
That module name is fixed. Do not copy it into an application: use package mode.

The model emits a `Record` struct with the declared string fields. Application
code calls `New([]Record)` and uses the returned `Model` as a Bubble Tea model.
Application code owns loading and actions. It can replace a snapshot with
`Replace(Snapshot{Status: Ready, Records: rows})`, or show `Loading`/`Failed` with a
plain-text `Message`. Replace returns a new model and an error; invalid input
returns the original model unchanged. It copies record content into the browser.
Record IDs must be unique and nonempty. A snapshot allows at most 4,096 records,
4,096 bytes per field/message and 1 MiB of total source text.

Replacement preserves an existing selected ID after reordering; removal selects
the first new record. Empty ready data clears selection. Loading and failed
snapshots retain the previous identity, focus and scroll positions for a later
ready snapshot. A same-size reload preserves valid positions. During loading or
failure, supported resizes clamp list/detail offsets against the last ready
snapshot's row counts, not any records supplied with a status snapshot. The
browser retains those counts, not the old record content. Expanding and then
shrinking cannot restore a clamped detail offset; a ready list can still advance
its offset as needed to show the selected record. Tiny viewports freeze offsets.
A new ready snapshot refreshes the counts, and an empty snapshot or removed
selected ID clears the old positions. The host must store the returned model.
There are no demo fixture-switch keys in the package.

Read-only accessors, added in generator 1.2.0, let a host or a test observe the
state without reading rendered text: `Selected() string` is the selected record's
identity, or `""` while loading, failed or empty; `Focus() string` is `"list"` or
`"detail"` (loading and failed snapshots retain it); `Layout() string` is
`"wide"`, `"narrow"` or `"tiny"` for the current size (`"tiny"` before the first
size); `HelpOpen() bool`; and `Status() Status` is the current snapshot's status.

## Expectation cases

Behaviour checks for a model are data too. A [cases file](CASES.md) states, per
case, a start size, a record set, key, resize and replace steps, and what must
then be true. `tuicases` runs it against the browser generated from the model:

```sh
go run ./cmd/tuicases validate examples/browser.json examples/browser.cases.json
go run ./cmd/tuicases check examples/browser.json examples/browser.cases.json
```

`validate` only reads and checks the files. `check` also builds and runs the
generated browser, so it needs `go` and the dependencies in the module cache or
network. The cases are bound to the model's canonical digest and are refused
when the model changes. The `acceptance/*.txt` suites stay as they are.

## Files and provenance

Module-mode output contains formatted Go, a pinned dependency manifest/checksums,
normalized `model.json`, a synthetic `cmd/demo`, and `foundry-manifest.json`.
Package-mode output has the Go files, `model.json` and `foundry-manifest.json`. The
manifest lists every other emitted file with its SHA-256 and size. It is an
integrity record, not an approval or proof of correctness. Every emitted `.go`
file starts with `// Code generated by tuic 1.2.0 from model <canonicalSHA256>. DO NOT EDIT.`

`tuic` follows the [Foundry generator contract](../CONTRACT.md):

- `validate MODEL` reads the model and prints one identity line.
- `generate MODEL OUT` creates a new `OUT` and prints the same line.
- `check MODEL OUT` compares `OUT` with what `generate` would write, writes
  nothing, and exits 3 with `drift:` lines on stderr when they differ.

Exit codes are 0 success, 1 I/O failure, 2 refused (usage, bad model, existing
`OUT`, missing parent of `OUT`) and 3 drift. An existing `OUT`, including a
symlink, is refused. An I/O failure after `OUT` is created leaves a reported
partial directory for inspection. The old form `tuic MODEL OUT` is removed.
Field declarations normalize by ID; detail order remains meaningful.

The renderer templates adapt the generic public Flexgo responsive-example source
proposal against [Flexgo main 312eeee4](https://github.com/spik3r/flexgo/commit/312eeee4e9f70e526dea10e673d8c68c955fe83b).
That proposal was example-local, so these templates are a Foundry-owned bounded
backend, not an import of a new stable Flexgo component API. A future shared
component extraction needs its own compatibility and consumer checks. This does
not change Flexgo core or upgrade any downstream application.

## Go versions

- Running the generator (`tuic`): Go 1.27.0 or later, as declared by
  `modeling/tui/go.mod`. CI and the manifest's toolchain constant use Go 1.27.4.
  The compiler itself uses only the standard library. As a check, the compiler's
  unit tests also passed with the directive lowered to 1.25.5 on Go 1.25.5, but
  that is not a supported configuration.
- Building generated output: Go 1.25.5. The emitted code uses nothing newer than
  Go 1.21 (`min`, `max`, `slices`), but the dependency graph does: `flexgo` v0.1.2
  declares `go 1.25.5`, and bubbletea, lipgloss, ultraviolet and others declare
  1.25.0. A lower `go` directive in the emitted `go.mod` fails with
  `-mod=readonly` ("updates to go.mod needed"). The emitted module builds with
  Go 1.25.5, 1.26.0 and 1.27.4.
- Package mode: the host module needs `go 1.25.5` or later; a host with `go 1.26.0`
  builds it. The host's own toolchain must be at least its directive.

## Verification

```sh
go test -short -count=1 ./...  # model, emission and CLI checks, no generated build
go test -race -count=1 -p=1 ./...  # also compile/run both generated variants
go vet ./...
```

The `tuicases` tests that build generated code also skip under `-short`.

Generated acceptance runs use bounded subprocess timeouts and `-mod=readonly`.
They check the baseline, changed and wrapped-text models, typed records,
help/focus/selection, replacement, status recovery, clipping, wrapping, scrolling
and resize clamping, and viewport bounds. A separate test generates two
package-mode browsers with different package names and fields into a temporary
host module (`go 1.26.0`), builds and runs both, and runs the behavior checks
against one. The source templates
are not themselves Go files; each emitted Go file is parsed/formatted during
compilation and then compiled in the generated acceptance suite.

CI also runs pinned `govulncheck` v1.9.0 on Go 1.27.4 for both this compiler and
a freshly generated module, using Linux/amd64, Darwin/arm64 and Windows/amd64
source graphs. Normal scanner exit status gates reachable vulnerable symbols;
module-only advisories still need review. Cross-target source analysis does not
claim that a native Windows or Linux application was executed. Models cannot
change imports, so one generated baseline covers the dependency graph.

The generated `golang.org/x/sys` pin is v0.44.0, addressing
[GO-2026-5024](https://pkg.go.dev/vuln/GO-2026-5024). The prior v0.43.0 scan reported
that Windows-package advisory but no reachable vulnerable symbol in this browser.
The patch removes that known vulnerable version rather than relying on current
unreachability. Keep the build toolchain patched as well; a language-version
minimum in `go.mod` is not a security-update policy.

These checks do not replace manual review of colors and interactive resizing in a
real terminal, a downstream application trial, or a second-use productivity
measurement. See the [profile contract](PROFILE.md) for the exact input boundary.
