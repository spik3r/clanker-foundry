import { digest } from '../src/contract.mjs';

export const taskSpec = {
  schema_version: 1,
  task_id: 'CF-runtime-example',
  spec_revision: 1,
  outcome: 'Produce a reviewed, read-only pilot result.',
  repository: {
    repository_id: 'spik3r/clanker-foundry',
    revision: '0123456789abcdef0123456789abcdef01234567',
  },
  dependencies: [],
  write_scope: [],
  shared_resources: [],
  acceptance_checks: ['pack-validation'],
  workflow: {
    workflow_id: 'read-only-v1',
    version: 1,
    stages: [
      {
        id: 'review',
        role: 'reviewer-verifier',
        required_checks: ['pack-validation'],
        requires_approval: true,
      },
    ],
  },
  retry_policy: {
    execution_retries: 1,
    rework_retries: 1,
    retry_delay_ms: 1000,
  },
  lease_ms: 60_000,
  validation_ms: 60_000,
  accountable_owner: 'operator',
  integration_policy: 'read_only',
};

export const specDigest = digest(taskSpec);

export const taskExecution = {
  schema_version: 1,
  execution_id: 'exec-1',
  spec: taskSpec,
  spec_digest: specDigest,
  state: 'PENDING',
  stage_index: 0,
  revision: 0,
  attempt: null,
  fence: 0,
  retry_count: 0,
  rework_count: 0,
  result: null,
  validations: [],
  approvals: [],
  cleanup_pending: false,
  resources_held: false,
  validation_deadline: null,
  accepted_subject_digest: null,
  worker_history: [],
  retry_at: null,
  blocker: null,
  outcome: null,
  updated_at: '2026-10-30T00:00:00.000Z',
};

export const command = {
  schema_version: 1,
  command_id: 'cmd-1',
  execution_id: 'exec-1',
  expected_revision: 0,
  at: '2026-10-30T00:00:00.000Z',
  kind: 'ADMIT',
  payload: { spec: taskSpec },
};

export const successfulResult = {
  schema_version: 1,
  attempt_id: 'attempt-1',
  fencing_token: 1,
  spec_digest: specDigest,
  kind: 'SUCCEEDED',
  artifacts: [
    {
      path: 'result.md',
      digest: 'a'.repeat(64),
      repository_revision: '0123456789abcdef0123456789abcdef01234567',
    },
  ],
  subject_digest: digest([
    {
      path: 'result.md',
      digest: 'a'.repeat(64),
      repository_revision: '0123456789abcdef0123456789abcdef01234567',
    },
  ]),
};

export const taskEvent = {
  schema_version: 1,
  evaluator_version: 1,
  event_id: 'cmd-1',
  command_id: 'cmd-1',
  execution_id: 'exec-1',
  sequence: 1,
  prior_revision: 0,
  new_revision: 1,
  actor_id: 'operator',
  at: '2026-10-30T00:00:00.000Z',
  kind: 'ADMIT',
  from_state: null,
  to_state: 'PENDING',
  reason: null,
};

export const effect = {
  schema_version: 1,
  effect_id: 'cmd-1:0',
  execution_id: 'exec-1',
  kind: 'RUN_CHECKS',
  payload: {
    attempt_id: 'attempt-1',
    fencing_token: 1,
    subject_digest: 'b'.repeat(64),
    check_ids: ['pack-validation'],
  },
};
