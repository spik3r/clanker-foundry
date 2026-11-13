# Process diagram profile v1

This profile describes one business-process diagram: lanes, events, tasks,
gateways and sequence flows. Its input format is `diagram-json/v1`, with profile
`process-diagram/v1`. The generator `diagramgen` selects a profile by the
model's `profileVersion`; this is the only profile it supports today.

It is a BPMN-style subset. It claims no BPMN conformance beyond emitting the
elements listed under [BPMN output](#bpmn-output). It does not read BPMN.

## Input

```json
{
  "formatVersion": "diagram-json/v1",
  "profileVersion": "process-diagram/v1",
  "id": "OrderHandling",
  "title": "Order handling",
  "lanes": [{"id": "Sales", "label": "Sales"}],
  "nodes": [{"id": "Start", "kind": "start", "label": "Order received", "lane": "Sales"}],
  "flows": [{"id": "F1", "from": "Start", "to": "Check", "label": "optional"}]
}
```

The model holds meaning only: no coordinates, sizes, colours or markup. A full
example is [examples/order.json](examples/order.json). An authoring aid is the
[JSON Schema](authoring/process-diagram-v1.schema.json); Go validation is
authoritative and enforces rules the schema cannot express (unique ids,
degrees, reachability).

Node `kind` is one of:

| Kind | Meaning | Drawn as |
| --- | --- | --- |
| `start` | where the process begins | thin circle |
| `end` | where a path finishes | thick circle |
| `task` | a unit of work | rounded rectangle |
| `exclusive` | XOR gateway: one outgoing path is taken | diamond with a cross |
| `parallel` | AND gateway: all outgoing paths are taken, or all incoming are awaited | diamond with a plus |

Lane order is the visual order, top to bottom.

## Strict decoding

These are refused: unknown members (a case variant such as `Title` is unknown),
duplicate JSON keys, `null`, an object or array where a scalar belongs, trailing
data after the model, invalid UTF-8, nesting deeper than 16 and input larger
than 512 KiB. A flow `label` that is present must not be empty.

## Validation

- `formatVersion` and `profileVersion` are exactly the values above.
- `id`, and every lane, node and flow `id`, is an ASCII identifier
  `[A-Za-z][A-Za-z0-9_]{0,47}`. Lane, node and flow ids are unique across all
  three kinds together.
- `title` and every label are 1 to 120 UTF-8 bytes of plain text. They cannot
  be blank, contain control characters, U+2028, U+2029, U+FFFE or U+FFFF.
  A flow label is optional.
- 1 to 8 lanes, 2 to 64 nodes, 1 to 128 flows.
- Every node names an existing lane. Every flow names existing nodes.
- Exactly one `start` and at least one `end`.
- `start`: no incoming flow, exactly one outgoing flow.
- `end`: no outgoing flow, at least one incoming flow.
- `task`: at least one incoming flow, exactly one outgoing flow.
- `exclusive` and `parallel`: at least one incoming, at least one outgoing and
  at least three flows in total.
- No flow from a node to itself. No two flows with the same `from` and `to`.
- Every node is reachable from the start, and every node can reach some end.

Cycles are allowed, so a rework loop is valid. The profile does not check
that an `exclusive` gateway's outgoing flows are labelled, that a `parallel`
split has a matching join, or that the process can terminate. Those are design
questions for a reviewer.

## Canonical encoding

`canonicalSHA256` is the SHA-256 of the normalized model as compact JSON: no
whitespace between tokens, no HTML escaping (`<`, `>` and `&` stay literal),
no trailing newline. Members come in this order: `formatVersion`,
`profileVersion`, `id`, `title`, `lanes`, `nodes`, `flows`. A lane is `id`,
`label`. A node is `id`, `kind`, `label`, `lane`. A flow is `id`, `from`, `to`
and, when present, `label`. `nodes` and `flows` are sorted by `id` in byte
order. `lanes` keep their model order.

JSON whitespace, member order in the source, node order and flow order do not
change the digest. Lane order, any label, any id and any kind do. Emitted bytes
depend only on the canonical model and the generator version, not on source
bytes, file names or paths.

## Output files

| File | Content |
| --- | --- |
| `diagram.svg` | the picture |
| `diagram.bpmn` | BPMN 2.0 XML with diagram interchange |
| `model.json` | the canonical model, two-space indented, trailing newline |
| `foundry-manifest.json` | integrity record: generator `diagramgen` 1.2.0, toolchain constant `go1.27.4` |

No emitted file is `.go` or `.ts`, so the generated-file header rule has
nothing to apply to. The manifest is an integrity record, not an approval.

### SVG

An inert picture. It has no script, event attribute, `foreignObject`, image,
embedded font or external reference; the only references are `url(#arrow)` and
the SVG namespace. It has a `<title>` with the model title, a `<desc>` naming
the profile, a fixed `viewBox` computed from the layout, and one small inline
`<style>` element with plain declarations.

Every node is a `<g data-id data-kind>`. Every flow is a `<path data-id
data-from data-to>` ending in an arrowhead marker. Every lane band is a
`<rect data-lane>`. Every flow label is a `<text data-flow data-anchor
data-box>`: `data-anchor` is the index of the path segment the label belongs to
and `data-box` is the label's estimated extent as `x y width height`.

Labels are XML-escaped text, never markup. They wrap by a fixed character
budget and end with `…` when cut: task labels 3 lines of 20 characters, lane
labels 3 of 18, event labels 2 of 22 below the circle, gateway labels 3 of 10
above and left of the diamond, flow labels one line of 12. There is no font
measurement: a flow label's box assumes 7 units per character, and wide
characters can overflow a shape or a box.

### BPMN output

Only these elements are emitted: `definitions`, `collaboration`, `participant`,
`process`, `laneSet`, `lane`, `flowNodeRef`, `startEvent`, `endEvent`, `task`,
`exclusiveGateway`, `parallelGateway`, `incoming`, `outgoing`, `sequenceFlow`,
and the diagram interchange elements `BPMNDiagram`, `BPMNPlane`, `BPMNShape`,
`BPMNEdge`, `Bounds` and `waypoint`. The `participant` is the whole diagram and
gets its own `BPMNShape`, as importers expect for a pool; every lane and node
has one shape and every flow one edge, all taken from the same layout as the
SVG.

Model ids become element ids unchanged. Derived ids contain a dot, which model
ids cannot, so they never collide. `targetNamespace` is the fixed
`urn:foundry:diagramgen:process-diagram:v1`. `exporter` and `exporterVersion`
are the generator id and version. There are no timestamps.

This is a BPMN-style subset. It makes no claim to conform to BPMN 2.0 or to any
BPMN conformance class beyond emitting the elements above. No task types,
events with definitions, message flows, sub-processes, data objects, conditions
or execution semantics are emitted. Whether a particular tool imports the file
has not been checked here.

## Layout

The layout is a fixed algorithm with no randomness and no dependence on map
order. Positions are not part of the model and cannot be set.

1. Back edges are found by a depth-first search from the start that takes
   outgoing flows in flow-id order. A node's column is the longest path from
   the start over the remaining edges.
2. Each lane is a horizontal band. Nodes in the same lane and column stack in
   rows ordered by id; a lane is as tall as its tallest stack.
3. Columns are separated by gutters. A gutter starts with a label zone, which
   holds no vertical line, followed by slots. Each flow owns one slot, ordered
   by flow id, in each gutter its vertical segments may use, so vertical
   segments never coincide. A gutter is as wide as its labels and slots need.
4. A flow leaves a start or a task at the middle of its right side. A flow
   leaves a gateway by direction: at the top corner when its target is in a
   higher row, at the bottom corner when lower, at the right corner when in
   the same row. A back edge leaves towards its channel. Each side has three
   exits, the corner and one point on the diamond either side of it, and
   every flow gets its own, so flows from one gateway do not run along each
   other. When a side is full the flow takes a free exit on another side.
   A flow leaving the top or bottom runs to its own corridor line above or
   below the gateway and turns right there.
5. Every flow enters the middle of its target's left side. Flows into one
   target share the last stretch.
6. A forward flow first tries a route with one vertical segment next to the
   source and, when it leaves rightwards, one next to the target. If a route
   would cross a node, and for every back edge, the flow goes through a channel
   above the first lane or below the last lane, chosen by shorter total travel,
   with one distinct line per flow.
7. A flow label sits just above the flow's first horizontal run, in the label
   zone right after the source. That run is cut at the end of the label zone,
   so the labelled piece is a segment of its own that no other flow shares.
   All horizontal lines inside a row lie on fixed levels at least 18 units
   apart, and a label is 10 units high, so it fits between two levels.

Guarantees, tested on the examples, on several hundred generated models and on
gateways with up to 30 outgoing flows: no two node shapes overlap; every node
lies inside its lane band; flow segments are axis-aligned; a flow starts on its
source and ends on its target; no flow segment passes through the interior of
any node (a gateway exit beside a corner starts on the diamond, inside the
gateway's own bounding box); vertical segments and channel lines of different
flows never coincide; every labelled flow has exactly one label, anchored to a
segment no other flow shares; and no label box overlaps a node box, another
label box or any flow segment.

Limits. A gateway has nine exits. With ten or more outgoing flows, the first
eight get their own exit and the rest share the ninth, so those flows do run
along each other next to the gateway; each of them carries its label to the
right of its own vertical segment instead, in a slot widened for it. Flows may
cross each other. Rows are 176 units tall to keep the levels apart, so diagrams
are tall as well as wide.
