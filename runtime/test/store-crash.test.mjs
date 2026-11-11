import test from 'node:test';
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { chmodSync, mkdtempSync, realpathSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { DatabaseSync } from 'node:sqlite';
import { openStore } from '../src/store.mjs';
import { taskSpec } from './fixtures.mjs';

const owner = { actor_id: 'operator', role: 'owner' };
const policy = { worker_ids: ['worker-1'] };
const at = '2026-10-30T00:00:00.000Z';
const tables = ['executions', 'events', 'attempts', 'reservations', 'outbox', 'receipts'];
const beforeCounts = [1, 1, 0, 0, 0, 1];
const assignedCounts = [1, 2, 1, 1, 1, 2];
const wire = (value) => JSON.parse(JSON.stringify(value));
const code = (expected) => (error) => error.code === expected;

function snapshot(db) {
  return wire(Object.fromEntries(tables.map((table) => [table, db.prepare(`SELECT * FROM ${table} ORDER BY rowid`).all()])));
}
function counts(rows, expected) {
  assert.deepEqual(tables.map((table) => rows[table].length), expected);
}
function receipt(rows) {
  return JSON.parse(rows.receipts.find((row) => row.command_id === 'assign').body);
}

function fixture(t) {
  const directory = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-store-crash-')));
  chmodSync(directory, 0o700);
  const path = join(directory, 'authority.sqlite');
  const connections = [];
  const children = [];
  const open = () => { const store = openStore(path, policy); connections.push(store); return store; };
  // One cleanup owns ordering: reap children before closing readers/removing WALs.
  t.after(async () => {
    try {
      await Promise.all(children.map(async (child) => {
        if (!child.closed) child.process.kill('SIGKILL');
        await child.exit;
      }));
    } finally {
      try {
        const errors = [];
        for (const connection of connections) {
          try { connection.close(); } catch (error) { errors.push(error); }
        }
        if (errors.length) throw new AggregateError(errors, 'Crash fixture connection cleanup failed');
      } finally { rmSync(directory, { recursive: true, force: true }); }
    }
  });
  const store = open();
  const spec = structuredClone(taskSpec);
  spec.task_id = 'CF-synthetic-store-crash';
  spec.workflow.stages[0].role = 'explorer-researcher';
  spec.shared_resources = [{ resource_id: 'synthetic-crash-lock', mode: 'exclusive' }];
  store.execute({ schema_version: 1, command_id: 'admit', execution_id: 'crash-execution', expected_revision: 0, at, kind: 'ADMIT', payload: { spec } }, owner);
  const reader = new DatabaseSync(path, { readOnly: true });
  connections.push(reader);
  const assign = { schema_version: 1, command_id: 'assign', execution_id: 'crash-execution', expected_revision: 1, at, kind: 'ASSIGN', payload: {
    attempt_id: 'crash-attempt', worker_id: 'worker-1', input_revisions: [], workspace: 'synthetic-no-launch', capabilities: ['read'], lease_deadline: '2026-10-30T00:01:00.000Z',
  } };
  return { path, open, store, reader, assign, children };
}

function child(f, input = f.assign, method = 'execute', boundary = 'normal') {
  const processHandle = spawn(process.execPath, [fileURLToPath(new URL('./helpers/store-crash-child.mjs', import.meta.url)), f.path, JSON.stringify(input), method, boundary], { stdio: ['pipe', 'pipe', 'pipe'] });
  const state = { process: processHandle, closed: false, messages: [] };
  f.children.push(state);
  let output = '', errors = '', failure;
  const listeners = new Set();
  const notify = () => { for (const listener of listeners) listener(); };
  const watchdog = setTimeout(() => { failure = new Error('Owned crash child exceeded 25s'); processHandle.kill('SIGKILL'); }, 25_000);
  state.exit = new Promise((resolve) => {
    processHandle.on('error', (error) => { failure = error; notify(); });
    processHandle.on('close', (status, signal) => {
      clearTimeout(watchdog);
      state.closed = true;
      if (output) failure ??= new Error('Child exited with an incomplete message');
      state.end = { status, signal, errors, failure };
      notify();
      resolve(state.end);
    });
  });
  processHandle.stdin.on('error', (error) => { failure = error; notify(); });
  processHandle.stderr.on('data', (data) => { errors = (errors + data).slice(-16_384); });
  processHandle.stdout.on('data', (data) => {
    output += data;
    if (output.length > 1_048_576) { failure = new Error('Crash child output exceeded limit'); processHandle.kill('SIGKILL'); }
    let newline;
    while ((newline = output.indexOf('\n')) !== -1) {
      const line = output.slice(0, newline); output = output.slice(newline + 1);
      try {
        const message = JSON.parse(line);
        assert.ok(['ready', 'begin', 'before-commit', 'after-commit', 'result'].includes(message.point), 'Unexpected child marker');
        assert.equal(state.messages.some((value) => value.point === message.point), false, 'Duplicate child marker');
        assert.ok(state.messages.length < 3, 'Too many child messages');
        state.messages.push(message);
      } catch (error) { failure = error; }
    }
    notify();
  });
  state.wait = (point) => new Promise((resolve, reject) => {
    const timeout = setTimeout(() => finish(new Error(`Timed out waiting for ${point}: ${errors}`)), 10_000);
    function finish(error, value) { clearTimeout(timeout); listeners.delete(check); error ? reject(error) : resolve(value); }
    function check() {
      if (failure) return finish(failure);
      const message = state.messages.find((value) => value.point === point);
      if (message) return finish(null, message);
      if (state.closed) finish(new Error(`Child exited before ${point}: ${JSON.stringify(state.end)}`));
    }
    listeners.add(check); check();
  });
  state.go = () => processHandle.stdin.end('go\n');
  return state;
}
async function paused(f, boundary, input = f.assign, method = 'execute') {
  const processHandle = child(f, input, method, boundary);
  await processHandle.wait('ready'); processHandle.go();
  const message = await processHandle.wait(boundary);
  assert.equal(processHandle.messages.some((value) => value.point === 'result'), false);
  return { processHandle, rows: message.rows };
}
async function kill(processHandle) {
  assert.equal(processHandle.closed, false);
  assert.equal(processHandle.process.kill('SIGKILL'), true);
  const end = await processHandle.exit;
  assert.equal(end.failure, undefined);
  assert.equal(end.signal, 'SIGKILL', end.errors);
  assert.equal(end.status, null);
  assert.equal(processHandle.messages.some((value) => value.point === 'result'), false);
}
async function result(processHandle) {
  const message = await processHandle.wait('result');
  const end = await processHandle.exit;
  assert.equal(end.failure, undefined);
  assert.equal(end.status, 0, end.errors);
  assert.equal(end.signal, null);
  return message.result;
}
async function retries(f, input = f.assign, method = 'execute') {
  const processes = [child(f, input, method), child(f, input, method)];
  await Promise.all(processes.map((value) => value.wait('ready')));
  for (const value of processes) value.go();
  return Promise.all(processes.map(result));
}
function committed(f, expectedReceipt) {
  const rows = snapshot(f.reader);
  counts(rows, assignedCounts);
  assert.deepEqual(receipt(rows), expectedReceipt);
  const reopened = f.open();
  assert.equal(reopened.get('crash-execution').state, 'ASSIGNED');
  assert.equal(reopened.get('crash-execution').revision, 2);
  const replay = reopened.replay();
  assert.deepEqual(replay.executions, reopened.list());
  assert.deepEqual(wire(replay.effects), expectedReceipt.effects);
  return reopened;
}

// These POSIX process deaths do not simulate power loss or execute a provider.
test('death before COMMIT preserves every prior row and replay before a later retry', { timeout: 40_000 }, async (t) => {
  const f = fixture(t), before = snapshot(f.reader), replay = f.store.replay();
  counts(before, beforeCounts);
  const holder = await paused(f, 'before-commit');
  counts(holder.rows, assignedCounts);
  assert.deepEqual(snapshot(f.reader), before);
  assert.equal(f.store.get('crash-execution').state, 'PENDING');
  await kill(holder.processHandle);
  const reopened = f.open();
  assert.deepEqual(snapshot(f.reader), before);
  assert.deepEqual(reopened.replay(), replay);
  assert.deepEqual(reopened.pendingEffects(), []);
  const retried = wire(reopened.execute(f.assign, owner));
  assert.deepEqual(retried, receipt(holder.rows));
  committed(f, retried);
  assert.deepEqual(snapshot(f.reader), holder.rows);
});

test('pre-opened identical competitor commits once after uncommitted writer death', { timeout: 40_000 }, async (t) => {
  const f = fixture(t), before = snapshot(f.reader);
  const competitor = child(f);
  await competitor.wait('ready');
  const holder = await paused(f, 'before-commit');
  counts(holder.rows, assignedCounts);
  competitor.go();
  // This marker precedes native BEGIN. It proves attempted entry while the
  // writer owns the lock, not how long the native busy handler has run.
  await competitor.wait('begin');
  assert.equal(competitor.messages.some((value) => value.point === 'result'), false);
  assert.deepEqual(snapshot(f.reader), before);
  await kill(holder.processHandle);
  const retried = await result(competitor);
  assert.deepEqual(retried, receipt(holder.rows));
  assert.deepEqual(wire(f.store.execute(f.assign, owner)), retried);
  committed(f, retried);
  assert.deepEqual(snapshot(f.reader), holder.rows);
});

test('death after assignment COMMIT preserves the original receipt for concurrent retries', { timeout: 40_000 }, async (t) => {
  const f = fixture(t), holder = await paused(f, 'after-commit');
  counts(holder.rows, assignedCounts);
  const original = receipt(holder.rows);
  assert.deepEqual(snapshot(f.reader), holder.rows);
  await kill(holder.processHandle);
  const results = await retries(f);
  for (const value of results) assert.deepEqual(value, original);
  assert.throws(() => f.store.execute({ ...f.assign, payload: { ...f.assign.payload, workspace: 'changed' } }, owner), code('COMMAND_ID_REUSED'));
  committed(f, original);
});

test('death after acknowledgment COMMIT keeps delivery and receipt fencing across replay', { timeout: 40_000 }, async (t) => {
  const f = fixture(t), assigned = wire(f.store.execute(f.assign, owner));
  const effectId = assigned.effects[0].effect_id;
  const claim = f.store.claimEffect({ effectId, dispatcherId: 'synthetic-dispatcher', now: at, leaseMs: 1000 });
  const ack = { effectId, dispatcherId: 'synthetic-dispatcher', claimToken: claim.claim_token, outcome: 'DELIVERED', receipt: { synthetic: true, noProviderInvoked: true } };
  const holder = await paused(f, 'after-commit', ack, 'acknowledgeEffect');
  assert.deepEqual(snapshot(f.reader), holder.rows);
  assert.equal(holder.rows.outbox[0].status, 'DELIVERED');
  assert.deepEqual(JSON.parse(holder.rows.outbox[0].outcome_receipt), ack.receipt);
  const original = wire(f.store.acknowledgeEffect(ack));
  assert.equal(original.status, 'DELIVERED');
  await kill(holder.processHandle);
  for (const value of await retries(f, ack, 'acknowledgeEffect')) assert.deepEqual(value, original);
  const reopened = committed(f, assigned);
  assert.throws(() => reopened.acknowledgeEffect({ ...ack, claimToken: 'incorrect' }), code('STALE_EFFECT_CLAIM'));
  assert.throws(() => reopened.acknowledgeEffect({ ...ack, receipt: { changed: true } }), code('EFFECT_OUTCOME_CONFLICT'));
  assert.deepEqual(wire(reopened.claimEffect({ effectId, dispatcherId: 'would-be-duplicate', now: at, leaseMs: 1000 })), original);
  reopened.replay();
  assert.deepEqual(reopened.pendingEffects(), []);
  assert.deepEqual(snapshot(f.reader), holder.rows);
});
