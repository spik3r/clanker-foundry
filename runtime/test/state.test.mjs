import test from 'node:test';
import assert from 'node:assert/strict';
import { evaluate, TERMINAL_STATES } from '../src/state.mjs';
import { canonical, digest, validate } from '../src/contract.mjs';
import { taskSpec } from './fixtures.mjs';

const instant = '2026-10-30T00:00:00.000Z';
const later = (ms) => new Date(Date.parse(instant) + ms).toISOString();
const trustedChecks = [{ check_id: 'pack-validation', check_version: '1', runner_id: 'checker' }];
const owner = { actor_id: 'operator', role: 'owner', references_valid: true, dependencies_satisfied: true, resources_available: true, worker_eligible: true, trusted_checks: trustedChecks, approver_ids: ['operator'] };
const worker = (id) => ({ ...owner, actor_id: id, role: 'worker' });
const checker = { ...owner, actor_id: 'checker', role: 'validator' };

function specFor(roles = ['reviewer-verifier']) {
  const spec = structuredClone(taskSpec);
  spec.workflow.stages = roles.map((role, index) => ({ id: `stage-${index}`, role, required_checks: ['pack-validation'], requires_approval: role === 'reviewer-verifier' }));
  return spec;
}

function harness(spec = specFor()) {
  let execution = null;
  let sequence = 0;
  const transcript = [];
  const send = (kind, payload = {}, context = owner, at = instant, overrides = {}) => {
    const command = { schema_version: 1, execution_id: 'exec-1', command_id: `command-${++sequence}`, expected_revision: execution?.revision ?? 0, at, kind, payload, ...overrides };
    const before = canonical(execution);
    try {
      const answer = evaluate(execution, command, context);
      assert.equal(canonical(execution), before, 'evaluation must not mutate input');
      transcript.push({ command, context, answer: structuredClone(answer) });
      execution = answer.execution;
      return answer;
    } catch (error) {
      assert.equal(canonical(execution), before, 'rejection must not mutate input');
      transcript.push({ command, context, error: error.code });
      throw error;
    }
  };
  send('ADMIT', { spec });
  return { send, transcript, get execution() { return execution; } };
}

function assign(h, workerId = 'worker-1', at = instant) {
  const stage = h.execution.spec.workflow.stages[h.execution.stage_index];
  const reviewInput = stage.role === 'reviewer-verifier'
    ? h.execution.accepted_subject_digest ?? h.execution.spec.repository.revision
    : null;
  const inputRevisions = [...new Set([h.execution.accepted_subject_digest, reviewInput].filter(Boolean))];
  return h.send('ASSIGN', { attempt_id: `attempt-${h.execution.worker_history.length + 1}`, worker_id: workerId, input_revisions: inputRevisions, workspace: 'attempts/read-only', capabilities: ['read', 'analyze'], lease_deadline: new Date(Date.parse(at) + 60_000).toISOString() }, owner, at);
}

function acknowledge(h, at = instant) {
  const attempt = h.execution.attempt;
  return h.send('ACKNOWLEDGE', { attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token }, worker(attempt.worker_id), at);
}

function succeeded(h, at = instant) {
  const attempt = h.execution.attempt;
  const artifacts = [{ path: 'result.md', digest: 'a'.repeat(64), repository_revision: h.execution.spec.repository.revision }];
  return h.send('SUBMIT', { result: { schema_version: 1, attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, spec_digest: h.execution.spec_digest, kind: 'SUCCEEDED', artifacts, subject_digest: digest(artifacts) } }, worker(attempt.worker_id), at);
}

function evidence(h, overrides = {}, at = instant) {
  return { check_id: 'pack-validation', check_version: '1', runner_id: 'checker', subject_digest: h.execution.result.subject_digest, repository_revision: h.execution.spec.repository.revision, attempt_id: h.execution.attempt.attempt_id, spec_digest: h.execution.spec_digest, workflow_id: h.execution.spec.workflow.workflow_id, workflow_version: 1, verdict: 'PASS', at, evidence_ref: 'evidence/check-1', ...overrides };
}

function approval(h, overrides = {}) {
  return { approval_id: `approval-${h.execution.revision}`, actor_id: 'operator', action: 'ACCEPT_STAGE', execution_id: h.execution.execution_id, spec_digest: h.execution.spec_digest, workflow_id: h.execution.spec.workflow.workflow_id, workflow_version: 1, subject_digest: h.execution.result.subject_digest, decision: 'APPROVE', expires_at: later(30_000), revoked: false, ...overrides };
}

