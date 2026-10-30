import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash, randomUUID } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { chmodSync, mkdirSync, mkdtempSync, readFileSync, realpathSync, rmSync, statSync, symlinkSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { DatabaseSync } from 'node:sqlite';
import { createProjectMemoryBridge } from '../src/project-memory.mjs';
import { openStore } from '../src/store.mjs';
import { canonical, digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';
import { createGit, loadMemory, localMemory, writeImmutable } from '../../skills/project-memory/scripts/lib/git.mjs';
import { eventPath } from '../../skills/project-memory/scripts/lib/state.mjs';

const helper = fileURLToPath(new URL('../../skills/project-memory/scripts/project-memory.mjs', import.meta.url));
const owner = { actor_id: 'operator', role: 'owner' };
const repository = 'example/evidence';
const task = 'CF-evidence';
const sourcePath = `tasks/${task}-bridge.md`;
const code = (expected) => (error) => error.code === expected;

function fixture(t, { files = ['a.txt'], invokeHelper, otherTask = false, executablePaths = [], newPaths = [] } = {}) {
  const directory = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-evidence-')));
  chmodSync(directory, 0o700);
  const root = join(directory, 'code'); mkdirSync(root);
  const git = (...args) => execFileSync('git', args, { cwd: root, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
  git('init', '-b', 'implementation'); git('config', 'user.email', 'test@example.invalid'); git('config', 'user.name', 'Evidence test');
  git('remote', 'add', 'origin', `https://github.com/${repository}.git`);
  mkdirSync(join(root, 'tasks'));
  writeFileSync(join(root, sourcePath), `---\nid: ${task}\nstatus: in-progress\nowner: operator\nrepositories: [${repository}]\n---\n\n# Bridge fixture\n\n## Implementation claim\n\nChange the exact declared fixture paths.\n`);
  if (otherTask) writeFileSync(join(root, 'tasks/CF-other-bridge.md'), `---\nid: CF-other\nstatus: in-progress\nowner: operator\nrepositories: [${repository}]\n---\n\n# Other task\n\n## Implementation claim\n\nA separate manually reviewed task.\n`);
  for (const file of files.filter((path) => !newPaths.includes(path))) writeFileSync(join(root, file), `initial ${file}\n`);
  for (const file of executablePaths) chmodSync(join(root, file), 0o755);
  git('add', 'tasks', ...files.filter((path) => !newPaths.includes(path))); git('commit', '-m', 'Record reviewed fixture task');
  const revision = git('rev-parse', 'HEAD');
  const cli = (command, ...args) => JSON.parse(execFileSync(process.execPath, [helper, command, ...args], { cwd: root, encoding: 'utf8' }));
  cli('register', '--id', task, '--source-repo', root, '--ref', revision, '--record', sourcePath, '--source-repository', repository);
  const started = cli('start', '--id', task, '--repository', repository, '--files', ...files, '--summary', 'Implement the fixture', '--next', 'Edit and test');
  const spec = { ...structuredClone(taskSpec), task_id: task, repository: { repository_id: repository, revision },
    integration_policy: 'isolated_write', write_scope: files };
  let execution = { execution_id: 'execution-evidence', spec, spec_digest: digest(spec), revision: 1, state: 'PENDING', attempt: null,
    worker_history: [], result: null, accepted_subject_digest: null, cleanup_pending: false, resources_held: false };
  let binding;
  let acceptedManifest; let acceptedPlan;
  // This small authority double isolates helper recovery from namespace admission.
  // A separate store integration test exercises the managed admission and dispatch path.
  const store = {
    get(id) { return id === execution.execution_id ? structuredClone(execution) : null; },
    bindEvidence(input, auth) {
      assert.deepEqual(auth, owner); assert.equal(input.expected_revision, execution.revision);
      if (binding) assert.deepEqual(input.binding, binding);
      binding = structuredClone(input.binding); return binding;
    },
    evidenceBinding() { return structuredClone(binding); },
    getSubject() { return structuredClone(acceptedManifest); },
    getArtifact() { return acceptedPlan; },
  };
  const journalPath = join(directory, 'bridge.sqlite');
  const bridges = [];
  const open = (options = {}) => {
    const bridge = createProjectMemoryBridge({ store, codeRoot: root, sourceRoot: root, journalPath, invokeHelper, ...options });
    bridges.push(bridge); return bridge;
  };
  const prepare = (bridge, extra = {}, auth = owner) => bridge.prepare({ execution_id: execution.execution_id, expected_revision: execution.revision,
    source_repository: repository, source_revision: revision, source_path: sourcePath, claim_heading: 'Implementation claim', files, ...extra }, auth);
  const setPlan = (plan) => {
    acceptedPlan = canonical(plan);
    acceptedManifest = [{ path: 'workspace-plan.json', digest: createHash('sha256').update(acceptedPlan).digest('hex'), repository_revision: revision }];
    const subject = digest(acceptedManifest);
    execution = { ...execution, state: 'COMPLETED', revision: 5,
      worker_history: [{ worker_id: 'fixture-worker', role: 'builder-developer', attempt_id: 'builder-attempt' }],
      result: { kind: 'SUCCEEDED', subject_digest: subject }, accepted_subject_digest: subject };
  };
  const accept = (contents = {}) => {
    const operations = files.map((file) => {
      const repo = createGit(root); const entry = repo.entry(revision, file);
      const before = entry.blob === null ? null : repo.bytes('cat-file', 'blob', entry.blob);
      let content;
      if (Object.hasOwn(contents, file)) content = contents[file];
      else { try { content = readFileSync(join(root, file), 'utf8'); } catch (error) { if (error.code !== 'ENOENT') throw error; content = null; } }
      return { path: file, action: content === null ? 'delete' : 'write',
        before_digest: before === null ? null : createHash('sha256').update(before).digest('hex'), content };
    });
    const plan = { schema_version: 1, execution_id: execution.execution_id, spec_digest: execution.spec_digest,
      attempt_id: 'builder-attempt', repository_revision: revision, write_scope: files, operations };
    setPlan(plan); return plan;
  };
  const completeInput = { execution_id: execution.execution_id, summary: 'Observed fixture changes', validation: 'Fixture check passed against staged content', next: 'Review the committed change' };
  t.after(() => { for (const bridge of bridges) { try { bridge.close(); } catch {} } rmSync(directory, { recursive: true, force: true }); });
  return { directory, root, git, cli, revision, store, started, open, prepare, accept, setPlan, completeInput,
    execution: () => execution, update: (value) => { execution = { ...execution, ...value }; } };
}

test('genuine start binds before edits and final staged blobs survive completion and committed gate', (t) => {
  const f = fixture(t); const bridge = f.open(); const binding = f.prepare(bridge);
  assert.equal(binding.start_event_id, f.started.eventId);
  assert.equal(bridge.guard({ binding, execution: f.execution() }), true);
  assert.equal(statSync(join(f.directory, 'bridge.sqlite')).mode & 0o777, 0o600);
  writeFileSync(join(f.root, 'a.txt'), 'tested final text\n'); f.git('add', 'a.txt'); f.accept();
  const completion = bridge.complete(f.completeInput);
  assert.equal(loadMemory(localMemory(f.root)).graph.byId.get(completion.event_id).validation, f.completeInput.validation);
  f.git('add', '.project-memory'); f.git('commit', '-m', 'Change fixture with genuine evidence');
  const verified = bridge.verify({ execution_id: 'execution-evidence', base: f.revision });
  assert.equal(verified.paths, 1); assert.equal(verified.completion.event_id, completion.event_id);
  assert.deepEqual(bridge.complete(f.completeInput), completion);
  assert.throws(() => bridge.verify({ execution_id: 'execution-evidence', base: 'HEAD' }), code('BOUND_COMPLETION_NOT_FRESH'));
});

test('accepted executable mode and deletions keep helper snapshots unchanged', (t) => {
  const f = fixture(t, { files: ['a.txt', 'removed.txt'], executablePaths: ['a.txt'] }); const bridge = f.open(); f.prepare(bridge);
  writeFileSync(join(f.root, 'a.txt'), 'updated executable text\n'); chmodSync(join(f.root, 'a.txt'), 0o755);
  rmSync(join(f.root, 'removed.txt')); f.git('add', 'a.txt', 'removed.txt'); f.accept();
  const completion = bridge.complete(f.completeInput);
  assert.equal(completion.snapshot['a.txt'].mode, '100755');
  assert.deepEqual(completion.snapshot['removed.txt'], { blob: null, mode: null });
  f.git('add', '.project-memory'); f.git('commit', '-m', 'Record mode and deletion');
  assert.equal(bridge.verify({ execution_id: 'execution-evidence', base: f.revision }).paths, 2);
});

test('preflight rejects after-edit or after-dispatch bindings and inaccurate scope or source identity', (t) => {
  const f = fixture(t); const bridge = f.open();
  assert.throws(() => f.prepare(bridge, { source_repository: 'other/tasks' }), code('REPOSITORY_MISMATCH'));
  assert.throws(() => f.prepare(bridge, { files: ['other.txt'] }), code('EVIDENCE_IDENTITY_MISMATCH'));
  assert.throws(() => f.prepare(bridge, { source_revision: 'a'.repeat(40) }));
  writeFileSync(join(f.root, 'a.txt'), 'already edited\n'); f.git('add', 'a.txt');
  assert.throws(() => f.prepare(bridge), code('EVIDENCE_AFTER_EDITS'));
  f.git('restore', '--staged', '--worktree', 'a.txt');
  f.update({ worker_history: [{ worker_id: 'already-dispatched' }] });
  assert.throws(() => f.prepare(bridge), code('EVIDENCE_AFTER_DISPATCH'));
});

test('fresh guard detects checkpoint, fork, source change and altered start bytes', (t) => {
  const f = fixture(t); const bridge = f.open(); const binding = f.prepare(bridge);
  const memory = loadMemory(localMemory(f.root)); const start = memory.graph.byId.get(binding.start_event_id);
  const fork = { ...start, eventId: randomUUID(), kind: 'checkpoint', parents: [start.eventId], sequence: 1, summary: 'Another writer' };
  writeImmutable(f.root, eventPath(fork), fork);
  assert.throws(() => bridge.guard({ binding, execution: f.execution() }), code('STALE_EVIDENCE_FRONTIER'));
  rmSync(join(f.root, eventPath(fork)));
  const original = readFileSync(join(f.root, eventPath(start)), 'utf8');
  writeFileSync(join(f.root, eventPath(start)), JSON.stringify({ ...start, summary: 'Rewritten start' }));
  assert.throws(() => bridge.guard({ binding, execution: f.execution() }), code('START_BINDING_MISMATCH'));
  writeFileSync(join(f.root, eventPath(start)), original);
  writeFileSync(join(f.root, sourcePath), `${readFileSync(join(f.root, sourcePath), 'utf8')}Changed claim\n`);
  assert.throws(() => bridge.guard({ binding, execution: f.execution() }));
});

test('completion requires accepted runtime result, observed validation and final staged content', (t) => {
  const f = fixture(t); const bridge = f.open(); f.prepare(bridge);
  assert.throws(() => bridge.complete(f.completeInput), code('RUNTIME_RESULT_NOT_ACCEPTED'));
  f.accept({ 'a.txt': 'unstaged change\n' }); assert.throws(() => bridge.complete({ ...f.completeInput, validation: '' }), code('OBSERVED_VALIDATION_REQUIRED'));
  writeFileSync(join(f.root, 'a.txt'), 'unstaged change\n');
  assert.throws(() => bridge.complete(f.completeInput), /stage final tested changes/);
  f.git('add', 'a.txt'); const done = bridge.complete(f.completeInput);
  assert.ok(done.event_id);
  writeFileSync(join(f.root, 'a.txt'), 'later content\n'); f.git('add', 'a.txt');
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_CONTENT_MISMATCH'));
});

test('lost helper response recovers its exact immutable child once across restart', (t) => {
  let invocations = 0;
  const f = fixture(t, { invokeHelper({ run, command }) { if (command === 'complete') { invocations += 1; run(); throw new Error('Lost reply'); } return run(); } });
  let bridge = f.open(); f.prepare(bridge); writeFileSync(join(f.root, 'a.txt'), 'final\n'); f.git('add', 'a.txt'); f.accept();
  const done = bridge.complete(f.completeInput); bridge.close(); bridge = f.open();
  assert.deepEqual(bridge.complete(f.completeInput), done); assert.equal(invocations, 1);
});

test('helper failure without a child remains unknown and never blindly retries', (t) => {
  let calls = 0;
  const f = fixture(t, { invokeHelper() { calls += 1; throw new Error('Unavailable helper'); } });
  const bridge = f.open(); f.prepare(bridge); f.accept();
  assert.throws(() => bridge.complete(f.completeInput), code('COMPLETION_OUTCOME_UNKNOWN'));
  assert.throws(() => bridge.complete(f.completeInput), code('COMPLETION_OUTCOME_UNKNOWN'));
  assert.equal(calls, 1);
});

test('forked helper publication blocks recovery instead of choosing a completion', (t) => {
  const f = fixture(t, { invokeHelper({ run }) {
    run(); const memory = loadMemory(localMemory(f.root));
    const completion = memory.graph.tasks.get(task).heads[0];
    writeImmutable(f.root, eventPath({ ...completion, eventId: '00000000-0000-4000-8000-000000000000' }),
      { ...completion, eventId: '00000000-0000-4000-8000-000000000000' });
  } });
  const bridge = f.open(); f.prepare(bridge); f.accept();
  assert.throws(() => bridge.complete(f.completeInput), code('AMBIGUOUS_COMPLETION'));
});

test('committed gate rejects a different committed blob even when older evidence exists', (t) => {
  const f = fixture(t); const bridge = f.open(); f.prepare(bridge);
  writeFileSync(join(f.root, 'a.txt'), 'first\n'); f.git('add', 'a.txt'); f.accept(); bridge.complete(f.completeInput);
  f.git('add', '.project-memory'); f.git('commit', '-m', 'First completion');
  writeFileSync(join(f.root, 'a.txt'), 'changed after completion\n'); f.git('add', 'a.txt'); f.git('commit', '-m', 'Uncovered change');
  assert.throws(() => bridge.verify({ execution_id: 'execution-evidence', base: f.revision }));
});

test('another task covering the files cannot substitute for the bound completion', (t) => {
  const f = fixture(t, { otherTask: true }); const bridge = f.open(); f.prepare(bridge);
  writeFileSync(join(f.root, 'a.txt'), 'final\n'); f.git('add', 'a.txt'); f.accept(); const done = bridge.complete(f.completeInput);
  f.cli('register', '--id', 'CF-other', '--source-repo', f.root, '--ref', f.revision,
    '--record', 'tasks/CF-other-bridge.md', '--source-repository', repository);
  f.cli('start', '--id', 'CF-other', '--repository', repository, '--files', 'a.txt', '--summary', 'Review other result', '--next', 'Check result');
  f.cli('complete', '--id', 'CF-other', '--summary', 'Other task completion', '--validation', 'Observed result', '--next', 'Review');
  f.git('add', '.project-memory'); f.git('restore', '--staged', '--', done.event_path);
  f.git('commit', '-m', 'Commit other task evidence only');
  assert.throws(() => bridge.verify({ execution_id: 'execution-evidence', base: f.revision }), code('BOUND_COMPLETION_NOT_FRESH'));
});

test('committed gate compares the bound start content as well as its event identity', (t) => {
  const f = fixture(t); const bridge = f.open(); const binding = f.prepare(bridge);
  writeFileSync(join(f.root, 'a.txt'), 'final\n'); f.git('add', 'a.txt'); f.accept(); bridge.complete(f.completeInput);
  const start = loadMemory(localMemory(f.root)).graph.byId.get(binding.start_event_id);
  const original = readFileSync(join(f.root, eventPath(start)), 'utf8');
  writeFileSync(join(f.root, eventPath(start)), JSON.stringify({ ...start, summary: 'Changed only in committed tree' }));
  f.git('add', '.project-memory'); f.git('commit', '-m', 'Commit altered start');
  writeFileSync(join(f.root, eventPath(start)), original);
  assert.throws(() => bridge.verify({ execution_id: 'execution-evidence', base: f.revision }), code('COMMITTED_START_MISMATCH'));
});

test('journal refuses workspace, symlink and public parent paths', (t) => {
  const f = fixture(t);
  assert.throws(() => f.open({ journalPath: join(f.root, 'journal.sqlite') }), code('UNSAFE_JOURNAL_PATH'));
  const target = join(f.directory, 'target'); writeFileSync(target, '', { mode: 0o600 });
  const link = join(f.directory, 'link.sqlite'); symlinkSync(target, link);
  assert.throws(() => f.open({ journalPath: link }), code('UNSAFE_JOURNAL_PATH'));
  const publicDir = join(f.directory, 'public'); mkdirSync(publicDir, { mode: 0o755 });
  assert.throws(() => f.open({ journalPath: join(publicDir, 'journal.sqlite') }), code('UNSAFE_JOURNAL_PATH'));
});

test('denied evidence preparation writes no durable operation before owner authentication', (t) => {
  const f = fixture(t); const bridge = f.open();
  assert.throws(() => f.prepare(bridge, {}, { actor_id: 'other-owner', role: 'owner' }), code('UNAUTHORIZED'));
  const journal = new DatabaseSync(join(f.directory, 'bridge.sqlite'), { readOnly: true });
  try { assert.equal(journal.prepare('SELECT count(*) AS count FROM operations').get().count, 0); }
  finally { journal.close(); }
  assert.equal(f.prepare(bridge).task_id, task);
});

test('accepted plan binds exact staged bytes including scoped files with no operation', (t) => {
  let calls = 0;
  const f = fixture(t, { files: ['a.txt', 'b.txt'], invokeHelper({ run }) { calls += 1; return run(); } });
  const bridge = f.open(); f.prepare(bridge);
  const plan = f.accept({ 'a.txt': 'reviewed text\n' });
  plan.operations = plan.operations.filter((operation) => operation.path === 'a.txt'); f.setPlan(plan);
  writeFileSync(join(f.root, 'a.txt'), 'unreviewed text\n'); f.git('add', 'a.txt');
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_CONTENT_MISMATCH')); assert.equal(calls, 0);
  writeFileSync(join(f.root, 'a.txt'), 'reviewed text\n'); writeFileSync(join(f.root, 'b.txt'), 'unreviewed second file\n'); f.git('add', 'a.txt', 'b.txt');
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_CONTENT_MISMATCH')); assert.equal(calls, 0);
  f.git('restore', '--staged', '--worktree', 'b.txt'); assert.ok(bridge.complete(f.completeInput).event_id); assert.equal(calls, 1);
});

test('accepted plan rejects altered modes, missing artifacts, stale base and foreign attempts', (t) => {
  const f = fixture(t); const bridge = f.open(); f.prepare(bridge); const plan = f.accept();
  chmodSync(join(f.root, 'a.txt'), 0o755); f.git('add', 'a.txt');
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_CONTENT_MISMATCH'));
  chmodSync(join(f.root, 'a.txt'), 0o644); f.git('add', 'a.txt');
  f.setPlan({ ...plan, attempt_id: 'unrelated-builder' });
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_PLAN_IDENTITY_MISMATCH'));
  f.setPlan({ ...plan, operations: [{ ...plan.operations[0], before_digest: 'a'.repeat(64) }] });
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_PLAN_BASE_MISMATCH'));
  f.setPlan(plan); f.store.getSubject = () => null;
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_PLAN_REQUIRED'));
});

test('accepted creation requires staged regular mode and the exact new file bytes', (t) => {
  const f = fixture(t, { files: ['new.txt'], newPaths: ['new.txt'] }); const bridge = f.open(); f.prepare(bridge);
  writeFileSync(join(f.root, 'new.txt'), 'accepted new file\n', { mode: 0o755 }); f.git('add', 'new.txt'); f.accept();
  assert.throws(() => bridge.complete(f.completeInput), code('ACCEPTED_CONTENT_MISMATCH'));
  chmodSync(join(f.root, 'new.txt'), 0o644); f.git('add', 'new.txt');
  assert.equal(bridge.complete(f.completeInput).snapshot['new.txt'].mode, '100644');
});

test('accepted terminal output waits for confirmed host cleanup and released resources', (t) => {
  const f = fixture(t); const bridge = f.open(); f.prepare(bridge); f.accept();
  f.update({ cleanup_pending: true, resources_held: true });
  assert.throws(() => bridge.complete(f.completeInput), code('RUNTIME_CLEANUP_REQUIRED'));
  f.update({ cleanup_pending: false });
  assert.throws(() => bridge.complete(f.completeInput), code('RUNTIME_CLEANUP_REQUIRED'));
  f.update({ resources_held: false }); assert.ok(bridge.complete(f.completeInput).event_id);
});

test('managed store binds genuine evidence, rechecks assignment and launch, then completes the committed gate', (t) => {
  const f = fixture(t); let bridge;
  const store = openStore(join(f.directory, 'authority.sqlite'), {
    worker_ids: ['builder', 'reviewer'],
    evidence_guard: (binding, execution) => bridge.guard({ binding, execution }),
    // Host enforcement has its own tests. This proof double isolates the evidence boundary.
    isolation_guard: (execution, attempt) => ({ verified: true, execution_id: execution.execution_id,
      spec_digest: execution.spec_digest, attempt_id: attempt.attempt_id, workspace: attempt.workspace,
      repository_revision: execution.spec.repository.revision, write_scope: execution.spec.write_scope }),
  });
  t.after(() => store.close());
  const governanceRoot = join(f.directory, 'governance'); mkdirSync(governanceRoot);
  const governanceGit = (...args) => execFileSync('git', args, { cwd: governanceRoot, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
  governanceGit('init', '-b', 'decisions'); governanceGit('config', 'user.name', 'Fixture'); governanceGit('config', 'user.email', 'fixture@example.invalid');
  governanceGit('remote', 'add', 'origin', `https://github.com/${repository}.git`);
  for (const [index, mode] of ['SHADOW', 'MANAGED'].entries()) {
    const decision = { schema_version: 1, decision_id: `evidence-decision-${index}`, namespace_id: 'evidence', task_prefix: 'CF-',
      repository, source_repository: repository, accountable_owner: 'operator', authority_instance_id: store.instanceId(),
      expected_generation: index, mode, require_evidence: true, allow_writes: true, prior_authority_drained: true,
      claim_disposition: 'new_namespace', previous_authority: 'none: invented fixture namespace', reason: 'Exercise local evidence binding' };
    const path = `decision-${index}.json`; writeFileSync(join(governanceRoot, path), JSON.stringify(decision));
    governanceGit('add', path); governanceGit('commit', '-m', 'Record fixture authority decision');
    store.recordNamespaceDecision({ decision, source_root: governanceRoot, source: { repository,
      revision: governanceGit('rev-parse', 'HEAD'), path, blob: governanceGit('rev-parse', `HEAD:${path}`) } }, owner);
  }
  const spec = { ...f.execution().spec, acceptance_checks: [], workflow: { workflow_id: 'isolated-write-v1', version: 1,
    stages: [{ id: 'build', role: 'builder-developer', required_checks: [], requires_approval: false },
      { id: 'review', role: 'reviewer-verifier', required_checks: [], requires_approval: false }] } };
  const source = [...loadMemory(localMemory(f.root)).sources.values()][0];
  store.registerAdmission({ namespace_id: 'evidence', generation: 2, spec, source, source_root: f.root, reviewed_by: 'operator' }, owner);
  let sequence = 0; const execution_id = 'execution-evidence';
  const command = (kind, payload, auth = owner) => store.execute({ schema_version: 1, command_id: `evidence-workflow-${++sequence}`,
    execution_id, expected_revision: store.get(execution_id)?.revision ?? 0, at: '2026-10-30T00:00:00.000Z', kind, payload }, auth);
  command('ADMIT', { spec }); bridge = f.open({ store }); f.prepare(bridge);
  writeFileSync(join(f.root, 'a.txt'), 'premature edit\n');
  const assignment = { attempt_id: 'builder-attempt', worker_id: 'builder', input_revisions: [], workspace: 'isolated-fixture',
    capabilities: ['read', 'write'], lease_deadline: '2026-10-30T00:01:00.000Z' };
  assert.throws(() => command('ASSIGN', assignment), /stage final tested changes/);
  f.git('restore', 'a.txt'); command('ASSIGN', assignment);
  const launch = store.pendingEffects().find((item) => item.effect.kind === 'LAUNCH_ATTEMPT');
  writeFileSync(join(f.root, 'a.txt'), 'changed before launch\n');
  assert.throws(() => store.claimEffect({ effectId: launch.effect.effect_id, dispatcherId: 'dispatcher', now: '2026-10-30T00:00:00.000Z', leaseMs: 1000 }), /stage final tested changes/);
  f.git('restore', 'a.txt');
  assert.equal(store.claimEffect({ effectId: launch.effect.effect_id, dispatcherId: 'dispatcher', now: '2026-10-30T00:00:00.000Z', leaseMs: 1000 }).status, 'CLAIMED');
  const plan = { schema_version: 1, execution_id, spec_digest: digest(spec), attempt_id: 'builder-attempt',
    repository_revision: f.revision, write_scope: ['a.txt'], operations: [{ path: 'a.txt', action: 'write',
      before_digest: createHash('sha256').update('initial a.txt\n').digest('hex'), content: 'observed final implementation\n' }] };
  const content = canonical(plan); const artifactDigest = createHash('sha256').update(content).digest('hex');
  store.putArtifact({ digest: artifactDigest, content });
  const artifacts = [{ path: 'workspace-plan.json', digest: artifactDigest, repository_revision: f.revision }]; const subject = digest(artifacts);
  const reviewContent = 'Independent reviewer accepted the candidate plan';
  const reviewDigest = createHash('sha256').update(reviewContent).digest('hex'); store.putArtifact({ digest: reviewDigest, content: reviewContent });
  const reviewArtifacts = [{ path: 'review.md', digest: reviewDigest, repository_revision: f.revision }];
  const reviewSubject = digest(reviewArtifacts);
  for (const worker_id of ['builder', 'reviewer']) {
    if (worker_id === 'reviewer') command('ASSIGN', { ...assignment, attempt_id: 'reviewer-attempt', worker_id,
      capabilities: ['read'], input_revisions: [subject] });
    const attempt = store.get(execution_id).attempt; const auth = { actor_id: worker_id, role: 'worker' };
    const attemptRef = { attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token };
    command('ACKNOWLEDGE', attemptRef, auth);
    const resultArtifacts = worker_id === 'builder' ? artifacts : reviewArtifacts;
    const resultSubject = worker_id === 'builder' ? subject : reviewSubject;
    command('SUBMIT', { result: { schema_version: 1, ...attemptRef, spec_digest: digest(spec), kind: 'SUCCEEDED',
      artifacts: resultArtifacts, subject_digest: resultSubject } }, auth);
    command('ACCEPT_STAGE', { subject_digest: resultSubject }); command('CLEANUP_CONFIRMED', attemptRef);
  }
  assert.equal(store.get(execution_id).state, 'COMPLETED');
  assert.equal(store.get(execution_id).accepted_subject_digest, subject);
  assert.equal(store.get(execution_id).result.subject_digest, reviewSubject);
  writeFileSync(join(f.root, 'a.txt'), 'observed final implementation\n'); f.git('add', 'a.txt');
  bridge.complete(f.completeInput); f.git('add', '.project-memory'); f.git('commit', '-m', 'Commit runtime-bound evidence');
  assert.equal(bridge.verify({ execution_id, base: f.revision }).paths, 1);
  assert.equal(store.replay().executions[0].state, 'COMPLETED');
});
