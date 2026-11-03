import test from 'node:test';
import assert from 'node:assert/strict';
import { chmodSync, mkdtempSync, realpathSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { createHash } from 'node:crypto';
import { openStore } from '../src/store.mjs';
import { canonical, digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

// Independently specified from the approved routing-integrity contract before
// production edits. Raw SQL models damaged owned fixtures, not supported writes.
const owner = { actor_id: 'operator', role: 'owner' };
const worker = { actor_id: 'worker-1', role: 'worker' };
const at = '2026-10-30T00:00:00.000Z';
const diagnostic = 'Stored execution routing fields disagree with its validated body';
let sequence = 0;
const sha = (text) => createHash('sha256').update(text, 'utf8').digest('hex');
function spec(id = 'CF-integrity') {
  const s = structuredClone(taskSpec); s.task_id = id;
  s.workflow.stages = [{ id: 'inspect', role: 'explorer-researcher', required_checks: [], requires_approval: false }];
  s.acceptance_checks = [];
  return s;
}
function command(kind, id, revision, payload) {
  return { schema_version: 1, command_id: `integrity-command-${++sequence}`, execution_id: id, expected_revision: revision, at, kind, payload };
}
function fixture(t) {
  const dir = realpathSync(mkdtempSync(join(tmpdir(), 'foundry-index-integrity-'))); chmodSync(dir, 0o700);
  const path = join(dir, 'authority.sqlite'); const stores = [];
  const open = () => { const s = openStore(path, { worker_ids: ['worker-1', 'reviewer-1'] }); stores.push(s); return s; };
  const store = open(); const raw = new DatabaseSync(path); raw.exec('PRAGMA foreign_keys=OFF');
  t.after(() => { raw.close(); for (const s of stores) { try { s.close(); } catch {} } rmSync(dir, { recursive: true, force: true }); });
  const admission = command('ADMIT', 'first', 0, { spec: spec() });
  const receipt = store.execute(admission, owner);
  return { store, raw, open, admission, receipt };
}
function tables(raw) {
  const names = raw.prepare("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").all().map((r) => r.name);
  const quoted = (name) => '"' + name.replaceAll('"', '""') + '"';
  return names.map((name) => {
    const columns = raw.prepare(`PRAGMA table_info(${quoted(name)})`).all().map((r) => r.name);
    const projection = columns.map((column) => `quote(${quoted(column)}) AS ${quoted(column)}`).join(',');
    return [name, raw.prepare(`SELECT ${projection} FROM ${quoted(name)}`).all().map((r) => canonical({ ...r })).sort()];
  });
}
function rejectUnchanged(f, operation) {
  const before = tables(f.raw);
  assert.throws(operation, (error) => {
    assert.equal(error.code, 'EXECUTION_INDEX_CORRUPTED');
    assert.equal(error.message, diagnostic);
    return true;
  });
  assert.deepEqual(tables(f.raw), before, 'refusal changed a logical row or stored byte string');
}
function body(f) { return JSON.parse(f.raw.prepare("SELECT body FROM executions WHERE execution_id='first'").get().body); }
function replaceBody(f, b) { f.raw.prepare("UPDATE executions SET body=? WHERE execution_id='first'").run(typeof b === 'string' ? b : canonical(b)); }
const corruptions = [
  ['execution-id', (f) => f.raw.exec("UPDATE executions SET execution_id='changed-execution'"), 'changed-execution'],
  ['task-id', (f) => f.raw.exec("UPDATE executions SET task_id='CF-hidden-task'"), 'first'],
  ['spec-digest', (f) => f.raw.prepare('UPDATE executions SET spec_digest=?').run('f'.repeat(64)), 'first'],
  ['known-state', (f) => f.raw.exec("UPDATE executions SET state='COMPLETED'"), 'first'],
  ['revision', (f) => f.raw.exec('UPDATE executions SET revision=8'), 'first'],
  ['revision-real', (f) => f.raw.exec('UPDATE executions SET revision=1.5'), 'first'],
  ['revision-text', (f) => f.raw.exec("UPDATE executions SET revision='invented invalid revision'"), 'first'],
  ['revision-wide-integer', (f) => f.raw.exec('UPDATE executions SET revision=9007199254740993'), 'first'],
  ['malformed-body', (f) => replaceBody(f, '{broken invented body'), 'first'],
  ['unknown-body-field', (f) => { const b = body(f); b.unexpected = 'invented private marker'; replaceBody(f, b); }, 'first'],
  ['missing-body-field', (f) => { const b = body(f); delete b.revision; replaceBody(f, b); }, 'first'],
  ['null-body-field', (f) => { const b = body(f); b.spec = null; replaceBody(f, b); }, 'first'],
  ['duplicate-body-field', (f) => replaceBody(f, '{"revision":1,' + canonical(body(f)).slice(1)), 'first'],
  ['changed-spec-content', (f) => { const b = body(f); b.spec.outcome = 'Invented changed immutable intent'; replaceBody(f, b); }, 'first'],
  ['matching-invalid-digest-index-and-body', (f) => { const b = body(f); b.spec_digest = 'e'.repeat(64); replaceBody(f, b); f.raw.prepare('UPDATE executions SET spec_digest=?').run(b.spec_digest); }, 'first'],
];

for (const [name, damage, lookup] of corruptions) {
  test(`routing integrity ${name}: selected reads, replay and cached receipt reject without mutation`, (t) => {
    const f = fixture(t); damage(f);
    rejectUnchanged(f, () => f.store.get(lookup));
    rejectUnchanged(f, () => f.store.list());
    rejectUnchanged(f, () => f.store.replay());
    rejectUnchanged(f, () => f.store.execute(f.admission, owner));
  });
}
for (const field of ['state', 'task_id']) {
  test(`routing integrity ${field}: hidden active task cannot be readmitted`, (t) => {
    const f = fixture(t);
    f.raw.exec(field === 'state' ? "UPDATE executions SET state='COMPLETED'" : "UPDATE executions SET task_id='CF-hidden-task'");
    const next = spec(); next.spec_revision = 2;
    rejectUnchanged(f, () => f.store.execute(command('ADMIT', 'replacement', 0, { spec: next }), owner));
  });
}
function assign(f, id = 'first') {
  return f.store.execute(command('ASSIGN', id, f.store.get(id).revision, {
    attempt_id: `attempt-${id}`, worker_id: 'worker-1', input_revisions: [], workspace: 'owned-read-only',
    capabilities: ['read'], lease_deadline: '2026-10-30T00:01:00.000Z',
  }), owner);
}
test('routing integrity guards transactional artifact and outbox mutations before work', (t) => {
  const f = fixture(t); assign(f);
  const effectId = f.store.pendingEffects()[0].effect.effect_id;
  const claim = f.store.claimEffect({ effectId, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 });
  f.raw.exec("UPDATE executions SET state='COMPLETED'");
  rejectUnchanged(f, () => f.store.putArtifact({ digest: sha('new artifact'), content: 'new artifact' }));
  rejectUnchanged(f, () => f.store.claimEffect({ effectId, dispatcherId: 'second', now: '2026-10-30T00:00:02.000Z', leaseMs: 1000 }));
  rejectUnchanged(f, () => f.store.acknowledgeEffect({ effectId, dispatcherId: 'dispatcher', claimToken: claim.claim_token, outcome: 'UNKNOWN', receipt: { observed: 'invented absence' } }));
});
test('routing integrity guards owner reconciliation before an outbox update', (t) => {
  const f = fixture(t); assign(f); const effectId = f.store.pendingEffects()[0].effect.effect_id;
  const claim = f.store.claimEffect({ effectId, dispatcherId: 'dispatcher', now: at, leaseMs: 1000 });
  f.store.acknowledgeEffect({ effectId, dispatcherId: 'dispatcher', claimToken: claim.claim_token, outcome: 'UNKNOWN', receipt: {} });
  f.raw.exec("UPDATE executions SET task_id='CF-hidden-task'");
  rejectUnchanged(f, () => f.store.reconcileEffect({ effectId, outcome: 'PENDING', receipt: { confirmed: 'absent' } }, owner));
});
test('routing integrity scans unrelated rows before a transactional cached receipt shortcut', (t) => {
  const f = fixture(t); f.store.execute(command('ADMIT', 'second', 0, { spec: spec('CF-second') }), owner);
  f.raw.exec("UPDATE executions SET task_id='CF-hidden-task' WHERE execution_id='second'");
  assert.equal(f.store.get('first').execution_id, 'first', 'selected reads do not claim to scan unselected rows');
  rejectUnchanged(f, () => f.store.execute(f.admission, owner));
  rejectUnchanged(f, () => f.store.dependencyOutputs('first'));
});
test('routing integrity checks dependency selection even when damaged indexes hide the dependency', (t) => {
  const f = fixture(t); const dependency = spec();
  const dependent = spec('CF-dependent'); dependent.dependencies = [{ task_id: dependency.task_id, spec_digest: digest(dependency) }];
  f.store.execute(command('ADMIT', 'dependent', 0, { spec: dependent }), owner);
  const b = body(f); b.state = 'COMPLETED'; b.accepted_subject_digest = 'a'.repeat(64);
  replaceBody(f, b); f.raw.exec("UPDATE executions SET state='COMPLETED' WHERE execution_id='first'");
  assert.deepEqual(f.store.dependencyOutputs('dependent'), [{ ...dependent.dependencies[0], execution_id: 'first', subject_digest: 'a'.repeat(64) }]);
  f.raw.exec("UPDATE executions SET task_id='CF-hidden-task' WHERE execution_id='first'");
  rejectUnchanged(f, () => f.store.dependencyOutputs('dependent'));
});
test('routing integrity preserves valid restart, old receipts and ordinary admission rules', (t) => {
  const f = fixture(t); assign(f); f.store.close(); const reopened = f.open();
  const before = tables(f.raw);
  assert.deepEqual(reopened.execute(f.admission, owner), f.receipt);
  assert.deepEqual(tables(f.raw), before);
  assert.equal(reopened.get('first').state, 'ASSIGNED'); assert.equal(reopened.list().length, 1);
  assert.equal(reopened.replay().executions[0].state, 'ASSIGNED');
  assert.equal(reopened.get('missing'), null);
  const next = spec(); next.spec_revision = 2;
  assert.throws(() => reopened.execute(command('ADMIT', 'replacement', 0, { spec: next }), owner), { code: 'TASK_ALREADY_ACTIVE' });
});
test('routing integrity accepts a valid stored execution larger than the new-command wire limit', (t) => {
  const f = fixture(t); const ids = Array.from({ length: 256 }, (_, n) => `check-${n}-`.padEnd(128, 'x'));
  const s = spec('CF-large-integrity'); s.outcome = '😀'.repeat(4096); s.acceptance_checks = ids;
  s.shared_resources = Array.from({ length: 256 }, (_, n) => ({ resource_id: `resource-${n}`.padEnd(128, 'x'), mode: 'shared' }));
  s.workflow.stages = Array.from({ length: 16 }, (_, n) => ({ id: `stage-${n}`, role: n === 0 ? 'builder-developer' : [1, 15].includes(n) ? 'reviewer-verifier' : 'explorer-researcher', required_checks: n === 0 ? [] : ids, requires_approval: false }));
  const send = (kind, payload, actor = owner) => f.store.execute(command(kind, 'large', f.store.get('large')?.revision ?? 0, payload), actor);
  send('ADMIT', { spec: s }); assign(f, 'large'); send('ACKNOWLEDGE', { attempt_id: 'attempt-large', fencing_token: 1 }, worker);
  const content = 'fixed large-fixture output'; f.store.putArtifact({ digest: sha(content), content });
  const artifacts = [{ path: 'result.md', digest: sha(content), repository_revision: s.repository.revision }]; const subject = digest(artifacts);
  send('SUBMIT', { result: { schema_version: 1, attempt_id: 'attempt-large', fencing_token: 1, spec_digest: digest(s), kind: 'SUCCEEDED', artifacts, subject_digest: subject } }, worker);
  send('ACCEPT_STAGE', { subject_digest: subject }); send('CLEANUP_CONFIRMED', { attempt_id: 'attempt-large', fencing_token: 1 });
  const reviewer = { actor_id: 'reviewer-1', role: 'worker' };
  const assigned = send('ASSIGN', { attempt_id: 'large-review', worker_id: reviewer.actor_id, input_revisions: [subject], workspace: 'read-only', capabilities: ['read'], lease_deadline: '2026-10-30T00:01:00.000Z' });
  const token = assigned.execution.attempt.fencing_token; send('ACKNOWLEDGE', { attempt_id: 'large-review', fencing_token: token }, reviewer);
  const submit = command('SUBMIT', 'large', f.store.get('large').revision, { result: { schema_version: 1, attempt_id: 'large-review', fencing_token: token, spec_digest: digest(s), kind: 'CHANGES_REQUIRED', subject_digest: subject, findings: Array.from({ length: 256 }, (_, n) => ({ code: `finding-${n}`, message: 'x'.repeat(2048) })) } });
  assert.ok(Buffer.byteLength(canonical(submit)) < 1_048_576);
  const receipt = f.store.execute(submit, reviewer); assert.ok(Buffer.byteLength(canonical(receipt.execution)) > 1_048_576);
  f.store.close(); const reopened = f.open();
  assert.deepEqual(reopened.execute(submit, reviewer), receipt);
  assert.deepEqual(reopened.get('large'), receipt.execution);
  assert.equal(reopened.list().length, 2); assert.equal(reopened.replay().executions.length, 2);
});
