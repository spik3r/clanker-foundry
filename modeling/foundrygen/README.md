# foundrygen

`foundrygen` is the project entrypoint for Foundry generators. A repository that
uses generated code lists its outputs in one file, `foundry.json`. Three commands
then do the work:

- `plan` previews which outputs would be created, replaced or left alone;
- `generate` rebuilds every declared output;
- `check` fails when any output differs from what its model generates.

It builds each generator from a Foundry checkout and runs it through the
[generator contract](../CONTRACT.md) (`validate`, `generate`, `check`). It has
no generator logic of its own and uses only the Go standard library.

## Project file

`foundry.json` sits in the project root:

```json
{
  "projectVersion": "foundry-project/v1",
  "foundry": {"revision": "0123456789abcdef0123456789abcdef01234567"},
  "outputs": [
    {"generator": "tuic", "model": "design/browser.json", "out": "internal/browser"}
  ]
}
```

- `foundry.revision` is the 40-digit lowercase git commit of the Foundry
  checkout whose generators the project uses.
- `outputs` holds 1 to 64 entries. `generator` must be an id in the checkout's
  [registry](../gencontract/generators.json).
- `model` and `out` are relative paths with `/` separators, inside the project.
  They may not be absolute, empty, contain `.` or `..` segments, or contain
  backslashes.
- No two `out` paths may be equal or nested. An `out` may not equal or contain
  any `model` path or `foundry.json`.
- Existing path components must not be symlinks.
- The file is decoded strictly: unknown keys, duplicate keys, wrong types,
  trailing data and sizes over 64 KiB are refused.

## Commands

```sh
foundrygen plan     --project DIR --foundry CHECKOUT [--unpinned]
foundrygen check    --project DIR --foundry CHECKOUT [--unpinned]
foundrygen generate --project DIR --foundry CHECKOUT [--unpinned]
```

`--project` defaults to `.`. `--foundry` is required. There are no other
arguments.

All three commands load the project file and the checkout's registry, check the pin,
and build each needed generator once into a private temporary directory with
`go build`. Generators run with an empty environment and a 60 second limit.

Successful generator responses must use the exact compact identity line from
the contract: four string members in order and one final newline. Trailing
data, duplicate or missing members, wrong types and extra whitespace fail with
exit 1.

`check` runs the generator's `check` for every output and reports all of them.
Drift lines are printed to stderr with the output path in front:

```text
internal/browser: drift: modified view.go
```

On success it prints `ok OUT GENERATOR CANONICAL_SHA256` per output. It never
writes inside the project.

`generate` validates every model first, so one bad model changes nothing. Then,
per output:

- if `check` passes, the output is left alone (`unchanged`);
- if `OUT` does not exist, it is generated (`generated`);
- if `OUT` exists and holds a regular `foundry-manifest.json`, it is replaced
  (`regenerated`);
- any other existing `OUT` is refused and left untouched. A directory that was
  not generated is never replaced.

New output is written to a temporary sibling directory next to `OUT`, checked,
and then moved into place. If a step fails, the old output is restored where
possible and the message names any leftover directory. Missing parent
directories of `OUT` are created.

Output lines are `generated|regenerated|unchanged OUT GENERATOR CANONICAL_SHA256`.
On any nonzero exit stdout is empty.

## Preview before generation

`plan` uses the same preflight as `generate`: it validates every model, checks
output ownership and classifies existing output with the generator's `check`.
It writes no project outputs or output parent directories and never invokes
`generate`. It still builds the generators into a temporary directory, so Go's
build cache may change. A trusted Foundry checkout is required, just as for
`check` and `generate`.

Success returns one JSON document, in declared output order:

```json
{
  "planVersion": "foundry-plan/v1",
  "outputs": [{
    "action": "create",
    "generator": "tuic",
    "model": "design/browser.json",
    "out": "internal/browser",
    "identity": {
      "contract": "foundry-generator/v1",
      "profile": "record-browser/v1",
      "sourceSHA256": "<64 lowercase hexadecimal digits>",
      "canonicalSHA256": "<64 lowercase hexadecimal digits>"
    }
  }]
}
```

The actions are `create` for an absent output, `replace` for an owned output
with drift, and `unchanged` when check passes. A valid plan exits 0 even when it
contains changes. Invalid models, unowned outputs and other failures produce
no JSON on stdout; diagnostics and exit codes follow `generate`.

This is an observation, not a saved execution plan or permission to overwrite
files. It does not list per-file edits, guarantee generation will succeed, or
lock project inputs. Run `generate` to apply changes; it repeats the preflight
against the current files. Review generated changes before committing them.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | Everything passed or was written. |
| 1 | Failure: build, I/O, timeout, or a generator error. |
| 2 | Refused: usage, bad project file, pin mismatch, bad model, unowned `OUT`, symlinked path. |
| 3 | `check` only: at least one output drifted and nothing else went wrong. |

When outputs end differently, 2 wins over 1, and 1 over 3.

## The pin and `--unpinned`

Without `--unpinned`, `git -C CHECKOUT rev-parse HEAD` must equal
`foundry.revision`, and
`git status --porcelain --untracked-files=all -- modeling` must print nothing.
Untracked files under `modeling/` are checked even when Git is configured to
hide them with `status.showUntrackedFiles=no`. This ties the generated bytes
to a reviewed Foundry commit. A mismatch is a refusal (exit 2).

`--unpinned` skips both checks and prints a warning on stderr. It is for local
development against a working copy, and for this repository's own tests, which
use the repository itself as the checkout. Do not use it in CI for a consuming
project.

## CI example

For a consuming repository, in `.github/workflows/foundry.yml`. Replace
`REVISION` with the same commit as `foundry.revision`. Pin the action
references to commit SHAs as you do elsewhere.

```yaml
name: foundry-drift
on: [push, pull_request]
permissions:
  contents: read
jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false
      - uses: actions/checkout@v4
        with:
          repository: spik3r/clanker-foundry
          ref: REVISION
          path: .foundry
          persist-credentials: false
      - uses: actions/setup-go@v6
        with:
          go-version: '1.27.4'
          cache: false
      - name: Build foundrygen
        working-directory: .foundry/modeling/foundrygen
        run: go build -o "$RUNNER_TEMP/foundrygen" ./cmd/foundrygen
      - name: Check generated output
        run: '"$RUNNER_TEMP/foundrygen" check --project . --foundry .foundry'
```

The pin check compares the side checkout's `HEAD` with `foundry.revision`, so a
stale revision in `foundry.json` fails the job. Do not commit `.foundry/` in the
consuming repository; the job only needs it on the runner.

## Limits

- It does not review models. A model is whatever the generator's `validate`
  accepts.
- It does not build or test generated code. `check` compares bytes only.
- It trusts the generator source at the pinned revision. Isolation comes from
  the [contract](../CONTRACT.md) and review of that source, not from this tool.
- A manifest is an integrity record, not an approval.

Run its own tests with Go 1.27.4:

```sh
cd modeling/foundrygen
go test -race -count=1 ./...
```
