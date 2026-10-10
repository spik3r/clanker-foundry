import test from 'node:test';
import fs from 'node:fs';
import childProcess from 'node:child_process';
import { syncBuiltinESMExports } from 'node:module';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync, chmodSync, symlinkSync, lstatSync, readdirSync, renameSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawn, spawnSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import {
  ROOT, TASK, metadataPath, blobHash, digest, claimSections, makeSource, sourcePath, validateSource,
  validateEvent, buildGraph, completionHeads, eventPath,
} from '../skills/project-memory/scripts/lib/state.mjs';
import { createGit, localMemory, loadMemory, writeImmutable } from '../skills/project-memory/scripts/lib/git.mjs';

const REPO = 'example/application';
const BACKLOG = 'example/backlog';
const REV = '1'.repeat(40);
const uuid = (n) => `00000000-0000-4000-8000-${n.toString(16).padStart(12, '0')}`;
const record = (id = 'CF-101', status = 'in-progress', extra = '') => `---\nid: ${id}\ncategory: tooling\nstatus: ${status}\nowner: Synthetic Fixture\nrepositories: [${REPO}]\ndepends_on: []\ncredentials: none\n---\n\n# Synthetic fixture task\n\n## Claim\n\nFixture-only claim for isolated tests.\n${extra}`;
const entry = (text) => ({ text, blob: blobHash(text), mode: '100644' });
const source = (id = 'CF-101', revision = REV, extra = '') => makeSource({
  repository: BACKLOG, revision, id, path: `tasks/${id}-fixture.md`,
  records: new Map([[`tasks/${id}-fixture.md`, entry(record(id, 'in-progress', extra))]]),
});
const task = (s) => ({ repository: s.repository, id: s.id, path: s.path, revision: s.revision, blob: s.blob,
  claim: claimSections(s.record).get('Claim') });
const ev = (s, n, kind = 'start', parent = null, overrides = {}) => ({
  formatVersion: 2, eventId: uuid(n), runId: parent?.runId || uuid(n + 1000), kind,
  parents: parent ? [parent.eventId] : [], sequence: parent ? parent.sequence + 1 : 0,
  task: task(s), repository: REPO, branch: 'codex/fixture', pullRequest: null,
  at: '2026-01-01T00:00:00.000Z', summary: 'Synthetic fixture', next: 'Review', files: ['file.txt'],
  ...(kind === 'complete' ? { validation: 'Synthetic checks passed', snapshot: { 'file.txt': { blob: 'a'.repeat(40), mode: '100644' } } } : {}),
  ...overrides,
});
const sources = (...items) => new Map(items.map((s) => [sourcePath(s), s]));
const graph = (events, ...items) => buildGraph(events, sources(...items));

