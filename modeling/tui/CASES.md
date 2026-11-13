# Expectation cases for record-browser/v1

A cases file states what the browser generated from a model must do and show.
It is data. A model author writes cases, never test code, and one fixed
program, `tuicases`, runs them. Format name: `record-browser-cases/v1`.

## How to write cases

Write each case from the requirement and from the [profile document](PROFILE.md)
and the [README](README.md), the statements about keys, panes, selection,
resize, replacement and help. Do not copy a case from what the generated
browser happens to do. A browser generated from a wrong model repeats the wrong
model perfectly, and cases copied from it repeat the same mistake.

A failing case is a discrepancy to investigate. Decide whether the model, the
generator or the case is wrong, from the requirement. Do not edit an
expectation to make a failing run pass.

## The file

A strict JSON document of at most 256 KiB, UTF-8. Duplicate keys, unknown keys,
keys that differ only in case, `null`, trailing data, invalid UTF-8 and numbers
written as non-integers are refused. Every member below is required unless it
says optional.

| Member | Meaning |
| --- | --- |
| `formatVersion` | `"record-browser-cases/v1"` |
| `profileVersion` | `"record-browser/v1"` |
| `model.canonicalSHA256` | 64 lowercase hex digits: the canonical digest of the model these cases were written for (see the binding rule). |
| `records` | 1 to 8 named record sets. A name matches `[A-Za-z][A-Za-z0-9_]{0,31}`. Each set is a list of 0 to 200 records. |
| `cases` | 1 to 200 cases. |

A record is an object with exactly the model's declared field ids, each a
string of at most 4096 UTF-8 bytes. The identity field must be nonempty and
unique in the set. Fields that are missing or extra, in any record, refuse the
file when it is loaded with the model.

### A case

| Member | Meaning |
| --- | --- |
| `id` | Unique, `[A-Za-z][A-Za-z0-9_]{0,63}`. |
| `requirement` | 1 to 240 bytes of plain text (no control characters): the statement the case checks. Quote or paraphrase the documented behaviour. |
| `size` | `{"width": W, "height": H}` in cells, 1 to 400 wide and 1 to 200 high. The terminal size before the first step. |
| `records` | Name of the record set the browser starts with. |
| `steps` | 0 to 64 steps, run in order. |
| `expect` | At least one expectation, checked after the last step. |

### Steps

Each step has exactly one member.

- `{"key": K}` presses a key. `K` is one of the keys the README documents, spelled
  as written here: `tab`, `shift+tab`, `enter`, `esc`, `?`, `q`, `ctrl+c`, `j`,
  `k`, `up`, `down`, `pgup`, `pgdown`, `home`, `end`. Any other spelling is
  refused, including aliases the browser accepts but the README does not
  document.
- `{"resize": {"width": W, "height": H}}` resizes the terminal (same limits as `size`).
- `{"replace": {"status": S, "records": "SetName", "message": "..."}}` hands the
  browser a new snapshot, as application code does. `S` is `ready`, `loading` or
  `failed`. `records` names a record set and is allowed only with `ready`; without
  it the snapshot has no records. `message` is optional plain text of 1 to 240
  bytes shown for `loading` and `failed`.

### Expectations

All members are optional, but at least one must be present. Each is checked
after the last step and reported separately.

| Member | Observes |
| --- | --- |
| `layout` | `wide`, `narrow` or `tiny`: both panes, one pane, or the resize message. |
| `focus` | `list` or `detail`: the focused pane. Loading and failed snapshots keep it. |
| `help` | `true` when the help overlay is open. |
| `selected` | The identity value of the selected record, or `""` when none is shown (loading, failed or empty data). |
| `status` | `ready`, `loading` or `failed`: the status of the current snapshot. |
| `visible` | 1 to 16 strings of 1 to 240 bytes. Each must appear in the rendered frame after ANSI sequences are removed. |
| `absent` | 1 to 16 such strings. None may appear in the frame. |
| `quit` | `true` when the last step was a key and the browser returned the quit command; `false` otherwise. |
| `capturesKeys` | `true` when the browser's `CapturesKeys()` is true after the last step: the help overlay is open and a host must forward every key. |
| `keyHints` | 1 to 12 objects `{"key": "...", "action": "..."}` (each 1 to 240 bytes, plain text). Each must appear in the browser's `KeyHints()`, and they must appear in the order given; other hints may lie between them. A standalone browser lists its footer hints (including `q quit`); a hosted browser has no quit hint. |

