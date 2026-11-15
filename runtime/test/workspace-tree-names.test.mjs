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

async function fixture(t, names) {
  const root = await realpath(await mkdtemp(path.join(os.tmpdir(), 'foundry-tree-names-')));
  t.after(() => rm(root, { recursive: true, force: true }));
  const repo = path.join(root, 'repo');
  await mkdir(repo);
  const git = (args, input) => execFileSync(realGit, ['-C', repo, ...args], {
    input, timeout: 10_000, env: { ...process.env,
      GIT_AUTHOR_NAME: 'Tree Fixture', GIT_AUTHOR_EMAIL: 'fixture@example.invalid',
      GIT_COMMITTER_NAME: 'Tree Fixture', GIT_COMMITTER_EMAIL: 'fixture@example.invalid' },
  }).toString().trim();
  git(['init', '-q']);
  const contents = names.map((_, index) => Buffer.from(`owned content ${index}\n`));
  const records = names.map((name, index) => Buffer.concat([
    Buffer.from(`100644 blob ${git(['hash-object', '-w', '--stdin'], contents[index])}\t`),
    Buffer.from(name), Buffer.from([0]),
  ]));
  const tree = git(['mktree', '-z'], Buffer.concat(records));
  const revision = git(['commit-tree', tree, '-m', 'Owned tree-name fixture']);
  git(['update-ref', 'HEAD', revision]);
  const log = path.join(root, 'git.log');
  const bin = path.join(root, 'bin');
  await mkdir(bin);
  await writeFile(log, '');
  await writeFile(path.join(bin, 'git'), `#!/bin/sh\nprintf '%s\\n' "$*" >> ${quote(log)}\nexec ${quote(realGit)} "$@"\n`, { mode: 0o755 });
  const sourceTree = git(['rev-parse', 'HEAD^{tree}']);
  return { repo, revision, contents, log, bin, git, sourceTree };
}

async function create(f, reads, host) {
  const prior = process.env.PATH;
  process.env.PATH = f.bin;
  try {
    return await createCandidateWorkspace({ repositoryRoot: f.repo, repositoryRevision: f.revision,
      executionId: 'tree-execution', attemptId: 'tree-attempt', specDigest: '0'.repeat(64),
      writeScope: ['output.txt'], readPaths: reads, host });
  } finally { process.env.PATH = prior; }
}

const malformed = [
  ['invalid byte', Buffer.from([0xff])],
  ['invalid continuation byte', Buffer.from([0x80])],
  ['truncated sequence', Buffer.from([0xe2, 0x82])],
  ['encoded surrogate', Buffer.from([0xed, 0xa0, 0x80])],
];
for (const [label, bytes] of malformed) {
  for (const collision of [false, true]) {
    test(`candidate refuses ${label}${collision ? ' alongside a valid replacement name' : ''}`, async t => {
      const bad = Buffer.concat([Buffer.from('bad'), bytes, Buffer.from('.txt')]);
      const names = collision ? [Buffer.from('bad�.txt'), bad] : [bad];
      const f = await fixture(t, names);
      let verifies = 0;
      let applies = 0;
      const host = { async verify() { verifies++; return { verified: true, verificationId: 'fake' }; },
        async applyPlan() { applies++; }, async revokeCapabilities() {}, async disposeWorkspace() { return true; } };
      let workspace;
      try {
        await assert.rejects(async () => { workspace = await create(f, ['bad�.txt'], host); }, {
          code: 'WORKSPACE_GIT_ERROR', message: 'Pinned repository tree is not valid UTF-8',
        });
      } finally { if (workspace) await workspace.dispose(); }
      assert.equal(verifies, 0);
      assert.equal(applies, 0);
      assert.doesNotMatch(await readFile(f.log, 'utf8'), /cat-file/);
      assert.equal(f.git(['rev-parse', 'HEAD^{tree}']), f.sourceTree);
    });
  }
}

test('candidate refuses an unselected malformed tree name', async t => {
  const f = await fixture(t, [Buffer.from('safe.txt'), Buffer.from([0xff])]);
  let verifies = 0;
  let workspace;
  try {
    await assert.rejects(async () => { workspace = await create(f, ['safe.txt'], {
      async verify() { verifies++; return { verified: true, verificationId: 'fake' }; },
      async applyPlan() { assert.fail('must not apply'); }, async revokeCapabilities() {},
      async disposeWorkspace() { return true; },
    }); }, { code: 'WORKSPACE_GIT_ERROR', message: 'Pinned repository tree is not valid UTF-8' });
  } finally { if (workspace) await workspace.dispose(); }
  assert.equal(verifies, 0);
  assert.doesNotMatch(await readFile(f.log, 'utf8'), /cat-file/);
});

for (const name of ['literal�.txt', '🙂.txt', '日本語.txt', 'é.txt', 'e\u0301.txt', '\ufeffname.txt', 'tab\tname.txt', 'line\nname.txt']) {
  test(`candidate preserves exact valid Git name ${JSON.stringify(name)}`, async t => {
    const names = [name];
    const f = await fixture(t, names);
    let verifies = 0;
    const workspace = await create(f, names, {
      async verify() { verifies++; return { verified: true, verificationId: 'owned-fake' }; },
      async applyPlan() { assert.fail('must not apply'); }, async revokeCapabilities() {},
      async disposeWorkspace() { return true; },
    });
    try {
      assert.equal(verifies, 1);
      assert.deepEqual(await workspace.readFiles(names), names.map((name, index) => ({
        path: name, content: f.contents[index].toString('utf8'), digest: hash(f.contents[index]),
      })));
      assert.equal(f.git(['rev-parse', 'HEAD^{tree}']), f.sourceTree);
    } finally { await workspace.dispose(); }
  });
}