test('source validates the actual canonical ID, repository membership and exact record blob', () => {
  const s = source(); assert.equal(validateSource(s), s);
  assert.throws(() => validateSource({ ...s, id: 'CF-102' }), /identity|path/);
  assert.throws(() => validateSource({ ...s, record: s.record.replace('Synthetic Fixture', 'Changed Owner') }), /blob/);
});
for (const status of ['ready', 'in-progress', 'blocked', 'review', 'done', 'closed']) {
  test(`a second canonical ID still conflicts when its status is ${status}`, () => {
    const records = new Map([
      ['tasks/CF-101-fixture.md', entry(record())],
      ['tasks/CF-101-unrelated.md', entry(record('CF-101', status))],
    ]);
    assert.throws(() => makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records }), /ambiguous/);
  });
}
test('incidental prose mentions in another canonical task are not identity collisions', () => {
  const records = new Map([
    ['tasks/CF-101-fixture.md', entry(record())],
    ['tasks/CF-102-other.md', entry(record('CF-102', 'done', 'This refers to CF-101 as a dependency.'))],
  ]);
  assert.equal(makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records }).id, 'CF-101');
});
test('explicit noncanonical redirect resolves to its exact existing canonical target', () => {
  const redirect = '# Historical task — noncanonical redirect\n\nThe canonical record is [CF-102](CF-102-other.md).\n';
  const records = new Map([
    ['tasks/CF-101-fixture.md', entry(record())],
    ['tasks/CF-101-old.md', entry(redirect)],
    ['tasks/CF-102-other.md', entry(record('CF-102'))],
  ]);
  const s = makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records });
  assert.equal(s.identityEvidence.redirects[0].targetPath, 'tasks/CF-102-other.md');
  const bad = structuredClone(s); bad.identityEvidence.redirects[0].targetPath = 'tasks/CF-102-spoof.md';
  assert.throws(() => validateSource(bad), /redirect/);
});
test('a redirect label cannot hide a canonical ID or missing target', () => {
  const records = new Map([
    ['tasks/CF-101-fixture.md', entry(record())],
    ['tasks/CF-101-old.md', entry('# Historical task — noncanonical redirect\nThe canonical record is [CF-102](CF-102-missing.md).\n')],
  ]);
  assert.throws(() => makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records }), /existing/);
  records.set('tasks/CF-101-old.md', entry(record('CF-101', 'done') + '# noncanonical redirect\n'));
  assert.throws(() => makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records }), /ambiguous/);
});
test('fresh unique complete frontier supplies coverage', () => {
  const s = source(); const start = ev(s, 1); const complete = ev(s, 2, 'complete', start);
  const g = graph([complete, start], s);
  assert.deepEqual(completionHeads(g, new Set([complete.eventId])), [complete]);
  assert.deepEqual(completionHeads(g, new Set()), []);
});
test('new active run supersedes old completion and removes its coverage', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'complete', a);
  const c = ev(s, 3, 'start', null, { parents: [b.eventId] });
  assert.deepEqual(completionHeads(graph([a, b, c], s), new Set([b.eventId, c.eventId])), []);
});
for (const kind of ['checkpoint', 'complete', 'abandon']) {
  test(`complete plus competing same-run ${kind} head fails until explicit reconciliation`, () => {
    const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'complete', a); const c = ev(s, 3, kind, a);
    const g = graph([a, b, c], s); assert.equal(g.tasks.get(s.id).heads.length, 2);
    assert.throws(() => completionHeads(g, new Set([b.eventId])), /unresolved frontier/);
    const d = ev(s, 4, 'start', null, { parents: [b.eventId, c.eventId], files: ['remaining.txt'] });
    const e = ev(s, 5, 'complete', d, { files: ['remaining.txt'], snapshot: { 'remaining.txt': { blob: 'b'.repeat(40), mode: '100644' } } });
    assert.deepEqual(completionHeads(graph([a, b, c, d, e], s), new Set([e.eventId])), [e]);
  });
}
test('unrelated separate runs of one task are conflicting frontiers even with identical snapshots', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'complete', a); const c = ev(s, 3); const d = ev(s, 4, 'complete', c);
  assert.throws(() => completionHeads(graph([a, b, c, d], s), new Set([b.eventId, d.eventId])), /unresolved frontier/);
});
test('reconciliation that misses a competing head cannot pick a winner', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'complete', a); const c = ev(s, 3, 'checkpoint', a);
  const d = ev(s, 4, 'start', null, { parents: [b.eventId] }); const e = ev(s, 5, 'complete', d);
  assert.throws(() => completionHeads(graph([a, b, c, d, e], s), new Set([e.eventId])), /unresolved frontier/);
});
test('timestamps do not select or order causal heads', () => {
  const s = source(); const a = ev(s, 1, 'start', null, { at: '2099-01-01T00:00:00.000Z' });
  const b = ev(s, 2, 'complete', a, { at: '2000-01-01T00:00:00.000Z' });
  assert.deepEqual(completionHeads(graph([b, a], s), new Set([b.eventId])), [b]);
});
test('missing/cross-task causal parents and causal cycles fail structurally', () => {
  const s = source(); const other = source('CF-102'); const a = ev(s, 1); const b = ev(other, 2);
  assert.throws(() => graph([{ ...a, parents: [uuid(99)] }], s), /missing parent/);
  assert.throws(() => graph([{ ...a, parents: [b.eventId] }, b], s, other), /cross-task/);
  assert.throws(() => graph([{ ...a, parents: [uuid(3)] }, ev(s, 3, 'start', null, { parents: [a.eventId] })], s), /cycle/);
});
test('duplicate event IDs and multiple starts in a run fail', () => {
  const s = source(); const a = ev(s, 1);
  assert.throws(() => graph([a, { ...a }], s), /duplicate event/);
  assert.throws(() => graph([a, ev(s, 2, 'start', null, { runId: a.runId })], s), /exactly one start/);
});
test('run task/source/claim/branch identity cannot be swapped', () => {
  const s = source(); const updated = source('CF-101', '2'.repeat(40), 'A revised claim.'); const a = ev(s, 1);
  for (const changed of [{ task: task(updated) }, { branch: 'other' }]) {
    const b = ev(s, 2, 'complete', a, changed);
    assert.throws(() => graph([a, b], s, updated), /identity/);
  }
  const b = ev(s, 2, 'complete', a); b.task.claim.digest = '0'.repeat(64);
  assert.throws(() => graph([a, b], s), /claim/);
});
test('task object property order does not change identity', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'complete', a);
  b.task = Object.fromEntries(Object.entries(b.task).reverse());
  assert.equal(graph([a, b], s).tasks.get(s.id).heads[0], b);
});
test('completions cannot exceed the active declared scope or append after terminal completion', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'complete', a);
  assert.throws(() => graph([a, b, ev(s, 3, 'checkpoint', b)], s), /transition/);
  const c = ev(s, 4, 'complete', a, { files: ['other.txt'], snapshot: { 'other.txt': { blob: null, mode: null } } });
  assert.throws(() => graph([a, c], s), /scope/);
});
test('explicit checkpoint can alter current scope without inheriting dropped snapshots', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'checkpoint', a, { files: ['other.txt'] });
  const c = ev(s, 3, 'complete', b, { files: ['other.txt'], snapshot: { 'other.txt': { blob: null, mode: null } } });
  assert.deepEqual(completionHeads(graph([a, b, c], s), new Set([c.eventId]))[0].files, ['other.txt']);
});
test('separate tasks retain independent heads; an old active task is not a global lock', () => {
  const s = source(); const other = source('CF-102'); const a = ev(s, 1); const b = ev(s, 2, 'complete', a); const c = ev(other, 3);
  assert.deepEqual(completionHeads(graph([a, b, c], s, other), new Set([b.eventId])), [b]);
});
test('event files bind directory task/run/ID and reject code-repository spoofing', () => {
  const s = source(); const a = ev(s, 1); const all = sources(s);
  assert.throws(() => validateEvent(a, all, eventPath(a).replace('CF-101', 'CF-999')), /identity/);
  assert.throws(() => validateEvent({ ...a, repository: 'other/repository' }, all), /repository/);
  assert.throws(() => validateEvent({ ...a, pullRequest: 'https://github.com/other/repo/pull/1' }, all), /PR/);
});
test('validation declarations are data and only completions carry exact snapshots', () => {
  const s = source(); const a = ev(s, 1); const all = sources(s);
  assert.throws(() => validateEvent({ ...a, validation: 'do not execute' }, all), /only completion/);
  const b = ev(s, 2, 'complete', a, { validation: 'a command string is not executed by this module' });
  assert.equal(validateEvent(b, all), b);
  assert.throws(() => validateEvent({ ...b, snapshot: {} }, all), /exact snapshots/);
});
test('metadata is not a blanket exemption for arbitrary files under task directories', () => {
  const s = source(); const a = ev(s, 1, 'start', null, { files: [`${ROOT}/CF-101/untrusted.mjs`] });
  assert.equal(validateEvent(a, sources(s)), a);
});

