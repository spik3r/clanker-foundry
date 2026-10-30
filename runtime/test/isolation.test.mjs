import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { canonical, digest } from '../src/contract.mjs';
import { createIsolatedAdapter } from '../src/isolation.mjs';

const revision = 'a'.repeat(40);
const workspaceId = 'workspace-test';
const hash = (value) => createHash('sha256').update(value, 'utf8').digest('hex');

function fixture() {
  const spec = {
    task_id: 'CF-isolation-adapter-test', spec_revision: 1,
    repository: { repository_id: 'test/repository', revision },
    write_scope: ['src/file.txt'], integration_policy: 'isolated_write',
    workflow: { workflow_id: 'builder-v1', stages: [{ id: 'build', role: 'builder-developer' }] },
  };
  const specDigest = digest(spec);
  const workspace = {
    execution_id: 'execution-test', spec_digest: specDigest, attempt_id: 'attempt-test',
    workspace: workspaceId, repository_revision: revision, write_scope: spec.write_scope,
    base_digests: { 'src/file.txt': null },
    guard(proof) { assert.equal(proof.workspace, workspaceId); return proof; },
    async readFiles() { return []; },
    async applyPlan(plan) {
      applied.push(plan);
      return { operations: plan.operations.map((operation) => ({ path: operation.path, action: operation.action,
        after_digest: operation.action === 'write' ? hash(operation.content) : null })) };
    },
    async getArtifact() { return null; },
    holdArtifact() {}, releaseArtifact() { return true; },
    async revokeCapabilities() { return { stopped: true }; },
  };
  const applied = [];
  const attempt = { attempt_id: workspace.attempt_id, stage_id: 'build', worker_id: 'builder-test', fencing_token: 1,
    input_revisions: [], workspace: workspaceId, capabilities: ['read', 'analyze', 'write'] };
  const plan = { schema_version: 1, execution_id: workspace.execution_id, spec_digest: specDigest,
    attempt_id: attempt.attempt_id, repository_revision: revision, write_scope: spec.write_scope,
    operations: [{ path: 'src/file.txt', action: 'write', before_digest: null, content: 'candidate\n' }] };
  return { spec, workspace, attempt, plan, applied };
}

test('isolated adapter gives the builder exact plan identity and target before digests', async () => {
  const state = fixture();
  let request;
  const adapter = createIsolatedAdapter({ workerId: state.attempt.worker_id, workspace: state.workspace,
    invoke: async (input) => { request = input; return canonical(state.plan); } });
  const result = await adapter.runAttempt({ attempt: state.attempt, spec: state.spec });
  assert.deepEqual(request.plan_identity, {
    schema_version: 1, execution_id: state.workspace.execution_id, spec_digest: state.workspace.spec_digest,
    attempt_id: state.attempt.attempt_id, repository_revision: revision, write_scope: ['src/file.txt'],
  });
  assert.deepEqual(request.worker_contract.before_digests, { 'src/file.txt': null });
  assert.equal(state.applied.length, 1);
  assert.equal(result.kind, 'SUCCEEDED');
});

test('isolated adapter rejects oversized plan artifacts before the Docker host runs', async () => {
  const state = fixture();
  const adapter = createIsolatedAdapter({ workerId: state.attempt.worker_id, workspace: state.workspace,
    invoke: async () => ' '.repeat(1024 * 1024 + 1) });
  await assert.rejects(adapter.runAttempt({ attempt: state.attempt, spec: state.spec }), { code: 'INVALID_WORKSPACE_PLAN' });
  assert.equal(state.applied.length, 0);
});
