# Deterministic diagram starters

Copy a JSON starter into your project. Edit the model, not the generated
picture. The four starters show different views of the same order-intake
system, so you can see which facts belong in each view.

| Start with | Use it to show | Output |
| --- | --- | --- |
| [UML class](uml-class.json) | types, interfaces, members and dependencies | SVG |
| [UML sequence](uml-sequence.json) | who sends what, in what order | SVG |
| [BPMN process](bpmn-process.json) | work ownership, branches and outcomes | SVG and BPMN 2.0 XML |
| [Logical architecture](architecture.json) | actors, services, data stores and external dependencies | SVG |

From a Foundry checkout with the Go version specified by the
[diagram generator](../../modeling/diagram/README.md):

```sh
cd modeling/diagram
model=../../templates/diagrams/architecture.json
out="$(mktemp -d)/diagram"
go run ./cmd/diagramgen validate "$model"
go run ./cmd/diagramgen generate "$model" "$out"
go run ./cmd/diagramgen check "$model" "$out"
```

1. Give each item a stable ID. Update references when an item is renamed.
2. Keep the scope narrow: one type slice, interaction, process or system view.
   Sequence steps/participants, BPMN lanes and architecture nodes keep their
   authored order. Architecture links sort by ID.
3. Review the JSON for meaning and direction. Replace placeholder details with
   confirmed facts. Have the generator validate it, then open the SVG.
4. Keep the source model and, if you commit the output, commit the entire bundle.
   The canonical model and hash manifest are always generated. Run `check` in
   CI to catch hand edits or missing output; regenerate into a fresh directory
   after a reviewed change.

These are bounded profiles, not formal UML or C4 editors, a deployment map, or
an executable BPMN engine. See the
[profile limits](../../modeling/diagram/README.md) before extending a model.
