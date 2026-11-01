import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, readFileSync, writeFileSync, mkdirSync, rmSync, lstatSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';
import * as api from '../src/consumer-shadow-v1.mjs';
import { evaluate } from '../src/state.mjs';
import { importReviewedTask } from '../src/import-export.mjs';
import { makeSource, blobHash } from '../../skills/project-memory/scripts/lib/state.mjs';
import { command, taskSpec } from './fixtures.mjs';

const root = fileURLToPath(new URL('../../', import.meta.url));
const revision = 'a'.repeat(40);
const context = { actor_id: 'operator', role: 'owner', references_valid: true };
const wire = (value) => JSON.stringify(value);
function reviewed() {
  const record = `---\nid: ${taskSpec.task_id}\nstatus: in-progress\nowner: operator\nrepositories: [${taskSpec.repository.repository_id}]\n---\n\n# Invented consumer fixture\n\n## Implementation claim\n\nRead-only fixture; no live task.\n`;
  const path = `tasks/${taskSpec.task_id}-fixture.md`;
  return { version: 1, spec: structuredClone(taskSpec), reviewedBy: 'operator',
    expectedRepository: 'example/tasks', expectedRevision: revision,
    source: makeSource({ repository: 'example/tasks', revision, id: taskSpec.task_id, path,
      records: new Map([[path, { text: record, blob: blobHash(record) }]]) }) };
}
const transition = (changes = {}) => ({ version: 1, current: null,
  command: structuredClone(command), context: { ...context }, ...changes });
function sameError(direct, throughFacade) {
  let first;
  try { direct(); } catch (error) { first = error; }
  assert.ok(first, 'Direct shared engine must refuse the fixture');
  assert.throws(throughFacade, (error) => error.code === first.code && error.message === first.message);
}

test('only the narrow versioned shadow surface is exported', () => {
  assert.deepEqual(Object.keys(api).sort(),
    ['SHADOW_API_VERSION', 'SHADOW_LIMITS', 'previewReviewedTask', 'previewTransition'].sort());
  assert.equal(api.SHADOW_API_VERSION, 1);
  assert.ok(Object.isFrozen(api.SHADOW_LIMITS));
});

test('reviewed-source preview delegates without granting admission or copying source prose', () => {
  const input = reviewed(), before = structuredClone(input);
  const { version, ...direct } = input;
  const output = api.previewReviewedTask(wire(input));
  assert.deepEqual(output, { schema: 'clanker-foundry/shadow-preview', version: 1,
    mode: 'shadow-only', authority: 'none', kind: 'reviewed-task', value: importReviewedTask(direct) });
  assert.equal(Object.hasOwn(output.value.source, 'record'), false);
  assert.deepEqual(input, before);
  assert.deepEqual(api.previewReviewedTask(wire(Object.fromEntries(Object.entries(input).reverse()))), output);
});

test('source provenance, genuine-claim and identity refusals are preserved', () => {
  const altered = [];
  for (const change of [
    { expectedRepository: 'other/tasks' }, { expectedRevision: 'b'.repeat(40) },
    { reviewedBy: 'another-owner' }, { spec: { ...taskSpec, task_id: 'CF-other' } },
  ]) altered.push({ ...reviewed(), ...change });
  for (const mutate of [
    (s) => s.replace('status: in-progress', 'status: done'),
    (s) => s.replace('## Implementation claim', '## Background'),
    (s) => `${s}\n## Another claim\nUnrelated fixture.\n`,
  ]) {
    const input = reviewed();
    input.source.record = mutate(input.source.record);
    input.source.blob = blobHash(input.source.record);
    input.source.identityEvidence.canonicals[0].blob = input.source.blob;
    input.source.identityEvidence.canonicals[0].status = input.source.record.includes('status: done') ? 'done' : 'in-progress';
    altered.push(input);
  }
  const forged = reviewed(); forged.source.record += 'Different bytes'; altered.push(forged);
  for (const input of altered) {
    const { version, ...direct } = input;
    sameError(() => importReviewedTask(direct), () => api.previewReviewedTask(wire(input)));
  }
});

