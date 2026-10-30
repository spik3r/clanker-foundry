#!/usr/bin/env node
import { constants, openSync, fstatSync, readFileSync, closeSync, mkdtempSync, realpathSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
import { randomUUID, createHash } from 'node:crypto';
import { parseJSON, validate } from './contract.mjs';
import { openStore } from './store.mjs';
import { FakeAdapter } from './adapters.mjs';
import { drainOutbox } from './runner.mjs';
import { importReviewedTask, exportTaskView } from './import-export.mjs';

const usage = `Usage:
  node runtime/src/cli.mjs demo [ABSOLUTE_DATABASE_PATH]
  node runtime/src/cli.mjs validate TYPE JSON_FILE
  node runtime/src/cli.mjs command DATABASE POLICY_FILE COMMAND_FILE ACTOR ROLE
  node runtime/src/cli.mjs show DATABASE EXECUTION_ID
  node runtime/src/cli.mjs export DATABASE EXECUTION_ID
  node runtime/src/cli.mjs replay DATABASE
  node runtime/src/cli.mjs import REVIEW_INPUT_FILE
  node runtime/src/cli.mjs authority-id DATABASE
  node runtime/src/cli.mjs authority-status DATABASE NAMESPACE_ID
  node runtime/src/cli.mjs authority-record DATABASE REQUEST_FILE ACTOR
  node runtime/src/cli.mjs authority-admission DATABASE REQUEST_FILE ACTOR

Local trusted-operator interface. Import emits a proposal; it never acquires a claim.
Demo uses fake workers only. No command publishes, merges, deploys or migrates tasks.`;

function fail(code, message = code) { throw Object.assign(new Error(message), { code }); }

function readWire(path) {
  const fd = openSync(path, constants.O_RDONLY | constants.O_NOFOLLOW);
  try {
    const stat = fstatSync(fd);
    if (!stat.isFile() || stat.size > 1_048_576) fail('INVALID_INPUT_FILE');
    const bytes = readFileSync(fd);
    if (bytes.length > 1_048_576) fail('INVALID_INPUT_FILE');
    return parseJSON(new TextDecoder('utf-8', { fatal: true }).decode(bytes));
  } finally { closeSync(fd); }
}

export async function runDemo(databasePath) {
  const directory = databasePath ? null : realpathSync(mkdtempSync(join(tmpdir(), 'foundry-demo-')));
  const path = databasePath ?? join(directory, 'authority.sqlite');
  const id = randomUUID();
  const at = new Date().toISOString();
  const later = (ms) => new Date(Date.parse(at) + ms).toISOString();
  const owner = { actor_id: 'demo-owner', role: 'owner' };
  const check = {
    check_id: 'artifact-integrity', check_version: '1', runner_id: 'demo-validator',
    run({ result }) {
      const valid = result.artifacts.every((artifact) => {
        const text = store.getArtifact(artifact.digest);
        return typeof text === 'string' && createHash('sha256').update(text, 'utf8').digest('hex') === artifact.digest;
      });
      return valid ? 'PASS' : 'FAIL';
    },
  };
  const workers = ['demo-explorer', 'demo-planner', 'demo-builder', 'demo-reviewer'];
  const store = openStore(path, {
    trusted_checks: [{ check_id: check.check_id, check_version: check.check_version, runner_id: check.runner_id }],
    worker_ids: workers, approver_ids: [owner.actor_id],
  });
  const adapter = new FakeAdapter();
  let sequence = 0;
  const executionId = `demo-${id}`;
  const send = (kind, payload) => store.execute({
    schema_version: 1, command_id: `${id}-${++sequence}`, execution_id: executionId,
    expected_revision: store.get(executionId)?.revision ?? 0, at, kind, payload,
  }, owner);
  const drain = async () => {
    const summary = await drainOutbox({ store, adapter, checks: [check], dispatcherId: `demo-${id}`, now: at });
    if (summary.unknown.length) fail('DEMO_DISPATCH_UNKNOWN');
  };
  try {
    if (store.list().length !== 0) fail('DEMO_DATABASE_NOT_EMPTY');
    const roles = ['explorer-researcher', 'architect-planner', 'builder-developer', 'reviewer-verifier'];
    const spec = {
      schema_version: 1, task_id: `DEMO-${id}`, spec_revision: 1,
      outcome: 'Simulate a four-stage read-only workflow; this is not a live task.',
      repository: { repository_id: 'example/demo', revision: '0'.repeat(40) },
      dependencies: [], write_scope: [], shared_resources: [], acceptance_checks: [check.check_id],
      workflow: { workflow_id: 'read-only-v1', version: 1, stages: roles.map((role, index) => ({
        id: `stage-${index}`, role, required_checks: [check.check_id], requires_approval: true,
      })) },
      retry_policy: { execution_retries: 1, rework_retries: 1, retry_delay_ms: 1000 },
      lease_ms: 60_000, validation_ms: 60_000,
      accountable_owner: owner.actor_id, integration_policy: 'read_only',
    };
    send('ADMIT', { spec });
    for (let stage = 0; stage < roles.length; stage += 1) {
      const previous = store.get(executionId).accepted_subject_digest;
      send('ASSIGN', {
        attempt_id: `attempt-${stage}`, worker_id: workers[stage], lease_deadline: later(60_000),
        workspace: 'demo-only', capabilities: ['read', 'analyze'],
        input_revisions: previous ? [previous] : [spec.repository.revision],
      });
      await drain();
      await drain();
      const current = store.get(executionId);
      if (current.state !== 'VALIDATING') fail('DEMO_STAGE_NOT_VALIDATING');
      send('RECORD_APPROVAL', { approval: {
        approval_id: `approval-${stage}`, actor_id: owner.actor_id, action: 'ACCEPT_STAGE',
        execution_id: executionId, spec_digest: current.spec_digest,
        workflow_id: spec.workflow.workflow_id, workflow_version: 1,
        subject_digest: current.result.subject_digest, decision: 'APPROVE',
        expires_at: later(60_000), revoked: false,
      } });
      send('ACCEPT_STAGE', { subject_digest: current.result.subject_digest });
      await drain();
    }
    const completed = store.get(executionId);
    const replay = store.replay();
    return {
      mode: 'SIMULATION', database: path, execution_id: executionId,
      state: completed.state, cleanup_pending: completed.cleanup_pending,
      stages: roles.length, events: store.events(executionId).length,
      replayed_executions: replay.executions.length,
    };
  } finally { store.close(); }
}

export async function main(argv = process.argv.slice(2)) {
  const [action, ...args] = argv;
  if (action === '--help' || action === 'help' || !action) return usage;
  if (action === 'demo' && args.length <= 1) return runDemo(args[0]);
  if (action === 'validate' && args.length === 2) {
    validate(args[0], readWire(args[1]));
    return { valid: true, type: args[0] };
  }
  if (action === 'import' && args.length === 1) return importReviewedTask(readWire(args[0]));
  if (action === 'authority-id' && args.length === 1) {
    const store = openStore(args[0]);
    try { return { authority_instance_id: store.instanceId() }; }
    finally { store.close(); }
  }
  if (action === 'authority-status' && args.length === 2) {
    const store = openStore(args[0]);
    try { return store.namespaceStatus(args[1]); }
    finally { store.close(); }
  }
  if (['authority-record', 'authority-admission'].includes(action) && args.length === 3) {
    const store = openStore(args[0]);
    try {
      const request = readWire(args[1]);
      const auth = { actor_id: args[2], role: 'owner' };
      return action === 'authority-record'
        ? store.recordNamespaceDecision(request, auth)
        : store.registerAdmission(request, auth);
    } finally { store.close(); }
  }
  if (action === 'command' && args.length === 5) {
    const store = openStore(args[0], readWire(args[1]));
    try { return store.execute(readWire(args[2]), { actor_id: args[3], role: args[4] }); }
    finally { store.close(); }
  }
  if (['show', 'export', 'replay'].includes(action) && args.length === (action === 'replay' ? 1 : 2)) {
    const store = openStore(args[0]);
    try {
      if (action === 'replay') return store.replay();
      const execution = store.get(args[1]);
      if (!execution) fail('EXECUTION_NOT_FOUND');
      return action === 'export' ? exportTaskView(execution) : execution;
    } finally { store.close(); }
  }
  fail('INVALID_ARGUMENT', usage);
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    const result = await main();
    process.stdout.write(typeof result === 'string' ? `${result}\n` : `${JSON.stringify(result, null, 2)}\n`);
  } catch (error) {
    process.stderr.write(`orchestration: ${error.code ?? 'ERROR'}\n`);
    process.exitCode = 1;
  }
}
