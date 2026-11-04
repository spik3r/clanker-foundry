import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, realpath, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { openStore } from '../src/store.mjs';
import { drainOutbox } from '../src/runner.mjs';
import { FakeAdapter } from '../src/adapters.mjs';
import { taskSpec } from './fixtures.mjs';

// Independent preproduction oracle: invented checks, real durable store, no tools.
const start = '2026-11-05T00:00:00.000Z';
async function fixture(t, ids = ['pack-validation']) {
  const dir = await realpath(await mkdtemp(join(tmpdir(), 'foundry-check-oracle-')));
  const trusted = ids.map(check_id => ({ check_id, check_version: 'v1', runner_id: 'validator' }));
  const store = openStore(join(dir, 'runtime.sqlite'), { worker_ids: ['worker'], trusted_checks: trusted });
  t.after(async () => { store.close(); await rm(dir, { recursive: true, force: true }); });
  const spec = structuredClone(taskSpec);
  spec.acceptance_checks = ids; spec.workflow.stages[0].required_checks = ids;
  let clock = start, seq = 0;
  const send = (kind, payload) => store.execute({ schema_version: 1, command_id: `oracle-${++seq}`, execution_id: 'execution',
    expected_revision: store.get('execution')?.revision ?? 0, at: clock, kind, payload }, { actor_id: 'operator', role: 'owner' });
  send('ADMIT', { spec });
  send('ASSIGN', { attempt_id: 'attempt', worker_id: 'worker', input_revisions: [spec.repository.revision],
    workspace: 'read-only', capabilities: ['read'], lease_deadline: '2026-11-05T00:01:00.000Z' });
  await drainOutbox({ store, adapter: new FakeAdapter(), dispatcherId: 'oracle', now: () => clock, limit: 1 });
  assert.equal(store.get('execution').state, 'VALIDATING');
  return { store, trusted, send, now: () => clock, advance: value => { clock = value; } };
}
const run = (f, store, checks) => drainOutbox({ store, checks, dispatcherId: 'oracle', now: f.now, limit: 1 });

test('current prepared check receives owned input and records evidence', async t => {
  const f = await fixture(t); let calls = 0;
  const result = await run(f, f.store, [{ ...f.trusted[0], run(input) {
    calls++; assert.equal(input.execution.state, 'VALIDATING');
    assert.equal(input.artifacts.length, 1); assert.equal(typeof input.artifacts[0].content, 'string');
    input.spec.outcome = 'local mutation'; input.execution.spec.outcome = 'local mutation';
    return 'PASS';
  } }]);
  assert.equal(calls, 1); assert.equal(result.delivered.length, 1);
  assert.equal(f.store.get('execution').validations.length, 1);
  assert.equal(f.store.get('execution').spec.outcome, taskSpec.outcome);
});

test('check at exact validation deadline is never invoked', async t => {
  const f = await fixture(t); let calls = 0;
  f.advance(f.store.get('execution').validation_deadline);
  const checks = [{ ...f.trusted[0], run() { calls++; return 'PASS'; } }];
  const result = await run(f, f.store, checks);
  assert.equal(calls, 0); assert.equal(result.unknown[0]?.receipt?.code, 'ATTEMPT_EXPIRED');
  assert.equal(f.store.get('execution').validations.length, 0);
  await run(f, f.store, checks); assert.equal(calls, 0, 'UNKNOWN does not automatically retry');
  assert.equal(f.store.get('execution').resources_held, true);
});

for (const change of ['cancel', 'worker', 'fence', 'subject', 'spec']) {
  test(`fresh eligibility after artifact preparation rejects ${change}`, async t => {
    const f = await fixture(t); let prepared = false, calls = 0;
    const wrapped = { ...f.store,
      getArtifact(hash) {
        const content = f.store.getArtifact(hash);
        if (!prepared) { prepared = true; if (change === 'cancel') f.send('CANCEL', { code: 'OWNER_CANCELLED' }); }
        return content;
      },
      get(id) {
        const current = f.store.get(id);
        if (prepared && change !== 'cancel') {
          // Controlled projection changes test identity checks without corrupting SQLite.
          if (change === 'worker') current.attempt.worker_id = 'other-worker';
          if (change === 'fence') { current.fence++; current.attempt.fencing_token++; }
          if (change === 'subject') current.result.subject_digest = 'f'.repeat(64);
          if (change === 'spec') current.spec_digest = 'f'.repeat(64);
        }
        return current;
      },
    };
    const result = await run(f, wrapped, [{ ...f.trusted[0], run() { calls++; return 'PASS'; } }]);
    assert.equal(prepared, true); assert.equal(calls, 0);
    assert.equal(result.unknown[0]?.receipt?.status, 'ATTEMPT_CHANGED_DURING_CHECKS');
    assert.equal(f.store.get('execution').validations.length, 0);
    assert.equal(f.store.get('execution').resources_held, true, 'refusal does not invent cleanup');
  });
}

test('second check is suppressed when clock reaches deadline after first durable evidence', async t => {
  const f = await fixture(t, ['pack-validation', 'second-check']); const calls = [];
  const wrapped = { ...f.store, execute(command, actor) {
    const result = f.store.execute(command, actor);
    if (command.kind === 'RECORD_VALIDATION') f.advance(f.store.get('execution').validation_deadline);
    return result;
  } };
  const result = await run(f, wrapped, f.trusted.map(check => ({ ...check, run() { calls.push(check.check_id); return 'PASS'; } })));
  assert.deepEqual(calls, ['pack-validation']);
  assert.deepEqual(f.store.get('execution').validations.map(x => x.check_id), ['pack-validation']);
  assert.equal(result.unknown[0]?.receipt?.code, 'ATTEMPT_EXPIRED');
});

test('already invoked check crossing deadline cannot record late evidence', async t => {
  const f = await fixture(t); let calls = 0;
  const result = await run(f, f.store, [{ ...f.trusted[0], async run() {
    calls++; await Promise.resolve(); f.advance(f.store.get('execution').validation_deadline); return 'PASS';
  } }]);
  assert.equal(calls, 1); assert.equal(result.unknown.length, 1);
  assert.equal(f.store.get('execution').validations.length, 0);
});
