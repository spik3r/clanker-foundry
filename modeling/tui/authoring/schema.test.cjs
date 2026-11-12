'use strict';

// Checks the record-browser/v1 authoring schema against the registry examples,
// against independently written invalid models and against the real tuic
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
const entry = registry.profiles.find(p => p.profile === 'record-browser/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const base = JSON.parse(fs.readFileSync(path.join(repo, 'modeling/tui/examples/browser.json'), 'utf8'));

const clone = o => JSON.parse(JSON.stringify(o));
const mut = fn => { const m = clone(base); fn(m); return JSON.stringify(m); };
const fieldIDs = n => Array.from({ length: n }, (_, i) => 'F' + String.fromCharCode(65 + i));
const manyFields = n => ids => ids.map(id => ({ id, label: id }));
const sixteen = fieldIDs(16);
const reserved = ['string', 'len', 'nil', 'main', 'func', 'init', 'int'];

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'example browser', expect: 'valid', text: JSON.stringify(base) },
  { name: 'package-mode example', expect: 'valid', text: fs.readFileSync(path.join(repo, 'modeling/tui/examples/browser-package.json'), 'utf8') },
  { name: 'detailText wrap', expect: 'valid', text: mut(m => { m.detailText = 'wrap'; }) },
  { name: 'detailText clip', expect: 'valid', text: mut(m => { m.detailText = 'clip'; }) },
  { name: 'explicit module emit', expect: 'valid', text: mut(m => { m.emit = { mode: 'module' }; }) },
  { name: 'package emit', expect: 'valid', text: mut(m => { m.emit = { mode: 'package', package: 'browser' }; }) },
  { name: 'package name of 32 characters', expect: 'valid', text: mut(m => { m.emit = { mode: 'package', package: 'a' + 'b'.repeat(31) }; }) },
  { name: 'package name of 1 character', expect: 'valid', text: mut(m => { m.emit = { mode: 'package', package: 'x' }; }) },
  { name: '16 fields and 16 details', expect: 'valid', text: mut(m => { m.fields = manyFields(16)(sixteen); m.identity = 'FA'; m.summary = 'FB'; m.details = sixteen; }) },
  { name: 'minimum layout', expect: 'valid', text: mut(m => { m.layout = { narrowWidth: 24, narrowHeight: 6, wideWidth: 24, wideHeight: 6, listWidth: 8 }; }) },
  { name: 'maximum layout', expect: 'valid', text: mut(m => { m.layout = { narrowWidth: 240, narrowHeight: 80, wideWidth: 400, wideHeight: 120, listWidth: 391 }; }) },
  { name: 'id of 48 characters', expect: 'valid', text: mut(m => { m.id = 'A' + 'b'.repeat(47); }) },
  { name: 'label of 120 characters', expect: 'valid', text: mut(m => { m.title = 'x'.repeat(120); }) },
  { name: 'blank label is allowed', expect: 'valid', text: mut(m => { m.title = '   '; }) },
  { name: 'multibyte label within 120 bytes', expect: 'valid', text: mut(m => { m.title = 'é'.repeat(60); }) },

  { name: 'wrong formatVersion', expect: 'invalid', text: mut(m => { m.formatVersion = 'tui-json/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(m => { m.profileVersion = 'record-browser/v2'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.extra = 1; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(m => { m.Title = m.title; }) },
  { name: 'missing layout', expect: 'invalid', text: mut(m => { delete m.layout; }) },
  { name: 'missing details', expect: 'invalid', text: mut(m => { delete m.details; }) },
  { name: 'id starts lowercase', expect: 'invalid', text: mut(m => { m.id = 'catalog'; }) },
  { name: 'id of 49 characters', expect: 'invalid', text: mut(m => { m.id = 'A' + 'b'.repeat(48); }) },
  { name: 'id with underscore', expect: 'invalid', text: mut(m => { m.id = 'My_Browser'; }) },
  { name: 'empty title', expect: 'invalid', text: mut(m => { m.title = ''; }) },
  { name: 'label of 121 characters', expect: 'invalid', text: mut(m => { m.listLabel = 'x'.repeat(121); }) },
  { name: 'newline in label', expect: 'invalid', text: mut(m => { m.detailLabel = 'a\nb'; }) },
  { name: 'U+2028 in label', expect: 'invalid', text: mut(m => { m.title = 'a b'; }) },
  { name: 'U+0085 in field label', expect: 'invalid', text: mut(m => { m.fields[0].label = 'a\u0085b'; }) },
  { name: 'no fields', expect: 'invalid', text: mut(m => { m.fields = []; m.details = []; }) },
  { name: '17 fields', expect: 'invalid', text: mut(m => { m.fields = manyFields(17)(fieldIDs(17)); m.identity = 'FA'; m.summary = 'FB'; m.details = ['FA']; }) },
  { name: 'no details', expect: 'invalid', text: mut(m => { m.details = []; }) },
  { name: '17 details', expect: 'invalid', text: mut(m => { m.fields = manyFields(16)(sixteen); m.identity = 'FA'; m.summary = 'FB'; m.details = sixteen.concat(['FA']); }) },
  { name: 'duplicate detail', expect: 'invalid', text: mut(m => { m.details = ['Name', 'Name']; }) },
  { name: 'field with unknown member', expect: 'invalid', text: mut(m => { m.fields[0].type = 'int'; }) },
  { name: 'field id lowercase', expect: 'invalid', text: mut(m => { m.fields[1].id = 'name'; }) },
  { name: 'layout with unknown member', expect: 'invalid', text: mut(m => { m.layout.gap = 1; }) },
  { name: 'narrowWidth 23', expect: 'invalid', text: mut(m => { m.layout.narrowWidth = 23; }) },
  { name: 'narrowWidth 241', expect: 'invalid', text: mut(m => { m.layout.narrowWidth = 241; }) },
  { name: 'narrowHeight 5', expect: 'invalid', text: mut(m => { m.layout.narrowHeight = 5; }) },
  { name: 'narrowHeight 81', expect: 'invalid', text: mut(m => { m.layout.narrowHeight = 81; }) },
  { name: 'wideWidth 401', expect: 'invalid', text: mut(m => { m.layout.wideWidth = 401; }) },
  { name: 'wideHeight 121', expect: 'invalid', text: mut(m => { m.layout.wideHeight = 121; }) },
  { name: 'listWidth 7', expect: 'invalid', text: mut(m => { m.layout.listWidth = 7; }) },
  { name: 'listWidth 392', expect: 'invalid', text: mut(m => { m.layout.listWidth = 392; }) },
  { name: 'layout number as string', expect: 'invalid', text: mut(m => { m.layout.wideWidth = '80'; }) },
  { name: 'layout number with fraction', expect: 'invalid', text: mut(m => { m.layout.wideWidth = 80.5; }) },
  { name: 'null label', expect: 'invalid', text: mut(m => { m.title = null; }) },
  { name: 'null emit', expect: 'invalid', text: mut(m => { m.emit = null; }) },
  { name: 'empty detailText', expect: 'invalid', text: mut(m => { m.detailText = ''; }) },
  { name: 'unknown detailText', expect: 'invalid', text: mut(m => { m.detailText = 'Wrap'; }) },
  { name: 'empty emit', expect: 'invalid', text: mut(m => { m.emit = {}; }) },
  { name: 'unknown emit mode', expect: 'invalid', text: mut(m => { m.emit = { mode: 'Module' }; }) },
  { name: 'package mode without package', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package' }; }) },
  { name: 'module mode with package', expect: 'invalid', text: mut(m => { m.emit = { mode: 'module', package: 'browser' }; }) },
  { name: 'module mode with empty package', expect: 'invalid', text: mut(m => { m.emit = { mode: 'module', package: '' }; }) },
  { name: 'package name of 33 characters', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: 'a' + 'b'.repeat(32) }; }) },
  { name: 'package name with uppercase', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: 'Browser' }; }) },
  { name: 'package name starting with a digit', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: '1abc' }; }) },
  { name: 'package name with underscore', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: 'my_pkg' }; }) },
  { name: 'empty package name', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: '' }; }) },
  { name: 'emit with import path', expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: 'browser', path: 'example.com/x' }; }) },
  ...reserved.map(word => ({ name: `reserved package name ${word}`, expect: 'invalid', text: mut(m => { m.emit = { mode: 'package', package: word }; }) })),

  { name: 'duplicate field ids', expect: 'looser', reason: 'unique field ids need a keyed uniqueness check JSON Schema lacks',
    text: mut(m => { m.fields[1].id = 'ID'; m.details = ['Name']; }) },
  { name: 'identity names no field', expect: 'looser', reason: 'references to declared fields are not expressible',
    text: mut(m => { m.identity = 'Missing'; }) },
  { name: 'summary names no field', expect: 'looser', reason: 'references to declared fields are not expressible',
    text: mut(m => { m.summary = 'Missing'; }) },
  { name: 'detail names no field', expect: 'looser', reason: 'references to declared fields are not expressible',
    text: mut(m => { m.details = ['Name', 'Missing']; }) },
  { name: 'wide width below narrow width', expect: 'looser', reason: 'comparison between two numbers is not expressible',
    text: mut(m => { m.layout.narrowWidth = 100; m.layout.wideWidth = 90; m.layout.listWidth = 10; }) },
  { name: 'wide height below narrow height', expect: 'looser', reason: 'comparison between two numbers is not expressible',
    text: mut(m => { m.layout.narrowHeight = 20; m.layout.wideHeight = 10; }) },
  { name: 'list width leaves under 8 columns', expect: 'looser', reason: 'listWidth <= wideWidth - 9 is not expressible',
    text: mut(m => { m.layout.listWidth = 72; }) },
  { name: 'integer written as 24.0', expect: 'looser', reason: 'JSON parsers give 24 for 24.0; Go refuses the float lexeme',
    text: JSON.stringify(base).replace('"narrowWidth":24,', '"narrowWidth":24.0,') },
  { name: 'integer written with exponent', expect: 'looser', reason: 'JSON parsers give 24 for 2.4e1; Go refuses the float lexeme',
    text: JSON.stringify(base).replace('"narrowWidth":24,', '"narrowWidth":2.4e1,') },
  { name: 'multibyte label over 120 bytes', expect: 'looser', reason: 'maxLength counts characters, the limit is in bytes',
    text: mut(m => { m.title = 'é'.repeat(61); }) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: JSON.stringify(base).replace('"title":"Catalog",', '"title":"Catalog","title":"Again",') },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data',
    text: JSON.stringify(base) + '{}' , parseText: JSON.stringify(base) },
  { name: 'input over 32 KiB', expect: 'looser', reason: 'the size limit applies to source bytes',
    text: mut(m => { m.fields = manyFields(16)(sixteen).map(f => ({ id: f.id, label: 'x'.repeat(120) })); m.identity = 'FA'; m.summary = 'FB'; m.details = sixteen; m.padding = 0; }).replace(/,"padding":0/, ' '.repeat(33000)) },
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
test('schema agrees with tuic validate', { skip: hasGo ? false : 'go is not available; agreement with the generator was not checked' }, () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'tui-schema-'));
  try {
    const bin = path.join(dir, 'tuic');
    const build = spawnSync('go', ['build', '-o', bin, './cmd/tuic'], { cwd: path.join(repo, 'modeling/tui'), encoding: 'utf8', env: { ...process.env, GOFLAGS: '-mod=readonly' } });
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
