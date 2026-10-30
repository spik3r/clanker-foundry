import test from 'node:test';
import assert from 'node:assert/strict';
import { chmodSync, mkdtempSync, readFileSync, realpathSync, rmSync, symlinkSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawn } from 'node:child_process';
import { DatabaseSync } from 'node:sqlite';
import { createHash } from 'node:crypto';
import { openStore } from '../src/store.mjs';
import { digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

const owner = { actor_id: 'operator', role: 'owner' };
const worker = { actor_id: 'worker-1', role: 'worker' };
const at = '2026-10-30T00:00:00.000Z';
const policy = { worker_ids: ['worker-1'] };
let commandSequence = 0;
function command(kind, execution_id, expected_revision, payload, extra = {}) {
  return { schema_version: 1, command_id: `store-command-${++commandSequence}`, execution_id, expected_revision, at, kind, payload, ...extra };
}
function spec(id = 'CF-store') {
  const value = { ...structuredClone(taskSpec), task_id: id };
  value.workflow.stages[0].role = 'explorer-researcher';
  return value;
}
function assign(id, revision = 1, attempt = `attempt-${id}`) {
  return command('ASSIGN', id, revision, { attempt_id: attempt, worker_id: 'worker-1', input_revisions: [], workspace: 'isolated-workspace', capabilities: ['read'], lease_deadline: '2026-10-30T00:01:00.000Z' });
}
function fixture(t, options = policy) {
  const directory = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-store-')));
  chmodSync(directory, 0o700);
  const path = join(directory, 'authority.sqlite');
  const stores = [];
  const open = () => { const store = openStore(path, options); stores.push(store); return store; };
  t.after(() => { for (const store of stores) { try { store.close(); } catch {} } rmSync(directory, { recursive: true, force: true }); });
  return { path, directory, open, store: open() };
}
function code(expected) { return (error) => error.code === expected; }

// Separate processes compete for the same writer lock, not just the JS event loop.
function childExecute(path, input, method = 'execute') {
  return new Promise((resolve, reject) => {
    const script = `import {openStore} from ${JSON.stringify(new URL('../src/store.mjs', import.meta.url).href)}; const s=openStore(process.argv[1],JSON.parse(process.argv[3]));try {const r=s[process.argv[4]](JSON.parse(process.argv[2]),{actor_id:'operator',role:'owner'});console.log(JSON.stringify({ok:true,revision:r.execution?.revision,status:r.status}));}catch(e){console.log(JSON.stringify({ok:false,code:e.code}));}finally{s.close();}`;
    const child = spawn(process.execPath, ['--input-type=module', '-e', script, path, JSON.stringify(input), JSON.stringify(policy), method], { stdio: ['ignore', 'pipe', 'pipe'] });
    let output = ''; let errors = '';
    child.stdout.on('data', (data) => { output += data; }); child.stderr.on('data', (data) => { errors += data; });
    child.on('error', reject); child.on('close', (status) => { if (status !== 0) reject(new Error(errors)); else resolve(JSON.parse(output)); });
  });
}

test('receipts replay exactly after restart and remain bound to principal and payload', (t) => {
  const f = fixture(t); const admit = command('ADMIT', 'execution-1', 0, { spec: spec() });
  const receipt = f.store.execute(admit, owner);
  f.store.execute(assign('execution-1'), owner);
  f.store.close(); const reopened = f.open();
  assert.deepEqual(reopened.execute(admit, owner), receipt);
  assert.equal(reopened.events().length, 2);
  assert.throws(() => reopened.execute(admit, worker), code('UNAUTHORIZED'));
  assert.throws(() => reopened.execute(admit, { actor_id: 'another-owner', role: 'owner' }), code('UNAUTHORIZED'));
  const changed = structuredClone(admit); changed.payload.spec.outcome = 'Changed';
  assert.throws(() => reopened.execute(changed, owner), code('COMMAND_ID_REUSED'));
  assert.equal(reopened.get('execution-1').revision, 2);
});

test('concurrent process assignments have one committed winner', async (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'race', 0, { spec: spec() }), owner);
  const results = await Promise.all([childExecute(f.path, assign('race', 1, 'attempt-a')), childExecute(f.path, assign('race', 1, 'attempt-b'))]);
  assert.equal(results.filter((result) => result.ok).length, 1);
  assert.equal(results.find((result) => !result.ok).code, 'REVISION_CONFLICT');
  assert.equal(f.store.get('race').revision, 2);
  assert.equal(f.store.pendingEffects().length, 1);
});

