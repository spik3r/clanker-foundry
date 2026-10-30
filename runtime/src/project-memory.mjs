import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { closeSync, constants, lstatSync, openSync, realpathSync } from 'node:fs';
import { dirname, isAbsolute, join, parse, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { DatabaseSync } from 'node:sqlite';
import { canonical, digest, parseJSON, validate } from './contract.mjs';
import { validateWorkspacePlan } from './workspace.mjs';
import { createGit, loadMemory, localMemory, validateCommitted } from '../../skills/project-memory/scripts/lib/git.mjs';
import { blobHash, eventPath, filesList, frontmatter, makeSource, sourcePath } from '../../skills/project-memory/scripts/lib/state.mjs';

const helper = fileURLToPath(new URL('../../skills/project-memory/scripts/project-memory.mjs', import.meta.url));
const reject = (code, message) => { throw Object.assign(new Error(message), { code }); };
const requireThat = (condition, code, message = code) => { if (!condition) reject(code, message); };
const same = (a, b) => canonical(a) === canonical(b);
const ordered = (files) => [...files].sort();
const inside = (root, path) => { const rel = relative(root, path); return rel === '' || (!rel.startsWith('../') && !isAbsolute(rel)); };

function privateJournal(path, roots) {
  requireThat(typeof path === 'string' && isAbsolute(path), 'UNSAFE_JOURNAL_PATH');
  const normalized = resolve(path);
  requireThat(!roots.some((root) => inside(root, normalized)), 'UNSAFE_JOURNAL_PATH', 'Journal must be outside worker and task-source workspaces');
  let cursor = parse(normalized).root;
  for (const part of normalized.slice(cursor.length).split('/').slice(0, -1)) {
    cursor = join(cursor, part);
    requireThat(!lstatSync(cursor).isSymbolicLink(), 'UNSAFE_JOURNAL_PATH');
  }
  const check = (name, directory = false) => {
    const stat = lstatSync(name);
    requireThat(!stat.isSymbolicLink() && (directory ? stat.isDirectory() : stat.isFile())
      && stat.uid === process.getuid() && !(stat.mode & 0o077) && (directory || stat.nlink === 1), 'UNSAFE_JOURNAL_PATH');
    return stat;
  };
  check(dirname(normalized), true);
  try { closeSync(openSync(normalized, constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW, 0o600)); }
  catch (error) { if (error.code !== 'EEXIST') throw error; }
  for (const suffix of ['-wal', '-shm', '-journal']) {
    try { check(`${normalized}${suffix}`); } catch (error) { if (error.code !== 'ENOENT') throw error; }
  }
  const before = check(normalized);
  const db = new DatabaseSync(normalized);
  const after = check(normalized);
  if (before.dev !== after.dev || before.ino !== after.ino) { db.close(); reject('UNSAFE_JOURNAL_PATH', 'Journal changed while opening'); }
  try {
    db.exec('PRAGMA busy_timeout=5000; PRAGMA trusted_schema=OFF;');
    const version = db.prepare('PRAGMA user_version').get().user_version;
    const app = db.prepare('PRAGMA application_id').get().application_id;
    if (version === 0 && app === 0 && !db.prepare("SELECT name FROM sqlite_master WHERE type='table'").get()) {
      db.exec(`BEGIN IMMEDIATE;
        CREATE TABLE operations (execution_id TEXT PRIMARY KEY, run_id TEXT NOT NULL UNIQUE, binding TEXT NOT NULL, intent TEXT, completion TEXT);
        PRAGMA application_id=1128678733; PRAGMA user_version=1; COMMIT;`);
    } else requireThat(version === 1 && app === 1128678733, 'UNSUPPORTED_JOURNAL');
    db.exec('PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL;');
  } catch (error) { db.close(); throw error; }
  return db;
}

function repositoryIdentity(git, expected) {
  let remote;
  try { remote = git.git('remote', 'get-url', 'origin').trim(); } catch { return; }
  const observed = /^(?:https:\/\/github\.com\/|git@github\.com:)([A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+?)(?:\.git)?\/?$/.exec(remote)?.[1];
  requireThat(!observed || observed === expected, 'REPOSITORY_MISMATCH');
}

/** Bind existing genuine events; this bridge never registers sources, starts work or stages files. */
export function createProjectMemoryBridge({ store, codeRoot, sourceRoot, journalPath, invokeHelper } = {}) {
  requireThat(store && typeof store.get === 'function' && typeof store.bindEvidence === 'function', 'INVALID_STORE');
  const code = createGit(codeRoot); const sourceGit = createGit(sourceRoot);
  requireThat(realpathSync(code.root) === resolve(codeRoot) && realpathSync(sourceGit.root) === resolve(sourceRoot), 'INVALID_WORKSPACE_ROOT');
  const db = privateJournal(journalPath, [code.root, sourceGit.root]);
  const invoke = (command, args) => {
    const run = () => execFileSync(process.execPath, [helper, command, ...args], {
      cwd: code.root, encoding: 'utf8', maxBuffer: 8 * 1024 * 1024, stdio: ['ignore', 'pipe', 'pipe'],
    });
    return invokeHelper ? invokeHelper({ command, args: [...args], run }) : run();
  };
  const transaction = (fn) => {
    db.exec('BEGIN IMMEDIATE');
    try { const result = fn(); db.exec('COMMIT'); return result; }
    catch (error) { db.exec('ROLLBACK'); throw error; }
  };
  const operation = (id) => {
    const row = db.prepare('SELECT * FROM operations WHERE execution_id=?').get(id);
    return row && { binding: JSON.parse(row.binding), intent: row.intent && JSON.parse(row.intent), completion: row.completion && JSON.parse(row.completion) };
  };
  const executionFor = (id) => {
    const execution = store.get(id);
    requireThat(execution && digest(execution.spec) === execution.spec_digest, 'EXECUTION_NOT_FOUND');
    return execution;
  };
  const bound = (id) => {
    const binding = store.evidenceBinding(id);
    requireThat(binding, 'EVIDENCE_NOT_BOUND');
    validate('EvidenceBinding', binding);
    return binding;
  };
  const authentic = (binding, execution) => {
    validate('EvidenceBinding', binding);
    filesList(binding.files);
    requireThat(binding.code_root === code.root && binding.source_root === sourceGit.root, 'WORKSPACE_MISMATCH');
    requireThat(binding.execution_id === execution.execution_id && binding.spec_digest === execution.spec_digest
      && binding.task_id === execution.spec.task_id && binding.repository === execution.spec.repository.repository_id
      && same(ordered(binding.files), ordered(execution.spec.write_scope)), 'EVIDENCE_IDENTITY_MISMATCH');
    repositoryIdentity(code, binding.repository); repositoryIdentity(sourceGit, binding.source_repository);
    requireThat(sourceGit.commit(binding.source_revision) === binding.source_revision, 'SOURCE_REVISION_MISMATCH');
    const records = new Map();
    for (const [path, entry] of sourceGit.tree(binding.source_revision, 'tasks')) {
      if (!path.endsWith('.md')) continue;
      requireThat(entry.type === 'blob' && ['100644', '100755'].includes(entry.mode), 'INVALID_SOURCE_RECORD');
      records.set(path, { ...entry, text: sourceGit.readBlob(entry.blob) });
    }
    const actual = makeSource({ repository: binding.source_repository, revision: binding.source_revision,
      id: binding.task_id, path: binding.source_path, records });
    requireThat(actual.blob === binding.source_blob, 'SOURCE_BLOB_MISMATCH');
    const currentRecords = new Map();
    for (const [path, entry] of sourceGit.tree('HEAD', 'tasks')) {
      if (path.endsWith('.md')) {
        requireThat(entry.type === 'blob' && ['100644', '100755'].includes(entry.mode), 'INVALID_SOURCE_RECORD');
        currentRecords.set(path, { ...entry, text: sourceGit.readBlob(entry.blob) });
      }
    }
    makeSource({ repository: binding.source_repository, revision: sourceGit.commit('HEAD'), id: binding.task_id,
      path: binding.source_path, records: currentRecords });
    const current = sourceGit.entry('HEAD', binding.source_path);
    requireThat(current.blob === actual.blob && sourceGit.snapshotIndex([binding.source_path])[binding.source_path].blob === actual.blob,
      'SOURCE_CHANGED', 'Canonical task changed since its reviewed pin');
    requireThat(!['done', 'closed', 'cancelled'].includes(frontmatter(actual.record)?.status), 'SOURCE_NOT_ACTIONABLE');
    const memory = loadMemory(localMemory(code.root));
    requireThat(same(memory.sources.get(sourcePath(actual)), actual), 'REGISTERED_SOURCE_MISMATCH');
    const start = memory.graph.byId.get(binding.start_event_id);
    const head = memory.graph.byId.get(binding.head_event_id);
    requireThat(start?.kind === 'start' && start.runId === binding.run_id && head?.runId === binding.run_id
      && digest(start) === binding.start_event_digest && digest(head) === binding.head_event_digest
      && ['start', 'checkpoint'].includes(head.kind) && head.task.id === binding.task_id
      && head.task.revision === binding.source_revision && head.task.blob === binding.source_blob
      && head.task.repository === binding.source_repository && head.task.path === binding.source_path
      && head.task.claim.heading === binding.claim_heading && head.task.claim.digest === binding.claim_digest
      && head.repository === binding.repository && same(ordered(head.files), ordered(binding.files)), 'START_BINDING_MISMATCH');
    return { memory, head, start };
  };
  const active = (binding, execution, clean = false) => {
    const state = authentic(binding, execution);
    const heads = state.memory.graph.tasks.get(binding.task_id)?.heads;
    requireThat(heads?.length === 1 && heads[0].eventId === binding.head_event_id, 'STALE_EVIDENCE_FRONTIER');
    if (clean) {
      requireThat(code.commit('HEAD') === execution.spec.repository.revision, 'CODE_REVISION_MISMATCH');
      const snapshot = code.snapshotIndex(binding.files);
      requireThat(binding.files.every((file) => same(snapshot[file], code.entry(execution.spec.repository.revision, file))),
        'EVIDENCE_AFTER_EDITS', 'Bind the genuine start before implementation changes');
    }
    return state;
  };
  const accepted = (execution) => {
    requireThat(execution.state === 'COMPLETED' && execution.result?.kind === 'SUCCEEDED'
      && !!execution.accepted_subject_digest, 'RUNTIME_RESULT_NOT_ACCEPTED');
    requireThat(execution.cleanup_pending === false && execution.resources_held === false, 'RUNTIME_CLEANUP_REQUIRED');
  };
  const acceptedSnapshot = (execution, binding) => {
    requireThat(typeof store.getSubject === 'function' && typeof store.getArtifact === 'function', 'ACCEPTED_PLAN_REQUIRED');
    const manifest = store.getSubject(execution.accepted_subject_digest);
    requireThat(Array.isArray(manifest) && manifest.length === 1 && digest(manifest) === execution.accepted_subject_digest
      && manifest[0].path === 'workspace-plan.json' && manifest[0].repository_revision === execution.spec.repository.revision,
    'ACCEPTED_PLAN_REQUIRED');
    const text = store.getArtifact(manifest[0].digest);
    requireThat(typeof text === 'string' && createHash('sha256').update(text).digest('hex') === manifest[0].digest, 'ACCEPTED_PLAN_REQUIRED');
    const plan = parseJSON(text); validateWorkspacePlan(plan);
    requireThat(canonical(plan) === text && plan.execution_id === execution.execution_id && plan.spec_digest === execution.spec_digest
      && plan.repository_revision === execution.spec.repository.revision && same(ordered(plan.write_scope), ordered(binding.files))
      && execution.worker_history.some((worker) => worker.role === 'builder-developer' && worker.attempt_id === plan.attempt_id),
    'ACCEPTED_PLAN_IDENTITY_MISMATCH');
    requireThat(plan.operations.every((operation) => binding.files.includes(operation.path)), 'ACCEPTED_PLAN_SCOPE_MISMATCH');
    const format = code.git('rev-parse', '--show-object-format').trim();
    const snapshot = Object.fromEntries(binding.files.map((file) => [file, code.entry(execution.spec.repository.revision, file)]));
    for (const file of binding.files) requireThat(snapshot[file].mode === null || ['100644', '100755'].includes(snapshot[file].mode), 'ACCEPTED_PLAN_UNSUPPORTED_MODE');
    for (const operation of plan.operations) {
      const before = snapshot[operation.path];
      const beforeDigest = before.blob === null ? null : createHash('sha256').update(code.bytes('cat-file', 'blob', before.blob)).digest('hex');
      requireThat(operation.before_digest === beforeDigest, 'ACCEPTED_PLAN_BASE_MISMATCH');
      snapshot[operation.path] = operation.action === 'delete' ? { blob: null, mode: null }
        : { blob: blobHash(operation.content, format), mode: before.mode || '100644' };
    }
    return snapshot;
  };
  const matchCompletion = (event, intent, binding) => event.kind === 'complete' && event.runId === binding.run_id
    && same(event.parents, [binding.head_event_id]) && same(ordered(event.files), ordered(binding.files))
    && event.summary === intent.summary && event.validation === intent.validation && event.next === intent.next
    && same(event.snapshot, intent.snapshot);
  const recover = (binding, execution, intent) => {
    const { memory } = authentic(binding, execution);
    const children = (memory.graph.children.get(binding.head_event_id) || []).map((id) => memory.graph.byId.get(id));
    if (!children.length) return null;
    requireThat(children.length === 1 && matchCompletion(children[0], intent, binding), 'AMBIGUOUS_COMPLETION');
    const heads = memory.graph.tasks.get(binding.task_id).heads;
    requireThat(heads.length === 1 && heads[0].eventId === children[0].eventId, 'STALE_EVIDENCE_FRONTIER');
    const completion = { event_id: children[0].eventId, event_path: eventPath(children[0]), run_id: binding.run_id,
      accepted_subject_digest: intent.accepted_subject_digest, snapshot: intent.snapshot };
    db.prepare('UPDATE operations SET completion=? WHERE execution_id=?').run(canonical(completion), binding.execution_id);
    return completion;
  };
  return {
    prepare(input, auth) {
      const execution = executionFor(input.execution_id);
      requireThat(auth?.role === 'owner' && auth.actor_id === execution.spec.accountable_owner, 'UNAUTHORIZED');
      requireThat(execution.revision === input.expected_revision, 'REVISION_CONFLICT');
      requireThat(!execution.worker_history.length && execution.attempt === null && !['COMPLETED', 'FAILED', 'CANCELLED'].includes(execution.state), 'EVIDENCE_AFTER_DISPATCH');
      filesList(input.files);
      const memory = loadMemory(localMemory(code.root)); const heads = memory.graph.tasks.get(execution.spec.task_id)?.heads;
      requireThat(heads?.length === 1 && ['start', 'checkpoint'].includes(heads[0].kind), 'GENUINE_START_REQUIRED');
      const head = heads[0]; const starts = memory.graph.runs.get(head.runId).filter((event) => event.kind === 'start');
      const binding = { schema_version: 1, execution_id: execution.execution_id, spec_digest: execution.spec_digest,
        task_id: execution.spec.task_id, repository: execution.spec.repository.repository_id,
        source_repository: input.source_repository, source_revision: input.source_revision, source_path: input.source_path,
        source_blob: head.task.blob, claim_heading: input.claim_heading, claim_digest: head.task.claim.digest,
        run_id: head.runId, start_event_id: starts[0].eventId, start_event_digest: digest(starts[0]),
        head_event_id: head.eventId, head_event_digest: digest(head),
        files: ordered(input.files), code_root: code.root, source_root: sourceGit.root };
      active(binding, execution, true);
      const prior = operation(execution.execution_id);
      requireThat(!prior || same(prior.binding, binding), 'BINDING_IMMUTABLE');
      transaction(() => {
        db.prepare('INSERT OR IGNORE INTO operations(execution_id,run_id,binding) VALUES(?,?,?)').run(execution.execution_id, binding.run_id, canonical(binding));
        requireThat(operation(execution.execution_id), 'RUN_ALREADY_BOUND');
      });
      store.bindEvidence({ execution_id: execution.execution_id, expected_revision: input.expected_revision, binding }, auth);
      return structuredClone(binding);
    },
    guard({ binding, execution }) { active(binding, execution, true); return true; },
    complete({ execution_id, summary, validation, next }) {
      const execution = executionFor(execution_id); accepted(execution);
      const binding = bound(execution_id); const prior = operation(execution_id);
      requireThat(prior && same(prior.binding, binding), 'JOURNAL_BINDING_MISMATCH');
      for (const value of [summary, validation, next]) requireThat(typeof value === 'string' && value.trim() && !value.startsWith('--'), 'OBSERVED_VALIDATION_REQUIRED');
      const request = { summary, validation, next, accepted_subject_digest: execution.accepted_subject_digest,
        snapshot: code.snapshotIndex(binding.files) };
      requireThat(same(request.snapshot, acceptedSnapshot(execution, binding)), 'ACCEPTED_CONTENT_MISMATCH',
        'Final staged files differ from the accepted workspace plan');
      if (prior.intent) {
        requireThat(same(prior.intent, request), 'COMPLETION_INTENT_CHANGED');
        const recovered = recover(binding, execution, prior.intent);
        requireThat(recovered, 'COMPLETION_OUTCOME_UNKNOWN', 'A prior helper invocation has no matching immutable child; inspect before retrying');
        return recovered;
      }
      active(binding, execution);
      transaction(() => {
        const fresh = operation(execution_id);
        requireThat(!fresh.intent, 'COMPLETION_IN_PROGRESS');
        db.prepare('UPDATE operations SET intent=? WHERE execution_id=?').run(canonical(request), execution_id);
      });
      // The durable intent precedes the helper. Its reply is never accepted as event proof.
      let helperError;
      try { invoke('complete', ['--id', binding.task_id, '--summary', summary, '--validation', validation, '--next', next]); }
      catch (error) { helperError = error; }
      const completion = recover(binding, execution, request);
      if (!completion) reject('COMPLETION_OUTCOME_UNKNOWN', helperError ? 'Helper failed without a matching immutable completion' : 'Helper returned without a matching immutable completion');
      return completion;
    },
    verify({ execution_id, base }) {
      requireThat(typeof base === 'string' && base.trim() && !base.startsWith('--'), 'EXPLICIT_BASE_REQUIRED');
      const execution = executionFor(execution_id); accepted(execution);
      const binding = bound(execution_id); const op = operation(execution_id);
      requireThat(op?.intent && op?.completion && same(op.binding, binding), 'COMPLETION_REQUIRED');
      requireThat(op.completion.accepted_subject_digest === execution.accepted_subject_digest, 'ACCEPTED_RESULT_CHANGED');
      requireThat(same(op.intent.snapshot, acceptedSnapshot(execution, binding)), 'ACCEPTED_CONTENT_MISMATCH');
      const localCompletion = recover(binding, execution, op.intent);
      requireThat(same(localCompletion, op.completion), 'COMPLETION_ASSOCIATION_MISMATCH');
      invoke('validate', ['--base', base, '--repository', binding.repository, '--source-repository', binding.source_repository]);
      const result = validateCommitted(code, { base, repository: binding.repository, backlogRepository: binding.source_repository });
      const committed = loadMemory(code.memoryAt(result.head));
      const event = committed.graph.byId.get(op.completion.event_id);
      requireThat(committed.graph.byId.has(binding.start_event_id) && committed.graph.byId.has(binding.head_event_id)
        && digest(committed.graph.byId.get(binding.start_event_id)) === binding.start_event_digest
        && digest(committed.graph.byId.get(binding.head_event_id)) === binding.head_event_digest, 'COMMITTED_START_MISMATCH');
      requireThat(event && matchCompletion(event, op.intent, binding)
        && !code.memoryAt(result.mergeBase).has(eventPath(event)), 'BOUND_COMPLETION_NOT_FRESH');
      const heads = committed.graph.tasks.get(binding.task_id)?.heads;
      requireThat(heads?.length === 1 && heads[0].eventId === event.eventId
        && binding.files.every((file) => same(code.entry(result.head, file), event.snapshot[file])), 'BOUND_COMPLETION_MISMATCH');
      return { ...result, execution_id, completion: structuredClone(op.completion) };
    },
    close() { db.close(); },
  };
}