function cleanup(h, at = instant) {
  const attempt = h.execution.attempt;
  return h.send('CLEANUP_CONFIRMED', { attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token }, owner, at);
}

function accept(h, at = instant, context = owner) {
  return h.send('ACCEPT_STAGE', { subject_digest: h.execution.result?.subject_digest ?? '0'.repeat(64) }, context, at);
}

function ready(h, workerId = 'worker-1', at = instant) {
  assign(h, workerId, at); acknowledge(h, at); succeeded(h, at);
  h.send('RECORD_VALIDATION', { evidence: evidence(h, {}, at) }, checker, at);
}

function rejectsCode(action, code) {
  assert.throws(action, (error) => error.code === code, `expected ${code}`);
}

test('maximum-length accepted command IDs produce bounded stable effect IDs', () => {
  const h = harness(specFor(['explorer-researcher']));
  const payload = { attempt_id: 'long-id-attempt', worker_id: 'worker-1',
    input_revisions: [], workspace: 'read-only', capabilities: ['read'],
    lease_deadline: later(60_000) };
  const input = { schema_version: 1, command_id: 'a'.repeat(128), execution_id: 'exec-1',
    expected_revision: h.execution.revision, at: instant, kind: 'ASSIGN', payload };
  const first = evaluate(h.execution, input, owner);
  const replayed = evaluate(h.execution, input, owner);
  assert.deepEqual(first, replayed);
  assert.equal(first.effects.length, 1);
  assert.ok(first.effects[0].effect_id.length <= 128);
  validate('Effect', first.effects[0]);
});

test('admission bounds the effective check union before a worker can start', () => {
  const spec = specFor(['explorer-researcher']);
  spec.workflow.stages[0].required_checks = Array.from({ length: 256 }, (_, index) => `check-${index}`);
  spec.acceptance_checks = ['extra-acceptance-check'];
  rejectsCode(() => harness(spec), 'CHECK_LIMIT_EXCEEDED');
  spec.acceptance_checks = ['check-0'];
  const h = harness(spec);
  assign(h); acknowledge(h);
  const result = succeeded(h);
  assert.equal(result.effects[0].payload.check_ids.length, 256);
  validate('Effect', result.effects[0]);
});

test('ordered stages require independent validation and approval, then confirmed cleanup', () => {
  const h = harness(specFor(['explorer-researcher', 'architect-planner', 'builder-developer', 'reviewer-verifier']));
  for (let index = 0; index < 4; index += 1) {
    ready(h, `worker-${index}`);
    if (index === 3) {
      rejectsCode(() => accept(h), 'APPROVAL_REQUIRED');
      h.send('RECORD_APPROVAL', { approval: approval(h) });
    }
    const accepted = accept(h);
    assert.equal(accepted.effects[0].kind, 'REVOKE_CAPABILITIES');
    assert.equal(h.execution.cleanup_pending, true);
    assert.equal(h.execution.resources_held, true);
    if (index < 3) rejectsCode(() => assign(h), 'CLEANUP_REQUIRED');
    cleanup(h);
    assert.equal(h.execution.resources_held, false);
  }
  assert.equal(h.execution.state, 'COMPLETED');
  assert.equal(h.execution.stage_index, 3);
  rejectsCode(() => h.send('CANCEL', { code: 'late-cancel' }), 'TERMINAL_EXECUTION');
  for (const entry of h.transcript.filter((item) => item.answer)) {
    validate('TaskExecution', entry.answer.execution);
    validate('TaskEvent', entry.answer.event);
  }
});

test('admission is typed, immutable and restricted to supported read-only workflows', () => {
  const command = { schema_version: 1, command_id: 'admit', execution_id: 'exec-1', expected_revision: 0, at: instant, kind: 'ADMIT', payload: { spec: specFor() } };
  rejectsCode(() => evaluate(null, { ...command, actor_id: 'operator' }, owner), 'INVALID_CONTRACT');
  rejectsCode(() => evaluate(null, command, { ...owner, actor_id: 'stranger' }), 'UNAUTHORIZED');
  rejectsCode(() => evaluate(null, command, { ...owner, references_valid: false }), 'INVALID_REFERENCES');
  for (const modify of [
    (spec) => { spec.write_scope = ['runtime/**']; },
    (spec) => { spec.workflow.workflow_id = 'arbitrary'; },
    (spec) => { spec.workflow.stages.push(structuredClone(spec.workflow.stages[0])); },
  ]) {
    const changed = structuredClone(command); modify(changed.payload.spec);
    assert.throws(() => evaluate(null, changed, owner));
  }
  const unreviewedBuild = { ...command, payload: { spec: specFor(['builder-developer']) } };
  rejectsCode(() => evaluate(null, unreviewedBuild, owner), 'REVIEW_REQUIRED');
});

