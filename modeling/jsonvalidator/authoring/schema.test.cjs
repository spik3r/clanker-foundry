'use strict';

// Checks the json-boolean/v1 authoring schema against the registry examples,
// against independently written invalid models and against the real jsonvalidator
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
const entry = registry.profiles.find(p => p.profile === 'json-boolean/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const base = JSON.parse(fs.readFileSync(path.join(repo, 'modeling/jsonvalidator/testdata/sample/model.json'), 'utf8'));

const J = JSON.stringify;
const str = { type: 'string' };
const obj = (...props) => ({ type: 'object', additionalProperties: true, properties: props });
const prop = (name, schema, required = true) => ({ name, required, schema });
const doc = root => J({ profile: 'json-boolean/v1', root });
const nest = depth => { let n = str; for (let i = 1; i < depth; i++) n = { type: 'array', items: n }; return n; };
const names = n => Array.from({ length: n }, (_, i) => 'p' + i);
const mut = fn => { const m = JSON.parse(J(base)); fn(m); return J(m); };

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'sample model', expect: 'valid', text: J(base) },
  { name: 'string root', expect: 'valid', text: doc(str) },
  { name: 'boolean root', expect: 'valid', text: doc({ type: 'boolean' }) },
  { name: 'boolean const true', expect: 'valid', text: doc({ type: 'boolean', const: true }) },
  { name: 'boolean const false', expect: 'valid', text: doc({ type: 'boolean', const: false }) },
  { name: 'array of strings', expect: 'valid', text: doc({ type: 'array', items: str }) },
  { name: 'empty object', expect: 'valid', text: doc(obj()) },
  { name: 'optional property', expect: 'valid', text: doc(obj(prop('a', str, false))) },
  { name: 'names with underscore and dollar', expect: 'valid', text: doc(obj(prop('_a', str), prop('$b', str), prop('c1_$', str))) },
  { name: 'property name of 128 characters', expect: 'valid', text: doc(obj(prop('a'.repeat(128), str))) },
  { name: '128 properties', expect: 'valid', text: doc(obj(...names(128).map(n => prop(n, str)))) },
  { name: 'schema nested to depth 32', expect: 'valid', text: doc(nest(32)) },
  { name: 'string const', expect: 'valid', text: doc({ type: 'string', const: 'a' }) },
  { name: 'string const empty', expect: 'valid', text: doc({ type: 'string', const: '' }) },
  { name: 'string const of 256 characters', expect: 'valid', text: doc({ type: 'string', const: 'a'.repeat(256) }) },
  { name: 'string enum', expect: 'valid', text: doc({ type: 'string', enum: ['scanner', 'reporter', 'exporter'] }) },
  { name: 'string enum of one value', expect: 'valid', text: doc({ type: 'string', enum: ['a'] }) },
  { name: 'string enum of 32 values', expect: 'valid', text: doc({ type: 'string', enum: names(32) }) },
  { name: 'string minLength', expect: 'valid', text: doc({ type: 'string', minLength: 1 }) },
  { name: 'string maxLength', expect: 'valid', text: doc({ type: 'string', maxLength: 65535 }) },
  { name: 'string both bounds', expect: 'valid', text: doc({ type: 'string', minLength: 0, maxLength: 0 }) },
  { name: 'closed object', expect: 'valid', text: doc({ type: 'object', additionalProperties: false, properties: [] }) },
  { name: 'manifest example model', expect: 'valid', text: fs.readFileSync(path.join(repo, 'modeling/jsonvalidator/testdata/manifest/model.json'), 'utf8') },
  { name: 'lengths example model', expect: 'valid', text: fs.readFileSync(path.join(repo, 'modeling/jsonvalidator/testdata/lengths/model.json'), 'utf8') },
  { name: 'object in array in object', expect: 'valid', text: doc(obj(prop('xs', { type: 'array', items: obj(prop('y', { type: 'boolean', const: false })) }))) },

  { name: 'wrong profile', expect: 'invalid', text: mut(m => { m.profile = 'json-boolean/v2'; }) },
  { name: 'lifecycle profile name', expect: 'invalid', text: mut(m => { m.profile = 'lifecycle/v1'; }) },
  { name: 'missing profile', expect: 'invalid', text: mut(m => { delete m.profile; }) },
  { name: 'missing root', expect: 'invalid', text: mut(m => { delete m.root; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.extra = true; }) },
  { name: 'root is an array', expect: 'invalid', text: mut(m => { m.root = [m.root]; }) },
  { name: 'root without type', expect: 'invalid', text: doc({ additionalProperties: true, properties: [] }) },
  { name: 'unknown type number', expect: 'invalid', text: doc({ type: 'number' }) },
  { name: 'type in upper case', expect: 'invalid', text: doc({ type: 'String' }) },
  { name: 'type null', expect: 'invalid', text: doc({ type: 'null' }) },
  { name: 'string with unknown member', expect: 'invalid', text: doc({ type: 'string', pattern: '.*' }) },
  { name: 'string with default', expect: 'invalid', text: doc({ type: 'string', default: 'a' }) },
  { name: 'string const as number', expect: 'invalid', text: doc({ type: 'string', const: 1 }) },
  { name: 'string const null', expect: 'invalid', text: doc({ type: 'string', const: null }) },
  { name: 'string const 257 characters', expect: 'invalid', text: doc({ type: 'string', const: 'a'.repeat(257) }) },
  { name: 'string const with enum', expect: 'invalid', text: doc({ type: 'string', const: 'a', enum: ['a'] }) },
  { name: 'string const with minLength', expect: 'invalid', text: doc({ type: 'string', const: 'a', minLength: 1 }) },
  { name: 'string enum with maxLength', expect: 'invalid', text: doc({ type: 'string', enum: ['a'], maxLength: 3 }) },
  { name: 'string enum empty', expect: 'invalid', text: doc({ type: 'string', enum: [] }) },
  { name: 'string enum with 33 values', expect: 'invalid', text: doc({ type: 'string', enum: names(33) }) },
  { name: 'string enum with a number', expect: 'invalid', text: doc({ type: 'string', enum: ['a', 1] }) },
  { name: 'string enum not an array', expect: 'invalid', text: doc({ type: 'string', enum: 'a' }) },
  { name: 'string enum value of 257 characters', expect: 'invalid', text: doc({ type: 'string', enum: ['a'.repeat(257)] }) },
  { name: 'string minLength negative', expect: 'invalid', text: doc({ type: 'string', minLength: -1 }) },
  { name: 'string minLength fraction', expect: 'invalid', text: doc({ type: 'string', minLength: 1.5 }) },
  { name: 'string minLength as text', expect: 'invalid', text: doc({ type: 'string', minLength: '1' }) },
  { name: 'string maxLength 65536', expect: 'invalid', text: doc({ type: 'string', maxLength: 65536 }) },
  { name: 'string maxLength null', expect: 'invalid', text: doc({ type: 'string', maxLength: null }) },
  { name: 'string with only an unknown length member', expect: 'invalid', text: doc({ type: 'string', length: 3 }) },
  { name: 'boolean const as string', expect: 'invalid', text: doc({ type: 'boolean', const: 'true' }) },
  { name: 'boolean const as number', expect: 'invalid', text: doc({ type: 'boolean', const: 1 }) },
  { name: 'boolean const null', expect: 'invalid', text: doc({ type: 'boolean', const: null }) },
  { name: 'boolean with extra member', expect: 'invalid', text: doc({ type: 'boolean', nullable: true }) },
  { name: 'array without items', expect: 'invalid', text: doc({ type: 'array' }) },
  { name: 'array with null items', expect: 'invalid', text: doc({ type: 'array', items: null }) },
  { name: 'array with extra member', expect: 'invalid', text: doc({ type: 'array', items: str, minItems: 1 }) },
  { name: 'object without properties', expect: 'invalid', text: doc({ type: 'object', additionalProperties: true }) },
  { name: 'object with null properties', expect: 'invalid', text: doc({ type: 'object', additionalProperties: true, properties: null }) },
  { name: 'object without additionalProperties', expect: 'invalid', text: doc({ type: 'object', properties: [] }) },
  { name: 'additionalProperties as null', expect: 'invalid', text: doc({ type: 'object', additionalProperties: null, properties: [] }) },
  { name: 'additionalProperties as string', expect: 'invalid', text: doc({ type: 'object', additionalProperties: 'true', properties: [] }) },
  { name: 'object with extra member', expect: 'invalid', text: doc({ type: 'object', additionalProperties: true, properties: [], required: [] }) },
  { name: 'properties as a map', expect: 'invalid', text: doc({ type: 'object', additionalProperties: true, properties: { a: str } }) },
  { name: '129 properties', expect: 'invalid', text: doc(obj(...names(129).map(n => prop(n, str)))) },
  { name: 'property without required', expect: 'invalid', text: doc(obj({ name: 'a', schema: str })) },
  { name: 'property without schema', expect: 'invalid', text: doc(obj({ name: 'a', required: true })) },
  { name: 'property without name', expect: 'invalid', text: doc(obj({ required: true, schema: str })) },
  { name: 'property with extra member', expect: 'invalid', text: doc(obj({ name: 'a', required: true, schema: str, default: 'x' })) },
  { name: 'required as string', expect: 'invalid', text: doc(obj(prop('a', str, 'true'))) },
  { name: 'empty property name', expect: 'invalid', text: doc(obj(prop('', str))) },
  { name: 'property name of 129 characters', expect: 'invalid', text: doc(obj(prop('a'.repeat(129), str))) },
  { name: 'property name starting with a digit', expect: 'invalid', text: doc(obj(prop('1a', str))) },
  { name: 'property name with hyphen', expect: 'invalid', text: doc(obj(prop('a-b', str))) },
  { name: 'property name with space', expect: 'invalid', text: doc(obj(prop('a b', str))) },
  { name: 'non-ASCII property name', expect: 'invalid', text: doc(obj(prop('café', str))) },
  { name: 'property name with a newline', expect: 'invalid', text: doc(obj(prop('a\n', str))) },
  { name: 'invalid nested schema', expect: 'invalid', text: doc(obj(prop('a', { type: 'array', items: { type: 'integer' } }))) },

  { name: 'duplicate property names', expect: 'looser', reason: 'unique names within one object are not expressible',
    text: doc(obj(prop('a', str), prop('a', { type: 'boolean' }))) },
  { name: 'duplicate enum values', expect: 'looser', reason: 'unique enum values are not expressible',
    text: doc({ type: 'string', enum: ['a', 'b', 'a'] }) },
  { name: 'minLength above maxLength', expect: 'looser', reason: 'comparing two members is not expressible',
    text: doc({ type: 'string', minLength: 3, maxLength: 2 }) },
  { name: 'const of 100 three-byte characters', expect: 'looser', reason: 'the schema counts characters, Go counts the 256-byte limit',
    text: doc({ type: 'string', const: '\u20ac'.repeat(100) }) },
  { name: 'const with a control character', expect: 'looser', reason: 'refusing control characters is not expressed',
    text: doc({ type: 'string', const: 'a\u0001' }) },
  { name: 'const with U+FFFD', expect: 'looser', reason: 'refusing the replacement character is not expressed',
    text: doc({ type: 'string', const: '\ufffd' }) },
  { name: 'enum value with a newline', expect: 'looser', reason: 'refusing control characters is not expressed',
    text: doc({ type: 'string', enum: ['a\n'] }) },
  { name: 'schema nested to depth 33', expect: 'looser', reason: 'the schema depth limit of 32 is not expressed',
    text: doc(nest(33)) },
  { name: 'more than 512 schema nodes', expect: 'looser', reason: 'the node count limit of 512 is not expressed',
    text: doc(obj(...['a', 'b', 'c', 'd'].map(n => prop(n, obj(...names(128).map(p => prop(p, str))))))) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: '{"profile":"json-boolean/v1","root":{"type":"string"},"root":{"type":"boolean"}}' },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data',
    text: doc(str) + '{}', parseText: doc(str) },
  { name: 'input over 1 MiB', expect: 'looser', reason: 'the size limit applies to source bytes',
    text: doc(str).replace('}}', '}' + ' '.repeat(1 << 20) + '}') },
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
test('schema agrees with jsonvalidator validate', { skip: hasGo ? false : 'go is not available; agreement with the generator was not checked' }, () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'jsonboolean-schema-'));
  try {
    const bin = path.join(dir, 'jsonvalidator');
    const build = spawnSync('go', ['build', '-o', bin, './cmd/jsonvalidator'], { cwd: path.join(repo, 'modeling/jsonvalidator'), encoding: 'utf8', env: { ...process.env, GOFLAGS: '-mod=readonly' } });
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
