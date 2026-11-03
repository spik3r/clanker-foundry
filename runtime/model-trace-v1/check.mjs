// A bounded, source-pinned review trace. This does not evaluate commands.
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
const trace = 'runtime/model-trace-v1';
const knownSources = [
  'runtime/src/state.mjs', 'runtime/src/contract.mjs',
  'runtime/contracts/v1.schema.json', 'runtime/contracts/authority.schema.json',
  'runtime/contracts/evidence.schema.json', 'runtime/contracts/workspace.schema.json',
  'runtime/package-lock.json', 'runtime/test/state.test.mjs', 'runtime/test/fixtures.mjs',
  'runtime/test/consumer-shadow-v1.test.mjs',
  `${trace}/testdata/oracle.json`,
];
const sha = (data) => createHash('sha256').update(data).digest('hex');
const insist = (value, message) => { if (!value) throw new Error(message); };
const keys = (object, expected) => object && typeof object === 'object' && !Array.isArray(object)
  && Object.keys(object).sort().join('\0') === [...expected].sort().join('\0');
const unique = (items) => new Set(items).size === items.length;
const sameSet = (a, b) => a.length === b.length && unique(a) && unique(b)
  && [...a].sort().every((item, i) => item === [...b].sort()[i]);

// Pure checker accepts supplied bytes for testability; the CLI reads a closed
// set of fixed repository paths and never trusts manifest paths for file access.
export function verify(manifest, files) {
  insist(keys(manifest, ['formatVersion','mode','authority','evaluatorVersion','notice','sources','commands','states','requirements','flows']), 'manifest fields');
  insist(manifest.formatVersion === 'orchestration-model-trace/v1' && manifest.mode === 'shadow-only'
    && manifest.authority === 'none' && manifest.evaluatorVersion === 1 && typeof manifest.notice === 'string', 'manifest envelope');
  insist(Array.isArray(manifest.sources) && sameSet(manifest.sources.map((x) => x.path), knownSources), 'source inventory');
  for (const source of manifest.sources) {
    insist(keys(source, ['path','sha256']) && /^[a-f0-9]{64}$/.test(source.sha256), 'source shape');
    insist(typeof files[source.path] === 'string' && sha(files[source.path]) === source.sha256, `source pin: ${source.path}`);
  }
  const schema = JSON.parse(files['runtime/contracts/v1.schema.json']);
  const variants = schema.$defs.Command.oneOf;
  const commands = variants.map((x) => x.properties.kind.const);
  const states = schema.$defs.TaskExecution.properties.state.enum;
  insist(Array.isArray(variants) && commands.every((x) => typeof x === 'string') && unique(commands), 'schema command inventory');
  insist(Array.isArray(states) && states.every((x) => typeof x === 'string') && unique(states), 'schema state inventory');
  const inventory = (entries, observed, partial, label) => {
    insist(Array.isArray(entries) && entries.every((e) => keys(e,['id','coverage']) && typeof e.id === 'string'), `${label} shape`);
    insist(sameSet(entries.map((e) => e.id), observed), `${label} inventory`);
    for (const entry of entries) insist(entry.coverage === (partial.includes(entry.id) ? 'partial' : 'unmapped'), `${label} coverage`);
  };
  inventory(manifest.commands, commands, ['ADMIT','ASSIGN'], 'command');
  inventory(manifest.states, states, ['PENDING'], 'state');
  const fixture = JSON.parse(files[`${trace}/testdata/oracle.json`]);
  const ids = fixture.cases.map((c) => c.id);
  insist(unique(ids), 'fixture IDs');
  insist(Array.isArray(manifest.requirements) && sameSet(manifest.requirements.map((x) => x.id), ['TR-ADMIT','TR-DEPENDENCY','TR-RESOURCE']), 'requirement inventory');
  const allCases = [];
  for (const req of manifest.requirements) {
    insist(keys(req,['id','command','statement','anchor','schemaPointer','caseIds','existingTest','existingTestPath','expectedCode']), 'requirement shape');
    insist(commands.includes(req.command) && ['ADMIT','ASSIGN'].includes(req.command), 'requirement command');
    insist(typeof req.statement === 'string' && req.statement.length > 0, 'requirement statement');
    insist(typeof req.anchor === 'string' && req.anchor.length > 0 && files['runtime/src/state.mjs'].includes(req.anchor), 'source anchor');
    const match = /^#\/\$defs\/Command\/oneOf\/(0|[1-9][0-9]*)$/.exec(req.schemaPointer);
    insist(match && variants[Number(match[1])]?.properties?.kind?.const === req.command, 'schema reference');
    insist(['runtime/test/state.test.mjs','runtime/test/consumer-shadow-v1.test.mjs'].includes(req.existingTestPath)
      && typeof req.existingTest === 'string' && files[req.existingTestPath].includes(`test('${req.existingTest}'`), 'test reference');
    insist(Array.isArray(req.caseIds) && req.caseIds.length > 0 && unique(req.caseIds) && req.caseIds.every((id) => ids.includes(id)), 'case reference');
    insist(req.caseIds.every((id) => fixture.cases.find((c) => c.id === id)?.input.command.kind === req.command), 'case command');
    const expectedCode = {'TR-ADMIT':null, 'TR-DEPENDENCY':'DEPENDENCIES_NOT_SATISFIED', 'TR-RESOURCE':'RESOURCE_CONFLICT'}[req.id];
    insist(req.expectedCode === expectedCode, 'requirement outcome');
    if (expectedCode) insist(req.anchor.includes(`'${expectedCode}'`), 'guard outcome reference');
    insist(req.caseIds.every((id) => {
      const c = fixture.cases.find((item) => item.id === id);
      return expectedCode ? c.expected.completion === 'throw' && c.expected.error.code === expectedCode
        : c.expected.completion === 'return';
    }), 'case outcome');
    allCases.push(...req.caseIds);
  }
  insist(sameSet(allCases, ids), 'fixture coverage');
  insist(Array.isArray(manifest.flows) && sameSet(manifest.flows.map((f) => f.id), ['admission','blocked-assignment']), 'flow inventory');
  const expected = {
    admission: {command:'ADMIT', from:'NONE', to:'PENDING', outcome:'return', requirementIds:['TR-ADMIT']},
    'blocked-assignment': {command:'ASSIGN', from:'PENDING', to:'REFUSED', outcome:'throw', requirementIds:['TR-DEPENDENCY','TR-RESOURCE']},
  };
  for (const flow of manifest.flows) {
    insist(keys(flow,['id','command','from','to','outcome','requirementIds']), 'flow shape');
    const want = expected[flow.id];
    insist(want && flow.command === want.command && flow.from === want.from && flow.to === want.to
      && flow.outcome === want.outcome && Array.isArray(flow.requirementIds)
      && sameSet(flow.requirementIds, want.requirementIds), 'flow reference');
    insist(flow.requirementIds.every((id) => manifest.requirements.find((r) => r.id === id)?.command === flow.command), 'flow requirement');
  }
  return true;
}

