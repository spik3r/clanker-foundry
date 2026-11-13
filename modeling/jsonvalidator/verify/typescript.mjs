import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';

const sha256 = bytes => createHash('sha256').update(bytes).digest('hex');
// Independent canonical encoding, written from PROFILE.md: compact JSON, fixed
// member order, properties sorted by name, enum values sorted by UTF-8 bytes,
// strings with only the double quote and backslash escaped.
const quote = s => '"' + s.replace(/[\\"]/g, '\\$&') + '"';
const canonicalNode = node => {
  const out = [`"type":${quote(node.type)}`];
  if (node.type === 'string') {
    if ('const' in node) out.push(`"const":${quote(node.const)}`);
    if ('enum' in node) out.push(`"enum":[${[...node.enum].sort((a, b) => Buffer.compare(Buffer.from(a), Buffer.from(b))).map(quote).join(',')}]`);
    if ('minLength' in node) out.push(`"minLength":${node.minLength}`);
    if ('maxLength' in node) out.push(`"maxLength":${node.maxLength}`);
  }
  if (node.type === 'boolean' && 'const' in node) out.push(`"const":${node.const}`);
  if (node.type === 'array') out.push(`"items":${canonicalNode(node.items)}`);
  if (node.type === 'object') {
    const props = [...node.properties]
      .sort((a, b) => Buffer.compare(Buffer.from(a.name), Buffer.from(b.name)))
      .map(p => `{"name":${quote(p.name)},"required":${p.required},"schema":${canonicalNode(p.schema)}}`);
    out.push(`"additionalProperties":${node.additionalProperties}`, `"properties":[${props.join(',')}]`);
  }
  return '{' + out.join(',') + '}';
};

// The corpus hashes and counts match verify/acceptance_test.go.
const fixtures = [
  { dir: 'sample', corpus: '60fb575f37adb783564fbaae7535bf7d59587127dd638c8a0a87ce083d56ef41', cases: 46, valid: 12 },
  { dir: 'manifest', corpus: '22c5346e4f2aecb344431f8bc5f0249e63485379baf61970e90bcbbf148dc889', cases: 52, valid: 10 },
  { dir: 'lengths', corpus: 'a1f8df3e54b2802b6f41ce19a2f45147498be51f9c4eab37ed361a208004677f', cases: 50, valid: 25 },
];
const summary = [];
for (const fx of fixtures) {
  const base = new URL(`../testdata/${fx.dir}/`, import.meta.url);
  const read = name => readFileSync(new URL(name, base));
  const corpus = read('cases.json');
  assert.equal(sha256(corpus), fx.corpus, `Frozen expectations changed: ${fx.dir}`);
  const cases = JSON.parse(corpus);
  assert.equal(cases.length, fx.cases);
  assert.equal(new Set(cases.map(item => item.id)).size, cases.length);
  assert.equal(cases.filter(item => item.valid).length, fx.valid);

  const source = stripTypeScriptTypes(read('validator.ts').toString('utf8'), { mode: 'transform' });
  const generated = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
  const model = JSON.parse(read('model.json'));
  const canonical = `{"profile":${quote(model.profile)},"root":${canonicalNode(model.root)}}`;
  const manifest = JSON.parse(read('foundry-manifest.json'));
  assert.equal(manifest.generator.version, '1.1.0');
  assert.equal(manifest.model.canonicalSHA256, sha256(canonical));
  assert.equal(generated.MODEL_SHA256, sha256(canonical));
  for (const file of manifest.files) assert.equal(sha256(read(file.path)), file.sha256, file.path);
  for (const item of cases) {
    const value = JSON.parse(item.rawJson);
    const before = JSON.stringify(value);
    assert.equal(generated.validate(value), item.valid, `${fx.dir}/${item.id}`);
    assert.equal(JSON.stringify(value), before, `Input mutated: ${item.id}`);
  }
  summary.push({ fixture: fx.dir, cases: cases.length, passed: cases.length, corpusSHA256: sha256(corpus) });
}
console.log(JSON.stringify({ node: process.version, fixtures: summary }));