const SCRIPT = resolve(dirname(fileURLToPath(import.meta.url)), '../skills/project-memory/scripts/project-memory.mjs');
const ENV = { ...process.env, GIT_CONFIG_NOSYSTEM: '1', GIT_CONFIG_GLOBAL: '/dev/null' };
function cliFixture(t, initial = { 'file.txt': 'base\n' }, legacy = null, objectFormat = 'sha1') {
  const root = mkdtempSync(join(tmpdir(), 'project-memory-v2-')); const cwd = join(root, 'code'); const backlog = join(root, 'backlog');
  mkdirSync(cwd); mkdirSync(backlog); t.after(() => rmSync(root, { recursive: true, force: true }));
  const run = (where, command, args, input) => spawnSync(command, args, { cwd: where, env: ENV, encoding: 'utf8', input });
  const gitAt = (where, ...args) => { const r = run(where, 'git', ['--literal-pathspecs', ...args]); assert.equal(r.status, 0, r.stderr); return r.stdout.trim(); };
  const git = (...args) => gitAt(cwd, ...args);
  for (const directory of [cwd, backlog]) {
    gitAt(directory, 'init', '-q', '--object-format=' + objectFormat); gitAt(directory, 'config', 'user.name', 'Synthetic Fixture');
    gitAt(directory, 'config', 'user.email', 'fixture@example.invalid'); gitAt(directory, 'config', 'commit.gpgsign', 'false');
    gitAt(directory, 'config', 'core.autocrlf', 'false'); gitAt(directory, 'config', 'core.hooksPath', '.empty-hooks');
  }
  gitAt(backlog, 'remote', 'add', 'origin', `https://github.com/${BACKLOG}`);
  const write = (path, contents) => { mkdirSync(dirname(join(cwd, path)), { recursive: true }); writeFileSync(join(cwd, path), contents); };
  const writeBacklog = (path, contents) => { mkdirSync(dirname(join(backlog, path)), { recursive: true }); writeFileSync(join(backlog, path), contents); };
  for (const id of ['CF-101', 'CF-102', 'CF-103']) writeBacklog(`tasks/${id}-fixture.md`, record(id));
  gitAt(backlog, 'add', '--', 'tasks'); gitAt(backlog, 'commit', '-qm', 'synthetic canonical tasks');
  let sourceRevision = gitAt(backlog, 'rev-parse', 'HEAD');
  const present = (path) => {
    const parts = path.split('/'); let current = cwd;
    for (const [index, part] of parts.entries()) {
      current = join(current, part); let stat;
      try { stat = lstatSync(current); } catch (error) { if (['ENOENT', 'ENOTDIR'].includes(error.code)) return false; throw error; }
      if (index < parts.length - 1 && !stat.isDirectory()) return false;
    }
    return true;
  };
  const stage = (...paths) => {
    for (const path of paths) if (present(path) || git('ls-files', '--', path)) git('add', '-A', '--', path);
  };
  const commit = (...paths) => {
    stage(...paths);
    if (present('.project-memory') || git('ls-files', '--', '.project-memory')) git('add', '-A', '--', '.project-memory');
    git('commit', '-qm', 'synthetic fixture'); return git('rev-parse', 'HEAD');
  };
  for (const [path, contents] of Object.entries(initial)) write(path, contents);
  if (legacy !== null) write('.project-memory/tasks.jsonl', legacy);
  const base = commit(...Object.keys(initial));
  const cli = (...args) => run(cwd, process.execPath, [SCRIPT, ...args,
    ...(args[0] === 'validate' && !args.includes('--backlog-repository') ? ['--backlog-repository', BACKLOG] : [])]);
  const pass = (...args) => { const r = cli(...args); assert.equal(r.status, 0, r.stderr); return r; };
  const reject = (pattern, ...args) => { const r = cli(...args); assert.equal(r.status, 1, r.stdout); assert.match(r.stderr, pattern); return r; };
  const register = (id = 'CF-101') => pass('register', '--id', id, '--backlog', backlog, '--ref', sourceRevision, '--record', `tasks/${id}-fixture.md`);
  const start = (id = 'CF-101', files = ['file.txt']) => {
    register(id);
    return JSON.parse(pass('start', '--id', id, '--repository', REPO, '--branch', 'codex/fixture', '--files', ...files, '--summary', 'Synthetic scope', '--next', 'Implement').stdout);
  };
  const complete = (id = 'CF-101', files) => JSON.parse(pass('complete', '--id', id, '--summary', 'Synthetic checked result', '--next', 'Review', '--validation', 'Synthetic checks passed', ...(files ? ['--files', ...files] : [])).stdout);
  const readEvent = (result) => JSON.parse(readFileSync(join(cwd, result.event), 'utf8'));
  const rewriteEvent = (result, change) => write(result.event, JSON.stringify(change(readEvent(result)), null, 2) + '\n');
  const memory = () => loadMemory(localMemory(cwd));
  const updateBacklog = (path, text) => { writeBacklog(path, text); gitAt(backlog, 'add', '--', path); gitAt(backlog, 'commit', '-qm', 'synthetic task update'); sourceRevision = gitAt(backlog, 'rev-parse', 'HEAD'); return sourceRevision; };
  return { root, cwd, backlog, git, gitAt, write, stage, commit, base, cli, pass, reject, register, start, complete, readEvent, rewriteEvent, memory, updateBacklog, sourceRevision: () => sourceRevision };
}
const finishCLI = (f, files = ['file.txt'], id = 'CF-101') => { f.stage(...files); const e = f.complete(id, files); f.commit(...files); return e; };
const checkCLI = (f, ...extra) => f.pass('validate', '--base', f.base, '--repository', REPO, ...extra);
const missingCLI = (f) => f.reject(/fresh final completion/, 'validate', '--base', f.base, '--repository', REPO);