test('transition preview uses the same admission guards and produces hypothetical state only', () => {
  const input = transition(), before = structuredClone(input);
  const result = api.previewTransition(wire(input));
  assert.equal(result.mode, 'shadow-only'); assert.equal(result.authority, 'none');
  assert.equal(result.kind, 'transition'); assert.equal(result.evaluator_version, 1);
  assert.deepEqual(result.value, evaluate(input.current, input.command, input.context));
  assert.equal(result.value.execution.state, 'PENDING');
  assert.deepEqual(api.previewTransition(wire(input)), result);
  assert.deepEqual(input, before);
  for (const change of [
    { command: { ...command, expected_revision: 1 } },
    { context: { ...context, references_valid: false } },
    { context: { ...context, actor_id: 'other' } },
    { command: { ...command, payload: { spec: { ...taskSpec, write_scope: ['a.txt'] } } } },
    { command: { ...command, payload: { spec: { ...taskSpec, dependencies: [
      { task_id: taskSpec.task_id, spec_digest: 'b'.repeat(64) },
    ] } } } },
  ]) {
    const trial = transition(change);
    sameError(() => evaluate(trial.current, trial.command, trial.context),
      () => api.previewTransition(wire(trial)));
  }
});

test('assignment intents remain hypothetical and stale revisions/uncertain prerequisites refuse', () => {
  const current = evaluate(null, command, context).execution;
  const assignment = { schema_version: 1, command_id: 'assign-1', execution_id: current.execution_id,
    expected_revision: current.revision, at: '2026-10-30T00:00:01.000Z', kind: 'ASSIGN', payload: {
      attempt_id: 'attempt-1', worker_id: 'reviewer', input_revisions: [taskSpec.repository.revision],
      workspace: 'fixture-workspace', capabilities: ['read'], lease_deadline: '2026-10-30T00:00:30.000Z',
    } };
  const allowed = { ...context, dependencies_satisfied: true, resources_available: true, worker_eligible: true };
  const input = transition({ current, command: assignment, context: allowed });
  const output = api.previewTransition(wire(input));
  assert.deepEqual(output.value, evaluate(current, assignment, allowed));
  assert.ok(output.value.effects.some((effect) => effect.kind === 'LAUNCH_ATTEMPT'));
  assert.equal(output.authority, 'none');
  for (const trial of [
    { ...input, command: { ...assignment, expected_revision: 0 } },
    { ...input, context: { ...allowed, dependencies_satisfied: false } },
    { ...input, context: { ...allowed, resources_available: false } },
  ]) sameError(() => evaluate(trial.current, trial.command, trial.context),
    () => api.previewTransition(wire(trial)));
});

test('JSON-only boundary rejects unsupported, duplicate, oversized and deeply nested inputs', () => {
  const input = transition();
  for (const trial of [null, {}, [], 1, wire(null), wire([]), '{}',
    wire({ ...input, version: 2 }), wire({ ...input, surprise: true }),
    wire({ ...input, context: { ...context, inherited_authority: true } }),
    wire({ ...input, context: { ...context, references_valid: 'true' } }),
    wire({ ...input, context: { ...context, approver_ids: Array(257).fill('owner') } }),
    wire({ ...input, context: { ...context, trusted_checks: [{ check_id: 'a', check_version: '1', runner_id: 'x', extra: true }] } }),
    wire(input).replace('"version":1', '"version":1,"version":1'),
    '{"__proto__":1}', ' '.repeat(api.SHADOW_LIMITS.maxBytes + 1),
    '['.repeat(66) + '0' + ']'.repeat(66),
  ]) assert.throws(() => api.previewTransition(trial));
  for (const trial of [wire({ ...reviewed(), extra: 1 }), wire({ ...reviewed(), version: '1' })]) {
    assert.throws(() => api.previewReviewedTask(trial));
  }
});

test('root dependency lock preserves the existing pinned runtime dependency graph', () => {
  const runtime = JSON.parse(readFileSync(join(root, 'runtime/package-lock.json'), 'utf8'));
  const consumer = JSON.parse(readFileSync(join(root, 'package-lock.json'), 'utf8'));
  consumer.name = runtime.name; consumer.packages[''].name = runtime.name;
  assert.deepEqual(consumer, runtime);
});

