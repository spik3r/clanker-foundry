# Profile `class-diagram/v1`

A UML-style class diagram: classes, interfaces and enumerations with attributes
and operations, joined by seven kinds of relation. `diagramgen` turns the model
into one SVG picture. The model carries meaning only: no coordinates, sizes,
colours or markup. Selected by `"profileVersion": "class-diagram/v1"`; the input
format is `diagram-json/v1`, as for [`process-diagram/v1`](PROFILE.md).

The picture is UML-style. It draws the notation listed under "What is drawn"
and claims no UML conformance beyond that list.

Go validation is authoritative. The
[schema](authoring/class-diagram-v1.schema.json) is an authoring aid and is
looser in the places its own description lists.

## Input

```json
{
  "formatVersion": "diagram-json/v1",
  "profileVersion": "class-diagram/v1",
  "id": "Orders",
  "title": "Order model",
  "classes": [
    {"id": "Order", "name": "Order", "kind": "class",
     "attributes": [{"name": "id", "type": "OrderId", "visibility": "public"}],
     "operations": [{"name": "total", "returns": "Money", "visibility": "public"}]},
    {"id": "Line", "name": "OrderLine", "kind": "class", "attributes": [], "operations": []},
    {"id": "Priced", "name": "Priced", "kind": "interface", "attributes": [],
     "operations": [{"name": "price", "returns": "Money", "visibility": "public"}]}
  ],
  "relations": [
    {"id": "R1", "kind": "composition", "from": "Order", "to": "Line", "fromLabel": "1", "toLabel": "1..*"},
    {"id": "R2", "kind": "realization", "from": "Line", "to": "Priced"}
  ]
}
```

- Top level: `formatVersion`, `profileVersion`, `id`, `title`, `classes` are
  required; `relations` is optional.
- Class: `id`, `name`, `kind` (`class`, `interface` or `enum`); `attributes` and
  `operations` are optional lists. For an `enum` the attributes are the
  literals.
- Attribute: `name`; optional `type` and `visibility`.
- Operation: `name`; optional `parameters` (each `{name, type}`), `returns` and
  `visibility`.
- `visibility` is `public`, `private`, `protected` or `package`; absent means
  `public`.
- Relation: `id`, `kind`, `from`, `to`; optional `label`, `fromLabel`,
  `toLabel`.

## Strict decoding

The same scanner as the other profiles runs first: duplicate keys, unknown
members (matched case-sensitively), trailing data and input over 512 KiB are
refused before any value is read. An optional text member that is present must
not be empty.

## Validation

- `formatVersion` is `diagram-json/v1`; `profileVersion` is
  `class-diagram/v1`.
- Ids match `[A-Za-z][A-Za-z0-9_]{0,47}` and are unique across classes and
  relations together.
- `title`, class names, member names, types and parameter names are plain text
  of 1 to 120 UTF-8 bytes, not blank, with no control character, U+2028,
  U+2029, U+FFFE or U+FFFF.
- Relation `label`, `fromLabel` and `toLabel` follow the same rules with 1 to 24
  bytes.
- 1 to 40 classes; at most 20 attributes and 20 operations per class; at most 8
  parameters per operation; at most 80 relations.
- Attribute names are unique within a class; operation names are unique within
  a class; parameter names are unique within an operation.
- A relation names two existing, different classes. No two relations have the
  same unordered pair of classes and the same kind.
- Inheritance and realization together form no cycle.

## Canonical encoding

Compact JSON, fixed member order as in the input example, no HTML escaping, no
trailing newline. Classes are sorted by id and relations by id. Attribute,
operation and parameter order are significant and kept. Absent lists are
written as empty ones (`parameters` is left out when empty), and an absent
visibility is written as `public`. The digest covers these bytes.

## Output files

| File | Content |
| --- | --- |
| `diagram.svg` | the picture |
| `model.json` | the canonical model, two-space indented, trailing newline |
| `foundry-manifest.json` | integrity record: generator `diagramgen` 1.2.1, toolchain constant `go1.27.4` |

`diagram.svg` is inert: no script, no event attribute, no external reference, no
`foreignObject`, no embedded font or image. Every drawn element group carries
`data-id` and `data-kind` (`class`, `interface`, `enum`, or the relation kind),
and each box and relation carries its geometry (`data-box`, `data-points`), so
tests can read it back. All coordinates are integers.

## What is drawn

- A class box in three compartments (name, attributes, operations) with
  separator lines. Empty compartments are drawn thin.
- `«interface»` above the name of an interface; `«enumeration»` above the name
  of an enum. The name is bold.
- Attributes as `+ name: Type`; operations as `+ name(a: T, b: U): R`. The
  symbol is `+` public, `-` private, `#` protected, `~` package. Enum literals
  are drawn as their name only: type and visibility are accepted but not drawn.
- Relations as orthogonal polylines between box borders:

| Kind | Line | End shape |
| --- | --- | --- |
| `inheritance` | solid | hollow triangle at `to` (the parent) |
| `realization` | dashed | hollow triangle at `to` |
| `association` | solid | none |
| `directed` | solid | open arrow at `to` |
| `aggregation` | solid | hollow diamond at `from` |
| `composition` | solid | filled diamond at `from` |
| `dependency` | dashed | open arrow at `to` |

- The relation `label` beside a segment that no other relation uses; `fromLabel`
  and `toLabel` beside their ends. Labels stay clear of every class box. When
  the nearby positions are blocked, placement uses class edges to find a free
  baseline, keeping end labels on the outward side of their class. The canvas
  grows to contain the labels.

Nothing else is drawn: no abstract or static marks, no generics, no packages,
no notes, no association classes, no navigability marks beyond `directed`, no
constraints.

## Layout rules

Fixed algorithm, no randomness, no map-order dependence, integer arithmetic.

- Rank: longest path over inheritance and realization, so a parent or interface
  is always above its child. Other kinds do not force rank.
- Order within a rank: four barycenter sweeps over all relations, ties keep id
  order. A rank wider than 1400 units is split into several rows.
- Box width comes from the longest text line by character count (7 units per
  character, 8 for the bold name), between 120 and 300 units, and wide enough
  for one port per touching relation. Longer lines are cut and end with `…`.
  Height comes from the number of lines.
- A relation leaves the bottom of the upper class and enters the top of the lower
  one. Each touching relation gets its own port on the border so parallel edges
  do not coincide. A relation between rows with other rows in between runs in a
  vertical corridor right of all boxes. A relation inside one row arches over it.
  Routes stay in the free bands between rows and in the corridors, so they do not
  cross a box other than their own ends.

## What it cannot express

- Coordinates, sizes, colours or other layout choices.
- Packages, generics, abstract or static members, association classes, notes,
  constraints, qualified associations, derived members.
- Self-relations, two relations of one kind between the same pair of classes,
  inheritance or realization cycles.
- More than 40 classes, 20 members per list, 8 parameters or 80 relations.
- Relation ends on a particular side of a class.

## Limits of the layout

- Width is estimated from character counts, not measured; wide characters can
  overflow a box. A long member line is cut, not wrapped.
- Lines cross when relations cross between ranks; the ordering sweeps reduce but
  do not minimise crossings.
- Relations that only share a rank may still be drawn long, around the row.
- A diagram with many relations is tall, because each channel between rows grows
  with the number of routes in it.
- Labels sit beside segments by estimated extent and can touch another line.

## Limits of what tests show

Tests parse the SVG and check structure and geometry: no two boxes overlap,
parents above children, relation ends on box borders, axis-aligned segments,
no segment through another box's interior, labels clear of boxes. They do not
show that a picture reads well. A person should look at the rendered picture
before relying on it.
