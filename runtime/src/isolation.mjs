import { createHash, randomUUID } from 'node:crypto';
import { spawn, execFile } from 'node:child_process';
import { promisify } from 'node:util';
import { canonical, digest, parseJSON } from './contract.mjs';
import { validateWorkspacePlan } from './workspace.mjs';

const execFileAsync = promisify(execFile);
const DEFAULT_IMAGE = 'node@sha256:8a32c8cf649a49915a66e81e7a609223c708d0adabf288ee3f4077d765542f85';
const DOCKER_TIMEOUT_MS = 15_000;
const RUN_TIMEOUT_MS = 120_000;
const PROVIDER_SETTLE_TIMEOUT_MS = 1_000;
const MAX_PLAN_BYTES = 1024 * 1024;
const MAX_STDOUT_BYTES = 2 * 1024 * 1024;
const MAX_PIDS = 32;
const MEMORY_BYTES = 268_435_456;
const CPU_NANOS = 1_000_000_000;
const PROBE_SCRIPT = String.raw`const fs=require('node:fs');const net=require('node:net');
if(process.getuid?.()!==1000||process.geteuid?.()!==1000)process.exit(31);
if(fs.existsSync('/var/run/docker.sock')||fs.existsSync('/root/.ssh')||fs.existsSync('/host-canary'))process.exit(32);
try{fs.writeFileSync('/foundry-rootfs-canary','x');process.exit(33)}catch{}
const socket=net.connect({host:'192.0.2.1',port:80});
socket.setTimeout(500,()=>{socket.destroy();process.stdout.write('FOUNDRY_PROBE_OK\\n')});
socket.on('connect',()=>process.exit(34));socket.on('error',()=>process.stdout.write('FOUNDRY_PROBE_OK\\n'));`;

const APPLY_SCRIPT = String.raw`
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
let input = '';
process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk) => {
  input += chunk;
  if (Buffer.byteLength(input, 'utf8') > 4194304) process.exit(21);
});
process.stdin.on('end', () => {
  try {
    const plan = JSON.parse(input);
    if (plan.schema_version !== 1 || !Array.isArray(plan.operations) || !Array.isArray(plan.write_scope)) throw new Error('invalid');
    const allowed = new Set(plan.write_scope);
    const seen = new Set();
    const rejectOverlaps = (paths) => {
      const ordered = [...paths].sort();
      for (let i = 0; i < ordered.length; i += 1) for (let j = i + 1; j < ordered.length; j += 1) {
        if (ordered[j].startsWith(ordered[i] + '/')) throw new Error('overlapping paths');
      }
    };
    rejectOverlaps(plan.write_scope);
    rejectOverlaps(plan.operations.map((operation) => operation.path));
    const digest = (bytes) => crypto.createHash('sha256').update(bytes).digest('hex');
    const pathFor = (value) => {
      if (typeof value !== 'string' || !value || value.length > 512 || value.includes('\\') || value.includes('\0') || value.startsWith('/') || value.split('/').some((part) => !part || part === '.' || part === '..')) throw new Error('path');
      const parts = value.split('/');
      if (parts.some((part) => ['.git','.codex','.agents','.aws','.project-memory','.ssh','.gnupg','.npmrc','.netrc'].includes(part) || /^\.env(?:\.|$)/i.test(part) || /^id_(?:rsa|ed25519)$/i.test(part)) || /\.(?:pem|key|p12|pfx|kdbx)$/i.test(parts.at(-1))) throw new Error('metadata');
      const target = path.resolve('/workspace', ...parts);
      if (!target.startsWith('/workspace/')) throw new Error('escape');
      return { target, parts };
    };
    const checkParents = (parts, create) => {
      let current = '/workspace';
      for (const part of parts.slice(0, -1)) {
        current = path.join(current, part);
        try {
          const st = fs.lstatSync(current);
          if (st.isSymbolicLink() || !st.isDirectory()) throw new Error('parent');
        } catch (error) {
          if (error.code !== 'ENOENT' || !create) throw error;
          fs.mkdirSync(current, { mode: 0o777 });
        }
      }
    };
    const results = [];
    for (const operation of plan.operations) {
      if (!allowed.has(operation.path) || seen.has(operation.path)) throw new Error('scope');
      seen.add(operation.path);
      const { target, parts } = pathFor(operation.path);
      checkParents(parts, operation.action === 'write');
      let current = null;
      try {
        const st = fs.lstatSync(target);
        if (st.isSymbolicLink() || !st.isFile()) throw new Error('target');
        current = fs.readFileSync(target);
      } catch (error) {
        if (error.code !== 'ENOENT') throw error;
      }
      if ((current === null ? null : digest(current)) !== operation.before_digest) throw new Error('digest');
      if (operation.action === 'write' && typeof operation.content === 'string' && !operation.content.includes('\0')) {
        const fd = fs.openSync(target, fs.constants.O_WRONLY | fs.constants.O_CREAT | fs.constants.O_TRUNC | (fs.constants.O_NOFOLLOW || 0), 0o666);
        try { fs.writeFileSync(fd, operation.content, 'utf8'); fs.fsyncSync(fd); } finally { fs.closeSync(fd); }
      } else if (operation.action === 'delete' && operation.content === null && current !== null) {
        fs.unlinkSync(target);
      } else {
        throw new Error('operation');
      }
      let after = null;
      if (operation.action === 'write') after = digest(fs.readFileSync(target));
      results.push({ path: operation.path, action: operation.action, after_digest: after });
    }
    process.stdout.write(JSON.stringify({ schema_version: 1, execution_id: plan.execution_id, spec_digest: plan.spec_digest, attempt_id: plan.attempt_id, repository_revision: plan.repository_revision, operations: results }) + '\n');
  } catch {
    process.stderr.write('workspace plan rejected\n');
    process.exitCode = 22;
  }
});`;

