import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, realpath, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { openStore } from '../src/store.mjs';
import { drainOutbox } from '../src/runner.mjs';
import { FakeAdapter } from '../src/adapters.mjs';
import { taskSpec } from './fixtures.mjs';

const origin = Date.parse('2026-11-11T00:00:00.000Z');
const owner = { actor_id: 'operator', role: 'owner' };
const trusted = ['first-check', 'second-check'].map(check_id => ({
  check_id, check_version: '1', runner_id: 'validator',
}));

async function fixture(t, checkIds = ['first-check']) {
  const directory = await realpath(await mkdtemp(join(tmpdir(), 'foundry-check-completion-')));
  const store = openStore(join(directory, 'runtime.sqlite'), {
    worker_ids: ['worker'], approver_ids: ['operator'], trusted_checks: trusted,
  });
  t.after(async () => { store.close(); await rm(directory, { recursive: true, force: true }); });
  let milliseconds = origin;
  let sequence = 0;
  const now = () => new Date(milliseconds).toISOString();
  const current = () => store.get('execution');
  const send = (kind, payload, actor = owner) => store.execute({
    schema_version: 1, command_id: `command-${++sequence}`, execution_id: 'execution',
    expected_revision: current()?.revision ?? 0, at: now(), kind, payload,
  }, actor);
  const spec = structuredClone(taskSpec);
  spec.acceptance_checks = checkIds;
  spec.workflow.stages[0].required_checks = checkIds;
  const assign = attemptId => send('ASSIGN', {
    attempt_id: attemptId, worker_id: 'worker', input_revisions: [spec.repository.revision],
    workspace: 'read-only', capabilities: ['read'],
    lease_deadline: new Date(milliseconds + spec.lease_ms).toISOString(),
  });
  send('ADMIT', { spec });
  assign('first-attempt');
  await drainOutbox({ store, adapter: new FakeAdapter(), dispatcherId: 'launcher', now });
  assert.equal(current().state, 'VALIDATING');
  const approval = () => ({
    approval_id: 'owner-approval', actor_id: 'operator', action: 'ACCEPT_STAGE', execution_id: 'execution',
    spec_digest: current().spec_digest, workflow_id: spec.workflow.workflow_id, workflow_version: 1,
    subject_digest: current().result.subject_digest, decision: 'APPROVE', revoked: false,
    expires_at: new Date(milliseconds + spec.validation_ms).toISOString(),
  });
  const evidence = (checkId, verdict = 'PASS') => ({
    ...trusted.find(check => check.check_id === checkId),
    subject_digest: current().result.subject_digest, repository_revision: spec.repository.revision,
    attempt_id: current().attempt.attempt_id, spec_digest: current().spec_digest,
    workflow_id: spec.workflow.workflow_id, workflow_version: 1, verdict, at: now(),
    evidence_ref: `external-${checkId}`,
  });
  return { store, current, now, send, assign, approval, evidence,
    advanceTo(value) { milliseconds = Date.parse(value); },
    tick() { milliseconds += 1000; },
  };
}

async function pendingCheck(t, f) {
  const started = Promise.withResolvers();
  const finished = Promise.withResolvers();
  t.after(() => finished.resolve('PASS'));
  const calls = [];
  const running = drainOutbox({ store: f.store, dispatcherId: 'checker', now: f.now,
    checks: trusted.map(check => ({ ...check, run() {
      calls.push(check.check_id);
      if (check.check_id !== 'first-check') return 'PASS';
      started.resolve();
      return finished.promise;
    } })),
  });
  await started.promise;
  return { running, calls, finish: () => finished.resolve('PASS') };
}

test('an unchanged check records once and remains replayable', async t => {
  const f = await fixture(t);
  const pending = await pendingCheck(t, f);
  pending.finish();
  const result = await pending.running;
  assert.deepEqual(result.unknown, []);
  assert.equal(result.delivered.length, 1);
  assert.deepEqual(pending.calls, ['first-check']);
  assert.deepEqual(f.current().validations.map(item => item.verdict), ['PASS']);
  assert.doesNotThrow(() => f.store.replay('execution'));
});

for (const revoke of [false, true]) {
  test(`owner approval${revoke ? ' and revocation' : ''} during a check preserves its result`, async t => {
    const f = await fixture(t);
    const pending = await pendingCheck(t, f);
    f.tick();
    f.send('RECORD_APPROVAL', { approval: f.approval() });
    if (revoke) f.send('REVOKE_APPROVAL', { approval_id: 'owner-approval', reason: 'changed-decision' });
    const revision = f.current().revision;
    pending.finish();
    const result = await pending.running;
    assert.deepEqual(result.unknown, []);
    assert.equal(result.delivered.length, 1);
    assert.deepEqual(pending.calls, ['first-check']);
    assert.equal(f.current().revision, revision + 1);
    assert.deepEqual(f.current().validations.map(item => item.verdict), ['PASS']);
    assert.equal(f.current().approvals[0].revoked, revoke);
    assert.equal(f.current().state, 'VALIDATING');
    assert.doesNotThrow(() => f.store.replay('execution'));
    await drainOutbox({ store: f.store, dispatcherId: 'checker-again', now: f.now });
    assert.equal(f.store.events('execution').filter(event => event.kind === 'RECORD_VALIDATION').length, 1);
  });
}

