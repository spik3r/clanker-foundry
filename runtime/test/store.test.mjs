import test from 'node:test';
import assert from 'node:assert/strict';
import { chmodSync, mkdtempSync, readFileSync, realpathSync, rmSync, symlinkSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawn } from 'node:child_process';
import { DatabaseSync } from 'node:sqlite';
import { createHash } from 'node:crypto';
import { openStore } from '../src/store.mjs';
import { canonical, digest, parseJSON } from '../src/contract.mjs';
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

test('parsed multi-resource specs survive persistence, assignment and replay', (t) => {
  const f = fixture(t);
  const input = spec('CF-multiple-resources');
  input.shared_resources = [
    { resource_id: 'one', mode: 'exclusive' }, { resource_id: 'two', mode: 'exclusive' },
  ];
  const parsed = parseJSON(JSON.stringify(input));
  f.store.execute(command('ADMIT', 'multi-resource', 0, { spec: parsed }), owner);
  f.store.close(); const reopened = f.open();
  assert.equal(reopened.execute(assign('multi-resource'), owner).execution.state, 'ASSIGNED');
  assert.equal(reopened.get('multi-resource').spec.shared_resources.length, 2);
  assert.equal(reopened.replay().executions.length, 1);
});

test('replay accepts a combined database response larger than the external wire limit', (t) => {
  const f = fixture(t);
  for (let index = 0; index < 200; index += 1) {
    const input = { ...spec(`CF-large-replay-${index}`), outcome: 'x'.repeat(4096) };
    f.store.execute(command('ADMIT', `large-replay-${index}`, 0, { spec: input }), owner);
  }
  const response = f.store.replay();
  assert.equal(response.executions.length, 200);
  assert.equal(response.events.length, 200);
  const serialized = JSON.stringify(response);
  assert.ok(Buffer.byteLength(serialized, 'utf8') > 1_048_576);
  assert.throws(() => parseJSON(serialized), code('JSON_TOO_LARGE'));
});