test('different spec versions cannot create concurrent nonterminal task executions', async (t) => {
  const f = fixture(t); const first = spec(); const second = { ...first, spec_revision: 2 };
  const results = await Promise.all([
    childExecute(f.path, command('ADMIT', 'admission-a', 0, { spec: first })),
    childExecute(f.path, command('ADMIT', 'admission-b', 0, { spec: second })),
  ]);
  assert.equal(results.filter((result) => result.ok).length, 1);
  assert.equal(results.find((result) => !result.ok).code, 'TASK_ALREADY_ACTIVE');
  assert.equal(f.store.list().length, 1);
});

test('reservations conflict across task IDs and survive cancellation until confirmed cleanup', (t) => {
  const f = fixture(t); const resources = [{ resource_id: 'api:shared', mode: 'exclusive' }];
  f.store.execute(command('ADMIT', 'a', 0, { spec: { ...spec('CF-a'), shared_resources: resources } }), owner);
  f.store.execute(command('ADMIT', 'b', 0, { spec: { ...spec('CF-b'), shared_resources: resources } }), owner);
  f.store.execute(assign('a'), owner);
  const assignB = assign('b');
  assert.throws(() => f.store.execute(assignB, owner), code('RESOURCE_CONFLICT'));
  f.store.execute(command('CANCEL', 'a', 2, { code: 'OWNER_CANCELLED' }), owner);
  assert.throws(() => f.store.execute(assignB, owner), code('RESOURCE_CONFLICT'));
  assert.equal(f.store.get('a').resources_held, true);
  f.store.execute(command('CLEANUP_CONFIRMED', 'a', 3, { attempt_id: 'attempt-a', fencing_token: 1 }), owner);
  assert.equal(f.store.execute(assignB, owner).execution.state, 'ASSIGNED');
});

test('shared reservations coexist but an exclusive reservation waits', (t) => {
  const f = fixture(t);
  for (const [id, mode] of [['a', 'shared'], ['b', 'shared'], ['c', 'exclusive']]) {
    f.store.execute(command('ADMIT', id, 0, { spec: { ...spec(`CF-${id}`), shared_resources: [{ resource_id: 'shared-interface', mode }] } }), owner);
  }
  f.store.execute(assign('a'), owner); f.store.execute(assign('b'), owner);
  assert.throws(() => f.store.execute(assign('c'), owner), code('RESOURCE_CONFLICT'));
  assert.deepEqual(f.store.list().map((execution) => execution.execution_id), ['a', 'b', 'c']);
});

test('dependencies must exist and completion is checked again at assignment', (t) => {
  const f = fixture(t); const dependency = spec('CF-dependency');
  const dependent = { ...spec('CF-dependent'), dependencies: [{ task_id: dependency.task_id, spec_digest: digest(dependency) }] };
  const admit = command('ADMIT', 'dependent', 0, { spec: dependent });
  assert.throws(() => f.store.execute(admit, owner), code('DEPENDENCY_NOT_FOUND'));
  f.store.execute(command('ADMIT', 'dependency', 0, { spec: dependency }), owner);
  f.store.execute(admit, owner);
  assert.throws(() => f.store.execute(assign('dependent'), owner), code('DEPENDENCIES_NOT_SATISFIED'));
  f.store.execute(command('CANCEL', 'dependency', 1, { code: 'OWNER_CANCELLED' }), owner);
  assert.throws(() => f.store.execute(assign('dependent'), owner), code('DEPENDENCIES_NOT_SATISFIED'));
});

