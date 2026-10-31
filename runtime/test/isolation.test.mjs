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

test('cancellation during input reads prevents provider dispatch and confirms cleanup only after settlement', async () => {
  const state = fixture();
  const reads = Promise.withResolvers();
  state.workspace.readFiles = () => reads.promise;
  let invocations = 0;
  const adapter = createIsolatedAdapter({ workerId: state.attempt.worker_id, workspace: state.workspace,
    invoke: async () => { invocations += 1; return canonical(state.plan); } });
  const running = adapter.runAttempt({ attempt: state.attempt, spec: state.spec });
  const rejected = assert.rejects(running, { code: 'ATTEMPT_ABORTED' });
  const revoke = () => adapter.revokeCapabilities({ attempt_id: state.attempt.attempt_id,
    fencing_token: state.attempt.fencing_token, attempt: state.attempt });
  assert.deepEqual(await revoke(), { stopped: false });
  reads.resolve([]);
  await rejected;
  assert.equal(invocations, 0);
  assert.equal(state.applied.length, 0);
  assert.deepEqual(await revoke(), { stopped: true });
});

test('cancellation before the queued provider callback prevents invocation', async () => {
  const state = fixture();
  let invocations = 0;
  const adapter = createIsolatedAdapter({ workerId: state.attempt.worker_id, workspace: state.workspace,
    invoke: async () => { invocations += 1; return canonical(state.plan); } });
  const running = adapter.runAttempt({ attempt: state.attempt, spec: state.spec });
  const rejected = assert.rejects(running, { code: 'ATTEMPT_ABORTED' });
  // readFiles has resolved; its continuation queues invoke after this microtask.
  const revoked = new Promise((resolve, reject) => queueMicrotask(() => {
    adapter.revokeCapabilities({ attempt_id: state.attempt.attempt_id,
      fencing_token: state.attempt.fencing_token, attempt: state.attempt }).then(resolve, reject);
  }));
  await Promise.all([rejected, revoked]);
  assert.equal(invocations, 0);
  assert.equal(state.applied.length, 0);
});

// Exercise the real host control path without requiring Docker. Only the Docker
// executable and attached process launch are replaced; inspect/ownership/cleanup
// still pass through createDockerIsolationHost.
async function dockerControlFixture(t, onStart) {
  const { mkdtemp, writeFile, rm } = await import('node:fs/promises');
  const { readFileSync, writeFileSync, existsSync } = await import('node:fs');
  const { tmpdir } = await import('node:os');
  const { join } = await import('node:path');
  const { createDockerIsolationHost } = await import('../src/isolation.mjs');
  const root = await mkdtemp(join(tmpdir(), 'foundry-docker-control-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const statePath = join(root, 'container.json');
  const executable = join(root, 'docker');
  await writeFile(executable, `#!${process.execPath}
import fs from 'node:fs';
const file = ${JSON.stringify(statePath)};
const args = process.argv.slice(2);
const load = () => JSON.parse(fs.readFileSync(file, 'utf8'));
const save = (record) => fs.writeFileSync(file, JSON.stringify(record));
const output = (value) => process.stdout.write(typeof value === 'string' ? value : JSON.stringify(value));
if (args[0] === 'info') output('27.5.1');
else if (args[0] === 'image') output([{ RepoDigests: [args[2]], Config: { Env: [] } }]);
else if (args[0] === 'create') {
  const labels = {};
  for (let i = 0; i < args.length; i += 1) if (args[i] === '--label') {
    const [key, value] = args[++i].split('='); labels[key] = value;
  }
  const mount = args[args.indexOf('--mount') + 1];
  const record = {
    Id: 'a'.repeat(64),
    Config: { Image: args.find((arg) => arg.includes('@sha256:')), User: '1000:1000', Env: [], Labels: labels },
    HostConfig: { NetworkMode: 'none', ReadonlyRootfs: true, Privileged: false, PidsLimit: 32,
      Memory: 268435456, NanoCpus: 1000000000, CapDrop: ['ALL'], CapAdd: [],
      SecurityOpt: ['no-new-privileges:true'], Devices: [] },
    Mounts: [{ Type: 'bind', Source: mount.match(/src=([^,]+)/)[1], Destination: '/workspace', RW: true }],
    State: { Running: false, Status: 'created', ExitCode: 0 },
  };
  save(record); output(record.Id);
} else if (args[0] === 'inspect') output([load()]);
else if (args[0] === 'start') {
  const record = load(); record.State.Status = 'exited'; save(record); output('FOUNDRY_PROBE_OK');
} else if (args[0] === 'kill') {
  // Do not simulate successful termination: negative tests must remain blocked.
  process.exitCode = 1;
} else if (args[0] === 'rm') fs.unlinkSync(file);
else process.exitCode = 2;
`, { mode: 0o700 });
  const read = () => JSON.parse(readFileSync(statePath, 'utf8'));
  const write = (record) => writeFileSync(statePath, JSON.stringify(record));
  const imageDigest = `node@sha256:${'a'.repeat(64)}`;
  const host = createDockerIsolationHost({ dockerExecutable: executable, imageDigest,
    spawnImpl() { onStart?.({ read, write }); throw new Error('invented attached-process spawn failure'); } });
  const verified = await host.verify({ workspaceRoot: root, imageDigest, executionId: 'execution-test', attemptId: 'attempt-test' });
  const { plan, workspace } = fixture();
  const apply = () => host.applyPlan({ workspaceRoot: root, workspaceId, imageDigest, plan,
    guard: { verified: true, workspace: workspaceId, execution_id: workspace.execution_id,
      spec_digest: workspace.spec_digest, attempt_id: workspace.attempt_id,
      repository_revision: revision, write_scope: workspace.write_scope }, verificationId: verified.verificationId });
  const revoke = () => host.revokeCapabilities({ executionId: workspace.execution_id, attemptId: workspace.attempt_id,
    fencingToken: 1, verificationId: verified.verificationId });
  return { host, root, read, write, apply, revoke, exists: () => existsSync(statePath) };
}

test('Docker start spawn failure cleans up an inspected never-started container', async (t) => {
  const state = await dockerControlFixture(t);
  await assert.rejects(state.apply(), { code: 'ISOLATION_DOCKER_ERROR' });
  assert.equal(state.exists(), false);
  assert.equal((await state.revoke()).stopped, true);
  assert.equal(await state.host.disposeWorkspace({ workspaceRoot: state.root }), true);
});

test('Docker cleanup preserves running and unowned outcomes, then removes a verified created container', async (t) => {
  const state = await dockerControlFixture(t, ({ read, write }) => {
    const record = read(); record.State = { Running: true, Status: 'running', ExitCode: 0 }; write(record);
  });
  await assert.rejects(state.apply(), { code: 'ISOLATION_OUTCOME_UNKNOWN' });
  assert.equal((await state.revoke()).stopped, false);
  assert.equal(state.exists(), true);
  assert.equal(await state.host.disposeWorkspace({ workspaceRoot: state.root }), false);
  const record = state.read();
  const labels = { ...record.Config.Labels };
  record.State = { Running: false, Status: 'created', ExitCode: 0 };
  record.Config.Labels['foundry.attempt'] = 'other-attempt'; state.write(record);
  assert.equal((await state.revoke()).stopped, false);
  assert.equal(state.exists(), true);
  record.Config.Labels = labels; state.write(record);
  assert.equal((await state.revoke()).stopped, true);
  assert.equal(state.exists(), false);
  assert.equal(await state.host.disposeWorkspace({ workspaceRoot: state.root }), true);
});
