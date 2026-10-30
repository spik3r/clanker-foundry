import { digest, validate } from './contract.mjs';

const EFFECT_LEASE_MS = 30_000;

function fail(code, message = code) {
  throw Object.assign(new Error(message), { code });
}

function timestamp(value) {
  const raw = typeof value === 'function' ? value() : value ?? new Date().toISOString();
  const parsed = Date.parse(raw);
  if (!Number.isFinite(parsed)) fail('INVALID_TIME', 'Runner clock must return an ISO timestamp');
  return new Date(parsed).toISOString();
}

function commandId(effectId, step) {
  return `runner-${digest({ effect_id: effectId, step }).slice(0, 48)}`;
}

function commandFor(effect, step, execution, at, kind, payload) {
  return {
    schema_version: 1,
    command_id: commandId(effect.effect_id, step),
    execution_id: effect.execution_id,
    expected_revision: execution.revision,
    at,
    kind,
    payload,
  };
}

function isCurrentAttempt(execution, attemptId, fencingToken, states) {
  return execution !== null
    && states.includes(execution.state)
    && execution.attempt?.attempt_id === attemptId
    && execution.attempt?.fencing_token === fencingToken
    && execution.fence === fencingToken
    && !execution.cleanup_pending;
}

function receiptError(error) {
  return {
    status: 'UNKNOWN',
    code: typeof error?.code === 'string' && error.code.length <= 64 ? error.code : 'DISPATCH_UNKNOWN',
  };
}

function workerContext(attempt) {
  return { actor_id: attempt.worker_id, role: 'worker' };
}

function resolveInputSubjects(store, execution, attempt) {
  const requested = [...new Set([
    ...attempt.input_revisions,
    ...(execution.accepted_subject_digest ? [execution.accepted_subject_digest] : []),
  ])];
  const subjects = [];
  const dependencyOutputs = store.dependencyOutputs?.(execution.execution_id) ?? [];
  const dependencyDigests = dependencyOutputs.map((output) => output.subject_digest).filter(Boolean);
  for (const subjectDigest of dependencyDigests) requested.push(subjectDigest);
  const uniqueDigests = [...new Set(requested.filter((item) => /^[a-f0-9]{64}$/.test(item)))];
  if (uniqueDigests.length > 64) fail('INPUT_TOO_LARGE', 'Attempt references too many immutable subjects');
  let totalBytes = 0;
  for (const subjectDigest of uniqueDigests) {
    const manifest = store.getSubject?.(subjectDigest);
    if (!manifest) {
      if (subjectDigest === execution.accepted_subject_digest) {
        fail('SUBJECT_NOT_FOUND', 'Accepted review subject is missing from durable artifact storage');
      }
      continue;
    }
    const artifacts = manifest.map((artifact) => {
      const content = store.getArtifact?.(artifact.digest);
      if (typeof content !== 'string') fail('ARTIFACT_NOT_FOUND', 'Input subject artifact is missing from durable storage');
      totalBytes += Buffer.byteLength(content, 'utf8');
      if (totalBytes > 1024 * 1024) fail('INPUT_TOO_LARGE', 'Immutable input subjects exceed the 1MiB attempt limit');
      return { ...artifact, content };
    });
    const dependency = dependencyOutputs.find((output) => output.subject_digest === subjectDigest);
    subjects.push({
      ...(dependency ? { task_id: dependency.task_id, execution_id: dependency.execution_id } : {}),
      subject_digest: subjectDigest,
      artifacts,
    });
  }
  return subjects;
}

