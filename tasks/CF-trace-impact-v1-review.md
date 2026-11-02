---
id: CF-trace-impact-v1
status: in-progress
owner: Codex trace-impact contributor
repositories: [spik3r/clanker-foundry]
---

# Declared trace change review

## Implementation claim

Build a bounded, deterministic read-only comparison of two complete, valid
model/requirements/acceptance-catalog snapshots. This joins existing
`traceability.Build` results to existing `impact.Compare` output, so a reviewer
can find requirement changes and declarations connected to changed model elements
or transitions on either side. It does not replace or duplicate modeldiff's model
comparison, infer dynamic impact, execute scenarios or claim requirement coverage,
authorship, approval or review. A model digest change globally invalidates any
model-bound review regardless of candidate list. The catalog parser reads only
selected fields: any catalog-byte change conservatively flags all requirements
with a declared acceptance link, including formatting-only changes; it must not
pretend to identify changed case behavior or actual test execution.

Exact new scope: `modeling/lifecycle/traceimpact/`,
`modeling/lifecycle/cmd/modeltraceimpact/`, this record and its immutable Project
Memory source/events. Do not alter existing impact, traceability, scenario,
intake/model bridge, process view PR20, authoring, held reviewpack/exporter,
compiler or generated artifacts. Both snapshots must independently pass existing
model and trace validation; dangling links reject instead of being hidden as gaps.
Unchanged gaps remain visible; catalog status remains an unverified declaration.
Inspect current published and held claims again before publication.

## Comparison contract

`Compare(before, after Snapshot) (Report, error)` owns all returned data. Each
snapshot contains raw model, catalog and requirements bytes. Use the existing
parsers, canonical identities, trace report and impact comparison; refuse invalid
or oversized inputs and return no partial report. Report fixed `formatVersion`
`lifecycle-trace-impact/v1`, mode `example-only`, authority `none`, before/after
model digest, normalized requirements digest and raw catalog SHA-256, booleans
`modelChanged`, `requirementsChanged`, `catalogSourceChanged`, and
`reviewRequired` (any identity changed). Before/after gap lists retain all current
trace gaps, even for unaffected requirements. Candidate IDs and reasons are
sorted and deduplicated. Each candidate has id, change (`added`, `removed`,
`modified`, `unchanged`), fields (sorted `title`, `statement`, `modelLinks`,
`acceptanceLinks`, `gaps`), before/after presence and gaps, and reasons.

Reasons are `{code, side, target, cause}`. Codes: `requirement-added` (after),
`requirement-removed` (before), `requirement-changed` (both, target field),
`model-element-changed` (before/after, target `<trace-kind>:<id>`, cause impact
change type), `linked-transition-changed` (before/after, target transition ID,
cause exact existing impact cause), and `catalog-source-changed` (both, empty
target/cause). Direct changed model requirement kinds map to `modelRequirement`.
Propagate changed transitions through explicitly linked transitions only. Preserve
both before and after graph links; changed or removed links must not disappear.
Add catalog-source reason once to any requirement with a declared case link on
at least one side. An empty candidate list does not prove no impact; declarations
can be absent or wrong. IDs are not guessed across renames. An `executable` case
status means only what the catalog declares, never a demonstrated test run.

A bounded stdin-only JSON CLI carries the six exact UTF-8 artifact texts as JSON
strings. It has no paths, network, output files, execution, digest rebinding or
mutable report input. Deterministic output uses explicit typed fields and sorted
lists; no clock or random IDs. Counts, byte caps and error behavior are documented.

## Verification

An independent contributor freezes a minimal before/after model, requirements,
catalog and expected candidates before implementation. Cover direct and propagated
links, removed-side links, changed requirements and gaps, unknown catalog content,
unchanged/missing links, reorder/whitespace identity behavior, malformed input,
resource caps, determinism, nonmutation and CLI errors. Validate with pinned Go
1.27.3 offline, existing Go suite, vet, modelc check, repository and Project Memory
checks, then independent review of the source. One draft PR if permitted; no
merge, deployment, worker activation or bypass of any denied action.
