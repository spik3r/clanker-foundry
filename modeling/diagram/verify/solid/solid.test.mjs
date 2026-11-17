import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { after, before, test } from 'node:test';
import { transformSync } from '@babel/core';
import solid from 'babel-preset-solid';
import { parseFragment } from 'parse5';
import { renderToString } from 'solid-js/web';

const here = dirname(fileURLToPath(import.meta.url));
const diagram = resolve(here, '../..');
let work, binary;
let serial = 0;

before(() => {
  // Keep compiled modules under this package so their bare imports resolve
  // against its pinned dependencies, without NODE_PATH or consumer packages.
  work = mkdtempSync(join(here, '.generated-'));
  binary = join(work, process.platform === 'win32' ? 'diagramgen.exe' : 'diagramgen');
  execFileSync('go', ['build', '-o', binary, './cmd/diagramgen'], { cwd: diagram });
});
after(() => { if (work) rmSync(work, { recursive: true, force: true }); });

function model(name) {
  const value = JSON.parse(readFileSync(join(diagram, 'examples', name), 'utf8'));
  value.target = 'solid';
  return value;
}

function compile(source) {
  return transformSync(source, {
    filename: 'Dashboard.jsx', babelrc: false, configFile: false,
    presets: [[solid, { generate: 'ssr', hydratable: false }]],
  }).code;
}

async function render(value, props = {}) {
  const id = serial++;
  const input = join(work, `model-${id}.json`);
  const output = join(work, `output-${id}`);
  writeFileSync(input, JSON.stringify(value));
  execFileSync(binary, ['generate', input, output], { cwd: diagram });
  const jsx = readFileSync(join(output, 'Dashboard.jsx'), 'utf8');
  const modulePath = join(work, `dashboard-${id}.mjs`);
  writeFileSync(modulePath, compile(jsx));
  const { default: Dashboard } = await import(pathToFileURL(modulePath).href);
  return parseFragment(renderToString(() => Dashboard(props)));
}

function all(node, tag) {
  return (node.childNodes ?? []).flatMap(child => [
    ...(child.tagName === tag ? [child] : []), ...all(child, tag),
  ]);
}
function attr(node, name) { return node.attrs?.find(a => a.name === name)?.value; }
function text(node) {
  return node.nodeName === '#text' ? node.value : (node.childNodes ?? []).map(text).join('');
}
function widget(tree, id) {
  const result = all(tree, 'article').find(node => attr(node, 'data-id') === id);
  assert.ok(result, `missing widget ${id}`);
  return result;
}
function assertOps(tree) {
  assert.deepEqual(all(tree, 'article').map(n => attr(n, 'data-id')), ['Total', 'ByDay', 'Trend', 'Detail']);
  assert.equal(text(all(tree, 'h1')[0]), 'Weekly operations');
  assert.deepEqual(all(widget(tree, 'Total'), 'p').map(text), ['Orders placed', '872', '+4% on last week']);
  const table = widget(tree, 'Detail');
  assert.deepEqual(all(table, 'th').map(text), ['Day', 'Placed', 'Shipped']);
  assert.deepEqual(all(all(table, 'tbody')[0], 'tr').map(row => all(row, 'td').map(text)), [
    ['Mon', '120', '96'], ['Tue', '135', '130'], ['Wed', '128', '124'],
    ['Thu', '142', '133'], ['Fri', '151', '147'], ['Sat', '98', '102'], ['Sun', '98', '110'],
  ]);
  const bars = all(widget(tree, 'ByDay'), 'svg').find(n => attr(n, 'role') === 'img');
  assert.equal(attr(bars, 'viewBox'), '0 0 640 320');
  const marks = all(bars, 'rect').filter(n => all(n, 'title').length);
  assert.equal(marks.length, 14);
  assert.equal(text(all(marks[0], 'title')[0]), 'Placed, Mon: 120');
  // The y axis spans 0..200 over 264 units: 120 occupies 158.4 units.
  assert.equal(attr(marks[0], 'height'), '158.40');
  const line = all(widget(tree, 'Trend'), 'svg').find(n => attr(n, 'role') === 'img');
  assert.equal(all(line, 'path').length, 1);
  assert.equal(all(line, 'circle').length, 7);
  assert.equal(text(all(all(line, 'circle')[6], 'title')[0]), 'Shipped, Sun: 110');
}

test('fresh generated ops component compiles and renders literal content and SVG', async () => {
  const tree = await render(model('ops-dashboard-solid.json'), { class: 'host-panel' });
  assertOps(tree);
  assert.equal(attr(all(tree, 'div')[0], 'class'), 'fd-dashboard host-panel');
});

test('stat-only and multi-dataset components compile and render', async () => {
  const minimal = await render(model('minimal-dashboard.json'));
  assert.equal(text(all(minimal, 'h1')[0]), 'Minimal dashboard');
  assert.deepEqual(all(widget(minimal, 'Status'), 'p').map(text), ['Status', 'All systems normal']);
  assert.equal(all(minimal, 'table').length, 0);
  const sales = await render(model('sales-dashboard.json'));
  assert.equal(all(sales, 'article').length, 7);
  assert.equal(all(widget(sales, 'RegionTable'), 'td').length, 45);
  assert.equal(all(widget(sales, 'WideTable'), 'td').length, 72);
  assert.deepEqual(all(all(widget(sales, 'WideTable'), 'tbody')[0], 'tr').slice(0, 1).map(n => all(n, 'td').map(text)),
    [['Row 1', '20.25', '33.5', '46.75', '30', '73.25', '86.5', '2.75', '16', '-0.75', '42.5', '55.75']]);
});

test('hostile model text stays text after compilation and HTML parsing', async () => {
  const attack = '</script><img src=x onerror="alert(1)"><script>alert(2)</script>&"\'` ${globalThis.pwned} {x}';
  const value = model('ops-dashboard-solid.json');
  value.title = attack;
  value.widgets[0].label = attack;
  value.widgets[0].value = attack;
  value.widgets[3].title = attack;
  value.datasets[0].columns[0].label = attack;
  value.datasets[0].rows[0][0] = attack;
  const tree = await render(value);
  assert.equal(text(all(tree, 'h1')[0]), attack);
  assert.equal(text(all(widget(tree, 'Total'), 'p')[0]), attack);
  assert.equal(text(all(widget(tree, 'Total'), 'p')[1]), attack);
  assert.equal(text(all(widget(tree, 'Detail'), 'th')[0]), attack);
  assert.equal(text(all(widget(tree, 'Detail'), 'td')[0]), attack);
  assert.equal(attr(all(tree, 'div').find(n => attr(n, 'role') === 'region'), 'aria-label'), attack);
  assert.equal(all(tree, 'script').length, 0);
  assert.equal(all(tree, 'img').length, 0);
  function check(node) {
    for (const a of node.attrs ?? []) assert.ok(!/^on/i.test(a.name), `unexpected handler ${a.name}`);
    for (const child of node.childNodes ?? []) check(child);
  }
  check(tree);
});

test('negative controls reject invalid JSX and changed content', async () => {
  assert.throws(() => compile('export default function Broken() { return <div>; }'));
  const changed = model('ops-dashboard-solid.json');
  changed.widgets[0].value = '873';
  const tree = await render(changed);
  assert.throws(() => assertOps(tree), { name: 'AssertionError' });
});
