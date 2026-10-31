import { execFileSync } from 'node:child_process';
import { chmod, mkdtemp, mkdir, readFile, realpath, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { canonical, digest } from '../src/contract.mjs';
import { createDockerIsolationHost, createIsolatedAdapter } from '../src/isolation.mjs';
import { createCandidateWorkspace, workspaceDefaults } from '../src/workspace.mjs';

const hash = (value) => createHash('sha256').update(value, 'utf8').digest('hex');
const imageDigest = workspaceDefaults.imageDigest;
let root;
let workspace;

try {
  const createdRoot = await mkdtemp(path.join(os.tmpdir(), 'foundry-isolation-live-'));
  root = await realpath(createdRoot);
  await chmod(root, 0o700);
  await mkdir(path.join(root, 'src'), { mode: 0o755 });
  await writeFile(path.join(root, 'src/live.txt'), 'pinned base bytes\n');
  execFileSync('git', ['-C', root, 'init', '-q']);
  execFileSync('git', ['-C', root, 'config', 'user.name', 'Foundry Isolation Smoke']);
  execFileSync('git', ['-C', root, 'config', 'user.email', 'foundry-isolation@example.invalid']);
  execFileSync('git', ['-C', root, 'add', '--all']);
  execFileSync('git', ['-C', root, 'commit', '-q', '-m', 'isolated candidate smoke fixture']);
  const revision = execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], { encoding: 'utf8' }).trim();
  const spec = {
    task_id: 'CF-isolation-live-smoke', spec_revision: 1,
    repository: { repository_id: 'local/isolation-smoke', revision },
    write_scope: ['src/live.txt', 'src/new.txt'],
    integration_policy: 'isolated_write',
    workflow: { workflow_id: 'builder-v1', stages: [{ id: 'build', role: 'builder-developer' }] },
  };
  const specDigest = digest(spec);
  const executionId = 'isolation-live-smoke';
  const attemptId = 'isolation-live-smoke-1';
  const host = createDockerIsolationHost({ imageDigest });
  workspace = await createCandidateWorkspace({ repositoryRoot: root, repositoryRevision: revision,
    executionId, specDigest, attemptId, writeScope: spec.write_scope, readPaths: ['src/live.txt'], imageDigest, host });
  const plan = {
    schema_version: 1, execution_id: executionId, spec_digest: specDigest, attempt_id: attemptId,
    repository_revision: revision, write_scope: spec.write_scope,
    operations: [
      { path: 'src/live.txt', action: 'write', before_digest: hash('pinned base bytes\n'), content: 'isolated Docker write\n' },
      { path: 'src/new.txt', action: 'write', before_digest: null, content: 'new isolated file\n' },
    ],
  };
  const adapter = createIsolatedAdapter({ workerId: 'live-smoke-worker', workspace, inputPaths: ['src/live.txt'], invoke: async (request) => {
    if (request.files[0]?.content !== 'pinned base bytes\n') throw new Error('Pinned input did not match');
    return canonical(plan);
  } });
  const attempt = { attempt_id: attemptId, stage_id: 'build', worker_id: 'live-smoke-worker',
    fencing_token: 1, input_revisions: [], workspace: workspace.workspace, capabilities: ['read', 'analyze', 'write'] };
  const result = await adapter.runAttempt({ attempt, spec });
  const content = await adapter.getArtifact(result.artifacts[0].digest);
  if (content !== canonical(plan) || result.kind !== 'SUCCEEDED') throw new Error('Adapter artifact did not match the approved file plan');
  await workspace.promote({ plan, proof: { verified: true, execution_id: executionId, spec_digest: specDigest,
    attempt_id: attemptId, workspace: workspace.workspace, repository_revision: revision, write_scope: spec.write_scope },
  ownerId: 'isolation-live-owner', expectedOwner: 'isolation-live-owner', acceptedSubjectDigest: result.subject_digest, targetRoot: root });
  if (await readFile(path.join(root, 'src/live.txt'), 'utf8') !== 'isolated Docker write\n'
    || await readFile(path.join(root, 'src/new.txt'), 'utf8') !== 'new isolated file\n') throw new Error('Owner promotion did not apply the candidate files');
  process.stdout.write(`${JSON.stringify({ status: 'PASS', isolation: 'docker-inspected', image_digest: imageDigest,
    operations: result.artifacts.length === 1 ? plan.operations.length : 0, artifacts: result.artifacts.length })}\n`);
} catch (error) {
  const code = typeof error?.code === 'string' ? error.code : 'ISOLATION_LIVE_FAILED';
  process.stderr.write(`${JSON.stringify({ status: 'FAIL', code })}\n`);
  process.exitCode = 1;
} finally {
  if (workspace) await workspace.dispose().catch(() => {});
  if (root) await rm(root, { recursive: true, force: true });
}
