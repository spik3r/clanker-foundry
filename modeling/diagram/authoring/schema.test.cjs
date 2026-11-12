'use strict';

// Checks the process-diagram/v1 authoring schema against the registry examples,
// against independently written invalid models and against the real diagramgen
// generator. Go validation is authoritative; the schema may only be looser than
// Go for rules it cannot express, and each such case is listed with its reason.

const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { createRequire } = require('node:module');
const requireRuntime = createRequire(path.resolve(__dirname, '../../../runtime/package.json'));
const Ajv2020 = requireRuntime('ajv/dist/2020');

const repo = path.resolve(__dirname, '../../..');
const registry = JSON.parse(fs.readFileSync(path.join(repo, 'modeling/profiles.json'), 'utf8'));
const entry = registry.profiles.find(p => p.profile === 'process-diagram/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const readExample = name => fs.readFileSync(path.join(repo, 'modeling/diagram/examples', name), 'utf8');
const base = JSON.parse(readExample('order.json'));

const J = JSON.stringify;
const mut = fn => { const m = JSON.parse(J(base)); fn(m); return J(m); };
const wrap = (nodes, flows, lanes = [{ id: 'L', label: 'Lane' }]) => J({
  formatVersion: 'diagram-json/v1', profileVersion: 'process-diagram/v1', id: 'Gen', title: 'Generated', lanes, nodes, flows,
});
const node = (id, kind, lane = 'L') => ({ id, kind, label: id, lane });
const flow = (id, from, to) => ({ id, from, to });
const chain = tasks => {
  const nodes = [node('S', 'start')];
  const flows = [];
  let prev = 'S';
  for (let i = 0; i < tasks; i++) {
    nodes.push(node('T' + i, 'task'));
    flows.push(flow('F' + i, prev, 'T' + i));
    prev = 'T' + i;
  }
  nodes.push(node('E', 'end'));
  flows.push(flow('Fend', prev, 'E'));
  return wrap(nodes, flows);
};
// 16 exclusive gateways in a forward chain (120 flows), the start flow, one
// flow to the end and six more ends: exactly 128 flows. extra adds more ends.
const maxFlows = (extra = 0) => {
  const nodes = [node('S', 'start')];
  const flows = [flow('FS', 'S', 'G0')];
  for (let i = 0; i < 16; i++) nodes.push(node('G' + i, 'exclusive'));
  for (let i = 0; i < 16; i++) for (let j = i + 1; j < 16; j++) flows.push(flow(`Fg${i}_${j}`, 'G' + i, 'G' + j));
  nodes.push(node('E', 'end'));
  flows.push(flow('FE', 'G15', 'E'));
  for (let k = 0; k < 6 + extra; k++) { nodes.push(node('X' + k, 'end')); flows.push(flow('FX' + k, 'G0', 'X' + k)); }
  return wrap(nodes, flows);
};
const eightLanes = Array.from({ length: 8 }, (_, i) => ({ id: 'L' + i, label: 'Lane ' + i }));
const semantic = 'Go-only semantic rule: ids, references, degrees or reachability';

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'order example', expect: 'valid', text: J(base) },
  { name: 'parallel example', expect: 'valid', text: readExample('parallel.json') },
  { name: 'rework example', expect: 'valid', text: readExample('rework.json') },
  { name: 'smallest diagram', expect: 'valid', text: wrap([node('S', 'start'), node('E', 'end')], [flow('F', 'S', 'E')]) },
  { name: '64 nodes', expect: 'valid', text: chain(62) },
  { name: '8 lanes', expect: 'valid', text: wrap([node('S', 'start', 'L7'), node('E', 'end', 'L0')], [flow('F', 'S', 'E')], eightLanes) },
  { name: '128 flows', expect: 'valid', text: maxFlows() },
  { name: 'underscore in id', expect: 'valid', text: mut(m => { m.id = 'My_Process_2'; }) },
  { name: 'id of 48 characters', expect: 'valid', text: mut(m => { m.id = 'A' + 'b'.repeat(47); }) },
  { name: 'label of 120 characters', expect: 'valid', text: mut(m => { m.title = 'x'.repeat(120); }) },
  { name: 'label with inner and outer spaces', expect: 'valid', text: mut(m => { m.title = '  a b  '; }) },
  { name: 'label of one byte order mark', expect: 'valid', text: mut(m => { m.title = '﻿'; }) },
  { name: 'multibyte label within 120 bytes', expect: 'valid', text: mut(m => { m.title = 'é'.repeat(60); }) },
  { name: 'flow label present', expect: 'valid', text: mut(m => { m.flows[0].label = 'go'; }) },

  { name: 'wrong formatVersion', expect: 'invalid', text: mut(m => { m.formatVersion = 'diagram-json/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(m => { m.profileVersion = 'process-diagram/v2'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.layout = {}; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(m => { m.Title = m.title; }) },
  { name: 'missing flows', expect: 'invalid', text: mut(m => { delete m.flows; }) },
  { name: 'missing title', expect: 'invalid', text: mut(m => { delete m.title; }) },
  { name: 'id starts with a digit', expect: 'invalid', text: mut(m => { m.id = '1Order'; }) },
  { name: 'id of 49 characters', expect: 'invalid', text: mut(m => { m.id = 'A' + 'b'.repeat(48); }) },
  { name: 'id with hyphen', expect: 'invalid', text: mut(m => { m.id = 'My-Process'; }) },
  { name: 'node id with a dot', expect: 'invalid', text: mut(m => { m.nodes[0].id = 'S.1'; }) },
  { name: 'no lanes', expect: 'invalid', text: mut(m => { m.lanes = []; }) },
  { name: '9 lanes', expect: 'invalid', text: wrap([node('S', 'start'), node('E', 'end')], [flow('F', 'S', 'E')], eightLanes.concat([{ id: 'L8', label: 'Lane 8' }])) },
  { name: 'one node', expect: 'invalid', text: mut(m => { m.nodes = [m.nodes[0]]; }) },
  { name: '65 nodes', expect: 'invalid', text: chain(63) },
  { name: 'no flows', expect: 'invalid', text: mut(m => { m.flows = []; }) },
  { name: '129 flows', expect: 'invalid', text: maxFlows(1) },
  { name: 'unknown node kind', expect: 'invalid', text: mut(m => { m.nodes[1].kind = 'subprocess'; }) },
  { name: 'node kind in upper case', expect: 'invalid', text: mut(m => { m.nodes[0].kind = 'Start'; }) },
  { name: 'empty title', expect: 'invalid', text: mut(m => { m.title = ''; }) },
  { name: 'blank title', expect: 'invalid', text: mut(m => { m.title = '   '; }) },
  { name: 'title of only a no-break space', expect: 'invalid', text: mut(m => { m.title = ' '; }) },
  { name: 'title of only ideographic spaces', expect: 'invalid', text: mut(m => { m.title = '　　'; }) },
  { name: 'label of 121 characters', expect: 'invalid', text: mut(m => { m.nodes[1].label = 'x'.repeat(121); }) },
  { name: 'newline in label', expect: 'invalid', text: mut(m => { m.lanes[0].label = 'a\nb'; }) },
  { name: 'tab in label', expect: 'invalid', text: mut(m => { m.nodes[1].label = 'a\tb'; }) },
  { name: 'U+2028 in label', expect: 'invalid', text: mut(m => { m.title = 'a b'; }) },
  { name: 'U+FFFF in label', expect: 'invalid', text: mut(m => { m.title = 'a￿b'; }) },
  { name: 'U+FFFE in label', expect: 'invalid', text: mut(m => { m.title = 'a￾b'; }) },
  { name: 'empty flow label', expect: 'invalid', text: mut(m => { m.flows[0].label = ''; }) },
  { name: 'null flow label', expect: 'invalid', text: mut(m => { m.flows[0].label = null; }) },
  { name: 'lane without label', expect: 'invalid', text: mut(m => { delete m.lanes[0].label; }) },
  { name: 'lane with unknown member', expect: 'invalid', text: mut(m => { m.lanes[0].color = 'red'; }) },
  { name: 'node with coordinates', expect: 'invalid', text: mut(m => { m.nodes[0].x = 10; }) },
  { name: 'flow with unknown member', expect: 'invalid', text: mut(m => { m.flows[0].condition = 'x'; }) },
  { name: 'flow endpoint as number', expect: 'invalid', text: mut(m => { m.flows[0].from = 1; }) },
  { name: 'lanes as object', expect: 'invalid', text: mut(m => { m.lanes = { Sales: 'Sales' }; }) },

  { name: 'lane and node share an id', expect: 'looser', reason: semantic, text: mut(m => { m.nodes[0].id = 'Sales'; m.flows[0].from = 'Sales'; }) },
  { name: 'duplicate node id', expect: 'looser', reason: semantic, text: mut(m => { m.nodes[4].id = 'Ship'; }) },
  { name: 'flow to unknown node', expect: 'looser', reason: semantic, text: mut(m => { m.flows[5].to = 'Nowhere'; }) },
  { name: 'node in unknown lane', expect: 'looser', reason: semantic, text: mut(m => { m.nodes[1].lane = 'Nowhere'; }) },
  { name: 'two start nodes', expect: 'looser', reason: semantic, text: mut(m => { m.nodes[4].kind = 'start'; }) },
  { name: 'no end node', expect: 'looser', reason: semantic, text: mut(m => { m.nodes[5].kind = 'task'; }) },
  { name: 'start with an incoming flow', expect: 'looser', reason: semantic, text: mut(m => { m.flows.push(flow('F7', 'Reject', 'Start')); }) },
  { name: 'task with two outgoing flows', expect: 'looser', reason: semantic, text: mut(m => { m.flows.push(flow('F7', 'Check', 'Ship')); }) },
  { name: 'gateway with two flows in total', expect: 'looser', reason: semantic,
    text: wrap([node('S', 'start'), node('X', 'exclusive'), node('E', 'end')], [flow('F1', 'S', 'X'), flow('F2', 'X', 'E')]) },
  { name: 'self-loop flow', expect: 'looser', reason: semantic, text: mut(m => { m.flows.push(flow('F7', 'InStock', 'InStock')); }) },
  { name: 'repeated flow between two nodes', expect: 'looser', reason: semantic, text: mut(m => { m.flows.push(flow('F7', 'InStock', 'Ship')); }) },
  { name: 'node unreachable from the start', expect: 'looser', reason: semantic,
    text: mut(m => { m.nodes.push(node('Island', 'end', 'Sales')); m.flows.push(flow('F7', 'Reject', 'Island')); m.flows.push(flow('F8', 'Island', 'Reject')); }) },
  { name: 'node that cannot reach an end', expect: 'looser', reason: semantic,
    text: wrap([node('S', 'start'), node('A', 'task'), node('B', 'task'), node('E', 'end'), node('X', 'exclusive')],
      [flow('F1', 'S', 'X'), flow('F2', 'X', 'E'), flow('F3', 'X', 'A'), flow('F4', 'A', 'B'), flow('F5', 'B', 'A')]) },
  { name: 'multibyte label over 120 bytes', expect: 'looser', reason: 'maxLength counts characters, the limit is in bytes',
    text: mut(m => { m.title = 'é'.repeat(61); }) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: J(base).replace('"title":"Order handling",', '"title":"Order handling","title":"Again",') },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data',
    text: J(base) + '{}', parseText: J(base) },
  { name: 'input over 512 KiB', expect: 'looser', reason: 'the size limit applies to source bytes',
    text: J(base).replace(/\}$/, ' '.repeat(524289) + '}') },
];

