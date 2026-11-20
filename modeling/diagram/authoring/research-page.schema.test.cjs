'use strict';

// Checks the research-page/v1 authoring schema against the registry example,
// against independently written invalid models and against the real diagramgen
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
const entry = registry.profiles.find(p => p.profile === 'research-page/v1');
const schema = JSON.parse(fs.readFileSync(path.join(repo, entry.schema), 'utf8'));
const validate = new Ajv2020({ allErrors: true, strict: true }).compile(schema);
const readExample = name => fs.readFileSync(path.join(repo, 'modeling/diagram/examples', name), 'utf8');
const base = JSON.parse(readExample('evidence-atlas-page.json'));

const J = JSON.stringify;
const mut = fn => { const m = JSON.parse(J(base)); fn(m); return J(m); };
const sec = (m, tab) => m.sections.find(s => s.tab === tab);
// Example items: explore [0] slot coverage, [1] slot controls, [2] heading,
// [3] panel matrix, [4] heading, [5] panel detail; method [0] methodPanel;
// sources [0] heading, [1] panel provenance.
const explore = (m, i) => sec(m, 'explore').items[i];
const method = m => sec(m, 'method').items[0];
const many = (text, n) => Array.from({ length: n }, () => text);
const proseItems = n => Array.from({ length: n }, () => ({ kind: 'prose', paragraphs: ['x'] }));
const semantic = 'Go-only semantic rule: unique tabs, or slot ids unique across slot and panel items';
const bytesOnly = 'maxLength counts characters, the limit is in bytes';