test('CLI import, start, staged completion and committed validation are ergonomic', (t) => {
  const f = cliFixture(t); const start = f.start(); f.write('file.txt', 'changed'); const complete = finishCLI(f); checkCLI(f, '--branch', 'codex/fixture');
  assert.equal(f.readEvent(complete).parents[0], start.eventId);
  assert.equal(f.readEvent(complete).snapshot['file.txt'].blob, f.git('rev-parse', 'HEAD:file.txt'));
  assert.equal(f.memory().events.length, 2);
  const status = JSON.parse(f.pass('status').stdout); assert.equal(status.tasks[0].heads[0].kind, 'complete');
});
test('CLI refuses uncovered edits and unchanged historical completion reuse', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'first'); finishCLI(f); f.base = f.git('rev-parse', 'HEAD');
  f.write('file.txt', 'new unrecorded change'); f.commit('file.txt'); missingCLI(f);
});
test('CLI start after completion links the old frontier but requires a new completion', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'first'); const old = finishCLI(f);
  const next = f.start(); assert.deepEqual(f.readEvent(next).parents, [old.eventId]); f.commit(); missingCLI(f);
  f.write('file.txt', 'second'); finishCLI(f); checkCLI(f);
});
test('CLI resumes historical active work without another start or UUID copying', (t) => {
  const f = cliFixture(t); const first = f.start(); f.commit(); f.base = f.git('rev-parse', 'HEAD');
  const resumed = JSON.parse(f.pass('resume', '--id', 'CF-101', '--summary', 'Resume synthetic work', '--next', 'Validate').stdout);
  assert.equal(resumed.runId, first.runId); assert.deepEqual(f.readEvent(resumed).parents, [first.eventId]);
  f.write('file.txt', 'changed'); finishCLI(f); checkCLI(f);
});
test('CLI refuses accidental restart of an active task', (t) => {
  const f = cliFixture(t); f.start();
  f.reject(/active; use resume/, 'start', '--id', 'CF-101', '--files', 'file.txt', '--summary', 'Again', '--next', 'Continue');
});
test('same-path content edits, amendment and mode edits stale the recorded snapshot', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'first'); finishCLI(f);
  f.write('file.txt', 'later'); f.stage('file.txt'); f.git('commit', '--amend', '-qm', 'changed source'); missingCLI(f);
  f.start(); finishCLI(f); chmodSync(join(f.cwd, 'file.txt'), 0o755); f.commit('file.txt'); missingCLI(f);
});
test('later completed different task can update a shared file without timestamp selection', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'first'); finishCLI(f);
  f.start('CF-102'); f.write('file.txt', 'second'); finishCLI(f, ['file.txt'], 'CF-102'); checkCLI(f);
});
test('final completion omits old paths rather than unioning prior task cycles', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base', 'other.txt': 'base' }); f.start('CF-101', ['file.txt', 'other.txt']);
  f.write('file.txt', 'changed'); finishCLI(f, ['file.txt', 'other.txt']); f.start('CF-101', ['file.txt', 'other.txt']);
  f.complete('CF-101', ['other.txt']); f.commit(); missingCLI(f);
});
for (const mutation of ['rewrite', 'delete', 'rename', 'mode']) {
  test(`base event ${mutation} is rejected even when other source evidence remains valid`, (t) => {
    const f = cliFixture(t); const e = f.start(); f.write('file.txt', 'changed'); finishCLI(f); f.base = f.git('rev-parse', 'HEAD');
    if (mutation === 'rewrite') f.rewriteEvent(e, (event) => ({ ...event, summary: 'rewritten history' }));
    if (mutation === 'delete') rmSync(join(f.cwd, e.event));
    if (mutation === 'rename') renameSync(join(f.cwd, e.event), join(f.cwd, e.event.replace(e.eventId, randomUUID())));
    if (mutation === 'mode') chmodSync(join(f.cwd, e.event), 0o755);
    f.commit(); f.reject(/immutable|non-executable/, 'validate', '--base', f.base, '--repository', REPO);
  });
}
test('base source snapshot cannot be rewritten, including whitespace-only edits', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'changed'); finishCLI(f); f.base = f.git('rev-parse', 'HEAD');
  const path = [...f.memory().sources.keys()][0]; f.write(path, readFileSync(join(f.cwd, path), 'utf8') + ' '); f.commit();
  f.reject(/immutable/, 'validate', '--base', f.base, '--repository', REPO);
});
test('immutable writer never overwrites an existing event, and source registration is byte-idempotent', (t) => {
  const f = cliFixture(t); const first = f.start(); const before = readFileSync(join(f.cwd, first.event));
  assert.throws(() => writeImmutable(f.cwd, first.event, { injected: true }), /already exists/);
  f.register(); assert.deepEqual(readFileSync(join(f.cwd, first.event)), before);
});
test('source revision or claim changes require a new run and explicit source selection', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'first'); finishCLI(f);
  const revision = f.updateBacklog('tasks/CF-101-fixture.md', record('CF-101', 'in-progress', 'New synthetic claim evidence.')); f.register();
  f.reject(/select one registered source/, 'start', '--id', 'CF-101', '--files', 'file.txt', '--summary', 'Fresh source', '--next', 'Validate');
  f.pass('start', '--id', 'CF-101', '--source', revision, '--repository', REPO, '--branch', 'codex/fixture', '--files', 'file.txt', '--summary', 'Fresh source', '--next', 'Validate');
  finishCLI(f); checkCLI(f);
});
test('registered source refuses a closed duplicate canonical ID without writing a snapshot', (t) => {
  const f = cliFixture(t); const revision = f.updateBacklog('tasks/CF-101-unrelated-recovery.md', record('CF-101', 'done'));
  f.reject(/ambiguous/, 'register', '--id', 'CF-101', '--backlog', f.backlog, '--ref', revision, '--record', 'tasks/CF-101-fixture.md');
  assert.equal(f.memory().sources.size, 0);
});
test('scope expansion needs checkpoint; abandoned work does not cover edits', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base', 'other.txt': 'base' }); f.start(); f.write('other.txt', 'changed'); f.stage('other.txt');
  f.reject(/scope/, 'complete', '--id', 'CF-101', '--files', 'other.txt', '--summary', 'Done', '--next', 'Review', '--validation', 'Synthetic checks passed');
  f.pass('checkpoint', '--id', 'CF-101', '--files', 'file.txt', 'other.txt', '--summary', 'Expanded scope', '--next', 'Check');
  f.pass('abandon', '--id', 'CF-101', '--summary', 'Stopped synthetic work', '--next', 'Owner decision'); f.commit('other.txt'); missingCLI(f);
  f.start('CF-101', ['other.txt']); finishCLI(f, ['other.txt']); checkCLI(f);
});
test('status and reconcile remain usable for same-run fork; reconciliation snapshots only current scope', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base', 'other.txt': 'base' }); const first = f.start('CF-101', ['file.txt', 'other.txt']);
  f.write('file.txt', 'changed'); const completed = finishCLI(f, ['file.txt', 'other.txt']);
  const fork = { ...f.readEvent(first), eventId: randomUUID(), kind: 'checkpoint', parents: [first.eventId], sequence: 1, summary: 'Conflicting synthetic branch' };
  writeImmutable(f.cwd, eventPath(fork), fork); f.commit();
  assert.equal(JSON.parse(f.pass('status').stdout).tasks[0].conflict, true);
  f.reject(/unresolved frontier/, 'validate', '--base', f.base, '--repository', REPO);
  f.reject(/unresolved task frontier/, 'start', '--id', 'CF-101', '--files', 'file.txt', '--summary', 'No winner', '--next', 'Continue');
  const reconciled = JSON.parse(f.pass('reconcile', '--id', 'CF-101', '--repository', REPO, '--branch', 'codex/fixture', '--files', 'file.txt', '--summary', 'Reviewed both synthetic heads; dropped unchanged other path', '--next', 'Revalidate final file').stdout);
  assert.deepEqual(new Set(f.readEvent(reconciled).parents), new Set([completed.eventId, fork.eventId]));
  f.commit(); missingCLI(f); const final = finishCLI(f); checkCLI(f);
  assert.deepEqual(Object.keys(f.readEvent(final).snapshot), ['file.txt']);
});
test('independent task histories merge without a shared append-file conflict', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base', 'other.txt': 'base' }); f.git('branch', 'other-lane');
  f.start(); f.write('file.txt', 'lane one'); finishCLI(f); const originalBranch = f.git('symbolic-ref', '--short', 'HEAD');
  f.git('checkout', '-q', 'other-lane'); f.start('CF-102', ['other.txt']); f.write('other.txt', 'lane two'); finishCLI(f, ['other.txt'], 'CF-102');
  f.git('checkout', '-q', originalBranch); f.git('merge', '--no-ff', '-qm', 'merge independent synthetic histories', 'other-lane'); checkCLI(f);
  assert.equal(f.memory().graph.tasks.size, 2);
});
test('old active tasks are advisory and do not lock unrelated covered tasks', (t) => {
  const f = cliFixture(t); const old = f.start('CF-103', ['other.txt']); f.rewriteEvent(old, (event) => ({ ...event, at: '2000-01-01T00:00:00.000Z' }));
  f.start(); f.write('file.txt', 'changed'); finishCLI(f); checkCLI(f);
  const state = JSON.parse(f.pass('status', '--stale-days', '1').stdout); assert.equal(state.tasks.find((task) => task.id === 'CF-103').heads[0].stale, true);
});
for (const path of ['café.txt', 'space name.txt', 'line\nbreak.txt', '[literal].txt', '\ufffd.txt', '\ufeffleading.txt', 'tab\tfile.txt', ':(glob)*.txt', ' trailing space ']) {
  test(`literal UTF-8 Git path ${JSON.stringify(path)} remains exact`, (t) => {
    const f = cliFixture(t, { [path]: 'base' }); f.start('CF-101', [path]); f.write(path, 'changed'); finishCLI(f, [path]); checkCLI(f);
  });
}
test('invalid UTF-8 filename cannot become a replacement-character deletion', (t) => {
  const f = cliFixture(t); f.start('CF-101', ['\ufffd.txt']);
  writeFileSync(Buffer.concat([Buffer.from(f.cwd + '/'), Buffer.from([0xff]), Buffer.from('.txt')]), 'uncovered');
  f.git('add', '-A'); f.complete(); f.commit();
  f.reject(/non-UTF-8/, 'validate', '--base', f.base, '--repository', REPO);
});
test('rename needs old and new path evidence, with an explicit deletion snapshot', (t) => {
  const f = cliFixture(t); f.start('CF-101', ['file.txt', 'renamed.txt']); f.git('mv', 'file.txt', 'renamed.txt');
  const e = finishCLI(f, ['file.txt', 'renamed.txt']); checkCLI(f); assert.deepEqual(f.readEvent(e).snapshot['file.txt'], { blob: null, mode: null });
});
test('only a rename destination cannot cover the deleted source path', (t) => {
  const f = cliFixture(t); f.start('CF-101', ['renamed.txt']); f.git('mv', 'file.txt', 'renamed.txt'); finishCLI(f, ['renamed.txt']); missingCLI(f);
});
test('file-to-directory and directory-to-file refactors retain exact coverage', (t) => {
  const f = cliFixture(t); f.start('CF-101', ['file.txt', 'file.txt/child']); f.git('rm', 'file.txt'); f.write('file.txt/child', 'child');
  finishCLI(f, ['file.txt', 'file.txt/child']); checkCLI(f); f.base = f.git('rev-parse', 'HEAD');
  f.start('CF-101', ['file.txt/child', 'file.txt']); f.git('rm', 'file.txt/child'); f.write('file.txt', 'flat again');
  finishCLI(f, ['file.txt/child', 'file.txt']); checkCLI(f);
});
test('deleted child behind a new symlink ancestor is a deletion, not an untracked target', (t) => {
  const f = cliFixture(t, { 'dir/child': 'base', 'target/child': 'target' }); f.start('CF-101', ['dir/child', 'dir']);
  f.git('rm', 'dir/child'); symlinkSync('target', join(f.cwd, 'dir')); f.stage('dir'); const e = f.complete(); f.commit(); checkCLI(f);
  assert.deepEqual(f.readEvent(e).snapshot['dir/child'], { blob: null, mode: null });
});
test('staged CRLF normalization and symlink modes match committed Git representation', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base\n', '.gitattributes': '*.txt text eol=lf\n' }); f.start('CF-101', ['file.txt', 'link.txt']);
  f.write('file.txt', 'changed\r\n'); symlinkSync('file.txt', join(f.cwd, 'link.txt')); const e = finishCLI(f, ['file.txt', 'link.txt']); checkCLI(f);
  assert.equal(f.readEvent(e).snapshot['link.txt'].mode, '120000');
});
test('new, ignored and partially staged paths cannot be falsely completed', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base', '.gitignore': 'ignored.txt\n' }); f.start('CF-101', ['ignored.txt']); f.write('ignored.txt', 'untracked');
  f.reject(/stage the new file/, 'complete', '--id', 'CF-101', '--summary', 'Done', '--next', 'Review', '--validation', 'Synthetic checks');
  f.pass('checkpoint', '--id', 'CF-101', '--files', 'file.txt', '--summary', 'Change scope', '--next', 'Check');
  f.write('file.txt', 'staged'); f.stage('file.txt'); f.write('file.txt', 'unstaged');
  f.reject(/stage final tested/, 'complete', '--id', 'CF-101', '--summary', 'Done', '--next', 'Review', '--validation', 'Synthetic checks');
});
for (const configuration of ['gitmodules', 'local']) {
  test(`submodule ignore setting in ${configuration} does not hide changed gitlinks`, (t) => {
    const initial = { 'file.txt': 'base' };
    if (configuration === 'gitmodules') initial['.gitmodules'] = '[submodule "mod"]\n\tpath = mod\n\turl = https://example.invalid/mod.git\n\tignore = all\n';
    const f = cliFixture(t, initial); f.git('update-index', '--add', '--cacheinfo', `160000,${f.base},mod`); f.commit(); f.base = f.git('rev-parse', 'HEAD');
    f.git('update-index', '--cacheinfo', `160000,${f.base},mod`); f.commit();
    if (configuration === 'local') f.git('config', 'diff.ignoreSubmodules', 'all'); missingCLI(f);
  });
}
test('gitlink with a valid current completion passes and can use SHA-256 Git objects', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base' }, null, 'sha256'); mkdirSync(join(f.cwd, 'mod'));
  f.git('update-index', '--add', '--cacheinfo', `160000,${f.base},mod`); f.commit(); f.base = f.git('rev-parse', 'HEAD');
  f.start('CF-101', ['mod']); f.git('update-index', '--cacheinfo', `160000,${f.base},mod`); const e = f.complete(); f.commit(); checkCLI(f);
  assert.equal(f.readEvent(e).snapshot.mod.blob.length, 64); assert.equal(f.readEvent(e).snapshot.mod.mode, '160000');
});
test('legacy history is read-only and cannot cover fresh code changes', (t) => {
  const legacy = JSON.stringify({ id: 'historical', event: 'complete', files: ['file.txt'] }) + '\n';
  const f = cliFixture(t, { 'file.txt': 'base' }, legacy); f.write('file.txt', 'changed'); f.commit('file.txt'); missingCLI(f);
  f.start(); finishCLI(f); checkCLI(f); assert.equal(readFileSync(join(f.cwd, '.project-memory/tasks.jsonl'), 'utf8'), legacy);
  const state = JSON.parse(f.pass('status').stdout); assert.equal(state.legacy.readOnly, true); assert.equal(state.legacy.providesV2Coverage, false);
  f.write('.project-memory/tasks.jsonl', legacy + '{}\n'); f.commit(); f.reject(/immutable/, 'validate', '--base', f.base, '--repository', REPO);
});
test('new legacy ledgers are rejected rather than treated as migration evidence', (t) => {
  const f = cliFixture(t); f.write('.project-memory/tasks.jsonl', '{}\n'); f.commit();
  f.reject(/new legacy/, 'validate', '--base', f.base, '--repository', REPO);
});
test('only valid source/event files are exempt; arbitrary memory-directory code still needs evidence', (t) => {
  const f = cliFixture(t); f.write(`${ROOT}/CF-101/hidden.mjs`, 'export const hidden = true;'); f.commit(`${ROOT}/CF-101/hidden.mjs`); missingCLI(f);
});
test('ordinary symlinked code inside a task directory is not read or exempted as metadata', (t) => {
  const f = cliFixture(t); const path = `${ROOT}/CF-101/ordinary-link.js`; f.start('CF-101', [path]);
  symlinkSync('../../../file.txt', join(f.cwd, path)); finishCLI(f, [path]); checkCLI(f);
});
test('metadata-only start/checkpoint commits can preserve a real interrupted task', (t) => {
  const f = cliFixture(t); f.start(); f.commit(); checkCLI(f);
  f.pass('checkpoint', '--id', 'CF-101', '--summary', 'Interrupted synthetic work', '--next', 'Resume'); f.commit(); checkCLI(f);
});
test('committed validation ignores uncommitted source, index and ledger evidence', (t) => {
  const f = cliFixture(t); f.write('file.txt', 'changed'); f.commit('file.txt'); f.start(); f.complete(); missingCLI(f);
  f.commit(); checkCLI(f); f.write('file.txt', 'uncommitted later edit'); f.stage('file.txt'); checkCLI(f);
});
test('target advancement and clean rebases use the actual merge base', (t) => {
  const f = cliFixture(t); f.git('branch', 'target'); f.start(); f.write('file.txt', 'changed'); finishCLI(f); const branch = f.git('symbolic-ref', '--short', 'HEAD');
  f.git('checkout', '-q', 'target'); f.write('target-only.txt', 'new target'); f.commit('target-only.txt'); const target = f.git('rev-parse', 'HEAD');
  f.git('checkout', '-q', branch); f.pass('validate', '--base', target, '--repository', REPO); f.git('rebase', 'target');
  f.pass('validate', '--base', target, '--repository', REPO);
});
test('unknown options, invalid refs and code-context mismatches fail closed', (t) => {
  const f = cliFixture(t); f.reject(/unknown option/, 'status', '--typo', 'x'); f.reject(/needs a value/, 'validate', '--base');
  assert.equal(f.cli('validate', '--base', 'missing').status, 1); f.start(); f.write('file.txt', 'changed'); finishCLI(f);
  f.reject(/repository/, 'validate', '--base', f.base, '--repository', 'other/app');
  assert.match(f.pass('validate', '--base', f.base, '--repository', REPO, '--branch', 'integration-branch').stdout, /integration branch "integration-branch"/);
});
test('unsafe symlinked metadata ancestors are rejected before any write', (t) => {
  const f = cliFixture(t); mkdirSync(join(f.root, 'outside')); symlinkSync(join(f.root, 'outside'), join(f.cwd, '.project-memory'));
  f.reject(/real directory|symlink/, 'register', '--id', 'CF-101', '--backlog', f.backlog, '--ref', f.sourceRevision(), '--record', 'tasks/CF-101-fixture.md');
  assert.deepEqual(readdirSync(join(f.root, 'outside')), []);
});
test('claim digests bind exact CRLF bytes and stop at the next same-level heading', () => {
  const text = record().replaceAll('\n', '\r\n') + '\r\n## Acceptance\r\nOther text.\r\n';
  const section = text.slice(text.indexOf('## Claim'), text.indexOf('## Acceptance'));
  assert.equal(claimSections(text).get('Claim').digest, digest(section));
});
test('offline validation binds task sources to the configured canonical backlog repository', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'changed'); finishCLI(f);
  f.reject(/backlog repository/, 'validate', '--base', f.base, '--repository', REPO, '--backlog-repository', 'wrong/backlog');
});
test('concurrent source registration and separate-task starts do not require a shared lock', async (t) => {
  const f = cliFixture(t);
  const child = (args) => new Promise((resolveResult) => {
    const process = spawn(globalThis.process.execPath, [SCRIPT, ...args], { cwd: f.cwd, env: ENV });
    let stdout = ''; let stderr = '';
    process.stdout.on('data', (chunk) => { stdout += chunk; }); process.stderr.on('data', (chunk) => { stderr += chunk; });
    process.on('close', (status) => resolveResult({ status, stdout, stderr }));
  });
  const register = (id) => ['register', '--id', id, '--backlog', f.backlog, '--ref', f.sourceRevision(), '--record', `tasks/${id}-fixture.md`];
  const registration = await Promise.all([child(register('CF-101')), child(register('CF-101')), child(register('CF-102'))]);
  for (const r of registration) assert.equal(r.status, 0, r.stderr);
  const start = (id) => ['start', '--id', id, '--repository', REPO, '--branch', 'codex/fixture', '--files', 'file.txt', '--summary', 'Independent synthetic task', '--next', 'Work'];
  for (const r of await Promise.all([child(start('CF-101')), child(start('CF-102'))])) assert.equal(r.status, 0, r.stderr);
  assert.equal(f.memory().graph.tasks.size, 2);
});