test('legal small commands can grow a stored aggregate beyond 1MiB without breaking recovery', (t) => {
  const ids = Array.from({ length: 256 }, (_, index) => `check-${index}-`.padEnd(128, 'x'));
  const f = fixture(t, { worker_ids: ['worker-1', 'reviewer-1'] });
  const input = spec('CF-growing-aggregate');
  input.outcome = '😀'.repeat(4096);
  input.acceptance_checks = ids;
  input.shared_resources = Array.from({ length: 256 }, (_, index) => ({
    resource_id: `resource-${index}`.padEnd(128, 'x'), mode: 'shared',
  }));
  input.workflow.stages = Array.from({ length: 16 }, (_, index) => ({ id: `stage-${index}`,
    role: index === 0 ? 'builder-developer' : [1, 15].includes(index) ? 'reviewer-verifier' : 'explorer-researcher',
    required_checks: index === 0 ? [] : ids, requires_approval: false,
  }));
  const send = (kind, payload, actor = owner) => {
    const value = command(kind, 'growing', f.store.get('growing')?.revision ?? 0, payload);
    assert.ok(Buffer.byteLength(canonical(value), 'utf8') < 1_048_576);
    return f.store.execute(value, actor);
  };
  send('ADMIT', { spec: input });
  f.store.execute(assign('growing'), owner);
  send('ACKNOWLEDGE', { attempt_id: 'attempt-growing', fencing_token: 1 }, worker);
  const content = 'stored output';
  const artifactDigest = createHash('sha256').update(content, 'utf8').digest('hex');
  f.store.putArtifact({ digest: artifactDigest, content });
  const artifacts = [{ path: 'output.txt', digest: artifactDigest, repository_revision: input.repository.revision }];
  const subject = digest(artifacts);
  send('SUBMIT', { result: {
    schema_version: 1, attempt_id: 'attempt-growing', fencing_token: 1, spec_digest: digest(input),
    kind: 'SUCCEEDED', subject_digest: subject, artifacts,
  } }, worker);
  send('ACCEPT_STAGE', { subject_digest: subject });
  send('CLEANUP_CONFIRMED', { attempt_id: 'attempt-growing', fencing_token: 1 });
  const reviewer = { actor_id: 'reviewer-1', role: 'worker' };
  const assigned = send('ASSIGN', { attempt_id: 'large-review', worker_id: reviewer.actor_id,
    lease_deadline: '2026-10-30T00:01:00.000Z', workspace: 'read-only',
    capabilities: ['read'], input_revisions: [subject] });
  const fencingToken = assigned.execution.attempt.fencing_token;
  send('ACKNOWLEDGE', { attempt_id: 'large-review', fencing_token: fencingToken }, reviewer);
  const submit = command('SUBMIT', 'growing', f.store.get('growing').revision, { result: {
    schema_version: 1, attempt_id: 'large-review', fencing_token: fencingToken,
    spec_digest: digest(input), kind: 'CHANGES_REQUIRED', subject_digest: subject,
    findings: Array.from({ length: 256 }, (_, index) => ({ code: `finding-${index}`, message: 'x'.repeat(2048) })),
  } });
  assert.ok(Buffer.byteLength(canonical(submit), 'utf8') < 1_048_576);
  const receipt = f.store.execute(submit, reviewer);
  assert.ok(Buffer.byteLength(canonical(receipt.execution), 'utf8') > 1_048_576);
  f.store.close(); const reopened = f.open();
  assert.deepEqual(reopened.execute(submit, reviewer), receipt);
  assert.equal(reopened.get('growing').result.findings.length, 256);
  assert.equal(reopened.history('growing').length, 9);
  assert.equal(reopened.replay().executions.length, 1);
  const oversized = structuredClone(submit);
  oversized.command_id = 'over-wire-limit';
  oversized.payload.result.findings = oversized.payload.result.findings.map((item) => ({ ...item, message: '😀'.repeat(2048) }));
  assert.throws(() => reopened.execute(oversized, reviewer), code('JSON_TOO_LARGE'));
  assert.equal(reopened.list().length, 1);
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
    { evidence_guard: true }, { isolation_guard: {} },
    Object.create({ inherited: true }),
  ]) assert.throws(() => openStore(f.path, options), code('INVALID_CONFIGURATION'));
  const configuration = { worker_ids: ['worker-1'] };
  const frozenStore = openStore(f.path, configuration);
  try {
    configuration.worker_ids.length = 0;
    frozenStore.execute(command('ADMIT', 'policy-freeze', 0, { spec: spec() }), owner);
    assert.equal(frozenStore.execute(assign('policy-freeze'), owner).execution.state, 'ASSIGNED');
  } finally { frozenStore.close(); }
});