function fail(code, message) { throw Object.assign(new Error(message), { code }); }
function sha256(value) { return createHash('sha256').update(value).digest('hex'); }
function hashIdentity(value) { return sha256(String(value)).slice(0, 32); }

function imageParts(imageDigest) {
  const match = typeof imageDigest === 'string' && imageDigest.match(/^([a-zA-Z0-9._/:+-]+)@sha256:([a-f0-9]{64})$/);
  if (!match) fail('INVALID_IMAGE_DIGEST', 'Docker image must use a pinned sha256 digest');
  return { image: imageDigest, digest: match[2] };
}

function childEnv() {
  const env = {};
  for (const name of ['PATH', 'HOME', 'TMPDIR', 'LANG', 'LC_ALL']) if (typeof process.env[name] === 'string') env[name] = process.env[name];
  return env;
}

async function dockerExec(executable, args, { timeout = DOCKER_TIMEOUT_MS, maxBuffer = 2 * 1024 * 1024 } = {}) {
  try {
    const { stdout } = await execFileAsync(executable, args, { encoding: 'utf8', timeout, maxBuffer, env: childEnv() });
    return stdout;
  } catch {
    const operation = ({ info: 'INFO', image: 'IMAGE', create: 'CREATE', inspect: 'INSPECT', start: 'START', rm: 'CLEANUP', kill: 'KILL' })[args[0]] ?? 'CONTROL';
    fail(`ISOLATION_DOCKER_${operation}_FAILED`, 'Docker control operation failed');
  }
}

function labelSet({ executionId, attemptId, verificationId }) {
  return {
    'foundry.managed': 'true',
    'foundry.execution': hashIdentity(executionId),
    'foundry.attempt': hashIdentity(attemptId),
    'foundry.verification': hashIdentity(verificationId),
  };
}

function containerArgs({ imageDigest, workspaceRoot, labels, command }) {
  const args = [
    'create', '--interactive', '--network', 'none', '--read-only', '--cap-drop', 'ALL',
    '--security-opt', 'no-new-privileges:true', '--user', '1000:1000', '--pids-limit', String(MAX_PIDS),
    '--memory', String(MEMORY_BYTES), '--cpus', '1',
    '--mount', `type=bind,src=${workspaceRoot},dst=/workspace`,
  ];
  for (const [key, value] of Object.entries(labels)) args.push('--label', `${key}=${value}`);
  args.push(imageDigest, ...command);
  return args;
}

