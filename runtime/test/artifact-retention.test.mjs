import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { chmodSync, mkdtempSync, realpathSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { createTextModelAdapter, FakeAdapter } from '../src/adapters.mjs';
import { openStore } from '../src/store.mjs';
import { drainOutbox } from '../src/runner.mjs';
import { taskSpec } from './fixtures.mjs';

// Independent contract frozen before the repair. All text and identities are invented.
const output = 'retention oracle shared output';
const artifactDigest = createHash('sha256').update(output).digest('hex');
const worker = 'retention-worker';
const at = '2026-11-04T00:00:00.000Z';
const owner = { actor_id: 'operator', role: 'owner' };
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
function deferred() { let resolve; const promise = new Promise((r) => { resolve = r; }); return { promise, resolve }; }
async function bounded(promise) {
  let timer;
  try { return await Promise.race([promise, new Promise((_, reject) => { timer = setTimeout(() => reject(new Error('bounded oracle wait expired')), 2000); })]); }
  finally { clearTimeout(timer); }
}
function attempt(id) {
  return { attempt_id: id, worker_id: worker, fencing_token: 1, stage_id: 'inspect',
    input_revisions: [], workspace: 'read-only', capabilities: ['read'],
    lease_deadline: '2026-11-04T00:01:00.000Z', attempt_number: 1 };
}
function spec(id = 'direct') {
  const value = structuredClone(taskSpec);
  value.task_id = `CF-retention-${id}`;
  value.acceptance_checks = [];
  value.workflow.stages = [{ id: 'inspect', role: 'explorer-researcher', required_checks: [], requires_approval: false }];
  return value;
}
function makeAdapter(kind, enter = async () => {}) {
  let calls = 0;
  const invoke = async () => { calls += 1; await enter(calls); return output; };
  const adapter = kind === 'text'
    ? createTextModelAdapter({ workerId: worker, invoke })
    : new FakeAdapter({ output, readFiles: async () => { await invoke(); return []; } });
  return { adapter, calls: () => calls };
}
const run = (adapter, value) => adapter.runAttempt({ attempt: value, spec: spec(), inputSubjects: [] });
const revoke = (adapter, value, extra = {}) => adapter.revokeCapabilities({ attempt_id: value.attempt_id, fencing_token: value.fencing_token, attempt: value, ...extra });
function releaseAll(adapter) { for (let n = 0; n < 64 && adapter.releaseArtifact(artifactDigest); n += 1) {} }

for (const kind of ['text', 'fake']) {
  test(`${kind}: scoped release and settled cleanup cannot steal shared references`, async () => {
    const { adapter } = makeAdapter(kind);
    const a = attempt('shared-a'), b = attempt('shared-b');
    try {
      await run(adapter, a); await run(adapter, b);
      for (const identity of [undefined, null, {}, { ...a, worker_id: 'wrong' }, { ...a, attempt_id: 'absent' }, { ...a, fencing_token: 2 }]) {
        assert.equal(adapter.releaseArtifact(artifactDigest, identity), false, 'explicit invalid scope releases nothing');
        assert.equal(await adapter.getArtifact(artifactDigest), output);
      }
      assert.equal((await revoke(adapter, a, { fencing_token: 2 })).stopped, false, 'outer fence must agree');
      assert.equal(adapter.releaseArtifact(artifactDigest, a), true);
      assert.equal(adapter.releaseArtifact(artifactDigest, a), false, 'one attempt owns only one retained reference');
      await revoke(adapter, a);
      assert.equal(await adapter.getArtifact(artifactDigest), output, 'cleanup after release must not steal B');
      assert.equal((await revoke(adapter, b)).stopped, true);
      assert.equal(await adapter.getArtifact(artifactDigest), null, 'matching settled cleanup discards its own output');
      await revoke(adapter, b);
      assert.equal(adapter.releaseArtifact(artifactDigest), false);
    } finally { releaseAll(adapter); }
  });

  test(`${kind}: legacy release followed by late cleanup preserves a later identical output`, async () => {
    const { adapter } = makeAdapter(kind);
    const a = attempt('legacy-a'), b = attempt('legacy-b');
    try {
      await run(adapter, a);
      assert.equal(adapter.releaseArtifact(artifactDigest), true, 'one-argument API remains supported');
      assert.equal(await adapter.getArtifact(artifactDigest), null);
      await run(adapter, b);
      await revoke(adapter, a);
      assert.equal(await adapter.getArtifact(artifactDigest), output);
      await revoke(adapter, b);
      assert.equal(await adapter.getArtifact(artifactDigest), null);
    } finally { releaseAll(adapter); }
  });

  test(`${kind}: unsettled revocation keeps capacity until the owned invocation settles`, async () => {
    const gate = deferred(), entered = deferred();
    const { adapter, calls } = makeAdapter(kind, async (n) => { if (n === 1) { entered.resolve(); await gate.promise; } });
    const a = attempt('active-a');
    const active = run(adapter, a);
    let queued;
    try {
      await bounded(entered.promise);
      for (let n = 0; n < 15; n += 1) await run(adapter, attempt(`settled-${n}`));
      let finished = false;
      queued = run(adapter, attempt('capacity-waiter')).then((value) => { finished = true; return value; });
      assert.equal((await revoke(adapter, a)).stopped, false);
      await delay(25);
      assert.equal(calls(), 16);
      assert.equal(finished, false, 'unsettled abort must not free capacity');
      gate.resolve(); await bounded(active);
      await delay(25);
      assert.equal(finished, false, 'settlement alone retains output');
      assert.equal((await revoke(adapter, a)).stopped, true);
      await bounded(queued);
      assert.equal(calls(), 17, 'confirmed settled cleanup frees one slot');
    } finally {
      gate.resolve(); releaseAll(adapter);
      await bounded(active).catch(() => {});
      if (queued) await bounded(queued).catch(() => {});
      releaseAll(adapter);
    }
  });

  test(`${kind}: UNKNOWN retains output; owner cleanup recovers capacity after persistence faults`, async () => {
    const directory = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-retention-oracle-')));
    chmodSync(directory, 0o700);
    const store = openStore(join(directory, 'authority.sqlite'), { worker_ids: [worker] });
    const { adapter, calls } = makeAdapter(kind);
    const broken = { ...store, putArtifact() { throw Object.assign(new Error('invented disk full'), { code: 'ENOSPC' }); } };
    let commands = 0, pending;
    const send = (id, commandKind, payload) => store.execute({ schema_version: 1, command_id: `oracle-${++commands}`,
      execution_id: id, expected_revision: store.get(id)?.revision ?? 0, at, kind: commandKind, payload }, owner);
    const assign = (n) => {
      const id = `retention-${n}`;
      send(id, 'ADMIT', { spec: spec(String(n)) });
      send(id, 'ASSIGN', { attempt_id: `attempt-${n}`, worker_id: worker, input_revisions: [], workspace: 'read-only',
        capabilities: ['read'], lease_deadline: '2026-11-04T00:01:00.000Z' });
      return id;
    };
    const drain = (source) => drainOutbox({ store: source, adapter, dispatcherId: 'retention-dispatcher', now: at, limit: 1 });
    try {
      for (let n = 0; n < 16; n += 1) {
        const id = assign(n);
        const failure = await bounded(drain(broken));
        assert.equal(failure.unknown[0]?.receipt?.code, 'ENOSPC');
        assert.equal(await adapter.getArtifact(artifactDigest), output, 'unknown persistence preserves output');
        if (n === 0) {
          await bounded(drain(store));
          assert.equal(calls(), 1, 'UNKNOWN effect does not relaunch');
          assert.equal(await adapter.getArtifact(artifactDigest), output);
        }
        send(id, 'CANCEL', { code: 'OWNER_CANCELLED' });
        await bounded(drain(store));
        assert.equal(store.get(id).cleanup_pending, false);
        assert.equal(store.get(id).resources_held, false);
        assert.equal(await adapter.getArtifact(artifactDigest), null, 'confirmed owned cleanup releases retained output');
      }
      const healthy = assign(16);
      pending = drain(store);
      await bounded(pending);
      assert.equal(store.get(healthy).state, 'VALIDATING');
      assert.equal(calls(), 17);
      assert.equal(await adapter.getArtifact(artifactDigest), null, 'durable persistence releases output');
    } finally {
      releaseAll(adapter);
      if (pending) await bounded(pending).catch(() => {});
      store.close(); rmSync(directory, { recursive: true, force: true });
    }
  });
}

test('text: legacy id-only revocation can stop but cannot discard retained output', async () => {
  const { adapter } = makeAdapter('text');
  const a = attempt('legacy-stop');
  try {
    await run(adapter, a);
    assert.equal((await adapter.revokeCapabilities({ attempt_id: a.attempt_id })).stopped, true);
    assert.equal(await adapter.getArtifact(artifactDigest), output);
    assert.equal(adapter.releaseArtifact(artifactDigest, a), true);
    assert.equal(await adapter.getArtifact(artifactDigest), null);
  } finally { releaseAll(adapter); }
});
