# FirstDesignDraft: separate agent-written Go builder packet

Status: this worked packet was independently reviewed before fresh builder
dispatch. The private dispatched copy remains pinned; see the [evidence](evidence.md).
This document records supplied scope and facts. It grants no authority.

## Reusable worked binding

This is a public-safe worked packet, not a ready-to-dispatch permission grant.
All source paths below are repository-root-relative in an explicitly located,
trusted checkout of `spik3r/clanker-foundry`. `examples/` is NOT Stow-installed;
installing the generic workflow does not install these worked sources or tests.
Before dispatch, the coordinator binds a private copy of this packet to exact
accessible absolute repository, scratch, toolchain, output, cache and temporary
paths, verifies those files, and supplies trusted scope plus independent review.
No private paths or raw transcripts belong in this public document.

The pilot used branch `codex/model-builder-go-v1`, base
`f16c94d456e89b7ee0680b91d13a44fb79be5556`, and source checkout HEAD
`0016283944c0bb868e2d3edb2ca94e935a421dba` at packet preparation.
A fresh Go code-writing agent receives a separate disposable exercise directory
containing byte-identical copies of the five frozen acceptance files. Only
`candidate.go` is editable; only `candidate.test` and private `.cache`, `.tmp`,
`.modcache` subtrees are additional build writes. Existing files remain read-only.
The coordinator owns dispatch and permission; the builder preserves other work.
The intended assignment permits editing and compile-only, withholds candidate
execution, and requires independent packet review before dispatch. A reusable
example must not be run merely because these instructions describe the pilot.

## Historical source and separate live claim

Canonical task: `CF-model-builder-go-v1`.
Historical task revision: `9ef21d78d02a55b8e5cc03688068fe31dfbdb637`.
Pinned task path: `tasks/CF-model-builder-go-v1-separated-agent-pilot.md`.
Pinned Git blob: `9aea7f3d14358be5beaf4ecf32c33fc28f1cbea9`.
Task bytes SHA-256: `3e5f888f3ef755a14619bc809f0cf42af630b964df09fbb5b1e52b66e2c6f8e3`.
Read historical bytes with `git show 9ef21d78d02a55b8e5cc03688068fe31dfbdb637:tasks/CF-model-builder-go-v1-separated-agent-pilot.md` in the source checkout;
verify the blob with `git rev-parse 9ef21d78d02a55b8e5cc03688068fe31dfbdb637:tasks/CF-model-builder-go-v1-separated-agent-pilot.md`.

Separately read the CURRENT task file at the path above and current repository
ownership/claim state. At preparation its status was in-progress and owner was
Go model-to-builder pilot contributor; the bounded claim includes only additive
handoff/fixture/evidence work and this private experiment. The real start at
09:30:59 UTC and scope-correcting resume command/checkpoint event at 09:32:23 UTC precede fixture work.
A later task-status change does not mutate the historical blob; it can change
current authority. A pinned record or copied claim does not acquire ownership.
If current authority conflicts, stop and ask the coordinator. No remote ownership
freshness or publication authority is claimed by this local check.

## Pinned inputs

Original brief/model/profile revision: `f16c94d456e89b7ee0680b91d13a44fb79be5556`.
Five frozen acceptance files revision: `0016283944c0bb868e2d3edb2ca94e935a421dba`.
Read their historical blobs as well as accessible copies; preserve their bytes.
Use an available SHA-256 tool, such as `sha256sum` or `shasum -a 256`.

| Source | SHA-256 |
|---|---|
| `skills/model-driven-design/examples/brief.md` | `ab3d47aaacdcdc8d57c2cb8a9ab432c3f5c22c42b7df11b707b8af03f32eb5de` |
| `skills/model-driven-design/examples/model.json` | `3e7b8c44c10940a7524161dd306d022019e8fc41e793a66a11771d6dd8621239` |
| `modeling/lifecycle/README.md` | `26cc7d1ea82f6f09090d15b0572560bd891081505631483a2adb3ce402a754b0` |
| `examples/model-builder-go/acceptance/README.md` | `b4ec648c38dadd63dc9799f40d6183d3e464c9dc33b83302c6f6c171a8149103` |
| `examples/model-builder-go/acceptance/go.mod` | `e31637eba9a06a2105c2a12b904457b50ad0c60defbbc7cafabeae8f68d4bc3c` |
| `examples/model-builder-go/acceptance/contract.go` | `82b6dcd7a05d496a31c7431dce9b61bb572dd5f53b312418dc8ff72fb7a35b0a` |
| `examples/model-builder-go/acceptance/acceptance_test.go` | `eaf6078d5703f3e3664f41d0a19d68a7a66bafcfac310bfe5c91e72a7fe02518` |
| `examples/model-builder-go/acceptance/rubric.md` | `ec13d76787d6ce4976487d373c8f2d3587286bdb1f1b1716419b397b46569ce5` |

