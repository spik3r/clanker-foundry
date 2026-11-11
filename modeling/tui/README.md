# Generate a responsive Flexgo record browser

This bounded Foundry compiler turns a reviewed JSON presentation model into an
importable Go package and a runnable synthetic demo. It uses Flexgo v0.1.2 and
Bubble Tea v2. No private application source or data is included.

The model chooses string fields, the identity and summary fields, ordered detail
fields, labels and explicit wide/narrow breakpoints. The compiler supplies one
fixed read-only interaction contract. It does not generate an arbitrary app,
network client, database, editable form or general widget tree.

## Try the supplied model

Use a current patched Go release (CI uses Go 1.27.4; the compiler language
minimum is Go 1.27). From this directory:

```sh
output=$(mktemp -d)
go run ./cmd/tuic examples/browser.json "$output/catalog"
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
cell width, not wrapped. ANSI and control characters from data are sanitized.

## Change the requirement, then regenerate

Copy [the model](examples/browser.json), change `layout.wideWidth` to 100,
`layout.listWidth` to 32, and `details` to `["Status", "Name"]`. Compile it into a
new output directory. At 80 columns the changed browser now shows one pane, and
the details now show Status before Name. No generated Go file needs hand editing.
The acceptance test runs this exact change against independently stated expected
behavior, not an expectation generated from the model.

## Application integration

The generated package is `recordbrowser`, in the standalone module
`example.com/generated/recordbrowser`. The demo imports it directly. For a real
application, place the generated package in a reviewed module location or wire
that module through the application's existing dependency workflow. The emitted
module name is fixed and cannot be chosen by untrusted model input.

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
snapshots retain the previous identity for a later ready snapshot. The host must
store the returned model. There are no demo fixture-switch keys in the package.

## Files and provenance

Generated output contains formatted Go, a pinned dependency manifest/checksums,
normalized `model.json`, a synthetic `cmd/demo`, and `trace.json`. The trace binds
all other emitted files by SHA-256. It is an integrity manifest, not an approval
or proof of correctness. Field declarations normalize by ID; detail order remains
meaningful. Existing output directories, including symlinks, are refused. An I/O
failure after output creation leaves a clearly reported partial directory for
inspection; it does not overwrite an existing application.

The renderer templates adapt the generic public Flexgo responsive-example source
proposal against [Flexgo main 312eeee4](https://github.com/spik3r/flexgo/commit/312eeee4e9f70e526dea10e673d8c68c955fe83b).
That proposal was example-local, so these templates are a Foundry-owned bounded
backend, not an import of a new stable Flexgo component API. A future shared
component extraction needs its own compatibility and consumer checks. This does
not change Flexgo core or upgrade any downstream application.

## Verification

```sh
go test -short -count=1 ./...  # model, emission and CLI checks, no generated build
go test -race -count=1 -p=1 ./...  # also compile/run both generated variants
go vet ./...
```

Generated acceptance runs use bounded subprocess timeouts and `-mod=readonly`.
They check both baseline and changed models, typed records, help/focus/selection,
replacement, status recovery, clipping and viewport bounds. The source templates
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