test('trusted guard callbacks stay outside cloned registry data and cannot be supplied through accessors', (t) => {
  const f = fixture(t);
  const callback = () => true;
  const guarded = openStore(f.path, { worker_ids: ['worker-1'], evidence_guard: callback, isolation_guard: callback });
  try {
    guarded.execute(command('ADMIT', 'guard-config', 0, { spec: spec('CF-guard-config') }), owner);
    guarded.execute(assign('guard-config'), owner);
    assert.equal(guarded.verifyDispatch('guard-config', 'attempt-guard-config'), true);
    assert.throws(() => guarded.verifyDispatch('guard-config', 'another-attempt'), code('STALE_ATTEMPT'));
  } finally { guarded.close(); }
  let reads = 0;
  const configuration = Object.defineProperty({}, 'evidence_guard', { enumerable: true, get() { reads += 1; return callback; } });
  assert.throws(() => openStore(f.path, configuration), code('INVALID_CONFIGURATION'));
  assert.equal(reads, 0);
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

function completedInput(f, id, content) {
  const value = spec(`CF-${id}`);
  value.acceptance_checks = [];
  value.workflow.stages[0].required_checks = [];
  value.workflow.stages[0].requires_approval = false;
  const send = (kind, payload, actor = owner) => f.store.execute(command(kind, id, f.store.get(id)?.revision ?? 0, payload), actor);
  send('ADMIT', { spec: value });
  f.store.execute(assign(id), owner);
  send('ACKNOWLEDGE', { attempt_id: `attempt-${id}`, fencing_token: 1 }, worker);
  const hash = createHash('sha256').update(content).digest('hex');
  f.store.putArtifact({ digest: hash, content });
  const artifacts = [{ path: `${id}.txt`, digest: hash, repository_revision: value.repository.revision }];
  send('SUBMIT', { result: { schema_version: 1, attempt_id: `attempt-${id}`, fencing_token: 1,
    spec_digest: digest(value), kind: 'SUCCEEDED', artifacts, subject_digest: digest(artifacts) } }, worker);
  send('ACCEPT_STAGE', { subject_digest: digest(artifacts) });
  send('CLEANUP_CONFIRMED', { attempt_id: `attempt-${id}`, fencing_token: 1 });
  return { task_id: value.task_id, spec_digest: digest(value) };
}

function assignedDependent(f, id, dependencies) {
  const value = spec(`CF-${id}`);
  value.dependencies = dependencies;
  value.acceptance_checks = [];
  value.workflow.stages[0].required_checks = [];
  value.workflow.stages[0].requires_approval = false;
  f.store.execute(command('ADMIT', id, 0, { spec: value }), owner);
  f.store.execute(assign(id), owner);
}

async function settleCompletedInputs(store) {
  const { drainOutbox } = await import('../src/runner.mjs');
  while (store.pendingEffects().some((effect) => effect.status === 'PENDING')) {
    const answer = await drainOutbox({ store, dispatcherId: 'dependency-cleanup', now: at, limit: 256 });
    assert.equal(answer.unknown.length, 0);
    assert.ok(answer.delivered.length > 0);
  }
}

test('65 and 256 distinct completed dependencies reach the text worker and replay', async (t) => {
  const { drainOutbox } = await import('../src/runner.mjs');
  const { createTextModelAdapter } = await import('../src/adapters.mjs');
  const f = fixture(t);
  const dependencies = Array.from({ length: 256 }, (_, index) => completedInput(f, `many-${index}`, `input-${index}`));
  await settleCompletedInputs(f.store);
  for (const count of [65, 256]) {
    const id = `dependent-${count}`;
    assignedDependent(f, id, dependencies.slice(0, count));
    let invocations = 0;
    const adapter = createTextModelAdapter({ workerId: 'worker-1', invoke: async ({ inputSubjects }) => {
      invocations += 1;
      assert.equal(inputSubjects.length, count);
      assert.equal(new Set(inputSubjects.map((subject) => subject.subject_digest)).size, count);
      assert.equal(inputSubjects[count - 1].artifacts[0].content, `input-${count - 1}`);
      return 'reviewed dependency outputs';
    } });
    const answer = await drainOutbox({ store: f.store, adapter, dispatcherId: `dispatch-${count}`, now: at });
    assert.equal(answer.unknown.length, 0);
    assert.equal(invocations, 1);
    assert.equal(f.store.get(id).state, 'VALIDATING');
    f.store.execute(command('ACCEPT_STAGE', id, f.store.get(id).revision, { subject_digest: f.store.get(id).result.subject_digest }), owner);
    await drainOutbox({ store: f.store, adapter, dispatcherId: `cleanup-${count}`, now: at });
    assert.equal(f.store.get(id).cleanup_pending, false);
  }
  assert.equal(f.store.replay().executions.length, 258);
});

test('oversized immutable inputs stay unclaimed and unacknowledged with no unknown launch', async (t) => {
  const { drainOutbox } = await import('../src/runner.mjs');
  const f = fixture(t);
  const dependencies = [completedInput(f, 'large-a', 'a'.repeat(600_000)), completedInput(f, 'large-b', 'b'.repeat(600_000))];
  await settleCompletedInputs(f.store);
  assignedDependent(f, 'oversized-inputs', dependencies);
  const before = f.store.get('oversized-inputs');
  let invocations = 0;
  for (const round of [1, 2]) {
    const answer = await drainOutbox({ store: f.store, dispatcherId: `oversized-${round}`, now: at,
      adapter: { async runAttempt() { invocations += 1; } } });
    assert.equal(answer.unknown.length, 0);
    assert.equal(answer.skipped[0].code, 'INPUT_TOO_LARGE');
    assert.deepEqual(f.store.get('oversized-inputs'), before);
    const launch = f.store.pendingEffects().find((effect) => effect.effect.execution_id === 'oversized-inputs');
    assert.equal(launch.status, 'PENDING');
    assert.equal(launch.dispatcher_id, null);
  }
  assert.equal(invocations, 0);
  assert.equal(before.state, 'ASSIGNED');
  // A rejected oldest input must not monopolize a bounded dispatch drain.
  const { FakeAdapter } = await import('../src/adapters.mjs');
  assignedDependent(f, 'later-ready', []);
  const adapter = new FakeAdapter();
  const later = await drainOutbox({ store: f.store, adapter, dispatcherId: 'later-ready', now: at, limit: 1 });
  assert.equal(later.delivered.length, 1);
  assert.equal(later.unknown.length, 0);
  assert.equal(f.store.get('later-ready').state, 'VALIDATING');
  f.store.execute(command('ACCEPT_STAGE', 'later-ready', f.store.get('later-ready').revision,
    { subject_digest: f.store.get('later-ready').result.subject_digest }), owner);
  for (let round = 0; round < 2; round += 1) {
    const cleanup = await drainOutbox({ store: f.store, adapter, dispatcherId: `later-cleanup-${round}`, now: at, limit: 1 });
    assert.equal(cleanup.unknown.length, 0);
  }
  assert.equal(f.store.get('later-ready').cleanup_pending, false);
  assert.equal(f.store.get('later-ready').resources_held, false);
  assert.deepEqual(f.store.get('oversized-inputs'), before);
});

test('a missing dependency manifest blocks dispatch across reopen and recovers after restoration', async (t) => {
  const { drainOutbox } = await import('../src/runner.mjs');
  const { createTextModelAdapter, FakeAdapter } = await import('../src/adapters.mjs');
  const f = fixture(t);
  const dependency = completedInput(f, 'required-upstream', 'required evidence');
  await settleCompletedInputs(f.store);
  const subjectDigest = f.store.get('required-upstream').accepted_subject_digest;
  // A mandatory subject must not disappear when its digest also names the Git base.
  const downstream = spec('CF-missing-dependent');
  downstream.repository.revision = subjectDigest;
  downstream.dependencies = [dependency];
  f.store.execute(command('ADMIT', 'missing-dependent', 0, { spec: downstream }), owner);
  const assignment = assign('missing-dependent'); assignment.payload.input_revisions = [subjectDigest];
  f.store.execute(assignment, owner);
  const original = f.store.getSubject(subjectDigest);
  const before = f.store.get('missing-dependent');
  f.store.close();
  const raw = new DatabaseSync(f.path);
  raw.prepare('DELETE FROM subjects WHERE digest=?').run(subjectDigest);
  raw.close();
  f.store = f.open();
  let invocations = 0;
  const adapter = createTextModelAdapter({ workerId: 'worker-1', invoke: async ({ inputSubjects }) => {
    invocations += 1;
    assert.equal(inputSubjects.length, 1);
    assert.equal(inputSubjects[0].subject_digest, subjectDigest);
    assert.equal(inputSubjects[0].artifacts[0].content, 'required evidence');
    return 'reviewed required evidence';
  } });
  for (let round = 0; round < 2; round += 1) {
    const rejected = await drainOutbox({ store: f.store, adapter, dispatcherId: `missing-${round}`, now: at, limit: 1 });
    assert.equal(rejected.skipped[0].code, 'SUBJECT_NOT_FOUND');
    assert.equal(rejected.unknown.length, 0);
    assert.equal(rejected.delivered.length, 0);
    assert.deepEqual(f.store.get('missing-dependent'), before);
    const launch = f.store.pendingEffects().find((row) => row.effect.execution_id === 'missing-dependent');
    assert.equal(launch.status, 'PENDING');
    assert.equal(launch.dispatcher_id, null);
  }
  assert.equal(invocations, 0);
  assignedDependent(f, 'missing-later-ready', []);
  const later = await drainOutbox({ store: f.store, adapter: new FakeAdapter(), dispatcherId: 'later-missing', now: at, limit: 1 });
  assert.equal(later.delivered.length, 1);
  assert.equal(f.store.get('missing-later-ready').state, 'VALIDATING');
  const restored = new DatabaseSync(f.path);
  restored.prepare('INSERT INTO subjects(digest,manifest) VALUES(?,?)').run(subjectDigest, canonical(original));
  restored.close();
  const delivered = await drainOutbox({ store: f.store, adapter, dispatcherId: 'restored-input', now: at, limit: 1 });
  assert.equal(delivered.unknown.length, 0);
  assert.equal(invocations, 1);
  assert.equal(f.store.get('missing-dependent').state, 'VALIDATING');
  assert.equal(f.store.replay().executions.length, 3);
});

test('explicit inputs require subjects except for the exact pinned SHA-1 or SHA-256 Git base', async (t) => {
  const { drainOutbox } = await import('../src/runner.mjs');
  const { createTextModelAdapter } = await import('../src/adapters.mjs');
  for (const [label, base, revisions, expected] of [
    ['missing-subject', 'a'.repeat(40), ['b'.repeat(64)], 'SUBJECT_NOT_FOUND'],
    ['unknown-git-base', 'a'.repeat(40), ['b'.repeat(40)], 'UNSUPPORTED_INPUT_REVISION'],
    ['abbreviated-base', 'a'.repeat(40), ['a'.repeat(7)], 'UNSUPPORTED_INPUT_REVISION'],
    ['short-pinned-base', 'a'.repeat(7), ['a'.repeat(7)], 'UNSUPPORTED_INPUT_REVISION'],
    ['sha1-base', 'a'.repeat(40), ['a'.repeat(40)], null],
    ['sha256-base', 'a'.repeat(64), ['a'.repeat(64)], null],
  ]) {
    const f = fixture(t);
    const value = spec(`CF-${label}`); value.repository.revision = base;
    f.store.execute(command('ADMIT', label, 0, { spec: value }), owner);
    const assignment = assign(label); assignment.payload.input_revisions = revisions;
    f.store.execute(assignment, owner);
    const before = f.store.get(label);
    let invocations = 0;
    const adapter = createTextModelAdapter({ workerId: 'worker-1', invoke: async ({ inputSubjects, attempt }) => {
      invocations += 1; assert.deepEqual(inputSubjects, []); assert.deepEqual(attempt.input_revisions, revisions); return 'checked base';
    } });
    const result = await drainOutbox({ store: f.store, adapter, dispatcherId: label, now: at });
    assert.equal(result.unknown.length, 0);
    if (expected) {
      assert.equal(invocations, 0); assert.equal(result.skipped[0].code, expected);
      assert.deepEqual(f.store.get(label), before);
      assert.equal(f.store.pendingEffects()[0].status, 'PENDING');
    } else {
      assert.equal(invocations, 1); assert.equal(f.store.get(label).state, 'VALIDATING');
    }
  }
});

test('a missing completed dependency execution cannot become an empty input list', async (t) => {
  const { drainOutbox } = await import('../src/runner.mjs');
  const f = fixture(t);
  const dependency = completedInput(f, 'lost-upstream', 'required bytes');
  await settleCompletedInputs(f.store);
  assignedDependent(f, 'lost-dependent', [dependency]);
  const before = f.store.get('lost-dependent');
  // Simulate an incomplete restore in this owned fixture, not a supported mutation.
  const raw = new DatabaseSync(f.path);
  raw.exec('PRAGMA foreign_keys=OFF');
  raw.prepare('DELETE FROM executions WHERE execution_id=?').run('lost-upstream');
  raw.close();
  let invocations = 0;
  const result = await drainOutbox({ store: f.store, dispatcherId: 'lost-output', now: at,
    adapter: { async runAttempt() { invocations += 1; } } });
  assert.equal(result.skipped[0].code, 'DEPENDENCY_OUTPUT_NOT_FOUND');
  assert.equal(result.unknown.length, 0);
  assert.equal(invocations, 0);
  assert.deepEqual(f.store.get('lost-dependent'), before);
  assert.equal(f.store.pendingEffects().find((row) => row.effect.execution_id === 'lost-dependent').status, 'PENDING');
});