async function submitAttempt({ store, adapter, effect, claim, now }) {
  const { attempt } = effect.payload;
  let execution = store.get(effect.execution_id);
  if (!isCurrentAttempt(execution, attempt.attempt_id, attempt.fencing_token, ['ASSIGNED', 'EXECUTING'])) {
    return { status: 'DELIVERED', receipt: { status: 'SUPPRESSED_STALE' } };
  }
  if (adapter?.workerId && adapter.workerId !== attempt.worker_id) {
    fail('WORKER_IDENTITY_MISMATCH', 'Adapter identity does not match the assigned worker');
  }
  if (!adapter || typeof adapter.runAttempt !== 'function') {
    return { status: 'UNKNOWN', receipt: { status: 'NO_SUPERVISED_ADAPTER' } };
  }
  const ackAt = timestamp(now);
  if (Date.parse(ackAt) >= Date.parse(execution.attempt.lease_deadline)) {
    return { status: 'UNKNOWN', receipt: { status: 'ATTEMPT_EXPIRED' } };
  }
  const ack = commandFor(effect, 'acknowledge', execution, ackAt, 'ACKNOWLEDGE', {
    attempt_id: attempt.attempt_id,
    fencing_token: attempt.fencing_token,
  });
  if (execution.state === 'ASSIGNED') store.execute(ack, workerContext(execution.attempt));
  else if (claim.resume_after_ack !== true) {
    return { status: 'UNKNOWN', receipt: { status: 'RECOVERY_NOT_CONFIRMED' } };
  }

  execution = store.get(effect.execution_id);
  if (!isCurrentAttempt(execution, attempt.attempt_id, attempt.fencing_token, ['EXECUTING'])) {
    return { status: 'UNKNOWN', receipt: { status: 'ATTEMPT_CHANGED_AFTER_ACK' } };
  }
  const result = await adapter.runAttempt({
    attempt: structuredClone(execution.attempt),
    spec: structuredClone(execution.spec),
    inputSubjects: resolveInputSubjects(store, execution, execution.attempt),
  });
  validate('WorkerResult', result);
  if (result.attempt_id !== attempt.attempt_id
    || result.fencing_token !== attempt.fencing_token
    || result.spec_digest !== execution.spec_digest) {
    fail('WORKER_RESULT_MISMATCH', 'Worker result does not match the claimed attempt');
  }
  if (result.kind === 'SUCCEEDED') {
    if (typeof adapter.getArtifact !== 'function' || typeof store.putArtifact !== 'function') {
      fail('ARTIFACT_STORE_REQUIRED', 'Worker artifacts must be stored before submission');
    }
    for (const artifact of result.artifacts) {
      const content = await adapter.getArtifact(artifact.digest);
      if (typeof content !== 'string' || Buffer.byteLength(content, 'utf8') > 1024 * 1024) {
        fail('INVALID_ARTIFACT_CONTENT', 'Adapter did not return a bounded artifact body');
      }
      store.putArtifact({ digest: artifact.digest, content });
      if (typeof adapter.releaseArtifact === 'function') adapter.releaseArtifact(artifact.digest);
    }
  }

  execution = store.get(effect.execution_id);
  if (!isCurrentAttempt(execution, attempt.attempt_id, attempt.fencing_token, ['EXECUTING'])) {
    return { status: 'UNKNOWN', receipt: { status: 'ATTEMPT_CHANGED_BEFORE_SUBMIT' } };
  }
  const submit = commandFor(effect, 'submit', execution, timestamp(now), 'SUBMIT', { result });
  store.execute(submit, workerContext(execution.attempt));
  return {
    status: 'DELIVERED',
    receipt: { status: 'SUBMITTED', command_id: submit.command_id },
  };
}

function configuredCheck(checks, checkId) {
  const matches = checks.filter((item) => item?.check_id === checkId);
  if (matches.length !== 1) fail('CHECK_CONFIGURATION', `Expected exactly one configured check for ${checkId}`);
  const [check] = matches;
  if (typeof check.run !== 'function' || typeof check.check_version !== 'string' || typeof check.runner_id !== 'string') {
    fail('CHECK_CONFIGURATION', `Configured check ${checkId} is incomplete`);
  }
  return check;
}