const parsed = c => JSON.parse(c.parseText || c.text);

test('registry examples satisfy the schema', () => {
  assert.ok(entry.examples.length >= 1);
  for (const rel of entry.examples) {
    const doc = JSON.parse(fs.readFileSync(path.join(repo, rel), 'utf8'));
    assert.equal(validate(doc), true, `${rel}: ${JSON.stringify(validate.errors)}`);
  }
});

test('schema rejects at least 12 invalid models', () => {
  const invalid = cases.filter(c => c.expect === 'invalid');
  assert.ok(invalid.length >= 12, `only ${invalid.length} invalid cases`);
  for (const c of invalid) assert.equal(validate(parsed(c)), false, `${c.name} was accepted`);
});

test('schema accepts the valid cases', () => {
  for (const c of cases.filter(c => c.expect === 'valid')) {
    assert.equal(validate(parsed(c)), true, `${c.name}: ${JSON.stringify(validate.errors)}`);
  }
});

const hasGo = spawnSync('go', ['version'], { encoding: 'utf8' }).status === 0;
test('schema agrees with diagramgen validate', { skip: hasGo ? false : 'go is not available; agreement with the generator was not checked' }, () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'diagram-schema-'));
  try {
    const bin = path.join(dir, 'diagramgen');
    const build = spawnSync('go', ['build', '-o', bin, './cmd/diagramgen'], { cwd: path.join(repo, 'modeling/diagram'), encoding: 'utf8', env: { ...process.env, GOFLAGS: '-mod=readonly' } });
    assert.equal(build.status, 0, `go build failed: ${build.stderr}`);
    assert.ok(cases.length >= 20);
    cases.forEach((c, i) => {
      const file = path.join(dir, `model-${i}.json`);
      fs.writeFileSync(file, c.text);
      const run = spawnSync(bin, ['validate', file], { encoding: 'utf8', env: {} });
      assert.ok(run.status === 0 || run.status === 2, `${c.name}: unexpected exit ${run.status}: ${run.stderr}`);
      const goAccepts = run.status === 0;
      const schemaAccepts = validate(parsed(c)) === true;
      assert.ok(!goAccepts || schemaAccepts, `${c.name}: schema rejects a model Go accepts`);
      assert.equal(goAccepts, c.expect === 'valid', `${c.name}: Go verdict differs from the stated expectation (${run.stderr.trim()})`);
      if (c.expect === 'looser') {
        assert.ok(c.reason, `${c.name}: a looser case needs a reason`);
        assert.equal(schemaAccepts, true, `${c.name}: listed as looser but the schema rejects it`);
      } else {
        assert.equal(schemaAccepts, goAccepts, `${c.name}: schema and Go disagree and no reason is listed`);
      }
    });
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
});
