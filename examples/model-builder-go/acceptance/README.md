# Frozen Go submission acceptance

This is a fictional, limited local design-to-code exercise. It does not supply real
human acceptance, visual review, product approval, authentication or authority.
It uses the existing `lifecycle/v1` Go profile, not a second interpreter. Go was
chosen for this exercise; this is not a JavaScript-to-Go migration.

## Semantic sources and independence

The expectation author read only these semantic sources:

- [First-design brief](../../../skills/model-driven-design/examples/brief.md)
- [Go lifecycle profile](../../../modeling/lifecycle/README.md)

Repository process instructions were read separately. The author did not inspect
model expressions, model JSON, any candidate, generated code/tests, a prior oracle,
a JavaScript exercise or other acceptance artifacts. Expected results below were
written as literals from those two documents, not obtained by running code. Freeze
all five files in this directory before a fresh builder sees the accepted model.
A nil Submit lets compile-only review happen without a reference implementation.

## Preconditions and exclusions

`Submit(Task, Command) Result` accepts only runtime-valid typed Go values. Task
ID is nonblank UTF-8, at most 256 bytes. State is Draft or Ready. Revision and
expected revision are native uint64. Outcome is valid UTF-8 of at most 4096 bytes.
Acceptance is nil or an ordered list of at most 32 valid UTF-8, nonblank items,
each at most 1024 bytes. Duplicate items are valid. Owner is nil or a valid UTF-8
string of at most 256 bytes, including present empty or whitespace strings. Actor
is valid UTF-8 of at most 256 bytes, including blank strings. Command Type is submit.
Blank outcome and actor, and nil or empty acceptance, are valid inputs whose
modeled guards can fail. Nonblank uses Go strings.TrimSpace Unicode semantics.

Unknown states/commands, invalid UTF-8, excess sizes, invalid acceptance items,
JSON parsing, omitted-field detection and input validation are excluded. No
INVALID_INPUT case is tested or promised at this API boundary. Passing this suite
cannot establish production readiness or behavior outside these preconditions.
Persistence, authentication, concurrency, integration, digest APIs and generation
are outside this reducer fixture.

## Frozen fixture choices

The sources describe semantics but do not spell out a concrete Result struct or
the error name for a modeled guard refusal. contract.go therefore freezes a local
adapter interface: Task fields ID, State, Revision, Outcome, Acceptance and Owner;
Command fields Type, Actor and ExpectedRevision; Result fields Task, Error,
FailedGuards, TransitionID, Scope and Authority. Task and Command are passed by
value. Result.Task is a new nonnil pointer for every in-scope call.

Error is empty on success, GUARD_FAILED for modeled guards, ILLEGAL_TRANSITION
for terminal Ready, and REVISION_LIMIT for overflow after all guards pass.
TransitionID is T_submit whenever Draft selects submit, including guard and
revision-limit refusals; it is empty for terminal refusal. FailedGuards is nil
unless guard refusal returns sorted failing IDs. Scope is example-only and
Authority is none on every result. These spellings and nil-envelope choices are
fixture-level decisions, not assertions about unseen generated field names.

Success changes only Draft to Ready and increments revision exactly once. Owner
is preserved, including nil, empty and whitespace. No actor-owner equality is
required. Refusals preserve every task field and revision. Nil acceptance remains
nil; a nonnil empty list remains nonnil. Each result owns fresh acceptance and
owner storage. Mutating input storage, one result or its failure list must not
change any other retained result. Calls must not mutate input storage. No external
side effects belong in this pure reducer.

## Evidence and execution

acceptance_test.go contains 30 literal table cases and three isolation tests.
The cases include each guard alone, every pair, every triple, all four failures,
terminal and overflow ordering, Unicode whitespace, nil/present distinctions,
large exact integers and max-uint64 boundaries. Isolation assertions compare full
literal Results and full literal input Tasks. Helpers allocate only literal owner
strings; no helper computes expected transitions.

Compile-only review may use `go test -c -o /tmp/foundry-go-acceptance.test` with the
supplied toolchain and offline, private caches. This creates a test executable but
does not run package initializers or tests. Do not replace this with `go test`,
`go test -run '^$'` or direct candidate invocation when execution is withheld.
A later execution-authorized builder may add candidate.go in package acceptance,
register its private implementation in init, then run the frozen suite. Do not
change the five frozen files to fit candidate behavior. The seven separate fresh
agent controls are defined in rubric.md; test passes do not establish those controls.
