# BPMN import check

`bpmn-moddle.mjs` reads `diagram.bpmn` files with
[bpmn-moddle](https://github.com/bpmn-io/bpmn-moddle), the BPMN 2.0 model
library used by bpmn.io and Camunda tools. It is opt-in: it is not run in CI,
not run by `go test`, and this repository does not depend on bpmn-moddle.

## Run it

bpmn-moddle is ESM only and is installed outside the repository. The script
loads it through `NODE_PATH`. It needs Node 22 or later and network access to
the npm registry for the install. The version below is the one last used.

```sh
tmp=$(mktemp -d)
npm install --prefix "$tmp" bpmn-moddle@10.3.1
cd modeling/diagram
NODE_PATH="$tmp/node_modules" node verify/bpmn-moddle.mjs examples/generated/*/diagram.bpmn
```

The shell glob matches only the three process examples; the other generated
directories have no `diagram.bpmn`. Exit code 0 means every file passed, 1 means
a file failed, 2 means bad usage or bpmn-moddle could not be loaded.

## What it checks

For each file it fails on:

- any entry in bpmn-moddle's `warnings` (unknown elements, unresolved
  references, invalid attributes) or a parse error;
- a root that is not `bpmn:Definitions`;
- an unresolved `sourceRef`, `targetRef`, `incoming`, `outgoing`,
  `flowNodeRef`, `processRef`, or diagram `bpmnElement`;
- a `toXML` then `fromXML` round trip that produces warnings.

It prints the process element counts by type, the flow count and the
shape and edge counts.

## What it does not establish

It checks the BPMN 2.0 meta-model and reference resolution as implemented by
bpmn-moddle. It is not an OMG XSD validation, it does not check BPMN
conformance classes or execution semantics, and it does not render anything, so
it says nothing about how a tool draws the diagram.