function validateInspect(inspect, { id, imageDigest, workspaceRoot, labels, imageEnvironment }) {
  const host = inspect?.HostConfig;
  const config = inspect?.Config;
  const mounts = inspect?.Mounts;
  if (inspect?.Id !== id || config?.Image !== imageDigest || config?.User !== '1000:1000'
    || host?.NetworkMode !== 'none' || host?.ReadonlyRootfs !== true || host?.Privileged !== false
    || host?.PidsLimit !== MAX_PIDS || host?.Memory !== MEMORY_BYTES || host?.NanoCpus !== CPU_NANOS
    || !Array.isArray(host?.CapDrop) || !host.CapDrop.includes('ALL') || (host.CapAdd?.length ?? 0) !== 0
    || !Array.isArray(host?.SecurityOpt) || !host.SecurityOpt.some((item) => item.startsWith('no-new-privileges'))
    || !Array.isArray(mounts) || mounts.length !== 1 || mounts[0].Type !== 'bind'
    || mounts[0].Source !== workspaceRoot || mounts[0].Destination !== '/workspace' || mounts[0].RW !== true
    || canonical(config?.Env ?? []) !== canonical(imageEnvironment) || (host?.Devices?.length ?? 0) !== 0) {
    fail('ISOLATION_CONFIG_MISMATCH', 'Docker container does not match the required isolation profile');
  }
  for (const [key, value] of Object.entries(labels)) if (config.Labels?.[key] !== value) fail('ISOLATION_IDENTITY_MISMATCH', 'Docker container labels do not match the attempt');
  return inspect.State;
}

function cleanupUnconfirmed(verificationId) {
  const error = Object.assign(new Error('Docker probe cleanup could not be confirmed'), {
    code: 'ISOLATION_CLEANUP_UNCONFIRMED',
    recovery_ref: `isolation-${verificationId}`,
  });
  return error;
}

function spawnAttached({ executable, args, input = '', timeoutMs, maxOutputBytes, spawnImpl, onAbort }) {
  return new Promise((resolve, reject) => {
    let child;
    const stdout = [];
    let total = 0;
    let reason;
    let settled = false;
    const timer = setTimeout(() => abort('ISOLATION_TIMEOUT'), timeoutMs);
    const done = (error, value) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      if (error) reject(error); else resolve(value);
    };
    const abort = (code) => {
      if (reason || settled) return;
      reason = code;
      void onAbort?.(code);
      try { child?.kill('SIGTERM'); } catch { /* Docker command close settles the call. */ }
    };
    try { child = spawnImpl(executable, args, { env: childEnv(), shell: false, stdio: ['pipe', 'pipe', 'pipe'] }); }
    catch { done(Object.assign(new Error('Docker process could not be started'), { code: 'ISOLATION_DOCKER_ERROR' })); return; }
    child.stdout.on('data', (chunk) => {
      if (reason || settled) return;
      const bytes = Buffer.from(chunk); total += bytes.length;
      if (total > maxOutputBytes) { abort('ISOLATION_OUTPUT_TOO_LARGE'); return; }
      stdout.push(bytes);
    });
    child.stdout.on('error', () => abort('ISOLATION_DOCKER_ERROR'));
    child.stderr.on('data', () => {}); child.stderr.on('error', () => {});
    child.on('error', () => done(Object.assign(new Error('Docker process could not be started'), { code: 'ISOLATION_DOCKER_ERROR' })));
    child.on('close', (code) => {
      if (reason) {
        const messages = { ISOLATION_TIMEOUT: 'Docker operation timed out', ATTEMPT_ABORTED: 'Docker attempt was cancelled', ISOLATION_OUTPUT_TOO_LARGE: 'Docker output exceeded its byte limit', ISOLATION_DOCKER_ERROR: 'Docker output stream failed' };
        done(Object.assign(new Error(messages[reason]), { code: reason }));
      } else if (code !== 0) done(Object.assign(new Error('Docker operation exited unsuccessfully'), { code: 'ISOLATION_CONTAINER_START_FAILED' }));
      else done(null, Buffer.concat(stdout, total).toString('utf8'));
    });
    child.stdin.on('error', () => {}); child.stdin.end(input);
  });
}

