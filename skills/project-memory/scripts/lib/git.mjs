import { execFileSync } from 'node:child_process';
import { createHash, randomUUID } from 'node:crypto';
import { existsSync, lstatSync, mkdirSync, readFileSync, readdirSync, writeFileSync, linkSync, unlinkSync } from 'node:fs';
import { dirname, relative, resolve } from 'node:path';
import {
  ROOT, LEGACY, MODES, REPO, metadataPath, metadataContainer, parseJSON, validateSource, validateEvent, buildGraph,
  completionHeads, eventPath, canonicalPath, fail,
} from './state.mjs';

export function utf8(bytes, label = 'data') {
  try { return new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(bytes); }
  catch { fail(`${label} contains unsupported non-UTF-8 bytes`); }
}
export function createGit(cwd = process.cwd()) {
  const runBytes = (where, ...args) => execFileSync('git', ['--no-replace-objects', '--literal-pathspecs', ...args], {
    cwd: where, maxBuffer: 64 * 1024 * 1024, stdio: ['ignore', 'pipe', 'pipe'],
    // Empty allow-list denies every transport, even on Git 2.43 where the
    // newer no-lazy-fetch option/environment variable is not supported.
    env: { ...process.env, GIT_OPTIONAL_LOCKS: '0', GIT_NO_LAZY_FETCH: '1', GIT_ALLOW_PROTOCOL: '', GIT_NO_REPLACE_OBJECTS: '1' },
  });
  const root = utf8(runBytes(cwd, 'rev-parse', '--show-toplevel'), 'Git root').replace(/\r?\n$/, '');
  const bytes = (...args) => runBytes(root, ...args);
  const git = (...args) => utf8(bytes(...args), 'Git output');
  const commit = (ref) => git('rev-parse', '--verify', '--end-of-options', `${ref}^{commit}`).trim();
  const tree = (ref, path) => {
    const args = ['ls-tree', '-r', '-z', '--full-tree', ref];
    if (path) args.push('--', path);
    const entries = new Map();
    for (const row of git(...args).split('\0').filter(Boolean)) {
      const match = /^(\d+) (\w+) ([a-f0-9]+)\t([\s\S]+)$/.exec(row);
      if (!match) fail('invalid Git tree record');
      entries.set(match[4], { mode: match[1], type: match[2], blob: match[3] });
    }
    return entries;
  };
  const readBlob = (blob) => utf8(bytes('cat-file', 'blob', blob), 'Git blob');
  const entry = (ref, path) => {
    const output = git('ls-tree', '-z', '--full-tree', ref, '--', path);
    if (!output) return { blob: null, mode: null };
    const rows = output.split('\0').filter(Boolean);
    const match = rows.length === 1 && /^(\d+) (\w+) ([a-f0-9]+)\t([\s\S]+)$/.exec(rows[0]);
    if (!match || match[4] !== path) fail(`expected an exact Git entry at ${JSON.stringify(path)}`);
    if (match[1] === '040000' && match[2] === 'tree') return { blob: null, mode: null };
    if (!MODES.has(match[1])) fail(`unsupported Git mode at ${JSON.stringify(path)}`);
    return { blob: match[3], mode: match[1] };
  };
  const hasWorktreeFile = (file) => {
    const parts = file.split('/'); let path = root;
    for (const [index, part] of parts.entries()) {
      path = resolve(path, part); let stat;
      try { stat = lstatSync(path); }
      catch (error) { if (['ENOENT', 'ENOTDIR'].includes(error.code)) return false; throw error; }
      if (index === parts.length - 1) return !stat.isDirectory();
      if (!stat.isDirectory()) return false;
    }
    return false;
  };
  const snapshotIndex = (files) => Object.fromEntries(files.map((file) => {
    try { git('diff', '--quiet', '--no-ext-diff', '--no-textconv', '--ignore-submodules=none', '--', file); }
    catch { fail(`stage final tested changes before completion: ${JSON.stringify(file)}`); }
    const records = git('ls-files', '--stage', '-z', '--', file).split('\0').filter(Boolean).map((row) => {
      const match = /^(\d+) ([a-f0-9]+) (\d)\t([\s\S]+)$/.exec(row);
      if (!match) fail('invalid Git index record');
      return match;
    }).filter((match) => match[4] === file);
    if (!records.length) {
      if (hasWorktreeFile(file)) fail(`stage the new file before completion: ${JSON.stringify(file)}`);
      return [file, { blob: null, mode: null }];
    }
    const match = records.length === 1 && records[0];
    if (!match || match[3] !== '0' || !MODES.has(match[1])) fail(`resolve index conflicts at ${JSON.stringify(file)}`);
    return [file, { blob: match[2], mode: match[1] }];
  }));
  const memoryAt = (ref) => {
    const entries = new Map();
    for (const [path, item] of tree(ref, '.project-memory')) {
      if (!metadataPath(path) && path !== LEGACY) continue;
      if (item.type !== 'blob' || item.mode !== '100644') fail(`memory metadata must be a regular non-executable file: ${path}`);
      entries.set(path, { ...item, text: readBlob(item.blob) });
    }
    return entries;
  };
  const branch = () => {
    try { return git('symbolic-ref', '--quiet', '--short', 'HEAD').replace(/\r?\n$/, ''); }
    catch { fail('detached checkout: provide --branch for the implementation branch'); }
  };
  return { root, bytes, git, commit, tree, readBlob, entry, snapshotIndex, memoryAt, branch };
}