test('revision reuse cannot mutate an immutable spec', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'original', 0, { spec: spec() }), owner);
  f.store.execute(command('CANCEL', 'original', 1, { code: 'OWNER_CANCELLED' }), owner);
  const changed = { ...spec(), outcome: 'Different content under the old revision.' };
  assert.throws(() => f.store.execute(command('ADMIT', 'replacement', 0, { spec: changed }), owner), code('SPEC_IMMUTABLE'));
  assert.equal(f.store.get('replacement'), null);
});

test('caller readiness and registry fields cannot authorize an unregistered worker', (t) => {
  const f = fixture(t, {}); f.store.execute(command('ADMIT', 'unregistered', 0, { spec: spec() }), owner);
  assert.throws(() => f.store.execute(assign('unregistered'), { ...owner, worker_eligible: true, worker_ids: ['worker-1'] }), code('WORKER_INELIGIBLE'));
  assert.equal(f.store.get('unregistered').revision, 1);
  assert.equal(f.store.pendingEffects().length, 0);
});

test('failed receipt insert rolls back state, event, reservation and outbox together', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'rollback', 0, { spec: { ...spec(), shared_resources: [{ resource_id: 'rollback-lock', mode: 'exclusive' }] } }), owner);
  const raw = new DatabaseSync(f.path);
  raw.exec("CREATE TRIGGER fail_receipt BEFORE INSERT ON receipts BEGIN SELECT RAISE(ABORT,'injected receipt failure'); END");
  const assignment = assign('rollback');
  assert.throws(() => f.store.execute(assignment, owner), /injected receipt failure/);
  assert.equal(f.store.get('rollback').revision, 1);
  assert.equal(f.store.events().length, 1);
  assert.equal(f.store.pendingEffects().length, 0);
  assert.equal(raw.prepare('SELECT count(*) AS n FROM reservations').get().n, 0);
  raw.exec('DROP TRIGGER fail_receipt'); raw.close();
  assert.equal(f.store.execute(assignment, owner).execution.revision, 2);
});

test('commit before dispatch survives process exit and store reopen', async (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'crash', 0, { spec: spec() }), owner);
  await childExecute(f.path, assign('crash'));
  f.store.close(); const reopened = f.open();
  const [pending] = reopened.pendingEffects();
  assert.equal(pending.status, 'PENDING');
  assert.equal(pending.effect.kind, 'LAUNCH_ATTEMPT');
  assert.equal(reopened.get('crash').state, 'ASSIGNED');
});

test('outbox claims fence dispatchers and identical acknowledgements are idempotent', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'dispatch', 0, { spec: spec() }), owner);
  f.store.execute(assign('dispatch'), owner);
  const effectId = f.store.pendingEffects()[0].effect.effect_id;
  const claim = f.store.claimEffect({ effectId, dispatcherId: 'dispatcher-a', now: at, leaseMs: 1000 });
  assert.equal(claim.status, 'CLAIMED');
  assert.throws(() => f.store.claimEffect({ effectId, dispatcherId: 'dispatcher-a', now: at, leaseMs: 1000 }), code('EFFECT_ALREADY_CLAIMED'));
  assert.throws(() => f.store.claimEffect({ effectId, dispatcherId: 'dispatcher-b', now: at, leaseMs: 1000 }), code('EFFECT_ALREADY_CLAIMED'));
  const acknowledgement = { effectId, dispatcherId: 'dispatcher-a', claimToken: claim.claim_token, outcome: 'DELIVERED', receipt: { launched: 'attempt-dispatch' } };
  assert.throws(() => f.store.acknowledgeEffect({ ...acknowledgement, claimToken: 'wrong-token' }), code('STALE_EFFECT_CLAIM'));
  const delivered = f.store.acknowledgeEffect(acknowledgement);
  assert.deepEqual(f.store.acknowledgeEffect(acknowledgement), delivered);
  assert.throws(() => f.store.acknowledgeEffect({ ...acknowledgement, receipt: { launched: 'another-attempt' } }), code('EFFECT_OUTCOME_CONFLICT'));
  assert.equal(f.store.pendingEffects().length, 0);
});

