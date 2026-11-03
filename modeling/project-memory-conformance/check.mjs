import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';
import { isDeepStrictEqual } from 'node:util';
import { pathToFileURL } from 'node:url';
import { buildGraph, completionHeads, sourcePath } from '../../skills/project-memory/scripts/lib/state.mjs';

export const CORPUS_URL = new URL('./testdata/cases.json', import.meta.url);
export const COVERAGE_URL = new URL('./coverage.json', import.meta.url);
const PRODUCER_URL = new URL('../../skills/project-memory/scripts/lib/state.mjs', import.meta.url);
const CONTRACT_URL = new URL('../../skills/project-memory/FORMAT.md', import.meta.url);
const MAX_BYTES = 8 * 1024 * 1024;
const ID = /^[a-z][a-z0-9-]{0,79}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const hash = (value) => createHash('sha256').update(value).digest('hex');
const fail = (message) => { throw new Error(`conformance fixture: ${message}`); };
const object = (value) => value !== null && typeof value === 'object' && !Array.isArray(value);
const text = (value) => typeof value === 'string' && value.length > 0 && value.length <= 4096 && value.isWellFormed();
function keys(value, expected, label) {
  if (!object(value) || !isDeepStrictEqual(Object.keys(value).sort(), [...expected].sort())) fail(`${label}: fields`);
}
function array(value, maximum, label) {
  if (!Array.isArray(value) || value.length > maximum) fail(`${label}: array limit`);
}
function uniqueStrings(value, maximum, pattern, label) {
  array(value, maximum, label);
  if (new Set(value).size !== value.length || value.some((item) => typeof item !== 'string' || !pattern.test(item))) fail(`${label}: values`);
}
function boundedJSON(value) {
  const pending = [[value, 0]];
  let nodes = 0;
  while (pending.length) {
    const [item, depth] = pending.pop();
    if (++nodes > 250000 || depth > 32) fail('tree limit');
    if (item !== null && typeof item === 'object') {
      for (const child of Object.values(item)) pending.push([child, depth + 1]);
    } else if (typeof item === 'string' && !item.isWellFormed()) fail('invalid Unicode');
    else if (typeof item === 'number' && !Number.isFinite(item)) fail('non-finite JSON number');
    else if (item !== null && !['string', 'boolean', 'number'].includes(typeof item)) fail('non-JSON value');
  }
}

// This schema describes trusted checked-in test data, not the producer's policy.
export function validateCorpus(corpus) {
  boundedJSON(corpus);
  keys(corpus, ['formatVersion', 'rules', 'cases'], 'corpus');
  if (corpus.formatVersion !== 'project-memory-conformance-cases/v1') fail('version');
  array(corpus.rules, 128, 'rules'); array(corpus.cases, 128, 'cases');
  if (!corpus.rules.length || !corpus.cases.length) fail('empty inventory');
  const rules = new Set(); const cases = new Set();
  for (const rule of corpus.rules) {
    keys(rule, ['id', 'description', 'contractSection'], 'rule');
    if (typeof rule.id !== 'string' || !ID.test(rule.id) || rules.has(rule.id) || !text(rule.description) || !text(rule.contractSection)) fail('rule identity');
    rules.add(rule.id);
  }
  const covered = new Set();
  for (const item of corpus.cases) {
    keys(item, ['id', 'ruleIds', 'sources', 'events', 'freshIds', 'expected'], 'case');
    if (typeof item.id !== 'string' || !ID.test(item.id) || cases.has(item.id)) fail('case identity');
    cases.add(item.id);
    uniqueStrings(item.ruleIds, 128, ID, 'rule references');
    if (!item.ruleIds.length || item.ruleIds.some((id) => !rules.has(id))) fail('unknown or missing rule');
    for (const id of item.ruleIds) covered.add(id);
    array(item.sources, 8, 'sources'); array(item.events, 16, 'events');
    const sources = new Set();
    for (const source of item.sources) {
      if (!object(source) || typeof source.id !== 'string' || typeof source.revision !== 'string') fail('source map identity');
      const path = sourcePath(source);
      if (sources.has(path)) fail('duplicate source map entry');
      sources.add(path);
    }
    if (item.events.some((event) => !object(event))) fail('event object');
    uniqueStrings(item.freshIds, 32, UUID, 'fresh IDs');
    keys(item.expected, ['graph', 'completion', 'heads', 'completionIds'], 'expectation');
    const expected = item.expected;
    if (!['accepted', 'refused'].includes(expected.graph)
        || !['accepted', 'refused', 'not-run'].includes(expected.completion)) fail('expected disposition');
    array(expected.heads, 16, 'heads');
    const taskIds = new Set();
    for (const head of expected.heads) {
      keys(head, ['taskId', 'eventIds'], 'head');
      if (!text(head.taskId) || taskIds.has(head.taskId)) fail('head identity');
      taskIds.add(head.taskId);
      uniqueStrings(head.eventIds, 16, UUID, 'head events');
      if (!head.eventIds.length) fail('empty head');
    }
    uniqueStrings(expected.completionIds, 16, UUID, 'completion IDs');
    if (expected.graph === 'refused' && (expected.completion !== 'not-run' || expected.heads.length || expected.completionIds.length)) fail('refused graph expectation');
    if (expected.graph === 'accepted' && expected.completion === 'not-run') fail('missing completion expectation');
    if (expected.completion !== 'accepted' && expected.completionIds.length) fail('refused completion expectation');
  }
  if (covered.size !== rules.size) fail('rule without cases');
  return corpus;
}

export function readCorpus() {
  const bytes = readFileSync(CORPUS_URL);
  if (bytes.length > MAX_BYTES) fail('byte limit');
  const decoded = new TextDecoder('utf-8', { fatal: true }).decode(bytes);
  return { bytes, corpus: validateCorpus(JSON.parse(decoded)) };
}

