import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { chmod, mkdtemp, realpath, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { createTextModelAdapter, FakeAdapter } from '../src/adapters.mjs';
import { createIsolatedAdapter } from '../src/isolation.mjs';
import { openStore } from '../src/store.mjs';
import { drainOutbox } from '../src/runner.mjs';
import { digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

// Independent preproduction contract. Real store tests use invented callbacks;
// direct isolated-adapter controls use a fake workspace, never a container.
const at = '2026-11-05T00:00:00.000Z';
const deadline = '2026-11-05T00:01:00.000Z';
const worker = 'eligibility-worker';
const owner = { actor_id: 'operator', role: 'owner' };
const output = 'invented eligible output';
const outputDigest = createHash('sha256').update(output).digest('hex');
function deferred() { let resolve; const promise = new Promise(r => { resolve = r; }); return { promise, resolve }; }
async function bounded(promise) {
  let timer;
  try { return await Promise.race([promise, new Promise((_, reject) => { timer = setTimeout(() => reject(new Error('bounded eligibility oracle wait expired')), 2000); })]); }
  finally { clearTimeout(timer); }
}
const refusal = () => { throw Object.assign(new Error('invented eligibility changed'), { code: 'ORACLE_REFUSED' }); };
const directAttempt = id => ({ attempt_id: id, worker_id: worker, fencing_token: 1, stage_id: 'inspect', input_revisions: [] });
async function stored(t) {
  const directory = await realpath(await mkdtemp(join(tmpdir(), 'foundry-eligibility-oracle-')));
  await chmod(directory, 0o700);
  const store = openStore(join(directory, 'runtime.sqlite'), { worker_ids: [worker] });
  t.after(async () => { store.close(); await rm(directory, { recursive: true, force: true }); });
  const spec = structuredClone(taskSpec);
  spec.task_id = 'CF-eligibility-oracle'; spec.acceptance_checks = [];
  spec.workflow.stages = [{ id: 'inspect', role: 'explorer-researcher', required_checks: [], requires_approval: false }];
  let sequence = 0, now = at;
  const send = (kind, payload) => store.execute({ schema_version: 1, command_id: `eligibility-${++sequence}`,
    execution_id: 'execution', expected_revision: store.get('execution')?.revision ?? 0, at: now, kind, payload }, owner);
  send('ADMIT', { spec });
  send('ASSIGN', { attempt_id: 'stored-attempt', worker_id: worker, input_revisions: [spec.repository.revision],
    workspace: 'read-only', capabilities: ['read'], lease_deadline: deadline });
  return { store, spec, send, clock: () => now, advance: value => { now = value; } };
}
for (const kind of ['text', 'fake']) {
  for (const change of ['current', 'cancel', 'lease', 'dispatch']) {
    test(`${kind}: actual store rechecks ${change} after awaited input preparation`, async t => {
      const f = await stored(t), entered = deferred(), gate = deferred();
      let invokes = 0, blocked = false, verifyCalls = 0;
      const readFiles = async () => { entered.resolve(); await gate.promise; return []; };
      const adapter = kind === 'text' ? createTextModelAdapter({ workerId: worker, inputPaths: ['invented.txt'],
        workspace: { async verifyBaseRevision() {}, readFiles }, invoke: async () => { invokes++; return output; } })
        : new FakeAdapter({ output, readFiles });
      const checkedStore = { ...f.store, verifyDispatch(...args) {
        verifyCalls++;
        if (blocked) throw Object.assign(new Error('invented private guard detail'), { code: 'SOURCE_NOT_CURRENT' });
        return f.store.verifyDispatch(...args);
      } };
      const drain = () => drainOutbox({ store: checkedStore, adapter, dispatcherId: 'eligibility-dispatcher', now: f.clock, limit: 1 });
      const pending = drain();
      try {
        await bounded(entered.promise);
        assert.equal(f.store.get('execution').state, 'EXECUTING');
        if (change === 'cancel') f.send('CANCEL', { code: 'OWNER_CANCELLED' });
        if (change === 'lease') f.advance(deadline); // Strictly before, not equal to, the lease.
        if (change === 'dispatch') blocked = true;
        gate.resolve();
        const result = await bounded(pending);
        const count = () => kind === 'text' ? invokes : adapter.calls.filter(x => x.type === 'run').length;
        if (change === 'current') {
          assert.equal(count(), 1); assert.equal(f.store.get('execution').state, 'VALIDATING');
          assert.equal(result.delivered.length, 1); return;
        }
        assert.equal(count(), 0, 'a stale preparation must not invoke or publish an adapter result');
        const code = change === 'cancel' ? 'ATTEMPT_CHANGED_BEFORE_INVOKE' : change === 'lease' ? 'ATTEMPT_EXPIRED' : 'SOURCE_NOT_CURRENT';
        assert.equal(result.unknown[0]?.receipt?.code, code);
        assert.equal(f.store.get('execution').resources_held, true, 'refusal does not invent cleanup');
        if (change === 'dispatch') assert.ok(verifyCalls >= 3, 'dispatch guard refreshed after awaited reads');
        blocked = false;
        await bounded(drain());
        assert.equal(count(), 0, 'UNKNOWN never automatically relaunches');
        if (change !== 'cancel') f.send('CANCEL', { code: 'OWNER_CANCELLED' });
        if (kind === 'text') {
          await bounded(drain());
          assert.equal(f.store.get('execution').cleanup_pending, false, 'owner cleanup remains a separate action');
        }
      } finally { gate.resolve(); await bounded(pending).catch(() => {}); }
    });
  }
}

test('text: queued capacity cannot reuse eligibility checked before the wait', async () => {
  let eligible = true, invoked = 0;
  const adapter = createTextModelAdapter({ workerId: worker, invoke: async () => { invoked++; return output; } });
  const run = id => adapter.runAttempt({ attempt: directAttempt(id), spec: taskSpec,
    beforeInvoke() { if (!eligible) refusal(); } });
  let pending;
  try {
    for (let n = 0; n < 16; n++) await run(`retained-${n}`);
    pending = run('queued');
    const rejected = assert.rejects(pending, { code: 'ORACLE_REFUSED' });
    eligible = false;
    assert.equal(adapter.releaseArtifact(outputDigest), true);
    await bounded(rejected);
    assert.equal(invoked, 16);
    eligible = true; await bounded(run('fresh'));
    assert.equal(invoked, 17, 'a refused queued invocation returns its capacity slot');
  } finally {
    for (let n = 0; n < 64 && adapter.releaseArtifact(outputDigest); n++) {}
    if (pending) await bounded(pending).catch(() => {});
    for (let n = 0; n < 64 && adapter.releaseArtifact(outputDigest); n++) {}
  }
});
function isolated(readFiles = async () => []) {
  const spec = structuredClone(taskSpec);
  spec.integration_policy = 'isolated_write'; spec.write_scope = ['result.txt'];
  spec.workflow = { workflow_id: 'isolated-write-v1', version: 1, stages: [{ id: 'build', role: 'builder-developer' }, { id: 'review', role: 'reviewer-verifier' }] };
  const attempt = { ...directAttempt('isolated-attempt'), stage_id: 'build', workspace: 'invented-workspace', capabilities: ['read', 'write'] };
  let applied = 0;
  const workspace = { execution_id: 'isolated-execution', spec_digest: digest(spec), attempt_id: attempt.attempt_id,
    workspace: attempt.workspace, repository_revision: spec.repository.revision, write_scope: spec.write_scope, base_digests: { 'result.txt': null },
    guard: proof => proof, readFiles,
    async applyPlan(plan) { applied++; return { operations: plan.operations.map(x => ({ path: x.path, action: x.action })) }; },
    async getArtifact() { return null; }, async revokeCapabilities() { return { stopped: true }; } };
  const plan = { schema_version: 1, execution_id: workspace.execution_id, spec_digest: workspace.spec_digest,
    attempt_id: attempt.attempt_id, repository_revision: spec.repository.revision, write_scope: spec.write_scope,
    operations: [{ path: 'result.txt', action: 'write', before_digest: null, content: 'invented plan' }] };
  return { spec, attempt, workspace, plan, applied: () => applied };
}
for (const phase of ['read', 'output']) {
  test(`isolated: changed eligibility during ${phase} suppresses the next side effect`, async () => {
    const entered = deferred(), gate = deferred(); let eligible = true, invokes = 0;
    const f = isolated(async () => { if (phase === 'read') { entered.resolve(); await gate.promise; } return []; });
    const adapter = createIsolatedAdapter({ workerId: worker, workspace: f.workspace, invoke: async () => {
      invokes++; if (phase === 'output') { entered.resolve(); await gate.promise; } return JSON.stringify(f.plan);
    } });
    const pending = adapter.runAttempt({ attempt: f.attempt, spec: f.spec, beforeInvoke() { if (!eligible) refusal(); } });
    const rejected = assert.rejects(pending, { code: 'ORACLE_REFUSED' });
    try { await bounded(entered.promise); eligible = false; gate.resolve(); await bounded(rejected);
      assert.equal(invokes, phase === 'read' ? 0 : 1); assert.equal(f.applied(), 0);
    } finally { gate.resolve(); await bounded(pending).catch(() => {}); }
  });
}
test('isolated: omitted guard preserves direct-call success', async () => {
  const f = isolated();
  const adapter = createIsolatedAdapter({ workerId: worker, workspace: f.workspace, invoke: async () => JSON.stringify(f.plan) });
  assert.equal((await adapter.runAttempt({ attempt: f.attempt, spec: f.spec })).kind, 'SUCCEEDED');
  assert.equal(f.applied(), 1);
});
for (const kind of ['text', 'fake', 'isolated']) {
  test(`${kind}: optional guard must be synchronous and return only undefined`, async () => {
    for (const guard of [false, () => false, () => Promise.resolve(), () => Promise.reject(new Error('invented rejected guard'))]) {
      let invokes = 0;
      const f = isolated();
      const adapter = kind === 'isolated' ? createIsolatedAdapter({ workerId: worker, workspace: f.workspace, invoke: async () => { invokes++; return JSON.stringify(f.plan); } })
        : kind === 'text' ? createTextModelAdapter({ workerId: worker, invoke: async () => { invokes++; return output; } }) : new FakeAdapter();
      const attempt = kind === 'isolated' ? f.attempt : directAttempt('invalid-guard');
      const spec = kind === 'isolated' ? f.spec : taskSpec;
      await assert.rejects(adapter.runAttempt({ attempt, spec, beforeInvoke: guard }), { code: 'INVALID_INVOCATION_GUARD' });
      assert.equal(invokes, 0); assert.equal(f.applied(), 0);
    }
  });
}