/** Docker host with an inspected, pinned, least-privilege container profile. */
export function createDockerIsolationHost({ dockerExecutable = 'docker', imageDigest = DEFAULT_IMAGE, spawnImpl = spawn } = {}) {
  if (typeof dockerExecutable !== 'string' || !dockerExecutable || dockerExecutable.includes('\0')) fail('INVALID_ARGUMENT', 'dockerExecutable is invalid');
  const image = imageParts(imageDigest);
  const attempts = new Map();
  let imageEnvironment = [];

  async function inspect(id) {
    let parsed;
    try { parsed = JSON.parse(await dockerExec(dockerExecutable, ['inspect', id])); }
    catch (error) { if (error?.code) throw error; fail('ISOLATION_DOCKER_ERROR', 'Docker inspect returned invalid JSON'); }
    if (!Array.isArray(parsed) || parsed.length !== 1) fail('ISOLATION_DOCKER_ERROR', 'Docker inspect returned an invalid record');
    return parsed[0];
  }
  async function createContainer({ workspaceRoot, executionId, attemptId, verificationId, command }) {
    const labels = labelSet({ executionId, attemptId, verificationId });
    const id = (await dockerExec(dockerExecutable, containerArgs({ imageDigest, workspaceRoot, labels, command }))).trim();
    if (!/^[a-f0-9]{64}$/.test(id)) fail('ISOLATION_OUTCOME_UNKNOWN', 'Docker did not return a valid container identity');
    return { id, labels };
  }
  async function inspectOwned(entry) {
    const state = await inspect(entry.id);
    validateInspect(state, { id: entry.id, imageDigest, workspaceRoot: entry.workspaceRoot, labels: labelSet(entry), imageEnvironment });
    return state;
  }
  async function stopOwned(entry) {
    try {
      await dockerExec(dockerExecutable, ['kill', entry.id]).catch(() => {});
      const state = await inspect(entry.id);
      if (state.Id !== entry.id || state.State?.Running || !['created', 'exited', 'dead'].includes(state.State?.Status)) return false;
      await dockerExec(dockerExecutable, ['rm', '-v', entry.id]);
      entry.stopped = true;
      return true;
    } catch { return false; }
  }
  async function verify({ workspaceRoot, imageDigest: requestedImage, executionId, attemptId }) {
    if (requestedImage !== imageDigest) fail('ISOLATION_IMAGE_MISMATCH', 'Workspace requested a different Docker image');
    const version = (await dockerExec(dockerExecutable, ['info', '--format', '{{.ServerVersion}}'])).trim();
    if (!/^\d+\.\d+/.test(version)) fail('ISOLATION_HOST_UNAVAILABLE', 'Docker server version could not be verified');
    let imageRecords;
    try { imageRecords = JSON.parse(await dockerExec(dockerExecutable, ['image', 'inspect', imageDigest])); }
    catch { fail('ISOLATION_IMAGE_UNAVAILABLE', 'Pinned Docker image is not available'); }
    if (!(imageRecords?.[0]?.RepoDigests ?? []).some((item) => item.endsWith(`@sha256:${image.digest}`))) fail('ISOLATION_IMAGE_MISMATCH', 'Docker image digest does not match the pin');
    imageEnvironment = imageRecords[0]?.Config?.Env ?? [];
    if (!Array.isArray(imageEnvironment) || imageEnvironment.some((item) => typeof item !== 'string')) fail('ISOLATION_IMAGE_MISMATCH', 'Pinned image environment is invalid');
    const verificationId = randomUUID();
    const probe = await createContainer({ workspaceRoot, executionId, attemptId, verificationId, command: ['node', '-e', PROBE_SCRIPT] });
    const entry = { ...probe, workspaceRoot, executionId, attemptId, verificationId, stopped: false };
    attempts.set(`probe:${verificationId}`, entry);
    try {
      const created = await inspectOwned(entry);
      if (created.State?.Running) fail('ISOLATION_PROBE_INVALID', 'New Docker probe container was already running');
      const probeOutput = await dockerExec(dockerExecutable, ['start', '--attach', probe.id]);
      const stopped = await inspectOwned(entry);
      if (stopped.State?.Running || stopped.State?.Status !== 'exited' || stopped.State?.ExitCode !== 0 || !probeOutput.includes('FOUNDRY_PROBE_OK')) fail('ISOLATION_PROBE_FAILED', 'Docker capability probe did not confirm the required isolation');
      await dockerExec(dockerExecutable, ['rm', '-v', probe.id]);
      entry.stopped = true;
      attempts.delete(`probe:${verificationId}`);
      return { verified: true, verificationId, imageDigest, dockerServerVersion: version };
    } catch (error) {
      if (await stopOwned(entry)) {
        attempts.delete(`probe:${verificationId}`);
        throw error;
      }
      throw cleanupUnconfirmed(verificationId);
    }
  }
  async function applyPlan({ workspaceRoot, workspaceId, imageDigest: requestedImage, plan, guard, verificationId }) {
    validateWorkspacePlan(plan);
    if (requestedImage !== imageDigest || guard?.verified !== true || guard.workspace !== workspaceId
      || guard.execution_id !== plan.execution_id || guard.spec_digest !== plan.spec_digest || guard.attempt_id !== plan.attempt_id
      || guard.repository_revision !== plan.repository_revision || !Array.isArray(guard.write_scope)) fail('ISOLATION_GUARD_MISMATCH', 'Docker plan does not match its verified attempt');
    const container = await createContainer({ workspaceRoot, executionId: plan.execution_id, attemptId: plan.attempt_id, verificationId, command: ['node', '-e', APPLY_SCRIPT] });
    const entry = { ...container, workspaceRoot, executionId: plan.execution_id, attemptId: plan.attempt_id, verificationId, stopped: false };
    attempts.set(plan.attempt_id, entry);
    await inspectOwned(entry);
    let output; let startError;
    try {
      output = await spawnAttached({
        executable: dockerExecutable, args: ['start', '--attach', '--interactive', container.id], input: canonical(plan),
        timeoutMs: RUN_TIMEOUT_MS, maxOutputBytes: MAX_STDOUT_BYTES, spawnImpl,
        onAbort: () => dockerExec(dockerExecutable, ['kill', container.id]).catch(() => {}),
      });
    } catch (error) { startError = error; }
    let stopped;
    try { stopped = await inspectOwned(entry); } catch { fail('ISOLATION_OUTCOME_UNKNOWN', 'Could not verify whether the Docker attempt stopped'); }
    if (stopped.State?.Running || !['exited', 'dead'].includes(stopped.State?.Status)) fail('ISOLATION_OUTCOME_UNKNOWN', 'Docker attempt may still be running');
    await dockerExec(dockerExecutable, ['rm', '-v', container.id]);
    entry.stopped = true;
    attempts.set(plan.attempt_id, entry);
    if (startError) throw startError;
    if (stopped.State?.ExitCode !== 0) fail('ISOLATION_PLAN_REJECTED', 'Docker rejected the workspace plan');
    let result;
    try { result = JSON.parse(output.trim()); } catch { fail('ISOLATION_INVALID_OUTPUT', 'Docker returned invalid workspace evidence'); }
    if (result?.schema_version !== 1 || result.execution_id !== plan.execution_id || result.spec_digest !== plan.spec_digest
      || result.attempt_id !== plan.attempt_id || result.repository_revision !== plan.repository_revision
      || !Array.isArray(result.operations) || result.operations.length !== plan.operations.length) fail('ISOLATION_RESULT_MISMATCH', 'Docker workspace evidence does not match the attempt');
    return { verified: true, ...result };
  }
  async function revokeCapabilities({ executionId, attemptId, fencingToken, verificationId }) {
    const entry = attempts.get(attemptId);
    if (!entry || entry.executionId !== executionId || entry.verificationId !== verificationId || !Number.isSafeInteger(fencingToken)) return { stopped: false };
    if (entry.stopped) return { stopped: true, container_id: entry.id };
    try {
      await dockerExec(dockerExecutable, ['kill', entry.id]).catch(() => {});
      const state = await inspectOwned(entry);
      if (state.State?.Running || !['exited', 'dead'].includes(state.State?.Status)) return { stopped: false };
      await dockerExec(dockerExecutable, ['rm', '-v', entry.id]);
      entry.stopped = true;
      return { stopped: true, container_id: entry.id };
    } catch { return { stopped: false }; }
  }
  async function disposeWorkspace({ workspaceRoot }) {
    for (const entry of attempts.values()) if (entry.workspaceRoot === workspaceRoot && !entry.stopped) return false;
    return true;
  }
  return Object.freeze({ verify, applyPlan, revokeCapabilities, disposeWorkspace, imageDigest });
}

