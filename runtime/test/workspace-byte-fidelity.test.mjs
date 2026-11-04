import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { mkdtemp, writeFile, rm, realpath } from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';
import { createReadOnlyWorkspaceAdapter, createTextModelAdapter } from '../src/adapters.mjs';
import { createCandidateWorkspace } from '../src/workspace.mjs';
import { createIsolatedAdapter } from '../src/isolation.mjs';
import { digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

// Independently stated source literals, not decoded reader or provider outputs.
// A fake host supplies no OS-isolation evidence and runs no candidate process.
const vectors = [
  ['BOM only', [0xef, 0xbb, 0xbf], '\ufeff'],
  ['BOM and text', [0xef, 0xbb, 0xbf, 0x61], '\ufeffa'],
  ['plain', [0x61, 0x0a], 'a\n'],
  ['multibyte', [0xc3, 0xa9, 0xf0, 0x9f, 0x98, 0x80], 'é😀'],
  ['CRLF', [0x61, 0x0d, 0x0a], 'a\r\n'],
  ['literal replacement character', [0xef, 0xbf, 0xbd], '\ufffd'],
  ['interior BOM', [0x61, 0xef, 0xbb, 0xbf], 'a\ufeff'],
];
const hash = bytes => createHash('sha256').update(bytes).digest('hex');
const attempt = {
  attempt_id: 'byte-attempt', stage_id: 'review', worker_id: 'byte-worker',
  input_revisions: [], fencing_token: 1,
};

async function repository(t, files) {
  const root = await realpath(await mkdtemp(path.join(os.tmpdir(), 'foundry-byte-fixture-')));
  t.after(() => rm(root, { recursive: true, force: true }));
  for (const [name, bytes] of Object.entries(files)) await writeFile(path.join(root, name), bytes);
  const git = args => execFileSync('git', ['-C', root, ...args], { encoding: 'utf8', timeout: 10_000 });
  git(['init', '-q']);
  git(['config', 'user.name', 'Byte Fixture']);
  git(['config', 'user.email', 'byte-fixture@example.invalid']);
  git(['add', '--', ...Object.keys(files)]);
  git(['commit', '-qm', 'Freeze literal workspace bytes']);
  return { root, revision: git(['rev-parse', 'HEAD']).trim() };
}

async function candidate(t, repo, options = {}) {
  const spec = structuredClone(taskSpec);
  spec.repository.revision = repo.revision;
  spec.integration_policy = 'isolated_write';
  spec.write_scope = ['output.txt'];
  spec.workflow.stages = [{ id: 'review', role: 'builder-developer' }];
  const workspace = await createCandidateWorkspace({
    repositoryRoot: repo.root, repositoryRevision: repo.revision,
    executionId: 'byte-execution', attemptId: attempt.attempt_id,
    specDigest: digest(spec), writeScope: spec.write_scope,
    readPaths: ['input.txt'],
    host: {
      async verify() { return { verified: true, verificationId: 'invented-byte-proof' }; },
      async applyPlan() { assert.fail('byte-read checks must not apply a plan'); },
      async revokeCapabilities() { return { stopped: true }; },
      async disposeWorkspace() { return true; },
    },
    ...options,
  });
  t.after(() => workspace.dispose());
  return { workspace, spec };
}

for (const [name, source, text] of vectors) {
  for (const mode of ['read-only', 'candidate']) {
    test(`${mode} reader retains exact ${name} bytes and digest`, async t => {
      const bytes = Buffer.from(source);
      const repo = await repository(t, { 'input.txt': bytes });
      let workspace;
      if (mode === 'read-only') {
        workspace = createReadOnlyWorkspaceAdapter({ root: repo.root, allowedPaths: ['input.txt'] });
        await workspace.verifyBaseRevision(repo.revision);
      } else ({ workspace } = await candidate(t, repo));
      const files = await workspace.readFiles(['input.txt']);
      assert.deepEqual(files, [{ path: 'input.txt', content: text, digest: hash(bytes) }]);
      assert.deepEqual(Buffer.from(files[0].content, 'utf8'), bytes);
      assert.equal(hash(Buffer.from(files[0].content, 'utf8')), files[0].digest);
    });
  }
}

for (const mode of ['read-only', 'candidate']) {
  test(`${mode} adapter callback receives BOM source without byte drift`, async t => {
    const bytes = Buffer.from([0xef, 0xbb, 0xbf, 0x61, 0x0d, 0x0a]);
    const repo = await repository(t, { 'input.txt': bytes });
    let calls = 0;
    const stop = new Error('invented callback stops before any plan or host execution');
    const invoke = async request => {
      calls++;
      assert.deepEqual(request.files, [{ path: 'input.txt', content: '\ufeffa\r\n', digest: hash(bytes) }]);
      assert.equal(Object.isFrozen(request.files[0]), true);
      if (mode === 'candidate') throw stop;
      return 'byte fidelity observed';
    };
    if (mode === 'read-only') {
      const spec = structuredClone(taskSpec);
      spec.repository.revision = repo.revision;
      const workspace = createReadOnlyWorkspaceAdapter({ root: repo.root, allowedPaths: ['input.txt'] });
      const adapter = createTextModelAdapter({ workerId: attempt.worker_id, workspace, inputPaths: ['input.txt'], invoke });
      assert.equal((await adapter.runAttempt({ attempt, spec })).kind, 'SUCCEEDED');
    } else {
      const { workspace, spec } = await candidate(t, repo);
      const assigned = { ...attempt, workspace: workspace.workspace, capabilities: ['write'] };
      const adapter = createIsolatedAdapter({ workerId: attempt.worker_id, workspace, inputPaths: ['input.txt'], invoke });
      await assert.rejects(adapter.runAttempt({ attempt: assigned, spec }), error => error === stop);
    }
    assert.equal(calls, 1);
  });
}

test('read-only aggregate byte limit counts each source BOM', async t => {
  const repo = await repository(t, {
    'input.txt': Buffer.from([0xef, 0xbb, 0xbf]), 'other.txt': Buffer.from([0x61, 0x62]),
  });
  const workspace = createReadOnlyWorkspaceAdapter({ root: repo.root, allowedPaths: ['input.txt', 'other.txt'], maxBytes: 4 });
  await workspace.verifyBaseRevision(repo.revision);
  await assert.rejects(workspace.readFiles(['input.txt', 'other.txt']), { code: 'INPUT_TOO_LARGE' });
});

test('text adapter byte limit includes the retained BOM before invocation', async t => {
  const repo = await repository(t, { 'input.txt': Buffer.from([0xef, 0xbb, 0xbf, 0x61]) });
  const spec = structuredClone(taskSpec);
  spec.repository.revision = repo.revision;
  const workspace = createReadOnlyWorkspaceAdapter({ root: repo.root, allowedPaths: ['input.txt'] });
  let calls = 0;
  const adapter = createTextModelAdapter({ workerId: attempt.worker_id, workspace, inputPaths: ['input.txt'], maxInputBytes: 3,
    invoke: async () => { calls++; return 'unused'; } });
  await assert.rejects(adapter.runAttempt({ attempt, spec }), { code: 'INPUT_TOO_LARGE' });
  assert.equal(calls, 0);
});

test('candidate snapshot bound includes original BOM bytes', async t => {
  const bytes = Buffer.concat([Buffer.from([0xef, 0xbb, 0xbf]), Buffer.alloc(510, 0x61)]);
  const repo = await repository(t, { 'input.txt': bytes, 'other.txt': Buffer.alloc(512, 0x62) });
  await assert.rejects(candidate(t, repo, { maxBytes: 1024, readPaths: ['input.txt', 'other.txt'] }), { code: 'WORKSPACE_TOO_LARGE' });
});

for (const mode of ['read-only', 'candidate']) {
  test(`${mode} reader still rejects malformed UTF-8`, async t => {
    const repo = await repository(t, { 'input.txt': Buffer.from([0xef, 0xbb, 0xbf, 0xc3, 0x28]) });
    if (mode === 'read-only') {
      const workspace = createReadOnlyWorkspaceAdapter({ root: repo.root, allowedPaths: ['input.txt'] });
      await workspace.verifyBaseRevision(repo.revision);
      await assert.rejects(workspace.readFiles(['input.txt']), { code: 'INVALID_UTF8' });
    } else {
      const { workspace } = await candidate(t, repo);
      await assert.rejects(workspace.readFiles(['input.txt']), { code: 'ERR_ENCODING_INVALID_ENCODED_DATA' });
    }
  });
}
