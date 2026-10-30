import { createHash } from 'node:crypto';
import { constants as fsConstants } from 'node:fs';
import { lstat, open, realpath } from 'node:fs/promises';
import { execFile } from 'node:child_process';
import path from 'node:path';
import { promisify } from 'node:util';
import { digest } from './contract.mjs';

const DEFAULT_OUTPUT_LIMIT = 32 * 1024;
const DEFAULT_INPUT_LIMIT = 1024 * 1024;
const DEFAULT_FILE_LIMIT = 64;
const MAX_OUTSTANDING_ARTIFACTS = 16;
const execFileAsync = promisify(execFile);

function fail(code, message) {
  throw Object.assign(new Error(message), { code });
}

function boundedInteger(value, name, maximum) {
  if (!Number.isSafeInteger(value) || value < 1 || value > maximum) {
    fail('INVALID_ARGUMENT', `${name} must be between 1 and ${maximum}`);
  }
}

function artifactResult({ attempt, spec, text }) {
  if (typeof text !== 'string') fail('INVALID_MODEL_OUTPUT', 'Worker output must be text');
  const bytes = Buffer.from(text, 'utf8');
  const resultArtifact = {
    path: 'worker-output.txt',
    digest: createHash('sha256').update(bytes).digest('hex'),
    repository_revision: spec.repository.revision,
  };
  const artifacts = [resultArtifact];
  return {
    schema_version: 1,
    attempt_id: attempt.attempt_id,
    fencing_token: attempt.fencing_token,
    spec_digest: digest(spec),
    kind: 'SUCCEEDED',
    artifacts,
    subject_digest: digest(artifacts),
  };
}

function createArtifactCapacity(limit = MAX_OUTSTANDING_ARTIFACTS) {
  let outstanding = 0;
  const waiters = [];

  function release() {
    const next = waiters.shift();
    if (next) {
      next.cleanup();
      next.resolve();
    } else {
      outstanding -= 1;
    }
  }

  return Object.freeze({
    async acquire(signal) {
      if (signal?.aborted) fail('ATTEMPT_ABORTED', 'Attempt was revoked before invocation');
      if (outstanding < limit) {
        outstanding += 1;
        return release;
      }
      await new Promise((resolve, reject) => {
        const waiter = {
          resolve,
          reject,
          cleanup: () => signal?.removeEventListener('abort', onAbort),
        };
        const onAbort = () => {
          const index = waiters.indexOf(waiter);
          if (index >= 0) waiters.splice(index, 1);
          waiter.cleanup();
          reject(Object.assign(new Error('Attempt was revoked while waiting for output capacity'), { code: 'ATTEMPT_ABORTED' }));
        };
        waiters.push(waiter);
        signal?.addEventListener('abort', onAbort, { once: true });
        if (signal?.aborted) onAbort();
      });
      return release;
    },
    release,
    get outstanding() { return outstanding; },
  });
}

function retainArtifact(artifacts, artifactDigest, content) {
  const current = artifacts.get(artifactDigest);
  if (current && current.content !== content) fail('ARTIFACT_DIGEST_COLLISION', 'One artifact digest maps to different content');
  if (current) current.references += 1;
  else artifacts.set(artifactDigest, { content, references: 1 });
}

function releaseArtifactReference(artifacts, capacity, artifactDigest) {
  const current = artifacts.get(artifactDigest);
  if (!current) return false;
  current.references -= 1;
  if (current.references === 0) artifacts.delete(artifactDigest);
  capacity.release();
  return true;
}

async function gitOutput(root, args) {
  const { stdout } = await execFileAsync('git', ['-C', root, ...args], {
    encoding: 'buffer',
    timeout: 2_000,
    maxBuffer: 2 * 1024 * 1024,
    env: {
      PATH: process.env.PATH ?? '/usr/bin:/bin',
      HOME: '/nonexistent',
      GIT_CONFIG_NOSYSTEM: '1',
      GIT_CONFIG_GLOBAL: '/dev/null',
      GIT_NO_REPLACE_OBJECTS: '1',
      GIT_OPTIONAL_LOCKS: '0',
      GIT_TERMINAL_PROMPT: '0',
      GIT_ALLOW_PROTOCOL: '',
    },
  });
  return stdout;
}

