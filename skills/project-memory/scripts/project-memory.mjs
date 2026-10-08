#!/usr/bin/env node
import { appendFileSync, existsSync, lstatSync, mkdirSync, readFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { isAbsolute, relative, resolve } from 'node:path';

// A local continuity aid, not a proof that declarations or test results are true.
const LEDGER = '.project-memory/tasks.jsonl';
const TYPES = new Set(['start', 'checkpoint', 'complete', 'abandon']);
const MODES = new Set(['100644', '100755', '120000', '160000']);
const [command = 'status', ...argv] = process.argv.slice(2);
let root;
const gitBytes = (...args) => execFileSync('git', ['--literal-pathspecs', ...args], {
  cwd: root, maxBuffer: 16 * 1024 * 1024,
  stdio: ['ignore', 'pipe', 'pipe'], env: { ...process.env, GIT_OPTIONAL_LOCKS: '0' },
});
const utf8 = (bytes) => new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(bytes);
const fail = (message) => { throw new Error(message); };
const git = (...args) => {
  const bytes = gitBytes(...args);
  try { return utf8(bytes); }
  catch { fail('Git output contains unsupported non-UTF-8 bytes; use UTF-8 repository paths'); }
};
const text = (value) => typeof value === 'string' && value.trim().length > 0;
const canonicalPath = (path) => typeof path === 'string' && path.length > 0
  && !path.includes('\0') && !path.includes('\\') && !isAbsolute(path)
  && !/^[A-Za-z]:/.test(path) && path.split('/').every((part) => part && part !== '.' && part !== '..');
const assertFiles = (files, label, legacy = false) => {
  if (!Array.isArray(files) || files.length === 0 || new Set(files).size !== files.length
      || files.some((file) => !canonicalPath(file) || (!legacy && file === LEDGER))) {
    fail(`${label}: files must be unique repository-relative file paths, excluding ${LEDGER}`);
  }
};
const parse = (raw, label) => {
  const events = [];
  for (const [index, line] of raw.split('\n').entries()) {
    if (!line.trim()) continue;
    let event;
    try { event = JSON.parse(line); } catch { fail(`${label}:${index + 1}: invalid JSON`); }
    events.push(event);
  }
  return events;
};
const assertEvent = (event, label, fresh = false) => {
  if (!event || typeof event !== 'object' || Array.isArray(event)) fail(`${label}: expected an event object`);
  const legacy = event.version === undefined;
  if ((fresh && legacy) || (!legacy && event.version !== 1)) fail(`${label}: new events require version 1`);
  const fields = ['version', 'id', 'event', 'at', 'summary', 'next', 'files', 'task', 'snapshot', 'validation'];
  if (!legacy && Object.keys(event).some((key) => !fields.includes(key))) fail(`${label}: unknown event field`);
  if (!text(event.id) || !/^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$/.test(event.id)) fail(`${label}: invalid task id`);
  if (!TYPES.has(event.event) || (legacy && !['start', 'complete'].includes(event.event))) fail(`${label}: invalid event type`);
  if (!text(event.summary) || !text(event.next)) fail(`${label}: summary and next must be nonempty strings`);
  if (typeof event.at !== 'string' || !/^\d{4}-\d{2}-\d{2}T/.test(event.at)
      || !Number.isFinite(Date.parse(event.at))) fail(`${label}: invalid timestamp`);
  if (!legacy && new Date(event.at).toISOString() !== event.at) fail(`${label}: timestamp must be canonical UTC ISO 8601`);
  if (event.task !== undefined && (!text(event.task) || !/^https:\/\/[^\s]+$/.test(event.task))) fail(`${label}: task must be an HTTPS canonical task URL`);
  assertFiles(event.files, label, legacy);
  if (!legacy && event.event === 'complete') {
    if (!text(event.validation)) fail(`${label}: completion requires validation evidence`);
    if (!event.snapshot || typeof event.snapshot !== 'object' || Array.isArray(event.snapshot)
        || Object.keys(event.snapshot).length !== event.files.length) fail(`${label}: completion snapshot must match files`);
    for (const file of event.files) {
      if (!Object.hasOwn(event.snapshot, file)) fail(`${label}: missing snapshot for ${JSON.stringify(file)}`);
      const item = event.snapshot[file];
      if (!item || typeof item !== 'object' || Array.isArray(item)
          || Object.keys(item).sort().join(',') !== 'blob,mode') fail(`${label}: invalid snapshot entry`);
      const deleted = item.blob === null && item.mode === null;
      if (!deleted && (!MODES.has(item.mode) || typeof item.blob !== 'string'
          || !/^(?:[a-f0-9]{40}|[a-f0-9]{64})$/.test(item.blob))) fail(`${label}: invalid Git blob or mode`);
    }
  } else if (!legacy && (event.snapshot !== undefined || event.validation !== undefined)) {
    fail(`${label}: snapshot and validation belong only on completions`);
  }
};
const replay = (events, freshFrom = events.length) => {
  const states = new Map();
  for (const [index, event] of events.entries()) {
    const label = `event ${index + 1}`;
    assertEvent(event, label, index >= freshFrom);
    const previous = states.get(event.id);
    if (previous && Date.parse(event.at) < Date.parse(previous.event.at)) fail(`${label}: timestamp precedes this task's prior event`);
    const active = previous?.active === true;
    if (event.event === 'start') {
      if (active) fail(`${label}: ${event.id} is already active; use checkpoint`);
    } else {
      if (!active) fail(`${label}: ${event.id} needs an active start`);
      if (event.event !== 'checkpoint' && event.files.some((file) => !previous.event.files.includes(file))) {
        fail(`${label}: files exceed active scope; record a checkpoint with the full scope first`);
      }
    }
    if (previous?.event.task && event.task !== previous.event.task) fail(`${label}: canonical task URL cannot change`);
    states.set(event.id, { event, index, active: ['start', 'checkpoint'].includes(event.event) });
  }
  return states;
};
const treeEntry = (ref, path) => {
  const output = git('ls-tree', '-z', '--full-tree', ref, '--', path);
  if (!output) return { blob: null, mode: null };
  const records = output.split('\0').filter(Boolean);
  if (records.length !== 1) fail(`expected one Git entry for ${JSON.stringify(path)}`);
  const match = /^(\d+) (\w+) ([a-f0-9]+)\t([\s\S]+)$/.exec(records[0]);
  if (!match || match[4] !== path) fail(`expected one exact Git entry for ${JSON.stringify(path)}`);
  // Git tracks files, not directories. A tree at this path means the former file is gone.
  if (match[1] === '040000' && match[2] === 'tree') return { blob: null, mode: null };
  if (!MODES.has(match[1])) fail(`expected a file at ${JSON.stringify(path)}`);
  return { blob: match[3], mode: match[1] };
};
const committedLedger = (ref) => {
  const entry = treeEntry(ref, LEDGER);
  if (entry.blob === null) return '';
  if (entry.mode !== '100644' && entry.mode !== '100755') fail(`${LEDGER} must be a regular file`);
  return utf8(gitBytes('cat-file', 'blob', entry.blob));
};
const localLedger = () => {
  for (const path of ['.project-memory', LEDGER]) {
    try {
      if (lstatSync(resolve(root, path)).isSymbolicLink()) fail(`${path} must not be a symlink`);
    } catch (error) { if (error.code !== 'ENOENT') throw error; }
  }
  return existsSync(resolve(root, LEDGER)) ? utf8(readFileSync(resolve(root, LEDGER))) : '';
};
const hasWorktreeFile = (file) => {
  const parts = file.split('/');
  let path = root;
  for (const [index, part] of parts.entries()) {
    path = resolve(path, part);
    let stat;
    try { stat = lstatSync(path); }
    catch (error) { if (['ENOENT', 'ENOTDIR'].includes(error.code)) return false; throw error; }
    if (index === parts.length - 1) return !stat.isDirectory();
    // Git never tracks paths through a symlink or a regular-file ancestor.
    if (!stat.isDirectory()) return false;
  }
  return false;
};
const snapshotIndex = (files) => Object.fromEntries(files.map((file) => {
  // Use the staged Git representation, preserving modes, symlinks and line-ending filters.
  try { git('diff', '--quiet', '--no-ext-diff', '--no-textconv', '--ignore-submodules=none', '--', file); }
  catch { fail(`stage final tested changes before completion: ${JSON.stringify(file)}`); }
  const output = git('ls-files', '--stage', '-z', '--', file);
  const records = output.split('\0').filter(Boolean).map((record) => {
    const match = /^(\d+) ([a-f0-9]+) (\d)\t([\s\S]+)$/.exec(record);
    if (!match) fail(`invalid Git index entry for ${JSON.stringify(file)}`);
    return match;
  }).filter((match) => match[4] === file);
  if (!records.length) {
    // A directory replacing a former file has no exact index entry; its children
    // need their own coverage. ENOTDIR means an ancestor became a file instead.
    if (hasWorktreeFile(file)) fail(`stage the new file before completion: ${JSON.stringify(file)}`);
    return [file, { blob: null, mode: null }];
  }
  const match = records.length === 1 && records[0];
  if (!match || match[3] !== '0' || !MODES.has(match[1])) fail(`resolve index conflicts or name a file: ${JSON.stringify(file)}`);
  return [file, { blob: match[2], mode: match[1] }];
}));
const parseArgs = () => {
  const result = new Map();
  for (let i = 0; i < argv.length; i += 1) {
    if (!argv[i].startsWith('--') || argv[i] === '--') fail(`unexpected argument: ${argv[i]}`);
    const key = argv[i].slice(2);
    if (result.has(key)) fail(`duplicate --${key}`);
    const values = [];
    while (argv[i + 1] !== undefined && !argv[i + 1].startsWith('--')) values.push(argv[++i]);
    if (values.length === 0) fail(`--${key} needs a value`);
    result.set(key, values);
  }
  return result;
};
const allowed = {
  start: ['id', 'summary', 'next', 'files', 'task'],
  checkpoint: ['id', 'summary', 'next', 'files', 'task'],
  complete: ['id', 'summary', 'next', 'files', 'task', 'validation'],
  abandon: ['id', 'summary', 'next', 'task'],
  status: ['stale-days'], validate: ['base'],
};

try {
  root = git('rev-parse', '--show-toplevel').replace(/\r?\n$/, '');
  if (!Object.hasOwn(allowed, command)) fail(`unknown command: ${command}`);
  const args = parseArgs();
  for (const key of args.keys()) if (!allowed[command].includes(key)) fail(`unknown option for ${command}: --${key}`);
  const value = (key, required = false) => {
    const result = (args.get(key) || []).join(' ').trim();
    if (required && !result) fail(`--${key} is required`);
    return result;
  };
  if (command === 'validate') {
    const base = value('base', true);
    const baseCommit = git('rev-parse', '--verify', '--end-of-options', `${base}^{commit}`).trim();
    const head = git('rev-parse', '--verify', 'HEAD^{commit}').trim();
    const mergeBase = git('merge-base', baseCommit, head).trim();
    const before = committedLedger(mergeBase);
    const current = committedLedger(head);
    if (!current.startsWith(before)) fail(`${LEDGER} changed historical bytes; preserve the merge-base ledger and append events`);
    if (before && !before.endsWith('\n') && current !== before && current[before.length] !== '\n') fail('append a newline after the historical final event');
    const historical = parse(before, 'base ledger');
    const events = parse(current, 'HEAD ledger');
    const states = replay(events, historical.length);
    // --no-renames gives both old/deleted and new names; -z preserves Unicode and newlines.
    const changed = git('diff', '--no-ext-diff', '--no-textconv', '--ignore-submodules=none', '--no-renames', '--name-only', '-z', mergeBase, head, '--')
      .split('\0').filter((path) => path && path !== LEDGER);
    const covered = new Set();
    for (const { event, index } of states.values()) {
      if (event.event !== 'complete' || index < historical.length || event.version !== 1) continue;
      for (const file of event.files) {
        const actual = treeEntry(head, file);
        const recorded = event.snapshot[file];
        // A later task may legitimately replace this task's result on a shared file.
        // Only snapshots that still match HEAD contribute current coverage.
        if (recorded.blob === actual.blob && recorded.mode === actual.mode) covered.add(file);
      }
    }
    const missing = changed.filter((file) => !covered.has(file));
    if (missing.length) fail(`changed paths lack fresh final completion evidence:\n${missing.map((file) => JSON.stringify(file)).join('\n')}`);
    console.log(`project-memory: ${changed.length} changed paths covered; ${events.length - historical.length} new events; committed HEAD ${head.slice(0, 12)}`);
  } else {
    const raw = localLedger();
    const events = parse(raw, LEDGER);
    const states = replay(events);
    if (command === 'status') {
      const days = args.has('stale-days') ? Number(value('stale-days')) : 7;
      if (!Number.isFinite(days) || days < 0) fail('--stale-days must be a nonnegative number');
      console.log(JSON.stringify([...states.values()].map(({ event, active }) => ({
        ...event, state: active ? 'active' : event.event === 'abandon' ? 'abandoned' : 'complete',
        stale: active && Date.now() - Date.parse(event.at) > days * 86400000,
      })).sort((a, b) => a.at.localeCompare(b.at)), null, 2));
    } else {
      const id = value('id', true);
      const previous = states.get(id)?.event;
      const files = args.has('files') ? args.get('files').map((file) => relative(root, resolve(root, file)).split('\\').join('/')) : previous?.files;
      const event = { version: 1, id, event: command, at: new Date().toISOString(),
        summary: value('summary', true), next: value('next', true), files };
      const task = value('task') || previous?.task;
      if (task) event.task = task;
      assertFiles(files, 'new event');
      if (command === 'complete') {
        event.validation = value('validation', true);
        event.snapshot = snapshotIndex(files);
      }
      replay([...events, event], events.length);
      mkdirSync(resolve(root, '.project-memory'), { recursive: true });
      appendFileSync(resolve(root, LEDGER), `${raw && !raw.endsWith('\n') ? '\n' : ''}${JSON.stringify(event)}\n`);
      console.log(`${command}: ${id}`);
    }
  }
} catch (error) {
  console.error(`project-memory: ${error.message}`);
  process.exitCode = 1;
}
