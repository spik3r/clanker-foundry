---
id: CF-json-requirement-change-v1
status: in-progress
owner: Foundry requirement-change regression contributor
repositories: [spik3r/clanker-foundry]
---

# Retain JSON validator requirement-change coverage

## Implementation claim

The coordinator assigned this bounded maintained regression after reviewing
current public main `7778beb4038f081ca0f36415a286811e8627fac6`.
The receiving contributor accepts ownership on
`astra/json-requirement-change-v1`. The original standalone producer task is
done and releases only its completed producer scope. No open public pull request
or other current task claims this new regression. These checks do not reserve
unpublished work; recheck shared ownership before publication.

Publish this task source before implementation, verify its actual remote commit,
register that exact revision and record a genuine Project Memory start before
editing maintained tests or documentation. This claim authorizes a draft PR,
not merge, deployment, consumer adoption or changes to approval settings.

The exact maintained frontier is:

- `modeling/jsonvalidator/cmd/jsonvalidator/requirement_change_test.go` (new).
- `modeling/jsonvalidator/README.md`.
- This task, `tasks/CF-json-requirement-change-v1-maintained-regression.md`.
- New immutable source snapshots and lifecycle events for this task under
  `.project-memory/tasks/CF-json-requirement-change-v1/`.

Preserve production code, profile semantics, existing tests and frozen fixtures,
all workflows, other tasks, prior source/events and legacy history. Do not access
or import private consumer material. No new profile, wrapper, dependency or
runtime authority is introduced.

## Intended result

Retain three small public synthetic model revisions: one optional boolean field,
the same field required, then required with constant true. An independently
reviewed literal table states accepted and refused JSON values for every
revision, including explicit narrowing and unchanged behavior. Expectations
must not be derived from generated output or a second model interpreter.

Exercise the real generator and both emitted languages using existing test
helpers where suitable. Check exact raw-model provenance, changed generated
bytes, deterministic restoration of the original model and preservation of
earlier outputs. Keep the test focused and its runtime proportionate; do not
replay the separate exploratory campaign or expand production semantics.
The existing Go test workflow already discovers this test.

Document the maintained coverage and narrowly replace obsolete draft wording
with the verified merged-producer state. Do not claim a tagged release,
consumer activation, browser acceptance or universal language equivalence.

## Validation and handoff

Use the pinned Go 1.27.3 and Node 24. Run the focused regression, full module
`go test -race -count=1 ./...`, `go vet ./...`, frozen TypeScript acceptance,
formatting, repository validation and whitespace checks. Record unavailable
optional checks explicitly. Check the full committed diff with Project Memory
against actual `origin/main`, using code/source identity
`spik3r/clanker-foundry`.

Stage only the named scope, record observed completion and commit fresh
immutable evidence with the implementation. Publish a draft PR, verify its
actual head and report local and hosted checks separately. An independent
reviewer checks the literal expectations before execution and the exact final
source before integration. No result in this source is a claim that those
future checks have already run.