test('unknown dispatch outcomes require owner reconciliation before any new claim', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'unknown', 0, { spec: spec() }), owner);
  f.store.execute(assign('unknown'), owner);
  const effectId = f.store.pendingEffects()[0].effect.effect_id;
  const first = f.store.claimEffect({ effectId, dispatcherId: 'dead-dispatcher', now: at, leaseMs: 1000 });
  f.store.close(); const reopened = f.open();
  const expired = reopened.claimEffect({ effectId, dispatcherId: 'new-dispatcher', now: '2026-10-30T00:00:01.000Z', leaseMs: 1000 });
  assert.equal(expired.status, 'UNKNOWN');
  assert.equal(expired.claim_token, first.claim_token);
  assert.equal(reopened.claimEffect({ effectId, dispatcherId: 'new-dispatcher', now: '2026-10-30T00:00:02.000Z', leaseMs: 1000 }).status, 'UNKNOWN');
  assert.throws(() => reopened.reconcileEffect({ effectId, outcome: 'PENDING', receipt: { confirmed_absent: true } }, worker), code('UNAUTHORIZED'));
  reopened.reconcileEffect({ effectId, outcome: 'PENDING', receipt: { confirmed_absent: true } }, owner);
  const second = reopened.claimEffect({ effectId, dispatcherId: 'new-dispatcher', now: '2026-10-30T00:00:02.000Z', leaseMs: 1000 });
  assert.notEqual(second.claim_token, first.claim_token);
  assert.throws(() => reopened.acknowledgeEffect({ effectId, dispatcherId: 'dead-dispatcher', claimToken: first.claim_token, outcome: 'DELIVERED', receipt: {} }), code('STALE_EFFECT_CLAIM'));
});

test('late results after cancellation cannot alter a newer revision', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'stale', 0, { spec: spec() }), owner);
  f.store.execute(assign('stale'), owner);
  f.store.execute(command('CANCEL', 'stale', 2, { code: 'OWNER_CANCELLED' }), owner);
  assert.throws(() => f.store.execute(command('ACKNOWLEDGE', 'stale', 2, { attempt_id: 'attempt-stale', fencing_token: 1 }), worker), code('REVISION_CONFLICT'));
  assert.equal(f.store.get('stale').state, 'CANCELLED');
  assert.equal(f.store.events('stale').length, 3);
});

test('database reopen rejects symlinks, loose permissions and unsupported versions', (t) => {
  const f = fixture(t); f.store.close();
  const link = join(f.directory, 'link.sqlite'); symlinkSync(f.path, link);
  assert.throws(() => openStore(link), code('UNSAFE_DATABASE_PATH'));
  chmodSync(f.path, 0o644);
  assert.throws(() => openStore(f.path), code('UNSAFE_DATABASE_PATH')); chmodSync(f.path, 0o600);
  const raw = new DatabaseSync(f.path); raw.exec('PRAGMA user_version=99'); raw.close();
  assert.throws(() => openStore(f.path), code('UNSUPPORTED_DATABASE'));
  assert.ok(readFileSync(f.path).length > 0);
});

test('cancelled launch intents are fenced before dispatch and remain skipped on reopen', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'fenced', 0, { spec: spec() }), owner);
  f.store.execute(assign('fenced'), owner);
  const effectId = f.store.pendingEffects()[0].effect.effect_id;
  f.store.execute(command('CANCEL', 'fenced', 2, { code: 'OWNER_CANCELLED' }), owner);
  const skipped = f.store.claimEffect({ effectId, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 });
  assert.equal(skipped.status, 'DELIVERED'); assert.equal(skipped.receipt.skipped, 'STALE_ATTEMPT');
  f.store.close(); const reopened = f.open();
  assert.equal(reopened.pendingEffects().some((item) => item.effect.effect_id === effectId), false);
  assert.equal(reopened.pendingEffects()[0].effect.kind, 'REVOKE_CAPABILITIES');
});

