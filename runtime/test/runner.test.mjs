import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, realpath, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { openStore } from '../src/store.mjs';
import { drainOutbox } from '../src/runner.mjs';
import { FakeAdapter, createTextModelAdapter } from '../src/adapters.mjs';
import { digest } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

const origin = '2026-10-30T00:00:00.000Z';
const owner = { actor_id: 'operator', role: 'owner' };
const workerId = 'worker-1';
const check = {
  check_id: 'pack-validation',
  check_version: 'v1',
  runner_id: 'check-runner',
  run: async () => ({ verdict: 'PASS', evidence_ref: 'check:pack-validation:v1' }),
};
let caseNumber = 0;

async function setup(t, { leaseMs = 60_000, stages } = {}) {
  const directory = await mkdtemp(path.join(await realpath(os.tmpdir()), 'foundry-runner-'));
  const store = openStore(path.join(directory, 'runtime.sqlite'), {
    worker_ids: [workerId, 'reviewer-1'],
    trusted_checks: [{ check_id: check.check_id, check_version: check.check_version, runner_id: check.runner_id }],
  });
  t.after(() => {
    store.close();
    return rm(directory, { recursive: true, force: true });
  });
  const spec = structuredClone(taskSpec);
  caseNumber += 1;
  spec.task_id = `CF-runner-${caseNumber}`;
  spec.spec_revision = 1;
  spec.lease_ms = leaseMs;
  spec.workflow.stages = stages ?? [{
    id: 'explore', role: 'explorer-researcher', required_checks: [check.check_id], requires_approval: false,
  }];
  const executionId = `exec-${spec.task_id}`;
  store.execute({
    schema_version: 1,
    command_id: `admit-${spec.task_id}`,
    execution_id: executionId,
    expected_revision: 0,
    at: origin,
    kind: 'ADMIT',
    payload: { spec },
  }, owner);
  const assignmentAt = Date.parse(origin) + 1_000;
  const leaseDeadline = new Date(assignmentAt + leaseMs).toISOString();
  store.execute({
    schema_version: 1,
    command_id: `assign-${spec.task_id}`,
    execution_id: executionId,
    expected_revision: 1,
    at: new Date(assignmentAt).toISOString(),
    kind: 'ASSIGN',
    payload: {
      attempt_id: `attempt-${spec.task_id}`,
      worker_id: workerId,
      input_revisions: [spec.repository.revision],
      workspace: 'workspace-fixture',
      capabilities: ['read', 'analyze'],
      lease_deadline: leaseDeadline,
    },
  }, owner);
  return { store, spec, executionId, leaseDeadline };
}

const at = '2026-10-30T00:00:02.000Z';

test('drains launch and trusted checks through durable worker/validator commands', async (t) => {
  const { store, executionId } = await setup(t);
  const adapter = new FakeAdapter();
  const launched = await drainOutbox({ store, adapter, dispatcherId: 'dispatcher-happy', now: at });
  assert.equal(launched.delivered.length, 1);
  assert.equal(store.get(executionId).state, 'VALIDATING');
  assert.equal(store.get(executionId).result.kind, 'SUCCEEDED');

  let checkedContent;
  const contentCheck = {
    ...check,
    async run(input) {
      checkedContent = input.artifacts[0].content;
      return { verdict: 'PASS', evidence_ref: 'check:pack-validation:v1' };
    },
  };
  const checked = await drainOutbox({ store, checks: [contentCheck], dispatcherId: 'dispatcher-happy', now: at });
  assert.equal(checked.delivered.length, 1);
  assert.equal(checkedContent, 'fake worker result');
  assert.deepEqual(store.get(executionId).validations.map((item) => item.verdict), ['PASS']);
  assert.equal(store.get(executionId).state, 'VALIDATING');
  assert.equal(store.events(executionId).some((event) => event.kind === 'ACCEPT_STAGE'), false);
});

