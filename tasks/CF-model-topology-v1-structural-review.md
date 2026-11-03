---
id: CF-model-topology-v1
status: in-progress
owner: Codex structural model-analysis contributor
repositories: [spik3r/clanker-foundry]
depends_on: [lifecycle-v1 model normalization and canonical identity]
---

# Inspect lifecycle model topology without claiming liveness

## Prospective implementation claim

Local-only lane assigned after coordinator confirmation and live ownership check:
current Foundry main is `fcda4be584ebbbdf84239a5b849b1e6b8fba2b9b`, tree
`bb13598e1434f7646f9f3dd5afee63608c025a04`. This new branch starts from that
actual commit, not a synthetic tree-equivalent ancestor. Open #5 and #20, its
denied visual artifact, scenario-suite diff, UML, model merge and other claims
are excluded. A local claim is not a shared lock. Recheck live main and claims
before separately authorized publication; no push, PR, merge, deployment,
provider or worker activation is authorized by this local task.

An earlier local prototype was developed from a synthetic commit with the same
main tree. Its hand-authored oracle was frozen at `a8d29df` before its mapping;
prototype candidate was `21d8b8e7147d68860aeebf88180a4ee1537b21a4`. Those
commits and their events remain untouched in their original workspace and are
not ancestors of this branch. Their fixture and implementation were authored by
this contributor, not an independent reviewer, and predate this new start. After
this claim is committed, register it and record a genuine start for the six
separate paths below before bringing over the frozen fixture and prototype.
Record this provenance honestly; do not backdate or imply the oracle was first
authored after the new start. Obtain independent exact-source review of the new
candidate and its inherited oracle.

Exact maintained scope, in addition to this task's own new Project Memory source
and immutable events:

- `tasks/CF-model-topology-v1-structural-review.md`
- `modeling/lifecycle/topology/analysis.go`
- `modeling/lifecycle/topology/analysis_test.go`
- `modeling/lifecycle/topology/README.md`
- `modeling/lifecycle/cmd/modeltopology/main.go`
- `modeling/lifecycle/cmd/modeltopology/main_test.go`

No changes to existing model semantics, compiler, generated artifacts, existing
claims, scenario inventory or model merge. No browser or visual artifact.

## Outcome and semantics

Build a bounded pure Go inventory of a completely validated and normalized
`lifecycle/v1` model's **declared graph**. Reuse `model.Normalize`,
`canonical.Digest` and `jsonformat.Decode`; do not implement another validator,
reducer, guard evaluator or expression solver. Every state is structurally
reachable from the initial state because the current validator requires it, but
guards can make declared paths infeasible. No execution, actual reachability,
authentication, approval, fairness, safety, termination or liveness is proved.

For each state, report a shortest declared path to any terminal, with ties
broken by lexicographically smallest transition-ID sequence. A terminal has an
empty path to itself. A state with no terminal path has false status and empty
terminal/path fields. Report cyclic maximal strongly connected components:
multiple states or a self-loop, sorted by their sorted state IDs, with sorted
internal and directly outgoing transition IDs. No outgoing edge means no direct
structural exit, not proof that execution is trapped. A path or exit is not an
executable witness or guarantee of eventual completion.

Return deterministic, bounded, independently owned typed output. A versioned,
strict bounded stdin wrapper accepts an existing JSON-v1 model. Reject invalid
input without partial output. CLI uses Go 1.27.3, buffered output and fixed
private diagnostics; failed or short writes are errors. No filesystem, network,
subprocess, provider, worker or visual access by this implementation. IDs and
digests may be sensitive and require appropriate handling.

## Acceptance and validation

Retain the prior hand-authored pre-mapping vectors and their provenance. Cover
competing shortest paths, closed and exiting cycles, a self-loop, terminal and
no-path states, set-order invariance, model bounds, invalid and ambiguous JSON,
input nonmutation, output ownership and deterministic CLI behavior. Use
compiling negative controls for tie-breaking and cycle/terminal-path mapping.
Obtain independent exact-source review; distinguish independent checking from
fixture authorship. Run Go 1.27.3 focused and full race tests, vet,
modelc/modeltrace checks, repository validation and whitespace, and full
committed Project Memory validation against actual base `fcda4be`. Record exact
results and limitations. Hosted CI and any later publication remain separate.
