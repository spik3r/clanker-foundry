# Requirements trace

This package links the outcomes a person wants from the lifecycle example to
the parts of the accepted [model](../examples/lifecycle.json) and to the
independently written acceptance cases in the [oracle catalog](../examples/oracle.json).
It answers three questions for a reviewer:

- Which model elements and acceptance cases does each requirement point at?
- Which requirements have no model link, no acceptance case, or only pending cases?
- Which model elements and acceptance cases no requirement points at?

A link is a declaration made by a person. The tool checks that each link points
at something real and of the stated kind. It does not check that the model
satisfies the requirement, that a case tests it, or that the requirement is the
right one. A fully linked trace is a review aid, not evidence of correctness.
All output says `example-only` and `authority=none`.

## Run

From `modeling/lifecycle`, with Go 1.27.3 and the same offline settings as
`modelc`:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltrace validate
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltrace validate -complete
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltrace generate
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltrace check
go test ./traceability ./traceabilityacceptance ./cmd/modeltrace
```

The command takes no paths. It reads `examples/lifecycle.json`,
`examples/oracle.json` and [`traceability/requirements.json`](requirements.json),
and owns only [`trace.json`](view/trace.json), [`trace.md`](view/trace.md) and
[`trace.svg`](view/trace.svg) under `traceability/view/`.

| Exit | Meaning |
|---|---|
| 0 | Valid. `validate` lists gaps on stdout; `check` found the views current. |
| 1 | `check` found a missing or stale view. Nothing is written. |
| 2 | Invalid requirements, unreadable model or catalog, symlinked input or output, a view that would exceed the 4 MiB output limit, or bad usage. Issues go to stderr; stdout is empty. |
| 3 | `validate -complete` only: valid, but some requirement has a gap, a model element is untraced, or a case is unlinked. |

Every subcommand renders all three views first and refuses with exit 2 if any
would exceed 4 MiB, the same limit `check` uses to read them back. So
`generate` never writes views that `check` then rejects, and `validate` does not
accept inputs that `generate` cannot write.

`generate` refuses a symlinked or non-directory `view/` and non-regular output
files. It creates each temporary file exclusively (`O_EXCL`) under a random
hidden name in `view/`, so a planted file or symlink is never opened or
followed, then renames the temporaries over the outputs. On error it removes
only the temporaries it created. The three renames are not one atomic step. It
is a single-writer tool for a trusted checkout, like `modelc`; it makes no claim
against a concurrent hostile writer.

`TestCommittedViewsMatchGeneration` fails when the committed views are stale, so
the existing `model-lifecycle-go` CI job (`go test ./...`) covers drift without a
workflow change.

## Input format: `lifecycle-requirements-json/v1`

```json
{
  "formatVersion": "lifecycle-requirements-json/v1",
  "modelId": "TaskLifecycleExample",
  "acceptanceCatalog": "lifecycle-semantic-oracle-v2",
  "requirements": [
    {
      "id": "UR-002",
      "title": "Done is final",
      "statement": "A task in Done accepts no further lifecycle commands.",
      "modelLinks": [{ "kind": "state", "id": "Done" }],
      "acceptanceLinks": [{ "caseId": "refusals-15" }]
    }
  ]
}
```

- Decoding uses the existing `strictjson` package: duplicate, unknown or
  case-variant keys, missing fields, `null`, trailing data, invalid UTF-8 or
  surrogate escapes, more than 1 MiB or nesting deeper than 8 are rejected.
- `id`: 1-8 uppercase letters, a hyphen, 1-6 digits (`UR-001`). IDs are stable
  names; the tool never renumbers them.
- `title` (up to 120 bytes) and `statement` (up to 1024 bytes): nonblank UTF-8,
  no control characters other than tab.
- `modelLinks[].kind` is one of `state`, `transition`, `guard`,
  `modelRequirement`; `id` is an existing ID of that kind in the model. The kind
  is explicit so that a renamed element reusing an ID of another kind is caught.
- `acceptanceLinks[].caseId` is a `caseId` from the `sequences`, `refusals` or
  `special` sections of the oracle catalog. Other oracle sections have no case
  IDs and are not linkable.
- Both link lists are required and may be empty `[]`. At most 256 requirements
  and 64 links of each list per requirement.
- Order of requirements and links carries no meaning. The report sorts
  everything.

These requirements sit above the model's own `R_*` requirements. The model's
`R_*` entries are guard-level statements owned by the model and included in its
canonical digest. This file is a separate review artifact: it adds no state,
guard, effect or rule, and the generated reducer does not read it.

## Invalid input versus incomplete coverage

Invalid input returns `*InvalidError` with every issue found, sorted by
requirement, code and target, and no report:

| Code | Cause |
|---|---|
| `malformed-json` | Strict decoding failed (structure, types, keys, encoding, size). |
| `unsupported-format-version` / `model-id-mismatch` / `catalog-mismatch` | Header does not match this version, the model's `id` or the catalog `schema`. |
| `no-requirements` / `too-many-requirements` / `too-many-links` | Bounds. |
| `invalid-requirement-id` / `duplicate-requirement-id` | ID shape or reuse. |
| `invalid-title` / `invalid-statement` | Blank, too long, invalid UTF-8 or control characters. |
| `ambiguous-statement` | Two statements equal after case folding and collapsing whitespace. Reported on each later ID, target is the lowest ID. |
| `unknown-link-kind` / `dangling-model-link` / `link-kind-mismatch` / `duplicate-model-link` | Model link target `kind:id`. |
| `dangling-acceptance-link` / `duplicate-acceptance-link` | Acceptance link target `caseId`. |

A model that fails the existing `jsonformat`/`model` checks, or a catalog without
the expected schema, unique well-formed case IDs and `executable`/`pending`
statuses, returns a plain error. Those are inputs this tool reads, not ones it
reviews.

The catalog is read with Go's `encoding/json`, which lets a later duplicate key
win and matches keys case-insensitively. So the fields the trace consumes
(`schema`, `sequences`, `refusals`, `special`, and each entry's `caseId`, `name`
and `executionStatus`) must each appear at most once, spelled exactly. A
repeated key, an escaped spelling of the same key, or a case- or Unicode-folded
variant (for example `ExecutionStatus` or `ſequences`) is refused rather than
allowed to change a case's status. Keys the trace does not read are left to the
oracle and not checked.

Valid input always produces a report. Gaps are listed per requirement:

- `no-model-link`: no model element is linked.
- `no-acceptance-link`: no acceptance case is linked.
- `no-executable-acceptance`: cases are linked, but the catalog marks all of them
  `pending`.

A requirement with no gaps is `linked`; otherwise `incomplete`. The report also
lists model elements no requirement links and cases no requirement links. A
model requirement (`R_*`) counts as traced when a requirement links it directly
or links a guard that declares it; the report shows both separately.

## The example

[`requirements.json`](requirements.json) states twelve outcomes for the
task lifecycle. Nine are linked. Three are left incomplete on purpose:

- `UR-010` (malformed input is rejected first) has no model element, because
  input validation is not modelled, and its cases `special-03`/`special-04` are
  `pending` in the oracle catalog. The raw-input adapter later made equivalent
  checks executable in `rawinputacceptance/`, but the catalog still marks them
  pending, and this tool reports the catalog as written.
- `UR-011` (revision never wraps) has a case but no model element: the limit is
  in the generator, not in the model.
- `UR-012` (each accepted command advances the revision once) links all five
  transitions but no acceptance case. The oracle checks this in every step, but
  no catalog case is about it alone. This is the deliberately missing link.

`special-02` (multiple failed guards are sorted) is linked by no requirement.

Expected results for the example, two edited variants, 22 invalid inputs and one
valid control were written by hand in
[`expected.json`](../traceabilityacceptance/expected.json) and committed before
this package existed. The same author wrote both; "independent" here means the
expectations were not derived from tool output.

## Determinism and identity

Requirements, links, model elements (by kind, then ID) and cases are sorted.
Shuffling requirements and links, or reformatting the file, gives identical
bytes in all three views (tested with 20 shuffles). The report records:

- the model's canonical digest from the existing `canonical` package;
- SHA-256 of the exact catalog bytes;
- SHA-256 of the normalized requirements (sorted IDs, text and link targets),
  so it changes with content but not with order.

These identify inputs. They do not authenticate them or show that review took
place.

## The view

`trace.svg` is an inert SVG (no script, links or external resources) with three
columns: model states, transitions and guards; requirements; acceptance cases.
Lines are declared links. Requirements with gaps are dashed orange, unlinked
elements and cases dashed red, pending cases dashed grey. Model requirements are
left out of the drawing for space and appear in `trace.md`. The layout is loosely
modelled on a SysML requirement diagram with `satisfy` and `verify` edges. It is
not SysML or UML and makes no conformance claim.

Titles accept any valid UTF-8 without control characters, including characters
XML 1.0 cannot carry, such as U+FFFE and U+FFFF. In the SVG only, each such
character is drawn as U+FFFD, so the file stays well-formed XML. `trace.json`,
`trace.md` and the requirements digest keep the exact source text.

## Extension points

- **More acceptance sources.** `acceptanceCatalog` names one catalog schema. A
  later version could accept a list of named catalogs (for example the scenario
  fixtures in PR 12 once merged) with case IDs qualified by catalog.
- **Link relations.** v1 has one implicit relation per list. A later version
  could add a `relation` field (for example `satisfy`, `refine`, `verify`,
  `deriveReqt`) once there is a reviewed need to tell them apart.
- **Requirement hierarchy.** A `parent` or `derivedFrom` field would allow
  decomposition. Cycle and depth checks would be needed.
- **Binding to a reviewed model.** The report records the model digest. A later
  version could require the requirements file to name the digest it was reviewed
  against, so any model change forces re-review of the trace.
- **Other renderers.** `Report` is a plain Go value; additional views read it
  rather than re-deriving links.

## Toward SysML and UTP

Nothing here implements either standard. The mapping a later exporter might use:

| This format | SysML v1.x / v2 | UTP 2 |
|---|---|---|
| requirement `id`, `title`, `statement` | `«requirement»` with `id` and `text` | test requirement |
| `modelLinks` to a transition or state | `«satisfy»` from the model element | — |
| `modelLinks` to a model `R_*` | `«refine»` or `«deriveReqt»` | — |
| `acceptanceLinks` | `«verify»` from a test case | test case `«verifies»` the requirement |
| `executable` / `pending` | — | test case with or without an executable procedure |
| `trace.svg` | requirement diagram, loosely | — |

Such an exporter would need a reviewed choice of SysML version and serialization
(XMI, SysML v2 textual or API), a decision on whether lifecycle states and
transitions map to a state machine or activity, and real UTP verdict semantics.
A linked case is not a passed test verdict.

## Limits

- Links are only checked for existence and kind, not meaning.
- Only the oracle catalog's `sequences`, `refusals` and `special` case IDs are
  linkable. Raw-input and scenario acceptance tests have no catalog IDs here.
- The catalog's `executionStatus` is taken as written. The tool does not run the
  cases or check that `executable` cases pass.
- The view layout does not scale past a few dozen items per column.
- Requirement text is not checked for clarity beyond blank and duplicate
  statements.