function freeze(value) {
  if (value && typeof value === 'object' && !Object.isFrozen(value)) {
    for (const child of Object.values(value)) freeze(child);
    Object.freeze(value);
  }
  return value;
}

function validateRelativePath(value) {
  if (typeof value !== 'string' || value.length < 1 || value.length > 512 || value.includes('\\')) {
    fail('INVALID_PATH', 'Workspace paths must be bounded, relative POSIX paths');
  }
  if (value.startsWith('/') || value.split('/').some((part) => part === '' || part === '.' || part === '..')) {
    fail('INVALID_PATH', 'Workspace path contains an unsafe component');
  }
  return value;
}

/** A bounded filesystem reader. It exposes no write, process, or network methods. */
export function createReadOnlyWorkspaceAdapter({
  root,
  allowedPaths = [],
  maxBytes = DEFAULT_INPUT_LIMIT,
  maxFiles = DEFAULT_FILE_LIMIT,
} = {}) {
  if (typeof root !== 'string' || root.length === 0) fail('INVALID_ARGUMENT', 'root is required');
  boundedInteger(maxBytes, 'maxBytes', 16 * 1024 * 1024);
  boundedInteger(maxFiles, 'maxFiles', 1024);
  if (!Array.isArray(allowedPaths) || allowedPaths.length > maxFiles) fail('INVALID_ARGUMENT', 'allowedPaths must be a bounded array');
  const allowed = new Set(allowedPaths.map(validateRelativePath));
  if (allowed.size !== allowedPaths.length) fail('INVALID_ARGUMENT', 'allowedPaths must not contain duplicates');
  let verifiedRevision = null;

  const ready = (async () => {
    const info = await lstat(root);
    if (info.isSymbolicLink() || !info.isDirectory()) fail('UNSAFE_WORKSPACE', 'Workspace root must be a real directory');
    return realpath(root);
  })();

  async function readOne(relative) {
    const candidate = validateRelativePath(relative);
    if (!allowed.has(candidate)) fail('PATH_NOT_ALLOWED', `Workspace path is not in the read allowlist: ${candidate}`);
    if (!verifiedRevision) fail('UNVERIFIED_BASE', 'Verify the workspace base revision before reading files');
    const base = await ready;
    const segments = candidate.split('/');
    let current = base;
    for (let index = 0; index < segments.length; index += 1) {
      current = path.join(current, segments[index]);
      const info = await lstat(current);
      if (info.isSymbolicLink()) fail('UNSAFE_WORKSPACE', 'Workspace reads do not follow symlinks');
      if (index < segments.length - 1 && !info.isDirectory()) fail('INVALID_PATH', 'A workspace path parent is not a directory');
      if (index === segments.length - 1 && !info.isFile()) fail('INVALID_PATH', 'Workspace path is not a regular file');
    }
    const flags = fsConstants.O_RDONLY | (fsConstants.O_NOFOLLOW ?? 0);
    const handle = await open(current, flags);
    try {
      const info = await handle.stat();
      if (!info.isFile() || info.size > maxBytes) fail('INPUT_TOO_LARGE', 'Workspace file exceeds the configured byte limit');
      const chunks = [];
      let total = 0;
      while (total <= maxBytes) {
        const chunk = Buffer.alloc(Math.min(64 * 1024, maxBytes + 1 - total));
        const { bytesRead } = await handle.read(chunk, 0, chunk.length, total);
        if (bytesRead === 0) break;
        chunks.push(chunk.subarray(0, bytesRead));
        total += bytesRead;
      }
      if (total > maxBytes) fail('INPUT_TOO_LARGE', 'Workspace file exceeds the configured byte limit');
      const currentRealPath = await realpath(current);
      if (currentRealPath !== path.join(base, candidate)) fail('UNSAFE_WORKSPACE', 'Workspace path changed during the read');
      const bytes = Buffer.concat(chunks, total);
      const committedBytes = await gitOutput(base, ['show', `${verifiedRevision}:${candidate}`]);
      if (!committedBytes.equals(bytes)) fail('WORKSPACE_DIRTY', `Workspace bytes differ from pinned base: ${candidate}`);
      let content;
      try {
        content = new TextDecoder('utf-8', { fatal: true }).decode(bytes);
      } catch {
        fail('INVALID_UTF8', 'Workspace text must be valid UTF-8');
      }
      return { path: candidate, content, digest: createHash('sha256').update(bytes).digest('hex') };
    } finally {
      await handle.close();
    }
  }

  return Object.freeze({
    async verifyBaseRevision(expectedRevision) {
      const base = await ready;
      if (typeof expectedRevision !== 'string' || !/^[a-f0-9]{7,64}$/.test(expectedRevision)) {
        fail('INVALID_REVISION', 'Expected repository revision must be a Git object ID');
      }
      const output = await gitOutput(base, ['rev-parse', '--verify', 'HEAD']);
      const actual = output.toString('utf8').trim();
      if (!/^[a-f0-9]{40,64}$/.test(actual) || !actual.startsWith(expectedRevision)) {
        fail('BASE_REVISION_MISMATCH', 'Workspace HEAD does not match the task base revision');
      }
      verifiedRevision = actual;
      return actual;
    },
    async readFiles(pathsToRead) {
      if (!Array.isArray(pathsToRead) || pathsToRead.length > maxFiles) fail('INVALID_ARGUMENT', 'Read request must be a bounded path array');
      if (new Set(pathsToRead).size !== pathsToRead.length) fail('INVALID_ARGUMENT', 'Read request contains duplicate paths');
      const files = [];
      let total = 0;
      for (const relative of pathsToRead) {
        const file = await readOne(relative);
        total += Buffer.byteLength(file.content, 'utf8');
        if (total > maxBytes) fail('INPUT_TOO_LARGE', 'Workspace input set exceeds the configured byte limit');
        files.push(file);
      }
      return files;
    },
  });
}

