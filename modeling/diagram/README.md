# Diagram generator

`diagramgen` turns a reviewed JSON model into a picture. A person or an AI drafts
the model; the generator draws it the same way every time. It follows the
[Foundry generator contract](../CONTRACT.md) and is listed in the
[conformance registry](../gencontract/generators.json).

Four profiles exist. The generator picks one by the model's `profileVersion`:

- [`process-diagram/v1`](PROFILE.md): a BPMN-style process diagram with lanes,
  tasks, events and gateways, as SVG and BPMN XML.
- [`dashboard/v1`](DASHBOARD.md): stat tiles, bar and line charts and tables
  over small embedded datasets, as one self-contained HTML file.
- [`sequence-diagram/v1`](SEQUENCE.md): a UML-style sequence diagram of
  participants, messages, notes and alt, opt and loop frames, as SVG.
- [`class-diagram/v1`](CLASS.md): a UML-style class diagram of classes,
  interfaces, enums and seven kinds of relation, as SVG.

Other profiles can be added as sibling packages that implement
`contract.Profile`.

For recurring reports, [reuse a reviewed dashboard with fresh data](REPORT-BINDING.md)
using `diagramgen bind TEMPLATE DATA` before generation.

## Run it

Go 1.27.4 is required.

```sh
cd modeling/diagram
go run ./cmd/diagramgen validate examples/order.json
go run ./cmd/diagramgen generate examples/order.json /tmp/order-diagram
go run ./cmd/diagramgen check examples/order.json /tmp/order-diagram
go test -race -count=1 ./...
```

`generate` needs an `OUT` that does not exist yet. `check` reports drift and
changes nothing. Exit codes are 0 success, 1 failure, 2 refused, 3 drift.

## What it writes

- `diagram.svg`: the picture. Inert: no script, no external reference.
- `diagram.bpmn`: BPMN 2.0 XML, with positions, for import into other tools.
- `model.json`: the canonical model.
- `foundry-manifest.json`: file hashes and the generator identity.

For a `dashboard/v1` model the files are `dashboard.html`, `model.json` and the
manifest; see the dashboard section below.

Generated examples sit under [examples/generated](examples/generated), one
directory per model: [order](examples/order.json), [parallel](examples/parallel.json)
and [rework](examples/rework.json) (a loop back to an earlier step), plus the
dashboard models [ops-dashboard](examples/ops-dashboard.json),
[sales-dashboard](examples/sales-dashboard.json) and
[minimal-dashboard](examples/minimal-dashboard.json). A test regenerates them
and compares bytes.

## Dashboards

`dashboard/v1` models hold datasets and widgets: `stat`, `bar`, `line` and
`table`, arranged in rows of up to four. The output is `dashboard.html`: inline
CSS, inline SVG charts, no script, no external reference, light and dark themes,
one column on narrow screens. Chart axes and coordinates are computed with
integer arithmetic, so output is byte-stable. Try it:

```sh
go run ./cmd/diagramgen validate examples/ops-dashboard.json
go run ./cmd/diagramgen generate examples/ops-dashboard.json /tmp/ops-dashboard
```

Input, number rules, canonical encoding, chart rules and what a dashboard
cannot express are in [DASHBOARD.md](DASHBOARD.md). The
[schema](authoring/dashboard-v1.schema.json) helps an editor; Go validation
decides.

## UML-style diagrams

`sequence-diagram/v1` and `class-diagram/v1` write `diagram.svg`, `model.json`
and the manifest. The pictures are UML-style: each profile document lists the
notation it draws and claims no UML conformance beyond that list. Examples:
[checkout-sequence](examples/checkout-sequence.json),
[retry-sequence](examples/retry-sequence.json),
[orders-class](examples/orders-class.json) and
[plugins-class](examples/plugins-class.json).

```sh
go run ./cmd/diagramgen generate examples/orders-class.json /tmp/orders-class
```

Input, validation, canonical encoding, what is drawn, what it cannot express and
known layout limits are in [SEQUENCE.md](SEQUENCE.md) and [CLASS.md](CLASS.md).
The schemas ([sequence](authoring/sequence-diagram-v1.schema.json),
[class](authoring/class-diagram-v1.schema.json)) help an editor; Go validation
decides.

## Change a diagram

Edit the model, then regenerate into a new directory and replace the old output.
Never edit `diagram.svg`, `diagram.bpmn`, `dashboard.html` or `model.json` by hand; `check` reports
the edit as drift. The [schema](authoring/process-diagram-v1.schema.json) helps
an editor, but the Go validation decides.

## Limits

- Layout is a fixed, simple algorithm. It is correct and stable, not elegant:
  diagrams are wide and tall, and lines cross. There is no manual positioning;
  to change the picture, change the model.
- A flow leaves a gateway at the top, bottom or right depending on where its
  target is, and its label sits on the first stretch that belongs to that flow
  alone. A gateway has nine exits; flows beyond that share one.
- Label wrapping counts characters. It does not measure fonts. Class relation
  labels move to a free baseline when a taller neighboring class blocks their
  nearby positions.
- BPMN output is a subset and claims no conformance. See [PROFILE.md](PROFILE.md).
- Tests parse the SVG and check structure and geometry. They do not show that a
  picture reads well. A person should look at the rendered picture before
  relying on it.
