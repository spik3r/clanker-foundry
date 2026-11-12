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
// Independent canonical encoding, written from PROFILE.md: compact JSON, fixed
// member order, properties sorted by name.
const canonicalNode = node => {
  const out = { type: node.type };
  if (node.type === 'boolean' && 'const' in node) out.const = node.const;
  if (node.type === 'array') out.items = canonicalNode(node.items);
  if (node.type === 'object') {
    out.additionalProperties = true;
    out.properties = [...node.properties]
      .sort((a, b) => (a.name < b.name ? -1 : a.name > b.name ? 1 : 0))
      .map(p => ({ name: p.name, required: p.required, schema: canonicalNode(p.schema) }));
  }
  return out;
};
const model = JSON.parse(read('model.json'));
const canonical = JSON.stringify({ profile: model.profile, root: canonicalNode(model.root) });
const manifest = JSON.parse(read('foundry-manifest.json'));
assert.equal(manifest.model.canonicalSHA256, sha256(canonical));
assert.equal(generated.MODEL_SHA256, sha256(canonical));
for (const file of manifest.files) assert.equal(sha256(read(file.path)), file.sha256, file.path);
for (const item of cases) {
  const value = JSON.parse(item.rawJson);
  const before = JSON.stringify(value);
  assert.equal(generated.validate(value), item.valid, item.id);
  assert.equal(JSON.stringify(value), before, `Input mutated: ${item.id}`);
}
console.log(JSON.stringify({ cases: cases.length, passed: cases.length, node: process.version, corpusSHA256: sha256(corpus) }));
