import { DatabaseSync } from 'node:sqlite';
import { createHash, randomUUID } from 'node:crypto';
import { closeSync, constants, lstatSync, openSync } from 'node:fs';
import { dirname, isAbsolute, join, parse, resolve } from 'node:path';
import { canonical, digest, parseJSON, validate } from './contract.mjs';
import { evaluate } from './state.mjs';

const DATABASE_VERSION = 1;
const APPLICATION_ID = 0x43464f52;
const ROLES = new Set(['owner', 'coordinator', 'scheduler', 'worker', 'validator', 'approver']);

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
  canonical(options);
  const keys = ['trusted_checks', 'approver_ids', 'worker_ids'];
  if (options === null || typeof options !== 'object' || Array.isArray(options) || Object.keys(options).some((key) => !keys.includes(key))) {
    reject('INVALID_CONFIGURATION', 'Store configuration accepts only trusted_checks, approver_ids and worker_ids');
  }
  const text = (value) => typeof value === 'string' && /^[A-Za-z0-9][A-Za-z0-9._:/-]{0,127}$/.test(value);
  const policy = structuredClone({ trusted_checks: [], approver_ids: [], worker_ids: [], ...options });
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
    if (version !== 0 && (version !== DATABASE_VERSION || application !== APPLICATION_ID)) {
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
        } else if (lockedVersion !== DATABASE_VERSION || db.prepare('PRAGMA application_id').get().application_id !== APPLICATION_ID) {
          reject('UNSUPPORTED_DATABASE', 'Database version changed while opening');
        }
        db.exec('COMMIT');
      } catch (error) { db.exec('ROLLBACK'); throw error; }
    }
    db.exec('PRAGMA journal_mode=WAL; PRAGMA synchronous=FULL;');
  } catch (error) { db.close(); throw error; }

  const transaction = (operation) => {
    db.exec('BEGIN IMMEDIATE');
    try { const result = operation(); db.exec('COMMIT'); return result; }
    catch (error) { db.exec('ROLLBACK'); throw error; }
  };
  const get = (id) => {
    const row = db.prepare('SELECT body FROM executions WHERE execution_id=?').get(id);
    return row ? parseJSON(row.body) : null;
  };
  const specFor = (dependency) => {
    const row = db.prepare('SELECT body FROM specs WHERE task_id=? AND digest=?').get(dependency.task_id, dependency.spec_digest);
    if (!row) return null;
    const spec = parseJSON(row.body);
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
    const execution = parseJSON(row.body);
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
    const manifest = parseJSON(row.manifest);
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
  const effectView = (row) => ({ effect: parseJSON(row.body), status: row.status, dispatcher_id: row.dispatcher_id, claim_token: row.claim_token, lease_deadline: row.lease_deadline, receipt: row.outcome_receipt === null ? null : parseJSON(row.outcome_receipt) });

  const history = (executionId) => {
    const query = 'SELECT events.command,events.context,receipts.body FROM events JOIN receipts ON receipts.command_id=events.event_id';
    const rows = executionId === undefined
      ? db.prepare(`${query} ORDER BY events.global_sequence`).all()
      : db.prepare(`${query} WHERE events.execution_id=? ORDER BY events.revision`).all(executionId);
    return rows.map((row) => ({ command: parseJSON(row.command), context: parseJSON(row.context), ...parseJSON(row.body) }));
  };

  return {
    execute(command, authContext) {
      const actor = identity(authContext);
      // Canonicalization rejects JS values that JSON/schema validation cannot represent.
      const commandDigest = digest(command);
      validate('Command', command);
      return transaction(() => {
        const priorReceipt = db.prepare('SELECT * FROM receipts WHERE command_id=?').get(command.command_id);
        if (priorReceipt) {
          if (priorReceipt.actor_id !== actor.actor_id || priorReceipt.role !== actor.role) reject('UNAUTHORIZED', 'Command receipt belongs to another principal');
          if (priorReceipt.command_digest !== commandDigest) reject('COMMAND_ID_REUSED', 'Command ID was already used with another payload');
          return parseJSON(priorReceipt.body);
        }
        const previous = get(command.execution_id);
        if ((previous?.revision ?? 0) !== command.expected_revision) reject('REVISION_CONFLICT', 'Expected revision does not match committed state');
        const spec = command.kind === 'ADMIT' ? command.payload.spec : previous?.spec;
        if (!spec) reject('EXECUTION_NOT_FOUND', 'No execution has that identity');
        const context = { ...actor, trusted_checks: policy.trusted_checks, approver_ids: policy.approver_ids,
          references_valid: command.kind === 'ADMIT' ? referencesValid(spec) : true,
          dependencies_satisfied: dependenciesSatisfied(spec), dependency_outputs: selectedDependencyOutputs(spec),
          resources_available: resourcesAvailable(command.execution_id, spec),
          worker_eligible: command.kind === 'ASSIGN' && policy.worker_ids.includes(command.payload.worker_id) };
        if (command.kind === 'ADMIT') {
          const active = db.prepare("SELECT execution_id FROM executions WHERE task_id=? AND state NOT IN ('COMPLETED','FAILED','CANCELLED')").get(spec.task_id);
          if (active) reject('TASK_ALREADY_ACTIVE', 'Task already has a nonterminal execution');
          const retained = db.prepare('SELECT body FROM executions WHERE task_id=?').all(spec.task_id).some((row) => {
            const prior = parseJSON(row.body);
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
        return parseJSON(canonical(receipt));
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
      const persisted = db.prepare('SELECT body FROM executions ORDER BY admission_sequence,execution_id').all().map((row) => parseJSON(row.body));
      if (digest(persisted) !== digest([...reconstructed.values()])) reject('REPLAY_MISMATCH', 'Replayed state does not match persisted execution state');
      const recordedAttempts = db.prepare('SELECT * FROM attempts').all();
      const sortedRecords = (items) => items.map(canonical).sort();
      if (digest(sortedRecords(recordedAttempts)) !== digest(sortedRecords([...attempts.values()]))) reject('REPLAY_MISMATCH', 'Replayed attempt identities do not match their permanent registry');
      const persistedEvents = db.prepare('SELECT body FROM events ORDER BY global_sequence').all().map((row) => parseJSON(row.body));
      if (digest(persistedEvents) !== digest(events)) reject('REPLAY_MISMATCH', 'Replayed events do not match durable events');
      const expectedReservations = [...reconstructed.values()].filter((execution) => execution.resources_held).flatMap((execution) => requestedResources(execution.spec).map((resource) => ({ execution_id: execution.execution_id, ...resource })));
      const actualReservations = db.prepare('SELECT * FROM reservations').all();
      const sortReservations = (items) => items.map(canonical).sort();
      if (digest(sortReservations(expectedReservations)) !== digest(sortReservations(actualReservations))) reject('REPLAY_MISMATCH', 'Replayed resource reservations do not match durable reservations');
      const persistedEffects = db.prepare('SELECT body FROM outbox ORDER BY sequence').all().map((row) => parseJSON(row.body));
      if (digest(persistedEffects) !== digest(effects)) reject('REPLAY_MISMATCH', 'Replayed effects do not match durable outbox intents');
      return parseJSON(canonical({ executions: [...reconstructed.values()], events, effects }));
    }); },
    list() { return db.prepare('SELECT body FROM executions ORDER BY admission_sequence,execution_id').all().map((row) => parseJSON(row.body)); },
    events(executionId) {
      const rows = executionId === undefined ? db.prepare('SELECT body FROM events ORDER BY global_sequence').all() : db.prepare('SELECT body FROM events WHERE execution_id=? ORDER BY revision').all(executionId);
      return rows.map((row) => parseJSON(row.body));
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
        const intent = parseJSON(row.body);
        const execution = get(row.execution_id);
        const attempt = intent.kind === 'LAUNCH_ATTEMPT' ? intent.payload.attempt : intent.payload;
        if (intent.kind !== 'REVOKE_CAPABILITIES' && (execution.cleanup_pending ||
            execution.attempt?.attempt_id !== attempt.attempt_id || execution.fence !== attempt.fencing_token ||
            (intent.kind === 'LAUNCH_ATTEMPT' ? execution.state !== 'ASSIGNED' : execution.state !== 'VALIDATING'))) {
          db.prepare("UPDATE outbox SET status='DELIVERED',outcome_receipt=? WHERE effect_id=?").run(canonical({ skipped: 'STALE_ATTEMPT' }), effectId);
          return effectView(effectRow(effectId));
        }
        const token = randomUUID();
        db.prepare("UPDATE outbox SET status='CLAIMED',dispatcher_id=?,claim_token=?,lease_deadline=? WHERE effect_id=?").run(dispatcherId, token, new Date(at + leaseMs).toISOString(), effectId);
        return effectView(effectRow(effectId));
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
        db.prepare('UPDATE outbox SET status=?,dispatcher_id=NULL,claim_token=NULL,lease_deadline=NULL,outcome_receipt=? WHERE effect_id=?').run(outcome, serialized, effectId);
        return effectView(effectRow(effectId));
      });
    },
    close() { db.close(); },
  };
}
