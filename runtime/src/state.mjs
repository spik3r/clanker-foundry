import { canonical, digest, validate, RUNTIME_LIMITS } from './contract.mjs';

export const EVALUATOR_VERSION = 1;
export const TERMINAL_STATES = Object.freeze(['COMPLETED', 'FAILED', 'CANCELLED']);
const active = new Set(['ASSIGNED', 'EXECUTING', 'VALIDATING']);
const roles = new Set(['owner', 'coordinator', 'scheduler', 'worker', 'validator', 'approver']);

function requireGuard(condition, code, message = code) {
  if (!condition) throw Object.assign(new Error(message), { code });
}

function millis(value) {
  const parsed = Date.parse(value);
  requireGuard(Number.isFinite(parsed) && new Date(parsed).toISOString().slice(0, 19) === value.slice(0, 19), 'INVALID_TIME');
  return parsed;
}

function uniqueBy(items, key) {
  return new Set(items.map((item) => item[key])).size === items.length;
}

function admitSpec(spec, context) {
  const writable = spec.integration_policy === 'isolated_write';
  if (writable) {
    requireGuard(context.namespace_managed === true && context.can_write === true, 'MANAGED_AUTHORITY_REQUIRED');
    requireGuard(spec.write_scope.length > 0 && spec.write_scope.every((path) =>
      !path.includes('*') && !path.includes('//') && !path.endsWith('/')
      && path.split('/').every((part) => part !== '.' && part !== '..' && !['.git', '.codex', '.agents', '.project-memory'].includes(part))), 'INVALID_WRITE_SCOPE');
    requireGuard(!spec.write_scope.some((path, index) => spec.write_scope.some((other, otherIndex) =>
      index !== otherIndex && other.startsWith(`${path}/`))), 'INVALID_WRITE_SCOPE');
  } else requireGuard(spec.write_scope.length === 0, 'UNSUPPORTED_CAPABILITY');
  requireGuard(spec.workflow.version === 1 && spec.workflow.workflow_id === (writable ? 'isolated-write-v1' : 'read-only-v1'), 'UNSUPPORTED_WORKFLOW');
  requireGuard(uniqueBy(spec.workflow.stages, 'id') && uniqueBy(spec.dependencies, 'task_id') && uniqueBy(spec.shared_resources, 'resource_id'), 'DUPLICATE_ID');
  const stages = spec.workflow.stages;
  requireGuard(!stages.some((stage) => stage.role === 'builder-developer') || stages.at(-1).role === 'reviewer-verifier', 'REVIEW_REQUIRED');
  requireGuard(!spec.dependencies.some((dependency) => dependency.task_id === spec.task_id), 'DEPENDENCY_CYCLE');
  const candidate = spec.workflow.stages.findLastIndex((stage) => stage.role !== 'reviewer-verifier');
  const acceptanceStage = candidate < 0 ? spec.workflow.stages.length - 1 : candidate;
  for (const [index, stage] of spec.workflow.stages.entries()) {
    const effective = new Set([...stage.required_checks, ...(index === acceptanceStage ? spec.acceptance_checks : [])]);
    requireGuard(effective.size <= 256, 'CHECK_LIMIT_EXCEEDED');
  }
}

