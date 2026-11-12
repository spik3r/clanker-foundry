# Direction: models in, generated code out

Reviewed against `origin/main` at `3d33f68` on 12 November 2026. The owner
accepted this plan the same day; the [decisions](#owner-decisions) are recorded
at the end. Each step below still needs its own reviewed task and claim. The
plan answers the route question in the [current direction](current-direction.md)
and reorders the priorities in the [holistic plan](holistic-improvement-plan.md).

## The aim

The owner's stated aim: a deterministic pipeline in which **AI drafts models and
reviewed generators produce the code. AI does not write the feature code.**

Stated as a rule for the pipeline:

> On the generated path, AI output is data: a model, expectation literals, or a
> gap report. Every byte of feature code is a function of the accepted model,
> the pinned generator revision and the pinned toolchain.

Three consequences follow, and the plan depends on each.

1. **AI still writes code, but in one place: the generators.** A generator or
   template change is ordinary engineering with ordinary review. It is reviewed
   once and reused for every model. The per-feature step contains no AI-written
   code.
2. **The pipeline is only as useful as its profiles.** A requirement no profile
   can express has three honest outcomes: extend a profile, write the code by
   hand in a marked seam, or leave the feature out of the pipeline. An AI-written
   candidate is not one of them.
3. **Determinism gives reproducible output, not correct output.** A wrong model
   regenerates perfectly. Independent expectations and human review of model
   meaning stay in place. The [pipeline guide](model-driven-pipeline.md) already
   says this and it does not change.

## What main has now

Facts observed in the tree at `3d33f68`.

| Generator | Model in | Code out | Size (non-test lines) |
| --- | --- | --- | --- |
| [lifecycle](../modeling/lifecycle/README.md) | `lifecycle/v1`: states, guards, transitions, effects over one fixed task/command shape | Go reducer, structural tests, SVG, decision table, trace | compiler 269, IR 343; whole module about 16,800 |
| [jsonvalidator](../modeling/jsonvalidator/README.md) | `json-boolean/v1`: bounded JSON shape | Go and TypeScript boolean predicates | about 530 |
| [tui](../modeling/tui/README.md) | `record-browser/v1`: fields, labels, breakpoints | Go package for Flexgo and Bubble Tea, demo, trace | about 340 plus nine templates |

- **Three generators, three contracts.** Each is a separate Go module with its
  own strict decoder, CLI shape, output rules and provenance record. `modelc`
  reads one fixed path and writes one fixed directory. `jsonvalidator` takes
  `-model` and `-out` and embeds a hash of raw model bytes with no manifest.
  `tuic` takes positional arguments and writes a `trace.json` over all outputs.
  Lifecycle has a canonical model digest; the JSON generator does not.
- **Most lifecycle code is not the generator.** The module has 33 commands. The
  compiler and IR are about 600 lines. The rest prepares, compares, rehearses
  and reports on one fictional task model.
- **AI authoring support covers one profile.** The
  [model-driven-design skill](../skills/model-driven-design/SKILL.md) and the
  only authoring schema target `lifecycle/v1`. The other two profiles have a
  `PROFILE.md` and Go validation, but no schema and no skill handoff.
- **Expectations are data in two of three.** Lifecycle suites and the JSON
  validator cases are literal JSON. TUI acceptance is a Go test template.
- **The profiles do not compose.** The lifecycle reducer accepts typed Go
  values, not JSON. The validator checks parsed JSON and returns a boolean. The
  browser declares its own string fields. No model refers to another.
- **Two implementation routes are documented as equals.** The pipeline guide
  describes deterministic generation and an agent-written
  [builder route](../examples/model-builder-go/README.md).
- **No consumer runs a generator.** No repository pins a generator revision,
  regenerates in CI or fails on drift. The TUI compiler targets the owner's
  public Flexgo library, which is the nearest thing to a real target.
- **The [runtime](../runtime/README.md) is separate and inactive.** It does not
  run modeling steps.
- Open pull requests: draft
  [spik3r/clanker-foundry#101](https://github.com/spik3r/clanker-foundry/pull/101)
  (JSON validator requirement-change regression) and draft
  [spik3r/clanker-foundry#110](https://github.com/spik3r/clanker-foundry/pull/110)
  (runtime CLI input refusal). Neither conflicts with this plan.

## Assessment

Opinion, for the owner to accept or reject.

The parts needed for the aim exist, and the newest one shows the right shape.
The TUI compiler is 340 lines, takes a model, emits a compilable package, and its
README demonstrates a requirement change with no hand edit. That is the aim in
miniature.

What is missing is everything that makes three generators into one pipeline:
a common contract, an authoring contract an AI can follow for any profile, one
command that regenerates and detects drift, and a consumer that depends on it.

The earlier plans block all work on the owner choosing a consumer and a route.
The owner has now chosen the route. That removes one of the two blocking
decisions and makes several items concrete that were previously deferred as
speculative. The consumer decision still stands, and step 4 needs it.

Effort has gone mostly into review and rehearsal reports around the lifecycle
example. Under this aim those reports are secondary. The generator contract and
profile coverage are primary.

## Direction changes proposed

1. **Generated code is the pipeline's only implementation route.** Keep the
   builder exercise as dated evidence. Stop describing it as a pipeline route
   and add nothing to it.
2. **Profiles stay declarative.** No loops, user-defined functions or embedded
   source in any model. If a profile needs general computation, the answer is a
   typed seam where a person writes Go, not a larger expression language. A
   model language that grows into a programming language returns the pipeline
   to AI-written code in a worse language.
3. **Unsupported means refused, in writing.** When a requirement does not fit,
   the AI's output is a gap report naming the requirement and the missing
   construct. Gaps drive profile work. They are not filled with prose or code.
4. **Pause new lifecycle report commands and runtime features.** Fixes continue.
   This matches the existing direction.

## Next steps in dependency order

### 1. Generator contract v1

One written contract and one conformance suite that all three generators pass.

- Same verbs and arguments: `validate MODEL`, `generate MODEL OUT`,
  `check MODEL OUT`. Explicit model path, fresh or manifest-owned output only.
- No network, no execution of generated code, no reading outside the model.
- Canonical model digest for every profile, separate from the raw-byte hash.
- One manifest schema binding model digest, profile, generator identity and
  revision, toolchain and a hash of every emitted file.
- A generated-file header naming the model and generator, so a hand edit is
  detectable and a reader knows not to make one.
- Shared refusal classes and exit codes.

The contract is checked from outside. A separate conformance module builds each
generator and drives its command line; generators do not import shared code, so
each keeps its own dependency-free module.

**Parts, one pull request each:**

1. The written contract, the conformance harness, a small reference generator
   that the harness tests itself against, and a CI workflow.
2. `tuic` conforms and joins the harness registry.
3. `jsonvalidator` conforms and joins the registry.
4. `modelc` conforms and joins the registry. Its fixed-path commands keep
   working, because the other lifecycle commands depend on them.

**Work:** `modelc` accepts a model path and output directory. `jsonvalidator`
gains a canonical digest, a manifest and `check`. `tuic` gains `validate` and
`check`. Extract the strict JSON decoder and manifest writer only if the
conformance suite shows the copies disagree.

**Done when:** one test suite runs each generator twice on its sample and gets
identical bytes, verifies the manifest, and confirms `check` fails after a
one-byte edit to any output. CI runs it for all three.

### 2. Authoring contract for every profile

What an AI needs to draft a valid model without reading generator source.

- A registry file listing each profile, its `PROFILE.md`, schema, example model
  and generator command.
- A JSON Schema for `json-boolean/v1` and `record-browser/v1`, tested against
  the Go validator the way the lifecycle schema is. Go stays authoritative.
- The model-driven-design skill reads the registry instead of assuming
  lifecycle. Its fit step produces either a model or a gap report.
- Expectation literals as data for the browser profile: key sequence and
  terminal size in, expected focus, selection and visible text out. One fixed
  runner per profile executes them. AI writes cases, never test code.

**Done when:** from a written brief and the registry alone, a fresh agent
session produces a model and literal cases for each profile that pass
`validate`, or a gap report. Run this as a recorded trial per profile and keep
the failures.

### 3. One entrypoint and a drift check

- A project manifest in the consuming repository lists each model, its profile,
  its output directory and the pinned generator revision.
- `foundry gen` regenerates everything listed. `foundry gen --check`
  regenerates to a temporary directory and compares bytes.
- A reusable CI job runs the check, so a hand edit or stale output fails the
  build.

**Done when:** in a scratch repository, editing a generated file fails CI, and
editing a model without regenerating fails CI.

### 4. First consumer, end to end

The owner named `~/code/monstera`. That checkout is on the owner's machine and
is not published where this plan's author can read it, so the feature is not
chosen yet. Picking it needs read access to the repository. The selection rule:
prefer a feature one existing profile covers whole, with the surrounding code
as an explicit hand-written seam, and with a requirement change already wanted.

**Done when:** the consumer builds from a pinned generator revision with the
drift check in its CI; one requirement change lands by editing only the model
and its expectation literals; and the record lists every line written by hand
and every gap report raised. The acceptance and stop rules in the
[current direction](current-direction.md) milestones 1 and 2 apply unchanged.

### 4b. Second consumer: diagram and dashboard generators

Requested by the owner. An AI drafts a declarative model; a generator emits an
inert picture or page. This does not depend on step 4 and can start once the
step 1 contract exists, so the new generators conform from their first commit.

Proposed profiles, each a finite documented subset and a separate claim:

- `process-diagram/v1`: start and end events, tasks, exclusive and parallel
  gateways, sequence flows and lanes. Output is SVG, and optionally BPMN 2.0
  XML for import into other tools.
- `class-diagram/v1` and `sequence-diagram/v1`: classes with attributes,
  operations and typed relations; participants with ordered messages.
- `dashboard/v1`: stat tiles, tables and bar or line charts over data supplied
  as JSON. Output is one self-contained HTML file with no script and no network
  reference.

Rules that follow from the aim. Layout is a fixed algorithm in the generator,
so the same model gives the same bytes; models carry no coordinates unless a
profile says so. Labels are escaped data, never markup. Each profile states
which BPMN or UML elements it supports and claims conformance to nothing
beyond that list. The existing lifecycle [process view](../modeling/lifecycle/processview/README.md)
stays as it is; a later change may re-express it as a `process-diagram/v1`
model.

**Done when:** each profile has a `PROFILE.md`, schema, example model,
expectation literals over the model's structure (elements present, edges
connected, no overlap), passes the step 1 conformance suite, and a person has
looked at the rendered examples.

### 5. Grow profiles from recorded gaps

Start only after step 4 produces gaps. Two extensions are likely enough to plan
for, but neither should start before a consumer asks.

- **A shared data-shape profile.** One model declares a record's fields and
  types. Lifecycle, validator and browser models refer to it. Today lifecycle is
  fixed to one task shape and the browser accepts strings only. This is the
  change that lets three generators describe one feature.
- **Composition at the boundaries.** A generated validator in front of a
  generated reducer, and browser actions bound to lifecycle triggers. Each
  binding is declared in a model and checked by the generator.

**Done when:** each extension closes a named gap from a consumer record and
passes the step 1 conformance suite.

### 6. Run the pipeline through the runtime, later

Draft model (AI), validate, generate, run literals (all deterministic), review
(person) map onto typed runtime tasks. Do this only after steps 1 to 4, and
only under the [rollout guide](deterministic-orchestration-rollout.md). Until
then the pipeline is a CLI and a CI job.

## What to measure

- Byte-identical regeneration in CI for every generator (step 1).
- Share of feature lines generated versus hand-written in the consumer (step 4).
- Gap reports per requirement, and how each was resolved (steps 2, 4, 5).
- Elapsed time and review findings for one requirement change, against the
  owner's baseline from the current direction's question 4.

## Risks

- **Profile creep.** Covered by direction change 2. Review every profile
  extension for whether it adds computation.
- **Generator bugs repeat everywhere.** Independent literals catch behavior;
  the conformance suite catches contract breaks. Neither replaces review of
  generator changes.
- **Generated dependency pins age.** The TUI module has already needed an
  `x/sys` patch. A generator release becomes the update channel, and consumers
  need the drift check to pick it up.
- **Too little is expressible to be worth it.** Step 4 measures this. If the
  generated share is small and gaps are many, narrow the aim to the profiles
  that pay rather than widening the language.

## Owner decisions

Recorded on 12 November 2026.

1. The rule in "The aim" and direction changes 1 to 4 are accepted.
2. The first consumer is `~/code/monstera`. The feature is open; see step 4.
3. Steps 1 to 3 start now, before the feature is chosen.
4. This plan answers the route question in the current direction, which now
   carries a dated note saying so.
5. Diagram and dashboard generation is the next consumer; see step 4b.
