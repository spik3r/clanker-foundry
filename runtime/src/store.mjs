import { DatabaseSync } from 'node:sqlite';
import { createHash, randomUUID } from 'node:crypto';
import { closeSync, constants, lstatSync, openSync } from 'node:fs';
import { dirname, isAbsolute, join, parse, resolve } from 'node:path';
import { canonical, digest, parseJSON, validate } from './contract.mjs';
import { evaluate } from './state.mjs';
import { validateEvidenceBinding, validateNamespaceDecision, verifyLocalPin, verifyReviewedAdmission } from './authority.mjs';

const DATABASE_VERSION = 2;
const APPLICATION_ID = 0x43464f52;
const ROLES = new Set(['owner', 'coordinator', 'scheduler', 'worker', 'validator', 'approver']);

function parseStored(text) {
  // Validated aggregates and receipts grow across commands. Their serialized
  // size is independent of the 1MiB limit for a new external wire request.
  return parseJSON(text, { maxBytes: Math.max(1, Buffer.byteLength(text, 'utf8')) });
}

function reject(code, message) {
  const error = new Error(message);
  error.code = code;
  throw error;
}

function identity(auth) {
  if (!auth || typeof auth.actor_id !== 'string' || !auth.actor_id || !ROLES.has(auth.role)) {
    reject('UNAUTHORIZED', 'A trusted local actor identity and role are required');
  }
  return { actor_id: auth.actor_id, role: auth.role };
}

function privateFile(path, { directory = false } = {}) {
  const stat = lstatSync(path);
  if (stat.isSymbolicLink() || (directory ? !stat.isDirectory() : !stat.isFile()) ||
      stat.uid !== process.getuid() || (stat.mode & 0o077) !== 0 || (!directory && stat.nlink !== 1)) {
    reject('UNSAFE_DATABASE_PATH', 'Database and containing directory must be private, owned, regular paths');
  }
  return stat;
}

function preparePath(path) {
  if (typeof path !== 'string' || !isAbsolute(path)) reject('UNSAFE_DATABASE_PATH', 'Database path must be absolute');
  const normalized = resolve(path);
  let cursor = parse(normalized).root;
  for (const part of normalized.slice(cursor.length).split('/').slice(0, -1)) {
    cursor = join(cursor, part);
    if (lstatSync(cursor).isSymbolicLink()) reject('UNSAFE_DATABASE_PATH', 'Database path cannot traverse symlinks');
  }
  privateFile(dirname(normalized), { directory: true });
  try {
    const descriptor = openSync(normalized, constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW, 0o600);
    closeSync(descriptor);
  } catch (error) {
    if (error.code !== 'EEXIST') throw error;
  }
  for (const suffix of ['-wal', '-shm', '-journal']) {
    try { privateFile(`${normalized}${suffix}`); }
    catch (error) { if (error.code !== 'ENOENT') throw error; }
  }
  return { path: normalized, stat: privateFile(normalized) };
}

function timestamp(value) {
  const millis = typeof value === 'string' ? Date.parse(value) : NaN;
  if (!Number.isSafeInteger(millis) || new Date(millis).toISOString() !== value) reject('INVALID_TIME', 'Use a canonical UTC timestamp');
  return millis;
}

function requiredText(value, name) {
  if (typeof value !== 'string' || !value || value.length > 512) reject('INVALID_ARGUMENT', `${name} must be a bounded nonempty string`);
}

function registryPolicy(options) {
  const keys = ['trusted_checks', 'approver_ids', 'worker_ids'];
  const guards = ['evidence_guard', 'isolation_guard'];
  if (options === null || typeof options !== 'object' || Array.isArray(options) || ![Object.prototype, null].includes(Object.getPrototypeOf(options)) || Object.keys(options).some((key) => ![...keys, ...guards].includes(key))) {
    reject('INVALID_CONFIGURATION', 'Store configuration accepts registries and trusted evidence/isolation guards');
  }
  const values = {};
  for (const key of Reflect.ownKeys(options)) {
    const descriptor = Object.getOwnPropertyDescriptor(options, key);
    if (typeof key !== 'string' || !descriptor.enumerable || !Object.hasOwn(descriptor, 'value')) reject('INVALID_CONFIGURATION', 'Configuration accepts data properties only');
    if (guards.includes(key)) {
      if (typeof descriptor.value !== 'function') reject('INVALID_CONFIGURATION', 'Trusted guards must be functions');
    } else values[key] = descriptor.value;
  }
  canonical(values);
  const text = (value) => typeof value === 'string' && /^[A-Za-z0-9][A-Za-z0-9._:/-]{0,127}$/.test(value);
  const policy = structuredClone({ trusted_checks: [], approver_ids: [], worker_ids: [], ...values });
  for (const key of keys) {
    const registry = policy[key];
    if (!Array.isArray(registry) || registry.length > 256) reject('INVALID_CONFIGURATION', 'Registries must be arrays of at most 256 entries');
    if (key === 'trusted_checks') {
      const fields = ['check_id', 'check_version', 'runner_id'];
      if (registry.some((check) => check === null || typeof check !== 'object' || Array.isArray(check) || Object.keys(check).length !== fields.length || !fields.every((field) => Object.hasOwn(check, field) && text(check[field])))) {
        reject('INVALID_CONFIGURATION', 'Trusted checks must contain exact bounded check_id, check_version and runner_id triples');
      }
      for (const check of registry) Object.freeze(check);
    } else if (!registry.every(text)) reject('INVALID_CONFIGURATION', 'Identity registries must contain bounded identity strings');
    if (new Set(registry.map(canonical)).size !== registry.length) reject('INVALID_CONFIGURATION', 'Registries cannot contain duplicate entries');
    Object.freeze(registry);
  }
  for (const key of guards) if (Object.hasOwn(options, key)) policy[key] = options[key];
  return Object.freeze(policy);
}