test('readiness, trusted identity, revisions and fencing reject stale workers', () => {
  const h = harness();
  rejectsCode(() => h.send('ASSIGN', { attempt_id: 'attempt-1', worker_id: 'worker-1', input_revisions: [], workspace: 'attempts/one', capabilities: ['read'], lease_deadline: later(60_000) }, { ...owner, dependencies_satisfied: false }), 'DEPENDENCIES_NOT_SATISFIED');
  assign(h);
  rejectsCode(() => h.send('ACKNOWLEDGE', { attempt_id: 'attempt-1', fencing_token: 0 }, worker('worker-1')), 'INVALID_CONTRACT');
  rejectsCode(() => h.send('ACKNOWLEDGE', { attempt_id: 'attempt-1', fencing_token: 2 }, worker('worker-1')), 'STALE_ATTEMPT');
  rejectsCode(() => h.send('ACKNOWLEDGE', { attempt_id: 'attempt-1', fencing_token: 1 }, worker('imposter')), 'UNAUTHORIZED');
  acknowledge(h);
  rejectsCode(() => h.send('CANCEL', { code: 'cancel' }, owner, instant, { expected_revision: 1 }), 'REVISION_CONFLICT');
  h.send('CANCEL', { code: 'cancel' });
  assert.equal(h.execution.fence, 2);
  rejectsCode(() => succeeded(h), 'TERMINAL_EXECUTION');
  cleanup(h);
});

test('candidate digest, revision and independent registered evidence bind exact output', () => {
  const h = harness(); assign(h); acknowledge(h);
  const result = { schema_version: 1, attempt_id: 'attempt-1', fencing_token: 1, spec_digest: h.execution.spec_digest, kind: 'SUCCEEDED', artifacts: [{ path: 'result.md', digest: 'a'.repeat(64), repository_revision: h.execution.spec.repository.revision }], subject_digest: 'b'.repeat(64) };
  rejectsCode(() => h.send('SUBMIT', { result }, worker('worker-1')), 'ARTIFACT_DIGEST_MISMATCH');
  succeeded(h);
  rejectsCode(() => h.send('RECORD_VALIDATION', { evidence: evidence(h) }, { ...checker, trusted_checks: [] }), 'UNTRUSTED_CHECK');
  rejectsCode(() => h.send('RECORD_VALIDATION', { evidence: evidence(h, { subject_digest: 'd'.repeat(64) }) }, checker), 'SUBJECT_MISMATCH');
  rejectsCode(() => h.send('RECORD_VALIDATION', { evidence: evidence(h, { attempt_id: 'old' }) }, checker), 'SUBJECT_MISMATCH');
  rejectsCode(() => h.send('RECORD_VALIDATION', { evidence: evidence(h, { repository_revision: 'f'.repeat(40) }) }, checker), 'SUBJECT_MISMATCH');
  rejectsCode(() => accept(h), 'CHECKS_NOT_PASSED');
  h.send('RECORD_VALIDATION', { evidence: evidence(h) }, checker);
  rejectsCode(() => h.send('RECORD_VALIDATION', { evidence: evidence(h) }, checker), 'DUPLICATE_EVIDENCE');
  rejectsCode(() => h.send('RECORD_APPROVAL', { approval: approval(h, { subject_digest: 'd'.repeat(64) }) }), 'SUBJECT_MISMATCH');
  const grant = approval(h);
  h.send('RECORD_APPROVAL', { approval: grant });
  h.send('REVOKE_APPROVAL', { approval_id: grant.approval_id, reason: 'owner-revoked' });
  rejectsCode(() => accept(h), 'APPROVAL_REQUIRED');
  h.send('RECORD_APPROVAL', { approval: approval(h) });
  rejectsCode(() => accept(h, later(30_000)), 'APPROVAL_REQUIRED');
});

