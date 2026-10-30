import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, rm, symlink, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { FakeAdapter, createReadOnlyWorkspaceAdapter, createTextModelAdapter } from '../src/adapters.mjs';
import { digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

const attempt = {
  attempt_id: 'attempt-1', stage_id: 'review', worker_id: 'worker-1',
  input_revisions: ['0123456789abcdef0123456789abcdef01234567'], fencing_token: 3,
};

async function gitWorkspace(t, files = { 'notes.txt': 'read only' }) {
  const root = await mkdtemp(path.join(os.tmpdir(), 'foundry-git-workspace-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  for (const [name, content] of Object.entries(files)) {
    const target = path.join(root, name);
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, content);
  }
  execFileSync('git', ['-C', root, 'init', '-q']);
  execFileSync('git', ['-C', root, 'config', 'user.name', 'Runtime Test']);
  execFileSync('git', ['-C', root, 'config', 'user.email', 'runtime-test@example.invalid']);
  execFileSync('git', ['-C', root, 'add', '--all']);
  execFileSync('git', ['-C', root, 'commit', '-q', '-m', 'fixture']);
  const revision = execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim();
  return { root, revision };
}

test('text adapter hashes its bounded output and passes a frozen, tool-less request', async (t) => {
  const { root, revision } = await gitWorkspace(t);
  const workspace = createReadOnlyWorkspaceAdapter({ root, allowedPaths: ['notes.txt'], maxBytes: 32 });
  const spec = structuredClone(taskSpec);
  spec.repository.revision = revision;
  let invocation;
  const adapter = createTextModelAdapter({
    workerId: 'text-worker',
    workspace,
    inputPaths: ['notes.txt'],
    invoke: async (request) => {
      invocation = request;
      assert.equal(Object.isFrozen(request), true);
      assert.equal(Object.isFrozen(request.task.workflow.stages[0]), true);
      assert.equal(request.files[0].content, 'read only');
      return 'review output';
    },
  });
  const result = await adapter.runAttempt({ attempt, spec });
  assert.equal(result.kind, 'SUCCEEDED');
  assert.equal(result.artifacts[0].digest, createHash('sha256').update('review output').digest('hex'));
  assert.equal(result.artifacts[0].repository_revision, revision);
  assert.equal(result.spec_digest, digest(spec));
  assert.equal(await adapter.getArtifact(result.artifacts[0].digest), 'review output');
  assert.equal(invocation.attempt.worker_id, 'worker-1');
  assert.deepEqual(Object.keys(adapter).sort(), ['getArtifact', 'releaseArtifact', 'revokeCapabilities', 'runAttempt', 'workerId']);
});

test('text adapter rejects non-text and oversized output', async () => {
  const invalid = createTextModelAdapter({ workerId: 'w', invoke: async () => ({ text: 'no' }) });
  await assert.rejects(invalid.runAttempt({ attempt, spec: taskSpec }), { code: 'INVALID_MODEL_OUTPUT' });
  const oversized = createTextModelAdapter({ workerId: 'w', maxOutputBytes: 2, invoke: async () => 'long' });
  await assert.rejects(oversized.runAttempt({ attempt, spec: taskSpec }), { code: 'OUTPUT_TOO_LARGE' });
});

test('text adapter bounds aggregate immutable subject bytes before invocation', async () => {
  let invoked = false;
  const content = '123456789';
  const artifact = {
    path: 'builder.txt',
    digest: createHash('sha256').update(content).digest('hex'),
    repository_revision: taskSpec.repository.revision,
    content,
  };
  const manifest = [{ path: artifact.path, digest: artifact.digest, repository_revision: artifact.repository_revision }];
  const adapter = createTextModelAdapter({
    workerId: 'text-worker',
    maxInputBytes: 8,
    invoke: async () => { invoked = true; return 'unused'; },
  });
  await assert.rejects(adapter.runAttempt({
    attempt,
    spec: taskSpec,
    inputSubjects: [{ subject_digest: digest(manifest), artifacts: [artifact] }],
  }), { code: 'INPUT_TOO_LARGE' });
  assert.equal(invoked, false);
});

test('text adapter does not claim revocation until an invocation has settled', async () => {
  let finish;
  const adapter = createTextModelAdapter({
    workerId: 'text-worker',
    invoke: async (_request, { signal }) => new Promise((resolve) => {
      finish = resolve;
      signal.addEventListener('abort', () => {}, { once: true });
    }),
  });
  const running = adapter.runAttempt({ attempt, spec: taskSpec });
  await Promise.resolve();
  const pending = await adapter.revokeCapabilities({ attempt_id: attempt.attempt_id });
  assert.equal(pending.stopped, false);
  finish('settled output');
  await running;
  assert.equal((await adapter.revokeCapabilities({ attempt_id: attempt.attempt_id })).stopped, true);
  const fresh = createTextModelAdapter({ workerId: 'text-worker', invoke: async () => 'unused' });
  assert.equal((await fresh.revokeCapabilities({ attempt_id: attempt.attempt_id, attempt })).stopped, false);
});

test('text adapter applies backpressure and keeps undelivered artifacts until release', async () => {
  const adapter = createTextModelAdapter({
    workerId: 'text-worker',
    invoke: async (request) => `output:${request.attempt.attempt_id}`,
  });
  const attempts = Array.from({ length: 17 }, (_, index) => ({
    ...attempt,
    attempt_id: `capacity-${index}`,
  }));
  const firstSixteen = await Promise.all(attempts.slice(0, 16).map((item) => adapter.runAttempt({ attempt: item, spec: taskSpec })));
  let seventeenthSettled = false;
  const seventeenth = adapter.runAttempt({ attempt: attempts[16], spec: taskSpec }).then((result) => {
    seventeenthSettled = true;
    return result;
  });
  await Promise.resolve();
  assert.equal(seventeenthSettled, false);

  const firstDigest = firstSixteen[0].artifacts[0].digest;
  assert.equal(await adapter.getArtifact(firstDigest), 'output:capacity-0');
  assert.equal(adapter.releaseArtifact(firstDigest), true);
  const last = await seventeenth;
  assert.equal(await adapter.getArtifact(last.artifacts[0].digest), 'output:capacity-16');
  for (let index = 1; index < 16; index += 1) {
    const artifactDigest = firstSixteen[index].artifacts[0].digest;
    assert.equal(await adapter.getArtifact(artifactDigest), `output:capacity-${index}`);
  }
  for (const result of firstSixteen.slice(1)) adapter.releaseArtifact(result.artifacts[0].digest);
  adapter.releaseArtifact(last.artifacts[0].digest);
});

test('text adapter reference counts identical undelivered artifact bytes', async () => {
  const adapter = createTextModelAdapter({ workerId: 'text-worker', invoke: async () => 'same bytes' });
  const first = await adapter.runAttempt({ attempt: { ...attempt, attempt_id: 'same-1' }, spec: taskSpec });
  const second = await adapter.runAttempt({ attempt: { ...attempt, attempt_id: 'same-2' }, spec: taskSpec });
  const artifactDigest = first.artifacts[0].digest;
  assert.equal(second.artifacts[0].digest, artifactDigest);
  assert.equal(adapter.releaseArtifact(artifactDigest), true);
  assert.equal(await adapter.getArtifact(artifactDigest), 'same bytes');
  assert.equal(adapter.releaseArtifact(artifactDigest), true);
  assert.equal(await adapter.getArtifact(artifactDigest), null);
  assert.equal(adapter.releaseArtifact(artifactDigest), false);
});

test('read-only workspace adapter enforces pinned Git bytes, allowlist, symlink, and limits', async (t) => {
  const { root, revision } = await gitWorkspace(t, { 'docs/guide.md': 'safe text' });
  const outside = path.join(root, '..', 'outside.txt');
  await writeFile(outside, 'outside');
  await symlink(outside, path.join(root, 'linked.txt'));

  const reader = createReadOnlyWorkspaceAdapter({
    root,
    allowedPaths: ['docs/guide.md', 'linked.txt'],
    maxBytes: 32,
  });
  assert.equal(await reader.verifyBaseRevision(revision), revision);
  const files = await reader.readFiles(['docs/guide.md']);
  assert.equal(files[0].content, 'safe text');
  assert.equal(files[0].digest, createHash('sha256').update('safe text').digest('hex'));
  assert.deepEqual(Object.keys(reader).sort(), ['readFiles', 'verifyBaseRevision']);
  await assert.rejects(reader.readFiles(['not-allowed.txt']), { code: 'PATH_NOT_ALLOWED' });
  await assert.rejects(reader.readFiles(['linked.txt']), { code: 'UNSAFE_WORKSPACE' });
  await assert.rejects(reader.readFiles(['../outside.txt']), { code: 'INVALID_PATH' });

  const limited = createReadOnlyWorkspaceAdapter({ root, allowedPaths: ['docs/guide.md'], maxBytes: 4 });
  await limited.verifyBaseRevision(revision);
  await assert.rejects(limited.readFiles(['docs/guide.md']), { code: 'INPUT_TOO_LARGE' });
  const wrongBase = createReadOnlyWorkspaceAdapter({ root, allowedPaths: ['docs/guide.md'] });
  await assert.rejects(wrongBase.verifyBaseRevision('f'.repeat(40)), { code: 'BASE_REVISION_MISMATCH' });
});

test('fake adapter returns a deterministic result tied to read input bytes and revokes its attempt', async () => {
  const adapter = new FakeAdapter({
    output: 'fake',
    inputPaths: ['input.txt'],
    readFiles: async () => [{ path: 'input.txt', content: 'abc', digest: createHash('sha256').update('abc').digest('hex') }],
  });
  const first = await adapter.runAttempt({ attempt, spec: taskSpec });
  const secondAttempt = { ...attempt, attempt_id: 'attempt-2' };
  const second = await adapter.runAttempt({ attempt: secondAttempt, spec: taskSpec });
  assert.equal(first.artifacts[0].digest, second.artifacts[0].digest);
  assert.equal(first.subject_digest, second.subject_digest);
  assert.notEqual(first.artifacts[0].digest, createHash('sha256').update('fake').digest('hex'));
  assert.deepEqual(await adapter.revokeCapabilities({ attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, attempt }), {
    stopped: true,
    reason: 'Fake attempt stopped',
  });
  assert.equal(adapter.activeAttempts.has(attempt.attempt_id), false);
  assert.equal((await adapter.revokeCapabilities({
    attempt_id: secondAttempt.attempt_id,
    fencing_token: secondAttempt.fencing_token,
    attempt: secondAttempt,
  })).stopped, true);
  assert.equal(await adapter.getArtifact(first.artifacts[0].digest), 'fake\ninput.txt:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad');
});

test('fake adapter confirms revocation only for the exact attempt it owns', async () => {
  const fresh = new FakeAdapter();
  const unknown = await fresh.revokeCapabilities({ attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, attempt });
  assert.equal(unknown.stopped, false);
  assert.equal(fresh.calls.length, 0);

  const adapter = new FakeAdapter();
  await adapter.runAttempt({ attempt, spec: taskSpec });
  await assert.rejects(adapter.runAttempt({ attempt, spec: taskSpec }), { code: 'DUPLICATE_ATTEMPT' });

  const wrongWorkerAttempt = { ...attempt, worker_id: 'other-worker' };
  assert.equal((await adapter.revokeCapabilities({
    attempt_id: attempt.attempt_id,
    fencing_token: attempt.fencing_token,
    attempt: wrongWorkerAttempt,
  })).stopped, false);
  assert.equal((await adapter.revokeCapabilities({
    attempt_id: attempt.attempt_id,
    fencing_token: attempt.fencing_token + 1,
    attempt,
  })).stopped, false);
  assert.equal(adapter.activeAttempts.has(attempt.attempt_id), true);
  assert.equal(adapter.calls.filter((call) => call.type === 'revoke').length, 0);

  assert.deepEqual(await adapter.revokeCapabilities({
    attempt_id: attempt.attempt_id,
    fencing_token: attempt.fencing_token,
    attempt,
  }), { stopped: true, reason: 'Fake attempt stopped' });
  assert.equal(adapter.activeAttempts.has(attempt.attempt_id), false);
  assert.equal((await adapter.revokeCapabilities({
    attempt_id: attempt.attempt_id,
    fencing_token: attempt.fencing_token,
    attempt,
  })).stopped, false);
});