test('unknown adapter outcomes become non-replayable and do not invoke the adapter again', async (t) => {
  const { store, executionId } = await setup(t);
  let calls = 0;
  const adapter = { async runAttempt() { calls += 1; throw new Error('network result unknown'); } };
  const first = await drainOutbox({ store, adapter, dispatcherId: 'dispatcher-unknown', now: at });
  assert.equal(first.unknown.length, 1);
  assert.equal(store.get(executionId).state, 'EXECUTING');
  assert.equal(calls, 1);
  const second = await drainOutbox({ store, adapter, dispatcherId: 'dispatcher-unknown', now: at });
  assert.equal(second.delivered.length, 0);
  assert.equal(second.unknown.length, 1);
  assert.equal(calls, 1);
});

test('cancellation suppresses a stale launch before the adapter is called', async (t) => {
  const { store, executionId } = await setup(t);
  const current = store.get(executionId);
  store.execute({
    schema_version: 1,
    command_id: 'cancel-before-launch',
    execution_id: executionId,
    expected_revision: current.revision,
    at,
    kind: 'CANCEL',
    payload: { code: 'operator-cancelled' },
  }, owner);
  const adapter = new FakeAdapter();
  const summary = await drainOutbox({ store, adapter, dispatcherId: 'dispatcher-cancel', now: at });
  assert.equal(adapter.calls.some((call) => call.type === 'run'), false);
  assert.equal(store.get(executionId).state, 'CANCELLED');
  assert.ok(summary.delivered.length >= 1);
});

test('schema-invalid worker results become unknown and are not submitted', async (t) => {
  const { store, executionId } = await setup(t);
  let calls = 0;
  const adapter = { async runAttempt() {
    calls += 1;
    return {
      schema_version: 1,
      attempt_id: 'wrong-attempt',
      fencing_token: 1,
      spec_digest: digest(store.get(executionId).spec),
      kind: 'SUCCEEDED',
      artifacts: [],
      subject_digest: 'a'.repeat(64),
    };
  } };
  const summary = await drainOutbox({ store, adapter, dispatcherId: 'dispatcher-invalid', now: at });
  assert.equal(summary.unknown.length, 1);
  assert.equal(calls, 1);
  assert.equal(store.get(executionId).state, 'EXECUTING');
  assert.equal(store.get(executionId).result, null);
});

test('an expired assignment does not invoke the adapter', async (t) => {
  const { store, executionId, leaseDeadline } = await setup(t, { leaseMs: 1000 });
  let calls = 0;
  const adapter = { async runAttempt() { calls += 1; return null; } };
  const summary = await drainOutbox({
    store,
    adapter,
    dispatcherId: 'dispatcher-expired',
    now: new Date(Date.parse(leaseDeadline) + 1).toISOString(),
  });
  assert.equal(calls, 0);
  assert.ok(summary.unknown.length + summary.delivered.length + summary.skipped.length >= 1);
  assert.equal(store.get(executionId).state, 'ASSIGNED');
});

test('effect command IDs are deterministic from the durable effect identity', async (t) => {
  const { store, executionId } = await setup(t);
  const adapter = new FakeAdapter();
  await drainOutbox({ store, adapter, dispatcherId: 'dispatcher-idempotent', now: at });
  const events = store.events(executionId);
  const workerEvents = events.filter((event) => event.kind === 'ACKNOWLEDGE' || event.kind === 'SUBMIT');
  assert.equal(workerEvents.length, 2);
  assert.equal(workerEvents[0].command_id.startsWith('runner-'), true);
  assert.notEqual(workerEvents[0].command_id, workerEvents[1].command_id);
});

