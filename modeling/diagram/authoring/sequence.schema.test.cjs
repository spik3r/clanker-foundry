'use strict';

// Checks the sequence-diagram/v1 authoring schema against the registry
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
const entry = registry.profiles.find(p => p.profile === 'sequence-diagram/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const readExample = name => fs.readFileSync(path.join(repo, 'modeling/diagram/examples', name), 'utf8');
const base = JSON.parse(readExample('checkout-sequence.json'));

const J = JSON.stringify;
const mut = fn => { const m = JSON.parse(J(base)); fn(m); return J(m); };
const head = { formatVersion: 'diagram-json/v1', profileVersion: 'sequence-diagram/v1', id: 'Gen', title: 'Generated' };
const part = (id, kind = 'participant') => ({ id, label: id, kind });
const parts = n => Array.from({ length: n }, (_, i) => part('P' + i));
const msg = (id, from = 'P0', to = 'P1', style = 'sync') => ({ id, kind: 'message', from, to, label: 'm', style });
const note = (id, over) => ({ id, kind: 'note', over, label: 'n' });
const model = (participants, steps, groups) => J({ ...head, participants, steps, ...(groups === undefined ? {} : { groups }) });
const steps = n => Array.from({ length: n }, (_, i) => msg('S' + i));
const group = (id, first, last, kind = 'opt') => ({ id, kind, label: 'g', first, last });
const semantic = 'Go-only semantic rule: unique ids or references to participants and steps';
const groupRule = 'Go-only rule on group ranges: order of first and last, and overlap between groups';
const bytes = 'maxLength counts characters, the limit is in bytes';

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'checkout example', expect: 'valid', text: readExample('checkout-sequence.json') },
  { name: 'retry example', expect: 'valid', text: readExample('retry-sequence.json') },
  { name: 'no groups member', expect: 'valid', text: mut(m => { delete m.groups; }) },
  { name: 'empty groups list', expect: 'valid', text: mut(m => { m.groups = []; }) },
  { name: 'two participants one message', expect: 'valid', text: model(parts(2), [msg('S')]) },
  { name: '12 participants', expect: 'valid', text: model(parts(12), [msg('S')]) },
  { name: '80 steps', expect: 'valid', text: model(parts(2), steps(80)) },
  { name: 'every message style', expect: 'valid', text: model(parts(2), [msg('A', 'P0', 'P1', 'sync'), msg('B', 'P1', 'P0', 'async'), msg('C', 'P1', 'P0', 'reply')]) },
  { name: 'self message', expect: 'valid', text: model(parts(2), [msg('S', 'P0', 'P0')]) },
  { name: 'note over two participants', expect: 'valid', text: model(parts(2), [note('N', ['P0', 'P1'])]) },
  { name: '8 groups', expect: 'valid', text: model(parts(2), steps(8), Array.from({ length: 8 }, (_, i) => group('G' + i, 'S' + i, 'S' + i))) },
  { name: 'every group kind', expect: 'valid', text: model(parts(2), steps(3), [group('A', 'S0', 'S0', 'alt'), group('B', 'S1', 'S1', 'opt'), group('C', 'S2', 'S2', 'loop')]) },
  { name: 'label of 120 characters', expect: 'valid', text: mut(m => { m.title = 'x'.repeat(120); }) },

  { name: 'wrong formatVersion', expect: 'invalid', text: mut(m => { m.formatVersion = 'diagram-json/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(m => { m.profileVersion = 'sequence-diagram/v2'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.theme = 'dark'; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(m => { m.Title = m.title; }) },
  { name: 'missing steps', expect: 'invalid', text: mut(m => { delete m.steps; }) },
  { name: 'missing participants', expect: 'invalid', text: mut(m => { delete m.participants; }) },
  { name: 'id starts with a digit', expect: 'invalid', text: mut(m => { m.id = '1Checkout'; }) },
  { name: 'id with hyphen', expect: 'invalid', text: mut(m => { m.steps[0].id = 'S-1'; }) },
  { name: 'id of 49 characters', expect: 'invalid', text: mut(m => { m.id = 'A' + 'b'.repeat(48); }) },
  { name: 'empty title', expect: 'invalid', text: mut(m => { m.title = ''; }) },
  { name: 'blank title', expect: 'invalid', text: mut(m => { m.title = '   '; }) },
  { name: 'label of 121 characters', expect: 'invalid', text: mut(m => { m.title = 'x'.repeat(121); }) },
  { name: 'newline in a label', expect: 'invalid', text: mut(m => { m.steps[0].label = 'a\nb'; }) },
  { name: 'U+2028 in a label', expect: 'invalid', text: mut(m => { m.participants[0].label = 'a b'; }) },
  { name: 'one participant', expect: 'invalid', text: model(parts(1), [msg('S', 'P0', 'P0')]) },
  { name: '13 participants', expect: 'invalid', text: model(parts(13), [msg('S')]) },
  { name: 'unknown participant kind', expect: 'invalid', text: mut(m => { m.participants[0].kind = 'database'; }) },
  { name: 'participant without label', expect: 'invalid', text: mut(m => { delete m.participants[0].label; }) },
  { name: 'participant with a colour', expect: 'invalid', text: mut(m => { m.participants[0].color = 'red'; }) },
  { name: 'no steps', expect: 'invalid', text: model(parts(2), []) },
  { name: '81 steps', expect: 'invalid', text: model(parts(2), steps(81)) },
  { name: 'unknown step kind', expect: 'invalid', text: mut(m => { m.steps[0].kind = 'fork'; }) },
  { name: 'unknown message style', expect: 'invalid', text: mut(m => { m.steps[0].style = 'loud'; }) },
  { name: 'message without style', expect: 'invalid', text: mut(m => { delete m.steps[0].style; }) },
  { name: 'message without from', expect: 'invalid', text: mut(m => { delete m.steps[0].from; }) },
  { name: 'message with over', expect: 'invalid', text: mut(m => { m.steps[0].over = ['Web']; }) },
  { name: 'message with an x coordinate', expect: 'invalid', text: mut(m => { m.steps[0].x = 10; }) },
  { name: 'note without over', expect: 'invalid', text: mut(m => { delete m.steps[3].over; }) },
  { name: 'note with empty over', expect: 'invalid', text: mut(m => { m.steps[3].over = []; }) },
  { name: 'note over three participants', expect: 'invalid', text: mut(m => { m.steps[3].over = ['User', 'Web', 'Pay']; }) },
  { name: 'note over the same participant twice', expect: 'invalid', text: mut(m => { m.steps[3].over = ['Web', 'Web']; }) },
  { name: 'note with from', expect: 'invalid', text: mut(m => { m.steps[3].from = 'Web'; }) },
  { name: 'note with style', expect: 'invalid', text: mut(m => { m.steps[3].style = 'sync'; }) },
  { name: 'note with empty label', expect: 'invalid', text: mut(m => { m.steps[3].label = ''; }) },
  { name: '9 groups', expect: 'invalid', text: model(parts(2), steps(9), Array.from({ length: 9 }, (_, i) => group('G' + i, 'S' + i, 'S' + i))) },
  { name: 'unknown group kind', expect: 'invalid', text: mut(m => { m.groups[0].kind = 'par'; }) },
  { name: 'group without last', expect: 'invalid', text: mut(m => { delete m.groups[0].last; }) },
  { name: 'group with empty label', expect: 'invalid', text: mut(m => { m.groups[0].label = ''; }) },
  { name: 'groups as an object', expect: 'invalid', text: mut(m => { m.groups = { G1: m.groups[0] }; }) },
  { name: 'null label', expect: 'invalid', text: mut(m => { m.steps[0].label = null; }) },

  { name: 'duplicate participant id', expect: 'looser', reason: semantic, text: mut(m => { m.participants[1].id = 'User'; }) },
  { name: 'duplicate step id', expect: 'looser', reason: semantic, text: mut(m => { m.steps[1].id = 'S1'; }) },
  { name: 'step id equal to a participant id', expect: 'looser', reason: semantic, text: mut(m => { m.steps[0].id = 'Web'; }) },
  { name: 'group id equal to a step id', expect: 'looser', reason: semantic, text: mut(m => { m.groups[0].id = 'S1'; }) },
  { name: 'message from an unknown participant', expect: 'looser', reason: semantic, text: mut(m => { m.steps[0].from = 'Nobody'; }) },
  { name: 'message to an unknown participant', expect: 'looser', reason: semantic, text: mut(m => { m.steps[0].to = 'Nobody'; }) },
  { name: 'note over an unknown participant', expect: 'looser', reason: semantic, text: mut(m => { m.steps[3].over = ['Nobody']; }) },
  { name: 'group over an unknown step', expect: 'looser', reason: semantic, text: mut(m => { m.groups[0].first = 'Nope'; }) },
  { name: 'group with first after last', expect: 'looser', reason: groupRule, text: mut(m => { m.groups[0].first = 'S3'; m.groups[0].last = 'S2'; }) },
  { name: 'overlapping groups', expect: 'looser', reason: groupRule, text: mut(m => { m.groups.push({ id: 'G2', kind: 'opt', label: 'g', first: 'S3', last: 'S4' }); }) },
  { name: 'nested groups', expect: 'looser', reason: groupRule, text: mut(m => { m.groups.push({ id: 'G2', kind: 'loop', label: 'g', first: 'S1', last: 'S6' }); }) },
  { name: 'multibyte label over 120 bytes', expect: 'looser', reason: bytes, text: mut(m => { m.title = 'é'.repeat(61); }) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: J(base).replace('"title":"Checkout",', '"title":"Checkout","title":"Again",') },
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