test('accurate origin branches survive integration of independent tasks', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base', 'other.txt': 'base' });
  f.git('branch', 'lane-two'); const integrationBranch = f.git('symbolic-ref', '--short', 'HEAD');
  f.register(); f.pass('start', '--id', 'CF-101', '--repository', REPO, '--files', 'file.txt', '--summary', 'Task one', '--next', 'Implement');
  f.write('file.txt', 'one'); const first = finishCLI(f);
  f.git('checkout', '-q', 'lane-two'); f.register('CF-102');
  f.pass('start', '--id', 'CF-102', '--repository', REPO, '--files', 'other.txt', '--summary', 'Task two', '--next', 'Implement');
  f.write('other.txt', 'two'); const second = finishCLI(f, ['other.txt'], 'CF-102');
  f.git('checkout', '-q', integrationBranch); f.git('merge', '--no-ff', '-qm', 'integrate independent tasks', 'lane-two');
  checkCLI(f, '--branch', integrationBranch);
  assert.equal(f.readEvent(first).branch, integrationBranch); assert.equal(f.readEvent(second).branch, 'lane-two');
});
for (const wrapper of [
  (text) => '```markdown\n' + text + '\n```',
  (text) => '~~~markdown\n' + text + '\n~~~',
  (text) => '<!--\n' + text + '\n-->',
  (text) => '    ' + text.replaceAll('\n', '\n    '),
]) {
  test(`example-only claims are ignored: ${wrapper('example').split('\n')[0]}`, (t) => {
    const f = cliFixture(t);
    const text = record().replace('## Claim\n\nFixture-only claim for isolated tests.', '## Examples\n\n' + wrapper('## Claim\n\nExample only.'));
    const revision = f.updateBacklog('tasks/CF-101-fixture.md', text);
    f.reject(/claim/i, 'register', '--id', 'CF-101', '--backlog', f.backlog, '--ref', revision, '--record', 'tasks/CF-101-fixture.md');
  });
}
test('fenced and hidden headings retain exact real claim digest and boundaries', () => {
  const text = record().replace('Fixture-only claim for isolated tests.', 'Owner and scope.\n\n````sh\n## Not a heading\n```\n## Still inside longer fence\n````\n<!--\n## Hidden heading\n-->\nRemaining claim.')
    + '\n## Acceptance\nActual next section.\n';
  assert.equal(claimSections(text).get('Claim').digest, digest(text.slice(text.indexOf('## Claim'), text.indexOf('## Acceptance'))));
});
test('local scans ignore temporary publications before stat while retaining recognized metadata checks', (t) => {
  const f = cliFixture(t); f.register();
  const temporary = `${ROOT}/CF-101/sources/${f.sourceRevision()}.json.${uuid(99)}.tmp`; f.write(temporary, 'pending publication');
  const original = fs.lstatSync; let inspected = false;
  fs.lstatSync = (path, ...args) => {
    if (path === join(f.cwd, temporary)) { inspected = true; fs.unlinkSync(path); }
    return original(path, ...args);
  };
  syncBuiltinESMExports();
  try { assert.equal(loadMemory(localMemory(f.cwd)).sources.size, 1); assert.equal(inspected, false); }
  finally { fs.lstatSync = original; syncBuiltinESMExports(); }
  const recognized = join(f.cwd, [...f.memory().sources.keys()][0]);
  fs.lstatSync = (path, ...args) => { if (path === recognized) fs.unlinkSync(path); return original(path, ...args); };
  syncBuiltinESMExports();
  try { assert.throws(() => localMemory(f.cwd), /ENOENT/); }
  finally { fs.lstatSync = original; syncBuiltinESMExports(); }
});
test('transport denial prevents missing-object fetch even without newer lazy-fetch suppression', (t) => {
  const f = cliFixture(t); const blob = f.git('rev-parse', 'HEAD:file.txt');
  const remote = join(f.root, 'local-promisor.git'); f.git('clone', '--bare', '--no-hardlinks', f.cwd, remote);
  f.git('remote', 'add', 'origin', `file://${remote}`); f.git('config', 'remote.origin.promisor', 'true');
  f.git('config', 'remote.origin.partialclonefilter', 'blob:none'); f.git('config', 'protocol.file.allow', 'always');
  const missing = join(f.cwd, '.git/objects', blob.slice(0, 2), blob.slice(2)); rmSync(missing);
  const original = childProcess.execFileSync;
  childProcess.execFileSync = (command, args, options) => {
    assert.equal(args.includes('--no-lazy-fetch'), false, 'remain compatible with Git 2.43');
    const env = { ...options.env }; delete env.GIT_NO_LAZY_FETCH;
    return original(command, args, { ...options, env });
  };
  syncBuiltinESMExports();
  try { assert.throws(() => createGit(f.cwd).readBlob(blob), /Command failed/); }
  finally { childProcess.execFileSync = original; syncBuiltinESMExports(); }
  assert.equal(fs.existsSync(missing), false);
  // Prove this fixture really can lazy-fetch the exact missing blob when the
  // local-file transport is allowed; no external network is involved.
  const env = { ...ENV, GIT_ALLOW_PROTOCOL: 'file' }; delete env.GIT_NO_LAZY_FETCH;
  const control = spawnSync('git', ['cat-file', 'blob', blob], { cwd: f.cwd, env, encoding: 'utf8' });
  assert.equal(control.status, 0, control.stderr); assert.equal(control.stdout, 'base\n');
});

