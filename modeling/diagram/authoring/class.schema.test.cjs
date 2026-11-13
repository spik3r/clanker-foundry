'use strict';

// Checks the class-diagram/v1 authoring schema against the registry
// examples, against independently written invalid models and against the real
// diagramgen generator. Go validation is authoritative; the schema may only be
// looser than Go for rules it cannot express, and each such case is listed with
// its reason.

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
const entry = registry.profiles.find(p => p.profile === 'class-diagram/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const readExample = name => fs.readFileSync(path.join(repo, 'modeling/diagram/examples', name), 'utf8');
const base = JSON.parse(readExample('orders-class.json'));

const J = JSON.stringify;
const mut = fn => { const m = JSON.parse(J(base)); fn(m); return J(m); };
const head = { formatVersion: 'diagram-json/v1', profileVersion: 'class-diagram/v1', id: 'Gen', title: 'Generated' };
const cls = (id, extra = {}) => ({ id, name: id, kind: 'class', attributes: [], operations: [], ...extra });
const classes = n => Array.from({ length: n }, (_, i) => cls('C' + i));
const rel = (id, kind = 'association', from = 'C0', to = 'C1', extra = {}) => ({ id, kind, from, to, ...extra });
const model = (cs, rs) => J({ ...head, classes: cs, ...(rs === undefined ? {} : { relations: rs }) });
const members = (n, key = 'name') => Array.from({ length: n }, (_, i) => ({ [key]: 'm' + i }));
const semantic = 'Go-only semantic rule: unique ids, unique member names, or references to classes';
const relRule = 'Go-only rule on relations: self-relations, repeated pairs, or inheritance cycles';
const bytes = 'maxLength counts characters, the limit is in bytes';

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'orders example', expect: 'valid', text: readExample('orders-class.json') },
  { name: 'plugins example', expect: 'valid', text: readExample('plugins-class.json') },
  { name: 'no relations member', expect: 'valid', text: mut(m => { delete m.relations; }) },
  { name: 'empty relations list', expect: 'valid', text: mut(m => { m.relations = []; }) },
  { name: 'one class without members', expect: 'valid', text: J({ ...head, classes: [{ id: 'A', name: 'A', kind: 'class' }] }) },
  { name: '40 classes', expect: 'valid', text: model(classes(40)) },
  { name: '80 relations', expect: 'valid', text: model(classes(40), Array.from({ length: 80 }, (_, i) => rel('R' + i, i % 2 ? 'association' : 'dependency', 'C' + (i % 40), 'C' + ((i + 1 + (i >= 40 ? 1 : 0)) % 40)))) },
  { name: '20 attributes and 20 operations', expect: 'valid', text: model([cls('A', { attributes: members(20), operations: members(20) })]) },
  { name: '8 parameters', expect: 'valid', text: model([cls('A', { operations: [{ name: 'f', parameters: Array.from({ length: 8 }, (_, i) => ({ name: 'p' + i, type: 'T' })) }] })]) },
  { name: 'every visibility', expect: 'valid', text: model([cls('A', { attributes: ['public', 'private', 'protected', 'package'].map(v => ({ name: v, visibility: v })) })]) },
  { name: 'every class kind', expect: 'valid', text: model([cls('A', { kind: 'class' }), cls('B', { kind: 'interface' }), cls('C', { kind: 'enum' })]) },
  { name: 'every relation kind', expect: 'valid', text: model(classes(2), ['inheritance', 'realization', 'association', 'directed', 'aggregation', 'composition', 'dependency'].map((k, i) => rel('R' + i, k, 'C0', 'C1'))) },
  { name: 'relation with all three labels', expect: 'valid', text: model(classes(2), [rel('R', 'association', 'C0', 'C1', { label: 'owns', fromLabel: '1', toLabel: '0..*' })]) },
  { name: 'end label of 24 characters', expect: 'valid', text: model(classes(2), [rel('R', 'association', 'C0', 'C1', { toLabel: 'x'.repeat(24) })]) },
  { name: 'label of 120 characters', expect: 'valid', text: mut(m => { m.title = 'x'.repeat(120); }) },

  { name: 'wrong formatVersion', expect: 'invalid', text: mut(m => { m.formatVersion = 'diagram-json/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(m => { m.profileVersion = 'class-diagram/v2'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.theme = 'dark'; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(m => { m.Title = m.title; }) },
  { name: 'missing classes', expect: 'invalid', text: mut(m => { delete m.classes; }) },
  { name: 'no classes', expect: 'invalid', text: model([]) },
  { name: '41 classes', expect: 'invalid', text: model(classes(41)) },
  { name: 'id starts with a digit', expect: 'invalid', text: mut(m => { m.id = '1Orders'; }) },
  { name: 'class id with hyphen', expect: 'invalid', text: mut(m => { m.classes[0].id = 'My-Class'; }) },
  { name: 'id of 49 characters', expect: 'invalid', text: mut(m => { m.id = 'A' + 'b'.repeat(48); }) },
  { name: 'empty title', expect: 'invalid', text: mut(m => { m.title = ''; }) },
  { name: 'blank class name', expect: 'invalid', text: mut(m => { m.classes[0].name = '   '; }) },
  { name: 'label of 121 characters', expect: 'invalid', text: mut(m => { m.title = 'x'.repeat(121); }) },
  { name: 'newline in a class name', expect: 'invalid', text: mut(m => { m.classes[0].name = 'a\nb'; }) },
  { name: 'U+2028 in an attribute name', expect: 'invalid', text: mut(m => { m.classes[0].attributes[0].name = 'a\u2028b'; }) },
  { name: 'unknown class kind', expect: 'invalid', text: mut(m => { m.classes[0].kind = 'record'; }) },
  { name: 'class without kind', expect: 'invalid', text: mut(m => { delete m.classes[0].kind; }) },
  { name: 'class with a colour', expect: 'invalid', text: mut(m => { m.classes[0].color = 'red'; }) },
  { name: 'class with a position', expect: 'invalid', text: mut(m => { m.classes[0].x = 10; }) },
  { name: '21 attributes', expect: 'invalid', text: model([cls('A', { attributes: members(21) })]) },
  { name: '21 operations', expect: 'invalid', text: model([cls('A', { operations: members(21) })]) },
  { name: 'unknown visibility', expect: 'invalid', text: mut(m => { m.classes[0].attributes[0].visibility = 'friend'; }) },
  { name: 'attribute without name', expect: 'invalid', text: mut(m => { delete m.classes[0].attributes[0].name; }) },
  { name: 'empty attribute type', expect: 'invalid', text: mut(m => { m.classes[0].attributes[0].type = ''; }) },
  { name: 'empty returns', expect: 'invalid', text: mut(m => { m.classes[0].operations[0].returns = ''; }) },
  { name: '9 parameters', expect: 'invalid', text: model([cls('A', { operations: [{ name: 'f', parameters: Array.from({ length: 9 }, (_, i) => ({ name: 'p' + i, type: 'T' })) }] })]) },
  { name: 'parameter without type', expect: 'invalid', text: model([cls('A', { operations: [{ name: 'f', parameters: [{ name: 'p' }] }] })]) },
  { name: 'attribute with extra member', expect: 'invalid', text: mut(m => { m.classes[0].attributes[0].static = true; }) },
  { name: '81 relations', expect: 'invalid', text: model(classes(2), Array.from({ length: 81 }, (_, i) => rel('R' + i))) },
  { name: 'unknown relation kind', expect: 'invalid', text: mut(m => { m.relations[0].kind = 'friendship'; }) },
  { name: 'relation without to', expect: 'invalid', text: mut(m => { delete m.relations[0].to; }) },
  { name: 'empty relation label', expect: 'invalid', text: mut(m => { m.relations[0].label = ''; }) },
  { name: 'end label of 25 characters', expect: 'invalid', text: model(classes(2), [rel('R', 'association', 'C0', 'C1', { fromLabel: 'x'.repeat(25) })]) },
  { name: 'relations as an object', expect: 'invalid', text: mut(m => { m.relations = { R1: m.relations[0] }; }) },
  { name: 'null label', expect: 'invalid', text: mut(m => { m.relations[0].label = null; }) },

  { name: 'duplicate class id', expect: 'looser', reason: semantic, text: mut(m => { m.classes[1].id = m.classes[0].id; }) },
  { name: 'duplicate relation id', expect: 'looser', reason: semantic, text: mut(m => { m.relations[1].id = 'R1'; }) },
  { name: 'relation id equal to a class id', expect: 'looser', reason: semantic, text: mut(m => { m.relations[0].id = 'Order'; }) },
  { name: 'duplicate attribute name', expect: 'looser', reason: semantic, text: mut(m => { m.classes[0].attributes.push({ ...m.classes[0].attributes[0] }); }) },
  { name: 'duplicate operation name', expect: 'looser', reason: semantic, text: model([cls('A', { operations: [{ name: 'f' }, { name: 'f' }] })]) },
  { name: 'duplicate parameter name', expect: 'looser', reason: semantic, text: model([cls('A', { operations: [{ name: 'f', parameters: [{ name: 'p', type: 'T' }, { name: 'p', type: 'U' }] }] })]) },
  { name: 'relation from an unknown class', expect: 'looser', reason: semantic, text: mut(m => { m.relations[0].from = 'Nobody'; }) },
  { name: 'relation to an unknown class', expect: 'looser', reason: semantic, text: mut(m => { m.relations[0].to = 'Nobody'; }) },
  { name: 'self relation', expect: 'looser', reason: relRule, text: model(classes(2), [rel('R', 'association', 'C0', 'C0')]) },
  { name: 'repeated pair and kind', expect: 'looser', reason: relRule, text: model(classes(2), [rel('R1', 'association', 'C0', 'C1'), rel('R2', 'association', 'C1', 'C0')]) },
  { name: 'inheritance cycle', expect: 'looser', reason: relRule, text: model(classes(3), [rel('R1', 'inheritance', 'C0', 'C1'), rel('R2', 'inheritance', 'C1', 'C2'), rel('R3', 'realization', 'C2', 'C0')]) },
  { name: 'multibyte label over 120 bytes', expect: 'looser', reason: bytes, text: mut(m => { m.title = 'é'.repeat(61); }) },
  { name: 'multibyte end label over 24 bytes', expect: 'looser', reason: bytes, text: model(classes(2), [rel('R', 'association', 'C0', 'C1', { label: 'é'.repeat(13) })]) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: J(base).replace('"title":"Order model",', '"title":"Order model","title":"Again",') },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data', text: J(base) + '{}', parseText: J(base) },
  { name: 'input over 512 KiB', expect: 'looser', reason: 'the size limit applies to source bytes',
    text: J(base).replace(/\}$/, ' '.repeat(524289) + '}') },
];

const parsed = c => JSON.parse(c.parseText || c.text);

test('registry examples satisfy the schema', () => {
  assert.ok(entry.examples.length >= 2);
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
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'sequence-schema-'));
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