test('another check recorded during the callback is retained and not invoked again', async t => {
  const f = await fixture(t, ['first-check', 'second-check']);
  const pending = await pendingCheck(t, f);
  f.tick();
  const external = f.evidence('second-check');
  f.send('RECORD_VALIDATION', { evidence: external }, { actor_id: 'validator', role: 'validator' });
  pending.finish();
  const result = await pending.running;
  assert.deepEqual(result.unknown, []);
  assert.equal(result.delivered.length, 1);
  assert.deepEqual(pending.calls, ['first-check']);
  assert.deepEqual(f.current().validations.map(item => item.check_id), ['second-check', 'first-check']);
  assert.deepEqual(structuredClone(f.current().validations[0]), external);
  assert.doesNotThrow(() => f.store.replay('execution'));
});

test('a concurrent result for the same check is not replaced by the pending callback', async t => {
  const f = await fixture(t);
  const pending = await pendingCheck(t, f);
  f.tick();
  const external = f.evidence('first-check', 'FAIL');
  f.send('RECORD_VALIDATION', { evidence: external }, { actor_id: 'validator', role: 'validator' });
  pending.finish();
  const result = await pending.running;
  assert.equal(result.unknown[0]?.receipt?.code, 'DUPLICATE_EVIDENCE');
  assert.deepEqual(structuredClone(f.current().validations), [external]);
  assert.deepEqual(pending.calls, ['first-check']);
  assert.doesNotThrow(() => f.store.replay('execution'));
});

test('cancellation while a check is pending still refuses its result and retains cleanup', async t => {
  const f = await fixture(t);
  const pending = await pendingCheck(t, f);
  f.tick();
  f.send('CANCEL', { code: 'OWNER_CANCELLED' });
  pending.finish();
  const result = await pending.running;
  assert.equal(result.unknown[0]?.receipt?.status, 'ATTEMPT_CHANGED_DURING_CHECKS');
  assert.equal(f.current().state, 'CANCELLED');
  assert.equal(f.current().validations.length, 0);
  assert.equal(f.current().cleanup_pending, true);
  assert.equal(f.current().resources_held, true);
  assert.doesNotThrow(() => f.store.replay('execution'));
});

test('a replaced attempt cannot receive a check result from the earlier attempt', async t => {
  const f = await fixture(t);
  const pending = await pendingCheck(t, f);
  const retired = f.current().attempt;
  f.tick();
  const retryAt = new Date(Date.parse(f.now()) + 1000).toISOString();
  f.send('RETRYABLE_ERROR', { code: 'RESTART_REQUIRED', retry_at: retryAt });
  f.send('CLEANUP_CONFIRMED', { attempt_id: retired.attempt_id, fencing_token: retired.fencing_token });
  f.advanceTo(retryAt);
  f.send('RETRY_TIMER', {});
  f.assign('replacement-attempt');
  await drainOutbox({ store: f.store, adapter: new FakeAdapter(), dispatcherId: 'replacement-launcher', now: f.now });
  assert.equal(f.current().state, 'VALIDATING');
  assert.equal(f.current().attempt.attempt_id, 'replacement-attempt');
  pending.finish();
  const result = await pending.running;
  assert.ok(result.unknown.some(item => item.receipt?.status === 'ATTEMPT_CHANGED_DURING_CHECKS'));
  assert.equal(f.current().validations.length, 0);
  assert.equal(f.current().resources_held, true);
  assert.deepEqual(pending.calls, ['first-check']);
  assert.doesNotThrow(() => f.store.replay('execution'));
});

test('an approval during the callback does not extend the validation deadline', async t => {
  const f = await fixture(t);
  const pending = await pendingCheck(t, f);
  f.tick();
  f.send('RECORD_APPROVAL', { approval: f.approval() });
  f.advanceTo(f.current().validation_deadline);
  pending.finish();
  const result = await pending.running;
  assert.equal(result.unknown[0]?.receipt?.code, 'ATTEMPT_EXPIRED');
  assert.equal(f.current().validations.length, 0);
  assert.equal(f.current().resources_held, true);
  assert.deepEqual(pending.calls, ['first-check']);
  assert.doesNotThrow(() => f.store.replay('execution'));
});