// Generic integration and v1 continuity regressions.
for (const id of ['CF-001', 'TASK-7', 'CF-project-memory-evidence', 'TEAM2-build-tools']) {
  test(`generic task ID ${id} binds exact canonical source and metadata paths`, () => {
    const s = source(id); const event = ev(s, 1);
    assert.equal(validateSource(s).id, id); assert.equal(TASK.test(id), true);
    assert.equal(metadataPath(sourcePath(s)), true); assert.equal(metadataPath(eventPath(event)), true);
  });
}
for (const id of ['../CF-1', 'CF-1/escape', 'cf-1', 'CF-', 'CF--1', 'CF-1.', 'CF-1\\other', 'CF-' + 'a'.repeat(78)]) {
  test(`unsafe or noncanonical task ID ${JSON.stringify(id)} is refused`, () => {
    assert.equal(TASK.test(id), false); assert.throws(() => source(id), /identity/);
  });
}
test('different valid generic IDs may share a filename prefix without identity collisions', () => {
  const records = new Map([
    ['tasks/CF-build-fixture.md', entry(record('CF-build'))],
    ['tasks/CF-build-tools-fixture.md', entry(record('CF-build-tools'))],
  ]);
  for (const id of ['CF-build', 'CF-build-tools']) {
    assert.equal(makeSource({ repository: BACKLOG, revision: REV, id, path: `tasks/${id}-fixture.md`, records }).id, id);
  }
  records.set('tasks/CF-build-impostor.md', entry(record('CF-unrelated')));
  assert.throws(() => makeSource({ repository: BACKLOG, revision: REV, id: 'CF-build', path: 'tasks/CF-build-fixture.md', records }), /conflicting/);
});
test('generic slug IDs support source registration, lifecycle and committed coverage', (t) => {
  const f = cliFixture(t); const id = 'CF-project-memory-evidence';
  f.updateBacklog(`tasks/${id}-fixture.md`, record(id)); f.start(id); f.write('file.txt', 'changed');
  finishCLI(f, ['file.txt'], id); checkCLI(f);
});
test('source aliases are explicit and simultaneous aliases are refused without metadata writes', (t) => {
  const f = cliFixture(t);
  const args = ['register', '--id', 'CF-101', '--ref', f.sourceRevision(), '--record', 'tasks/CF-101-fixture.md'];
  f.reject(/only --source-repo/, ...args, '--source-repo', f.backlog, '--backlog', f.backlog);
  f.reject(/only --source-repository/, ...args, '--source-repo', f.backlog, '--source-repository', BACKLOG, '--backlog-repository', BACKLOG);
  assert.equal(f.memory().sources.size, 0);
  f.pass(...args, '--source-repo', f.backlog, '--source-repository', BACKLOG);
});
test('source origin mismatch is refused even with explicit source identity', (t) => {
  const f = cliFixture(t);
  f.reject(/disagrees/, 'register', '--id', 'CF-101', '--source-repo', f.backlog, '--ref', f.sourceRevision(), '--record', 'tasks/CF-101-fixture.md', '--source-repository', 'other/tasks');
  assert.equal(f.memory().sources.size, 0);
});
test('no source origin requires an explicit declared identity', (t) => {
  const f = cliFixture(t); f.gitAt(f.backlog, 'remote', 'remove', 'origin');
  const args = ['register', '--id', 'CF-101', '--source-repo', f.backlog, '--ref', f.sourceRevision(), '--record', 'tasks/CF-101-fixture.md'];
  f.reject(/provide --source-repository/, ...args);
  f.pass(...args, '--source-repository', BACKLOG);
});
test('code identity is explicit without a remote and binds recognized origin', (t) => {
  const f = cliFixture(t); f.register();
  const args = ['start', '--id', 'CF-101', '--files', 'file.txt', '--summary', 'Synthetic scope', '--next', 'Check'];
  f.reject(/provide --repository/, ...args);
  f.git('remote', 'add', 'origin', `https://github.com/${REPO}.git`);
  f.reject(/disagrees/, ...args, '--repository', 'other/app');
  f.pass(...args); f.write('file.txt', 'changed'); finishCLI(f); checkCLI(f);
});
test('validation never guesses a canonical source repository', (t) => {
  const f = cliFixture(t); f.start(); f.write('file.txt', 'changed'); finishCLI(f);
  const r = spawnSync(process.execPath, [SCRIPT, 'validate', '--base', f.base, '--repository', REPO], { cwd: f.cwd, env: ENV, encoding: 'utf8' });
  assert.equal(r.status, 1); assert.match(r.stderr, /--source-repository is required/);
});
test('malformed JSON and unsupported source/event versions fail closed', (t) => {
  const f = cliFixture(t); const started = f.start(); const sourceFile = [...f.memory().sources.keys()][0];
  const originalSource = readFileSync(join(f.cwd, sourceFile), 'utf8');
  f.write(sourceFile, '{invalid'); f.reject(/invalid JSON/, 'status');
  f.write(sourceFile, JSON.stringify({ ...JSON.parse(originalSource), sourceVersion: 2 })); f.reject(/identity/, 'status');
  f.write(sourceFile, originalSource);
  f.rewriteEvent(started, (event) => ({ ...event, formatVersion: 3 })); f.reject(/invalid event/, 'status');
});
test('mixed v1 and unversioned legacy rows remain exact read-only bytes without newline repair', (t) => {
  const legacy = JSON.stringify({ id: 'CF-old', event: 'start', files: ['old.txt'] }) + '\r\n'
    + JSON.stringify({ version: 1, id: 'CF-old', event: 'complete', files: ['old.txt'] });
  const f = cliFixture(t, { 'file.txt': 'base' }, legacy);
  f.start(); f.write('file.txt', 'changed'); finishCLI(f); checkCLI(f);
  assert.equal(readFileSync(join(f.cwd, '.project-memory/tasks.jsonl'), 'utf8'), legacy);
  const status = JSON.parse(f.pass('status').stdout);
  assert.deepEqual(status.legacy, { path: '.project-memory/tasks.jsonl', records: 2, readOnly: true, providesV2Coverage: false });
});
test('Git replacement objects cannot change the recorded source bytes', (t) => {
  const f = cliFixture(t); const original = f.git('rev-parse', 'HEAD:file.txt');
  f.write('replacement.txt', 'replacement'); const replacement = f.git('hash-object', '-w', 'replacement.txt');
  f.git('replace', original, replacement);
  assert.equal(f.git('cat-file', 'blob', original), 'replacement');
  assert.equal(createGit(f.cwd).readBlob(original), 'base\n');
});