function checkAncestors(root, path, create = false) {
  let current = root;
  for (const part of path.split('/')) {
    current = resolve(current, part);
    let stat;
    try { stat = lstatSync(current); }
    catch (error) {
      if (error.code !== 'ENOENT') throw error;
      if (!create) return false;
      try { mkdirSync(current); } catch (mkdirError) { if (mkdirError.code !== 'EEXIST') throw mkdirError; }
      stat = lstatSync(current);
    }
    if (!stat.isDirectory() || stat.isSymbolicLink()) fail(`memory ancestor must be a real directory: ${relative(root, current)}`);
  }
  return true;
}
export function writeImmutable(root, path, value, { identicalOK = false } = {}) {
  if (!canonicalPath(path) || !metadataPath(path)) fail('refusing a non-metadata write path');
  checkAncestors(root, dirname(path), true);
  const destination = resolve(root, path); const text = `${JSON.stringify(value, null, 2)}\n`;
  if (existsSync(destination) || (() => { try { return !!lstatSync(destination); } catch (error) { if (error.code === 'ENOENT') return false; throw error; } })()) {
    const stat = lstatSync(destination);
    if (identicalOK && stat.isFile() && !stat.isSymbolicLink() && !(stat.mode & 0o111) && utf8(readFileSync(destination)) === text) return false;
    fail(`immutable metadata already exists: ${path}`);
  }
  const temporary = `${destination}.${randomUUID()}.tmp`;
  try {
    writeFileSync(temporary, text, { flag: 'wx', mode: 0o644 });
    // A hard-link publication is exclusive and never exposes a partially written event.
    try { linkSync(temporary, destination); }
    catch (error) {
      if (error.code === 'EEXIST' && identicalOK) {
        const stat = lstatSync(destination);
        if (stat.isFile() && !stat.isSymbolicLink() && !(stat.mode & 0o111) && utf8(readFileSync(destination)) === text) return false;
      }
      throw error;
    }
  } finally {
    try { unlinkSync(temporary); } catch (error) { if (error.code !== 'ENOENT') throw error; }
  }
  return true;
}
export function localMemory(root) {
  const entries = new Map();
  if (!checkAncestors(root, '.project-memory')) return entries;
  const inspect = (path) => {
    const container = metadataContainer(path);
    // Atomic writers may remove temporary files after readdir. They are not
    // metadata, so do not stat/read them or ordinary project files here.
    if (!container && !metadataPath(path) && path !== LEGACY) return;
    const full = resolve(root, path); const stat = lstatSync(full);
    if (stat.isSymbolicLink()) {
      if (container || metadataPath(path) || path === LEGACY) fail(`memory metadata/ancestor must not be a symlink: ${path}`);
      return;
    }
    if (stat.isDirectory()) {
      if (container) for (const child of readdirSync(full)) inspect(`${path}/${child}`);
    } else if (metadataPath(path) || path === LEGACY) {
      if (!stat.isFile() || stat.mode & 0o111) fail(`memory metadata must be a regular non-executable file: ${path}`);
      const text = utf8(readFileSync(full), path);
      entries.set(path, { mode: '100644', type: 'blob', text,
        blob: createHash('sha1').update(`blob ${Buffer.byteLength(text)}\0`).update(text).digest('hex') });
    }
  };
  for (const child of readdirSync(resolve(root, '.project-memory'))) {
    const path = `.project-memory/${child}`;
    if (path === ROOT || path === LEGACY) inspect(path);
  }
  return entries;
}
export function loadMemory(entries) {
  const sources = new Map(); const events = []; const eventFiles = new Map();
  for (const [path, item] of entries) {
    if (path === LEGACY) continue;
    if (/\/sources\//.test(path)) {
      const source = validateSource(parseJSON(item.text, path), path); sources.set(path, source);
    }
  }
  for (const [path, item] of entries) {
    if (path === LEGACY || /\/sources\//.test(path)) continue;
    const event = validateEvent(parseJSON(item.text, path), sources, path);
    events.push(event); eventFiles.set(event.eventId, path);
  }
  return { sources, events, eventFiles, graph: buildGraph(events, sources) };
}
export function preserveMetadata(before, after) {
  for (const [path, old] of before) {
    const current = after.get(path);
    if (!current || current.mode !== old.mode || current.blob !== old.blob || current.text !== old.text) fail(`immutable base metadata changed or removed: ${path}`);
  }
  if (!before.has(LEGACY) && after.has(LEGACY)) fail('new legacy tasks.jsonl is not accepted; write v2 events');
}
export function validateCommitted(repo, { base, repository, branch, backlogRepository } = {}) {
  if (!base) fail('--base is required; no comparison base is guessed');
  if (!REPO.test(repository || '') || !REPO.test(backlogRepository || '')) fail('explicit code and source repository identities are required');
  const head = repo.commit('HEAD'); const mergeBase = repo.git('merge-base', repo.commit(base), head).trim();
  const before = repo.memoryAt(mergeBase); const after = repo.memoryAt(head);
  preserveMetadata(before, after);
  const memory = loadMemory(after);
  if (backlogRepository && [...memory.sources.values()].some((source) => source.repository !== backlogRepository)) fail('task-source backlog repository disagrees with the reviewed contract');
  const fresh = new Set(memory.events.filter((event) => !before.has(eventPath(event))).map((event) => event.eventId));
  if (repository && memory.events.some((event) => fresh.has(event.eventId) && event.repository !== repository)) fail('new event code repository disagrees with validation context');
  const completions = completionHeads(memory.graph, fresh);
  const changed = repo.git('diff', '--no-ext-diff', '--no-textconv', '--ignore-submodules=none', '--no-renames', '--name-only', '-z', mergeBase, head, '--')
    .split('\0').filter((path) => path && !after.has(path));
  const covered = new Set();
  for (const event of completions) {
    // Event branches describe where work was recorded. Integrating another
    // branch must not require rewriting valid origin provenance.
    for (const file of event.files) {
      const actual = repo.entry(head, file); const expected = event.snapshot[file];
      if (actual.blob === expected.blob && actual.mode === expected.mode) covered.add(file);
    }
  }
  const missing = changed.filter((path) => !covered.has(path));
  if (missing.length) fail(`changed paths lack fresh final completion evidence:\n${missing.map((path) => JSON.stringify(path)).join('\n')}`);
  return { head, mergeBase, paths: changed.length, newEvents: fresh.size, integrationBranch: branch || null };
}