test('timeouts use recorded deadlines, finite budgets and cleanup before retry', () => {
  const h = harness(); assign(h);
  rejectsCode(() => h.send('TIMEOUT', { retry_at: later(60_999) }, owner, later(59_999)), 'TIMEOUT_NOT_DUE');
  h.send('TIMEOUT', { retry_at: later(61_000) }, owner, later(60_000));
  assert.equal(h.execution.state, 'RETRY_WAIT');
  rejectsCode(() => h.send('RETRY_TIMER', {}, owner, later(61_000)), 'NOT_READY');
  cleanup(h, later(61_000));
  h.send('RETRY_TIMER', {}, owner, later(61_000));
  assign(h, 'worker-2', later(61_000));
  h.send('TIMEOUT', { retry_at: later(122_000) }, owner, later(121_000));
  assert.equal(h.execution.state, 'FAILED');
  assert.equal(h.execution.outcome.code, 'RETRIES_EXHAUSTED');
  assert.equal(h.execution.resources_held, true);
  cleanup(h, later(121_000));
});

test('validation deadline does not slide with evidence, and worker lifecycle commands fail', () => {
  const h = harness(); assign(h); acknowledge(h); succeeded(h);
  h.send('RECORD_VALIDATION', { evidence: evidence(h, {}, later(10_000)) }, checker, later(10_000));
  rejectsCode(() => h.send('CANCEL', { code: 'worker-cancel' }, worker('worker-1'), later(10_000)), 'UNAUTHORIZED');
  rejectsCode(() => accept(h, later(60_000)), 'ATTEMPT_EXPIRED');
  h.send('TIMEOUT', { retry_at: later(61_000) }, owner, later(60_000));
  assert.equal(h.execution.state, 'RETRY_WAIT');
});

test('review cannot reuse builder and rework exhaustion blocks owner decisions', () => {
  const spec = specFor(['builder-developer', 'reviewer-verifier']);
  const h = harness(spec);
  ready(h, 'builder'); accept(h); cleanup(h);
  rejectsCode(() => assign(h, 'builder'), 'REVIEW_NOT_INDEPENDENT');
  for (let round = 0; round < 2; round += 1) {
    assign(h, 'reviewer'); acknowledge(h);
    const result = { schema_version: 1, attempt_id: h.execution.attempt.attempt_id, fencing_token: h.execution.attempt.fencing_token, spec_digest: h.execution.spec_digest, kind: 'CHANGES_REQUIRED', subject_digest: h.execution.accepted_subject_digest, findings: [{ code: 'review', message: 'Revise the candidate.' }] };
    h.send('SUBMIT', { result }, worker('reviewer'));
    h.send('REQUEST_CHANGES', { reason: 'review-findings' });
    cleanup(h);
    if (round === 0) { assert.equal(h.execution.state, 'PENDING'); ready(h, 'builder'); accept(h); cleanup(h); }
  }
  assert.equal(h.execution.state, 'BLOCKED');
  assert.equal(h.execution.blocker.code, 'REWORK_EXHAUSTED');
  h.send('RESOLVE');
  assert.equal(h.execution.state, 'PENDING');
  assert.equal(h.execution.attempt, null);
});

test('blocking fences active work and resolution never restores old attempts', () => {
  const h = harness(); assign(h); acknowledge(h);
  h.send('BLOCK', { code: 'owner-input', resolver: 'operator' });
  assert.equal(h.execution.cleanup_pending, true);
  rejectsCode(() => h.send('RESOLVE'), 'NOT_READY');
  cleanup(h); h.send('RESOLVE');
  rejectsCode(() => h.send('SUBMIT', { result: { schema_version: 1, attempt_id: 'attempt-1', fencing_token: 1, spec_digest: h.execution.spec_digest, kind: 'FAILED', failure_code: 'late' } }, worker('worker-1')), 'ILLEGAL_TRANSITION');
  assert.equal(h.execution.state, 'PENDING');
});

