# Generator conformance harness

This directory holds a black-box test harness for the
[Foundry generator contract v1](../CONTRACT.md), a reference generator that
conforms to it, and the registry of generators the harness checks.

The harness builds a generator, runs it from the command line, and compares what
it observes with the contract. It does not read generator source and generators
do not import it. `Check(g, scratch)` returns a list of findings, each with a
stable rule ID, and returns an error only for harness problems such as an
unstartable binary.

Child processes run with an empty environment, a 30 second limit and an empty
working directory.

## Run it

Go 1.27.4 is required.

```sh
cd modeling/gencontract
go test -race -count=1 ./...
```

`TestRegisteredGeneratorsConform` builds every generator in
[generators.json](generators.json) and fails with all of its findings.
`TestHarnessDetectsFaults` builds `cmd/refgen` with `-tags refgenfault`, injects
one fault at a time through `REFGEN_FAULT`, and requires the matching rule to
appear. Normal builds of `refgen` contain no fault code.

## Rules

| Rule | What is checked | Contract section |
| --- | --- | --- |
| `validate.exit` | `validate` on the sample model exits 0 | Commands |
| `validate.identity` | stdout is exactly one identity line: four members in order, byte-exact encoding, contract value, profile, `sourceSHA256` of the model bytes, 64 lowercase hex `canonicalSHA256` | Identity line |
| `validate.whitespace` | a whitespace-only reformatted copy has the same `canonicalSHA256` and a different `sourceSHA256` | Model identity |
| `generate.exit` | `generate` exits 0 into a new `OUT` | generate |
| `generate.identity` | the identity line is valid and agrees with `validate` | Identity line |
| `generate.deterministic` | two runs with other output paths, model paths, file names and whitespace give byte-identical trees | Model identity |
| `generate.tree` | output holds only regular files and non-empty directories | check |
| `generate.existing-out` | an existing empty directory, file or symlink at `OUT` gives exit 2 and is left untouched | generate |
| `manifest.encoding` | the manifest exists, decodes strictly, has no trailing data and re-encodes to the same bytes | Manifest |
| `manifest.fields` | version, profile, `generator.id`, `version`, `toolchain` and `model.canonicalSHA256` are valid and consistent | Manifest |
| `manifest.files` | files are sorted, nonempty, have valid paths, matching `sha256` and `bytes`, and are exactly the files on disk | Manifest |
| `header` | every `.go` and `.ts` file starts with the generated-file header | Generated file header |
| `check.clean` | `check` exits 0 on fresh output | check |
| `check.identity` | `check` prints a valid identity line equal to `validate`'s | Identity line |
| `check.writes` | `check` leaves fresh output unchanged | check |
| `check.modified` | a changed byte in any file, manifest included, gives exit 3 and `drift: modified PATH` | check |
| `check.missing` | a deleted file gives exit 3 and `drift: missing PATH` | check |
| `check.unexpected` | an extra file or directory, at the top level or in a subdirectory, gives exit 3 and `drift: unexpected PATH` | check |
| `check.symlink` | a symlink in place of a file gives exit 3 and `drift: unexpected PATH` | check |
| `check.absent` | a missing `OUT` gives exit 3 and `drift: missing .` | check |
| `check.repaired` | a failed `check` leaves the tampered tree as it found it | check |
| `refuse.invalid-model` | empty, truncated, `{}` and `[]` models give exit 2 from all three commands, and `generate` creates no `OUT` | Exit codes, generate |
| `refuse.missing-model` | a missing model file gives exit 2 from all three commands | Exit codes |
| `refuse.usage` | wrong argument counts, an unknown verb and `--help` give exit 2 | Commands, Exit codes |
| `readonly` | `validate` and `check` add or change nothing in the model's directory or the working directory | Commands |
| `run.timeout` | a command finishes within 30 seconds | not in the contract; harness limit |

Every nonzero exit is also checked for empty stdout and a nonempty stderr, and
is reported under the rule of the case being run.

## Add a generator

1. Write the generator so that it builds with `go build -o BIN PACKAGE` from its
   module directory, and add a small valid sample model to its module.
2. Add an entry to [generators.json](generators.json). Paths are relative to the
   repository root.

   ```json
   {
     "id": "mygen",
     "profile": "my/v1",
     "module": "modeling/mygen",
     "package": "./cmd/mygen",
     "model": "modeling/mygen/examples/model.json"
   }
   ```

3. Run `go test -race -count=1 ./...` here. The registry file is decoded
   strictly, so unknown members fail the test.

`cmd/refgen` is the smallest working example.

## Limits

- The harness cannot show isolation. No network, no child processes, no
  execution of emitted code and no model-controlled output paths are properties
  of the generator source. Review establishes them, not this harness.
- Conformance is not correctness. A conforming generator is repeatable and its
  output can be checked for drift. It may still emit code that is wrong, from a
  model that is wrong.
- Determinism is observed on one host and one input. Dependence on time, host
  or iteration order that did not show up in these runs is not ruled out.
- A generator that detects the harness can pass dishonestly. The harness does
  not defend against that.
