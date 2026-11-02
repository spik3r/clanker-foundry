'use strict';

const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
const requireRuntime = createRequire(path.resolve(__dirname, '../../../runtime/package.json'));
const Ajv = requireRuntime('ajv');
const read = name => JSON.parse(fs.readFileSync(path.join(__dirname, name), 'utf8'));
const schema = read('lifecycle-v1.schema.json');
const cases = read('cases.json');
const validate = new Ajv({ allErrors: true, strict: true }).compile(schema);

test('schema compiles and classifies independently stated cases', () => {
  assert.equal(new Set(cases.map(c => c.name)).size, cases.length);
  for (const c of cases) {
    assert.equal(validate(c.document), c.schemaValid, `${c.name}: ${JSON.stringify(validate.errors)}`);
    assert.ok(!c.modelValid || c.schemaValid, `${c.name}: accepted model must pass schema`);
  }
});

test('existing lifecycle source is representable', () => {
  const example = JSON.parse(fs.readFileSync(path.join(__dirname, '../examples/lifecycle.json'), 'utf8'));
  assert.equal(validate(example), true, JSON.stringify(validate.errors));
});
