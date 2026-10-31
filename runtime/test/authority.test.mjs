import test from 'node:test';
import assert from 'node:assert/strict';
import { execFileSync, spawn } from 'node:child_process';
import { chmodSync, mkdirSync, mkdtempSync, realpathSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { randomUUID } from 'node:crypto';
import { DatabaseSync } from 'node:sqlite';
import { openStore } from '../src/store.mjs';
import { digest } from '../src/contract.mjs';
import { verifyLocalPin } from '../src/authority.mjs';
import { makeSource } from '../../skills/project-memory/scripts/lib/state.mjs';
import { taskSpec } from './fixtures.mjs';

const owner = { actor_id: 'operator', role: 'owner' };
const at = '2026-10-30T00:00:00.000Z';
const code = (expected) => (error) => error.code === expected;
let sequence = 0;
const command = (kind, id, revision, payload) => ({ schema_version: 1, command_id: `authority-${++sequence}`, execution_id: id, expected_revision: revision, at, kind, payload });
const assignment = (id, revision = 1) => command('ASSIGN', id, revision, { attempt_id: `attempt-${id}`, worker_id: 'worker', lease_deadline: '2026-10-30T00:01:00.000Z', workspace: 'private-workspace', capabilities: ['read'], input_revisions: [] });

function fixture(t, { writable = false, guards = {}, prefix = 'CF-AUTO-' } = {}) {
  const directory = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-authority-')));
  chmodSync(directory, 0o700);
  const root = join(directory, 'source'); mkdirSync(root); mkdirSync(join(root, 'tasks'));
  const git = (...args) => execFileSync('git', args, { cwd: root, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
  git('init', '--quiet'); git('config', 'user.name', 'Fixture'); git('config', 'user.email', 'fixture@example.test');
  git('remote', 'add', 'origin', 'https://github.com/example/tasks.git');
  const path = join(directory, 'authority.sqlite');
  const stores = [];
  const open = (options = {}) => { const value = openStore(path, { worker_ids: ['worker'], ...guards, ...options }); stores.push(value); return value; };
  const store = open();
  t.after(() => { for (const value of stores) { try { value.close(); } catch {} } rmSync(directory, { recursive: true, force: true }); });
  const spec = structuredClone(taskSpec); spec.task_id = `${prefix}one`; spec.workflow.stages[0].role = 'explorer-researcher';
  if (writable) { spec.integration_policy = 'isolated_write'; spec.write_scope = ['result.txt']; spec.workflow.workflow_id = 'isolated-write-v1'; }
  const taskPath = `tasks/${spec.task_id}-example.md`;
  const record = `---\nid: ${spec.task_id}\nstatus: in-progress\nowner: operator\nrepositories: [${spec.repository.repository_id}]\n---\n\n# Bounded local task\n\n## Implementation claim\n\nOwner operator; inspect the exact fixture.\n`;
  writeFileSync(join(root, taskPath), record); git('add', '--', taskPath); git('commit', '--quiet', '-m', 'Record fixture claim');
  const base = { schema_version: 1, namespace_id: 'auto', task_prefix: prefix, repository: spec.repository.repository_id,
    source_repository: 'example/tasks', accountable_owner: 'operator', authority_instance_id: store.instanceId(),
    require_evidence: writable, allow_writes: writable, prior_authority_drained: true,
    claim_disposition: 'new_namespace', previous_authority: 'none: new fixture namespace', reason: 'Explicit fixture decision' };
  function proposal(mode, changes = {}) {
    const decision = { ...base, decision_id: `decision-${randomUUID()}`, expected_generation: store.namespaceStatus(changes.namespace_id ?? base.namespace_id)?.generation ?? 0, mode, ...changes };
    const relative = `tasks/${decision.decision_id}.json`; writeFileSync(join(root, relative), JSON.stringify(decision));
    git('add', '--', relative); git('commit', '--quiet', '-m', 'Record fixture authority decision');
    return { decision, source_root: root, source: { repository: 'example/tasks', revision: git('rev-parse', 'HEAD'), path: relative, blob: git('rev-parse', `HEAD:${relative}`) } };
  }
  const decide = (mode, changes = {}) => store.recordNamespaceDecision(proposal(mode, changes), owner);
  const activate = () => { decide('SHADOW'); return decide('MANAGED'); };
  function source() {
    const revision = git('rev-parse', 'HEAD');
    return makeSource({ repository: 'example/tasks', revision, id: spec.task_id, path: taskPath,
      records: new Map([[taskPath, { text: record, blob: git('rev-parse', `HEAD:${taskPath}`) }]]) });
  }
  const register = (value = spec, snapshot = source()) => store.registerAdmission({ namespace_id: 'auto', generation: store.namespaceStatus('auto').generation, spec: value, source: snapshot, source_root: root, reviewed_by: 'operator' }, owner);
  const admit = (id = 'execution') => store.execute(command('ADMIT', id, 0, { spec }), owner);
  function binding(id = 'execution') {
    const admission = register(); const source = admission.proposal.source;
    return { schema_version: 1, execution_id: id, spec_digest: digest(spec), task_id: spec.task_id, repository: spec.repository.repository_id,
      source_repository: source.repository, source_revision: source.revision, source_path: source.path, source_blob: source.blob,
      claim_heading: source.claim.heading, claim_digest: source.claim.digest, run_id: randomUUID(), start_event_id: randomUUID(), head_event_id: randomUUID(),
      start_event_digest: 'a'.repeat(64), head_event_digest: 'b'.repeat(64), files: ['result.txt'], code_root: root, source_root: root };
  }
  return { directory, root, path, git, store, open, spec, taskPath, record, source, proposal, decide, activate, register, admit, binding };
}

// Public fixture recipes use real committed JSON decisions and Markdown claims.
test('managed admission binds an actual local source and survives reopen and later draining', (t) => {
  const f = fixture(t); f.activate();
  assert.throws(() => f.admit(), code('REVIEWED_ADMISSION_REQUIRED'));
  f.register(); f.admit(); f.store.execute(assignment('execution'), owner);
  const replay = f.store.replay();
  f.decide('DRAINING');
  assert.deepEqual(f.store.replay(), replay);
  const reopened = f.open();
  assert.equal(reopened.instanceId(), f.store.instanceId());
  assert.equal(reopened.namespaceStatus('auto').decision.mode, 'DRAINING');
  assert.throws(() => reopened.execute(command('ADMIT', 'bypass', 0, { spec: { ...f.spec, spec_revision: 2 } }), owner), code('NAMESPACE_NOT_MANAGED'));
  assert.throws(() => reopened.claimEffect({ effectId: f.store.pendingEffects()[0].effect.effect_id, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 }), code('NAMESPACE_NOT_MANAGED'));
  assert.equal(f.store.pendingEffects()[0].status, 'PENDING');
});

test('namespace decisions verify actual commit bytes, origin, blob, instance and actor', (t) => {
  const f = fixture(t); const request = f.proposal('SHADOW');
  assert.throws(() => f.store.recordNamespaceDecision({ ...request, decision: { ...request.decision, reason: 'Forged' } }, owner), code('SOURCE_CONTENT_MISMATCH'));
  assert.throws(() => f.store.recordNamespaceDecision({ ...request, source: { ...request.source, blob: 'f'.repeat(40) } }, owner), code('SOURCE_BLOB_MISMATCH'));
  assert.throws(() => verifyLocalPin({ root: f.root, source: { ...request.source, repository: 'other/tasks' } }), code('SOURCE_REPOSITORY_MISMATCH'));
  assert.throws(() => f.store.recordNamespaceDecision(request, { actor_id: 'other', role: 'owner' }), code('UNAUTHORIZED'));
  const first = f.store.recordNamespaceDecision(request, owner);
  assert.deepEqual(f.store.recordNamespaceDecision(request, owner), first);
  const other = f.proposal('MANAGED', { authority_instance_id: randomUUID() });
  assert.throws(() => f.store.recordNamespaceDecision(other, owner), code('AUTHORITY_INSTANCE_MISMATCH'));
  assert.equal(f.store.namespaceStatus('auto').generation, 1);
});

test('overlapping namespaces, implicit activation, policy mutation and unconfirmed drain fail', (t) => {
  const f = fixture(t);
  assert.throws(() => f.decide('MANAGED'), code('ILLEGAL_AUTHORITY_TRANSITION'));
  f.decide('SHADOW');
  assert.throws(() => f.decide('SHADOW', { namespace_id: 'overlap', task_prefix: 'CF-' }), code('NAMESPACE_OVERLAP'));
  assert.throws(() => f.decide('MANAGED', { prior_authority_drained: false }), code('PRIOR_AUTHORITY_NOT_DRAINED'));
  assert.throws(() => f.decide('MANAGED', { require_evidence: true }), code('NAMESPACE_IMMUTABLE'));
  f.decide('MANAGED');
  assert.throws(() => f.decide('MANUAL'), code('ILLEGAL_AUTHORITY_TRANSITION'));
});

test('source registration refuses substituted snapshot, missing actual object and duplicate canonical IDs', (t) => {
  const f = fixture(t); f.activate(); const source = f.source();
  assert.throws(() => f.register(f.spec, { ...source, record: `${source.record}\nforged` }));
  assert.throws(() => f.register(f.spec, { ...source, revision: 'f'.repeat(40) }));
  const duplicate = `tasks/${f.spec.task_id}-duplicate.md`;
  writeFileSync(join(f.root, duplicate), f.record); f.git('add', '--', duplicate); f.git('commit', '--quiet', '-m', 'Duplicate fixture canonical ID');
  const forged = { ...source, revision: f.git('rev-parse', 'HEAD') };
  assert.throws(() => f.register(f.spec, forged), /ambiguous/);
  assert.equal(f.store.list().length, 0);
});

test('managed admission and dispatch reject a stale pin when the current claim changes', (t) => {
  const f = fixture(t); f.activate(); const snapshot = f.source(); f.register(f.spec, snapshot); f.admit();
  writeFileSync(join(f.root, f.taskPath), f.record.replace('status: in-progress', 'status: closed').replace('owner: operator', 'owner: other'));
  f.git('add', '--', f.taskPath); f.git('commit', '--quiet', '-m', 'Close and reassign fixture claim');
  assert.throws(() => f.register(f.spec, snapshot), code('CURRENT_CLAIM_MISMATCH'));
  assert.throws(() => f.store.execute(assignment('execution'), owner), code('CURRENT_CLAIM_MISMATCH'));
  assert.equal(f.store.get('execution').state, 'PENDING');
  assert.equal(f.store.replay().executions[0].state, 'PENDING');
});

test('current source owner and active status must match managed authority', (t) => {
  for (const [before, after, error] of [['owner: operator', 'owner: other', 'NAMESPACE_OWNER_MISMATCH'], ['status: in-progress', 'status: closed', 'CURRENT_CLAIM_NOT_ACTIVE']]) {
    const f = fixture(t); f.activate(); const text = f.record.replace(before, after);
    writeFileSync(join(f.root, f.taskPath), text); f.git('add', '--', f.taskPath); f.git('commit', '--quiet', '-m', 'Change fixture claim');
    const snapshot = makeSource({ repository: 'example/tasks', revision: f.git('rev-parse', 'HEAD'), id: f.spec.task_id, path: f.taskPath,
      records: new Map([[f.taskPath, { text, blob: f.git('rev-parse', `HEAD:${f.taskPath}`) }]]) });
    assert.throws(() => f.register(f.spec, snapshot), code(error));
  }
});

test('binding scope compares exact paths independent of scope array order', (t) => {
  const f = fixture(t, { writable: true }); f.spec.write_scope = ['z.txt', 'a.txt']; f.activate(); f.register(); f.admit();
  const binding = { ...f.binding(), files: ['a.txt', 'z.txt'] };
  assert.doesNotThrow(() => f.store.bindEvidence({ execution_id: 'execution', expected_revision: 1, binding }, owner));
});

test('legacy read-only tasks stay usable and cannot silently become managed claims', (t) => {
  const f = fixture(t);
  f.admit('legacy');
  f.decide('SHADOW');
  assert.throws(() => f.decide('MANAGED'), code('EXPLICIT_MIGRATION_REQUIRED'));
  assert.throws(() => f.store.execute(assignment('legacy'), owner), code('NAMESPACE_NOT_MANAGED'));
  const old = { ...f.spec, task_id: 'CF-001' };
  f.store.execute(command('ADMIT', 'old-cf', 0, { spec: old }), owner);
  f.store.execute(assignment('old-cf'), owner);
  assert.equal(f.store.get('old-cf').state, 'ASSIGNED');
});

test('isolated writes require a managed namespace, fresh evidence and exact attempt host proof', (t) => {
  let evidenceGood = true; let isolationGood = false;
  const f = fixture(t, { writable: true, guards: {
    evidence_guard: () => evidenceGood,
    isolation_guard: (execution, attempt) => ({ verified: true, execution_id: execution.execution_id, spec_digest: execution.spec_digest,
      attempt_id: attempt.attempt_id, workspace: attempt.workspace, repository_revision: execution.spec.repository.revision,
      write_scope: isolationGood ? execution.spec.write_scope : ['another.txt'] }),
  } });
  assert.throws(() => f.admit(), code('NAMESPACE_REQUIRED')); f.activate(); f.register(); f.admit();
  assert.throws(() => f.store.execute(assignment('execution'), owner), code('EVIDENCE_REQUIRED'));
  const binding = f.binding(); f.store.bindEvidence({ execution_id: 'execution', expected_revision: 1, binding }, owner);
  assert.equal(digest(f.store.evidenceBinding('execution')), digest(binding));
  assert.throws(() => f.store.execute(assignment('execution'), owner), code('ISOLATION_PROOF_MISMATCH'));
  isolationGood = true; f.store.execute(assignment('execution'), owner);
  const context = f.store.history().at(-1).context;
  assert.equal(context.namespace_managed, true); assert.equal(context.isolation_verified, true); assert.equal(context.evidence_binding_digest, digest(binding));
  assert.equal(JSON.stringify(context).includes(f.root), false);
  evidenceGood = false;
  assert.throws(() => f.store.claimEffect({ effectId: f.store.pendingEffects()[0].effect.effect_id, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 }), code('EVIDENCE_GUARD_FAILED'));
  assert.equal(f.store.pendingEffects()[0].status, 'PENDING');
  assert.equal(f.store.replay().executions[0].state, 'ASSIGNED');
});

test('binding is immutable, must precede assignment, and requires fresh guards after reopening', (t) => {
  const f = fixture(t, { writable: true, guards: { evidence_guard: () => true } }); f.activate(); f.register(); f.admit();
  const binding = f.binding(); f.store.bindEvidence({ execution_id: 'execution', expected_revision: 1, binding }, owner);
  assert.throws(() => f.store.bindEvidence({ execution_id: 'execution', expected_revision: 1, binding: { ...binding, head_event_id: randomUUID() } }, owner), code('EVIDENCE_BINDING_IMMUTABLE'));
  assert.throws(() => f.open({ evidence_guard: () => true }).execute(assignment('execution'), owner), code('ISOLATION_REQUIRED'));
  const reopened = openStore(f.path, { worker_ids: ['worker'] }); t.after(() => reopened.close());
  assert.throws(() => reopened.execute(assignment('execution'), owner), code('EVIDENCE_GUARD_REQUIRED'));
});

test('return to manual requires terminal executions and settled outbox, preserving replay', (t) => {
  const f = fixture(t); f.activate(); f.register(); f.admit(); f.decide('DRAINING');
  assert.throws(() => f.decide('MANUAL'), code('NAMESPACE_NOT_DRAINED'));
  f.store.execute(command('CANCEL', 'execution', 1, { code: 'RETURN_TO_MANUAL' }), owner);
  f.decide('MANUAL');
  assert.equal(f.store.replay().executions[0].state, 'CANCELLED');
  f.decide('SHADOW'); f.decide('MANAGED');
  assert.throws(() => f.admit('next-generation'), code('STALE_ADMISSION'));
});

test('draining after an outbox claim stops dispatch and unresolved outcomes block manual return', (t) => {
  const f = fixture(t); f.activate(); f.register(); f.admit(); f.store.execute(assignment('execution'), owner);
  const effectId = f.store.pendingEffects()[0].effect.effect_id;
  const claim = f.store.claimEffect({ effectId, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 });
  assert.equal(f.store.verifyDispatch('execution', 'attempt-execution'), true);
  f.decide('DRAINING');
  assert.throws(() => f.store.verifyDispatch('execution', 'attempt-execution'), code('NAMESPACE_NOT_MANAGED'));
  assert.equal(f.store.pendingEffects()[0].status, 'CLAIMED');
  f.store.acknowledgeEffect({ effectId, dispatcherId: 'dispatcher', claimToken: claim.claim_token, outcome: 'UNKNOWN', receipt: { reason: 'Host outcome requires owner review' } });
  f.store.execute(command('CANCEL', 'execution', 2, { code: 'RETURN_TO_MANUAL' }), owner);
  assert.throws(() => f.decide('MANUAL'), code('NAMESPACE_NOT_DRAINED'));
  f.store.execute(command('CLEANUP_CONFIRMED', 'execution', 3, { attempt_id: 'attempt-execution', fencing_token: 1 }), owner);
  assert.throws(() => f.decide('MANUAL'), code('NAMESPACE_NOT_DRAINED'));
  f.store.reconcileEffect({ effectId, outcome: 'DELIVERED', receipt: { operator_observed: 'Synthetic fixture outcome checked' } }, owner);
  for (const pending of f.store.pendingEffects()) {
    const cleanup = f.store.claimEffect({ effectId: pending.effect.effect_id, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 });
    f.store.acknowledgeEffect({ effectId: pending.effect.effect_id, dispatcherId: 'dispatcher', claimToken: cleanup.claim_token, outcome: 'DELIVERED', receipt: { fixture: 'settled' } });
  }
  f.decide('MANUAL'); assert.equal(f.store.replay().executions[0].state, 'CANCELLED');
});

test('migration preserves existing v1 event and receipt bytes and grants no new write authority', (t) => {
  const f = fixture(t); const legacy = { ...f.spec, task_id: 'CF-legacy' }; f.store.execute(command('ADMIT', 'legacy', 0, { spec: legacy }), owner);
  const raw = new DatabaseSync(f.path); const before = raw.prepare('SELECT body,command,context FROM events').all(); const receipts = raw.prepare('SELECT body FROM receipts').all();
  raw.exec('DROP TABLE isolation_proofs; DROP TABLE evidence_bindings; DROP TABLE execution_authority; DROP TABLE admission_sources; DROP TABLE admissions; DROP TABLE namespace_decisions; DROP TABLE namespaces; DROP TABLE authority_instance; PRAGMA user_version=1'); raw.close();
  const reopened = f.open(); const inspect = new DatabaseSync(f.path);
  assert.deepEqual(inspect.prepare('SELECT body,command,context FROM events').all(), before); assert.deepEqual(inspect.prepare('SELECT body FROM receipts').all(), receipts);
  inspect.close(); assert.equal(reopened.replay().executions.length, 1);
});

function childDecision(path, request) {
  return new Promise((resolve, reject) => {
    const script = `import {openStore} from ${JSON.stringify(new URL('../src/store.mjs', import.meta.url).href)};const s=openStore(process.argv[1]);try{s.recordNamespaceDecision(JSON.parse(process.argv[2]),{actor_id:'operator',role:'owner'});console.log(JSON.stringify({ok:true}));}catch(e){console.log(JSON.stringify({ok:false,code:e.code}));}finally{s.close();}`;
    const child = spawn(process.execPath, ['--input-type=module', '-e', script, path, JSON.stringify(request)], { stdio: ['ignore', 'pipe', 'pipe'] });
    let output = ''; let errors = ''; child.stdout.on('data', (data) => { output += data; }); child.stderr.on('data', (data) => { errors += data; });
    child.on('error', reject); child.on('close', (status) => status === 0 ? resolve(JSON.parse(output)) : reject(new Error(errors)));
  });
}

test('separate process namespace decisions have one generation-checked winner', async (t) => {
  const f = fixture(t); f.activate();
  const a = f.proposal('DRAINING'); const b = f.proposal('DRAINING');
  const results = await Promise.all([childDecision(f.path, a), childDecision(f.path, b)]);
  assert.equal(results.filter((result) => result.ok).length, 1);
  assert.equal(results.find((result) => !result.ok).code, 'GENERATION_CONFLICT');
  assert.equal(f.store.namespaceStatus('auto').generation, 3);
});
