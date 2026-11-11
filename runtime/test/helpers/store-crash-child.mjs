import assert from 'node:assert/strict';
import { writeSync } from 'node:fs';
import { once } from 'node:events';
import { DatabaseSync } from 'node:sqlite';
import { openStore } from '../../src/store.mjs';

const [path, inputJSON, method, boundary] = process.argv.slice(2);
const input = JSON.parse(inputJSON);
const store = openStore(path, { worker_ids: ['worker-1'] });
const tables = ['executions', 'events', 'attempts', 'reservations', 'outbox', 'receipts'];
const report = (message) => writeSync(1, `${JSON.stringify(message)}\n`);
let reached = false;

function pause(db) {
  reached = true;
  const rows = Object.fromEntries(tables.map((table) => [table, db.prepare(`SELECT * FROM ${table} ORDER BY rowid`).all()]));
  report({ point: boundary, rows });
  Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 20_000);
  throw new Error(`Crash boundary ${boundary} timed out without parent interruption`);
}

// Open/schema transactions have finished. Only the requested operation is hooked.
// Keep the receiver, arguments and real SQL execution unchanged.
const originalExec = DatabaseSync.prototype.exec;
DatabaseSync.prototype.exec = function (sql, ...args) {
  if (sql.trim().toUpperCase() === 'BEGIN IMMEDIATE') report({ point: 'begin' });
  if (sql.trim().toUpperCase() === 'COMMIT' && boundary === 'before-commit') pause(this);
  const result = originalExec.call(this, sql, ...args);
  if (sql.trim().toUpperCase() === 'COMMIT' && boundary === 'after-commit') pause(this);
  return result;
};

try {
  report({ point: 'ready' });
  const [data] = await once(process.stdin, 'data');
  assert.equal(data.toString(), 'go\n');
  process.stdin.pause();
  const result = store[method](input, { actor_id: 'operator', role: 'owner' });
  assert.ok(boundary === 'normal' || reached, `Requested crash boundary ${boundary} was not reached`);
  report({ point: 'result', result });
} finally {
  DatabaseSync.prototype.exec = originalExec;
  store.close();
}