/** Evaluate one authenticated, ordered command without reading external state. */
export function evaluate(current, command, context = {}) {
  // Canonicalization rejects accessors and non-JSON values before validation reads them.
  canonical(command);
  validate('Command', command);
  requireGuard(roles.has(context.role) && typeof context.actor_id === 'string', 'UNAUTHORIZED');
  const now = millis(command.at);
  const { kind, payload } = command;
  const effects = [];
  let execution;
  let reason = null;
  if (current === null) {
    requireGuard(kind === 'ADMIT', 'EXECUTION_NOT_FOUND');
    requireGuard(command.expected_revision === 0, 'REVISION_CONFLICT');
    const spec = payload.spec;
    requireGuard((context.role === 'owner' && context.actor_id === spec.accountable_owner) || context.role === 'coordinator', 'UNAUTHORIZED');
    admitSpec(spec, context);
    requireGuard(context.references_valid === true, 'INVALID_REFERENCES');
    execution = {
      schema_version: 1, execution_id: command.execution_id, spec: structuredClone(spec), spec_digest: digest(spec),
      state: 'PENDING', stage_index: 0, revision: 0, attempt: null, fence: 0,
      retry_count: 0, rework_count: 0, result: null, validations: [], approvals: [],
      cleanup_pending: false, resources_held: false, retry_at: null, blocker: null,
      outcome: null, updated_at: command.at, worker_history: [], validation_deadline: null,
      accepted_subject_digest: null,
    };
  } else {
    canonical(current);
    validate('TaskExecution', current);
    requireGuard(digest(current.spec) === current.spec_digest, 'SPEC_DIGEST_MISMATCH');
    requireGuard(command.execution_id === current.execution_id, 'EXECUTION_MISMATCH');
    requireGuard(command.expected_revision === current.revision, 'REVISION_CONFLICT');
    requireGuard(now >= millis(current.updated_at), 'TIME_REGRESSION');
    requireGuard(!TERMINAL_STATES.includes(current.state) || kind === 'CLEANUP_CONFIRMED', 'TERMINAL_EXECUTION');
    execution = structuredClone(current);
    const owner = (context.role === 'owner' && context.actor_id === execution.spec.accountable_owner) || context.role === 'coordinator';
    const scheduled = owner || context.role === 'scheduler';
    const stage = execution.spec.workflow.stages[execution.stage_index];
    requireGuard(stage !== undefined, 'INVALID_STAGE');
    const state = (...states) => requireGuard(states.includes(execution.state), 'ILLEGAL_TRANSITION');
    const authorized = (allowed = owner) => requireGuard(allowed, 'UNAUTHORIZED');
    const effect = (effectKind, effectPayload) => effects.push({
      schema_version: 1, effect_id: `effect-${digest({ command_id: command.command_id, index: effects.length })}`,
      execution_id: execution.execution_id, kind: effectKind, payload: structuredClone(effectPayload),
    });
    const matchingAttempt = (input, live = true) => {
      requireGuard(execution.attempt !== null && input.attempt_id === execution.attempt.attempt_id && input.fencing_token === execution.attempt.fencing_token, 'STALE_ATTEMPT');
      if (live) requireGuard(!execution.cleanup_pending && input.fencing_token === execution.fence, 'STALE_ATTEMPT');
    };
    const unexpired = () => {
      const deadline = execution.state === 'VALIDATING' ? execution.validation_deadline : execution.attempt.lease_deadline;
      requireGuard(now < millis(deadline), 'ATTEMPT_EXPIRED');
    };
    const retire = () => {
      execution.validation_deadline = null;
      if (execution.attempt && !execution.cleanup_pending) {
        execution.fence += 1;
        execution.cleanup_pending = true;
        effect('REVOKE_CAPABILITIES', { attempt_id: execution.attempt.attempt_id, fencing_token: execution.attempt.fencing_token });
      }
    };
    const clearSubject = () => { execution.result = null; execution.validations = []; execution.approvals = []; };
    const candidateStageIndex = execution.spec.workflow.stages.findLastIndex((item) => item.role !== 'reviewer-verifier');
    const acceptanceStageIndex = candidateStageIndex < 0 ? execution.spec.workflow.stages.length - 1 : candidateStageIndex;
    const checks = () => [...new Set([
      ...stage.required_checks,
      ...(execution.stage_index === acceptanceStageIndex ? execution.spec.acceptance_checks : []),
    ])];
    const sameSubject = (record) => {
      requireGuard(record.subject_digest === execution.result?.subject_digest && record.spec_digest === execution.spec_digest && record.workflow_id === execution.spec.workflow.workflow_id && record.workflow_version === execution.spec.workflow.version, 'SUBJECT_MISMATCH');
    };
    switch (kind) {
      case 'ASSIGN': {
        authorized(scheduled); state('PENDING');
        requireGuard(execution.worker_history.length < RUNTIME_LIMITS.workerHistory, 'ATTEMPT_HISTORY_EXHAUSTED');
        requireGuard(!execution.cleanup_pending && !execution.resources_held && execution.attempt === null, 'CLEANUP_REQUIRED');
        requireGuard(context.dependencies_satisfied === true, 'DEPENDENCIES_NOT_SATISFIED');
        requireGuard(context.resources_available === true, 'RESOURCE_CONFLICT');
        requireGuard(context.worker_eligible === true, 'WORKER_INELIGIBLE');
        if (execution.spec.integration_policy === 'isolated_write') {
          requireGuard(context.namespace_managed === true && context.can_write === true && context.isolation_verified === true, 'ISOLATION_REQUIRED');
          requireGuard(payload.capabilities.every((capability) => ['read', 'analyze', 'write'].includes(capability)), 'UNSUPPORTED_CAPABILITY');
          requireGuard(stage.role === 'builder-developer' || !payload.capabilities.includes('write'), 'UNSUPPORTED_CAPABILITY');
        } else requireGuard(payload.capabilities.every((capability) => ['read', 'analyze'].includes(capability)), 'UNSUPPORTED_CAPABILITY');
        requireGuard(!execution.worker_history.some((item) => item.attempt_id === payload.attempt_id), 'DUPLICATE_ATTEMPT');
        requireGuard(millis(payload.lease_deadline) > now && millis(payload.lease_deadline) <= now + execution.spec.lease_ms, 'INVALID_LEASE');
        if (stage.role === 'reviewer-verifier') {
          const reviewSubject = execution.accepted_subject_digest ?? execution.spec.repository.revision;
          requireGuard(!execution.worker_history.some((item) => item.role === 'builder-developer' && item.worker_id === payload.worker_id), 'REVIEW_NOT_INDEPENDENT');
          requireGuard(payload.input_revisions.includes(reviewSubject), 'SUBJECT_MISMATCH');
        }
        execution.fence += 1;
        execution.attempt = { ...structuredClone(payload), stage_id: stage.id, attempt_number: execution.worker_history.length + 1, fencing_token: execution.fence };
        execution.worker_history.push({ stage_id: stage.id, role: stage.role, worker_id: payload.worker_id, attempt_id: payload.attempt_id });
        execution.resources_held = true;
        execution.state = 'ASSIGNED';
        effect('LAUNCH_ATTEMPT', { attempt: execution.attempt });
        break;
      }
      case 'ACKNOWLEDGE':
        state('ASSIGNED'); matchingAttempt(payload); unexpired();
        authorized(context.role === 'worker' && context.actor_id === execution.attempt.worker_id);
        execution.state = 'EXECUTING';
        break;
      case 'SUBMIT': {
        state('EXECUTING'); matchingAttempt(payload.result); unexpired();
        authorized(context.role === 'worker' && context.actor_id === execution.attempt.worker_id);
        const result = payload.result;
        requireGuard(result.spec_digest === execution.spec_digest, 'SPEC_DIGEST_MISMATCH');
        if (result.kind === 'SUCCEEDED') {
          requireGuard(result.artifacts.length > 0 && uniqueBy(result.artifacts, 'path'), 'INVALID_ARTIFACTS');
          requireGuard(result.artifacts.every((artifact) => artifact.repository_revision === execution.spec.repository.revision), 'SUBJECT_MISMATCH');
          requireGuard(result.subject_digest === digest(result.artifacts), 'ARTIFACT_DIGEST_MISMATCH');
        } else if (result.kind === 'CHANGES_REQUIRED') {
          requireGuard(stage.role === 'reviewer-verifier' && result.findings.length > 0 && result.subject_digest === execution.accepted_subject_digest, 'SUBJECT_MISMATCH');
        }
        execution.result = structuredClone(result);
        execution.validations = []; execution.approvals = [];
        execution.state = 'VALIDATING';
        execution.validation_deadline = new Date(now + execution.spec.validation_ms).toISOString();
        if (result.kind === 'SUCCEEDED' || result.kind === 'CHANGES_REQUIRED') effect('RUN_CHECKS', {
          attempt_id: execution.attempt.attempt_id, fencing_token: execution.attempt.fencing_token,
          subject_digest: result.subject_digest, check_ids: checks(),
        });
        break;
      }
      case 'RECORD_VALIDATION': {
        state('VALIDATING'); unexpired();
        const evidence = payload.evidence;
        authorized(context.role === 'validator' && context.actor_id === evidence.runner_id);
        requireGuard(evidence.runner_id !== execution.attempt.worker_id, 'REVIEW_NOT_INDEPENDENT');
        requireGuard((context.trusted_checks ?? []).some((check) => check.check_id === evidence.check_id && check.check_version === evidence.check_version && check.runner_id === evidence.runner_id), 'UNTRUSTED_CHECK');
        sameSubject(evidence);
        requireGuard(evidence.attempt_id === execution.attempt.attempt_id && evidence.repository_revision === execution.spec.repository.revision, 'SUBJECT_MISMATCH');
        requireGuard(checks().includes(evidence.check_id), 'UNEXPECTED_CHECK');
        requireGuard(millis(evidence.at) <= now && millis(evidence.at) >= millis(execution.updated_at), 'INVALID_EVIDENCE_TIME');
        requireGuard(!execution.validations.some((item) => item.check_id === evidence.check_id), 'DUPLICATE_EVIDENCE');
        execution.validations.push(structuredClone(evidence));
        break;
      }
      case 'RECORD_APPROVAL': {
        state('VALIDATING'); unexpired();
        const approval = payload.approval;
        authorized((owner || context.role === 'approver') && context.actor_id === approval.actor_id && (context.approver_ids ?? []).includes(context.actor_id));
        sameSubject(approval);
        requireGuard(approval.execution_id === execution.execution_id && millis(approval.expires_at) > now && approval.revoked === false, 'INVALID_APPROVAL');
        requireGuard(!execution.approvals.some((item) => item.approval_id === approval.approval_id), 'DUPLICATE_APPROVAL');
        execution.approvals.push(structuredClone(approval));
        break;
      }
      case 'REVOKE_APPROVAL': {
        state('VALIDATING');
        const approval = execution.approvals.find((item) => item.approval_id === payload.approval_id);
        requireGuard(approval !== undefined && !approval.revoked, 'APPROVAL_NOT_FOUND');
        authorized(owner || (context.role === 'approver' && context.actor_id === approval.actor_id));
        approval.revoked = true;
        break;
      }
      case 'ACCEPT_STAGE': {
        authorized(); state('VALIDATING'); unexpired();
        requireGuard(execution.result.kind === 'SUCCEEDED', 'RESULT_NOT_SUCCESS');
        requireGuard(payload.subject_digest === execution.result.subject_digest, 'SUBJECT_MISMATCH');
        requireGuard(checks().every((check) => execution.validations.some((item) => item.check_id === check && item.verdict === 'PASS')), 'CHECKS_NOT_PASSED');
        if (stage.requires_approval) requireGuard(execution.approvals.some((item) => item.decision === 'APPROVE' && !item.revoked && millis(item.expires_at) > now && (context.approver_ids ?? []).includes(item.actor_id)), 'APPROVAL_REQUIRED');
        if (stage.role !== 'reviewer-verifier' || execution.accepted_subject_digest === null) {
          execution.accepted_subject_digest = execution.result.subject_digest;
        }
        retire();
        if (execution.stage_index === execution.spec.workflow.stages.length - 1) {
          execution.state = 'COMPLETED'; execution.outcome = { code: 'ACCEPTED' };
        } else {
          execution.stage_index += 1; execution.state = 'PENDING'; clearSubject();
        }
        break;
      }
      case 'REQUEST_CHANGES': {
        authorized(); state('VALIDATING'); unexpired();
        requireGuard(execution.result.kind === 'CHANGES_REQUIRED' || execution.validations.some((item) => item.verdict !== 'PASS'), 'CHANGES_NOT_JUSTIFIED');
        const build = execution.spec.workflow.stages.findIndex((item) => item.role === 'builder-developer');
        retire(); reason = payload.reason;
        if (execution.rework_count >= execution.spec.retry_policy.rework_retries || build < 0) {
          execution.state = 'BLOCKED'; execution.blocker = { code: 'REWORK_EXHAUSTED', resolver: execution.spec.accountable_owner };
        } else {
          execution.rework_count += 1; execution.stage_index = build; execution.state = 'PENDING'; clearSubject();
        }
        break;
      }
      case 'RETRYABLE_ERROR':
      case 'TIMEOUT': {
        authorized(scheduled); requireGuard(active.has(execution.state), 'ILLEGAL_TRANSITION');
        if (kind === 'TIMEOUT') requireGuard(now >= millis(execution.state === 'VALIDATING' ? execution.validation_deadline : execution.attempt.lease_deadline), 'TIMEOUT_NOT_DUE');
        requireGuard(millis(payload.retry_at) === now + execution.spec.retry_policy.retry_delay_ms, 'INVALID_RETRY_TIME');
        retire(); reason = kind === 'TIMEOUT' ? 'TIMEOUT' : payload.code;
        if (execution.retry_count >= execution.spec.retry_policy.execution_retries) {
          execution.state = 'FAILED'; execution.outcome = { code: 'RETRIES_EXHAUSTED' }; execution.retry_at = null;
        } else {
          execution.retry_count += 1; execution.state = 'RETRY_WAIT'; execution.retry_at = payload.retry_at;
        }
        break;
      }
      case 'RETRY_TIMER':
        authorized(scheduled); state('RETRY_WAIT');
        requireGuard(now >= millis(execution.retry_at), 'RETRY_NOT_DUE');
        requireGuard(!execution.cleanup_pending && context.dependencies_satisfied === true && context.resources_available === true, 'NOT_READY');
        execution.state = 'PENDING'; execution.retry_at = null; clearSubject();
        break;
      case 'BLOCK':
        authorized(); retire(); execution.state = 'BLOCKED'; execution.blocker = structuredClone(payload); reason = payload.code;
        break;
      case 'RESOLVE':
        authorized(); state('BLOCKED');
        requireGuard(!execution.cleanup_pending && context.dependencies_satisfied === true && context.resources_available === true, 'NOT_READY');
        execution.state = 'PENDING'; execution.blocker = null; clearSubject();
        break;
      case 'FAIL':
      case 'CANCEL':
        authorized(); retire(); execution.state = kind === 'FAIL' ? 'FAILED' : 'CANCELLED'; execution.outcome = { code: payload.code }; reason = payload.code;
        break;
      case 'CLEANUP_CONFIRMED':
        authorized(); requireGuard(execution.cleanup_pending, 'NO_CLEANUP_PENDING'); matchingAttempt(payload, false);
        execution.cleanup_pending = false; execution.resources_held = false; execution.attempt = null;
        break;
      default: requireGuard(false, 'ILLEGAL_TRANSITION');
    }
  }
  const priorRevision = current?.revision ?? 0;
  execution.revision = priorRevision + 1;
  execution.updated_at = command.at;
  const event = {
    schema_version: 1, evaluator_version: EVALUATOR_VERSION, event_id: command.command_id,
    command_id: command.command_id, execution_id: execution.execution_id,
    sequence: execution.revision, prior_revision: priorRevision, new_revision: execution.revision,
    actor_id: context.actor_id, at: command.at, kind, from_state: current?.state ?? null,
    to_state: execution.state, reason,
  };
  validate('TaskExecution', execution); validate('TaskEvent', event);
  for (const effect of effects) validate('Effect', effect);
  return { execution, event, effects };
}