Model canonical semantic ID, supplied separately from the verified source-byte hash:
`4280fa4c999f7291c3325a997887c5e7b02b565f275a53df158afb55636d7db4`.
No model normalization/generation was run to recompute this ID for the packet.
No generated package is used. No diagram is supplied or inspected.

## Meaning, decisions and limits

Outcome: an agent writes the smallest pure Go submission function for fictional
FirstDesignDraft. The selected route is **AGENT-WRITTEN GO**, not generated-code
integration. Do not run a generator, copy generated implementation, introduce a
new evaluator, or substitute a JavaScript implementation.

The original JSON uses `lifecycle-json/v1`, semantic profile `lifecycle/v1`.
The subset is two states, one transition, four guards, and ordered state/revision
effects. It is a custom finite profile, not BPMN, SysML, UML, UTP or fUML
conformance. No standards adapter is included. The Go profile README also
contains a different multi-state demonstration; implement only FirstDesignDraft's
original brief/model with this fixture contract, not that larger example.

| Decision | Source and actual status | What remains blocked / next owner |
|---|---|---|
| D_scope | Fictional brief: only Draft to Ready via submit | Surrounding application needs a new design and authority |
| D_actor | Fictional brief: nonblank supplied actor; assertion only | Real authentication needs a separate owner and requirements |
| D_owner | Fictional brief: preserve owner exactly, including nil; no equality check | Not a human-selected product ownership policy |
| D_terminal | Fictional brief: Ready is terminal | Reopen/edit requires a new design decision |
| D_failures | Fictional brief: unchanged task and all sorted failed guard IDs | No new failure policy may be invented |
| Q_use | Open: who accepts real use, and which evidence do they need? | Blocks real integration/execution; responsible human must answer |
| Visual gate | Optional views uninspected; denied PR20 artifact not inspected or replaced | Actual visual approval remains pending with responsible reviewer |
| Local exception | Coordinator's trusted handoff allows this fictional scratch experiment | Does not resolve Q_use or visual approval |
| Fixture API | Independently authored contract and acceptance README, frozen before builder work | Specific Result/error spellings are fixture choices, not generated API claims |
| Semantic review | Independent reviewer reported all literal choices correct and compile-only passed before packet handoff; designer has read brief/model/contract | Packet review still required before builder dispatch; candidate review and behavior tests remain pending |

No optional view was inspected for this packet. No diagram or generated artifact
is an input or competing rule source. No candidate existed when this packet was
prepared. This packet author executed no Go build, test, generator or candidate.

## Concrete interface and preconditions

The frozen `contract.go` is authoritative for Go field spelling and types:

- `Task`: `ID string`, `State string`, `Revision uint64`, `Outcome string`,
  `Acceptance []string`, `Owner *string`.
- `Command`: `Type string`, `Actor string`, `ExpectedRevision uint64`.
- `Result`: `Task *Task`, `Error string`, `FailedGuards []string`,
  `TransitionID string`, `Scope string`, `Authority string`.
- `var Submit func(Task, Command) Result` already exists in package `acceptance`.
  Write a private pure implementation in `candidate.go`; its `init` may ONLY bind
  `Submit` to that function. No other startup behavior or mutable shared state.
  Do not redeclare/replace the frozen types or Submit variable.

