# Profile `sequence-diagram/v1`

A UML-style sequence diagram: participants across the top, one lifeline under
each, and steps (messages and notes) in order from top to bottom. `diagramgen`
turns the model into one SVG picture. The model carries meaning only: no
coordinates, sizes, colours or markup. Selected by
`"profileVersion": "sequence-diagram/v1"`; the input format is
`diagram-json/v1`, as for [`process-diagram/v1`](PROFILE.md).

The picture is UML-style. It draws the notation listed under "What is drawn"
and claims no UML conformance beyond that list.

Go validation is authoritative. The
[schema](authoring/sequence-diagram-v1.schema.json) is an authoring aid and is
looser in the places its own description lists.

## Input

```json
{
  "formatVersion": "diagram-json/v1",
  "profileVersion": "sequence-diagram/v1",
  "id": "Checkout",
  "title": "Checkout",
  "participants": [
    {"id": "User", "label": "Customer", "kind": "actor"},
    {"id": "Web", "label": "Web app", "kind": "participant"},
    {"id": "Pay", "label": "Payment service", "kind": "participant"}
  ],
  "steps": [
    {"id": "S1", "kind": "message", "from": "User", "to": "Web", "label": "Place order", "style": "sync"},
    {"id": "S2", "kind": "message", "from": "Web", "to": "Pay", "label": "Charge card", "style": "sync"},
    {"id": "S3", "kind": "message", "from": "Pay", "to": "Web", "label": "Charged", "style": "reply"},
    {"id": "S4", "kind": "note", "over": ["Web"], "label": "Order is now confirmed"}
  ],
  "groups": [
    {"id": "G1", "kind": "opt", "label": "card accepted", "first": "S2", "last": "S3"}
  ]
}
```

- Top level: `formatVersion`, `profileVersion`, `id`, `title`, `participants`,
  `steps` are required; `groups` is optional.
- Participant: `id`, `label`, `kind` (`actor` or `participant`). The array order
  is the left-to-right order.
- Step: `id`, `kind` (`message` or `note`), `label`. The array order is the
  top-to-bottom order.
  - A message also has `from`, `to` (participant ids) and `style`: `sync`,
    `async` or `reply`. It must not have `over`.
  - A note also has `over`: one or two distinct participant ids. It must not
    have `from`, `to` or `style`.
- Group: `id`, `kind` (`alt`, `opt` or `loop`), `label`, `first`, `last` (step
  ids).

## Strict decoding

The same scanner as the other profiles runs first: duplicate keys, unknown
members (matched case-sensitively), trailing data and input over 512 KiB are
refused before any value is read.

## Validation

- `formatVersion` is `diagram-json/v1`; `profileVersion` is
  `sequence-diagram/v1`.
- Ids (model, participants, steps, groups) match `[A-Za-z][A-Za-z0-9_]{0,47}`
  and are unique across participants, steps and groups together.
- `title` and every label are plain text of 1 to 120 UTF-8 bytes, not blank,
  with no control character, U+2028, U+2029, U+FFFE or U+FFFF.
- 2 to 12 participants; 1 to 80 steps; at most 8 groups.
- A message names existing participants in `from` and `to` (they may be equal)
  and has a known `style`. A note names 1 or 2 existing, distinct participants
  in `over`.
- A group names existing steps; `first` is not after `last`; group ranges do not
  overlap each other. Frames are not nested.

## Canonical encoding

Compact JSON, fixed member order as in the input example, no HTML escaping, no
trailing newline. Participant order and step order are significant and kept.
Groups are sorted by id; an empty `groups` is left out. The digest covers these
bytes, so layout code changes never change it, and a change of meaning always
does.

## Output files

| File | Content |
| --- | --- |
| `diagram.svg` | the picture |
| `model.json` | the canonical model, two-space indented, trailing newline |
| `foundry-manifest.json` | integrity record: generator `diagramgen` 1.2.1, toolchain constant `go1.27.4` |

`diagram.svg` is inert: no script, no event attribute, no external reference, no
`foreignObject`, no embedded font or image. Every drawn element group carries
`data-id` and `data-kind` (`participant` or `actor`, `lifeline`, `message`,
`note`, `alt`, `opt`, `loop`), so tests and tools can read geometry back.
All coordinates are integers.

## What is drawn

- Participants: a box for `participant`, a stick figure for `actor`, with the
  label under it (up to three lines in a box, two under a figure), in a row at
  the top.
- A dashed lifeline under each participant, down to the bottom of the diagram.
- `sync` message: solid horizontal line, filled arrowhead.
- `async` message: solid horizontal line, open arrowhead.
- `reply` message: dashed horizontal line, open arrowhead.
- A message to the same participant: a small loop right of the lifeline.
- The message label above the arrow, centred between its ends.
- A note: a box over one or two lifelines with the label wrapped to at most
  three lines.
- A group: a labelled rectangle with a tab holding `alt`, `opt` or `loop` and
  the label in brackets, around the steps `first` to `last`, spanning the
  participants those steps use.

Nothing else is drawn. In particular there are no activation bars, no `else`
dividers, no found or lost messages, no creation or destruction and no message
numbers.

## Layout rules

Fixed algorithm, no randomness, no map-order dependence, integer arithmetic.

- Participants keep model order. Columns are 160 units apart; a gap is widened
  evenly when a message label between those columns needs more room.
- Text width is estimated from a character count (7 units per character), not
  measured from a font.
- A message label is cut at 40 characters (30 for a self-message) and ends with
  `…` where cut. A group label is cut at 40 characters.
- One row per step; a row is taller for a self-message, for a note with more
  lines and where a group starts or ends.
- The picture grows right and down to fit.

## What it cannot express

- Coordinates, sizes, colours or other layout choices.
- Activation bars, creation or destruction of lifelines.
- Found, lost, combined or numbered messages; return values on a message.
- Nested frames, `else` branches, `par`, `ref` and other frame kinds.
- More than 12 participants, 80 steps or 8 groups.
- Notes over more than two participants.

## Limits of the layout

- Because width is estimated, a label of wide characters (for example CJK) can
  run past its arrow, note or frame.
- A long label is cut, not wrapped; the full text stays in the model.
- A message between far-apart participants crosses the lifelines between them.
- A note over two participants covers any lifeline between them.
- A frame spans the steps' own extent, so it can cross lifelines of participants
  that take no part in its steps.

## Limits of what tests show

Tests parse the SVG and check structure and geometry: participant headers on one
baseline, horizontal arrows ending exactly on lifelines, strictly increasing
step rows, notes and frames enclosing what they should. They do not show that a
picture reads well. A person should look at the rendered picture before relying
on it.
