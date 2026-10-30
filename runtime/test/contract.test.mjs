import test from 'node:test';
import assert from 'node:assert/strict';
import { canonical, digest, parseJSON, validate } from '../src/contract.mjs';
import {
  command,
  effect,
  specDigest,
  successfulResult,
  taskEvent,
  taskExecution,
  taskSpec,
} from './fixtures.mjs';

test('validates the shared v1 records without changing them', () => {
  for (const [type, value] of [
    ['TaskSpec', taskSpec],
    ['TaskExecution', taskExecution],
    ['Command', command],
    ['WorkerResult', successfulResult],
    ['TaskEvent', taskEvent],
    ['Effect', effect],
  ]) {
    assert.equal(validate(type, value), value);
  }
  assert.equal(specDigest, digest(taskSpec));
});

test('rejects unknown fields, wrong types, and mismatched result variants', () => {
  assert.throws(
    () => validate('TaskSpec', { ...taskSpec, unexpected: true }),
    { code: 'INVALID_CONTRACT' },
  );
  assert.throws(
    () => validate('TaskSpec', { ...taskSpec, spec_revision: '1' }),
    { code: 'INVALID_CONTRACT' },
  );
  assert.throws(
    () => validate('WorkerResult', { ...successfulResult, failure_code: 'oops' }),
    { code: 'INVALID_CONTRACT' },
  );
  assert.throws(() => validate('unknown', {}), { code: 'UNKNOWN_CONTRACT' });
});

test('canonical JSON sorts object keys and preserves array order', () => {
  assert.equal(canonical({ z: [2, 1], a: { y: true, x: null } }), '{"a":{"x":null,"y":true},"z":[2,1]}');
  assert.equal(digest({ b: 2, a: 1 }), digest({ a: 1, b: 2 }));
  assert.notEqual(digest([1, 2]), digest([2, 1]));
  assert.equal(digest({ a: 1 }), '015abd7f5cc57a2dd94b7590f04ad8084273905ee33ec5cebeae62276a97f862');
});

test('canonical JSON rejects values outside its deterministic data model', () => {
  const cyclic = {};
  cyclic.self = cyclic;
  for (const value of [undefined, 1n, Number.NaN, Infinity, cyclic, new Date(), [, 1]]) {
    assert.throws(() => canonical(value), { code: 'UNSUPPORTED_VALUE' });
  }
  assert.throws(() => canonical({ ['__proto__']: 'unsafe' }), { code: 'UNSAFE_JSON_KEY' });
});

test('canonical JSON rejects sparse arrays, accessors, and hidden properties without reading them', () => {
  const sparse = new Array(1);
  sparse.extra = 'value';
  let getterCalls = 0;
  const accessor = [];
  Object.defineProperty(accessor, '0', {
    enumerable: true,
    get() { getterCalls += 1; return 'unsafe'; },
  });
  accessor.length = 1;
  const hidden = [1];
  Object.defineProperty(hidden, 'extra', { value: 'hidden', enumerable: false });
  assert.throws(() => canonical(sparse), { code: 'UNSUPPORTED_VALUE' });
  assert.throws(() => canonical(accessor), { code: 'UNSUPPORTED_VALUE' });
  assert.throws(() => canonical(hidden), { code: 'UNSUPPORTED_VALUE' });
  assert.equal(getterCalls, 0);
});

test('parseJSON rejects duplicate keys, prototype keys, depth and byte overflow', () => {
  assert.throws(() => parseJSON('{"task":1,"task":2}'), { code: 'DUPLICATE_JSON_KEY' });
  assert.throws(() => parseJSON('{"__proto__":{"polluted":true}}'), { code: 'UNSAFE_JSON_KEY' });
  assert.throws(() => parseJSON('[[[0]]]', { maxDepth: 2 }), { code: 'JSON_TOO_DEEP' });
  assert.throws(() => parseJSON('{"x":"long"}', { maxBytes: 4 }), { code: 'JSON_TOO_LARGE' });
  assert.throws(() => parseJSON('{"x":}'), { code: 'MALFORMED_JSON' });
});

test('parseJSON parses valid JSON with a null-prototype object', () => {
  const value = parseJSON('{"items":[1,true,null],"text":"hello"}');
  assert.equal(Object.getPrototypeOf(value), null);
  assert.deepEqual(value.items, [1, true, null]);
});
