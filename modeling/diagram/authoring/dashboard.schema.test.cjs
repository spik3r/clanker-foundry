'use strict';

// Checks the dashboard/v1 authoring schema against the registry examples,
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
const entry = registry.profiles.find(p => p.profile === 'dashboard/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const readExample = name => fs.readFileSync(path.join(repo, 'modeling/diagram/examples', name), 'utf8');
const base = JSON.parse(readExample('ops-dashboard.json'));

const J = JSON.stringify;
const mut = fn => { const m = JSON.parse(J(base)); fn(m); return J(m); };
const head = { formatVersion: 'diagram-json/v1', profileVersion: 'dashboard/v1', id: 'Gen', title: 'Generated' };
const stat = id => ({ id, kind: 'stat', label: 'Label', value: 'Value' });
const col = (id, type = 'number') => ({ id, label: id, type });
const textCol = col('K', 'text');
const wrap = (datasets, widgets, layout) => J({ ...head, datasets, widgets, layout });
const stats = n => Array.from({ length: n }, (_, i) => stat('W' + i));
const dataset = (id, ncols, rows) => ({
  id,
  columns: Array.from({ length: ncols }, (_, i) => col('C' + i)),
  rows: Array.from({ length: rows }, () => Array.from({ length: ncols }, () => 1)),
});
const chartOver = rows => wrap(
  [{ id: 'D', columns: [textCol, col('A')], rows: Array.from({ length: rows }, (_, i) => ['r' + i, i]) }],
  [{ id: 'W', kind: 'bar', title: 'T', dataset: 'D', category: 'K', series: ['A'] }], [['W']]);
const oneCell = cell => J({ ...head,
  datasets: [{ id: 'D', columns: [col('A')], rows: [[cell]] }],
  widgets: [{ id: 'W', kind: 'table', title: 'T', dataset: 'D', columns: ['A'] }], layout: [['W']] });
const tableOf = (cols, type = 'number') => J({ ...head,
  datasets: [{ id: 'D', columns: cols.map(c => col(c, type)), rows: [] }],
  widgets: [{ id: 'W', kind: 'table', title: 'T', dataset: 'D', columns: cols }], layout: [['W']] });
const cellModel = (type, cell) => J({ ...head,
  datasets: [{ id: 'D', columns: [col('A', type)], rows: [[cell]] }],
  widgets: [{ id: 'W', kind: 'table', title: 'T', dataset: 'D', columns: ['A'] }], layout: [['W']] });
const semantic = 'Go-only semantic rule: ids, references, column types, row shape or layout coverage';
const numberText = 'Go-only rule on number spelling; JSON.parse has already turned the token into a number';

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'ops example', expect: 'valid', text: readExample('ops-dashboard.json') },
  { name: 'sales example', expect: 'valid', text: readExample('sales-dashboard.json') },
  { name: 'minimal example', expect: 'valid', text: readExample('minimal-dashboard.json') },
  { name: 'no datasets member', expect: 'valid', text: wrap(undefined, [stat('W')], [['W']]) },
  { name: 'empty datasets list', expect: 'valid', text: wrap([], [stat('W')], [['W']]) },
  { name: 'dataset without rows member', expect: 'valid', text: wrap([{ id: 'D', columns: [col('A')] }], [stat('W')], [['W']]) },
  { name: '8 datasets', expect: 'valid', text: wrap(Array.from({ length: 8 }, (_, i) => ({ id: 'D' + i, columns: [col('A')], rows: [] })), [stat('W')], [['W']]) },
  { name: '12 columns', expect: 'valid', text: tableOf(Array.from({ length: 12 }, (_, i) => 'C' + i)) },
  { name: '200 rows', expect: 'valid', text: wrap([dataset('D', 2, 200)], [{ id: 'W', kind: 'table', title: 'T', dataset: 'D', columns: ['C0'] }], [['W']]) },
  { name: '60 chart rows', expect: 'valid', text: chartOver(60) },
  { name: '24 widgets in 12 rows', expect: 'valid', text: wrap([], stats(24), Array.from({ length: 12 }, (_, i) => ['W' + 2 * i, 'W' + (2 * i + 1)])) },
  { name: '4 widgets in a row', expect: 'valid', text: wrap([], stats(4), [['W0', 'W1', 'W2', 'W3']]) },
  { name: 'number spelled 1.50', expect: 'valid', text: cellModel('number', 0).replace('[[0]]', '[[1.50]]') },
  { name: 'negative zero', expect: 'valid', text: cellModel('number', 0).replace('[[0]]', '[[-0]]') },
  { name: 'number 0.000001', expect: 'valid', text: cellModel('number', 0).replace('[[0]]', '[[0.000001]]') },
  { name: 'number 999999999999', expect: 'valid', text: cellModel('number', 0).replace('[[0]]', '[[999999999999]]') },
  { name: 'empty text cell', expect: 'valid', text: cellModel('text', '') },
  { name: 'text cell of 120 characters', expect: 'valid', text: cellModel('text', 'x'.repeat(120)) },
  { name: 'blank text cell', expect: 'valid', text: cellModel('text', '   ') },
  { name: 'label of 120 characters', expect: 'valid', text: mut(m => { m.title = 'x'.repeat(120); }) },

  { name: 'wrong formatVersion', expect: 'invalid', text: mut(m => { m.formatVersion = 'diagram-json/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(m => { m.profileVersion = 'dashboard/v2'; }) },
  { name: 'process profile', expect: 'invalid', text: mut(m => { m.profileVersion = 'process-diagram/v1'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.theme = 'dark'; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(m => { m.Title = m.title; }) },
  { name: 'missing widgets', expect: 'invalid', text: mut(m => { delete m.widgets; }) },
  { name: 'missing layout', expect: 'invalid', text: mut(m => { delete m.layout; }) },
  { name: 'missing title', expect: 'invalid', text: mut(m => { delete m.title; }) },
  { name: 'id starts with a digit', expect: 'invalid', text: mut(m => { m.id = '1Ops'; }) },
  { name: 'id of 49 characters', expect: 'invalid', text: mut(m => { m.id = 'A' + 'b'.repeat(48); }) },
  { name: 'id with hyphen', expect: 'invalid', text: mut(m => { m.widgets[0].id = 'Tot-al'; }) },
  { name: 'empty title', expect: 'invalid', text: mut(m => { m.title = ''; }) },
  { name: 'blank title', expect: 'invalid', text: mut(m => { m.title = '  '; }) },
  { name: 'label of 121 characters', expect: 'invalid', text: mut(m => { m.title = 'x'.repeat(121); }) },
  { name: 'newline in label', expect: 'invalid', text: mut(m => { m.widgets[0].label = 'a\nb'; }) },
  { name: 'U+2028 in label', expect: 'invalid', text: mut(m => { m.title = 'a\u2028b'; }) },
  { name: 'empty subtitle', expect: 'invalid', text: mut(m => { m.subtitle = ''; }) },
  { name: 'null subtitle', expect: 'invalid', text: mut(m => { m.subtitle = null; }) },
  { name: '9 datasets', expect: 'invalid', text: wrap(Array.from({ length: 9 }, (_, i) => ({ id: 'D' + i, columns: [col('A')], rows: [] })), [stat('W')], [['W']]) },
  { name: 'dataset without columns', expect: 'invalid', text: mut(m => { m.datasets[0].columns = []; }) },
  { name: '13 columns', expect: 'invalid', text: tableOf(Array.from({ length: 13 }, (_, i) => 'C' + i)) },
  { name: 'unknown column type', expect: 'invalid', text: mut(m => { m.datasets[0].columns[0].type = 'date'; }) },
  { name: 'column without label', expect: 'invalid', text: mut(m => { delete m.datasets[0].columns[1].label; }) },
  { name: '201 rows', expect: 'invalid', text: wrap([dataset('D', 2, 201)], [{ id: 'W', kind: 'table', title: 'T', dataset: 'D', columns: ['C0'] }], [['W']]) },
  { name: 'empty row', expect: 'invalid', text: mut(m => { m.datasets[0].rows[0] = []; }) },
  { name: 'boolean cell', expect: 'invalid', text: oneCell(true) },
  { name: 'null cell', expect: 'invalid', text: oneCell(null) },
  { name: 'object cell', expect: 'invalid', text: oneCell({}) },
  { name: 'number of magnitude 1e12', expect: 'invalid', text: oneCell(1000000000000) },
  { name: 'text cell of 121 characters', expect: 'invalid', text: oneCell('x'.repeat(121)) },
  { name: 'control character in a cell', expect: 'invalid', text: oneCell('a\u0001b') },
  { name: 'no widgets', expect: 'invalid', text: wrap([], [], [['W']]) },
  { name: '25 widgets', expect: 'invalid', text: wrap([], stats(25), [['W0']]) },
  { name: 'unknown widget kind', expect: 'invalid', text: mut(m => { m.widgets[1].kind = 'pie'; }) },
  { name: 'stat without value', expect: 'invalid', text: mut(m => { delete m.widgets[0].value; }) },
  { name: 'stat with a dataset', expect: 'invalid', text: mut(m => { m.widgets[0].dataset = 'Orders'; }) },
  { name: 'stat with empty note', expect: 'invalid', text: mut(m => { m.widgets[0].note = ''; }) },
  { name: 'bar without series', expect: 'invalid', text: mut(m => { delete m.widgets[1].series; }) },
  { name: 'bar with empty series', expect: 'invalid', text: mut(m => { m.widgets[1].series = []; }) },
  { name: 'bar with 5 series', expect: 'invalid', text: mut(m => { m.widgets[1].series = ['A', 'B', 'C', 'D', 'E']; }) },
  { name: 'bar with duplicate series', expect: 'invalid', text: mut(m => { m.widgets[1].series = ['Placed', 'Placed']; }) },
  { name: 'bar with columns', expect: 'invalid', text: mut(m => { m.widgets[1].columns = ['Day']; }) },
  { name: 'line without category', expect: 'invalid', text: mut(m => { delete m.widgets[2].category; }) },
  { name: 'table without columns', expect: 'invalid', text: mut(m => { delete m.widgets[3].columns; }) },
  { name: 'table with duplicate columns', expect: 'invalid', text: mut(m => { m.widgets[3].columns = ['Day', 'Day']; }) },
  { name: 'table with series', expect: 'invalid', text: mut(m => { m.widgets[3].series = ['Placed']; }) },
  { name: 'widget with a colour', expect: 'invalid', text: mut(m => { m.widgets[1].color = 'red'; }) },
  { name: 'empty layout', expect: 'invalid', text: mut(m => { m.layout = []; }) },
  { name: '13 layout rows', expect: 'invalid', text: wrap([], stats(13), Array.from({ length: 13 }, (_, i) => ['W' + i])) },
  { name: 'empty layout row', expect: 'invalid', text: mut(m => { m.layout.push([]); }) },
  { name: '5 widgets in a row', expect: 'invalid', text: wrap([], stats(5), [['W0', 'W1', 'W2', 'W3', 'W4']]) },
  { name: 'layout row as string', expect: 'invalid', text: mut(m => { m.layout[0] = 'Total'; }) },

  { name: 'duplicate dataset id', expect: 'looser', reason: semantic, text: wrap([{ id: 'D', columns: [col('A')] }, { id: 'D', columns: [col('A')] }], [stat('W')], [['W']]) },
  { name: 'duplicate widget id', expect: 'looser', reason: semantic, text: wrap([], [stat('W'), stat('W')], [['W']]) },
  { name: 'duplicate column id', expect: 'looser', reason: semantic, text: wrap([{ id: 'D', columns: [col('A'), col('A')] }], [stat('W')], [['W']]) },
  { name: 'row with too few cells', expect: 'looser', reason: semantic, text: mut(m => { m.datasets[0].rows[0] = ['Mon', 120]; }) },
  { name: 'row with too many cells', expect: 'looser', reason: semantic, text: mut(m => { m.datasets[0].rows[0] = ['Mon', 120, 96, 1]; }) },
  { name: 'number in a text column', expect: 'looser', reason: semantic, text: mut(m => { m.datasets[0].rows[0][0] = 5; }) },
  { name: 'text in a number column', expect: 'looser', reason: semantic, text: mut(m => { m.datasets[0].rows[0][1] = '120'; }) },
  { name: 'bar over an unknown dataset', expect: 'looser', reason: semantic, text: mut(m => { m.widgets[1].dataset = 'Nope'; }) },
  { name: 'bar over an unknown category', expect: 'looser', reason: semantic, text: mut(m => { m.widgets[1].category = 'Nope'; }) },
  { name: 'bar with a number category', expect: 'looser', reason: semantic, text: mut(m => { m.widgets[1].category = 'Placed'; m.widgets[1].series = ['Shipped']; }) },
  { name: 'bar with a text series', expect: 'looser', reason: semantic, text: mut(m => { m.widgets[1].series = ['Day']; }) },
  { name: 'table with an unknown column', expect: 'looser', reason: semantic, text: mut(m => { m.widgets[3].columns = ['Day', 'Nope']; }) },
  { name: 'layout names an unknown widget', expect: 'looser', reason: semantic, text: mut(m => { m.layout[2] = ['Nope']; }) },
  { name: 'layout omits a widget', expect: 'looser', reason: semantic, text: mut(m => { m.layout.pop(); }) },
  { name: 'layout repeats a widget', expect: 'looser', reason: semantic, text: mut(m => { m.layout.push(['Total']); }) },
  { name: 'chart over 61 rows', expect: 'looser', reason: 'Go-only limit: a chart dataset has at most 60 rows', text: chartOver(61) },
  { name: 'exponent in a number', expect: 'looser', reason: numberText, text: oneCell(0).replace('[[0]]', '[[1e3]]') },
  { name: 'number with 7 fractional digits', expect: 'looser', reason: numberText, text: oneCell(0).replace('[[0]]', '[[0.0000001]]') },
  { name: 'number with 16 significant digits', expect: 'looser', reason: numberText, text: oneCell(0).replace('[[0]]', '[[123456789012.3456]]') },
  { name: 'multibyte label over 120 bytes', expect: 'looser', reason: 'maxLength counts characters, the limit is in bytes', text: mut(m => { m.title = 'é'.repeat(61); }) },
  { name: 'multibyte cell over 120 bytes', expect: 'looser', reason: 'maxLength counts characters, the limit is in bytes', text: oneCell('é'.repeat(61)) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: J(base).replace('"title":"Weekly operations",', '"title":"Weekly operations","title":"Again",') },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data', text: J(base) + '{}', parseText: J(base) },
  { name: 'input over 512 KiB', expect: 'looser', reason: 'the size limit applies to source bytes',
    text: J(base).replace(/\}$/, ' '.repeat(524289) + '}') },
];

const parsed = c => JSON.parse(c.parseText || c.text);

test('registry examples satisfy the schema', () => {
  assert.ok(entry.examples.length >= 3);
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
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'dashboard-schema-'));
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