test('cross-task process reservation races choose one winner', async (t) => {
  const f = fixture(t);
  for (const id of ['one', 'two']) f.store.execute(command('ADMIT', id, 0, { spec: { ...spec(`CF-${id}`), shared_resources: [{ resource_id: 'race-resource', mode: 'exclusive' }] } }), owner);
  const results = await Promise.all([childExecute(f.path, assign('one')), childExecute(f.path, assign('two'))]);
  assert.equal(results.filter((result) => result.ok).length, 1);
  assert.equal(results.find((result) => !result.ok).code, 'RESOURCE_CONFLICT');
  assert.equal(f.store.pendingEffects().length, 1);
});

test('completed pinned dependency unlocks assignment and spec-version graph cycles reject', (t) => {
  const f = fixture(t);
  const dependency = spec('CF-completing'); dependency.acceptance_checks = [];
  dependency.workflow.stages[0].required_checks = []; dependency.workflow.stages[0].requires_approval = false;
  f.store.execute(command('ADMIT', 'completing', 0, { spec: dependency }), owner);
  f.store.execute(assign('completing'), owner);
  f.store.execute(command('ACKNOWLEDGE', 'completing', 2, { attempt_id: 'attempt-completing', fencing_token: 1 }), worker);
  const content = 'Reviewed dependency result';
  const artifactDigest = createHash('sha256').update(content).digest('hex');
  f.store.putArtifact({ digest: artifactDigest, content });
  const artifacts = [{ path: 'result.md', digest: artifactDigest, repository_revision: dependency.repository.revision }];
  f.store.execute(command('SUBMIT', 'completing', 3, { result: { schema_version: 1, attempt_id: 'attempt-completing', fencing_token: 1, spec_digest: digest(dependency), kind: 'SUCCEEDED', artifacts, subject_digest: digest(artifacts) } }), worker);
  f.store.execute(command('ACCEPT_STAGE', 'completing', 4, { subject_digest: digest(artifacts) }), owner);
  const dependent = { ...spec('CF-next'), dependencies: [{ task_id: dependency.task_id, spec_digest: digest(dependency) }] };
  f.store.execute(command('ADMIT', 'next', 0, { spec: dependent }), owner);
  const selected = f.store.dependencyOutputs('next');
  assert.equal(selected.length, 1); assert.equal(selected[0].execution_id, 'completing');
  assert.equal(selected[0].subject_digest, digest(artifacts));
  assert.equal(f.store.execute(assign('next'), owner).execution.state, 'ASSIGNED');
  assert.equal(digest(f.store.history('next')[1].context.dependency_outputs), digest(selected));
  const cyclic = { ...dependency, spec_revision: 2, dependencies: [{ task_id: dependent.task_id, spec_digest: digest(dependent) }] };
  assert.throws(() => f.store.execute(command('ADMIT', 'cycle', 0, { spec: cyclic }), owner), code('DEPENDENCY_CYCLE'));
});

test('database path rejects symlink sidecars before SQLite can follow them', (t) => {
  const f = fixture(t); f.store.close();
  const target = join(f.directory, 'authority.sqlite-journal');
  symlinkSync(f.path, target);
  assert.throws(() => openStore(f.path), code('UNSAFE_DATABASE_PATH'));
});

