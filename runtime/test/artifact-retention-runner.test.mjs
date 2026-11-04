import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { chmodSync, mkdtempSync, realpathSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createTextModelAdapter, FakeAdapter } from '../src/adapters.mjs';
import { openStore } from '../src/store.mjs';
import { drainOutbox } from '../src/runner.mjs';
import { digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

// Post-freeze integration controls, added after the initial nine-case oracle and
// implementation draft. These check the approved persistence/ownership contract.
const owner = { actor_id: 'operator', role: 'owner' };
const workerId = 'retention-worker';
const at = '2026-11-04T00:00:00.000Z';
const sha = (value) => createHash('sha256').update(value).digest('hex');
function fixture(t) {
  const dir = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-retention-runner-')));
  chmodSync(dir, 0o700);
  const store = openStore(join(dir, 'authority.sqlite'), { worker_ids: [workerId] });
  t.after(() => { store.close(); rmSync(dir, { recursive: true, force: true }); });
  let sequence = 0;
  const send = (id, kind, payload) => store.execute({ schema_version: 1,
    execution_id: id, command_id: `retention-runner-${++sequence}`,
    expected_revision: store.get(id)?.revision ?? 0, at, kind, payload }, owner);
  const assign = (id) => {
    const spec = structuredClone(taskSpec);
    spec.task_id = `CF-retention-runner-${id}`;
    spec.acceptance_checks = [];
    spec.workflow.stages = [{ id: 'inspect', role: 'explorer-researcher', required_checks: [], requires_approval: false }];
    send(id, 'ADMIT', { spec });
    return send(id, 'ASSIGN', { attempt_id: `attempt-${id}`, worker_id: workerId,
      input_revisions: [], workspace: 'read-only', capabilities: ['read'],
      lease_deadline: '2026-11-04T00:01:00.000Z' }).execution.attempt;
  };
  const drain = (adapter, source = store) => drainOutbox({ store: source, adapter,
    dispatcherId: 'retention-runner-dispatcher', now: at, limit: 8 });
  return { store, send, assign, drain };
}

for (const kind of ['text', 'fake']) {
  test(`${kind}: runner releases persisted B rather than older UNKNOWN A with identical bytes`, async (t) => {
    const f = fixture(t);
    const text = 'shared runner output';
    const adapter = kind === 'text'
      ? createTextModelAdapter({ workerId, invoke: async () => text })
      : new FakeAdapter({ output: text });
    f.assign('A');
    const broken = { ...f.store, putArtifact() { throw Object.assign(new Error('invented disk full'), { code: 'ENOSPC' }); } };
    const failed = await f.drain(adapter, broken);
    assert.equal(failed.unknown[0].receipt.code, 'ENOSPC');
    assert.equal(await adapter.getArtifact(sha(text)), text);
    f.assign('B');
    await f.drain(adapter);
    assert.equal(f.store.get('B').state, 'VALIDATING');
    assert.equal(f.store.getArtifact(sha(text)), text);
    assert.equal(await adapter.getArtifact(sha(text)), text, 'A remains retained while its outcome is unknown');
    f.send('A', 'CANCEL', { code: 'OWNER_CANCELLED' });
    await f.drain(adapter);
    assert.equal(f.store.get('A').cleanup_pending, false);
    assert.equal(await adapter.getArtifact(sha(text)), null, 'B already released its own reference; A cleanup leaves none');
    f.send('B', 'CANCEL', { code: 'OWNER_CANCELLED' });
    await f.drain(adapter);
    assert.equal(await adapter.getArtifact(sha(text)), null);
  });
}

test('runner releases only individually persisted artifacts with the exact assigned attempt', async (t) => {
  const f = fixture(t);
  const assigned = f.assign('partial');
  const first = 'first persisted output', second = 'second unpersisted output';
  const releases = [];
  const retained = new Map([[sha(first), first], [sha(second), second]]);
  const adapter = {
    workerId,
    async runAttempt({ attempt, spec }) {
      const artifacts = [first, second].map((text, index) => ({ path: `output-${index}.txt`,
        digest: sha(text), repository_revision: spec.repository.revision }));
      return { schema_version: 1, attempt_id: attempt.attempt_id,
        fencing_token: attempt.fencing_token, spec_digest: digest(spec), kind: 'SUCCEEDED',
        artifacts, subject_digest: digest(artifacts) };
    },
    async getArtifact(hash) { return retained.get(hash) ?? null; },
    releaseArtifact(hash, attempt) {
      releases.push({ digest: hash, attempt: structuredClone(attempt), persisted: f.store.getArtifact(hash) });
      return retained.delete(hash);
    },
  };
  const partiallyFailing = { ...f.store, putArtifact(artifact) {
    if (artifact.digest === sha(second)) throw Object.assign(new Error('invented disk full'), { code: 'ENOSPC' });
    return f.store.putArtifact(artifact);
  } };
  const failed = await f.drain(adapter, partiallyFailing);
  assert.equal(failed.unknown[0].receipt.code, 'ENOSPC');
  assert.equal(f.store.get('partial').state, 'EXECUTING');
  assert.equal(f.store.getArtifact(sha(first)), first);
  assert.equal(f.store.getArtifact(sha(second)), null);
  assert.equal(await adapter.getArtifact(sha(first)), null);
  assert.equal(await adapter.getArtifact(sha(second)), second);
  assert.deepEqual(releases, [{ digest: sha(first), attempt: structuredClone(assigned), persisted: first }]);
});
