'use strict';

// Checks the record-browser-cases/v1 authoring schema against the registry
// examples, against independently written invalid documents and against the
// real tuicases validate command. Go validation is authoritative; the schema may
// only be looser than Go for rules it cannot express, and each such case is
// listed with its reason.

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
assert.ok(entry.cases, 'registry entry has no cases member');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.cases.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const modelPath = path.join(repo, 'modeling/tui/examples/browser.json');
const baseText = fs.readFileSync(path.join(repo, 'modeling/tui/examples/browser.cases.json'), 'utf8');
const base = JSON.parse(baseText);

const clone = o => JSON.parse(JSON.stringify(o));
const mut = fn => { const d = clone(base); fn(d); return JSON.stringify(d); };
const firstCase = d => d.cases[0];
const oneCase = (d, fn) => { d.cases = [clone(d.cases[0])]; fn(d.cases[0], d); };
const rec = (id) => ({ ID: id, Name: id, Description: 'd', Status: 's' });
const keys = ['tab', 'shift+tab', 'enter', 'esc', '?', 'q', 'ctrl+c', 'j', 'k', 'up', 'down', 'pgup', 'pgdown', 'home', 'end'];

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'browser example', expect: 'valid', text: baseText },
  { name: 'package example is for another model but valid for the schema', expect: 'looser', reason: 'binding to the model digest is not expressible',
    text: fs.readFileSync(path.join(repo, 'modeling/tui/examples/browser-package.cases.json'), 'utf8') },
  { name: 'hosted package example is for another model but valid for the schema', expect: 'looser', reason: 'binding to the model digest is not expressible',
    text: fs.readFileSync(path.join(repo, 'modeling/tui/examples/browser-hosted.cases.json'), 'utf8') },
  { name: 'capturesKeys expectation', expect: 'valid', text: mut(d => oneCase(d, c => { c.expect = { capturesKeys: true }; })) },
  { name: 'keyHints expectation', expect: 'valid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ key: 'j/k', action: 'select' }, { key: '?', action: 'help' }] }; })) },
  { name: 'keyHints by id', expect: 'valid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ id: 'select' }, { id: 'help.open', key: '?' }, { id: 'quit', key: 'q', action: 'quit' }] }; })) },
  { name: '12 keyHints', expect: 'valid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: Array.from({ length: 12 }, () => ({ key: '?', action: 'help' })) }; })) },
  { name: 'every documented key', expect: 'valid', text: mut(d => oneCase(d, c => { c.steps = keys.map(key => ({ key })); })) },
  { name: 'resize step', expect: 'valid', text: mut(d => oneCase(d, c => { c.steps = [{ resize: { width: 400, height: 200 } }, { resize: { width: 1, height: 1 } }]; })) },
  { name: 'replace steps', expect: 'valid', text: mut(d => oneCase(d, c => { c.steps = [{ replace: { status: 'loading', message: 'Fetching' } }, { replace: { status: 'failed' } }, { replace: { status: 'ready', records: 'Main' } }, { replace: { status: 'ready' } }]; })) },
  { name: 'no steps', expect: 'valid', text: mut(d => oneCase(d, c => { c.steps = []; })) },
  { name: 'empty record set', expect: 'valid', text: mut(d => { d.records.Empty = []; }) },
  { name: 'expectation of every kind', expect: 'valid', text: mut(d => oneCase(d, c => { c.expect = { layout: 'tiny', focus: 'detail', help: true, selected: '', status: 'loading', visible: ['x'], absent: ['y'], quit: false }; })) },
  { name: '16 visible strings', expect: 'valid', text: mut(d => oneCase(d, c => { c.expect = { visible: Array.from({ length: 16 }, (_, i) => 's' + i) }; })) },
  { name: '64 steps', expect: 'valid', text: mut(d => oneCase(d, c => { c.steps = Array.from({ length: 64 }, () => ({ key: 'j' })); })) },
  { name: '8 record sets', expect: 'valid', text: mut(d => { for (let i = 0; i < 4; i++) d.records['S' + i] = [rec('x' + i)]; }) },
  { name: '200 records in a set', expect: 'valid', text: mut(d => { d.records.Main = Array.from({ length: 200 }, (_, i) => rec('id' + i)); }) },
  { name: 'case id of 64 characters', expect: 'valid', text: mut(d => oneCase(d, c => { c.id = 'C' + 'a'.repeat(63); })) },
  { name: 'requirement of 240 characters', expect: 'valid', text: mut(d => oneCase(d, c => { c.requirement = 'x'.repeat(240); })) },

  { name: 'capturesKeys not a boolean', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { capturesKeys: 'yes' }; })) },
  { name: 'keyHints entry without id, key or action', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{}] }; })) },
  { name: 'keyHints entry with an unknown id', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ id: 'nope' }] }; })) },
  { name: 'keyHints entry with an empty id', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ id: '' }] }; })) },
  { name: 'keyHints id with an empty key', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ id: 'quit', key: '' }] }; })) },
  { name: 'empty keyHints', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [] }; })) },
  { name: '13 keyHints', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: Array.from({ length: 13 }, () => ({ key: '?', action: 'help' })) }; })) },
  { name: 'keyHints entry as string', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: ['?'] }; })) },
  { name: 'keyHints entry without action', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ key: '?' }] }; })) },
  { name: 'keyHints entry with empty key', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ key: '', action: 'help' }] }; })) },
  { name: 'keyHints entry with unknown member', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ key: '?', action: 'help', note: 'x' }] }; })) },
  { name: 'keyHints action with a control character', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { keyHints: [{ key: '?', action: 'a\tb' }] }; })) },
  { name: 'wrong formatVersion', expect: 'invalid', text: mut(d => { d.formatVersion = 'record-browser-cases/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(d => { d.profileVersion = 'record-browser/v2'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(d => { d.extra = 1; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(d => { d.Cases = []; }) },
  { name: 'missing model', expect: 'invalid', text: mut(d => { delete d.model; }) },
  { name: 'short digest', expect: 'invalid', text: mut(d => { d.model.canonicalSHA256 = 'abcd'; }) },
  { name: 'uppercase digest', expect: 'invalid', text: mut(d => { d.model.canonicalSHA256 = d.model.canonicalSHA256.toUpperCase(); }) },
  { name: 'no record sets', expect: 'invalid', text: mut(d => { d.records = {}; }) },
  { name: 'nine record sets', expect: 'invalid', text: mut(d => { for (let i = 0; i < 5; i++) d.records['S' + i] = []; }) },
  { name: 'record set name with a digit first', expect: 'invalid', text: mut(d => { d.records['1Main'] = []; }) },
  { name: '201 records in a set', expect: 'invalid', text: mut(d => { d.records.Main = Array.from({ length: 201 }, (_, i) => rec('id' + i)); }) },
  { name: 'record value not a string', expect: 'invalid', text: mut(d => { d.records.Main[0].Name = 5; }) },
  { name: 'record value null', expect: 'invalid', text: mut(d => { d.records.Main[0].Name = null; }) },
  { name: 'no cases', expect: 'invalid', text: mut(d => { d.cases = []; }) },
  { name: '201 cases', expect: 'invalid', text: mut(d => { d.cases = Array.from({ length: 201 }, (_, i) => Object.assign(clone(d.cases[0]), { id: 'C' + i })); }) },
  { name: 'case id with space', expect: 'invalid', text: mut(d => oneCase(d, c => { c.id = 'C one'; })) },
  { name: 'case id of 65 characters', expect: 'invalid', text: mut(d => oneCase(d, c => { c.id = 'C' + 'a'.repeat(64); })) },
  { name: 'empty requirement', expect: 'invalid', text: mut(d => oneCase(d, c => { c.requirement = ''; })) },
  { name: 'requirement of 241 characters', expect: 'invalid', text: mut(d => oneCase(d, c => { c.requirement = 'x'.repeat(241); })) },
  { name: 'newline in requirement', expect: 'invalid', text: mut(d => oneCase(d, c => { c.requirement = 'a\nb'; })) },
  { name: 'unknown member in a case', expect: 'invalid', text: mut(d => oneCase(d, c => { c.note = 'x'; })) },
  { name: 'size of zero', expect: 'invalid', text: mut(d => oneCase(d, c => { c.size.width = 0; })) },
  { name: 'size too high', expect: 'invalid', text: mut(d => oneCase(d, c => { c.size.height = 201; })) },
  { name: 'size as fraction', expect: 'invalid', text: mut(d => oneCase(d, c => { c.size.width = 80.5; })) },
  { name: 'steps missing', expect: 'invalid', text: mut(d => oneCase(d, c => { delete c.steps; })) },
  { name: '65 steps', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = Array.from({ length: 65 }, () => ({ key: 'j' })); })) },
  { name: 'empty step', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{}]; })) },
  { name: 'step with two members', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ key: 'j', resize: { width: 30, height: 8 } }]; })) },
  { name: 'key outside the documented set', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ key: 'x' }]; })) },
  { name: 'undocumented alias key', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ key: 'G' }]; })) },
  { name: 'key with different case', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ key: 'Enter' }]; })) },
  { name: 'resize without height', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ resize: { width: 30 } }]; })) },
  { name: 'replace with unknown status', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ replace: { status: 'done' } }]; })) },
  { name: 'replace records with loading', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ replace: { status: 'loading', records: 'Main' } }]; })) },
  { name: 'replace with empty message', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ replace: { status: 'failed', message: '' } }]; })) },
  { name: 'replace with unknown member', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ replace: { status: 'failed', note: 'x' } }]; })) },
  { name: 'empty expect', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = {}; })) },
  { name: 'unknown layout', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { layout: 'huge' }; })) },
  { name: 'unknown focus', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { focus: 'both' }; })) },
  { name: 'help not a boolean', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { help: 'yes' }; })) },
  { name: 'empty visible list', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { visible: [] }; })) },
  { name: '17 absent strings', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { absent: Array.from({ length: 17 }, (_, i) => 's' + i) }; })) },
  { name: 'empty visible string', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { visible: [''] }; })) },
  { name: 'tab in an absent string', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { absent: ['a\tb'] }; })) },
  { name: 'null expectation', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { selected: null }; })) },
  { name: 'unknown expectation', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { selection: 'a' }; })) },

  { name: 'text step needs a filter', expect: 'looser', reason: 'whether the model has a filter is not visible to the schema',
    text: mut(d => oneCase(d, c => { c.steps = [{ key: '/' }, { text: 'abc' }, { key: 'backspace' }]; })) },
  { name: 'text step of 64 code points', expect: 'looser', reason: 'whether the model has a filter is not visible to the schema',
    text: mut(d => oneCase(d, c => { c.steps = [{ text: 'é'.repeat(64) }]; })) },
  { name: 'filter expectations need a filter', expect: 'looser', reason: 'whether the model has a filter is not visible to the schema',
    text: mut(d => oneCase(d, c => { c.expect = { filter: '', filterOpen: false, shown: 4096 }; })) },
  { name: 'empty text step', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ text: '' }]; })) },
  { name: 'text step of 65 code points', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ text: 'é'.repeat(65) }]; })) },
  { name: 'newline in a text step', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ text: 'a\nb' }]; })) },
  { name: 'step with key and text', expect: 'invalid', text: mut(d => oneCase(d, c => { c.steps = [{ key: 'j', text: 'a' }]; })) },
  { name: 'filter expectation of 65 code points', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { filter: 'x'.repeat(65) }; })) },
  { name: 'control character in a filter expectation', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { filter: 'a\u0007' }; })) },
  { name: 'shown above 4096', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { shown: 4097 }; })) },
  { name: 'shown negative', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { shown: -1 }; })) },
  { name: 'filterOpen not a boolean', expect: 'invalid', text: mut(d => oneCase(d, c => { c.expect = { filterOpen: 'yes' }; })) },
  { name: 'filter example is for another model but valid for the schema', expect: 'looser', reason: 'binding to the model digest is not expressible',
    text: fs.readFileSync(path.join(repo, 'modeling/tui/examples/browser-filter.cases.json'), 'utf8') },

  { name: 'stale digest', expect: 'looser', reason: 'binding to the digest of the model is not expressible',
    text: mut(d => { d.model.canonicalSHA256 = '0'.repeat(64); }) },
  { name: 'record missing a field', expect: 'looser', reason: 'the field set comes from the model, which the schema does not see',
    text: mut(d => { delete d.records.Main[0].Name; }) },
  { name: 'record with an extra field', expect: 'looser', reason: 'the field set comes from the model, which the schema does not see',
    text: mut(d => { d.records.Main[0].Extra = 'x'; }) },
  { name: 'duplicate record identity', expect: 'looser', reason: 'unique identities need the identity field of the model',
    text: mut(d => { d.records.Main[1].ID = d.records.Main[0].ID; }) },
  { name: 'empty record identity', expect: 'looser', reason: 'the identity field comes from the model',
    text: mut(d => { d.records.Main[1].ID = ''; }) },
  { name: 'case names an undeclared record set', expect: 'looser', reason: 'references to record sets are not expressible',
    text: mut(d => oneCase(d, c => { c.records = 'Other'; })) },
  { name: 'replace names an undeclared record set', expect: 'looser', reason: 'references to record sets are not expressible',
    text: mut(d => oneCase(d, c => { c.steps = [{ replace: { status: 'ready', records: 'Other' } }]; })) },
  { name: 'duplicate case ids', expect: 'looser', reason: 'unique case ids need a keyed uniqueness check JSON Schema lacks',
    text: mut(d => { d.cases = [d.cases[0], clone(d.cases[0])]; }) },
  { name: 'integer written as 100.0', expect: 'looser', reason: 'JSON parsers give 100 for 100.0; Go refuses the float lexeme',
    text: baseText.replace('"width": 100,', '"width": 100.0,') },
  { name: 'multibyte requirement over 240 bytes', expect: 'looser', reason: 'maxLength counts characters, the limit is in bytes',
    text: mut(d => oneCase(d, c => { c.requirement = 'é'.repeat(121); })) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: baseText.replace('"profileVersion": "record-browser/v1",', '"profileVersion": "record-browser/v1", "profileVersion": "record-browser/v1",') },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data',
    text: baseText + '{}', parseText: baseText },
];

