import test from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { realpathSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { openStore } from '../src/store.mjs';
import { taskSpec } from './fixtures.mjs';

const cli = fileURLToPath(new URL('../src/cli.mjs', import.meta.url));
const run = (...args) => spawnSync(process.execPath, [cli, ...args], { encoding: 'utf8' });

test('CLI simulates all four stages with approval, cleanup, persistence and replay', (t) => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-cli-')));
  t.after(() => rmSync(root, { force: true, recursive: true }));
  const db = join(root, 'authority.sqlite');
  const output = run('demo', db);
  assert.equal(output.status, 0, output.stderr);
  const summary = JSON.parse(output.stdout);
  assert.equal(summary.mode, 'SIMULATION');
  assert.equal(summary.state, 'COMPLETED');
  assert.equal(summary.cleanup_pending, false);
  assert.equal(summary.stages, 4);
  assert.equal(summary.replayed_executions, 1);
  const shown = run('show', db, summary.execution_id);
  assert.equal(shown.status, 0, shown.stderr);
  const execution = JSON.parse(shown.stdout);
  assert.equal(execution.worker_history.length, 4);
  assert.notEqual(execution.worker_history[2].worker_id, execution.worker_history[3].worker_id);
  const exported = run('export', db, summary.execution_id);
  assert.equal(exported.status, 0);
  assert.match(exported.stdout, /Read-only projection/);
  assert.equal(run('replay', db).status, 0);
});

test('CLI rejects invalid commands and duplicate JSON without echoing input secrets', (t) => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-cli-invalid-')));
  t.after(() => rmSync(root, { force: true, recursive: true }));
  const file = join(root, 'invalid.json');
  writeFileSync(file, '{"credential":"test-secret-value","credential":"duplicate"}');
  const output = run('validate', 'TaskSpec', file);
  assert.equal(output.status, 1);
  assert.doesNotMatch(output.stderr, /test-secret-value/);
  assert.equal(run('command').status, 1);
  assert.equal(run('--help').status, 0);
});

test('demo refuses a nonempty authority without dispatching its real pending work', (t) => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-cli-existing-')));
  t.after(() => rmSync(root, { force: true, recursive: true }));
  const db = join(root, 'authority.sqlite');
  const store = openStore(db, { worker_ids: ['actual-worker'] });
  const at = '2026-10-30T00:00:00.000Z';
  const spec = structuredClone(taskSpec);
  spec.workflow.stages[0].role = 'explorer-researcher';
  const send = (kind, payload, revision) => store.execute({ schema_version: 1,
    command_id: `actual-${kind}`, execution_id: 'actual-execution', expected_revision: revision,
    at, kind, payload }, { actor_id: 'operator', role: 'owner' });
  send('ADMIT', { spec }, 0);
  send('ASSIGN', { attempt_id: 'actual-attempt', worker_id: 'actual-worker',
    lease_deadline: '2026-10-30T00:01:00.000Z', workspace: 'real-work',
    capabilities: ['read'], input_revisions: [] }, 1);
  const output = run('demo', db);
  assert.equal(output.status, 1);
  assert.match(output.stderr, /DEMO_DATABASE_NOT_EMPTY/);
  assert.equal(store.get('actual-execution').state, 'ASSIGNED');
  assert.equal(store.events().length, 2);
  assert.equal(store.pendingEffects()[0].status, 'PENDING');
  store.close();
});

test('CLI exposes stable database identity and absent namespace without activating authority', (t) => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-cli-authority-')));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  const db = join(root, 'authority.sqlite');
  const first = run('authority-id', db);
  assert.equal(first.status, 0, first.stderr);
  const identity = JSON.parse(first.stdout);
  assert.match(identity.authority_instance_id, /^[a-f0-9-]{36}$/);
  assert.deepEqual(JSON.parse(run('authority-id', db).stdout), identity);
  const status = run('authority-status', db, 'cf-auto');
  assert.equal(status.status, 0, status.stderr);
  assert.equal(JSON.parse(status.stdout), null);
});
