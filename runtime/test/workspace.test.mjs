import test from 'node:test';
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { chmod, lstat, mkdtemp, mkdir, readFile, readdir, realpath, rm, symlink, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { canonical, digest } from '../src/contract.mjs';
import { createCandidateWorkspace, validateWorkspacePlan } from '../src/workspace.mjs';

const hash = (text) => createHash('sha256').update(text, 'utf8').digest('hex');
function taskFor(revision) {
  return {
    task_id: 'CF-isolation-test',
    spec_revision: 1,
    repository: { repository_id: 'example/repo', revision },
    write_scope: ['src/base.txt', 'src/new.txt'],
    integration_policy: 'isolated_write',
    workflow: { workflow_id: 'build-v1', stages: [{ id: 'build', role: 'builder-developer' }] },
  };
}

async function repository(t, files = { 'src/base.txt': 'base bytes\n', 'README.md': 'not in workspace\n' }) {
  const root = await mkdtemp(path.join(os.tmpdir(), 'foundry-repository-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  for (const [name, content] of Object.entries(files)) {
    const file = path.join(root, name);
    await mkdir(path.dirname(file), { recursive: true });
    if (content?.symlink) await symlink(content.symlink, file);
    else await writeFile(file, content);
  }
  execFileSync('git', ['-C', root, 'init', '-q']);
  execFileSync('git', ['-C', root, 'config', 'user.name', 'Foundry Test']);
  execFileSync('git', ['-C', root, 'config', 'user.email', 'foundry-test@example.invalid']);
  execFileSync('git', ['-C', root, 'add', '--all']);
  execFileSync('git', ['-C', root, 'commit', '-q', '-m', 'pinned workspace fixture']);
  return { root: await realpath(root), revision: execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim() };
}

function fakeHost({ failApply = null } = {}) {
  const calls = { verify: [], apply: [], revoke: [] };
  const host = {
    async verify(input) { calls.verify.push(input); return { verified: true, verificationId: 'verified-attempt-1' }; },
    async applyPlan({ workspaceRoot, plan, guard }) {
      calls.apply.push({ workspaceRoot, plan, guard });
      if (failApply) throw Object.assign(new Error('mock outcome'), { code: failApply });
      const operations = [];
      for (const operation of plan.operations) {
        const target = path.join(workspaceRoot, ...operation.path.split('/'));
        await mkdir(path.dirname(target), { recursive: true });
        if (operation.action === 'delete') await rm(target);
        else await writeFile(target, operation.content);
        operations.push({ path: operation.path, action: operation.action, after_digest: operation.action === 'write' ? hash(operation.content) : null });
      }
      return { verified: true, execution_id: plan.execution_id, spec_digest: plan.spec_digest,
        attempt_id: plan.attempt_id, repository_revision: plan.repository_revision, operations };
    },
    async revokeCapabilities(input) { calls.revoke.push(input); return { stopped: true, container_id: 'mock-container' }; },
    async disposeWorkspace() { return true; },
  };
  return { host, calls };
}

async function setup(t, options = {}) {
  const repo = await repository(t, options.files);
  const fake = fakeHost(options);
  const spec = taskFor(repo.revision);
  const workspace = await createCandidateWorkspace({
    repositoryRoot: repo.root,
    repositoryRevision: repo.revision,
    executionId: 'exec-isolated-1',
    specDigest: digest(spec),
    attemptId: 'attempt-isolated-1',
    writeScope: ['src/base.txt', 'src/new.txt'],
    readPaths: ['src/base.txt'],
    host: fake.host,
    imageDigest: 'node@sha256:'.concat('a'.repeat(64)),
  });
  t.after(() => workspace.dispose().catch(() => {}));
  return { ...repo, ...fake, workspace, spec, specDigest: digest(spec) };
}

function planFor(revision, currentSpecDigest = 'b'.repeat(64), overrides = {}) {
  return {
    schema_version: 1,
    execution_id: 'exec-isolated-1',
    spec_digest: currentSpecDigest,
    attempt_id: 'attempt-isolated-1',
    repository_revision: revision,
    write_scope: ['src/base.txt', 'src/new.txt'],
    operations: [
      { path: 'src/base.txt', action: 'write', before_digest: hash('base bytes\n'), content: 'candidate bytes\n' },
      { path: 'src/new.txt', action: 'write', before_digest: null, content: 'new candidate\n' },
    ],
    ...overrides,
  };
}

function proof(workspace) {
  return {
    verified: true,
    execution_id: workspace.execution_id,
    spec_digest: workspace.spec_digest,
    attempt_id: workspace.attempt_id,
    workspace: workspace.workspace,
    repository_revision: workspace.repository_revision,
    write_scope: workspace.write_scope,
  };
}

test('candidate snapshot includes only the exact authorized committed paths', async (t) => {
  const state = await setup(t);
  await writeFile(path.join(state.root, 'src/base.txt'), 'uncommitted change\n');
  const input = await state.workspace.readFiles(['src/base.txt']);
  assert.equal(input[0].content, 'base bytes\n');
  assert.equal(input[0].digest, hash('base bytes\n'));
  const candidate = state.calls.verify[0].workspaceRoot;
  assert.deepEqual(await readdir(candidate), ['src']);
  assert.deepEqual(await readdir(path.join(candidate, 'src')), ['base.txt']);
  assert.equal(await readFile(path.join(candidate, 'src/base.txt'), 'utf8'), 'base bytes\n');
  assert.equal((await lstat(path.dirname(candidate))).mode & 0o077, 0);
  assert.equal(state.calls.verify.length, 1);
  assert.match(state.workspace.workspace, /^workspace-/);
  assert.deepEqual(state.workspace.base_digests, {
    'src/base.txt': hash('base bytes\n'),
    'src/new.txt': null,
  });
});

test('workspace guard binds the verified execution, spec, attempt, base and exact scope', async (t) => {
  const { workspace } = await setup(t);
  assert.equal(workspace.verified, true);
  assert.match(workspace.workspace, /^workspace-[0-9a-f-]{36}$/);
  assert.equal(workspace.guard(proof(workspace)).verified, true);
  for (const changes of [
    { execution_id: 'other' }, { spec_digest: 'c'.repeat(64) }, { attempt_id: 'other' },
    { repository_revision: 'f'.repeat(40) }, { write_scope: ['src/base.txt'] }, { workspace: '/tmp/workspace' },
  ]) assert.throws(() => workspace.guard({ ...proof(workspace), ...changes }), { code: 'ISOLATION_GUARD_MISMATCH' });
});

test('workspace plan schema rejects unsafe paths, extra fields, duplicate paths and invalid operation variants', () => {
  const plan = planFor('a'.repeat(40));
  assert.equal(validateWorkspacePlan(plan), plan);
  for (const changes of [
    { unsafe: { operations: [{ ...plan.operations[0], path: '../escape' }] } },
    { extra: { ignored: true } },
    { duplicate: { operations: [plan.operations[0], plan.operations[0]] } },
    { deleteContent: { operations: [{ ...plan.operations[0], action: 'delete' }] } },
  ]) {
    const [key, value] = Object.entries(changes)[0];
    assert.throws(() => validateWorkspacePlan({ ...plan, ...value }), { code: 'INVALID_WORKSPACE_PLAN' }, key);
  }
});

test('candidate applies only exact authorized operations and returns a plan artifact subject', async (t) => {
  const { root, revision, workspace, calls, spec, specDigest } = await setup(t);
  const assignedPlan = planFor(revision, specDigest);
  const invoke = async (input) => {
    assert.equal(Object.isFrozen(input), true);
    assert.equal(input.files[0].content, 'base bytes\n');
    return canonical(assignedPlan);
  };
  const { createIsolatedAdapter } = await import('../src/isolation.mjs');
  const adapter = createIsolatedAdapter({ workerId: 'builder-1', invoke, workspace, inputPaths: ['src/base.txt'] });
  const attempt = { attempt_id: 'attempt-isolated-1', stage_id: 'build', worker_id: 'builder-1', fencing_token: 8, input_revisions: [], workspace: workspace.workspace, capabilities: ['read', 'analyze', 'write'] };
  const result = await adapter.runAttempt({ attempt, spec });
  assert.equal(result.kind, 'SUCCEEDED');
  assert.deepEqual(result.artifacts.map((item) => item.path), ['workspace-plan.json']);
  assert.equal(result.artifacts[0].digest, hash(canonical(assignedPlan)));
  assert.equal(result.subject_digest, digest(result.artifacts));
  assert.equal(await adapter.getArtifact(result.artifacts[0].digest), canonical(assignedPlan));
  assert.equal(calls.apply.length, 1);
  assert.equal(await readFile(path.join(calls.apply[0].workspaceRoot, 'src/base.txt'), 'utf8'), 'candidate bytes\n');
  assert.equal((await adapter.revokeCapabilities({ attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, attempt })).stopped, true);
  assert.equal(calls.revoke.length, 0, 'settled Docker attempts are already inspected and stopped');
  assert.equal(adapter.releaseArtifact(result.artifacts[0].digest), true);
  const acceptedDigest = result.subject_digest;
  await workspace.promote({ plan: assignedPlan, proof: proof(workspace), ownerId: 'operator', expectedOwner: 'operator', acceptedSubjectDigest: acceptedDigest, targetRoot: root });
  assert.equal(await readFile(path.join(root, 'src/base.txt'), 'utf8'), 'candidate bytes\n');
  assert.equal(await readFile(path.join(root, 'src/new.txt'), 'utf8'), 'new candidate\n');
});

test('scope, duplicate-key, untrusted owner, wrong artifact and changed targets fail before promotion', async (t) => {
  const { root, revision, workspace, calls, spec, specDigest } = await setup(t);
  const { createIsolatedAdapter } = await import('../src/isolation.mjs');
  const attempt = { attempt_id: 'attempt-isolated-1', stage_id: 'build', worker_id: 'builder-1', fencing_token: 4, input_revisions: [], workspace: workspace.workspace, capabilities: ['read', 'analyze', 'write'] };
  const invalid = planFor(revision, specDigest, { operations: [{ path: 'README.md', action: 'write', before_digest: null, content: 'out of scope' }] });
  const outside = createIsolatedAdapter({ workerId: 'builder-1', workspace, inputPaths: [], invoke: async () => canonical(invalid) });
  await assert.rejects(outside.runAttempt({ attempt, spec }), { code: 'WRITE_SCOPE_VIOLATION' });
  assert.equal(calls.apply.length, 0);

  const dup = createIsolatedAdapter({ workerId: 'builder-1', workspace, inputPaths: [], invoke: async () => '{"a":1,"a":2}' });
  await assert.rejects(dup.runAttempt({ attempt, spec }), { code: 'INVALID_WORKSPACE_PLAN' });

  const plan = planFor(revision, specDigest);
  const subject = digest([{ path: 'workspace-plan.json', digest: hash(canonical(plan)), repository_revision: revision }]);
  await assert.rejects(workspace.promote({ plan, proof: proof(workspace), ownerId: 'worker', expectedOwner: 'operator', acceptedSubjectDigest: subject, targetRoot: root }), { code: 'PROMOTION_OWNER_REQUIRED' });
  await assert.rejects(workspace.promote({ plan, proof: proof(workspace), ownerId: 'operator', expectedOwner: 'operator', acceptedSubjectDigest: 'f'.repeat(64), targetRoot: root }), { code: 'PROMOTION_SUBJECT_MISMATCH' });
  await writeFile(path.join(root, 'src/base.txt'), 'changed after snapshot\n');
  await assert.rejects(workspace.promote({ plan, proof: proof(workspace), ownerId: 'operator', expectedOwner: 'operator', acceptedSubjectDigest: subject, targetRoot: root }), { code: 'PROMOTION_TARGET_CHANGED' });
  await assert.rejects(readFile(path.join(root, 'src/new.txt')), { code: 'ENOENT' });
});

test('snapshot rejects symlinks and protected metadata paths', async (t) => {
  const symlinkRepo = await repository(t, { 'src/link.txt': { symlink: '../outside' }, outside: 'not followed' });
  const fake = fakeHost();
  await assert.rejects(createCandidateWorkspace({ ...symlinkRepo, repositoryRoot: symlinkRepo.root, repositoryRevision: symlinkRepo.revision,
    executionId: 'e', attemptId: 'a', specDigest: 'b'.repeat(64), writeScope: ['src/link.txt'], readPaths: ['src/link.txt'], host: fake.host }), { code: 'WORKSPACE_SYMLINK_OR_SPECIAL_FILE' });

  const normal = await repository(t);
  await assert.rejects(createCandidateWorkspace({ ...normal, repositoryRoot: normal.root, repositoryRevision: normal.revision,
    executionId: 'e', attemptId: 'a', specDigest: 'b'.repeat(64), writeScope: ['.git/config'], host: fake.host }), { code: 'WORKSPACE_PATH_BLOCKED' });
});

test('isolated builder rejects read-only attempts before invoking the model', async (t) => {
  const { workspace, spec } = await setup(t);
  const { createIsolatedAdapter } = await import('../src/isolation.mjs');
  let invoked = false;
  const adapter = createIsolatedAdapter({ workerId: 'builder-1', workspace, invoke: async () => { invoked = true; return ''; } });
  const attempt = { attempt_id: workspace.attempt_id, stage_id: 'build', worker_id: 'builder-1', fencing_token: 1,
    input_revisions: [], workspace: workspace.workspace, capabilities: ['read', 'analyze'] };
  await assert.rejects(adapter.runAttempt({ attempt, spec }), { code: 'ATTEMPT_IDENTITY_MISMATCH' });
  assert.equal(invoked, false);
});

test('revocation aborts a pending provider before Docker receives a plan', async (t) => {
  const { workspace, spec, calls } = await setup(t);
  const { createIsolatedAdapter } = await import('../src/isolation.mjs');
  let markStarted;
  const started = new Promise((resolve) => { markStarted = resolve; });
  const adapter = createIsolatedAdapter({ workerId: 'builder-1', workspace, invoke: async (_request, { signal }) => {
    markStarted();
    return new Promise((_resolve, reject) => signal.addEventListener('abort', () => reject(Object.assign(new Error('aborted'), { code: 'ATTEMPT_ABORTED' })), { once: true }));
  } });
  const attempt = { attempt_id: workspace.attempt_id, stage_id: 'build', worker_id: 'builder-1', fencing_token: 1,
    input_revisions: [], workspace: workspace.workspace, capabilities: ['read', 'analyze', 'write'] };
  const running = adapter.runAttempt({ attempt, spec });
  await started;
  const revocation = await adapter.revokeCapabilities({ attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, attempt });
  await assert.rejects(running, { code: 'ATTEMPT_ABORTED' });
  assert.deepEqual(revocation, { stopped: true });
  assert.equal(calls.apply.length, 0);
});

test('revocation stays unconfirmed when a provider ignores abort', async (t) => {
  const { workspace, spec, calls } = await setup(t);
  const { createIsolatedAdapter } = await import('../src/isolation.mjs');
  let markStarted;
  const started = new Promise((resolve) => { markStarted = resolve; });
  const adapter = createIsolatedAdapter({ workerId: 'builder-1', workspace, invoke: async () => {
    markStarted();
    return new Promise(() => {});
  } });
  const attempt = { attempt_id: workspace.attempt_id, stage_id: 'build', worker_id: 'builder-1', fencing_token: 1,
    input_revisions: [], workspace: workspace.workspace, capabilities: ['read', 'analyze', 'write'] };
  void adapter.runAttempt({ attempt, spec }).catch(() => {});
  await started;
  const revocation = await adapter.revokeCapabilities({ attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, attempt });
  assert.deepEqual(revocation, { stopped: false });
  assert.equal(calls.apply.length, 0);
});