const parsed = c => JSON.parse(c.parseText || c.text);

test('registry cases examples satisfy the schema', () => {
  assert.ok(entry.cases.examples.length >= 1);
  for (const rel of entry.cases.examples) {
    const doc = JSON.parse(fs.readFileSync(path.join(repo, rel), 'utf8'));
    assert.equal(validate(doc), true, `${rel}: ${JSON.stringify(validate.errors)}`);
  }
});

test('schema rejects at least 12 invalid documents', () => {
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
test('schema agrees with tuicases validate', { skip: hasGo ? false : 'go is not available; agreement with the runner was not checked' }, () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'tui-cases-schema-'));
  try {
    const bin = path.join(dir, 'tuicases');
    const build = spawnSync('go', ['build', '-o', bin, './cmd/tuicases'], { cwd: path.join(repo, 'modeling/tui'), encoding: 'utf8', env: { ...process.env, GOFLAGS: '-mod=readonly' } });
    assert.equal(build.status, 0, `go build failed: ${build.stderr}`);
    assert.ok(cases.length >= 15);
    cases.forEach((c, i) => {
      const file = path.join(dir, `cases-${i}.json`);
      fs.writeFileSync(file, c.text);
      const run = spawnSync(bin, ['validate', modelPath, file], { encoding: 'utf8', env: {} });
      assert.ok(run.status === 0 || run.status === 2, `${c.name}: unexpected exit ${run.status}: ${run.stderr}`);
      const goAccepts = run.status === 0;
      const schemaAccepts = validate(parsed(c)) === true;
      assert.ok(!goAccepts || schemaAccepts, `${c.name}: schema rejects a document Go accepts`);
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
