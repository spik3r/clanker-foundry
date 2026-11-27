# Logical architecture profile v1

Copy the [architecture starter](../../templates/diagrams/architecture.json).
Use this view to answer “who depends on what?” for one small system. A node is
an `actor`, `service`, `store`, or `external` dependency. A link points from
the caller to its dependency; its label says what interaction takes place.
Cycles and distinct interactions between the same pair are allowed. For
message ordering use the [sequence profile](SEQUENCE.md); for process
ownership and branching use the [BPMN-style process profile](PROFILE.md).

## Input and output

- `formatVersion` is `diagram-json/v1`; `profileVersion` is
  `architecture-diagram/v1`. The root has an `id`, `title`, `nodes`, and
  `links`, all required. No unknown or duplicate members, nulls, or invalid
  Unicode are accepted.
- Provide 1–12 nodes and 0–24 links. Each node has `id`, `kind`, `label`;
  each link has `id`, `from`, `to`, `label`. IDs match
  `[A-Za-z][A-Za-z0-9_]{0,47}`. Node/link IDs share one namespace; endpoints
  must be two different existing nodes. Plain text is limited to 120 UTF-8
  bytes for the title and 96 for each label.
- The [authoring schema](authoring/architecture-diagram-v1.schema.json) helps
  editors; Go validation decides byte length, uniqueness and references.
- Output is `diagram.svg`, canonical `model.json` and
  `foundry-manifest.json`. Canonical data keeps node order (top to bottom) and
  sorts links by ID. Equivalent formatting and link order produce identical
  output bytes. The CLI input identity separately reports the source digest.

The fixed layout uses one row per node and one numbered routing track per link.
Arrows end at the dependency; a numbered key shows the interaction and its
endpoints. Tracks may cross but do not run through boxes. Text wraps by
character count, not font measurement, and can be shortened visually; full
labels remain in the model. There are no coordinates, custom styles, nested
boundaries, network/deployment views or C4 conformance claims.
