# Prepare the fixed Go builder exercise

This package connects the unchanged [Go worked packet](../../../examples/model-builder-go/packet.md)
to the [pinned input preflight](../builderpreflight/README.md). It creates a new
private workspace containing exact source copies, the five frozen Go module
files, three byte reports and a fully path-bound handoff draft. It stops before
worker dispatch, candidate implementation, compilation or execution.

This is a FirstDesignDraft example preparer, not a general workspace manager or
an agent runtime. It uses only the standard library. The existing `foundry` routes,
preflight, source model, public packet, acceptance files and historical evidence
are unchanged. Follow the [coordinator walkthrough](../../../examples/model-builder-go/preparation.md)
for the complete local preparation flow.

## Fixed inputs, not refreshed pins

[inputs.json](inputs.json) lists 15 files: the eight original packet inputs, the
public packet, repository/global instructions, the implementation workflow,
handoff and prompt templates, and Project Memory guidance. The original eight
hashes come from the existing packet. All 15 were independently compared with
main `882b5002bd4dc70b4be0060e5e5668d1faa8bbdd` and frozen before implementation.
The exact manifest pin is
`6a87f4db69a9c43376b405aca0d1252a662a3a4cdc7496329f2b41fac2896032`.

The trusted binary embeds these bytes and their separate fixed digest. The
operator does not supply substitute hashes. A later legitimate source edit can
block this old example; use the reviewed baseline or obtain separately reviewed
support for the new input. Never refresh a pin merely to obtain a match. This
pinning is byte correlation, not source authentication or author verification.

The source root must be explicitly authorized and contain those exact files.
Preparation validates locations, runs the existing full source preflight, then
collects only the listed bytes through a rooted filesystem handle. Every second
read checks the expected digest, bounded size, regular-file metadata and named
file identity. Source directory enumeration, globbing, fallback search, Git,
subprocesses, tool probing and candidate execution are absent.

## API and CLI

`Prepare(sourceRoot, workParent, goBinary string) (Receipt, error)` requires three
absolute UTF-8 paths without control characters. The source and work-parent
locations must exist as real directory leaves; final symlinks, including trailing
slash spellings, are refused. Ancestor aliases retain the host's actual traversal
through `..`. The work parent must resolve outside the source root. This boundary
compares the directory identities of the parent and its ancestors with the source,
so case aliases cannot hide overlap. It reads ancestor metadata without enumerating
directories and refuses a metadata failure. It does not fold path case or reject
distinct directories merely because their names differ only in case. The Go path
is recorded literally; preparation does not open it or verify its version.

The standalone CLI takes exactly:

```text
modelbuilderdemo prepare --root ABSOLUTE-SOURCE --work-parent ABSOLUTE-PARENT --go ABSOLUTE-GO-BINARY
```

Flag pairs may appear in any order, once each. There are no default locations,
stdin input, discovery mode, execution mode or automatic retry. The caller's
shell must separately have permission to write a redirected receipt.

After all source bytes are collected, preparation creates an exclusive
`foundry-builder-*` directory under the resolved work parent. New directories
use 0700; new regular files use 0600, subject to a stricter umask. Nothing is
reused, overwritten, deleted or cleaned up automatically. The layout is:

```text
foundry-builder-<unique>/
  inputs/                 # 15 original paths and exact file bytes
  exercise/               # README.md, go.mod, contract.go, acceptance_test.go, rubric.md
  cache/ tmp/ modcache/    # empty proposed build locations
  source-manifest.json
  source-report.json
  exercise-manifest.json
  inputs-report.json
  exercise-report.json
  handoff.md
```

No `candidate.go` or `candidate.test` is created. The exercise manifest retains
the five frozen hashes and changes only their paths to module-local filenames.
The existing preflight checks both destination trees. Both must match before a
handoff is written and `prepared-draft` is returned. The saved report files retain
the original native JSON bytes. The receipt embeds their complete JSON objects.