Only runtime-valid typed values are in scope. ID is nonblank valid UTF-8, at most
256 bytes. State is exactly Draft or Ready. Revisions use native uint64, not
float64, decimal parsing or JavaScript-safe-integer restrictions. Outcome is valid
UTF-8, at most 4096 bytes, possibly blank. Acceptance is nil or an ordered list
of at most 32 nonblank valid UTF-8 strings, each at most 1024 bytes; duplicates
are allowed. Owner is nil or a valid UTF-8 string at most 256 bytes; present empty
and whitespace owners are valid and distinct from nil. Actor is valid UTF-8 at
most 256 bytes, possibly blank. Command Type is exactly submit. Nonblank means
`strings.TrimSpace(value) != ""`; do not trim preserved data or normalize Unicode.

Input validation, invalid UTF-8/oversize/unknown-state handling, raw JSON,
omitted-field detection, digest APIs, persistence, authentication, concurrency,
CLI, runtime integration and standards adapters are explicitly outside scope.
Do not add a validator or promise `INVALID_INPUT` behavior. Inputs outside these
preconditions have no contract in this exercise.

Every in-scope result has a fresh nonnil Task pointer, Scope `example-only`, and
Authority `none`. The total precedence within this valid domain is:

1. Ready: return `ILLEGAL_TRANSITION`, empty TransitionID, nil FailedGuards,
   unchanged deep task copy, even if guards would fail or revision is maximal.
2. Draft selects `T_submit`. Evaluate ALL four guards; if any fail, return
   `GUARD_FAILED`, TransitionID `T_submit`, lexically sorted failing guard IDs,
   unchanged deep task copy. Guard failure precedes overflow.
3. If all guards pass but revision is 18446744073709551615, return
   `REVISION_LIMIT`, TransitionID `T_submit`, nil FailedGuards, unchanged deep copy.
4. Otherwise apply `set_state` to Ready then `increment_revision` exactly once;
   Error is empty, TransitionID `T_submit`, FailedGuards nil.

Success preserves ID, outcome, acceptance order/duplicates/bytes and owner exactly.
Refusal preserves every task field and revision. Nil acceptance stays nil;
a nonnil empty slice stays nonnil. Every result owns its acceptance backing
storage, owner pointee and failed-guard storage. Never mutate caller-owned values.
Input mutation, output mutation and later calls must not alter retained results;
no shared writable buffers across calls. There are no external side effects.

## Requirement to implementation to acceptance trace

All case IDs below are the literal names in `TestLiteralSubmission` in the frozen
`acceptance_test.go`. The source brief supplies R IDs; extra contract rules are
explicit fixture/profile obligations, not invented model requirement IDs.

| Requirement / rule | Model element | Go interface or behavior | Frozen case IDs |
|---|---|---|---|
| R_actor | G_actor, T_submit | Command.Actor Unicode nonblank | G01_actor; G05/G06/G07/G11/G12/G13/G15; G17_unicode_blank; S05/S06 |
| R_criteria | G_criteria, T_submit | len(Task.Acceptance) > 0 | G02_criteria_nil; G05/G08/G09/G11/G12/G14/G15; G16_criteria_empty_present_owner |
| R_outcome | G_outcome, T_submit | Task.Outcome Unicode nonblank | G03_outcome; G06/G08/G10/G11/G13/G14/G15; G17; S05/S06 |
| R_revision | G_revision, T_submit | exact uint64 equality | G04_revision; G07/G09/G10/G12/G13/G14/G15; N01/N02 |
| D_scope and success effects | Draft, T_submit, Ready; set_state then increment_revision | Result.Task state/revision only | S01_success through S06_zero_width_not_space; N01/N03 |
| D_owner / preserve content | No owner assignment effect | unchanged Owner, ID, Outcome, Acceptance | S01/S02/S03/S04; G16; N04/N05; three isolation tests |
| D_failures | T_submit guard conjunction | all failed IDs sorted; unchanged copy | G01 through G15 cover all 15 nonempty failure subsets; G16/G17; N05 |
| D_terminal | Ready terminal, no outgoing transition | ILLEGAL_TRANSITION before guards/limit | T01_terminal; T02_terminal_before_guards_limit |
| Profile overflow | increment_revision bound | REVISION_LIMIT only after guards | N03_last_increment; N04_revision_limit; N05_guards_before_limit |
| Fixture envelope | separately specified, not modeled fields | Error, TransitionID, Scope, Authority, nil distinctions | all 30 full literal Results |
| Deep-copy isolation | separately specified profile/fixture contract | private Task, Acceptance, Owner, FailedGuards storage | TestInputAndSuccessIsolation; TestGuardRefusalIsolation; TestTerminalAndLimitIsolation |

