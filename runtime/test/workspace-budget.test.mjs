import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { mkdtemp, realpath, rm, mkdir, writeFile, readFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { createCandidateWorkspace } from '../src/workspace.mjs';

const realGit = execFileSync('which', ['git'], { encoding: 'utf8' }).trim();
const quote = text => `'${text.replaceAll("'", "'\\''")}'`;
const hash = bytes => createHash('sha256').update(bytes).digest('hex');

async function fixture(t, contents) {
  const root = await realpath(await mkdtemp(path.join(os.tmpdir(), 'foundry-candidate-budget-')));
  t.after(() => rm(root, { recursive: true, force: true }));
  const repo = path.join(root, 'repo');
  await mkdir(repo);
  const git = args => execFileSync(realGit, ['-C', repo, ...args], { encoding: 'utf8', timeout: 10_000 });
  git(['init', '-q']);
  git(['config', 'user.name', 'Budget Fixture']);
  git(['config', 'user.email', 'fixture@example.invalid']);
  for (const [name, bytes] of Object.entries(contents)) await writeFile(path.join(repo, name), bytes);
  // This unselected blob exceeds every small content budget in the table.
  await writeFile(path.join(repo, 'unselected-metadata-entry.txt'), 'unselected private fixture bytes'.repeat(40));
  git(['add', '--all']);
  git(['commit', '-qm', 'Owned candidate budget fixture']);
  const revision = git(['rev-parse', 'HEAD']).trim();
  const sourceTree = git(['rev-parse', 'HEAD^{tree}']).trim();
  const unselected = git(['rev-parse', 'HEAD:unselected-metadata-entry.txt']).trim();
  const log = path.join(root, 'git.log');
  const bin = path.join(root, 'bin');
  await mkdir(bin);
  await writeFile(log, '');
  await writeFile(path.join(bin, 'git'), `#!/bin/sh\nprintf '%s\\n' "$*" >> ${quote(log)}\nexec ${quote(realGit)} "$@"\n`, { mode: 0o755 });
  return { repo, git, revision, sourceTree, log, bin, unselected };
}

async function candidate(f, options) {
  let verifies = 0;
  const previousPath = process.env.PATH;
  process.env.PATH = f.bin;
  try {
    const workspace = await createCandidateWorkspace({ repositoryRoot: f.repo, repositoryRevision: f.revision,
      executionId: 'budget-execution', attemptId: 'budget-attempt', specDigest: '0'.repeat(64),
      writeScope: ['output.txt'], readPaths: ['input.txt'], ...options,
      host: { async verify() { verifies++; return { verified: true, verificationId: 'owned-fake' }; },
        async applyPlan() { assert.fail('must not apply'); }, async revokeCapabilities() {},
        async disposeWorkspace() { return true; } },
    });
    return { workspace, verifies };
  } catch (error) {
    assert.equal(verifies, 0, 'refusal must occur before host verification');
    throw error;
  } finally { process.env.PATH = previousPath; }
}

const fitting = [
  ['one byte', 1, Buffer.from('x')],
  ['four ASCII bytes', 4, Buffer.from('abcd')],
  ['sixteen bytes', 16, Buffer.from('0123456789abcdef')],
  ['empty content', 1, Buffer.alloc(0)],
  ['astral character', 4, Buffer.from([0xf0, 0x9f, 0x99, 0x82])],
  ['BOM plus ASCII', 4, Buffer.from([0xef, 0xbb, 0xbf, 0x61])],
];
for (const [label, maxBytes, bytes] of fitting) {
  test(`candidate content budget admits ${label} without charging tree metadata`, async t => {
    const f = await fixture(t, { 'input.txt': bytes });
    const { workspace, verifies } = await candidate(f, { maxBytes });
    try {
      assert.equal(verifies, 1);
      assert.deepEqual(await workspace.readFiles(['input.txt']), [{ path: 'input.txt',
        content: bytes.toString('utf8'), digest: hash(bytes) }]);
      assert.equal(f.git(['rev-parse', 'HEAD^{tree}']).trim(), f.sourceTree);
      assert.ok(!(await readFile(f.log, 'utf8')).includes(`cat-file blob ${f.unselected}`));
    } finally { await workspace.dispose(); }
  });
}

for (const overflow of [false, true]) {
  test(`candidate content budget ${overflow ? 'refuses overflow of' : 'admits exact'} selected aggregate`, async t => {
    const f = await fixture(t, { 'input.txt': 'ab', 'second.txt': overflow ? 'cde' : 'cd' });
    if (overflow) {
      await assert.rejects(candidate(f, { maxBytes: 4, readPaths: ['input.txt', 'second.txt'] }), { code: 'WORKSPACE_TOO_LARGE' });
    } else {
      const { workspace, verifies } = await candidate(f, { maxBytes: 4, readPaths: ['input.txt', 'second.txt'] });
      try {
        assert.equal(verifies, 1);
        assert.deepEqual((await workspace.readFiles(['input.txt', 'second.txt'])).map(file => file.content), ['ab', 'cd']);
      } finally { await workspace.dispose(); }
    }
  });
}

test('candidate content budget still refuses a single oversized blob through Git capture', async t => {
  const f = await fixture(t, { 'input.txt': 'abcde' });
  await assert.rejects(candidate(f, { maxBytes: 4 }), { code: 'WORKSPACE_GIT_ERROR' });
});

test('candidate content budget charges overlapping read and write-base selection once', async t => {
  const f = await fixture(t, { 'input.txt': 'abcd' });
  const { workspace, verifies } = await candidate(f, { maxBytes: 4, writeScope: ['input.txt'] });
  try {
    assert.equal(verifies, 1);
    assert.equal(workspace.base_digests['input.txt'], hash(Buffer.from('abcd')));
    assert.equal((await workspace.readFiles(['input.txt']))[0].content, 'abcd');
  } finally { await workspace.dispose(); }
});
