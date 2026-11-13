---
id: CF-record-browser-cases-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Add expectation cases as data for the record browser

## Implementation claim

Owner: Claude pipeline coordinator, with delegated implementation by a
supervised subagent and coordinator review of the diff. This closes the item
left open in step 2 of the
[generator pipeline plan](../docs/generator-pipeline-plan.md): expectation
literals as data for `record-browser/v1`.

Scope: a cases format `record-browser-cases/v1` bound to a model's canonical
digest; a `tuicases` command under `modeling/tui/cmd/` whose `validate` checks a
cases file against a model and whose `check` generates the browser, builds one
fixed runner and reports each case; read-only accessors on the generated model
so behaviour is observable; `tuic` 1.2.0 with unchanged canonical digests; two
example cases files; a format document, an authoring schema and tests; an
optional `cases` member in the profile registry and its consistency test; one
paragraph in the design skill's profiles reference; the workflow line that runs
the new schema test. Also this task record and new immutable Project Memory
files.

Exclusions: the existing Go acceptance suites are kept and not ported; no
change to the generator contract, the harness or other generators; `tuicases`
is not a generator and is not registered as one.

Checks: gofmt, `go vet` and `go test -race` in `modeling/tui`, `go test -race`
in `modeling/gencontract` and `modeling/foundrygen`, the node schema tests, the
workflow inventory test, the repository validator, whitespace check and
committed Project Memory validation.
