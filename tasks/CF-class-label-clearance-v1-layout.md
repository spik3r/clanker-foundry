---
id: CF-class-label-clearance-v1
status: completed
owner: Codex diagram maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Keep class relation labels clear of class boxes

## Implementation claim

Owner: Codex diagram maintenance contributor, delegated by the coordinator under
the owner's continued maintenance and reviewed-merge request. Base: origin/main
at `9371406`. The coordinator reviewed active work; the output-root symlink task
owns `modeling/diagram/contract/contract.go`, excluded from this change.

A valid three-class model places a short parent beside an unrelated class with
eight attributes. Its child's inheritance relation has the end label "Parent
reference". The finite eight-position search fails, then returns its first
candidate despite overlap: label box `94 116 112 12` intersects class box
`192 68 120 172`. The class is painted over the label. This violates the
profile's label-clearance rule without involving estimated font widths.

Keep model acceptance and public interfaces stable. Preserve existing successful
candidates; add a deterministic finite fallback derived from class extents for
end and centre labels, with canvas containment. Add the literal three-class
regression and endpoint, mirror, maximum-label and tall-class variants using
existing SVG geometry checks. Publish diagramgen 1.2.1 because emitted bytes
change, and regenerate the ten committed examples with updated provenance.

Scope: `modeling/diagram/class/layout.go`, its output and model tests;
`modeling/diagram/diagram.go`, `diagram_test.go`; version-pinned test helpers and
assertions in process, sequence and dashboard; the four profile documents;
`modeling/diagram/README.md`; changed generated SVG, BPMN, HTML and manifest files
under `modeling/diagram/examples/generated/`; this task and immutable Project
Memory files. No other generator, contract I/O, schema, consumer or runtime edits.

Checks: observe the regression fail before the fix; diagram gofmt, vet, full
race suite and authoring schema suite; generator-contract and foundrygen race
suites; repository validator, whitespace and committed Project Memory validation.
Render and visually inspect the repro and generated examples; state any visual
limits. Commit, push and open a PR; independent review and merge belong to the
coordinator.

## Integration result

The coordinating agent reviewed `de02aafc5b783c594fd2bb55ac8025896761984e`;
all ten hosted workflows passed on that head. [PR #148](https://github.com/spik3r/clanker-foundry/pull/148)
merged on 14 November 2026 as `f98b825f971fc253b687c28c660453ea4f0bfc45`.
This bounded implementation claim is complete and reserves no further work.