export function eventOrders(count) {
  if (!Number.isInteger(count) || count < 0 || count > 16) fail('event count');
  const base = Array.from({ length: count }, (_, index) => index);
  if (count <= 5) {
    const all = [];
    const visit = (prefix, rest) => {
      if (!rest.length) { all.push(prefix); return; }
      rest.forEach((value, index) => visit([...prefix, value], rest.filter((_, at) => at !== index)));
    };
    visit([], base); return all;
  }
  // Larger graphs have bounded ordering probes, not exhaustive permutations.
  const orders = [base, [...base].reverse()];
  for (let offset = 1; offset < count; offset += 1) orders.push([...base.slice(offset), ...base.slice(0, offset)]);
  return orders;
}

function normalized(expected) {
  return { graph: expected.graph, completion: expected.completion,
    heads: expected.heads.map((head) => ({ taskId: head.taskId, eventIds: [...head.eventIds].sort() }))
      .sort((a, b) => a.taskId < b.taskId ? -1 : a.taskId > b.taskId ? 1 : 0),
    completionIds: [...expected.completionIds].sort() };
}

// Only these imports decide graph validity and completion eligibility.
export function observeCase(item, order = item.events.map((_, index) => index)) {
  if (order.length !== item.events.length || new Set(order).size !== order.length
      || order.some((index) => !Number.isInteger(index) || index < 0 || index >= item.events.length)) fail('event order');
  const input = structuredClone({ sources: item.sources, events: order.map((index) => item.events[index]), freshIds: item.freshIds });
  const before = structuredClone(input);
  const sources = new Map(input.sources.map((source) => [sourcePath(source), source]));
  const sourceMapBefore = structuredClone(sources);
  const freshIds = new Set(input.freshIds);
  const freshIdsBefore = new Set(freshIds);
  const actual = { graph: 'refused', completion: 'not-run', heads: [], completionIds: [] };
  let graphError = null; let completionError = null; let graph;
  try { graph = buildGraph(input.events, sources); actual.graph = 'accepted'; }
  catch (error) {
    graphError = String(error?.message ?? error);
    if (error?.constructor !== Error) actual.graph = 'error';
  }
  if (graph) {
    actual.heads = [...graph.tasks.entries()].map(([taskId, task]) => ({ taskId, eventIds: task.heads.map((event) => event.eventId) }));
    try { actual.completionIds = completionHeads(graph, freshIds).map((event) => event.eventId); actual.completion = 'accepted'; }
    catch (error) { actual.completion = error?.constructor === Error ? 'refused' : 'error'; completionError = String(error?.message ?? error); }
  }
  return { actual: normalized(actual), inputUnchanged: isDeepStrictEqual(input, before)
    && isDeepStrictEqual(sources, sourceMapBefore) && isDeepStrictEqual(freshIds, freshIdsBefore), graphError, completionError };
}

export function coverageFor(corpus) {
  validateCorpus(corpus);
  return { formatVersion: 'project-memory-conformance-coverage/v1', mode: 'declared-test-coverage', authority: 'none',
    ruleCount: corpus.rules.length, caseCount: corpus.cases.length,
    orderingRuns: corpus.cases.reduce((count, item) => count + eventOrders(item.events.length).length, 0),
    rules: [...corpus.rules].sort((a, b) => a.id.localeCompare(b.id)).map((rule) => ({ ...rule,
      cases: corpus.cases.filter((item) => item.ruleIds.includes(rule.id)).map((item) => item.id).sort() })) };
}

export function runCorpus(corpus) {
  validateCorpus(corpus);
  const results = corpus.cases.map((item) => {
    const expected = normalized(item.expected);
    const orders = eventOrders(item.events.length);
    let failedOrderings = 0; let firstMismatch = null; let original;
    for (const order of orders) {
      const observed = observeCase(item, order);
      if (!original) original = observed;
      if (!observed.inputUnchanged || !isDeepStrictEqual(observed.actual, expected)) {
        failedOrderings += 1;
        firstMismatch ??= { order, expected, ...observed };
      }
    }
    return { id: item.id, ruleIds: item.ruleIds, matched: failedOrderings === 0,
      orderingRuns: orders.length, failedOrderings, original, firstMismatch };
  });
  return { formatVersion: 'project-memory-conformance-report/v1', mode: 'conformance-only', authority: 'none',
    matched: results.every((item) => item.matched), casesRun: results.length,
    orderingRuns: results.reduce((sum, item) => sum + item.orderingRuns, 0),
    failedCases: results.filter((item) => !item.matched).length, results };
}

export function main(args) {
  if (args.length > 1 || (args[0] && !['check', 'coverage'].includes(args[0]))) throw new Error('usage: node modeling/project-memory-conformance/check.mjs [check|coverage]');
  const { bytes, corpus } = readCorpus();
  const coverage = `${JSON.stringify(coverageFor(corpus), null, 2)}\n`;
  if (args[0] === 'coverage') { writeFileSync(COVERAGE_URL, coverage); return { written: 'modeling/project-memory-conformance/coverage.json', behavior: 'not-run' }; }
  const coverageMatches = readFileSync(COVERAGE_URL, 'utf8') === coverage;
  const report = runCorpus(corpus);
  return { ...report, coverageMatches, matched: report.matched && coverageMatches, corpusSHA256: hash(bytes),
    observedSourceSHA256: { producer: hash(readFileSync(PRODUCER_URL)), contract: hash(readFileSync(CONTRACT_URL)) } };
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    const result = main(process.argv.slice(2));
    process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
    if (result.matched === false) process.exitCode = 1;
  } catch (error) { process.stderr.write(`${error.message}\n`); process.exitCode = 2; }
}
