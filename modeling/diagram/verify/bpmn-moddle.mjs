// Opt-in check of diagram.bpmn files with bpmn-moddle. See README.md here.
// Usage: NODE_PATH=<dir>/node_modules node bpmn-moddle.mjs diagram.bpmn...
// Exits 1 on any parse error, warning or unresolved reference, 2 on bad usage.
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';

const files = process.argv.slice(2);
if (files.length === 0) {
  console.error('usage: bpmn-moddle.mjs diagram.bpmn...');
  process.exit(2);
}

let BpmnModdle;
try {
  // createRequire honours NODE_PATH; a bare ESM import does not.
  ({ BpmnModdle } = createRequire(import.meta.url)('bpmn-moddle'));
} catch (err) {
  console.error('cannot load bpmn-moddle (set NODE_PATH, see README.md): ' + err.message);
  process.exit(2);
}

const moddle = new BpmnModdle();
let failed = false;

async function check(file) {
  const problems = [];
  const { rootElement: root, warnings } = await moddle.fromXML(readFileSync(file, 'utf8'));
  for (const w of warnings) problems.push('warning: ' + w.message);
  if (root.$type !== 'bpmn:Definitions') problems.push('root is ' + root.$type);

  const processes = (root.rootElements || []).filter((e) => e.$type === 'bpmn:Process');
  const counts = {};
  let flows = 0;
  for (const p of processes) {
    for (const e of p.flowElements || []) {
      counts[e.$type] = (counts[e.$type] || 0) + 1;
      if (e.$type === 'bpmn:SequenceFlow') {
        flows++;
        if (!e.sourceRef) problems.push('flow ' + e.id + ' has no resolved sourceRef');
        if (!e.targetRef) problems.push('flow ' + e.id + ' has no resolved targetRef');
      } else {
        for (const k of ['incoming', 'outgoing']) {
          for (const f of e[k] || []) {
            if (!f || f.$type !== 'bpmn:SequenceFlow') problems.push(e.id + ' ' + k + ' is unresolved');
          }
        }
      }
    }
    for (const ls of p.laneSets || []) {
      for (const lane of ls.lanes || []) {
        for (const n of lane.flowNodeRef || []) {
          if (!n || !n.$type) problems.push('lane ' + lane.id + ' has an unresolved flowNodeRef');
        }
      }
    }
  }
  for (const c of (root.rootElements || []).filter((e) => e.$type === 'bpmn:Collaboration')) {
    for (const p of c.participants || []) {
      if (!p.processRef) problems.push('participant ' + p.id + ' has no resolved processRef');
    }
  }

  let shapes = 0;
  let edges = 0;
  if ((root.diagrams || []).length === 0) problems.push('no BPMNDiagram');
  for (const d of root.diagrams || []) {
    if (!d.plane || !d.plane.bpmnElement) problems.push('BPMNPlane has no resolved bpmnElement');
    for (const e of (d.plane && d.plane.planeElement) || []) {
      if (e.$type === 'bpmndi:BPMNShape') shapes++;
      else if (e.$type === 'bpmndi:BPMNEdge') edges++;
      if (!e.bpmnElement) problems.push(e.$type + ' ' + e.id + ' has no resolved bpmnElement');
    }
  }

  const { xml } = await moddle.toXML(root, { format: true });
  const again = await moddle.fromXML(xml);
  for (const w of again.warnings) problems.push('round-trip warning: ' + w.message);

  console.log(
    file + ': ' + (problems.length ? 'FAIL' : 'ok') + ' ' + root.$type +
      ' ' + JSON.stringify(counts) + ' flows=' + flows + ' shapes=' + shapes + ' edges=' + edges,
  );
  for (const p of problems) console.log('  ' + p);
  return problems.length === 0;
}

for (const f of files) {
  try {
    if (!(await check(f))) failed = true;
  } catch (err) {
    console.log(f + ': FAIL ' + err.message);
    failed = true;
  }
}
process.exit(failed ? 1 : 0);
