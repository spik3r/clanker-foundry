import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve, dirname } from 'node:path';
import { evaluate } from '../src/state.mjs';
import { verify, render, load, checkViews } from '../model-trace-v1/check.mjs';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
const read = (p) => readFileSync(resolve(root,p), 'utf8');
const path = 'runtime/model-trace-v1/';
const rawOracle = read(`${path}testdata/oracle.json`);
const oracle = JSON.parse(rawOracle);
const manifest = JSON.parse(read(`${path}manifest.json`));
const files = load();
const pinned = '09b152afafec087c7101e71b26ff89a9a531288a45cafa1382d0063c4c0f411b';

// The oracle was independently stated and committed before this test. These
// selected assertions are not exhaustive evaluator behavior or human review.
test('frozen shadow trace replays through the one existing evaluator', () => {
  assert.equal(createHash('sha256').update(rawOracle).digest('hex'), pinned);
  assert.deepEqual(oracle.cases.map((c) => c.id), [
    'admit_read_only_pending', 'assign_dependencies_missing', 'assign_dependencies_false',
    'assign_resources_missing', 'assign_resources_false',
  ]);
  for (const c of oracle.cases) {
    const {current, command, context} = structuredClone(c.input);
    const before = structuredClone({current, command, context});
    if (c.expected.completion === 'return') {
      const result = evaluate(current, command, context);
      for (const [key,value] of Object.entries(c.expected.executionSelected)) assert.deepEqual(result.execution[key], value, `${c.id}: ${key}`);
      assert.deepEqual(result.execution.spec, command.payload.spec);
      assert.notStrictEqual(result.execution.spec, command.payload.spec);
      assert.deepEqual(result.event, c.expected.event);
      assert.deepEqual(result.effects, c.expected.effects);
    } else {
      assert.equal(c.expected.completion, 'throw');
      let returned = false;
      assert.throws(() => { evaluate(current, command, context); returned = true; }, (error) => {
        assert.equal(error.code,c.expected.error.code, c.id);
        assert.equal(error.message,c.expected.error.message,c.id);
        return true;
      });
      assert.equal(returned,false);
      assert.equal(c.expected.resultReturned,false);
      assert.equal(c.expected.eventReturned,false);
      assert.equal(c.expected.effectsReturned,false);
      for (const [key,value] of Object.entries(c.expected.currentSelectedAfter)) assert.deepEqual(current[key],value,`${c.id}: ${key}`);
    }
    assert.deepEqual({current, command, context}, before, `${c.id}: inputs mutated`);
  }
});

test('source-pinned trace covers the schema inventory and generated views', () => {
  assert.equal(verify(manifest, files),true);
  for (const [name, expected] of Object.entries(render(manifest))) assert.equal(read(name),expected,`${name} is stale`);
});

test('trace controls reject drift, omissions and dangling evidence', () => {
  assert.throws(() => verify(manifest, {...files, 'runtime/src/state.mjs': `${files['runtime/src/state.mjs']}\n`}), /source pin/);
  const mutant = (change, expected) => { const copy = structuredClone(manifest); change(copy); assert.throws(() => verify(copy,files),expected); };
  mutant((m) => m.commands.pop(), /command inventory/);
  mutant((m) => m.states.push({id:'MADE_UP',coverage:'unmapped'}), /state inventory/);
  mutant((m) => {m.commands.find((x) => x.id === 'CANCEL').coverage = 'partial';}, /command coverage/);
  mutant((m) => {m.requirements[0].schemaPointer = '#/$defs/Command/oneOf/1';}, /schema reference/);
  mutant((m) => {m.requirements[0].anchor = 'no such evaluator source';}, /source anchor/);
  mutant((m) => {m.requirements[0].caseIds = ['not_a_case'];}, /case reference/);
  mutant((m) => {m.requirements[0].existingTest = 'not a test';}, /test reference/);
  mutant((m) => {m.requirements[1].caseIds = ['assign_resources_missing'];}, /case outcome/);
  mutant((m) => {m.flows[1].requirementIds = ['TR-ADMIT'];}, /flow reference/);
  const views = render(manifest);
  assert.equal(checkViews(manifest, views),true);
  assert.throws(() => checkViews(manifest, {...views, [`${path}views/sequence.md`]: `${views[`${path}views/sequence.md`]}tamper`}), /stale view/);
  assert.throws(() => checkViews(manifest, {...views, [`${path}views/state.dot`]: 'corrupted'}), /stale view/);
});