`visible` and `absent` match plain substrings within the text of the frame. For
a hosted model the frame is the body the browser draws, with no title line,
status line or footer, and `size` and every `resize` are the body size.
The frame is rows of text, and a string cannot contain a line break, so a
match lies inside one row. In a wide layout the two panes share rows. Only
observable behaviour is checked: the frame and the public read-only methods of
the generated package (`Selected`, `Focus`, `Layout`, `HelpOpen`, `Status`,
`CapturesKeys`, `KeyHints`).
Scroll offsets are not observable as numbers. Check them through `visible` and
`absent`. Colours and styles are removed from the frame and cannot be checked.

## The binding rule

`model.canonicalSHA256` binds the cases to one model. `tuicases` refuses to run
(exit 2) when it differs from the digest of the model it is given, and says the
cases are stale. This is deliberate. A changed model may change behaviour, and
the old expectations were written for the old model. Review every case against
the changed model and its requirements, and set the digest only after that
review. A digest changed without review turns the check into a copy of the
current behaviour. `tuic validate MODEL` prints the digest as `canonicalSHA256`.

### The digest when you have no tools

An author who cannot run `tuic` writes 64 zeros
(`0000000000000000000000000000000000000000000000000000000000000000`) for
`model.canonicalSHA256`. That passes the format check and is refused by
`tuicases` as stale, on purpose: a placeholder must never run as if it were
reviewed. The author does not guess a digest.

The digest is filled in when the cases are reviewed, by the reviewer: a person (or
an agent acting for one) who has read every case against the model and the
requirements and who can run `tuic validate MODEL`. The reviewer copies the
`canonicalSHA256` value from that output into the file. Nothing else may change at
that step. If the review finds a case that must change, change it first, finish the
review again, and only then set the digest. A digest filled in by the author
without that review turns the check into a copy of the current behaviour.

## Running

```sh
go run ./cmd/tuicases validate MODEL CASES   # format, binding and record checks; no build, no go needed
go run ./cmd/tuicases check MODEL CASES      # also generates, builds and runs the browser
```

`check` generates the browser from the model into a private temporary
directory (a temporary host module in package mode), writes one fixed runner
program, builds it with `go build -mod=readonly` and runs it with an empty
environment. The runner is generic: it reads the cases file at run time. Only
the import path and the model's field list come from the model, and nothing
from the cases becomes Go source. The `tuic` generator never runs code;
running is a separate command on purpose.

`go` must be installed (see Go versions in the README), and the generated
module's dependencies must be in the module cache or reachable through the
network, as for the acceptance tests. The timeouts are 5 minutes for the build
and 2 minutes for the run.

On success `check` prints `pass ID` for each case on stdout. When an
expectation fails it prints, for every case in order and on stderr,
`pass ID` or one `fail ID: EXPECTATION: want ... got ...` line per failed
expectation; for `visible` and `absent` the line names the string. Stdout is
empty on any nonzero exit.

| Exit | Meaning |
| --- | --- |
| 0 | Every case passed. |
| 1 | Build, run or internal failure. |
| 2 | Refused: usage, bad model, bad cases, digest mismatch, records that do not match the model's fields. |
| 3 | At least one expectation failed. |

The [authoring schema](authoring/record-browser-cases-v1.schema.json) is a
first check. It lists what it cannot express; `tuicases validate` is
authoritative. Examples: [browser](examples/browser.cases.json) and
[package browser](examples/browser-package.cases.json) and
[hosted browser](examples/browser-hosted.cases.json), whose cases are written from
the hosting contract in the README.
