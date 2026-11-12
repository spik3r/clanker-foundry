# Diagram generator

`diagramgen` turns a reviewed JSON model into a picture. A person or an AI drafts
the model; the generator draws it the same way every time. It follows the
[Foundry generator contract](../CONTRACT.md) and is listed in the
[conformance registry](../gencontract/generators.json).

One profile exists: [`process-diagram/v1`](PROFILE.md), a BPMN-style process
diagram with lanes, tasks, events and gateways. The generator picks a profile by
the model's `profileVersion`. Other diagram profiles can be added as sibling
packages that implement `contract.Profile`; nothing for them exists yet.

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

Generated examples sit under [examples/generated](examples/generated), one
directory per model: [order](examples/order.json), [parallel](examples/parallel.json)
and [rework](examples/rework.json) (a loop back to an earlier step). A test
regenerates them and compares bytes.

## Change a diagram

Edit the model, then regenerate into a new directory and replace the old output.
Never edit `diagram.svg`, `diagram.bpmn` or `model.json` by hand; `check` reports
the edit as drift. The [schema](authoring/process-diagram-v1.schema.json) helps
an editor, but the Go validation decides.

## Limits

- Layout is a fixed, simple algorithm. It is correct and stable, not elegant:
  diagrams are wide and tall, and lines cross. There is no manual positioning;
  to change the picture, change the model.
- A flow leaves a gateway at the top, bottom or right depending on where its
  target is, and its label sits on the first stretch that belongs to that flow
  alone. A gateway has nine exits; flows beyond that share one.
- Label wrapping counts characters. It does not measure fonts.
- BPMN output is a subset and claims no conformance. See [PROFILE.md](PROFILE.md).
- Tests parse the SVG and check structure and geometry. They do not show that a
  picture reads well. A person should look at the rendered picture before
  relying on it.