test('replay after restart reconstructs accepted inputs and intents without redispatch', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'replay-a', 0, { spec: spec('CF-replay-a') }), owner);
  f.store.execute(command('ADMIT', 'replay-b', 0, { spec: spec('CF-replay-b') }), owner);
  f.store.execute(assign('replay-a'), owner);
  f.store.execute(assign('replay-b'), owner);
  const effectId = f.store.pendingEffects()[0].effect.effect_id;
  const claim = f.store.claimEffect({ effectId, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 });
  f.store.acknowledgeEffect({ effectId, dispatcherId: 'dispatcher', claimToken: claim.claim_token, outcome: 'DELIVERED', receipt: { launched: true } });
  f.store.execute(command('CANCEL', 'replay-a', 2, { code: 'OWNER_CANCELLED' }), owner);
  const original = f.store.replay(); const pending = f.store.pendingEffects(); const history = f.store.history();
  assert.equal(history.length, 5); assert.equal(f.store.history('replay-a').length, 3);
  assert.equal(history[2].context.worker_eligible, true);
  assert.equal(history[2].command.kind, 'ASSIGN');
  f.store.close(); const reopened = f.open();
  assert.deepEqual(reopened.replay(), original);
  assert.deepEqual(reopened.replay().executions, reopened.list());
  assert.deepEqual(reopened.pendingEffects(), pending);
});

test('replay detects a corrupted accepted outcome rather than trusting snapshots', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'tampered', 0, { spec: spec() }), owner);
  const raw = new DatabaseSync(f.path); const corrupted = structuredClone(f.store.get('tampered'));
  corrupted.retry_count = 1;
  raw.prepare('UPDATE executions SET body=? WHERE execution_id=?').run(JSON.stringify(corrupted), 'tampered'); raw.close();
  assert.throws(() => f.store.replay(), code('REPLAY_MISMATCH'));
});

test('artifact bytes persist immutably with raw UTF-8 digests across restart', (t) => {
  const f = fixture(t); const content = 'Verified output\nUnicode: café.';
  const artifactDigest = createHash('sha256').update(content, 'utf8').digest('hex');
  assert.equal(f.store.getArtifact(artifactDigest), null);
  assert.equal(f.store.putArtifact({ digest: artifactDigest, content }), artifactDigest);
  assert.equal(f.store.putArtifact({ digest: artifactDigest, content }), artifactDigest);
  assert.throws(() => f.store.putArtifact({ digest: artifactDigest, content: 'Changed bytes' }), code('ARTIFACT_DIGEST_MISMATCH'));
  assert.throws(() => f.store.putArtifact({ digest: artifactDigest, content: 'x'.repeat(1_048_577) }), code('INVALID_ARTIFACT'));
  f.store.close(); const reopened = f.open();
  assert.equal(reopened.getArtifact(artifactDigest), content);
});

test('successful submission requires stored bytes and replay checks their hashes', (t) => {
  const f = fixture(t); const admittedSpec = spec();
  f.store.execute(command('ADMIT', 'artifact', 0, { spec: admittedSpec }), owner);
  f.store.execute(assign('artifact'), owner);
  f.store.execute(command('ACKNOWLEDGE', 'artifact', 2, { attempt_id: 'attempt-artifact', fencing_token: 1 }), worker);
  const content = 'Actual output'; const artifactDigest = createHash('sha256').update(content).digest('hex');
  const artifacts = [{ path: 'output.md', digest: artifactDigest, repository_revision: admittedSpec.repository.revision }];
  const submit = command('SUBMIT', 'artifact', 3, { result: { schema_version: 1, attempt_id: 'attempt-artifact', fencing_token: 1, spec_digest: digest(admittedSpec), kind: 'SUCCEEDED', artifacts, subject_digest: digest(artifacts) } });
  assert.throws(() => f.store.execute(submit, worker), code('ARTIFACT_NOT_FOUND'));
  assert.equal(f.store.get('artifact').revision, 3);
  f.store.putArtifact({ digest: artifactDigest, content });
  f.store.execute(submit, worker);
  assert.equal(f.store.replay().executions[0].state, 'VALIDATING');
  const raw = new DatabaseSync(f.path); raw.prepare('UPDATE artifacts SET content=? WHERE digest=?').run('Mutated output', artifactDigest); raw.close();
  assert.throws(() => f.store.getArtifact(artifactDigest), code('ARTIFACT_CORRUPTED'));
  assert.throws(() => f.store.replay(), code('ARTIFACT_CORRUPTED'));
});

