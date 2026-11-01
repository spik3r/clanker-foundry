---
id: CF-lifecycle-process-view-v1
status: in-progress
owner: Codex lifecycle process-view contributor
repositories: [spik3r/clanker-foundry]
---

# Lifecycle process-oriented review view

## Implementation claim

Build a separate offline, deterministic human-review projection of the existing
validated lifecycle model on branch `codex/lifecycle-process-view`. Existing
`views/` renders transition-per-row state connections and a decision table; this
view instead lays out each declared state/trigger as an explicit admission flow:
all guards, guard refusal, revision-limit refusal, ordered effects and destination.
The projection is a review aid, not executable process semantics, proof of
execution or approval, a scheduler, or BPMN conformance. Explicitly label invalid
input/illegal transitions as outside the per-transition lanes and unsupported
concepts such as parallelism, timers, pools and subprocesses. Recompute canonical
model identity from validated input; output must not depend on source ordering where
normalization makes order immaterial. Never modify compiler-generated artifacts.

Scope is new `modeling/lifecycle/processview/` and
`modeling/lifecycle/cmd/modelprocess/` files, this task and its own new immutable
Project Memory records only. Do not edit existing model/compiler/views/trace,
traceability, tutorial, scenario, reviewpack/exporter or Astra's `intakemodel/`
bridge paths. Main, visible claims and the held stack were inspected; the local
claim is not a shared lock. Recheck before publication.

## Projection contract and verification

`processview.Build(raw []byte) (Outputs, error)` accepts the existing strict
JSON-v1 model, normalizes it, computes its canonical identity and produces JSON
and inert SVG bytes. JSON has `formatVersion`, `mode`, `authority`, `modelDigest`,
`initial`, `states` (id, label, terminal), `transitions` (id, from, trigger, to,
normalized guards and ordered effect op names), and `guards` (id, requirement,
readable expression). Version is `lifecycle-process-view/v1`; mode is
`example-only`, authority is `none`. No source-byte hash in the projection, so
source formatting and semantic-set order do not change bytes. Effects retain their
meaningful order. Bound input and output, escape all untrusted text, never silently
truncate, and return no partial artifacts on failure. The CLI consumes bounded
stdin and emits either JSON or SVG; it never opens a path or writes files itself.

After genuine start, an independent contributor freezes expected transition,
guard and refusal cases before implementation. Test semantics against those fixed
literals, validate SVG structure and escaping, repeatability and harmless reorder,
malformed and resource-bounded inputs, CLI failure/short output, and actual
example model. Run the pinned Go race/vet/drift and repository checks, obtain
independent review, record final evidence, and aim for one draft PR. No merge,
deployment, worker activation or permission expansion.
