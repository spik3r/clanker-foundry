import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import {
  COVERAGE_URL, readCorpus, validateCorpus, coverageFor, eventOrders, observeCase, runCorpus, main,
} from '../modeling/project-memory-conformance/check.mjs';

const { bytes, corpus } = readCorpus();
const clone = () => structuredClone(corpus);
const frozenSHA256 = '7d8e85a6595e3c3d6b98e594f9167259cb5c6c302970a51e492ff4c545ac402e';

test('pre-execution literal corpus remains byte-frozen', () => {
  assert.equal(createHash('sha256').update(bytes).digest('hex'), frozenSHA256);
  assert.equal(corpus.rules.length, 14);
  assert.equal(corpus.cases.length, 50);
});

for (const item of corpus.cases) {
  test(`frozen graph contract: ${item.id}`, () => {
    const subset = { ...corpus, rules: corpus.rules.filter((rule) => item.ruleIds.includes(rule.id)), cases: [item] };
    const before = structuredClone(subset);
    const result = runCorpus(subset);
    assert.equal(result.matched, true, JSON.stringify(result.results[0].firstMismatch));
    assert.equal(result.casesRun, 1);
    assert.equal(result.orderingRuns, eventOrders(item.events.length).length);
    assert.equal(result.results[0].original.inputUnchanged, true);
    assert.deepEqual(subset, before);
  });
}

test('the real producer matches all literals in 598 event orders', () => {
  const report = main(['check']);
  assert.equal(report.matched, true);
  assert.equal(report.failedCases, 0);
  assert.equal(report.orderingRuns, 598);
  assert.equal(report.coverageMatches, true);
  assert.equal(report.corpusSHA256, frozenSHA256);
  assert.match(report.observedSourceSHA256.producer, /^[a-f0-9]{64}$/);
  assert.equal(report.authority, 'none');
});

test('coverage is a deterministic declaration inventory, not a test result', () => {
  const expected = `${JSON.stringify(coverageFor(corpus), null, 2)}\n`;
  assert.equal(readFileSync(COVERAGE_URL, 'utf8'), expected);
  const reordered = clone(); reordered.rules.reverse(); reordered.cases.reverse();
  for (const item of reordered.cases) item.ruleIds.reverse();
  assert.deepEqual(coverageFor(reordered), coverageFor(corpus));
  assert.equal(coverageFor(corpus).mode, 'declared-test-coverage');
  assert.equal(Object.hasOwn(coverageFor(corpus), 'matched'), false);
});

test('a wrong literal is a mismatch, never a repaired expectation', () => {
  const changed = clone();
  const item = changed.cases.find((candidate) => candidate.id === 'first-start');
  item.expected = { graph: 'refused', completion: 'not-run', heads: [], completionIds: [] };
  const before = structuredClone(changed);
  const report = runCorpus(changed);
  assert.equal(report.matched, false);
  assert.equal(report.failedCases, 1);
  assert.equal(report.results.find((result) => result.id === item.id).firstMismatch.actual.graph, 'accepted');
  assert.deepEqual(changed, before);
});

test('wrong frontier and completion IDs cannot pass on aggregate counts', () => {
  const changed = clone();
  const first = changed.cases.find((item) => item.id === 'first-start');
  first.expected.heads[0].eventIds = ['ffffffff-ffff-4fff-8fff-ffffffffffff'];
  const completed = changed.cases.find((item) => item.expected.completionIds.length);
  completed.expected.completionIds = ['ffffffff-ffff-4fff-8fff-ffffffffffff'];
  const result = runCorpus(changed);
  assert.equal(result.failedCases, 2);
  assert.equal(result.results.find((item) => item.id === first.id).matched, false);
  assert.equal(result.results.find((item) => item.id === completed.id).matched, false);
});

test('expected refusal never stops an actual completion selection call', () => {
  const changed = clone();
  const completed = changed.cases.find((item) => item.expected.completionIds.length);
  completed.expected.completion = 'refused'; completed.expected.completionIds = [];
  const result = runCorpus(changed).results.find((item) => item.id === completed.id);
  assert.equal(result.matched, false);
  assert.equal(result.original.actual.completion, 'accepted');
  assert.ok(result.original.actual.completionIds.length > 0);
});

test('unexpected producer TypeError cannot masquerade as a contract refusal', () => {
  const item = structuredClone(corpus.cases[0]);
  // Source-parser hardening is outside the graph corpus; this deliberately invalid
  // source makes the existing graph API throw outside its ordinary refusal helper.
  item.sources[0].record = null;
  const observed = observeCase(item);
  assert.equal(observed.actual.graph, 'error');
  assert.equal(observed.actual.completion, 'not-run');
  assert.equal(observed.inputUnchanged, true);
});