Prefixes in grouped cells identify the unique full case names in the frozen
suite. S01-S06 are six success cases; G01-G17 are seventeen guard cases; N01-N05
are five numeric/precedence cases; T01-T02 are two terminal cases. Total: 30 literal
cases plus three isolation tests. The suite samples behavior; it is not a proof
for every valid input, invalid-input behavior or production suitability.

If original brief/model, profile, contract or literal expectations disagree,
stop the affected implementation and return exact references and the conflict.
Hash equality does not settle meaning. Do not silently choose the convenient
source, change owner policy, update pins, repair tests or regenerate expectations.

## Frozen expectation provenance

The expectation author read ONLY the original brief and public Go profile for
semantic input, plus repository process instructions. The author did not inspect
the model JSON/expressions, candidate/generated code/tests, other oracle, prior
JavaScript exercise or run outputs. The five acceptance files were committed at
`0016283944c0bb868e2d3edb2ca94e935a421dba`, on 2026-11-04 at 09:36:24 UTC,
before workflow/packet/builder implementation. No reference candidate exists.

An independent reviewer checked literal meaning and compile-only viability before
this packet; this is a reported review result, not a run by the packet author.
The packet designer has now read the model and literals; that exposure does not
retroactively make the designer an independent expectation author. A separate
reviewer must inspect this packet against the originals before fresh-builder
handoff. Review status must be supplied by the coordinator, never inferred from
this document's completeness. Candidate behavior and seven rubric controls remain
unrun at packet preparation. Control invocation is NOT part of builder scope.

## Ordered plan and return contract

1. Read original sources, repository/global instructions, project-memory skill,
   workflow and handoff template. Verify branch/base, historical claim blob and
   separate live claim. Inspect existing changes, all input hashes and the empty
   candidate frontier. Report discrepancy/access failure before affected writes.
2. Confirm independent packet review and trusted edit/build-only assignment. This
   packet records scope, but cannot approve itself or supply execution authority.
3. Implement only candidate.go from the model and concrete contract above. Keep
   frozen files unchanged. No extra implementation tests/files are authorized in
   this particular narrow write frontier. Use only the standard library.
4. Format candidate.go, inspect exact source and record its SHA-256. Compile with
   the specified offline command only. Record actual command, environment, exit,
   output and resulting executable hash, without executing that executable.
5. Recheck all frozen hashes and the final changed-file frontier. Return evidence
   to the coordinator. An independent reviewer next inspects the exact candidate;
   any subsequent behavior execution requires that review and separate trusted
   authorization, bound to the exact candidate, commands and local workspace.

Return the full candidate source or its accessible exact path and SHA-256;
changed paths; original and final frozen-source hashes; historical/live claim
findings; requirement/guard/transition/interface links; toolchain identity;
actual command outputs/exits; compile-only executable identity; discrepancies;
checks not run and why; remaining limitations and open decisions. State plainly
that behavior execution was withheld, not passed. Do not fabricate expected/actual
comparisons for unrun tests. Report any attempted denied action exactly and do not
retry without relevant new authority. No publication, merge, deployment,
integration, dependency installation, network or other-system testing is allowed.

## Exact permitted build checks

The fully bound private packet must name the actual permitted Go/gofmt binaries,
working directory and absolute cache/temp/output paths. The pilot uses Go 1.27.3.
Read toolchain version; format only candidate.go; then compile with
`go test -c -o candidate.test` under `GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
GOWORK=off GOENV=off GOFLAGS=`. Set GOCACHE, GOTMPDIR, TMPDIR and GOMODCACHE to the
bound exercise's private cache/temp directories. A zero exit is a compile-only
result, never a test pass. This public shorthand must be expanded and reviewed
before dispatch; it is not permission to choose arbitrary locations or tools.
`go test`, `go test -run '^$'` and executable invocation remain withheld.