test('review assignment binds an earlier accepted subject and acceptance names its own result', () => {
  const h = harness(specFor(['builder-developer', 'reviewer-verifier']));
  ready(h, 'builder');
  const candidate = h.execution.result.subject_digest;
  accept(h); cleanup(h);

  const bad = {
    attempt_id: 'attempt-review', worker_id: 'reviewer', input_revisions: [h.execution.spec.repository.revision],
    workspace: 'attempts/review', capabilities: ['read'], lease_deadline: later(60_000),
  };
  rejectsCode(() => h.send('ASSIGN', bad), 'SUBJECT_MISMATCH');
  assign(h, 'reviewer'); acknowledge(h);
  const attempt = h.execution.attempt;
  const report = [{ path: 'review.md', digest: 'c'.repeat(64), repository_revision: h.execution.spec.repository.revision }];
  h.send('SUBMIT', { result: { schema_version: 1, attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, spec_digest: h.execution.spec_digest, kind: 'SUCCEEDED', artifacts: report, subject_digest: digest(report) } }, worker('reviewer'));
  h.send('RECORD_VALIDATION', { evidence: evidence(h) }, checker);
  h.send('RECORD_APPROVAL', { approval: approval(h) });
  rejectsCode(() => h.send('ACCEPT_STAGE', { subject_digest: candidate }), 'SUBJECT_MISMATCH');
  accept(h); cleanup(h);
  assert.equal(h.execution.accepted_subject_digest, candidate, 'review report must not replace the reviewed candidate digest');
});

test('task acceptance checks bind to the candidate, not a later review report', () => {
  const spec = specFor(['builder-developer', 'reviewer-verifier']);
  spec.acceptance_checks = ['task-acceptance'];
  spec.workflow.stages[0].required_checks = ['builder-check'];
  spec.workflow.stages[1].required_checks = ['review-check'];
  const h = harness(spec);
  assign(h, 'builder'); acknowledge(h); succeeded(h);
  const candidateDigest = h.execution.result.subject_digest;
  assert.deepEqual(h.transcript.at(-1).answer.effects[0].payload.check_ids, ['builder-check', 'task-acceptance']);

  const checkContext = (checkId) => ({ ...checker, trusted_checks: [{ check_id: checkId, check_version: '1', runner_id: 'checker' }] });
  h.send('RECORD_VALIDATION', { evidence: evidence(h, { check_id: 'builder-check' }) }, checkContext('builder-check'));
  h.send('RECORD_VALIDATION', { evidence: evidence(h, { check_id: 'task-acceptance', verdict: 'FAIL' }) }, checkContext('task-acceptance'));
  rejectsCode(() => accept(h), 'CHECKS_NOT_PASSED');
  rejectsCode(() => h.send('RECORD_VALIDATION', { evidence: evidence(h, { check_id: 'task-acceptance', verdict: 'PASS', subject_digest: digest([{ path: 'review.md', digest: 'c'.repeat(64), repository_revision: h.execution.spec.repository.revision }]) }) }, checkContext('task-acceptance')), 'SUBJECT_MISMATCH');
  assert.equal(h.execution.state, 'VALIDATING');
  assert.equal(h.execution.accepted_subject_digest, null);
  assert.notEqual(candidateDigest, digest([{ path: 'review.md', digest: 'c'.repeat(64), repository_revision: h.execution.spec.repository.revision }]));

  const accepted = harness(spec);
  assign(accepted, 'builder'); acknowledge(accepted); succeeded(accepted);
  const acceptedCandidate = accepted.execution.result.subject_digest;
  accepted.send('RECORD_VALIDATION', { evidence: evidence(accepted, { check_id: 'builder-check' }) }, checkContext('builder-check'));
  accepted.send('RECORD_VALIDATION', { evidence: evidence(accepted, { check_id: 'task-acceptance' }) }, checkContext('task-acceptance'));
  accept(accepted); cleanup(accepted);
  assign(accepted, 'reviewer'); acknowledge(accepted); succeeded(accepted);
  assert.deepEqual(accepted.transcript.at(-1).answer.effects[0].payload.check_ids, ['review-check']);
  assert.equal(accepted.execution.accepted_subject_digest, acceptedCandidate);
});