/**
 * Adapt a text-only invocation into a content-addressed result artifact.
 * The callback receives data only; this adapter starts no process and grants no tools.
 */
export function createTextModelAdapter({
  invoke,
  workerId,
  workspace = null,
  inputPaths = [],
  maxOutputBytes = DEFAULT_OUTPUT_LIMIT,
  maxInputBytes = DEFAULT_INPUT_LIMIT,
} = {}) {
  if (typeof invoke !== 'function') fail('INVALID_ARGUMENT', 'invoke must be a function');
  if (typeof workerId !== 'string' || workerId.length < 1 || workerId.length > 128) fail('INVALID_ARGUMENT', 'workerId is required and bounded');
  if (!Array.isArray(inputPaths) || inputPaths.length > DEFAULT_FILE_LIMIT || inputPaths.some((item) => typeof item !== 'string')) {
    fail('INVALID_ARGUMENT', 'inputPaths must be a bounded string array');
  }
  if (inputPaths.length && (!workspace || typeof workspace.verifyBaseRevision !== 'function' || typeof workspace.readFiles !== 'function')) {
    fail('INVALID_ARGUMENT', 'Input paths require a verified read-only workspace adapter');
  }
  boundedInteger(maxOutputBytes, 'maxOutputBytes', 1024 * 1024);
  boundedInteger(maxInputBytes, 'maxInputBytes', DEFAULT_INPUT_LIMIT);
  const artifacts = new Map();
  const pending = new Map();
  const ownedAttempts = new Map();
  const capacity = createArtifactCapacity();

  return Object.freeze({
    workerId,
    async runAttempt({ attempt, spec, inputSubjects = [] }) {
      if (ownedAttempts.has(attempt.attempt_id)) fail('DUPLICATE_ATTEMPT', 'Adapter already owns this attempt ID');
      const controller = new AbortController();
      const ownedAttempt = { controller, workerId: attempt.worker_id, status: 'preparing' };
      ownedAttempts.set(attempt.attempt_id, ownedAttempt);
      pending.set(attempt.attempt_id, { controller });
      let releaseCapacity;
      let retained = false;
      try {
        if (!Array.isArray(inputSubjects) || inputSubjects.length > DEFAULT_FILE_LIMIT) fail('INVALID_INPUT', 'Input subjects must be a bounded array');
        const files = inputPaths.length
          ? await (async () => {
            await workspace.verifyBaseRevision(spec.repository.revision);
            return workspace.readFiles([...inputPaths]);
          })()
          : [];
        if (!Array.isArray(files) || files.length > DEFAULT_FILE_LIMIT) fail('INVALID_INPUT', 'Workspace returned an invalid file list');
        let inputBytes = 0;
        for (const file of files) {
          if (typeof file?.path !== 'string' || typeof file.content !== 'string' || typeof file.digest !== 'string') {
            fail('INVALID_INPUT', 'Workspace returned a malformed file record');
          }
          const bytes = Buffer.from(file.content, 'utf8');
          inputBytes += bytes.length;
          if (inputBytes > maxInputBytes) fail('INPUT_TOO_LARGE', 'Model input exceeds the configured byte limit');
          const actualDigest = createHash('sha256').update(bytes).digest('hex');
          if (file.digest !== actualDigest) fail('INVALID_INPUT_DIGEST', `Workspace content digest does not match: ${file.path}`);
        }
        const checkedSubjects = [];
        for (const subject of inputSubjects) {
          if (typeof subject?.subject_digest !== 'string' || !Array.isArray(subject.artifacts) || subject.artifacts.length > DEFAULT_FILE_LIMIT) {
            fail('INVALID_INPUT', 'Input subject is malformed');
          }
          const manifest = [];
          for (const artifact of subject.artifacts) {
            if (typeof artifact?.path !== 'string' || typeof artifact.digest !== 'string'
              || typeof artifact.repository_revision !== 'string' || typeof artifact.content !== 'string') {
              fail('INVALID_INPUT', 'Input subject artifact is malformed');
            }
            const bytes = Buffer.from(artifact.content, 'utf8');
            inputBytes += bytes.length;
            if (inputBytes > maxInputBytes) fail('INPUT_TOO_LARGE', 'Combined model inputs exceed the configured byte limit');
            if (createHash('sha256').update(bytes).digest('hex') !== artifact.digest) {
              fail('INVALID_INPUT_DIGEST', `Input subject artifact digest does not match: ${artifact.path}`);
            }
            manifest.push({ path: artifact.path, digest: artifact.digest, repository_revision: artifact.repository_revision });
          }
          if (digest(manifest) !== subject.subject_digest) fail('INVALID_INPUT_SUBJECT', 'Input subject manifest does not match its digest');
          checkedSubjects.push({
            ...(typeof subject.task_id === 'string' ? { task_id: subject.task_id } : {}),
            ...(typeof subject.execution_id === 'string' ? { execution_id: subject.execution_id } : {}),
            subject_digest: subject.subject_digest,
            artifacts: subject.artifacts,
          });
        }
        if (controller.signal.aborted) fail('ATTEMPT_ABORTED', 'Attempt was revoked before invocation');
        const request = freeze(structuredClone({
          task: spec,
          attempt: {
            attempt_id: attempt.attempt_id,
            stage_id: attempt.stage_id,
            worker_id: attempt.worker_id,
            input_revisions: attempt.input_revisions,
          },
          files,
          inputSubjects: checkedSubjects,
        }));
        releaseCapacity = await capacity.acquire(controller.signal);
        if (controller.signal.aborted) fail('ATTEMPT_ABORTED', 'Attempt was revoked before invocation');
        ownedAttempt.status = 'invoking';
        const invocation = Promise.resolve().then(() => invoke(request, { signal: controller.signal }));
        pending.set(attempt.attempt_id, { controller, invocation });
        const output = await invocation;
        if (typeof output !== 'string') fail('INVALID_MODEL_OUTPUT', 'Text invocation must return a string');
        const bytes = Buffer.from(output, 'utf8');
        if (bytes.length > maxOutputBytes) fail('OUTPUT_TOO_LARGE', 'Worker output exceeds the configured byte limit');
        const result = artifactResult({ attempt, spec, text: output });
        retainArtifact(artifacts, result.artifacts[0].digest, output);
        retained = true;
        return result;
      } finally {
        pending.delete(attempt.attempt_id);
        ownedAttempt.status = 'settled';
        if (!retained) releaseCapacity?.();
      }
    },
    async getArtifact(artifactDigest) { return artifacts.get(artifactDigest)?.content ?? null; },
    releaseArtifact(artifactDigest) { return releaseArtifactReference(artifacts, capacity, artifactDigest); },
    async revokeCapabilities({ attempt_id, attempt }) {
      const owned = ownedAttempts.get(attempt_id);
      if (!owned || (attempt && owned.workerId !== attempt.worker_id)) {
        return { stopped: false, reason: 'This adapter cannot prove ownership of the attempt' };
      }
      const active = pending.get(attempt_id);
      if (!active && owned.status === 'settled') return { stopped: true, reason: 'Owned invocation has settled' };
      if (!active) return { stopped: false, reason: 'Attempt state is unknown' };
      active.controller.abort();
      await Promise.resolve();
      if (pending.has(attempt_id)) return { stopped: false, reason: 'Invocation has not settled after abort' };
      return { stopped: true, reason: 'Invocation settled after abort' };
    },
  });
}

