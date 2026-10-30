import { createHash, randomUUID } from 'node:crypto';
import { execFile } from 'node:child_process';
import { mkdir, mkdtemp, chmod, lstat, open, realpath, rm, writeFile, rename } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { promisify } from 'node:util';
import Ajv2020 from 'ajv/dist/2020.js';
import workspaceSchema from '../contracts/workspace.schema.json' with { type: 'json' };
import { canonical, digest } from './contract.mjs';

const execFileAsync = promisify(execFile);
const DEFAULT_IMAGE = 'node@sha256:8a32c8cf649a49915a66e81e7a609223c708d0adabf288ee3f4077d765542f85';
const DEFAULT_TOTAL_BYTES = 32 * 1024 * 1024;
const ajv = new Ajv2020({ allErrors: true, strict: true, strictTypes: true, coerceTypes: false, useDefaults: false, removeAdditional: false });
const validateSchema = ajv.compile(workspaceSchema);

function fail(code, message) {
  throw Object.assign(new Error(message), { code });
}

function sha256(bytes) {
  return createHash('sha256').update(bytes).digest('hex');
}

function validatePath(value) {
  if (typeof value !== 'string' || value.length < 1 || value.length > 512 || value.includes('\\') || value.includes('\0')
    || value.startsWith('/') || value.split('/').some((part) => part === '' || part === '.' || part === '..')) {
    fail('INVALID_WORKSPACE_PATH', 'Workspace paths must be bounded relative POSIX paths');
  }
  const parts = value.split('/');
  const blocked = new Set(['.git', '.codex', '.agents', '.aws', '.project-memory', '.ssh', '.gnupg', '.npmrc', '.netrc']);
  if (parts.some((part) => blocked.has(part) || /^\.env(?:\.|$)/i.test(part) || /^id_(?:rsa|ed25519)$/i.test(part))) {
    fail('WORKSPACE_PATH_BLOCKED', 'Workspace path names protected metadata or credential material');
  }
  if (/\.(?:pem|key|p12|pfx|kdbx)$/i.test(parts.at(-1))) fail('WORKSPACE_PATH_BLOCKED', 'Workspace path names protected key material');
  return value;
}

function sortedUniquePaths(paths, name, { allowEmpty = false, maximum = 256 } = {}) {
  if (!Array.isArray(paths) || paths.length > maximum || (!allowEmpty && paths.length === 0)) fail('INVALID_ARGUMENT', `${name} must be a bounded path array`);
  const result = paths.map(validatePath);
  if (new Set(result).size !== result.length) fail('INVALID_ARGUMENT', `${name} contains duplicate paths`);
  for (let left = 0; left < result.length; left += 1) {
    for (let right = left + 1; right < result.length; right += 1) {
      if (result[right].startsWith(`${result[left]}/`) || result[left].startsWith(`${result[right]}/`)) {
        fail('WORKSPACE_PATH_CONFLICT', `${name} cannot contain ancestor and descendant paths`);
      }
    }
  }
  return [...result].sort();
}

