import { validate, canonical, digest } from './contract.mjs';
import {
  validateSource, frontmatter, repositories, claimSections,
} from '../../skills/project-memory/scripts/lib/state.mjs';

function requireInput(condition, code) {
  if (!condition) throw Object.assign(new Error(code), { code });
}

/** Review a pinned source as a proposal; this never acquires or transfers a claim. */
export function importReviewedTask({ source, spec, reviewedBy, expectedRepository, expectedRevision }) {
  canonical(source);
  canonical(spec);
  validateSource(source);
  validate('TaskSpec', spec);
  requireInput(source.repository === expectedRepository && source.revision === expectedRevision,
    'SOURCE_PROVENANCE_MISMATCH');
  requireInput(reviewedBy === spec.accountable_owner, 'REVIEWER_MISMATCH');
  requireInput(source.id === spec.task_id, 'TASK_ID_MISMATCH');
  requireInput(repositories(source.record).includes(spec.repository.repository_id), 'REPOSITORY_MISMATCH');
  const meta = frontmatter(source.record);
  requireInput(!['done', 'closed', 'cancelled'].includes(meta.status), 'SOURCE_NOT_ACTIONABLE');
  const claims = claimSections(source.record);
  requireInput(claims.size === 1, 'SELECT_ONE_SOURCE_CLAIM');
  const [claim] = claims.values();
  return {
    schema_version: 1,
    mode: 'PROPOSAL',
    spec: structuredClone(spec),
    spec_digest: digest(spec),
    reviewed_by: reviewedBy,
    source: {
      repository: source.repository, revision: source.revision,
      path: source.path, blob: source.blob, claim,
    },
  };
}

/** Render an inert view. State changes go through commands, never Markdown edits. */
export function exportTaskView(execution) {
  canonical(execution);
  validate('TaskExecution', execution);
  requireInput(digest(execution.spec) === execution.spec_digest, 'SPEC_DIGEST_MISMATCH');
  // JSON literals keep embedded newlines and Markdown syntax out of view fields.
  const text = (value) => JSON.stringify(value);
  return [
    '# Generated execution view',
    '',
    'Read-only projection. This file grants no ownership or permission.',
    'Submit commands to the coordinator; editing this view changes no runtime state.',
    '',
    `Task: ${text(execution.spec.task_id)}`,
    `Execution: ${text(execution.execution_id)}`,
    `Revision: ${execution.revision}`,
    `Specification: ${execution.spec_digest}`,
    `State: ${execution.state}`,
    `Stage: ${text(execution.spec.workflow.stages[execution.stage_index]?.id ?? null)}`,
    `Accountable owner: ${text(execution.spec.accountable_owner)}`,
    `Worker: ${text(execution.attempt?.worker_id ?? null)}`,
    `Blocker: ${text(execution.blocker)}`,
    `Cleanup pending: ${execution.cleanup_pending}`,
    '',
  ].join('\n');
}
