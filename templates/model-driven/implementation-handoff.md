# Implementation packet: <task and bounded slice>

Status: <proposal / reviewed for the named scope / blocked, with source>.
This document records supplied facts and decisions; it does not grant authority.
Replace every placeholder before handoff. Do not fabricate absent evidence.

## Assignment and authority

- Canonical task, pinned historical claim commit/blob and separate current live claim source:
- Target repository, exact base revision and exclusive writable workspace:
- Existing changes to preserve; assigned builder and coordinating owner:
- Exact writable files; explicit exclusions:
- Route: **agent-written code** or **generated-code integration**. Why this route:
- Edit permission: source, exact scope, status:
- Compile/build permission: source, exact tool and commands, cache/temp/output paths, status:
- Candidate execution permission: source, named code/commands/host, status:
- Integration, publication and deployment: separate sources/statuses or outside scope:
- Required decisions/reviews before each permitted action:

Verify the current claim and trusted authorization independently of this packet.
If an experiment is allowed before real-use decisions, state the narrow experimental
scope and which gates still block real use. Do not turn that exception into approval.

## Read and discovery boundary

Carry the initial boundary in the trusted receiving prompt, before its first
filesystem action. Do not rely on this packet being read before routine discovery.

- First readable files: exact absolute packet and governing instruction paths:
- Subsequent semantic inputs: exact paths from the inventory below:
- Process guidance and historical/live claim files or pinned Git blobs:
- Permitted metadata operations, each with its exact repository/workspace root:
- Directory enumeration or search: none unless an exact subtree and purpose are granted:
- Missing-input or instruction-scope conflict: responsible owner and required resolution:

A named repository, current directory or writable workspace is not blanket read
permission. Filename listings, directory enumeration, glob expansion and recursive
instruction discovery count as reads. Do not search shared ancestors or siblings
or follow an unlisted source link to replace an absent input. A permitted metadata
command may reveal filenames without permitting their contents to be read.
Higher-priority instructions still apply; surface a conflict instead of widening
scope or ignoring them. Record actual exposure, not an assumed isolation guarantee.

## Pinned input inventory

| Role | Accessible location and source revision | Exact bytes SHA-256 | Semantic/profile identity if applicable | Actual status |
|---|---|---|---|---|
| Original requirements/brief | | | | |
| Source model/blueprint | | | | |
| Interface/data/error contract | | | | |
| Derived diagram/table | | | | |
| Independent literal expectations | | | | |
| Acceptance harness/adapters | | | | |
| Generated package, only for selected integration route | | | | |

Include needed original files, not only a summary or screenshot. State absent
artifacts explicitly. Distinguish source-byte and canonical identities. Hashes
bind supplied bytes; they do not authenticate author, execution or approval.
Read historical inputs from their pinned revision/blob. Check current task ownership
separately; a changed live task status is not drift in the historical source blob.
For changed or inaccessible input, stop affected work and ask for resolution.

## Meaning and scope

- Intended outcome and actors:
- Chosen modeling language/profile/version and supported semantic subset:
- Concrete behavior inside the model:
- Surrounding code explicitly included in this assignment:
- Unsupported/excluded behavior and follow-on owner:
- Views: relationship to source model, actual inspection and pending visual gates:
- Existing approvals/reviews with actual source and scope, not inferred from checks:

| Decision or question | Supplied answer and source, or still open | Affected slice and next owner |
|---|---|---|

## Trace and concrete code contract

| Requirement | Model element(s) | Implemented component/interface | Acceptance case(s) | Gap or limitation |
|---|---|---|---|---|

- Interface signatures, types, permitted input domain and units/encoding:
- Outputs, error/refusal representation and ordering/precedence:
- Mutation, copy/aliasing, state, persistence and external-effect rules:
- Compatibility, security and operational constraints:
- Source precedence: disagreements require a decision; do not silently choose a winner:
- Generated route only: trusted generator/toolchain, actual package binding and owned frontier:

## Plan and frozen checks

1. <Verify identities/claim and clarify any blocking choice.>
2. <Implement a named small slice within the write scope.>
3. <Run the permitted checks and preserve their results.>

| Check | Exact command and working directory | Does it execute candidate code? | Expected criterion and evidence location |
|---|---|---|---|

Expectation author/reviewer, materials actually seen, excluded implementation/run
outputs, freeze commit/time/identities before builder implementation and remaining independence limits:

Distinguish build-only commands from behavior execution. For Go, `go test -c`
compiles; `go test -run '^$'` still runs initializers. Name all output/cache/temp
paths and whether each write and run is allowed.

Preserve literals and input files. New implementation tests may supplement them.
For a discrepancy, record full expected/actual values and return the unresolved
choice. Do not regenerate expectations from output or repair a stale binding.

## Builder return

- Exact candidate revision/files and hashes when useful:
- Requirement/model/interface links for each implemented slice:
- Commands run, exits, full comparison reports and relevant environment:
- Input files actually read, discovery/metadata operations and unavailable inputs:
- Checks failed, blocked or not run; scope of actual validation:
- Remaining unsupported behavior, assumptions and decisions with owners:
- Next action and required authority:

An independent reviewer must be able to reproduce the named checks from the
retained sources when execution is authorized. Local evidence is not a release,
an approval or a promise about future agents. Keep private inputs and reports private.