test('worker BLOCKED and FAILED results require an owner transition and cleanup', () => {
  for (const kind of ['BLOCKED', 'FAILED']) {
    const h = harness(specFor(['builder-developer', 'reviewer-verifier']));
    assign(h); acknowledge(h);
    const attempt = h.execution.attempt;
    const result = kind === 'BLOCKED'
      ? { schema_version: 1, attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, spec_digest: h.execution.spec_digest, kind, blocker: { code: 'MISSING_INPUT', resolver: 'operator' } }
      : { schema_version: 1, attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token, spec_digest: h.execution.spec_digest, kind, failure_code: 'WORKER_ERROR' };
    h.send('SUBMIT', { result }, worker(attempt.worker_id));
    assert.equal(h.execution.state, 'VALIDATING');
    rejectsCode(() => accept(h), 'RESULT_NOT_SUCCESS');
    if (kind === 'BLOCKED') h.send('BLOCK', result.blocker);
    else h.send('FAIL', { code: result.failure_code });
    assert.equal(h.execution.state, kind === 'BLOCKED' ? 'BLOCKED' : 'FAILED');
    assert.equal(h.execution.cleanup_pending, true);
    cleanup(h);
    assert.equal(h.execution.resources_held, false);
  }
});

test('invalid calendar times, stale fences and early cleanup confirmations are rejected', () => {
  const h = harness(specFor(['builder-developer', 'reviewer-verifier']));
  const malformed = { schema_version: 1, command_id: 'bad-time', execution_id: 'exec-1', expected_revision: 1, at: '2026-13-30T00:00:00.000Z', kind: 'ASSIGN', payload: { attempt_id: 'a', worker_id: 'worker-1', input_revisions: [], workspace: 'attempts/a', capabilities: ['read'], lease_deadline: later(60_000) } };
  rejectsCode(() => evaluate(h.execution, malformed, owner), 'INVALID_TIME');
  assign(h); acknowledge(h);
  const attempt = h.execution.attempt;
  rejectsCode(() => h.send('CLEANUP_CONFIRMED', { attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token }, owner), 'NO_CLEANUP_PENDING');
  h.send('RETRYABLE_ERROR', { code: 'WORKER_RETRY', retry_at: later(1_000) });
  cleanup(h);
  h.send('RETRY_TIMER', {}, owner, later(1_000));
  assign(h, 'worker-2', later(1_000));
  rejectsCode(() => h.send('ACKNOWLEDGE', { attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token }, worker('worker-1'), later(1_000)), 'STALE_ATTEMPT');
});

test('generated accepted/rejected command sequences replay identical states, events and effects', () => {
  for (let seed = 1; seed <= 24; seed += 1) {
    const h = harness();
    let random = seed;
    const next = () => (random = (Math.imul(random, 1664525) + 1013904223) >>> 0);
    for (let step = 0; step < 18; step += 1) {
      const execution = h.execution;
      if (TERMINAL_STATES.includes(execution.state)) {
        if (execution.cleanup_pending) cleanup(h);
        break;
      }
      if (next() % 5 === 0) rejectsCode(() => accept(h, instant, worker('worker-1')), 'UNAUTHORIZED');
      if (execution.cleanup_pending) { cleanup(h); continue; }
      if (execution.state === 'PENDING') assign(h);
      else if (execution.state === 'ASSIGNED') acknowledge(h);
      else if (execution.state === 'EXECUTING') {
        if (next() % 3 === 0) h.send('CANCEL', { code: 'generated-cancel' });
        else succeeded(h);
      } else if (execution.state === 'VALIDATING') {
        if (execution.validations.length === 0) h.send('RECORD_VALIDATION', { evidence: evidence(h) }, checker);
        else if (execution.approvals.length === 0) h.send('RECORD_APPROVAL', { approval: approval(h) });
        else accept(h);
      }
      assert.equal(h.execution.resources_held, h.execution.attempt !== null);
      assert.ok(h.execution.fence >= (h.execution.attempt?.fencing_token ?? 0));
    }
    let replay = null;
    for (const entry of h.transcript) {
      if (entry.error) rejectsCode(() => evaluate(replay, entry.command, entry.context), entry.error);
      else {
        const evaluated = evaluate(replay, entry.command, entry.context);
        assert.deepEqual(evaluated, entry.answer);
        replay = evaluated.execution;
      }
    }
    assert.deepEqual(replay, h.execution);
  }
});

