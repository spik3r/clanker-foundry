# Structural lifecycle topology

This offline analyzer summarizes a validated model's **declared state graph**.
It does not execute transitions, evaluate guards or effects, solve satisfiability,
authenticate a model, or establish actual reachability, safety, deadlock freedom,
fairness, termination or liveness. A structurally present path can be infeasible
because of guards or runtime input. A structural cycle or lack of an exit is not
proof that an execution enters, leaves or remains in it. This is a review aid,
not a scheduler, visual projection, approval, or standards-conformance claim.

## Results

`Analyze(model.Model) (Report, error)` normalizes and validates the complete
existing `lifecycle/v1` IR and computes the existing canonical model digest. An
invalid model produces no partial report. Existing model validation requires all
states to be structurally reachable from the initial state, at least one terminal,
and no outgoing transition from a terminal. It does **not** require a terminal to
be structurally reachable from every state, nor prove any guard satisfiable.

The report uses `mode: structural-only`, `authority: none`, a model ID, digest,
initial state, sorted state entries, and sorted cyclic components. The digest is
a consistency identity for normalized content, not proof of provenance or review.
For each state:

- `hasTerminalPath` says whether the declared directed graph has a path to any
  terminal. `terminalId` and `transitionIds` identify one **shortest structural
  path**. Ties use the lexicographically smallest transition-ID sequence; source
  collection order does not matter.
- A terminal has a path to itself with an empty transition list. A state with no
  structural path has `hasTerminalPath: false`, an empty terminal ID, and an
  empty transition list. This distinction is intentional.
- The path is not an executable witness. It ignores every guard and does not
  select actual commands, actors, revisions or execution context.

Each cycle is a maximal strongly connected component of two or more states, or a
single state with an explicit self-loop. Components sort by their sorted state
IDs. State IDs, internal transition IDs and outgoing transition IDs are each
sorted. An outgoing transition is a directly declared edge from a component
state to a state outside it; an empty list means **no direct structural exit**.
Neither an exit nor a terminal path establishes an actual escape or eventual
completion. A cycle can include terminal-infeasible paths or impossible guards.

The typed result owns all returned slices. Callers must not mutate inputs
concurrently. IDs and digests can be sensitive; keep output private when inputs
are private. There is no persistence, network, child process, provider, worker,
image or visual operation in the analyzer or CLI.

## JSON and command

`Execute([]byte)` accepts exactly one wrapper, with both keys present and
non-null:

```json
{"formatVersion":"lifecycle-structural-topology-request/v1","model":{}}
```

Replace the placeholder with a complete existing `lifecycle-json/v1` model.
Strict decoding rejects duplicate, missing, case-variant or unknown fields,
trailing values, invalid UTF-8 or surrogate escapes, invalid models and profile
mismatches. Request input is capped at 2 MiB and depth 66, with the existing
1 MiB model bound; complete buffered output is capped at 1 MiB. The existing
model limits (16 states and 32 transitions) bound graph analysis.

From the module root with Go 1.27.3 installed:

```sh
python3 -c 'import json; print(json.dumps({"formatVersion":"lifecycle-structural-topology-request/v1","model":json.load(open("examples/lifecycle.json"))}))' | GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltopology
```

The command takes no arguments or paths and reads bounded stdin. Exit 0 reports a
valid graph, including graphs with cycles or states lacking terminal paths; it is
not an acceptance decision. Exit 2 rejects arguments, toolchain or input and
produces no stdout. Exit 3 reports failed or short output; a stream writer can
still expose a prefix despite buffering. Diagnostics are fixed categories and do
not echo input or I/O errors. The example shell command reads a file, but the Go
program itself does not.

## Validation

The graph fixture and literal expected paths and components were hand-authored by
the implementation contributor and frozen before mapping; they were not
independently authored. The fixture covers equal-length paths, a closed two-state
cycle, an exit-bearing cycle and an exit-bearing self-loop, as well as terminal
and no-terminal-path states. Additional tests exercise ordering invariance,
parallel edges, impossible guards, invalid models, input ownership, strict JSON,
deterministic output and failed I/O. These are mapping regressions, not a proof
of model behavior. With the exact local toolchain:

```sh
go test -race -count=1 ./topology ./cmd/modeltopology
```