// expect: 'valid' (schema and Go accept), 'invalid' (both refuse) or
// 'looser' (schema accepts, Go refuses; the reason names the rule the schema
// cannot express).
const cases = [
  { name: 'evidence atlas example', expect: 'valid', text: readExample('evidence-atlas-page.json') },
  { name: 'no class', expect: 'valid', text: mut(m => { delete m.class; }) },
  { name: 'no context', expect: 'valid', text: mut(m => { delete m.bottomLine.context; }) },
  { name: 'every status', expect: 'valid', text: mut(m => { m.status = 'not-yet-run'; }) },
  { name: 'status supported', expect: 'valid', text: mut(m => { m.status = 'supported'; }) },
  { name: 'status unknown', expect: 'valid', text: mut(m => { m.status = 'unknown'; }) },
  { name: 'one section only', expect: 'valid', text: mut(m => { m.sections = [sec(m, 'method')]; }) },
  { name: 'sections out of tab order', expect: 'valid', text: mut(m => { m.sections.reverse(); }) },
  { name: 'heading without description', expect: 'valid', text: mut(m => { delete explore(m, 2).description; }) },
  { name: 'methodPanel without description', expect: 'valid', text: mut(m => { delete method(m).description; }) },
  { name: 'prose item', expect: 'valid', text: mut(m => { sec(m, 'method').items.push({ kind: 'prose', paragraphs: ['One.', 'Two.'] }); }) },
  { name: '24 items in a tab', expect: 'valid', text: mut(m => { sec(m, 'method').items = proseItems(24); }) },
  { name: '12 points', expect: 'valid', text: mut(m => { method(m).points = many('x', 12); }) },
  { name: '6 paragraphs', expect: 'valid', text: mut(m => { sec(m, 'method').items = [{ kind: 'prose', paragraphs: many('x', 6) }]; }) },
  { name: '6 context paragraphs', expect: 'valid', text: mut(m => { m.bottomLine.context = many('x', 6); }) },
  { name: 'id of 48 characters', expect: 'valid', text: mut(m => { m.id = 'A' + 'b'.repeat(47); }) },
  { name: 'slot id of 32 characters', expect: 'valid', text: mut(m => { explore(m, 0).id = 'a' + 'B'.repeat(31); }) },
  { name: 'slot id class', expect: 'valid', text: mut(m => { explore(m, 0).id = 'class'; }) },
  { name: 'title of 120 characters', expect: 'valid', text: mut(m => { m.title = 'x'.repeat(120); }) },
  { name: 'purpose of 600 characters', expect: 'valid', text: mut(m => { m.purpose = 'x'.repeat(600); }) },
  { name: 'non-ASCII text', expect: 'valid', text: mut(m => { m.title = 'Évidence — 証拠'; }) },

  { name: 'wrong formatVersion', expect: 'invalid', text: mut(m => { m.formatVersion = 'diagram-json/v2'; }) },
  { name: 'wrong profileVersion', expect: 'invalid', text: mut(m => { m.profileVersion = 'research-page/v2'; }) },
  { name: 'dashboard profile', expect: 'invalid', text: mut(m => { m.profileVersion = 'dashboard/v1'; }) },
  { name: 'unknown top-level member', expect: 'invalid', text: mut(m => { m.theme = 'dark'; }) },
  { name: 'case variant of a member', expect: 'invalid', text: mut(m => { m.Title = m.title; }) },
  { name: 'missing purpose', expect: 'invalid', text: mut(m => { delete m.purpose; }) },
  { name: 'missing bottomLine', expect: 'invalid', text: mut(m => { delete m.bottomLine; }) },
  { name: 'missing sections', expect: 'invalid', text: mut(m => { delete m.sections; }) },
  { name: 'missing limitation', expect: 'invalid', text: mut(m => { delete m.bottomLine.limitation; }) },
  { name: 'unknown bottomLine member', expect: 'invalid', text: mut(m => { m.bottomLine.html = 'x'; }) },
  { name: 'id starts lower-case', expect: 'invalid', text: mut(m => { m.id = 'evidenceAtlas'; }) },
  { name: 'id starts with a digit', expect: 'invalid', text: mut(m => { m.id = '1Atlas'; }) },
  { name: 'id with underscore', expect: 'invalid', text: mut(m => { m.id = 'Evidence_Atlas'; }) },
  { name: 'id of 49 characters', expect: 'invalid', text: mut(m => { m.id = 'A' + 'b'.repeat(48); }) },
  { name: 'empty title', expect: 'invalid', text: mut(m => { m.title = ''; }) },
  { name: 'blank title', expect: 'invalid', text: mut(m => { m.title = '   '; }) },
  { name: 'title of 121 characters', expect: 'invalid', text: mut(m => { m.title = 'x'.repeat(121); }) },
  { name: 'purpose of 601 characters', expect: 'invalid', text: mut(m => { m.purpose = 'x'.repeat(601); }) },
  { name: 'newline in title', expect: 'invalid', text: mut(m => { m.title = 'a\nb'; }) },
  { name: 'U+2028 in purpose', expect: 'invalid', text: mut(m => { m.purpose = 'a\u2028b'; }) },
  { name: 'unknown status', expect: 'invalid', text: mut(m => { m.status = 'done'; }) },
  { name: 'status in another case', expect: 'invalid', text: mut(m => { m.status = 'Mixed'; }) },
  { name: 'status with underscores', expect: 'invalid', text: mut(m => { m.status = 'not_yet_run'; }) },
  { name: 'empty class', expect: 'invalid', text: mut(m => { m.class = ''; }) },
  { name: 'class with a space', expect: 'invalid', text: mut(m => { m.class = 'a b'; }) },
  { name: 'class with a quote', expect: 'invalid', text: mut(m => { m.class = 'a"b'; }) },
  { name: 'null class', expect: 'invalid', text: mut(m => { m.class = null; }) },
  { name: 'empty context', expect: 'invalid', text: mut(m => { m.bottomLine.context = []; }) },
  { name: '7 context paragraphs', expect: 'invalid', text: mut(m => { m.bottomLine.context = many('x', 7); }) },
  { name: 'no sections', expect: 'invalid', text: mut(m => { m.sections = []; }) },
  { name: '4 sections', expect: 'invalid', text: mut(m => { m.sections.push({ tab: 'summary', items: proseItems(1) }); }) },
  { name: 'unknown tab', expect: 'invalid', text: mut(m => { sec(m, 'method').tab = 'summary'; }) },
  { name: 'tab in another case', expect: 'invalid', text: mut(m => { sec(m, 'method').tab = 'Method'; }) },
  { name: 'section without items', expect: 'invalid', text: mut(m => { sec(m, 'method').items = []; }) },
  { name: '25 items in a tab', expect: 'invalid', text: mut(m => { sec(m, 'method').items = proseItems(25); }) },
  { name: 'unknown item kind', expect: 'invalid', text: mut(m => { explore(m, 0).kind = 'table'; }) },
  { name: 'slot without id', expect: 'invalid', text: mut(m => { delete explore(m, 0).id; }) },
  { name: 'slot with a title', expect: 'invalid', text: mut(m => { explore(m, 0).title = 'x'; }) },
  { name: 'slot id starts upper-case', expect: 'invalid', text: mut(m => { explore(m, 0).id = 'Coverage'; }) },
  { name: 'slot id with a hyphen', expect: 'invalid', text: mut(m => { explore(m, 0).id = 'cov-erage'; }) },
  { name: 'slot id of 33 characters', expect: 'invalid', text: mut(m => { explore(m, 0).id = 'a' + 'b'.repeat(32); }) },
  { name: 'slot id constructor', expect: 'invalid', text: mut(m => { explore(m, 0).id = 'constructor'; }) },
  { name: 'slot id toString', expect: 'invalid', text: mut(m => { explore(m, 0).id = 'toString'; }) },
  { name: 'panel without slot', expect: 'invalid', text: mut(m => { delete explore(m, 3).slot; }) },
  { name: 'panel with a class', expect: 'invalid', text: mut(m => { explore(m, 3).class = 'x'; }) },
  { name: 'heading without title', expect: 'invalid', text: mut(m => { delete explore(m, 2).title; }) },
  { name: 'heading with empty description', expect: 'invalid', text: mut(m => { explore(m, 2).description = ''; }) },
  { name: 'heading with points', expect: 'invalid', text: mut(m => { explore(m, 2).points = ['x']; }) },
  { name: 'methodPanel without points', expect: 'invalid', text: mut(m => { delete method(m).points; }) },
  { name: 'methodPanel with empty points', expect: 'invalid', text: mut(m => { method(m).points = []; }) },
  { name: '13 points', expect: 'invalid', text: mut(m => { method(m).points = many('x', 13); }) },
  { name: 'empty point', expect: 'invalid', text: mut(m => { method(m).points = ['a', '']; }) },
  { name: 'point of 601 characters', expect: 'invalid', text: mut(m => { method(m).points = ['x'.repeat(601)]; }) },
  { name: 'prose without paragraphs', expect: 'invalid', text: mut(m => { sec(m, 'method').items = [{ kind: 'prose' }]; }) },
  { name: '7 paragraphs', expect: 'invalid', text: mut(m => { sec(m, 'method').items = [{ kind: 'prose', paragraphs: many('x', 7) }]; }) },
  { name: 'items as a string', expect: 'invalid', text: mut(m => { sec(m, 'method').items = 'x'; }) },

  { name: 'duplicate tab', expect: 'looser', reason: semantic, text: mut(m => { sec(m, 'method').tab = 'explore'; }) },
  { name: 'duplicate slot id across slot and panel', expect: 'looser', reason: semantic, text: mut(m => { explore(m, 0).id = 'matrix'; }) },
  { name: 'duplicate slot id across tabs', expect: 'looser', reason: semantic, text: mut(m => { sec(m, 'sources').items[1].slot = 'coverage'; }) },
  { name: 'duplicate panel slots', expect: 'looser', reason: semantic, text: mut(m => { explore(m, 5).slot = 'matrix'; }) },
  { name: 'multibyte title over 120 bytes', expect: 'looser', reason: bytesOnly, text: mut(m => { m.title = 'é'.repeat(61); }) },
  { name: 'multibyte purpose over 600 bytes', expect: 'looser', reason: bytesOnly, text: mut(m => { m.purpose = 'é'.repeat(301); }) },
  { name: 'multibyte point over 600 bytes', expect: 'looser', reason: bytesOnly, text: mut(m => { method(m).points = ['é'.repeat(301)]; }) },
  { name: 'duplicate JSON key', expect: 'looser', reason: 'a parsed value has already dropped duplicate keys',
    text: J(base).replace('"title":"Evidence Atlas",', '"title":"Evidence Atlas","title":"Again",') },
  { name: 'trailing data', expect: 'looser', reason: 'a parsed value has no trailing data', text: J(base) + '{}', parseText: J(base) },
  { name: 'input over 512 KiB', expect: 'looser', reason: 'the size limit applies to source bytes',
    text: J(base).replace(/\}$/, ' '.repeat(524289) + '}') },
];

const parsed = c => JSON.parse(c.parseText || c.text);

test('registry example satisfies the schema', () => {
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
test('schema agrees with diagramgen validate', { skip: hasGo ? false : 'go is not available; agreement with the generator was not checked' }, () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'research-page-schema-'));
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