Receipt fields include `formatVersion: model-builder-demo-preparation/v1`,
`outcome`, nullable `workspace` and `handoffPath`, `sourceManifestSHA256`, and
nullable `sourceReport`, `inputsReport`, `exerciseReport`. Every meaningful receipt
retains `executionStatus: not-run`, `reviewStatus: not-established`, and
`authority: none`. Null reports mean unrun, not passed.

- Exit 0: delivered a `prepared-draft` receipt. This is not permission or readiness.
- Exit 1: delivered a `blocked` receipt, with no new workspace. An initial source
  failure retains its full preflight report. A second-read failure may retain an
  earlier matched report; the top-level block and fixed-file diagnostic remain
  decisive. No synthetic preflight observation is substituted.
- Exit 2: invalid arguments or locations, no receipt and no preparation output.
- Exit 3: failed or short stdout delivery. A created workspace is retained and its
  quoted path is written to stderr. Do not retry blindly and create duplicate work.
- Exit 4: incomplete preparation or internal failure. A meaningful receipt retains
  actual reports and the workspace path when one was created; `handoffPath` stays
  null until its write completes. The directory may contain partial files.

The API uses `ErrInvalid` and `ErrBlocked` for the first two failure classes.
Other errors mean incomplete preparation. A nil error means `prepared-draft` only.
There is no promise of durable storage across a host crash or successful delivery
when the output channel itself fails.

## What the draft does and does not supply

The first section gives an explicit proposed initial read boundary. All source
and exercise paths and hashes are bound. The only proposed candidate edit is
`exercise/candidate.go`; the proposed build product is `exercise/candidate.test`.
The exact offline compile-only command is a JSON program, argument array,
working directory and environment. User path strings are data, not shell code.
The environment names native preparation-host GOOS/GOARCH, disables CGO and remote
dependency/toolchain retrieval, and binds private caches and temporary directories.
The proposed environment replaces the inherited environment, rather than overlaying
it; unlisted Go settings and external cache/tool hooks must not be inherited.
No command runs. Do not evaluate JSON values as shell expressions.

The required Go version is 1.27.3, but tool identity/version remain unchecked.
No gofmt executable or formatting authority is inferred from a Go pathname.
`go test -c` is build-only; `go test -run` still runs initializers. The coordinator
must separately verify exact candidate source and authorize any later compilation
or behavior execution. A test harness and rooted reads are not an OS sandbox.

The copied packet remains historical, including its pilot exception, reviewed
status, earlier paths and suggested Git reads. None becomes a current grant.
The draft requires a current live claim, current governing instructions, trusted
scope, semantic packet review and independent expectation provenance checks.
It supplies no permission to follow links or search for missing inputs. `Q_use`,
visual approval and real-use acceptance remain open. A legitimate fictional
experiment must be explicitly bounded rather than presented as product approval.

## Observation limits and checks

Use trusted, exclusive input and output parents. Ordinary named-file replacement
and metadata changes are checked, but this is not an atomic multi-file snapshot,
a hostile concurrent-writer defense, mount or hardlink authentication, or a
complete system-call audit. Source reads may update access times. Check receipts
and preserved files again when the actual authorized builder receives them.

The separate author froze `prepare_test.go` and `cmd/modelbuilderdemo/main_test.go`
before production code. They check copied byte identity, the exact file frontier,
reports, permissions, stale/missing/refused inputs, path traversal, new workspaces,
JSON command binding, CLI errors and output failures. Run from `modeling/lifecycle` on the native host:

```bash
go test -race -count=1 ./builderdemo ./cmd/modelbuilderdemo
go vet ./builderdemo ./cmd/modelbuilderdemo
```

These tests do not compile or execute a builder candidate. Concurrent replacement
and post-creation I/O failures lack deterministic fault-injection coverage.
Permission probes skip when root would bypass the tested mode bits. Source review
of those paths is distinct from an executed failure probe. No cross-platform
qualification is implied by the standard-library implementation.