function freeze(value) {
  if (value && typeof value === 'object' && !Object.isFrozen(value)) {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}

/** Adapter compatible with runAttempt/getArtifact/releaseArtifact/revokeCapabilities. */
export function createIsolatedAdapter({ workerId, invoke, workspace, inputPaths = [] } = {}) {
  if (typeof workerId !== 'string' || !workerId || workerId.length > 128) fail('INVALID_ARGUMENT', 'workerId is required and bounded');
  if (typeof invoke !== 'function') fail('INVALID_ARGUMENT', 'invoke must be a function');
  if (!workspace || typeof workspace.guard !== 'function' || typeof workspace.readFiles !== 'function' || typeof workspace.applyPlan !== 'function'
    || typeof workspace.getArtifact !== 'function' || typeof workspace.revokeCapabilities !== 'function') fail('INVALID_ARGUMENT', 'workspace is not a verified candidate workspace');
  if (!Array.isArray(inputPaths) || inputPaths.length > 256 || new Set(inputPaths).size !== inputPaths.length) fail('INVALID_ARGUMENT', 'inputPaths must be a bounded unique array');
  const artifacts = new Map();
  const owned = new Map();
  return Object.freeze({
    workerId,
    async runAttempt({ attempt, spec, inputSubjects = [] }) {
      if (!attempt || attempt.worker_id !== workerId || attempt.attempt_id !== workspace.attempt_id
        || attempt.workspace !== workspace.workspace || !Array.isArray(attempt.capabilities) || !attempt.capabilities.includes('write')) {
        fail('ATTEMPT_IDENTITY_MISMATCH', 'Assigned attempt does not match the isolated workspace or write capability');
      }
      const stage = spec.workflow?.stages?.find((item) => item.id === attempt.stage_id);
      if (spec.integration_policy !== 'isolated_write' || stage?.role !== 'builder-developer') fail('UNSUPPORTED_CAPABILITY', 'Only managed builder stages can use isolated write workspaces');
      if (!Array.isArray(spec.write_scope) || !Array.isArray(workspace.write_scope)
        || canonical([...spec.write_scope].sort()) !== canonical([...workspace.write_scope].sort())) {
        fail('WRITE_SCOPE_MISMATCH', 'Task write scope does not match the prepared candidate workspace');
      }
      if (digest(spec) !== workspace.spec_digest || spec.repository.revision !== workspace.repository_revision) fail('SPEC_IDENTITY_MISMATCH', 'Task specification does not match the isolated workspace');
      if (owned.has(attempt.attempt_id)) fail('DUPLICATE_ATTEMPT', 'Isolated adapter already owns this attempt');
      const proof = workspace.guard({ verified: true, execution_id: workspace.execution_id, spec_digest: workspace.spec_digest,
        attempt_id: attempt.attempt_id, workspace: workspace.workspace, repository_revision: spec.repository.revision, write_scope: workspace.write_scope });
      const record = { status: 'invoking', fencingToken: attempt.fencing_token, controller: new AbortController(), invocation: null };
      owned.set(attempt.attempt_id, record);
      let dockerInvocationStarted = false;
      try {
        const files = await workspace.readFiles(inputPaths);
        if (record.controller.signal.aborted) fail('ATTEMPT_ABORTED', 'Attempt was revoked during input preparation');
        const request = freeze(structuredClone({ task: spec,
          attempt: { attempt_id: attempt.attempt_id, stage_id: attempt.stage_id, worker_id: attempt.worker_id, input_revisions: attempt.input_revisions },
          plan_identity: { schema_version: 1, execution_id: workspace.execution_id, spec_digest: digest(spec),
            attempt_id: attempt.attempt_id, repository_revision: spec.repository.revision, write_scope: workspace.write_scope },
          worker_contract: { output: 'one JSON WorkspacePlan object only', operation_fields: ['path', 'action', 'before_digest', 'content'],
            actions: ['write', 'delete'], before_digest: 'SHA-256 of the pinned current file, or null if absent',
            write_scope: 'Use exactly plan_identity.write_scope; do not invent IDs or revisions.',
            before_digests: workspace.base_digests },
          files, inputSubjects }));
        record.invocation = Promise.resolve().then(() => {
          if (record.controller.signal.aborted) fail('ATTEMPT_ABORTED', 'Attempt was revoked before provider invocation');
          return invoke(request, { signal: record.controller.signal });
        });
        const output = await record.invocation;
        if (record.controller.signal.aborted) fail('ATTEMPT_ABORTED', 'Attempt was revoked before workspace execution');
        if (typeof output !== 'string' || Buffer.byteLength(output, 'utf8') > MAX_PLAN_BYTES) fail('INVALID_WORKSPACE_PLAN', 'Worker must return a bounded JSON file plan');
        let plan;
        try { plan = parseJSON(output, { maxBytes: MAX_PLAN_BYTES, maxDepth: 24 }); }
        catch { fail('INVALID_WORKSPACE_PLAN', 'Worker returned invalid or duplicate-key JSON'); }
        validateWorkspacePlan(plan);
        const content = canonical(plan);
        if (Buffer.byteLength(content, 'utf8') > MAX_PLAN_BYTES) fail('INVALID_WORKSPACE_PLAN', 'Canonical workspace plan exceeds its artifact limit');
        dockerInvocationStarted = true;
        record.status = 'running';
        const summary = await workspace.applyPlan(plan, proof);
        const artifactDigest = sha256(Buffer.from(content, 'utf8'));
        const manifest = [{ path: 'workspace-plan.json', digest: artifactDigest, repository_revision: spec.repository.revision }];
        const result = { schema_version: 1, attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token,
          spec_digest: digest(spec), kind: 'SUCCEEDED', artifacts: manifest, subject_digest: digest(manifest) };
        artifacts.set(artifactDigest, content);
        workspace.holdArtifact?.(artifactDigest, content);
        if (summary.operations.some((item, index) => item.path !== plan.operations[index].path || item.action !== plan.operations[index].action)) fail('ISOLATION_RESULT_MISMATCH', 'Docker applied a different set of workspace operations');
        owned.get(attempt.attempt_id).status = 'settled';
        return result;
      } catch (error) {
        if (record.status !== 'settled') record.status = dockerInvocationStarted ? 'unknown' : 'settled';
        throw error;
      }
    },
    async getArtifact(artifactDigest) { return artifacts.get(artifactDigest) ?? null; },
    releaseArtifact(artifactDigest) { return artifacts.delete(artifactDigest) && (workspace.releaseArtifact?.(artifactDigest) ?? true); },
    async revokeCapabilities({ attempt_id, fencing_token, attempt } = {}) {
      const record = owned.get(attempt_id);
      if (!record || attempt?.attempt_id !== attempt_id || fencing_token !== record.fencingToken) return { stopped: false };
      if (record.status === 'settled') return { stopped: true };
      if (record.status === 'invoking') {
        record.controller.abort();
        if (record.invocation) {
          let timer;
          const settled = await Promise.race([record.invocation.then(() => true, () => true),
            new Promise((resolve) => { timer = setTimeout(() => resolve(false), PROVIDER_SETTLE_TIMEOUT_MS); })]);
          clearTimeout(timer);
          if (!settled) return { stopped: false };
        }
        if (record.status === 'invoking') return { stopped: false };
        if (record.status === 'settled') return { stopped: true };
      }
      const result = await workspace.revokeCapabilities({ attempt_id, fencing_token, attempt });
      if (result?.stopped === true) record.status = 'settled';
      return result;
    },
  });
}

export const isolationDefaults = Object.freeze({ imageDigest: DEFAULT_IMAGE, maxPlanBytes: MAX_PLAN_BYTES });