/** Open one private, local SQLite authority. Caller identities are trusted local API inputs. */
export function openStore(databasePath, options = {}) {
  const policy = registryPolicy(options);
  const prepared = preparePath(databasePath);
  const db = new DatabaseSync(prepared.path);
  const reopened = privateFile(prepared.path);
  if (reopened.dev !== prepared.stat.dev || reopened.ino !== prepared.stat.ino) {
    db.close();
    reject('UNSAFE_DATABASE_PATH', 'Database path changed while opening');
  }
  try {
    db.exec('PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON; PRAGMA trusted_schema=OFF;');
    const version = db.prepare('PRAGMA user_version').get().user_version;
    const application = db.prepare('PRAGMA application_id').get().application_id;
    if (version !== 0 && (![1, DATABASE_VERSION].includes(version) || application !== APPLICATION_ID)) {
      reject('UNSUPPORTED_DATABASE', 'Database version or application identity is unsupported');
    }
    if (version === 0) {
      db.exec('BEGIN IMMEDIATE');
      try {
        // Recheck after acquiring the writer lock: another opener may have initialized it.
        const lockedVersion = db.prepare('PRAGMA user_version').get().user_version;
        if (lockedVersion === 0) {
          if (application !== 0 || db.prepare("SELECT name FROM sqlite_master WHERE type='table'").all().length) {
            reject('UNSUPPORTED_DATABASE', 'Refusing an unversioned database containing other data');
          }
          db.exec(`
            CREATE TABLE artifacts (digest TEXT PRIMARY KEY, content TEXT NOT NULL);
            CREATE TABLE subjects (digest TEXT PRIMARY KEY, manifest TEXT NOT NULL);
            CREATE TABLE specs (digest TEXT PRIMARY KEY, task_id TEXT NOT NULL, spec_revision INTEGER NOT NULL, body TEXT NOT NULL, UNIQUE(task_id,spec_revision));
            CREATE TABLE executions (admission_sequence INTEGER PRIMARY KEY AUTOINCREMENT, execution_id TEXT NOT NULL UNIQUE, task_id TEXT NOT NULL, spec_digest TEXT NOT NULL REFERENCES specs(digest), state TEXT NOT NULL, revision INTEGER NOT NULL, body TEXT NOT NULL);
            CREATE UNIQUE INDEX one_nonterminal_task ON executions(task_id) WHERE state NOT IN ('COMPLETED','FAILED','CANCELLED');
            CREATE TABLE attempts (attempt_id TEXT PRIMARY KEY, execution_id TEXT NOT NULL REFERENCES executions(execution_id), fencing_token INTEGER NOT NULL);
            CREATE TABLE events (global_sequence INTEGER PRIMARY KEY AUTOINCREMENT, execution_id TEXT NOT NULL REFERENCES executions(execution_id), revision INTEGER NOT NULL, event_id TEXT NOT NULL UNIQUE, body TEXT NOT NULL, command TEXT NOT NULL, context TEXT NOT NULL, UNIQUE(execution_id,revision));
            CREATE TABLE receipts (command_id TEXT PRIMARY KEY, command_digest TEXT NOT NULL, actor_id TEXT NOT NULL, role TEXT NOT NULL, body TEXT NOT NULL);
            CREATE TABLE reservations (execution_id TEXT NOT NULL REFERENCES executions(execution_id), kind TEXT NOT NULL, namespace TEXT NOT NULL, resource TEXT NOT NULL, mode TEXT NOT NULL, PRIMARY KEY(execution_id,kind,namespace,resource));
            CREATE TABLE outbox (sequence INTEGER PRIMARY KEY AUTOINCREMENT, effect_id TEXT NOT NULL UNIQUE, execution_id TEXT NOT NULL REFERENCES executions(execution_id), body TEXT NOT NULL, status TEXT NOT NULL CHECK(status IN ('PENDING','CLAIMED','UNKNOWN','DELIVERED')), dispatcher_id TEXT, claim_token TEXT, lease_deadline TEXT, outcome_receipt TEXT);
            PRAGMA application_id=${APPLICATION_ID}; PRAGMA user_version=${DATABASE_VERSION};
          `);
        } else if (![1, DATABASE_VERSION].includes(lockedVersion) || db.prepare('PRAGMA application_id').get().application_id !== APPLICATION_ID) {
          reject('UNSUPPORTED_DATABASE', 'Database version changed while opening');
        }
        db.exec('COMMIT');
      } catch (error) { db.exec('ROLLBACK'); throw error; }
    }
    db.exec('BEGIN IMMEDIATE');
    try {
      db.exec(`
        CREATE TABLE IF NOT EXISTS authority_instance (singleton INTEGER PRIMARY KEY CHECK(singleton=1), instance_id TEXT NOT NULL UNIQUE);
        CREATE TABLE IF NOT EXISTS namespaces (namespace_id TEXT PRIMARY KEY, generation INTEGER NOT NULL, body TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS namespace_decisions (decision_id TEXT PRIMARY KEY, namespace_id TEXT NOT NULL, generation INTEGER NOT NULL, digest TEXT NOT NULL, body TEXT NOT NULL, UNIQUE(namespace_id,generation));
        CREATE TABLE IF NOT EXISTS admissions (spec_digest TEXT PRIMARY KEY, namespace_id TEXT NOT NULL REFERENCES namespaces(namespace_id), generation INTEGER NOT NULL, body TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS admission_sources (spec_digest TEXT PRIMARY KEY REFERENCES admissions(spec_digest), digest TEXT NOT NULL, body TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS execution_authority (execution_id TEXT PRIMARY KEY REFERENCES executions(execution_id), namespace_id TEXT NOT NULL REFERENCES namespaces(namespace_id), admission_digest TEXT NOT NULL, body TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS evidence_bindings (execution_id TEXT PRIMARY KEY REFERENCES executions(execution_id), digest TEXT NOT NULL, body TEXT NOT NULL);
        CREATE TABLE IF NOT EXISTS isolation_proofs (digest TEXT PRIMARY KEY, execution_id TEXT NOT NULL REFERENCES executions(execution_id), body TEXT NOT NULL);
      `);
      db.prepare('INSERT OR IGNORE INTO authority_instance(singleton,instance_id) VALUES(1,?)').run(randomUUID());
      db.exec(`PRAGMA user_version=${DATABASE_VERSION}; COMMIT;`);
    } catch (error) { db.exec('ROLLBACK'); throw error; }
    db.exec('PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL;');
  } catch (error) { db.close(); throw error; }

  const transaction = (operation) => {
    db.exec('BEGIN IMMEDIATE');
    try { const result = operation(); db.exec('COMMIT'); return result; }
    catch (error) { db.exec('ROLLBACK'); throw error; }
  };
  const instanceId = db.prepare('SELECT instance_id FROM authority_instance WHERE singleton=1').get().instance_id;
  const namespaceStatus = (id) => {
    const row = db.prepare('SELECT body FROM namespaces WHERE namespace_id=?').get(id);
    if (!row) return null;
    const current = parseStored(row.body);
    const historical = db.prepare('SELECT digest,body FROM namespace_decisions WHERE decision_id=?').get(current.decision.decision_id);
    if (!historical || historical.digest !== current.decision_digest || digest(parseStored(historical.body)) !== historical.digest) reject('AUTHORITY_CORRUPTED', 'Namespace decision is missing or changed');
    const record = parseStored(historical.body);
    if (current.decision.namespace_id !== id || canonical(record.decision) !== canonical(current.decision) ||
        canonical(record.source) !== canonical(current.source) || record.generation !== current.generation) reject('AUTHORITY_CORRUPTED', 'Current namespace disagrees with its immutable decision');
    return current;
  };
  const namespaceFor = (taskId) => {
    const matching = db.prepare('SELECT namespace_id FROM namespaces').all().map((row) => namespaceStatus(row.namespace_id)).filter((record) => taskId.startsWith(record.decision.task_prefix));
    if (matching.length > 1) reject('NAMESPACE_OVERLAP', 'A task cannot match multiple authorities');
    return matching[0] ?? null;
  };
  const evidenceBinding = (executionId) => {
    const row = db.prepare('SELECT digest,body FROM evidence_bindings WHERE execution_id=?').get(executionId);
    if (!row) return null;
    const binding = parseStored(row.body);
    if (digest(binding) !== row.digest) reject('EVIDENCE_BINDING_CORRUPTED', 'Evidence binding digest changed');
    return binding;
  };
  const get = (id) => {
    const row = db.prepare('SELECT body FROM executions WHERE execution_id=?').get(id);
    return row ? parseStored(row.body) : null;
  };
  const specFor = (dependency) => {
    const row = db.prepare('SELECT body FROM specs WHERE task_id=? AND digest=?').get(dependency.task_id, dependency.spec_digest);
    if (!row) return null;
    const spec = parseStored(row.body);
    if (digest(spec) !== dependency.spec_digest) reject('SPEC_DIGEST_MISMATCH', 'Stored dependency spec does not match its immutable digest');
    return spec;
  };
  const referencesValid = (spec) => {
    const visited = new Set();
    const visit = (candidate) => {
      if (candidate.task_id === spec.task_id) reject('DEPENDENCY_CYCLE', 'Dependencies cannot lead back to the admitted task');
      const key = digest(candidate);
      if (visited.has(key)) return;
      visited.add(key);
      for (const dependency of candidate.dependencies) {
        const target = specFor(dependency);
        if (!target) reject('DEPENDENCY_NOT_FOUND', 'Dependency must name an admitted immutable spec');
        visit(target);
      }
    };
    for (const dependency of spec.dependencies) {
      const target = specFor(dependency);
      if (!target) reject('DEPENDENCY_NOT_FOUND', 'Dependency must name an admitted immutable spec');
      visit(target);
    }
    return true;
  };
  const selectedDependencyOutputs = (spec) => spec.dependencies.flatMap((dependency) => {
    const row = db.prepare("SELECT execution_id,body FROM executions WHERE task_id=? AND spec_digest=? AND state='COMPLETED' ORDER BY admission_sequence,execution_id LIMIT 1").get(dependency.task_id, dependency.spec_digest);
    if (!row) return [];
    const execution = parseStored(row.body);
    if (!execution.accepted_subject_digest) reject('INVALID_DEPENDENCY_OUTPUT', 'Completed dependency has no accepted output');
    return [{ ...dependency, execution_id: row.execution_id, subject_digest: execution.accepted_subject_digest }];
  });
  const dependenciesSatisfied = (spec) => selectedDependencyOutputs(spec).length === spec.dependencies.length;
  const requestedResources = (spec) => [
    ...spec.write_scope.map((resource) => ({ kind: 'path', namespace: spec.repository.repository_id, resource, mode: 'exclusive' })),
    ...spec.shared_resources.map(({ resource_id, mode }) => ({ kind: 'shared', namespace: '', resource: resource_id, mode })),
  ];
  const resourcesAvailable = (executionId, spec) => {
    const held = db.prepare('SELECT * FROM reservations WHERE execution_id<>?').all(executionId);
    return requestedResources(spec).every((request) => !held.some((other) => {
      if (request.kind !== other.kind || request.namespace !== other.namespace) return false;
      if (request.kind === 'shared') return request.resource === other.resource && (request.mode === 'exclusive' || other.mode === 'exclusive');
      return request.resource === other.resource || request.resource.startsWith(`${other.resource}/`) || other.resource.startsWith(`${request.resource}/`);
    }));
  };
  const getArtifact = (artifactDigest) => {
    if (typeof artifactDigest !== 'string' || !/^[a-f0-9]{64}$/.test(artifactDigest)) reject('INVALID_ARTIFACT', 'Artifact identity must be a SHA-256 digest');
    const row = db.prepare('SELECT content FROM artifacts WHERE digest=?').get(artifactDigest);
    if (!row) return null;
    if (createHash('sha256').update(row.content, 'utf8').digest('hex') !== artifactDigest) reject('ARTIFACT_CORRUPTED', 'Stored artifact does not match its digest');
    return row.content;
  };
  const getSubject = (subjectDigest) => {
    if (typeof subjectDigest !== 'string' || !/^[a-f0-9]{64}$/.test(subjectDigest)) reject('INVALID_ARTIFACT', 'Subject identity must be a SHA-256 digest');
    const row = db.prepare('SELECT manifest FROM subjects WHERE digest=?').get(subjectDigest);
    if (!row) return null;
    const manifest = parseStored(row.manifest);
    if (digest(manifest) !== subjectDigest) reject('SUBJECT_CORRUPTED', 'Stored subject manifest does not match its digest');
    for (const artifact of manifest) if (getArtifact(artifact.digest) === null) reject('ARTIFACT_NOT_FOUND', 'Stored subject references missing artifact bytes');
    return manifest;
  };
  const verifyResultArtifacts = (command) => {
    if (command.kind === 'SUBMIT' && command.payload.result.kind === 'SUCCEEDED') {
      for (const artifact of command.payload.result.artifacts) {
        if (getArtifact(artifact.digest) === null) reject('ARTIFACT_NOT_FOUND', 'Successful output must reference stored artifact bytes');
      }
    }
  };
  const effectRow = (id) => {
    const row = db.prepare('SELECT * FROM outbox WHERE effect_id=?').get(id);
    if (!row) reject('EFFECT_NOT_FOUND', 'No effect has that identity');
    return row;
  };
  const effectView = (row) => ({ effect: parseStored(row.body), status: row.status, dispatcher_id: row.dispatcher_id, claim_token: row.claim_token, lease_deadline: row.lease_deadline, receipt: row.outcome_receipt === null ? null : parseStored(row.outcome_receipt) });
  const acknowledgedRecovery = (row, execution, intent) => {
    if (intent.kind !== 'LAUNCH_ATTEMPT' || execution.state !== 'EXECUTING' || !row.outcome_receipt) return false;
    const proof = parseStored(row.outcome_receipt).reconciliation;
    return proof?.decision === 'CONFIRMED_ABSENT'
      && proof.actor_id === execution.spec.accountable_owner
      && proof.attempt_id === intent.payload.attempt.attempt_id
      && proof.fencing_token === execution.fence;
  };

  const history = (executionId) => {
    const query = 'SELECT events.command,events.context,receipts.body FROM events JOIN receipts ON receipts.command_id=events.event_id';
    const rows = executionId === undefined
      ? db.prepare(`${query} ORDER BY events.global_sequence`).all()
      : db.prepare(`${query} WHERE events.execution_id=? ORDER BY events.revision`).all(executionId);
    return rows.map((row) => ({ command: parseStored(row.command), context: parseStored(row.context), ...parseStored(row.body) }));
  };

  const authorityContext = (spec, executionId, admitting = false) => {
    const namespace = namespaceFor(spec.task_id);
    if (!namespace) {
      if (spec.integration_policy === 'isolated_write' || spec.write_scope.length > 0) reject('NAMESPACE_REQUIRED', 'Repository writes require a managed namespace');
      return { namespace: null, context: {} };
    }
    const decision = namespace.decision;
    if (decision.mode !== 'MANAGED') reject('NAMESPACE_NOT_MANAGED', 'Namespace does not permit admission or dispatch');
    if (decision.repository !== spec.repository.repository_id || decision.accountable_owner !== spec.accountable_owner) reject('NAMESPACE_IDENTITY_MISMATCH', 'Specification disagrees with namespace authority');
    let binding;
    if (admitting) {
      const row = db.prepare('SELECT body FROM admissions WHERE spec_digest=?').get(digest(spec));
      if (!row) reject('REVIEWED_ADMISSION_REQUIRED', 'Register the exact reviewed source and specification before admission');
      binding = parseStored(row.body);
      if (binding.namespace_id !== decision.namespace_id || binding.generation !== namespace.generation || binding.spec_digest !== digest(spec)) reject('STALE_ADMISSION', 'Admission review belongs to another namespace generation');
    } else {
      const row = db.prepare('SELECT admission_digest,body FROM execution_authority WHERE execution_id=?').get(executionId);
      if (!row) reject('EXPLICIT_MIGRATION_REQUIRED', 'Existing executions do not acquire namespace authority implicitly');
      binding = parseStored(row.body);
      if (digest(binding) !== row.admission_digest || binding.namespace_id !== decision.namespace_id || binding.spec_digest !== digest(spec)) reject('AUTHORITY_CORRUPTED', 'Execution authority disagrees with its specification');
    }
    if ((spec.integration_policy === 'isolated_write' || spec.write_scope.length > 0) && !decision.allow_writes) reject('NAMESPACE_WRITES_DISABLED', 'Namespace has not enabled isolated repository writes');
    const sourceRow = db.prepare('SELECT digest,body FROM admission_sources WHERE spec_digest=?').get(binding.spec_digest);
    if (!sourceRow || sourceRow.digest !== binding.source_digest || digest(parseStored(sourceRow.body)) !== sourceRow.digest) reject('AUTHORITY_CORRUPTED', 'Reviewed local source binding is missing or changed');
    const reviewed = parseStored(sourceRow.body);
    verifyReviewedAdmission({ spec, source: reviewed.source, source_root: reviewed.source_root, reviewed_by: decision.accountable_owner, namespace: decision });
    return { namespace, binding, context: { namespace_managed: true, can_write: decision.allow_writes,
      authority_namespace: decision.namespace_id, authority_generation: binding.generation,
      authority_decision_digest: binding.decision_digest, admission_digest: digest(binding) } };
  };
  const guardEvidence = (execution, namespace) => {
    const binding = evidenceBinding(execution.execution_id);
    if (!binding) {
      if (namespace?.decision.require_evidence || execution.spec.integration_policy === 'isolated_write') reject('EVIDENCE_REQUIRED', 'A genuine source and start must be bound before dispatch');
      return {};
    }
    validateEvidenceBinding(binding, execution);
    if (typeof policy.evidence_guard !== 'function') reject('EVIDENCE_GUARD_REQUIRED', 'Fresh evidence verification is required for a bound execution');
    if (policy.evidence_guard(structuredClone(binding), structuredClone(execution)) !== true) reject('EVIDENCE_GUARD_FAILED', 'Evidence verifier did not confirm the current bound frontier');
    return { evidence_binding_digest: digest(binding) };
  };
  const guardIsolation = (execution, attemptPayload) => {
    if (execution.spec.integration_policy !== 'isolated_write') return {};
    if (typeof policy.isolation_guard !== 'function') reject('ISOLATION_REQUIRED', 'A trusted host must verify the exact isolated attempt');
    const proof = policy.isolation_guard(structuredClone(execution), structuredClone(attemptPayload));
    canonical(proof);
    const fields = ['verified', 'execution_id', 'spec_digest', 'attempt_id', 'workspace', 'repository_revision', 'write_scope'];
    if (!proof || Array.isArray(proof) || Object.keys(proof).length !== fields.length || !fields.every((field) => Object.hasOwn(proof, field)) ||
        proof.verified !== true || proof.execution_id !== execution.execution_id || proof.spec_digest !== execution.spec_digest ||
        proof.attempt_id !== attemptPayload.attempt_id || proof.workspace !== attemptPayload.workspace ||
        proof.repository_revision !== execution.spec.repository.revision || !Array.isArray(proof.write_scope) ||
        canonical([...proof.write_scope].sort()) !== canonical([...execution.spec.write_scope].sort())) {
      reject('ISOLATION_PROOF_MISMATCH', 'Host proof must match this execution, attempt, workspace, revision and exact write scope');
    }
    const proofDigest = digest(proof);
    db.prepare('INSERT OR IGNORE INTO isolation_proofs(digest,execution_id,body) VALUES(?,?,?)').run(proofDigest, execution.execution_id, canonical(proof));
    return { isolation_verified: true, isolation_proof_digest: proofDigest };
  };

  return {
    instanceId() { return instanceId; },
    namespaceStatus,
    recordNamespaceDecision({ decision, source, source_root }, authContext) {
      const actor = identity(authContext);
      validateNamespaceDecision(decision);
      validate('AuthoritySourcePin', source);
      if (actor.role !== 'owner' || actor.actor_id !== decision.accountable_owner) reject('UNAUTHORIZED', 'Only the accountable owner may change namespace authority');
      if (decision.authority_instance_id !== instanceId) reject('AUTHORITY_INSTANCE_MISMATCH', 'Decision names another authority database');
      if (source.repository !== decision.source_repository) reject('SOURCE_REPOSITORY_MISMATCH', 'Decision source must belong to the declared source repository');
      verifyLocalPin({ root: source_root, source, expectedValue: decision });
      return transaction(() => {
        const prior = db.prepare('SELECT digest,body FROM namespace_decisions WHERE decision_id=?').get(decision.decision_id);
        const record = { decision, source, generation: decision.expected_generation + 1, actor_id: actor.actor_id };
        const recordDigest = digest(record);
        if (prior) {
          if (prior.digest !== recordDigest) reject('DECISION_ID_REUSED', 'Decision identity already has different content');
          return parseStored(canonical({ decision, source, generation: record.generation, decision_digest: prior.digest }));
        }
        const current = namespaceStatus(decision.namespace_id);
        if ((current?.generation ?? 0) !== decision.expected_generation) reject('GENERATION_CONFLICT', 'Expected namespace generation does not match current authority');
        const overlapping = db.prepare('SELECT namespace_id FROM namespaces WHERE namespace_id<>?').all(decision.namespace_id).map((row) => namespaceStatus(row.namespace_id)).some((other) => decision.task_prefix.startsWith(other.decision.task_prefix) || other.decision.task_prefix.startsWith(decision.task_prefix));
        if (overlapping) reject('NAMESPACE_OVERLAP', 'Namespace prefixes must remain disjoint');
        if (current) {
          const immutable = ['namespace_id', 'task_prefix', 'repository', 'source_repository', 'accountable_owner', 'authority_instance_id', 'require_evidence', 'allow_writes'];
          if (immutable.some((field) => canonical(current.decision[field]) !== canonical(decision[field]))) reject('NAMESPACE_IMMUTABLE', 'Namespace identity and capability policy cannot change');
          const allowed = { SHADOW: ['MANAGED', 'MANUAL'], MANAGED: ['DRAINING'], DRAINING: ['MANAGED', 'MANUAL'], MANUAL: ['SHADOW'] };
          if (!allowed[current.decision.mode].includes(decision.mode)) reject('ILLEGAL_AUTHORITY_TRANSITION', 'Authority must activate explicitly and drain before returning to manual');
        } else if (!['SHADOW', 'MANUAL'].includes(decision.mode)) reject('ILLEGAL_AUTHORITY_TRANSITION', 'A namespace must first be recorded without managed dispatch');
        const executions = db.prepare('SELECT body FROM executions').all().map((row) => parseStored(row.body)).filter((execution) => execution.spec.task_id.startsWith(decision.task_prefix));
        if (decision.mode === 'MANAGED') {
          if (!decision.prior_authority_drained) reject('PRIOR_AUTHORITY_NOT_DRAINED', 'Previous authority must be explicitly confirmed drained');
          if (executions.some((execution) => !db.prepare('SELECT 1 FROM execution_authority WHERE execution_id=?').get(execution.execution_id))) reject('EXPLICIT_MIGRATION_REQUIRED', 'Existing executions cannot transfer into a namespace silently');
        }
        if (current?.decision.mode === 'DRAINING' && decision.mode === 'MANUAL') {
          if (executions.some((execution) => !['COMPLETED', 'FAILED', 'CANCELLED'].includes(execution.state) || execution.cleanup_pending || execution.resources_held)) reject('NAMESPACE_NOT_DRAINED', 'Resolve executions and retained cleanup before manual authority resumes');
          const unsettled = db.prepare("SELECT execution_id FROM outbox WHERE status<>'DELIVERED'").all();
          if (unsettled.some((row) => executions.some((execution) => execution.execution_id === row.execution_id))) reject('NAMESPACE_NOT_DRAINED', 'Resolve all pending and unknown effect outcomes before manual authority resumes');
        }
        db.prepare('INSERT INTO namespace_decisions(decision_id,namespace_id,generation,digest,body) VALUES(?,?,?,?,?)').run(decision.decision_id, decision.namespace_id, record.generation, recordDigest, canonical(record));
        const state = { decision, source, generation: record.generation, decision_digest: recordDigest };
        db.prepare('INSERT INTO namespaces(namespace_id,generation,body) VALUES(?,?,?) ON CONFLICT(namespace_id) DO UPDATE SET generation=excluded.generation,body=excluded.body').run(decision.namespace_id, record.generation, canonical(state));
        return parseStored(canonical(state));
      });
    },
    registerAdmission({ namespace_id, generation, spec, source, source_root, reviewed_by }, authContext) {
      const actor = identity(authContext);
      validate('TaskSpec', spec);
      return transaction(() => {
        const namespace = namespaceStatus(namespace_id);
        if (!namespace || namespace.generation !== generation) reject('GENERATION_CONFLICT', 'Admission must use the current namespace generation');
        if (namespace.decision.mode !== 'MANAGED') reject('NAMESPACE_NOT_MANAGED', 'Admission requires an active managed namespace');
        if (actor.role !== 'owner' || actor.actor_id !== namespace.decision.accountable_owner || reviewed_by !== actor.actor_id) reject('UNAUTHORIZED', 'Only the accountable owner may register a reviewed admission');
        const proposal = verifyReviewedAdmission({ spec, source, source_root, reviewed_by, namespace: namespace.decision });
        const sourceBinding = { source, source_root };
        const binding = { namespace_id, generation, decision_digest: namespace.decision_digest, spec_digest: digest(spec), source_digest: digest(sourceBinding), proposal };
        const old = db.prepare('SELECT body FROM admissions WHERE spec_digest=?').get(binding.spec_digest);
        if (old && canonical(parseStored(old.body)) !== canonical(binding)) reject('ADMISSION_IMMUTABLE', 'A reviewed specification admission cannot change');
        db.prepare('INSERT OR IGNORE INTO admissions(spec_digest,namespace_id,generation,body) VALUES(?,?,?,?)').run(binding.spec_digest, namespace_id, generation, canonical(binding));
        db.prepare('INSERT OR IGNORE INTO admission_sources(spec_digest,digest,body) VALUES(?,?,?)').run(binding.spec_digest, digest(sourceBinding), canonical(sourceBinding));
        return parseStored(canonical(binding));
      });
    },
    evidenceBinding,
    verifyDispatch(executionId, attemptId) {
      return transaction(() => {
        const execution = get(executionId);
        if (!execution || !['ASSIGNED', 'EXECUTING'].includes(execution.state) || execution.cleanup_pending ||
            execution.attempt?.attempt_id !== attemptId || execution.attempt.fencing_token !== execution.fence) reject('STALE_ATTEMPT', 'Dispatch must name the exact live assigned attempt');
        const authority = authorityContext(execution.spec, executionId);
        guardEvidence(execution, authority.namespace);
        guardIsolation(execution, execution.attempt);
        return true;
      });
    },
    bindEvidence({ execution_id, expected_revision, binding }, authContext) {
      const actor = identity(authContext);
      return transaction(() => {
        const execution = get(execution_id);
        if (!execution) reject('EXECUTION_NOT_FOUND', 'No execution has that identity');
        if (actor.role !== 'owner' || actor.actor_id !== execution.spec.accountable_owner) reject('UNAUTHORIZED', 'Only the accountable owner may bind evidence');
        if (execution.revision !== expected_revision) reject('REVISION_CONFLICT', 'Evidence preparation must use the current execution revision');
        validateEvidenceBinding(binding, execution);
        const prior = evidenceBinding(execution_id);
        if (prior) {
          if (canonical(prior) !== canonical(binding)) reject('EVIDENCE_BINDING_IMMUTABLE', 'An execution cannot switch evidence runs or sources');
          return prior;
        }
        if (execution.worker_history.length !== 0 || execution.state !== 'PENDING') reject('EVIDENCE_TOO_LATE', 'Bind a genuine source and start before the first assignment');
        const admission = db.prepare('SELECT body FROM execution_authority WHERE execution_id=?').get(execution_id);
        if (admission) {
          const source = parseStored(admission.body).proposal.source;
          if (binding.source_repository !== source.repository || binding.source_revision !== source.revision || binding.source_path !== source.path ||
              binding.source_blob !== source.blob || binding.claim_heading !== source.claim.heading || binding.claim_digest !== source.claim.digest) reject('EVIDENCE_SOURCE_MISMATCH', 'Evidence must use the exact admitted source and claim');
        }
        db.prepare('INSERT INTO evidence_bindings(execution_id,digest,body) VALUES(?,?,?)').run(execution_id, digest(binding), canonical(binding));
        return parseStored(canonical(binding));
      });
    },
    execute(command, authContext) {
      const actor = identity(authContext);
      // Canonicalization rejects JS values that JSON/schema validation cannot represent.
      const wire = canonical(command);
      if (Buffer.byteLength(wire, 'utf8') > 1_048_576) reject('JSON_TOO_LARGE', 'Command exceeds the 1MiB external wire limit');
      const commandDigest = createHash('sha256').update(wire, 'utf8').digest('hex');
      validate('Command', command);
      return transaction(() => {
        const priorReceipt = db.prepare('SELECT * FROM receipts WHERE command_id=?').get(command.command_id);
        if (priorReceipt) {
          if (priorReceipt.actor_id !== actor.actor_id || priorReceipt.role !== actor.role) reject('UNAUTHORIZED', 'Command receipt belongs to another principal');
          if (priorReceipt.command_digest !== commandDigest) reject('COMMAND_ID_REUSED', 'Command ID was already used with another payload');
          return parseStored(priorReceipt.body);
        }
        const previous = get(command.execution_id);
        if ((previous?.revision ?? 0) !== command.expected_revision) reject('REVISION_CONFLICT', 'Expected revision does not match committed state');
        const spec = command.kind === 'ADMIT' ? command.payload.spec : previous?.spec;
        if (!spec) reject('EXECUTION_NOT_FOUND', 'No execution has that identity');
        const authority = ['ADMIT', 'ASSIGN'].includes(command.kind) ? authorityContext(spec, command.execution_id, command.kind === 'ADMIT') : { context: {} };
        const context = { ...actor, trusted_checks: policy.trusted_checks, approver_ids: policy.approver_ids,
          ...authority.context,
          ...(command.kind === 'ASSIGN' ? guardEvidence(previous, authority.namespace) : {}),
          ...(command.kind === 'ASSIGN' ? guardIsolation(previous, command.payload) : {}),
          references_valid: command.kind === 'ADMIT' ? referencesValid(spec) : true,
          dependencies_satisfied: dependenciesSatisfied(spec), dependency_outputs: selectedDependencyOutputs(spec),
          resources_available: resourcesAvailable(command.execution_id, spec),
          worker_eligible: command.kind === 'ASSIGN' && policy.worker_ids.includes(command.payload.worker_id) };
        if (command.kind === 'ADMIT') {
          const active = db.prepare("SELECT execution_id FROM executions WHERE task_id=? AND state NOT IN ('COMPLETED','FAILED','CANCELLED')").get(spec.task_id);
          if (active) reject('TASK_ALREADY_ACTIVE', 'Task already has a nonterminal execution');
          const retained = db.prepare('SELECT body FROM executions WHERE task_id=?').all(spec.task_id).some((row) => {
            const prior = parseStored(row.body);
            return prior.cleanup_pending || prior.resources_held;
          });
          if (retained) reject('CLEANUP_REQUIRED', 'Prior execution must confirm cleanup before task resubmission');
          const frozen = db.prepare('SELECT digest FROM specs WHERE task_id=? AND spec_revision=?').get(spec.task_id, spec.spec_revision);
          if (frozen && frozen.digest !== digest(spec)) reject('SPEC_IMMUTABLE', 'A spec revision cannot change its digest');
        }
        if (command.kind === 'ASSIGN' && db.prepare('SELECT 1 FROM attempts WHERE attempt_id=?').get(command.payload.attempt_id)) {
          reject('DUPLICATE_ATTEMPT', 'Attempt identity was already assigned to an execution');
        }
        verifyResultArtifacts(command);
        const { execution, event, effects } = evaluate(previous, command, context);
        validate('TaskExecution', execution);
        validate('TaskEvent', event);
        for (const effect of effects) validate('Effect', effect);
        if (command.kind === 'SUBMIT' && command.payload.result.kind === 'SUCCEEDED') {
          const result = command.payload.result;
          const existing = getSubject(result.subject_digest);
          if (existing !== null && canonical(existing) !== canonical(result.artifacts)) reject('SUBJECT_IMMUTABLE', 'Subject manifest cannot change');
          db.prepare('INSERT OR IGNORE INTO subjects(digest,manifest) VALUES(?,?)').run(result.subject_digest, canonical(result.artifacts));
        }
        if (execution.resources_held && !resourcesAvailable(execution.execution_id, execution.spec)) reject('RESOURCE_CONFLICT', 'Another execution retains a conflicting reservation');
        if (!previous) {
          db.prepare('INSERT OR IGNORE INTO specs(digest,task_id,spec_revision,body) VALUES(?,?,?,?)').run(execution.spec_digest, execution.spec.task_id, execution.spec.spec_revision, canonical(execution.spec));
          db.prepare('INSERT INTO executions(execution_id,task_id,spec_digest,state,revision,body) VALUES(?,?,?,?,?,?)').run(execution.execution_id, execution.spec.task_id, execution.spec_digest, execution.state, execution.revision, canonical(execution));
          if (authority.binding) db.prepare('INSERT INTO execution_authority(execution_id,namespace_id,admission_digest,body) VALUES(?,?,?,?)').run(execution.execution_id, authority.binding.namespace_id, digest(authority.binding), canonical(authority.binding));
        } else {
          if (execution.spec_digest !== previous.spec_digest || digest(execution.spec) !== previous.spec_digest) reject('SPEC_IMMUTABLE', 'Execution cannot change its admitted spec');
          db.prepare('UPDATE executions SET state=?,revision=?,body=? WHERE execution_id=?').run(execution.state, execution.revision, canonical(execution), execution.execution_id);
        }
        if (command.kind === 'ASSIGN') {
          db.prepare('INSERT INTO attempts(attempt_id,execution_id,fencing_token) VALUES(?,?,?)').run(execution.attempt.attempt_id, execution.execution_id, execution.attempt.fencing_token);
        }
        if (execution.resources_held) {
          for (const resource of requestedResources(execution.spec)) db.prepare('INSERT OR IGNORE INTO reservations VALUES(?,?,?,?,?)').run(execution.execution_id, resource.kind, resource.namespace, resource.resource, resource.mode);
        } else if (previous?.resources_held) {
          if (command.kind !== 'CLEANUP_CONFIRMED') reject('CLEANUP_REQUIRED', 'Reservations can only be released by confirmed cleanup');
          db.prepare('DELETE FROM reservations WHERE execution_id=?').run(execution.execution_id);
        }
        db.prepare('INSERT INTO events(execution_id,revision,event_id,body,command,context) VALUES(?,?,?,?,?,?)').run(execution.execution_id, execution.revision, event.event_id, canonical(event), canonical(command), canonical(context));
        for (const effect of effects) db.prepare("INSERT INTO outbox(effect_id,execution_id,body,status) VALUES(?,?,?,'PENDING')").run(effect.effect_id, execution.execution_id, canonical(effect));
        const receipt = { command_id: command.command_id, execution, event, effects };
        db.prepare('INSERT INTO receipts VALUES(?,?,?,?,?)').run(command.command_id, commandDigest, actor.actor_id, actor.role, canonical(receipt));
        return parseStored(canonical(receipt));
      });
    },
    get,
    getArtifact,
    getSubject,
    dependencyOutputs(executionId) {
      const execution = get(executionId);
      if (!execution) reject('EXECUTION_NOT_FOUND', 'No execution has that identity');
      return selectedDependencyOutputs(execution.spec);
    },
    putArtifact({ digest: artifactDigest, content }) {
      if (typeof content !== 'string' || Buffer.byteLength(content, 'utf8') > 1_048_576) reject('INVALID_ARTIFACT', 'Artifact must be UTF-8 text no larger than 1MiB');
      canonical(content);
      if (createHash('sha256').update(content, 'utf8').digest('hex') !== artifactDigest) reject('ARTIFACT_DIGEST_MISMATCH', 'Artifact bytes do not match their declared digest');
      return transaction(() => {
        const existing = getArtifact(artifactDigest);
        if (existing !== null && existing !== content) reject('ARTIFACT_IMMUTABLE', 'Stored artifact bytes cannot change');
        db.prepare('INSERT OR IGNORE INTO artifacts(digest,content) VALUES(?,?)').run(artifactDigest, content);
        return artifactDigest;
      });
    },
    history,
    replay() { return transaction(() => {
      const reconstructed = new Map();
      const events = []; const effects = []; const attempts = new Map();
      for (const record of history()) {
        if (record.context.admission_digest) {
          const row = db.prepare('SELECT admission_digest,body FROM execution_authority WHERE execution_id=?').get(record.command.execution_id);
          if (!row || row.admission_digest !== record.context.admission_digest || digest(parseStored(row.body)) !== row.admission_digest) reject('REPLAY_MISMATCH', 'Historical admission binding is missing or changed');
          const binding = parseStored(row.body);
          const source = db.prepare('SELECT digest,body FROM admission_sources WHERE spec_digest=?').get(binding.spec_digest);
          if (!source || source.digest !== binding.source_digest || digest(parseStored(source.body)) !== source.digest) reject('REPLAY_MISMATCH', 'Historical reviewed source binding is missing or changed');
          const decision = db.prepare('SELECT digest,body FROM namespace_decisions WHERE namespace_id=? AND generation=?').get(binding.namespace_id, binding.generation);
          if (!decision || decision.digest !== binding.decision_digest || digest(parseStored(decision.body)) !== decision.digest) reject('REPLAY_MISMATCH', 'Historical namespace decision is missing or changed');
        }
        if (record.context.evidence_binding_digest) {
          const binding = evidenceBinding(record.command.execution_id);
          if (!binding || digest(binding) !== record.context.evidence_binding_digest) reject('REPLAY_MISMATCH', 'Historical evidence binding is missing or changed');
        }
        if (record.context.isolation_proof_digest) {
          const proof = db.prepare('SELECT execution_id,body FROM isolation_proofs WHERE digest=?').get(record.context.isolation_proof_digest);
          if (!proof || proof.execution_id !== record.command.execution_id || digest(parseStored(proof.body)) !== record.context.isolation_proof_digest) reject('REPLAY_MISMATCH', 'Historical isolation proof is missing or changed');
        }
        verifyResultArtifacts(record.command);
        if (record.command.kind === 'SUBMIT' && record.command.payload.result.kind === 'SUCCEEDED' && getSubject(record.command.payload.result.subject_digest) === null) reject('SUBJECT_NOT_FOUND', 'Accepted output subject is missing');
        const result = evaluate(reconstructed.get(record.command.execution_id) ?? null, record.command, record.context);
        if (record.command.kind === 'ASSIGN') {
          const attempt = result.execution.attempt;
          if (attempts.has(attempt.attempt_id)) reject('REPLAY_MISMATCH', 'Accepted commands reused a global attempt identity');
          attempts.set(attempt.attempt_id, { attempt_id: attempt.attempt_id, execution_id: result.execution.execution_id, fencing_token: attempt.fencing_token });
        }
        if (digest(result.execution) !== digest(record.execution) || digest(result.event) !== digest(record.event) || digest(result.effects) !== digest(record.effects)) {
          reject('REPLAY_MISMATCH', 'Recorded accepted input does not reproduce its receipt');
        }
        reconstructed.set(record.command.execution_id, result.execution);
        events.push(result.event); effects.push(...result.effects);
      }
      const persisted = db.prepare('SELECT body FROM executions ORDER BY admission_sequence,execution_id').all().map((row) => parseStored(row.body));
      if (digest(persisted) !== digest([...reconstructed.values()])) reject('REPLAY_MISMATCH', 'Replayed state does not match persisted execution state');
      const recordedAttempts = db.prepare('SELECT * FROM attempts').all();
      const sortedRecords = (items) => items.map(canonical).sort();
      if (digest(sortedRecords(recordedAttempts)) !== digest(sortedRecords([...attempts.values()]))) reject('REPLAY_MISMATCH', 'Replayed attempt identities do not match their permanent registry');
      const persistedEvents = db.prepare('SELECT body FROM events ORDER BY global_sequence').all().map((row) => parseStored(row.body));
      if (digest(persistedEvents) !== digest(events)) reject('REPLAY_MISMATCH', 'Replayed events do not match durable events');
      const expectedReservations = [...reconstructed.values()].filter((execution) => execution.resources_held).flatMap((execution) => requestedResources(execution.spec).map((resource) => ({ execution_id: execution.execution_id, ...resource })));
      const actualReservations = db.prepare('SELECT * FROM reservations').all();
      const sortReservations = (items) => items.map(canonical).sort();
      if (digest(sortReservations(expectedReservations)) !== digest(sortReservations(actualReservations))) reject('REPLAY_MISMATCH', 'Replayed resource reservations do not match durable reservations');
      const persistedEffects = db.prepare('SELECT body FROM outbox ORDER BY sequence').all().map((row) => parseStored(row.body));
      if (digest(persistedEffects) !== digest(effects)) reject('REPLAY_MISMATCH', 'Replayed effects do not match durable outbox intents');
      // These records were validated individually; a growing database response is
      // not an external JSON request subject to the parser's 1MiB wire limit.
      const response = canonical({ executions: [...reconstructed.values()], events, effects });
      return parseStored(response);
    }); },
    list() { return db.prepare('SELECT body FROM executions ORDER BY admission_sequence,execution_id').all().map((row) => parseStored(row.body)); },
    events(executionId) {
      const rows = executionId === undefined ? db.prepare('SELECT body FROM events ORDER BY global_sequence').all() : db.prepare('SELECT body FROM events WHERE execution_id=? ORDER BY revision').all(executionId);
      return rows.map((row) => parseStored(row.body));
    },
    pendingEffects() { return db.prepare("SELECT * FROM outbox WHERE status<>'DELIVERED' ORDER BY sequence").all().map(effectView); },
    claimEffect({ effectId, dispatcherId, now, leaseMs }) {
      requiredText(effectId, 'effectId'); requiredText(dispatcherId, 'dispatcherId');
      const at = timestamp(now);
      if (!Number.isSafeInteger(leaseMs) || leaseMs < 1 || leaseMs > 3_600_000) reject('INVALID_ARGUMENT', 'Effect lease must be between 1ms and one hour');
      return transaction(() => {
        const row = effectRow(effectId);
        if (row.status === 'DELIVERED' || row.status === 'UNKNOWN') return effectView(row);
        if (row.status === 'CLAIMED') {
          if (timestamp(row.lease_deadline) <= at) {
            db.prepare("UPDATE outbox SET status='UNKNOWN' WHERE effect_id=?").run(effectId);
            return effectView(effectRow(effectId));
          }
          reject('EFFECT_ALREADY_CLAIMED', 'Effect already has a live dispatcher claim');
        }
        const intent = parseStored(row.body);
        const execution = get(row.execution_id);
        const attempt = intent.kind === 'LAUNCH_ATTEMPT' ? intent.payload.attempt : intent.payload;
        const resumeAfterAck = acknowledgedRecovery(row, execution, intent);
        if (intent.kind !== 'REVOKE_CAPABILITIES' && (execution.cleanup_pending ||
            execution.attempt?.attempt_id !== attempt.attempt_id || execution.fence !== attempt.fencing_token ||
            (intent.kind === 'LAUNCH_ATTEMPT' ? execution.state !== 'ASSIGNED' && !resumeAfterAck : execution.state !== 'VALIDATING'))) {
          db.prepare("UPDATE outbox SET status='DELIVERED',outcome_receipt=? WHERE effect_id=?").run(canonical({ skipped: 'STALE_ATTEMPT' }), effectId);
          return effectView(effectRow(effectId));
        }
        if (intent.kind === 'LAUNCH_ATTEMPT') {
          const authority = authorityContext(execution.spec, execution.execution_id);
          guardEvidence(execution, authority.namespace);
          guardIsolation(execution, intent.payload.attempt);
        }
        const token = randomUUID();
        db.prepare("UPDATE outbox SET status='CLAIMED',dispatcher_id=?,claim_token=?,lease_deadline=? WHERE effect_id=?").run(dispatcherId, token, new Date(at + leaseMs).toISOString(), effectId);
        return { ...effectView(effectRow(effectId)), resume_after_ack: resumeAfterAck };
      });
    },
    acknowledgeEffect({ effectId, dispatcherId, claimToken, outcome, receipt }) {
      if (!['DELIVERED', 'UNKNOWN'].includes(outcome)) reject('INVALID_ARGUMENT', 'Effect outcome must be DELIVERED or UNKNOWN');
      const serialized = canonical(receipt);
      if (Buffer.byteLength(serialized) > 65_536) reject('INVALID_ARGUMENT', 'Effect receipt exceeds 64KiB');
      return transaction(() => {
        const row = effectRow(effectId);
        if (row.dispatcher_id !== dispatcherId || row.claim_token !== claimToken || !claimToken) reject('STALE_EFFECT_CLAIM', 'Effect acknowledgement does not match its claim');
        if (row.status === outcome && row.outcome_receipt === serialized) return effectView(row);
        if (row.status !== 'CLAIMED') reject('EFFECT_OUTCOME_CONFLICT', 'Effect already has a recorded outcome');
        db.prepare('UPDATE outbox SET status=?,outcome_receipt=? WHERE effect_id=?').run(outcome, serialized, effectId);
        return effectView(effectRow(effectId));
      });
    },
    reconcileEffect({ effectId, outcome, receipt }, authContext) {
      const actor = identity(authContext);
      if (!['DELIVERED', 'PENDING'].includes(outcome)) reject('INVALID_ARGUMENT', 'Reconciliation must confirm delivery or absence');
      const serialized = canonical(receipt);
      if (Buffer.byteLength(serialized) > 65_536) reject('INVALID_ARGUMENT', 'Effect receipt exceeds 64KiB');
      return transaction(() => {
        const row = effectRow(effectId);
        const execution = get(row.execution_id);
        if (actor.role !== 'owner' || actor.actor_id !== execution.spec.accountable_owner) reject('UNAUTHORIZED', 'Only the accountable owner may reconcile an unknown effect');
        if (row.status !== 'UNKNOWN') reject('EFFECT_OUTCOME_CONFLICT', 'Only unknown effects need reconciliation');
        // PENDING means the owner confirmed the effect never occurred, not a blind retry.
        const intent = parseStored(row.body);
        const reconciledReceipt = canonical({
          reconciliation: {
            decision: outcome === 'PENDING' ? 'CONFIRMED_ABSENT' : 'CONFIRMED_DELIVERED',
            actor_id: actor.actor_id,
            ...(intent.kind === 'LAUNCH_ATTEMPT' ? {
              attempt_id: intent.payload.attempt.attempt_id,
              fencing_token: intent.payload.attempt.fencing_token,
            } : {}),
          },
          evidence: receipt,
        });
        if (Buffer.byteLength(reconciledReceipt, 'utf8') > 65_536) reject('INVALID_ARGUMENT', 'Reconciliation receipt exceeds 64KiB');
        db.prepare('UPDATE outbox SET status=?,dispatcher_id=NULL,claim_token=NULL,lease_deadline=NULL,outcome_receipt=? WHERE effect_id=?').run(outcome, reconciledReceipt, effectId);
        return effectView(effectRow(effectId));
      });
    },
    close() { db.close(); },
  };
}