async function runChecks({ store, effect, checks, now }) {
  const execution = store.get(effect.execution_id);
  if (!isCurrentAttempt(execution, effect.payload.attempt_id, effect.payload.fencing_token, ['VALIDATING'])
    || execution.result?.subject_digest !== effect.payload.subject_digest) {
    return { status: 'DELIVERED', receipt: { status: 'SUPPRESSED_STALE' } };
  }

  for (const checkId of effect.payload.check_ids) {
    const current = store.get(effect.execution_id);
    if (!isCurrentAttempt(current, effect.payload.attempt_id, effect.payload.fencing_token, ['VALIDATING'])
      || current.result?.subject_digest !== effect.payload.subject_digest) {
      return { status: 'UNKNOWN', receipt: { status: 'ATTEMPT_CHANGED_DURING_CHECKS' } };
    }
    if (current.validations.some((item) => item.check_id === checkId)) continue;
    const check = configuredCheck(checks, checkId);
    const output = await check.run(Object.freeze({
      execution: structuredClone(current),
      spec: structuredClone(current.spec),
      result: structuredClone(current.result),
      subject_digest: effect.payload.subject_digest,
      artifacts: current.result.artifacts?.map((artifact) => {
        const content = store.getArtifact?.(artifact.digest);
        if (typeof content !== 'string') fail('ARTIFACT_NOT_FOUND', 'Validation artifact is not available in the durable store');
        return { ...artifact, content };
      }) ?? [],
    }));
    const verdict = typeof output === 'string' ? output : output?.verdict;
    const evidenceRef = typeof output === 'object' && typeof output.evidence_ref === 'string'
      ? output.evidence_ref
      : `check-${digest({ check_id: check.check_id, check_version: check.check_version })}`;
    if (!['PASS', 'FAIL', 'INCONCLUSIVE'].includes(verdict)) fail('INVALID_CHECK_RESULT', `Check ${checkId} returned an invalid verdict`);
    if (evidenceRef.length < 1 || evidenceRef.length > 128) fail('INVALID_CHECK_RESULT', `Check ${checkId} returned an invalid evidence reference`);

    const at = timestamp(now);
    const evidence = {
      check_id: check.check_id,
      check_version: check.check_version,
      runner_id: check.runner_id,
      subject_digest: current.result.subject_digest,
      repository_revision: current.spec.repository.revision,
      attempt_id: current.attempt.attempt_id,
      spec_digest: current.spec_digest,
      workflow_id: current.spec.workflow.workflow_id,
      workflow_version: current.spec.workflow.version,
      verdict,
      at,
      evidence_ref: evidenceRef,
    };
    validate('ValidationEvidence', evidence);
    const command = commandFor(effect, `validation:${checkId}`, current, at, 'RECORD_VALIDATION', { evidence });
    store.execute(command, { actor_id: check.runner_id, role: 'validator' });
  }
  return { status: 'DELIVERED', receipt: { status: 'CHECKS_RECORDED', check_ids: effect.payload.check_ids } };
}

async function revokeCapabilities({ store, adapter, effect, now }) {
  const execution = store.get(effect.execution_id);
  const { attempt_id: attemptId, fencing_token: fencingToken } = effect.payload;
  if (!execution?.cleanup_pending || execution.attempt?.attempt_id !== attemptId) {
    return { status: 'DELIVERED', receipt: { status: 'SUPPRESSED_STALE' } };
  }
  if (!adapter || typeof adapter.revokeCapabilities !== 'function') {
    return { status: 'UNKNOWN', receipt: { status: 'NO_REVOCATION_ADAPTER' } };
  }
  if (adapter.workerId && adapter.workerId !== execution.attempt.worker_id) {
    return { status: 'UNKNOWN', receipt: { status: 'REVOCATION_ADAPTER_IDENTITY_MISMATCH' } };
  }
  const confirmation = await adapter.revokeCapabilities({
    attempt_id: attemptId,
    fencing_token: fencingToken,
    attempt: structuredClone(execution.attempt),
  });
  if (confirmation?.stopped !== true) {
    return { status: 'UNKNOWN', receipt: { status: 'STOP_NOT_CONFIRMED' } };
  }
  const current = store.get(effect.execution_id);
  if (!current?.cleanup_pending || current.attempt?.attempt_id !== attemptId) {
    return { status: 'UNKNOWN', receipt: { status: 'ATTEMPT_CHANGED_DURING_REVOCATION' } };
  }
  const command = commandFor(effect, 'cleanup-confirmed', current, timestamp(now), 'CLEANUP_CONFIRMED', {
    attempt_id: attemptId,
    fencing_token: fencingToken,
  });
  store.execute(command, { actor_id: current.spec.accountable_owner, role: 'owner' });
  return { status: 'DELIVERED', receipt: { status: 'CLEANUP_CONFIRMED', command_id: command.command_id } };
}