test('writable workflow requires managed authority and per-attempt isolation', () => {
  const spec = specFor(['builder-developer', 'reviewer-verifier']);
  spec.integration_policy = 'isolated_write';
  spec.workflow.workflow_id = 'isolated-write-v1';
  spec.write_scope = ['src/example.txt'];
  const admit = { schema_version: 1, execution_id: 'write-1', command_id: 'admit-write', expected_revision: 0, at: instant, kind: 'ADMIT', payload: { spec } };
  assert.throws(() => evaluate(null, admit, owner), { code: 'MANAGED_AUTHORITY_REQUIRED' });
  const managed = { ...owner, namespace_managed: true, can_write: true };
  const pending = evaluate(null, admit, managed).execution;
  const assignment = { ...admit, command_id: 'assign-write', expected_revision: pending.revision, kind: 'ASSIGN', payload: {
    attempt_id: 'write-attempt', worker_id: 'builder', workspace: 'private-candidate', input_revisions: [spec.repository.revision], capabilities: ['read', 'write'], lease_deadline: later(60_000),
  } };
  assert.throws(() => evaluate(pending, assignment, managed), { code: 'ISOLATION_REQUIRED' });
  assert.equal(evaluate(pending, assignment, { ...managed, isolation_verified: true }).execution.state, 'ASSIGNED');
  const overlap = structuredClone(admit); overlap.payload.spec.write_scope = ['new', 'new/child.txt'];
  assert.throws(() => evaluate(null, overlap, managed), { code: 'INVALID_WRITE_SCOPE' });
  for (const path of ['src/*', '.git/config', 'src/../secret', '.project-memory/event.json', './file']) {
    const bad = structuredClone(admit); bad.payload.spec.write_scope = [path];
    assert.throws(() => evaluate(null, bad, managed));
  }
});

test('maximum stage, execution retry and rework budgets fit retained attempt history', () => {
  const spec = specFor([...Array(15).fill('builder-developer'), 'reviewer-verifier']);
  spec.acceptance_checks = [];
  spec.workflow.stages.forEach((stage) => { stage.required_checks = []; stage.requires_approval = false; });
  spec.retry_policy = { execution_retries: 20, rework_retries: 20, retry_delay_ms: 0 };
  const h = harness(spec);
  for (let retry = 0; retry < 20; retry += 1) {
    assign(h, 'builder'); acknowledge(h);
    h.send('RETRYABLE_ERROR', { code: 'temporary', retry_at: instant });
    cleanup(h); h.send('RETRY_TIMER');
  }
  for (let round = 0; round <= 20; round += 1) {
    for (let stage = 0; stage < 16; stage += 1) {
      const reviewer = stage === 15;
      assign(h, reviewer ? 'reviewer' : 'builder'); acknowledge(h);
      if (reviewer && round < 20) {
        h.send('SUBMIT', { result: { schema_version: 1, attempt_id: h.execution.attempt.attempt_id,
          fencing_token: h.execution.fence, spec_digest: h.execution.spec_digest, kind: 'CHANGES_REQUIRED',
          subject_digest: h.execution.accepted_subject_digest, findings: [{ code: 'revision', message: 'Revise candidate' }] } }, worker('reviewer'));
        h.send('REQUEST_CHANGES', { reason: 'review-findings' });
      } else { succeeded(h); accept(h); }
      cleanup(h);
    }
  }
  assert.equal(h.execution.state, 'COMPLETED');
  assert.equal(h.execution.worker_history.length, 356);
  assert.equal(h.execution.retry_count, 20);
  assert.equal(h.execution.rework_count, 20);
  assert.equal(h.execution.resources_held, false);
  let replay = null;
  for (const entry of h.transcript) {
    const answer = evaluate(replay, entry.command, entry.context);
    assert.deepEqual(answer, entry.answer);
    replay = answer.execution;
  }
  assert.deepEqual(replay, h.execution);
});

test('extra manual recovery at full history capacity fails explicitly without mutation', () => {
  const h = harness(specFor(['explorer-researcher']));
  const current = structuredClone(h.execution);
  current.worker_history = Array.from({ length: 356 }, (_, index) => ({
    stage_id: 'stage-0', role: 'explorer-researcher', worker_id: 'worker-1', attempt_id: `old-${index}`,
  }));
  const before = canonical(current);
  assert.throws(() => evaluate(current, { schema_version: 1, command_id: 'beyond-capacity',
    execution_id: current.execution_id, expected_revision: current.revision, at: instant, kind: 'ASSIGN', payload: {
      attempt_id: 'new-extra-attempt', worker_id: 'worker-1', input_revisions: [], workspace: 'private',
      capabilities: ['read'], lease_deadline: later(60_000),
    } }, owner), { code: 'ATTEMPT_HISTORY_EXHAUSTED' });
  assert.equal(canonical(current), before);
});