/** A deterministic test adapter; it never launches a process or changes the workspace. */
export class FakeAdapter {
  constructor({ output = 'fake worker result', readFiles = async () => [], inputPaths = [] } = {}) {
    if (typeof output !== 'string' || typeof readFiles !== 'function' || !Array.isArray(inputPaths)) {
      fail('INVALID_ARGUMENT', 'FakeAdapter configuration is invalid');
    }
    this.output = output;
    this.readFiles = readFiles;
    this.inputPaths = [...inputPaths];
    this.activeAttempts = new Map();
    this.calls = [];
    this.artifacts = new Map();
    this.capacity = createArtifactCapacity();
  }

  async runAttempt({ attempt, spec }) {
    const releaseCapacity = await this.capacity.acquire();
    let retained = false;
    try {
    const files = await this.readFiles(this.inputPaths);
    // Include stable input digests so test results are tied to the bytes actually read.
    const inputSummary = files.map(({ path: filePath, digest: fileDigest }) => `${filePath}:${fileDigest}`).join('\n');
    const text = inputSummary ? `${this.output}\n${inputSummary}` : this.output;
    if (Buffer.byteLength(text, 'utf8') > 1024 * 1024) fail('OUTPUT_TOO_LARGE', 'Fake output exceeds the 1MiB artifact limit');
    const result = artifactResult({ attempt, spec, text });
    retainArtifact(this.artifacts, result.artifacts[0].digest, text);
    retained = true;
    this.activeAttempts.set(attempt.attempt_id, attempt.fencing_token);
    this.calls.push({ type: 'run', attempt_id: attempt.attempt_id, fencing_token: attempt.fencing_token });
    return result;
    } finally {
      if (!retained) releaseCapacity();
    }
  }

  async revokeCapabilities({ attempt_id, fencing_token }) {
    const current = this.activeAttempts.get(attempt_id);
    if (current !== undefined && current !== fencing_token) return { stopped: false, reason: 'Fencing token mismatch' };
    this.activeAttempts.delete(attempt_id);
    this.calls.push({ type: 'revoke', attempt_id, fencing_token });
    return { stopped: true, reason: 'Fake attempt stopped' };
  }

  async getArtifact(artifactDigest) {
    return this.artifacts.get(artifactDigest)?.content ?? null;
  }

  releaseArtifact(artifactDigest) {
    return releaseArtifactReference(this.artifacts, this.capacity, artifactDigest);
  }
}