test('reviewer receives the exact immutable artifact bytes accepted from the builder', async (t) => {
  const { store, spec, executionId } = await setup(t, {
    stages: [
      { id: 'build', role: 'builder-developer', required_checks: [check.check_id], requires_approval: false },
      { id: 'review', role: 'reviewer-verifier', required_checks: [check.check_id], requires_approval: false },
    ],
  });
  const builder = new FakeAdapter({ output: 'builder artifact bytes' });
  await drainOutbox({ store, adapter: builder, dispatcherId: 'dispatcher-review', now: at });
  await drainOutbox({ store, checks: [check], dispatcherId: 'dispatcher-review', now: at });
  let execution = store.get(executionId);
  const subjectDigest = execution.result.subject_digest;
  store.execute({
    schema_version: 1,
    command_id: 'accept-builder-stage',
    execution_id: executionId,
    expected_revision: execution.revision,
    at: '2026-10-30T00:00:03.000Z',
    kind: 'ACCEPT_STAGE',
    payload: { subject_digest: subjectDigest },
  }, owner);
  await drainOutbox({ store, adapter: builder, dispatcherId: 'dispatcher-review', now: '2026-10-30T00:00:04.000Z' });

  execution = store.get(executionId);
  assert.equal(execution.accepted_subject_digest, subjectDigest);
  store.execute({
    schema_version: 1,
    command_id: 'assign-reviewer',
    execution_id: executionId,
    expected_revision: execution.revision,
    at: '2026-10-30T00:00:05.000Z',
    kind: 'ASSIGN',
    payload: {
      attempt_id: 'reviewer-attempt',
      worker_id: 'reviewer-1',
      input_revisions: [subjectDigest],
      workspace: 'workspace-fixture',
      capabilities: ['read', 'analyze'],
      lease_deadline: '2026-10-30T00:01:05.000Z',
    },
  }, owner);
  let suppliedSubject;
  const reviewer = createTextModelAdapter({
    workerId: 'reviewer-1',
    invoke: async (request) => {
      suppliedSubject = request.inputSubjects[0];
      return 'reviewer report';
    },
  });
  await drainOutbox({ store, adapter: reviewer, dispatcherId: 'dispatcher-review', now: '2026-10-30T00:00:06.000Z' });
  assert.equal(suppliedSubject.subject_digest, subjectDigest);
  assert.equal(suppliedSubject.artifacts[0].content, 'builder artifact bytes');
  assert.equal(store.get(executionId).state, 'VALIDATING');
  assert.equal(store.get(executionId).spec.task_id, spec.task_id);
});

test('wrong or fresh adapter cannot confirm revocation of another attempt', async (t) => {
  const { store, executionId } = await setup(t);
  const textAdapter = createTextModelAdapter({ workerId, invoke: async () => 'completed output' });
  await drainOutbox({ store, adapter: textAdapter, dispatcherId: 'dispatcher-revoke', now: at });
  let execution = store.get(executionId);
  store.execute({
    schema_version: 1,
    command_id: 'cancel-after-run',
    execution_id: executionId,
    expected_revision: execution.revision,
    at: '2026-10-30T00:00:03.000Z',
    kind: 'CANCEL',
    payload: { code: 'operator-cancelled' },
  }, owner);
  const wrongAdapter = createTextModelAdapter({ workerId: 'another-worker', invoke: async () => 'unused' });
  const summary = await drainOutbox({
    store,
    adapter: wrongAdapter,
    checks: [check],
    dispatcherId: 'dispatcher-revoke-wrong',
    now: '2026-10-30T00:00:04.000Z',
  });
  execution = store.get(executionId);
  assert.equal(summary.unknown.some((item) => item.receipt?.status === 'REVOCATION_ADAPTER_IDENTITY_MISMATCH'), true);
  assert.equal(execution.cleanup_pending, true);
  assert.equal(execution.resources_held, true);
});

test('old unknown effects do not starve later cleanup when the dispatch limit is one', async (t) => {
  const { store, executionId } = await setup(t);
  const failed = await drainOutbox({
    store,
    adapter: { async runAttempt() { throw new Error('outcome unavailable'); } },
    dispatcherId: 'dispatcher-starvation',
    now: at,
  });
  assert.equal(failed.unknown.length, 1);
  let execution = store.get(executionId);
  store.execute({
    schema_version: 1,
    command_id: 'cancel-after-unknown',
    execution_id: executionId,
    expected_revision: execution.revision,
    at: '2026-10-30T00:00:03.000Z',
    kind: 'CANCEL',
    payload: { code: 'owner-cancelled' },
  }, owner);
  const cleanup = await drainOutbox({
    store,
    adapter: new FakeAdapter(),
    dispatcherId: 'dispatcher-starvation-cleanup',
    now: '2026-10-30T00:00:04.000Z',
    limit: 1,
  });
  assert.equal(cleanup.unknown.length, 1);
  assert.equal(cleanup.delivered.length, 1);
  execution = store.get(executionId);
  assert.equal(execution.cleanup_pending, false);
  assert.equal(execution.resources_held, false);
});