test('packed artifact installs in a clean offline consumer with no operational exports or effects', { timeout: 60_000 }, () => {
  const temp = mkdtempSync(join(tmpdir(), 'foundry-consumer-'));
  try {
    const packed = JSON.parse(execFileSync('npm', ['pack', '--json', '--ignore-scripts', '--pack-destination', temp],
      { cwd: root, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }))[0];
    const metadata = JSON.parse(readFileSync(join(root, 'package.json'), 'utf8'));
    const expected = [...metadata.files, 'README.md', 'package.json'].sort();
    assert.deepEqual(packed.files.map((file) => file.path).sort(), expected);
    assert.equal(metadata.private, true); assert.equal(Object.hasOwn(metadata, 'scripts'), false);
    const consumer = join(temp, 'consumer'); mkdirSync(consumer);
    const archive = `file:${join(temp, packed.filename)}`;
    const consumerPackage = { name: 'foundry-consumer-fixture', version: '1.0.0', private: true,
      type: 'module', dependencies: { [metadata.name]: archive } };
    const lock = JSON.parse(readFileSync(join(root, 'package-lock.json'), 'utf8'));
    lock.name = consumerPackage.name; lock.version = consumerPackage.version;
    lock.packages[''] = { name: lock.name, version: lock.version, dependencies: consumerPackage.dependencies };
    lock.packages[`node_modules/${metadata.name}`] = { version: metadata.version, resolved: archive,
      integrity: packed.integrity, dependencies: metadata.dependencies, engines: metadata.engines };
    writeFileSync(join(consumer, 'package.json'), wire(consumerPackage));
    writeFileSync(join(consumer, 'package-lock.json'), wire(lock));
    execFileSync('npm', ['ci', '--offline', '--ignore-scripts', '--no-fund', '--no-audit'],
    { cwd: consumer, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
    assert.equal(lstatSync(join(consumer, 'node_modules/@clanker-foundry/orchestration')).isSymbolicLink(), false);
    const script = `
      import assert from 'node:assert/strict';
      import fs from 'node:fs'; import fsp from 'node:fs/promises';
      import cp from 'node:child_process'; import http from 'node:http'; import https from 'node:https';
      import net from 'node:net'; import { syncBuiltinESMExports } from 'node:module';
      const fail = () => { throw new Error('Unexpected consumer import effect'); };
      for (const key of ['writeFileSync','appendFileSync','mkdirSync','rmSync','unlinkSync']) fs[key] = fail;
      for (const key of ['writeFile','appendFile','mkdir','rm','unlink']) fsp[key] = fail;
      for (const key of ['spawn','spawnSync','exec','execSync','execFile','execFileSync','fork']) cp[key] = fail;
      http.request = http.get = https.request = https.get = net.connect = net.createConnection = fail;
      globalThis.fetch = fail; syncBuiltinESMExports();
      const api = await import('@clanker-foundry/orchestration/shadow/v1');
      assert.equal(api.previewTransition(${wire(wire(transition()))}).mode, 'shadow-only');
      assert.equal(api.previewReviewedTask(${wire(wire(reviewed()))}).authority, 'none');
      for (const subpath of ['', '/runtime/src/store.mjs', '/runtime/src/runner.mjs', '/shadow/v2']) {
        await assert.rejects(import('@clanker-foundry/orchestration' + subpath),
          error => error.code === 'ERR_PACKAGE_PATH_NOT_EXPORTED');
      }
      process.stdout.write('consumer-ok');
    `;
    const env = { ...process.env }; delete env.NODE_PATH; delete env.NODE_OPTIONS;
    const out = execFileSync(process.execPath, ['--input-type=module', '--eval', script],
      { cwd: consumer, encoding: 'utf8', env, stdio: ['ignore', 'pipe', 'pipe'] });
    assert.equal(out, 'consumer-ok');
  } finally { rmSync(temp, { recursive: true, force: true }); }
});