/** Drain a bounded number of durable effects. Unknown outcomes are never relaunched. */
export async function drainOutbox({
  store,
  adapter,
  checks = [],
  dispatcherId,
  now,
  limit = 32,
} = {}) {
  if (!store || typeof store.pendingEffects !== 'function' || typeof store.claimEffect !== 'function'
    || typeof store.acknowledgeEffect !== 'function' || typeof store.get !== 'function' || typeof store.execute !== 'function') {
    fail('INVALID_STORE', 'drainOutbox requires a runtime store');
  }
  if (typeof dispatcherId !== 'string' || dispatcherId.length < 1 || dispatcherId.length > 128) fail('INVALID_ARGUMENT', 'dispatcherId is required and bounded');
  if (!Number.isSafeInteger(limit) || limit < 1 || limit > 256) fail('INVALID_ARGUMENT', 'limit must be between 1 and 256');
  if (!Array.isArray(checks)) fail('INVALID_ARGUMENT', 'checks must be an array');

  const summary = { delivered: [], unknown: [], skipped: [] };
  const allEffects = store.pendingEffects();
  const candidates = allEffects.filter((item) => item.status === 'PENDING').slice(0, limit);
  for (const item of allEffects) {
    if (item.status === 'UNKNOWN') {
      summary.unknown.push({ effect_id: item.effect.effect_id, status: 'UNKNOWN', receipt: item.receipt });
    }
  }
  for (const pending of allEffects.filter((item) => item.status === 'CLAIMED')) {
    try {
      const claim = store.claimEffect({
        effectId: pending.effect.effect_id,
        dispatcherId,
        now: timestamp(now),
        leaseMs: EFFECT_LEASE_MS,
      });
      if (claim.status === 'UNKNOWN') {
        summary.unknown.push({ effect_id: claim.effect.effect_id, status: 'UNKNOWN', receipt: claim.receipt });
      } else {
        summary.skipped.push({ effect_id: pending.effect.effect_id, status: 'CLAIMED' });
      }
    } catch (error) {
      summary.skipped.push({ effect_id: pending.effect.effect_id, code: error.code ?? 'CLAIM_FAILED' });
    }
  }
  for (const pending of candidates) {
    if (!['LAUNCH_ATTEMPT', 'RUN_CHECKS', 'REVOKE_CAPABILITIES'].includes(pending.effect.kind)) {
      summary.skipped.push({ effect_id: pending.effect.effect_id, status: 'UNSUPPORTED_EFFECT' });
      continue;
    }
    const at = timestamp(now);
    let claim;
    try {
      claim = store.claimEffect({ effectId: pending.effect.effect_id, dispatcherId, now: at, leaseMs: EFFECT_LEASE_MS });
    } catch (error) {
      summary.skipped.push({ effect_id: pending.effect.effect_id, code: error.code ?? 'CLAIM_FAILED' });
      continue;
    }
    if (claim.status !== 'CLAIMED') {
      const item = { effect_id: claim.effect.effect_id, status: claim.status, receipt: claim.receipt };
      (claim.status === 'UNKNOWN' ? summary.unknown : summary.skipped).push(item);
      continue;
    }
    let outcome = 'UNKNOWN';
    let receipt;
    try {
      const effect = claim.effect;
      validate('Effect', effect);
      let result;
      if (effect.kind === 'LAUNCH_ATTEMPT') result = await submitAttempt({ store, adapter, effect, claim, now });
      else if (effect.kind === 'RUN_CHECKS') result = await runChecks({ store, effect, checks, now });
      else if (effect.kind === 'REVOKE_CAPABILITIES') result = await revokeCapabilities({ store, adapter, effect, now });
      else fail('UNSUPPORTED_EFFECT', `Unsupported effect kind: ${effect.kind}`);
      outcome = result.status;
      receipt = result.receipt;
    } catch (error) {
      const failed = receiptError(error);
      outcome = failed.status;
      receipt = failed;
    }

    try {
      const acknowledged = store.acknowledgeEffect({
        effectId: claim.effect.effect_id,
        dispatcherId,
        claimToken: claim.claim_token,
        outcome,
        receipt,
      });
      const item = { effect_id: claim.effect.effect_id, status: acknowledged.status, receipt: acknowledged.receipt };
      (acknowledged.status === 'UNKNOWN' ? summary.unknown : summary.delivered).push(item);
    } catch (error) {
      summary.unknown.push({ effect_id: claim.effect.effect_id, status: 'ACKNOWLEDGEMENT_UNKNOWN', code: error.code ?? 'ACK_FAILED' });
    }
  }
  return summary;
}
