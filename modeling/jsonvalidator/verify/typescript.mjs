import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';

const fixture = new URL('../testdata/sample/', import.meta.url);
const read = name => readFileSync(new URL(name, fixture));
const sha256 = bytes => createHash('sha256').update(bytes).digest('hex');
const corpus = read('cases.json');
assert.equal(sha256(corpus), '60fb575f37adb783564fbaae7535bf7d59587127dd638c8a0a87ce083d56ef41', 'Frozen synthetic expectations changed');
const cases = JSON.parse(corpus);
assert.equal(cases.length, 46);
assert.equal(new Set(cases.map(item => item.id)).size, cases.length);
assert.equal(cases.filter(item => item.valid).length, 12);

const source = stripTypeScriptTypes(read('validator.ts').toString('utf8'), { mode: 'transform' });
const generated = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
assert.equal(generated.MODEL_SHA256, sha256(read('model.json')));
for (const item of cases) {
  const value = JSON.parse(item.rawJson);
  const before = JSON.stringify(value);
  assert.equal(generated.validate(value), item.valid, item.id);
  assert.equal(JSON.stringify(value), before, `Input mutated: ${item.id}`);
}
console.log(JSON.stringify({ cases: cases.length, passed: cases.length, node: process.version, corpusSHA256: sha256(corpus) }));
