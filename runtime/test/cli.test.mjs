import test from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { realpathSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { openStore } from '../src/store.mjs';
import { taskSpec } from './fixtures.mjs';

const cli = fileURLToPath(new URL('../src/cli.mjs', import.meta.url));
const run = (...args) => spawnSync(process.execPath, [cli, ...args], { encoding: 'utf8' });

test('CLI reads at most one byte beyond its JSON file limit', (t) => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-cli-read-')));
  t.after(() => rmSync(root, { force: true, recursive: true }));
  const file = join(root, 'input.json');
  const preload = join(root, 'probe.cjs');
  const report = join(root, 'read-report.json');
  writeFileSync(preload, `
    const fs = require('node:fs');
    const { syncBuiltinESMExports } = require('node:module');
    const original = { open: fs.openSync, stat: fs.fstatSync, read: fs.readSync,
      close: fs.closeSync, write: fs.writeFileSync };
    let target;
    const report = { requested: 0, returned: 0, calls: 0, positiveCalls: 0,
      closed: false, statSize: null };
    fs.openSync = (path, ...args) => {
      const fd = original.open(path, ...args);
      if (target === undefined && path === process.env.FOUNDRY_PROBE_FILE) target = fd;
      return fd;
    };
    fs.fstatSync = (fd, ...args) => {
      const stat = original.stat(fd, ...args);
      if (fd === target) {
        report.statSize = stat.size;
        const growTo = Number(process.env.FOUNDRY_PROBE_GROW_TO || 0);
        if (growTo > stat.size) fs.appendFileSync(process.env.FOUNDRY_PROBE_FILE,
          Buffer.alloc(growTo - stat.size, 32));
      }
      return stat;
    };
    fs.readSync = (fd, buffer, offset, length, position) => {
      if (fd !== target) return original.read(fd, buffer, offset, length, position);
      report.calls += 1;
      report.requested += length;
      if (process.env.FOUNDRY_PROBE_FAIL === '1') throw Object.assign(new Error('read failed'), { code: 'EIO' });
      const count = original.read(fd, buffer, offset,
        process.env.FOUNDRY_PROBE_SHORT === '1' ? Math.min(length, 3) : length, position);
      report.returned += count;
      if (count > 0) report.positiveCalls += 1;
      return count;
    };
    fs.closeSync = (fd) => {
      if (fd === target) report.closed = true;
      return original.close(fd);
    };
    syncBuiltinESMExports();
    process.on('exit', () => original.write(process.env.FOUNDRY_PROBE_REPORT,
      JSON.stringify(report)));
  `);
  const limit = 1_048_576;
  const probe = (contents, options = {}) => {
    writeFileSync(file, contents);
    const result = spawnSync(process.execPath, ['--require', preload, cli,
      'validate', options.contract ?? 'NoSuchContract', file], {
      encoding: 'utf8', timeout: 10_000, killSignal: 'SIGKILL',
      env: { ...process.env, FOUNDRY_PROBE_FILE: file, FOUNDRY_PROBE_REPORT: report,
        FOUNDRY_PROBE_GROW_TO: String(options.growTo ?? 0),
        FOUNDRY_PROBE_SHORT: options.short ? '1' : '0',
        FOUNDRY_PROBE_FAIL: options.fail ? '1' : '0' },
    });
    assert.equal(result.error, undefined);
    return { result, reads: JSON.parse(readFileSync(report, 'utf8')) };
  };

  let outcome = probe('');
  assert.match(outcome.result.stderr, /MALFORMED_JSON/);
  assert.equal(outcome.reads.returned, 0);
  assert.equal(outcome.reads.closed, true);

  outcome = probe('{"a":1}', { short: true });
  assert.match(outcome.result.stderr, /UNKNOWN_CONTRACT/);
  assert.ok(outcome.reads.positiveCalls > 1);
  assert.equal(outcome.reads.returned, 7);

  const spec = JSON.stringify(taskSpec);
  const exact = spec + ' '.repeat(limit - Buffer.byteLength(spec));
  outcome = probe(exact, { contract: 'TaskSpec' });
  assert.equal(outcome.result.status, 0, outcome.result.stderr);
  assert.equal(outcome.reads.returned, limit);
  assert.equal(outcome.reads.closed, true);

  outcome = probe(exact + ' ');
  assert.match(outcome.result.stderr, /INVALID_INPUT_FILE/);
  assert.equal(outcome.reads.calls, 0);
  assert.equal(outcome.reads.closed, true);

  outcome = probe(exact, { growTo: limit + 1 });
  assert.equal(outcome.reads.statSize, limit);
  assert.match(outcome.result.stderr, /INVALID_INPUT_FILE/);
  assert.equal(outcome.reads.returned, limit + 1);
  assert.equal(outcome.reads.requested, limit + 1);
  assert.equal(outcome.reads.closed, true);

  outcome = probe('{}', { growTo: 2 * limit });
  assert.equal(outcome.reads.statSize, 2);
  assert.match(outcome.result.stderr, /INVALID_INPUT_FILE/);
  assert.equal(outcome.reads.returned, limit + 1);
  assert.equal(outcome.reads.requested, limit + 1);
  assert.equal(outcome.reads.closed, true);

  outcome = probe('{}', { fail: true });
  assert.match(outcome.result.stderr, /EIO/);
  assert.equal(outcome.reads.closed, true);

  const unicode = ' '.repeat(65_534) + '"é"';
  outcome = probe(unicode + ' '.repeat(limit - Buffer.byteLength(unicode)));
  assert.match(outcome.result.stderr, /UNKNOWN_CONTRACT/);
  assert.equal(outcome.reads.returned, limit);

  outcome = probe(Buffer.concat([Buffer.from('"' + ' '.repeat(limit - 2)),
    Buffer.from([0xc3])]));
  assert.match(outcome.result.stderr, /ERR_ENCODING_INVALID_ENCODED_DATA/);
  assert.doesNotMatch(outcome.result.stderr, /MALFORMED_JSON/);
  assert.equal(outcome.reads.returned, limit);
});

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