test('task resubmission waits for explicit cleanup even without declared shared resources', (t) => {
  const f = fixture(t); const first = spec();
  f.store.execute(command('ADMIT', 'old', 0, { spec: first }), owner);
  f.store.execute(assign('old'), owner);
  f.store.execute(command('CANCEL', 'old', 2, { code: 'OWNER_CANCELLED' }), owner);
  const replacement = command('ADMIT', 'new', 0, { spec: { ...first, spec_revision: 2 } });
  assert.throws(() => f.store.execute(replacement, owner), code('CLEANUP_REQUIRED'));
  f.store.execute(command('CLEANUP_CONFIRMED', 'old', 3, { attempt_id: 'attempt-old', fencing_token: 1 }), owner);
  assert.equal(f.store.execute(replacement, owner).execution.state, 'PENDING');
});

test('concurrent first open initializes one database without losing either admission', async (t) => {
  const f = fixture(t); const freshPath = join(f.directory, 'fresh.sqlite');
  const results = await Promise.all([
    childExecute(freshPath, command('ADMIT', 'fresh-one', 0, { spec: spec('CF-fresh-one') })),
    childExecute(freshPath, command('ADMIT', 'fresh-two', 0, { spec: spec('CF-fresh-two') })),
  ]);
  assert.ok(results.every((result) => result.ok));
  const reopened = openStore(freshPath, policy);
  try { assert.equal(reopened.list().length, 2); assert.equal(reopened.replay().executions.length, 2); }
  finally { reopened.close(); }
});

test('subject manifests remain readable after restart and reject manifest corruption', (t) => {
  const f = fixture(t); const admitted = spec();
  f.store.execute(command('ADMIT', 'subject', 0, { spec: admitted }), owner);
  f.store.execute(assign('subject'), owner);
  f.store.execute(command('ACKNOWLEDGE', 'subject', 2, { attempt_id: 'attempt-subject', fencing_token: 1 }), worker);
  const content = 'Input for the next stage'; const artifactDigest = createHash('sha256').update(content).digest('hex');
  f.store.putArtifact({ digest: artifactDigest, content });
  const artifacts = [{ path: 'input.md', digest: artifactDigest, repository_revision: admitted.repository.revision }];
  const subjectDigest = digest(artifacts);
  f.store.execute(command('SUBMIT', 'subject', 3, { result: { schema_version: 1, attempt_id: 'attempt-subject', fencing_token: 1, spec_digest: digest(admitted), kind: 'SUCCEEDED', artifacts, subject_digest: subjectDigest } }), worker);
  assert.equal(digest(f.store.getSubject(subjectDigest)), subjectDigest);
  f.store.close(); const reopened = f.open();
  const manifest = reopened.getSubject(subjectDigest);
  assert.equal(manifest[0].path, 'input.md'); assert.equal(reopened.getArtifact(manifest[0].digest), content);
  const raw = new DatabaseSync(f.path); raw.prepare('UPDATE subjects SET manifest=? WHERE digest=?').run('[]', subjectDigest); raw.close();
  assert.throws(() => reopened.getSubject(subjectDigest), code('SUBJECT_CORRUPTED'));
  assert.throws(() => reopened.replay(), code('SUBJECT_CORRUPTED'));
});

test('two processes using the same dispatcher cannot obtain the same live claim', async (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'claim-race', 0, { spec: spec() }), owner);
  f.store.execute(assign('claim-race'), owner);
  const input = { effectId: f.store.pendingEffects()[0].effect.effect_id, dispatcherId: 'same-dispatcher', now: at, leaseMs: 30_000 };
  const results = await Promise.all([childExecute(f.path, input, 'claimEffect'), childExecute(f.path, input, 'claimEffect')]);
  assert.equal(results.filter((result) => result.ok).length, 1);
  assert.equal(results.find((result) => !result.ok).code, 'EFFECT_ALREADY_CLAIMED');
});