const invalid = [
  ['unknown corpus field', (x) => { x.extra = true; }],
  ['unsupported version', (x) => { x.formatVersion = 'v0'; }],
  ['no rules', (x) => { x.rules = []; }],
  ['no cases', (x) => { x.cases = []; }],
  ['too many cases', (x) => { x.cases = Array(129).fill(x.cases[0]); }],
  ['null rule ID', (x) => { x.rules[0].id = null; }],
  ['null case ID', (x) => { x.cases[0].id = null; }],
  ['duplicate rule', (x) => { x.rules.push(x.rules[0]); }],
  ['duplicate case', (x) => { x.cases.push(x.cases[0]); }],
  ['uncovered rule', (x) => { x.rules.push({ id: 'missing', description: 'Missing', contractSection: 'Events and runs' }); }],
  ['unknown rule', (x) => { x.cases[0].ruleIds = ['missing']; }],
  ['empty rule references', (x) => { x.cases[0].ruleIds = []; }],
  ['duplicate references', (x) => { x.cases[0].ruleIds.push(x.cases[0].ruleIds[0]); }],
  ['duplicate source map path', (x) => { x.cases[0].sources.push(x.cases[0].sources[0]); }],
  ['missing source map identity', (x) => { delete x.cases[0].sources[0].id; }],
  ['too many sources', (x) => { x.cases[0].sources = Array(9).fill(x.cases[0].sources[0]); }],
  ['too many events', (x) => { x.cases[0].events = Array(17).fill(x.cases[0].events[0]); }],
  ['primitive event', (x) => { x.cases[0].events = [null]; }],
  ['invalid fresh ID', (x) => { x.cases[0].freshIds = ['missing']; }],
  ['duplicate fresh ID', (x) => { x.cases[0].freshIds = [x.cases[0].events[0].eventId, x.cases[0].events[0].eventId]; }],
  ['unknown expected field', (x) => { x.cases[0].expected.approved = true; }],
  ['unknown graph disposition', (x) => { x.cases[0].expected.graph = 'approved'; }],
  ['unrun accepted graph selection', (x) => { x.cases[0].expected.completion = 'not-run'; }],
  ['refused graph with heads', (x) => { x.cases[0].expected.graph = 'refused'; x.cases[0].expected.completion = 'not-run'; }],
  ['refused selection with IDs', (x) => { const c = x.cases.find((i) => i.expected.completionIds.length); c.expected.completion = 'refused'; }],
  ['duplicate head', (x) => { x.cases[0].expected.heads.push(x.cases[0].expected.heads[0]); }],
  ['empty head', (x) => { x.cases[0].expected.heads[0].eventIds = []; }],
  ['invalid Unicode', (x) => { x.rules[0].description = '\ud800'; }],
  ['non-JSON number', (x) => { x.cases[0].events[0].sequence = Infinity; }],
  ['non-JSON value', (x) => { x.cases[0].events[0].sequence = undefined; }],
  ['deep tree', (x) => { let p = x.cases[0].events[0]; for (let i = 0; i < 33; i += 1) { p.extra = {}; p = p.extra; } }],
];
for (const [name, mutate] of invalid) test(`fixture schema refuses ${name}`, () => {
  const changed = clone(); mutate(changed);
  assert.throws(() => validateCorpus(changed), /conformance fixture:/);
});

test('bounded ordering probes are complete only through five events', () => {
  for (const [count, length] of [[0, 1], [1, 1], [3, 6], [5, 120], [6, 7], [16, 17]]) {
    const orders = eventOrders(count);
    assert.equal(orders.length, length);
    assert.equal(new Set(orders.map(JSON.stringify)).size, length);
    for (const order of orders) assert.deepEqual([...order].sort((a, b) => a - b), Array.from({ length: count }, (_, i) => i));
  }
  for (const count of [-1, 17, 2.5, '3', null]) assert.throws(() => eventOrders(count));
  assert.throws(() => observeCase(corpus.cases[0], []), /event order/);
  assert.throws(() => observeCase(corpus.cases[0], [1]), /event order/);
});

test('CLI rejects unknown options before reading or writing files', () => {
  assert.throws(() => main(['--execute']), /usage:/);
  assert.throws(() => main(['check', 'extra']), /usage:/);
});

test('checks repeat byte-identically without temporal success fields', () => {
  assert.equal(JSON.stringify(main(['check'])), JSON.stringify(main(['check'])));
});