export function render(manifest) {
  const node = (name) => ({NONE:'start', PENDING:'pending', REFUSED:'refused'})[name];
  const lines = [
    'digraph bounded_shadow {', '  // Partial review trace; not an executable or complete state model.',
    '  rankdir=LR;', '  start [shape=point];', '  pending [label="PENDING (partial)"];',
    '  refused [shape=box,label="REFUSED outcome (no new state)"];',
    ...manifest.flows.map((f) => `  ${node(f.from)} -> ${node(f.to)} [${f.outcome === 'throw' ? 'style=dashed,' : ''}label="${f.command}: ${f.requirementIds.join(' / ')}"];`),
    '}', ''
  ];
  const md = [
    '# Partial orchestration sequence', '',
    'This is a declared review view, not an executable state machine or a complete account of the evaluator.',
    'REFUSED is an outcome marker, not a runtime state. In the invented refusal cases the input remains PENDING at revision 1.', '',
    ...manifest.flows.flatMap((f, i) => [`${i+1}. **${f.command}**: ${f.from} → ${f.to} (${f.outcome}).`,
      ...f.requirementIds.map((id) => { const r = manifest.requirements.find((x) => x.id === id); return `   - **${r.id}**: ${r.statement} Fixtures: ${r.caseIds.join(', ')}.`; })]), '',
    `Commands not modeled: ${manifest.commands.filter((x) => x.coverage === 'unmapped').map((x) => x.id).join(', ')}.`, '',
    `States not modeled: ${manifest.states.filter((x) => x.coverage === 'unmapped').map((x) => x.id).join(', ')}.`, '',
    'Partial entries do not claim all guards, outcomes, transitions, effects or ordering have been modeled.', ''
  ];
  return { [`${trace}/views/state.dot`]: lines.join('\n'), [`${trace}/views/sequence.md`]: md.join('\n') };
}

export function checkViews(manifest, actual) {
  for (const [path, expected] of Object.entries(render(manifest))) insist(actual[path] === expected, `stale view: ${path}`);
  return true;
}

export function load() {
  return Object.fromEntries(knownSources.map((path) => [path, readFileSync(resolve(root,path),'utf8')]));
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    insist(process.argv.length === 3 && ['--check','--write-views'].includes(process.argv[2]), 'usage: node runtime/model-trace-v1/check.mjs --check|--write-views');
    const manifest = JSON.parse(readFileSync(resolve(root,`${trace}/manifest.json`),'utf8'));
    verify(manifest, load());
    const views = render(manifest);
    if (process.argv[2] === '--write-views') {
      for (const [path, bytes] of Object.entries(views)) writeFileSync(resolve(root,path),bytes);
    } else checkViews(manifest, Object.fromEntries(Object.keys(views).map((path) => [path,readFileSync(resolve(root,path),'utf8')])));
    console.log('Partial orchestration trace pins, inventories and views are current.');
  } catch (error) {
    console.error(`trace check failed: ${error.message}`); process.exitCode = 1;
  }
}