test('registry configuration rejects coercion, unknown fields, duplicates and malformed triples', (t) => {
  const f = fixture(t);
  for (const options of [
    { worker_ids: 'worker-1' }, { approver_ids: 'operator' }, { workers: [] },
    { worker_ids: ['worker-1', 'worker-1'] }, { worker_ids: [''] },
    { trusted_checks: [{ check_id: 'check', check_version: 'v1' }] },
    { trusted_checks: [{ check_id: 'check', check_version: 'v1', runner_id: 'runner', extra: true }] },
    { trusted_checks: [{ check_id: 'check', check_version: 1, runner_id: 'runner' }] },
  ]) assert.throws(() => openStore(f.path, options), code('INVALID_CONFIGURATION'));
  const configuration = { worker_ids: ['worker-1'] };
  const frozenStore = openStore(f.path, configuration);
  try {
    configuration.worker_ids.length = 0;
    frozenStore.execute(command('ADMIT', 'policy-freeze', 0, { spec: spec() }), owner);
    assert.equal(frozenStore.execute(assign('policy-freeze'), owner).execution.state, 'ASSIGNED');
  } finally { frozenStore.close(); }
});

test('attempt identity cannot cross executions or be reused after cleanup and restart', (t) => {
  const f = fixture(t);
  for (const id of ['identity-one', 'identity-two']) f.store.execute(command('ADMIT', id, 0, { spec: spec(`CF-${id}`) }), owner);
  f.store.execute(assign('identity-one', 1, 'SAME-ATTEMPT'), owner);
  const duplicate = assign('identity-two', 1, 'SAME-ATTEMPT');
  assert.throws(() => f.store.execute(duplicate, owner), code('DUPLICATE_ATTEMPT'));
  assert.equal(f.store.get('identity-two').revision, 1);
  assert.equal(f.store.events().length, 3); assert.equal(f.store.pendingEffects().length, 1);
  f.store.execute(command('CANCEL', 'identity-one', 2, { code: 'OWNER_CANCELLED' }), owner);
  f.store.execute(command('CLEANUP_CONFIRMED', 'identity-one', 3, { attempt_id: 'SAME-ATTEMPT', fencing_token: 1 }), owner);
  f.store.close(); const reopened = f.open();
  assert.throws(() => reopened.execute(duplicate, owner), code('DUPLICATE_ATTEMPT'));
  assert.equal(reopened.execute(assign('identity-two', 1, 'NEW-ATTEMPT'), owner).execution.state, 'ASSIGNED');
  assert.equal(reopened.replay().executions.length, 2);
  const raw = new DatabaseSync(f.path); raw.prepare('DELETE FROM attempts WHERE attempt_id=?').run('SAME-ATTEMPT'); raw.close();
  assert.throws(() => reopened.replay(), code('REPLAY_MISMATCH'));
});

test('separate processes cannot assign one attempt identity to different executions', async (t) => {
  const f = fixture(t);
  for (const id of ['attempt-race-one', 'attempt-race-two']) f.store.execute(command('ADMIT', id, 0, { spec: spec(`CF-${id}`) }), owner);
  const results = await Promise.all([
    childExecute(f.path, assign('attempt-race-one', 1, 'GLOBAL-ATTEMPT')),
    childExecute(f.path, assign('attempt-race-two', 1, 'GLOBAL-ATTEMPT')),
  ]);
  assert.equal(results.filter((result) => result.ok).length, 1);
  assert.equal(results.find((result) => !result.ok).code, 'DUPLICATE_ATTEMPT');
  assert.equal(f.store.pendingEffects().length, 1); assert.equal(f.store.events().length, 3);
  assert.equal(f.store.replay().effects.length, 1);
});
