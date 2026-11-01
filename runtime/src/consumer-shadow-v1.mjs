import { parseJSON } from './contract.mjs';
import { importReviewedTask } from './import-export.mjs';
import { evaluate, EVALUATOR_VERSION } from './state.mjs';

export const SHADOW_API_VERSION = 1;
export const SHADOW_LIMITS = Object.freeze({ maxBytes: 1_048_576, maxDepth: 64 });

function requireInput(condition) {
  if (!condition) throw Object.assign(new Error('Invalid shadow request'), { code: 'INVALID_SHADOW_REQUEST' });
}

function fields(value, required, optional = []) {
  requireInput(value !== null && typeof value === 'object' && !Array.isArray(value));
  requireInput(required.every((key) => Object.hasOwn(value, key))
    && Object.keys(value).every((key) => required.includes(key) || optional.includes(key)));
}

function request(text, keys) {
  // The untrusted boundary is JSON text, not an arbitrarily deep object/accessor graph.
  const value = parseJSON(text, SHADOW_LIMITS);
  fields(value, ['version', ...keys]);
  requireInput(value.version === SHADOW_API_VERSION);
  return value;
}

function result(kind, value) {
  return { schema: 'clanker-foundry/shadow-preview', version: SHADOW_API_VERSION,
    mode: 'shadow-only', authority: 'none', kind, value };
}

/** Inspect a supplied genuine source/specification; this does not admit a task. */
export function previewReviewedTask(text) {
  const input = request(text, ['source', 'spec', 'reviewedBy', 'expectedRepository', 'expectedRevision']);
  const { version, ...proposal } = input;
  return result('reviewed-task', importReviewedTask(proposal));
}

function context(value) {
  const flags = ['references_valid', 'dependencies_satisfied', 'resources_available',
    'worker_eligible', 'namespace_managed', 'can_write', 'isolation_verified'];
  fields(value, ['actor_id', 'role'], [...flags, 'approver_ids', 'trusted_checks']);
  const identity = (item) => typeof item === 'string' && item.length > 0 && item.length <= 128;
  requireInput(identity(value.actor_id) && identity(value.role));
  requireInput(flags.every((key) => !Object.hasOwn(value, key) || typeof value[key] === 'boolean'));
  if (Object.hasOwn(value, 'approver_ids')) {
    requireInput(Array.isArray(value.approver_ids) && value.approver_ids.length <= 256
      && value.approver_ids.every(identity));
  }
  if (Object.hasOwn(value, 'trusted_checks')) {
    requireInput(Array.isArray(value.trusted_checks) && value.trusted_checks.length <= 256);
    for (const check of value.trusted_checks) {
      fields(check, ['check_id', 'check_version', 'runner_id']);
      requireInput(Object.values(check).every(identity));
    }
  }
  return value;
}

/** Simulate the shared evaluator. Caller assertions and returned effects are hypothetical. */
export function previewTransition(text) {
  const input = request(text, ['current', 'command', 'context']);
  const evaluation = evaluate(input.current, input.command, context(input.context));
  return { ...result('transition', evaluation), evaluator_version: EVALUATOR_VERSION };
}