async function gitOutput(root, args, maxBuffer = 16 * 1024 * 1024) {
  try {
    const { stdout } = await execFileAsync('git', ['-C', root, ...args], {
      encoding: 'buffer', timeout: 10_000, maxBuffer,
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
  } catch {
    fail('WORKSPACE_GIT_ERROR', 'Could not read the pinned repository objects');
  }
}

function parseTree(treeBytes) {
  const records = new Map();
  for (const record of treeBytes.toString('utf8').split('\0')) {
    if (!record) continue;
    const tab = record.indexOf('\t');
    if (tab < 0) fail('WORKSPACE_GIT_ERROR', 'Pinned repository tree is malformed');
    const [mode, type, oid] = record.slice(0, tab).split(' ');
    const name = record.slice(tab + 1);
    records.set(name, { mode, type, oid });
  }
  return records;
}

function assertSafeTreePath(filePath, records, { mustExist }) {
  const parts = filePath.split('/');
  let current = '';
  for (let index = 0; index < parts.length; index += 1) {
    current = current ? `${current}/${parts[index]}` : parts[index];
    const entry = records.get(current);
    if (index < parts.length - 1) {
      if (entry && (entry.type !== 'tree' || entry.mode !== '040000')) fail('WORKSPACE_PATH_CONFLICT', 'Workspace path traverses a non-directory repository entry');
      continue;
    }
    if (!entry) {
      if (mustExist) fail('WORKSPACE_INPUT_MISSING', `Pinned input is missing: ${filePath}`);
      if ([...records.keys()].some((candidate) => candidate.startsWith(`${filePath}/`))) fail('WORKSPACE_PATH_CONFLICT', 'Workspace output conflicts with an existing directory');
      return null;
    }
    if (entry.type !== 'blob' || !['100644', '100755'].includes(entry.mode)) fail('WORKSPACE_SYMLINK_OR_SPECIAL_FILE', 'Workspace inputs and targets must be regular files');
    return entry;
  }
  return null;
}

async function mkdirSafe(root, relativeDirectory) {
  let current = root;
  for (const part of relativeDirectory.split('/').filter(Boolean)) {
    current = path.join(current, part);
    try {
      await mkdir(current, { mode: 0o777 });
    } catch (error) {
      if (error.code !== 'EEXIST') throw error;
    }
    const info = await lstat(current);
    if (info.isSymbolicLink() || !info.isDirectory()) fail('UNSAFE_WORKSPACE', 'Workspace parent is not a real directory');
    await chmod(current, 0o777);
  }
}

function validatePlan(plan) {
  if (!validateSchema(plan)) fail('INVALID_WORKSPACE_PLAN', 'Workspace plan does not match its schema');
  sortedUniquePaths(plan.write_scope, 'plan write_scope');
  const paths = plan.operations.map((operation) => validatePath(operation.path));
  if (new Set(paths).size !== paths.length) fail('INVALID_WORKSPACE_PLAN', 'Workspace plan repeats a path');
  for (let left = 0; left < paths.length; left += 1) {
    for (let right = left + 1; right < paths.length; right += 1) {
      if (paths[right].startsWith(`${paths[left]}/`) || paths[left].startsWith(`${paths[right]}/`)) {
        fail('INVALID_WORKSPACE_PLAN', 'Workspace plan cannot contain ancestor and descendant paths');
      }
    }
  }
  for (const operation of plan.operations) {
    if (operation.content !== null && Buffer.byteLength(operation.content, 'utf8') > 1024 * 1024) fail('INVALID_WORKSPACE_PLAN', 'A file operation exceeds the 1MiB content limit');
    if (operation.content?.includes('\0')) fail('INVALID_WORKSPACE_PLAN', 'Workspace text cannot contain NUL bytes');
  }
  return plan;
}

function exactScope(left, right) {
  return left.length === right.length && left.every((item, index) => item === right[index]);
}

async function readRegularFile(root, relative) {
  const target = path.join(root, ...relative.split('/'));
  const segments = relative.split('/');
  let current = root;
  for (let index = 0; index < segments.length; index += 1) {
    current = path.join(current, segments[index]);
    let info;
    try { info = await lstat(current); } catch (error) {
      if (error.code === 'ENOENT') return null;
      throw error;
    }
    if (info.isSymbolicLink() || (index < segments.length - 1 && !info.isDirectory()) || (index === segments.length - 1 && !info.isFile())) {
      fail('UNSAFE_WORKSPACE', 'Workspace path contains a symlink or non-regular file');
    }
  }
  const handle = await open(target, 'r');
  try { return await handle.readFile(); } finally { await handle.close(); }
}

/** Snapshot only explicitly allowed Git blobs into a private candidate directory. */
export async function createCandidateWorkspace({
  repositoryRoot,
  repositoryRevision,
  executionId,
  specDigest,
  attemptId,
  writeScope,
  readPaths = [],
  imageDigest = DEFAULT_IMAGE,
  host,
  maxFiles = 256,
  maxBytes = DEFAULT_TOTAL_BYTES,
} = {}) {
  if (typeof repositoryRoot !== 'string' || !path.isAbsolute(repositoryRoot)) fail('INVALID_ARGUMENT', 'repositoryRoot must be absolute');
  if (!/^[a-f0-9]{40,64}$/.test(repositoryRevision ?? '')) fail('INVALID_ARGUMENT', 'repositoryRevision must be a full Git object ID');
  if (typeof executionId !== 'string' || !executionId || executionId.length > 128) fail('INVALID_ARGUMENT', 'executionId is required');
  if (typeof attemptId !== 'string' || !attemptId || attemptId.length > 128) fail('INVALID_ARGUMENT', 'attemptId is required');
  if (!/^[a-f0-9]{64}$/.test(specDigest ?? '')) fail('INVALID_ARGUMENT', 'specDigest must be a SHA-256 digest');
  const scope = sortedUniquePaths(writeScope, 'writeScope', { maximum: maxFiles });
  const reads = sortedUniquePaths(readPaths, 'readPaths', { allowEmpty: true, maximum: maxFiles });
  if (reads.length + scope.length > maxFiles) fail('INVALID_ARGUMENT', 'Combined workspace paths exceed the configured file limit');
  if (!host || typeof host.verify !== 'function' || typeof host.applyPlan !== 'function' || typeof host.revokeCapabilities !== 'function') {
    fail('INVALID_ARGUMENT', 'host must implement verify, applyPlan and revokeCapabilities');
  }
  if (!Number.isSafeInteger(maxBytes) || maxBytes < 1 || maxBytes > 128 * 1024 * 1024) fail('INVALID_ARGUMENT', 'maxBytes is outside the supported range');

  const rootInfo = await lstat(repositoryRoot);
  if (rootInfo.isSymbolicLink() || !rootInfo.isDirectory()) fail('UNSAFE_REPOSITORY', 'repositoryRoot must be a real directory');
  const sourceRoot = await realpath(repositoryRoot);
  const actualRevision = (await gitOutput(sourceRoot, ['rev-parse', '--verify', 'HEAD^{commit}'])).toString('utf8').trim();
  if (actualRevision !== repositoryRevision) fail('BASE_REVISION_MISMATCH', 'Repository HEAD does not match the pinned revision');
  const tree = parseTree(await gitOutput(sourceRoot, ['ls-tree', '-r', '-z', '--full-tree', repositoryRevision], maxBytes));
  const all = [...new Set([...reads, ...scope])].sort();
  for (let left = 0; left < all.length; left += 1) {
    for (let right = left + 1; right < all.length; right += 1) {
      if (all[right].startsWith(`${all[left]}/`)) fail('WORKSPACE_PATH_CONFLICT', 'Allowed paths cannot contain ancestor and descendant paths');
    }
  }
  const blobs = new Map();
  let total = 0;
  for (const filePath of all) {
    const entry = assertSafeTreePath(filePath, tree, { mustExist: reads.includes(filePath) });
    if (!entry) continue;
    const bytes = await gitOutput(sourceRoot, ['cat-file', 'blob', entry.oid], maxBytes);
    total += bytes.length;
    if (total > maxBytes) fail('WORKSPACE_TOO_LARGE', 'Allowed workspace snapshot exceeds its byte limit');
    blobs.set(filePath, bytes);
  }

  const tempRoot = process.platform === 'darwin' ? '/private/tmp' : os.tmpdir();
  const parent = await mkdtemp(path.join(tempRoot, 'foundry-candidate-'));
  await chmod(parent, 0o700);
  const workspaceRoot = path.join(parent, 'workspace');
  let hostProof;
  try {
    await mkdir(workspaceRoot, { mode: 0o777 });
    await chmod(workspaceRoot, 0o777);
    for (const [filePath, bytes] of blobs) {
      await mkdirSafe(workspaceRoot, path.posix.dirname(filePath));
      const target = path.join(workspaceRoot, ...filePath.split('/'));
      const handle = await open(target, 'wx', 0o666);
      try { await handle.writeFile(bytes); } finally { await handle.close(); }
      await chmod(target, 0o666);
    }
    hostProof = await host.verify({ workspaceRoot, imageDigest, executionId, attemptId });
    if (hostProof?.verified !== true || typeof hostProof.verificationId !== 'string') fail('ISOLATION_UNVERIFIED', 'Docker host did not return a verified capability proof');
  } catch (error) {
    if (error?.code !== 'ISOLATION_CLEANUP_UNCONFIRMED') await rm(parent, { recursive: true, force: true });
    throw error;
  }

  const identity = Object.freeze({
    verified: true,
    execution_id: executionId,
    spec_digest: specDigest,
    attempt_id: attemptId,
    workspace: `workspace-${randomUUID()}`,
    repository_revision: repositoryRevision,
    write_scope: scope,
  });
  const baseDigests = Object.freeze(Object.fromEntries(scope.map((filePath) => [
    filePath, blobs.has(filePath) ? sha256(blobs.get(filePath)) : null,
  ])));
  const artifacts = new Map();
  let disposed = false;
  const workspace = {
    ...identity,
    image_digest: imageDigest,
    base_digests: baseDigests,
    verificationId: hostProof.verificationId,
    guard(proof) {
      if (disposed || proof?.verified !== true || proof.execution_id !== executionId || proof.spec_digest !== specDigest
        || proof.attempt_id !== attemptId || proof.workspace !== identity.workspace || proof.repository_revision !== repositoryRevision
        || !Array.isArray(proof.write_scope) || !exactScope([...proof.write_scope].sort(), scope)) {
        fail('ISOLATION_GUARD_MISMATCH', 'Workspace proof does not match this execution, attempt, base and write scope');
      }
      return identity;
    },
    async readFiles(paths) {
      if (disposed || !Array.isArray(paths) || paths.length > reads.length || paths.some((item) => !reads.includes(item))) fail('PATH_NOT_ALLOWED', 'Workspace read is outside the configured input scope');
      const output = [];
      for (const filePath of paths) {
        const bytes = blobs.get(filePath);
        if (!bytes) fail('WORKSPACE_INPUT_MISSING', `Pinned input is missing: ${filePath}`);
        output.push({ path: filePath, content: new TextDecoder('utf-8', { fatal: true }).decode(bytes), digest: sha256(bytes) });
      }
      return output;
    },
    async applyPlan(plan, proof) {
      workspace.guard(proof);
      validatePlan(plan);
      if (plan.execution_id !== executionId || plan.spec_digest !== specDigest || plan.attempt_id !== attemptId
        || plan.repository_revision !== repositoryRevision || !exactScope([...plan.write_scope].sort(), scope)) {
        fail('WORKSPACE_PLAN_IDENTITY_MISMATCH', 'Workspace plan identity does not match its verified attempt');
      }
      const operationPaths = plan.operations.map((item) => item.path).sort();
      if (operationPaths.some((filePath) => !scope.includes(filePath))) fail('WRITE_SCOPE_VIOLATION', 'Workspace plan writes outside its exact scope');
      if (new Set(operationPaths).size !== operationPaths.length) fail('INVALID_WORKSPACE_PLAN', 'Workspace plan repeats a path');
      for (const operation of plan.operations) {
        const prior = await readRegularFile(workspaceRoot, operation.path);
        const beforeDigest = prior === null ? null : sha256(prior);
        if (beforeDigest !== operation.before_digest) fail('WORKSPACE_BASE_CHANGED', `Candidate bytes changed before applying ${operation.path}`);
      }
      const summary = await host.applyPlan({ workspaceRoot, workspaceId: identity.workspace, imageDigest, plan, guard: identity, verificationId: workspace.verificationId });
      if (summary?.verified !== true || summary.execution_id !== executionId || summary.attempt_id !== attemptId
        || summary.spec_digest !== specDigest || summary.repository_revision !== repositoryRevision
        || !Array.isArray(summary.operations) || summary.operations.length !== plan.operations.length) {
        fail('ISOLATION_RESULT_MISMATCH', 'Docker host returned an unverified workspace result');
      }
      for (let index = 0; index < plan.operations.length; index += 1) {
        const operation = plan.operations[index];
        const expectedDigest = operation.action === 'write' ? sha256(Buffer.from(operation.content, 'utf8')) : null;
        const result = summary.operations[index];
        if (result?.path !== operation.path || result.action !== operation.action || result.after_digest !== expectedDigest) {
          fail('ISOLATION_RESULT_MISMATCH', 'Docker host result does not match the authorized file plan');
        }
      }
      return summary;
    },
    holdArtifact(hash, content) { artifacts.set(hash, content); },
    async getArtifact(hash) { return artifacts.get(hash) ?? null; },
    releaseArtifact(hash) { return artifacts.delete(hash); },
    async promote({ plan, proof, ownerId, expectedOwner, acceptedSubjectDigest, targetRoot }) {
      workspace.guard(proof);
      validatePlan(plan);
      if (!ownerId || ownerId !== expectedOwner) fail('PROMOTION_OWNER_REQUIRED', 'Only the accountable owner may promote a workspace plan');
      if (plan.execution_id !== executionId || plan.spec_digest !== specDigest || plan.attempt_id !== attemptId
        || plan.repository_revision !== repositoryRevision || !exactScope([...plan.write_scope].sort(), scope)
        || plan.operations.some((operation) => !scope.includes(operation.path))) {
        fail('PROMOTION_PLAN_MISMATCH', 'Promotion plan does not match its verified execution and write scope');
      }
      if (acceptedSubjectDigest !== digest([{ path: 'workspace-plan.json', digest: sha256(Buffer.from(canonical(plan))), repository_revision: repositoryRevision }])) {
        fail('PROMOTION_SUBJECT_MISMATCH', 'Accepted subject does not match the candidate workspace plan');
      }
      if (targetRoot !== sourceRoot) fail('PROMOTION_TARGET_MISMATCH', 'Promotion target must be the original pinned repository');
      const head = (await gitOutput(sourceRoot, ['rev-parse', '--verify', 'HEAD^{commit}'])).toString('utf8').trim();
      if (head !== repositoryRevision) fail('PROMOTION_BASE_CHANGED', 'Repository revision changed before promotion');
      const preflight = [];
      for (const operation of plan.operations) {
        const current = await readRegularFile(sourceRoot, operation.path);
        if ((current === null ? null : sha256(current)) !== operation.before_digest) fail('PROMOTION_TARGET_CHANGED', `Target bytes changed before promoting ${operation.path}`);
        const parentPath = path.posix.dirname(operation.path);
        if (parentPath !== '.') {
          const parentTree = tree;
          for (let index = 1; index < parentPath.split('/').length; index += 1) {
            const prefix = parentPath.split('/').slice(0, index).join('/');
            if (parentTree.has(prefix)) fail('PROMOTION_TARGET_CHANGED', 'Target parent path is not a directory');
          }
        }
        const target = path.join(sourceRoot, ...operation.path.split('/'));
        const mode = current === null ? 0o644 : ((await lstat(target)).mode & 0o111) ? 0o755 : 0o644;
        preflight.push({ operation, target, mode });
      }
      const promotions = [];
      for (const { operation, target, mode } of preflight) {
        if (operation.action === 'delete') {
          await rm(target);
        } else {
          const containing = path.dirname(target);
          await mkdir(containing, { recursive: true });
          const temp = path.join(containing, `.foundry-promote-${randomUUID()}`);
          await writeFile(temp, operation.content, { flag: 'wx', mode });
          await rm(target, { force: true });
          await rename(temp, target);
        }
        promotions.push({ path: operation.path, action: operation.action });
      }
      return { promoted: true, execution_id: executionId, attempt_id: attemptId, operations: promotions };
    },
    async revokeCapabilities({ attempt_id, fencing_token, attempt } = {}) {
      if (disposed || attempt_id !== attemptId || attempt?.attempt_id !== attemptId || !Number.isSafeInteger(fencing_token)) return { stopped: false };
      return host.revokeCapabilities({ executionId, attemptId, fencingToken: fencing_token, verificationId: workspace.verificationId });
    },
    async dispose() {
      if (disposed) return false;
      if (typeof host.disposeWorkspace === 'function') {
        const safe = await host.disposeWorkspace({ workspaceRoot, executionId, attemptId, verificationId: workspace.verificationId });
        if (safe !== true) fail('ISOLATION_CLEANUP_UNCONFIRMED', 'Docker host has not confirmed all attempt processes stopped');
      }
      disposed = true;
      artifacts.clear();
      await rm(parent, { recursive: true, force: true });
      return true;
    },
  };
  return Object.freeze(workspace);
}

export function validateWorkspacePlan(plan) {
  return validatePlan(plan);
}

export const workspaceDefaults = Object.freeze({ imageDigest: DEFAULT_IMAGE, maxBytes: DEFAULT_TOTAL_BYTES });