test('a forged single-parent restart cannot silently supersede an active run', () => {
  const s = source(); const active = ev(s, 1);
  const restarted = ev(s, 2, 'start', null, { parents: [active.eventId] });
  assert.throws(() => graph([active, restarted], s), /active head/);
});
for (const path of ['../outside', '/absolute', 'C:/absolute', 'nested/../../outside', '.project-memory/tasks.jsonl']) {
  test(`unsafe/self-referential file scope ${JSON.stringify(path)} cannot become evidence`, (t) => {
    const f = cliFixture(t); f.register();
    f.reject(/paths/, 'start', '--id', 'CF-101', '--repository', REPO, '--files', path, '--summary', 'Invalid scope', '--next', 'Stop');
    assert.equal(f.memory().events.length, 0);
  });
}
test('directory scope cannot cover an unlisted changed child', (t) => {
  const f = cliFixture(t); f.start('CF-101', ['directory']); f.write('directory/new.txt', 'uncovered'); f.stage('directory/new.txt');
  f.complete(); f.commit('directory/new.txt'); missingCLI(f);
});
test('invalid UTF-8 legacy bytes fail closed without rewrite', (t) => {
  const f = cliFixture(t, { 'file.txt': 'base' }, '{}\n');
  writeFileSync(join(f.cwd, '.project-memory/tasks.jsonl'), Buffer.from([0xff, 0x0a]));
  f.reject(/non-UTF-8/, 'status');
  assert.deepEqual(readFileSync(join(f.cwd, '.project-memory/tasks.jsonl')), Buffer.from([0xff, 0x0a]));
});

