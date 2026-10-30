import { realpathSync } from 'node:fs';
import { resolve } from 'node:path';
import { canonical, parseJSON, validate } from './contract.mjs';
import { importReviewedTask } from './import-export.mjs';
import { createGit } from '../../skills/project-memory/scripts/lib/git.mjs';
import { canonicalPath, frontmatter, makeSource, TASK } from '../../skills/project-memory/scripts/lib/state.mjs';

function requireInput(condition, code) {
  if (!condition) throw Object.assign(new Error(code), { code });
}

function observedRepository(repo) {
  try {
    return /^(?:https:\/\/github\.com\/|git@github\.com:)([A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+?)(?:\.git)?\/?$/.exec(repo.git('remote', 'get-url', 'origin').trim())?.[1];
  } catch { return undefined; }
}

/** Verify actual local Git objects without transports, replacement refs or lazy fetch. */
export function verifyLocalPin({ root, source, expectedValue }) {
  canonical(source);
  validate('AuthoritySourcePin', source);
  requireInput(typeof root === 'string' && realpathSync(root) === resolve(root), 'UNSAFE_SOURCE_ROOT');
  requireInput(canonicalPath(source.path), 'INVALID_SOURCE_PATH');
  const repo = createGit(root);
  const observed = observedRepository(repo);
  requireInput(!observed || observed === source.repository, 'SOURCE_REPOSITORY_MISMATCH');
  requireInput(repo.commit(source.revision) === source.revision, 'SOURCE_REVISION_MISMATCH');
  const entry = repo.tree(source.revision).get(source.path);
  requireInput(entry && entry.type === 'blob' && ['100644', '100755'].includes(entry.mode) && entry.blob === source.blob, 'SOURCE_BLOB_MISMATCH');
  const text = repo.readBlob(entry.blob);
  requireInput(Buffer.byteLength(text, 'utf8') <= 1_048_576, 'SOURCE_TOO_LARGE');
  if (expectedValue !== undefined) requireInput(canonical(parseJSON(text)) === canonical(expectedValue), 'SOURCE_CONTENT_MISMATCH');
  return { repo, text };
}

/** Authenticate a reviewed Project Memory snapshot against its pinned local tree. */
export function verifyReviewedAdmission({ spec, source, source_root, reviewed_by, namespace }) {
  canonical(source);
  requireInput(TASK.test(spec.task_id) && spec.task_id.startsWith(namespace.task_prefix), 'NAMESPACE_TASK_MISMATCH');
  requireInput(spec.repository.repository_id === namespace.repository && source.repository === namespace.source_repository, 'NAMESPACE_REPOSITORY_MISMATCH');
  requireInput(reviewed_by === namespace.accountable_owner && spec.accountable_owner === namespace.accountable_owner, 'NAMESPACE_OWNER_MISMATCH');
  const pin = { repository: source.repository, revision: source.revision, path: source.path, blob: source.blob };
  const { repo } = verifyLocalPin({ root: source_root, source: pin });
  const sourceAt = (revision) => {
    const records = new Map();
    for (const [path, entry] of repo.tree(revision, 'tasks')) {
      if (!path.endsWith('.md')) continue;
      requireInput(entry.type === 'blob' && ['100644', '100755'].includes(entry.mode), 'INVALID_TASK_SOURCE');
      const text = repo.readBlob(entry.blob);
      requireInput(Buffer.byteLength(text, 'utf8') <= 1_048_576, 'SOURCE_TOO_LARGE');
      records.set(path, { text, blob: entry.blob, mode: entry.mode });
    }
    return makeSource({ repository: source.repository, revision, id: spec.task_id, path: source.path, records });
  };
  const actual = sourceAt(source.revision);
  requireInput(canonical(actual) === canonical(source), 'SOURCE_CONTENT_MISMATCH');
  const live = sourceAt(repo.commit('HEAD'));
  requireInput(live.blob === source.blob && live.record === source.record, 'CURRENT_CLAIM_MISMATCH');
  const meta = frontmatter(live.record);
  requireInput(meta.owner === namespace.accountable_owner, 'NAMESPACE_OWNER_MISMATCH');
  requireInput(meta.status === 'in-progress', 'CURRENT_CLAIM_NOT_ACTIVE');
  return importReviewedTask({ source, spec, reviewedBy: reviewed_by,
    expectedRepository: namespace.source_repository, expectedRevision: source.revision });
}

export function validateEvidenceBinding(binding, execution) {
  canonical(binding);
  validate('EvidenceBinding', binding);
  requireInput(binding.execution_id === execution.execution_id && binding.spec_digest === execution.spec_digest &&
    binding.task_id === execution.spec.task_id && binding.repository === execution.spec.repository.repository_id, 'EVIDENCE_IDENTITY_MISMATCH');
  requireInput(binding.files.every(canonicalPath) && canonicalPath(binding.source_path), 'INVALID_EVIDENCE_BINDING');
  if (execution.spec.integration_policy === 'isolated_write') {
    requireInput(canonical([...binding.files].sort()) === canonical([...execution.spec.write_scope].sort()), 'EVIDENCE_SCOPE_MISMATCH');
  }
  return binding;
}

export function validateNamespaceDecision(decision) {
  canonical(decision);
  validate('NamespaceDecision', decision);
  requireInput(!decision.allow_writes || decision.require_evidence, 'WRITES_REQUIRE_EVIDENCE');
  return decision;
}