test('ancestor padding cannot disguise a single active head as competing reconciliation heads', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'checkpoint', a);
  for (const parents of [[a.eventId, b.eventId], [b.eventId, a.eventId]]) {
    const c = ev(s, 3, 'start', null, { parents }); const d = ev(s, 4, 'complete', c);
    assert.throws(() => graph([a, b, c, d], s), /competing heads/);
  }
  const distant = ev(s, 5, 'checkpoint', b);
  const padded = ev(s, 6, 'start', null, { parents: [a.eventId, distant.eventId] });
  assert.throws(() => graph([a, b, distant, padded], s), /competing heads/);
});
test('genuine fork heads remain reconcilable after a shared checkpoint ancestor', () => {
  const s = source(); const a = ev(s, 1); const b = ev(s, 2, 'checkpoint', a);
  const left = ev(s, 3, 'checkpoint', b); const right = ev(s, 4, 'complete', b);
  const reconciled = ev(s, 5, 'start', null, { parents: [left.eventId, right.eventId] });
  const final = ev(s, 6, 'complete', reconciled);
  assert.deepEqual(completionHeads(graph([a, b, left, right, reconciled, final], s), new Set([final.eventId])), [final]);
});
for (const id of ['CF-101 # comment', '"CF-101" # comment', 'CF-101 trailing', '', 'CF-101/other']) {
  test(`unsupported source ID syntax ${JSON.stringify(id)} cannot hide in another filename`, () => {
    const records = new Map([
      ['tasks/CF-101-fixture.md', entry(record())],
      ['tasks/OTHER-1-duplicate.md', entry(record(id))],
    ]);
    assert.throws(() => makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records }), /unsupported canonical task ID/);
  });
}
test('quoted duplicate IDs are still detected across unrelated filenames', () => {
  const records = new Map([
    ['tasks/CF-101-fixture.md', entry(record())],
    ['tasks/OTHER-1-duplicate.md', entry(record().replace('id: CF-101', 'id: "CF-101"'))],
  ]);
  assert.throws(() => makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records }), /disagrees|ambiguous/);
});

test('source identity fields reject coercible arrays instead of accepting regex conversions', () => {
  for (const key of ['repository', 'revision', 'id', 'path', 'blob']) {
    const s = source(); s[key] = [s[key]];
    assert.throws(() => validateSource(s), /identity/);
  }
});
test('event and parent UUID fields require strings rather than coercible arrays', () => {
  const s = source(); const start = ev(s, 1);
  for (const key of ['eventId', 'runId']) {
    const bad = { ...start, [key]: [start[key]] };
    assert.throws(() => validateEvent(bad, sources(s)), /identity|causal/);
  }
  const badParent = ev(s, 2, 'checkpoint', start, { parents: [[start.eventId]] });
  assert.throws(() => validateEvent(badParent, sources(s)), /identity|causal/);
});

for (const key of ['path', 'blob', 'record', 'targetPath', 'targetId', 'targetBlob']) {
  test(`redirect ${key} rejects non-string identity values`, () => {
    const text = '# Historical task — noncanonical redirect\n\nThe canonical record is [CF-102](CF-102-other.md).\n';
    const records = new Map([
      ['tasks/CF-101-fixture.md', entry(record())],
      ['tasks/CF-101-old.md', entry(text)],
      ['tasks/CF-102-other.md', entry(record('CF-102'))],
    ]);
    const s = makeSource({ repository: BACKLOG, revision: REV, id: 'CF-101', path: 'tasks/CF-101-fixture.md', records });
    assert.equal(validateSource(s), s);
    for (const value of [[s.identityEvidence.redirects[0][key]], null, {}, 42, true]) {
      const bad = structuredClone(s); bad.identityEvidence.redirects[0][key] = value;
      assert.throws(() => validateSource(bad), undefined, `${key} must reject ${JSON.stringify(value)}`);
    }
  });
}
