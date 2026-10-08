import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync, chmodSync, symlinkSync, lstatSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const SCRIPT = resolve(dirname(fileURLToPath(import.meta.url)), '../skills/project-memory/scripts/project-memory.mjs');
const LEDGER = '.project-memory/tasks.jsonl';
const env = { ...process.env, GIT_CONFIG_NOSYSTEM: '1', GIT_CONFIG_GLOBAL: '/dev/null' };
const fixture = (t, initial = { 'file.txt': 'base\n' }, history = '') => {
  const cwd = mkdtempSync(join(tmpdir(), 'project-memory-test-'));
  t.after(() => rmSync(cwd, { recursive: true, force: true }));
  const command = (exe, args) => spawnSync(exe, args, { cwd, env, encoding: 'utf8' });
  const git = (...args) => {
    const result = command('git', ['--literal-pathspecs', ...args]);
    assert.equal(result.status, 0, result.stderr);
    return result.stdout.trim();
  };
  const write = (path, content) => { mkdirSync(dirname(join(cwd, path)), { recursive: true }); writeFileSync(join(cwd, path), content); };
  const stage = (...paths) => {
    for (const path of paths) {
      let present = false;
      try { lstatSync(join(cwd, path)); present = true; } catch (error) { if (!['ENOENT', 'ENOTDIR'].includes(error.code)) throw error; }
      if (present || git('ls-files', '--', path)) git('add', '-A', '--', path);
    }
  };
  const commit = (...paths) => { stage(...paths); git('commit', '-qm', 'fixture'); return git('rev-parse', 'HEAD'); };
  const cli = (...args) => command(process.execPath, [SCRIPT, ...args]);
  const pass = (...args) => { const result = cli(...args); assert.equal(result.status, 0, result.stderr); return result; };
  const fail = (pattern, ...args) => { const result = cli(...args); assert.equal(result.status, 1, result.stdout); assert.match(result.stderr, pattern); return result; };
  const events = () => readFileSync(join(cwd, LEDGER), 'utf8').trim().split('\n').filter(Boolean).map(JSON.parse);
  const replace = (items) => write(LEDGER, items.map((event) => JSON.stringify(event)).join('\n') + '\n');
  const start = (id = 'T-1', files = ['file.txt']) => pass('start', '--id', id, '--summary', 'Scope', '--next', 'Implement', '--files', ...files);
  const complete = (id = 'T-1', files) => pass('complete', '--id', id, '--summary', 'Validated fixture', '--next', 'Review', '--validation', 'fixture checks passed', ...(files ? ['--files', ...files] : []));
  git('init', '-q'); git('config', 'user.name', 'Fixture'); git('config', 'user.email', 'fixture@example.invalid');
  git('config', 'commit.gpgsign', 'false'); git('config', 'core.autocrlf', 'false'); git('config', 'core.hooksPath', '.empty-hooks');
  for (const [path, value] of Object.entries(initial)) write(path, value);
  if (history) write(LEDGER, history);
  const base = commit(...Object.keys(initial), ...(history ? [LEDGER] : []));
  return { cwd, git, write, stage, commit, cli, pass, fail, events, replace, start, complete, base };
};
const finish = (f, files = ['file.txt']) => { f.stage(...files); f.complete('T-1', files); f.commit(...files, LEDGER); };
const check = (f) => f.pass('validate', '--base', f.base);
const reject = (f) => f.fail(/fresh final completion/, 'validate', '--base', f.base);

test('fresh start, staged completion and committed validation pass', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'changed\n'); finish(f); check(f);
  const [start, complete] = f.events();
  assert.equal(start.version, 1); assert.equal(complete.snapshot['file.txt'].blob, f.git('rev-parse', 'HEAD:file.txt'));
  assert.equal(complete.validation, 'fixture checks passed');
});
test('uncovered code change fails', (t) => { const f = fixture(t); f.write('file.txt', 'changed'); f.commit('file.txt'); reject(f); });
test('historical complete cannot cover a new change', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f); f.base = f.git('rev-parse', 'HEAD');
  f.write('file.txt', 'second'); f.commit('file.txt'); reject(f);
});
test('new complete for historical open start can close interrupted work', (t) => {
  const f = fixture(t); f.start(); f.commit(LEDGER); f.base = f.git('rev-parse', 'HEAD');
  f.write('file.txt', 'change'); finish(f); check(f);
});
test('start complete start is unfinished and cannot cover the diff', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f); f.start(); f.commit(LEDGER); reject(f);
});
test('completion after restart refreshes coverage', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f); f.start(); f.write('file.txt', 'second'); finish(f); check(f);
});
test('later edit to the same path invalidates completion fingerprint', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f); f.write('file.txt', 'later'); f.commit('file.txt'); reject(f);
});
test('mode changes invalidate completion fingerprint', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f); chmodSync(join(f.cwd, 'file.txt'), 0o755); f.commit('file.txt'); reject(f);
});
test('a later completed task can update the same file without stale-task noise', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f);
  f.start('T-2'); f.write('file.txt', 'second'); f.git('add', '--', 'file.txt'); f.complete('T-2'); f.commit('file.txt', LEDGER); check(f);
});
test('final completion alone supplies coverage, not older task cycles', (t) => {
  const f = fixture(t, { 'file.txt': 'base', 'other.txt': 'base' });
  f.start('T-1', ['file.txt', 'other.txt']); f.write('file.txt', 'first'); finish(f, ['file.txt', 'other.txt']);
  f.start('T-1', ['file.txt', 'other.txt']); f.git('add', '--', 'other.txt'); f.complete('T-1', ['other.txt']); f.commit(LEDGER); reject(f);
});
test('historical ledger edits, deletion and reordering fail', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f); f.base = f.git('rev-parse', 'HEAD');
  const original = f.events();
  for (const events of [[...original].reverse(), [], original.map((event) => ({ ...event, summary: 'Changed history' }))]) {
    f.replace(events); f.commit(LEDGER); f.fail(/historical bytes/, 'validate', '--base', f.base);
  }
});
test('a forged append cannot finish a task without a start', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f);
  f.replace([f.events()[1]]); f.commit(LEDGER); f.fail(/active start/, 'validate', '--base', f.base);
});
test('completion before start timestamp fails', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'first'); finish(f);
  const events = f.events(); events[1].at = '2000-01-01T00:00:00.000Z'; f.replace(events); f.commit(LEDGER);
  f.fail(/timestamp precedes/, 'validate', '--base', f.base);
});
test('duplicate starts and completions fail without appending', (t) => {
  const f = fixture(t); f.start(); const first = f.events().length;
  f.fail(/already active/, 'start', '--id', 'T-1', '--summary', 'Again', '--next', 'Continue'); assert.equal(f.events().length, first);
  f.write('file.txt', 'first'); finish(f);
  f.fail(/active start/, 'complete', '--id', 'T-1', '--summary', 'Again', '--next', 'Review', '--validation', 'checks passed');
});
test('scope expansion requires a checkpoint before completion', (t) => {
  const f = fixture(t, { 'file.txt': 'base', 'other.txt': 'base' }); f.start(); f.write('other.txt', 'changed'); f.git('add', '--', 'other.txt');
  f.fail(/exceed active scope/, 'complete', '--id', 'T-1', '--summary', 'Done', '--next', 'Review', '--validation', 'passed', '--files', 'other.txt');
  f.pass('checkpoint', '--id', 'T-1', '--summary', 'Expanded scope', '--next', 'Validate', '--files', 'file.txt', 'other.txt');
  f.complete('T-1'); f.commit('other.txt', LEDGER); check(f);
});
test('abandoned work does not provide coverage', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'changed');
  f.pass('abandon', '--id', 'T-1', '--summary', 'Stopped', '--next', 'Owner review'); f.commit('file.txt', LEDGER); reject(f);
});
test('abandoned work can be restarted', (t) => {
  const f = fixture(t); f.start(); f.pass('abandon', '--id', 'T-1', '--summary', 'Paused decision', '--next', 'Review');
  f.start(); f.write('file.txt', 'changed'); finish(f); check(f);
});
test('stale open tasks are visible without blocking unrelated completed work', (t) => {
  const f = fixture(t); f.start('OLD', ['other.txt']); const events = f.events(); events[0].at = '2000-01-01T00:00:00.000Z'; f.replace(events);
  f.start(); f.write('file.txt', 'changed'); finish(f); check(f);
  const status = JSON.parse(f.pass('status', '--stale-days', '1').stdout);
  assert.equal(status.find((item) => item.id === 'OLD').stale, true);
  assert.equal(status.find((item) => item.id === 'OLD').state, 'active');
});
for (const path of ['café.txt', 'space name.txt', 'line\nbreak.txt', '[literal].txt', '\ufffd.txt']) {
  test(`literal path ${JSON.stringify(path)} passes`, (t) => {
    const f = fixture(t, { [path]: 'base' }); f.start('T-1', [path]); f.write(path, 'changed'); finish(f, [path]); check(f);
  });
}
test('rename requires old and new file coverage, including deletion snapshot', (t) => {
  const f = fixture(t); f.start('T-1', ['file.txt', 'renamed.txt']); f.git('mv', 'file.txt', 'renamed.txt'); finish(f, ['file.txt', 'renamed.txt']); check(f);
  assert.deepEqual(f.events()[1].snapshot['file.txt'], { mode: null, blob: null });
});
test('rename with only new-name coverage fails', (t) => {
  const f = fixture(t); f.start('T-1', ['renamed.txt']); f.git('mv', 'file.txt', 'renamed.txt'); finish(f, ['renamed.txt']); reject(f);
});
test('deleted files have explicit null snapshot', (t) => {
  const f = fixture(t); f.start(); f.git('rm', 'file.txt'); finish(f); check(f);
  assert.deepEqual(f.events()[1].snapshot['file.txt'], { mode: null, blob: null });
});
test('new paths must be staged before completion', (t) => {
  const f = fixture(t); f.start('T-1', ['new.txt']); f.write('new.txt', 'new');
  f.fail(/stage the new file/, 'complete', '--id', 'T-1', '--summary', 'Done', '--next', 'Review', '--validation', 'passed');
  finish(f, ['new.txt']); check(f);
});
test('unstaged or partially staged paths cannot be completed', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'staged'); f.git('add', '--', 'file.txt'); f.write('file.txt', 'unstaged');
  f.fail(/stage final tested/, 'complete', '--id', 'T-1', '--summary', 'Done', '--next', 'Review', '--validation', 'passed');
});
test('completion snapshots staged CRLF normalization correctly', (t) => {
  const f = fixture(t, { 'file.txt': 'base\n', '.gitattributes': '*.txt text eol=lf\n' }); f.start(); f.write('file.txt', 'changed\r\n'); finish(f); check(f);
});
test('symlink target bytes and mode are represented by the index', (t) => {
  const f = fixture(t); f.start('T-1', ['link.txt']); symlinkSync('file.txt', join(f.cwd, 'link.txt')); finish(f, ['link.txt']); check(f);
  assert.equal(f.events()[1].snapshot['link.txt'].mode, '120000');
});
test('ledger-only checkpoint changes are valid and never self-hashed', (t) => {
  const f = fixture(t); f.start(); f.commit(LEDGER); check(f);
  f.pass('checkpoint', '--id', 'T-1', '--summary', 'Interruption', '--next', 'Resume tomorrow'); f.commit(LEDGER); check(f);
});
test('only the ledger is exempt; other memory-directory files need evidence', (t) => {
  const f = fixture(t); f.write('.project-memory/hidden.mjs', 'export const unsafe = true;'); f.commit('.project-memory/hidden.mjs'); reject(f);
});
test('committed validation ignores forged uncommitted ledger changes', (t) => {
  const f = fixture(t); f.write('file.txt', 'changed'); f.commit('file.txt'); f.start(); f.complete(); reject(f);
});
test('append-only prefix uses merge base when target branch advanced independently', (t) => {
  const f = fixture(t); f.git('branch', 'target'); f.start(); f.write('file.txt', 'changed'); finish(f); const head = f.git('rev-parse', 'HEAD');
  f.git('checkout', '-q', 'target'); f.write('target-only.txt', 'new target'); f.commit('target-only.txt');
  const target = f.git('rev-parse', 'HEAD'); f.git('checkout', '-q', '--detach', head); f.pass('validate', '--base', target);
});
test('unknown arguments and invalid revisions fail closed', (t) => {
  const f = fixture(t); f.fail(/unknown option/, 'status', '--typo', 'x'); f.fail(/needs a value/, 'validate', '--base');
  assert.equal(f.cli('validate', '--base', 'not-a-revision').status, 1);
});
test('bad JSON and malformed v1 schemas fail', (t) => {
  const f = fixture(t); f.write(LEDGER, '{oops\n'); f.commit(LEDGER); f.fail(/invalid JSON/, 'validate', '--base', f.base);
  f.replace([{ version: 1, id: 'T-1', event: 'start', at: '2026-01-01T00:00:00.000Z', summary: {}, next: 'Resume', files: ['file.txt'] }]);
  f.commit(LEDGER); f.fail(/nonempty strings/, 'validate', '--base', f.base);
});
test('repository escape and ledger self-reference paths are rejected', (t) => {
  const f = fixture(t);
  for (const path of ['../outside', '..', LEDGER]) {
    f.fail(/repository-relative/, 'start', '--id', 'T-1', '--summary', 'Scope', '--next', 'Implement', '--files', path);
  }
});
test('legacy baseline remains readable, but fresh legacy events cannot bypass migration', (t) => {
  const legacy = ['start', 'complete'].map((event) => ({ id: 'OLD', event, at: '2026-01-01T00:00:00.000Z', summary: 'Historical', next: 'Review', files: ['file.txt'] }));
  const f = fixture(t, { 'file.txt': 'base' }, legacy.map(JSON.stringify).join('\n') + '\n'); f.pass('status');
  f.start(); f.write('file.txt', 'changed'); finish(f); check(f);
  f.replace([...f.events(), { ...legacy[0], id: 'BYPASS' }]); f.commit(LEDGER); f.fail(/require version 1/, 'validate', '--base', f.base);
});
test('legacy final line without newline can be preserved and appended safely', (t) => {
  const prior = { id: 'OLD', event: 'start', at: '2026-01-01T00:00:00.000Z', summary: 'Historical', next: 'Resume', files: ['other.txt'] };
  const f = fixture(t, { 'file.txt': 'base' }, JSON.stringify(prior)); f.start(); f.write('file.txt', 'changed'); finish(f); check(f);
});
test('canonical task link is retained across lifecycle events and cannot be changed', (t) => {
  const f = fixture(t); const task = 'https://github.com/example/backlog/blob/main/tasks/T-1.md';
  f.pass('start', '--id', 'T-1', '--summary', 'Scope', '--next', 'Implement', '--files', 'file.txt', '--task', task);
  f.write('file.txt', 'changed'); finish(f); check(f); assert.equal(f.events()[1].task, task);
  f.fail(/cannot change/, 'start', '--id', 'T-1', '--summary', 'Resume', '--next', 'Implement', '--task', 'https://example.org/other');
});
test('symlinked local ledger is rejected before append', (t) => {
  const f = fixture(t); f.write('other.txt', 'do not append'); mkdirSync(join(f.cwd, '.project-memory')); symlinkSync('../other.txt', join(f.cwd, LEDGER));
  f.fail(/must not be a symlink/, 'start', '--id', 'T-1', '--summary', 'Scope', '--next', 'Implement', '--files', 'file.txt');
  assert.equal(readFileSync(join(f.cwd, 'other.txt'), 'utf8'), 'do not append');
});
for (const change of [
  (event) => ({ ...event, version: 2 }),
  (event) => ({ ...event, id: 7 }),
  (event) => ({ ...event, files: ['/absolute.txt'] }),
  (event) => ({ ...event, files: ['a/../file.txt'] }),
  (event) => ({ ...event, files: ['file.txt', 'file.txt'] }),
  (event) => ({ ...event, at: 0 }),
  (event) => ({ ...event, at: '2026-02-30T00:00:00.000Z' }),
  (event) => ({ ...event, unexpected: 'ignored state' }),
]) {
  test(`malformed event variant ${change.toString()} fails closed`, (t) => {
    const f = fixture(t); f.start(); f.replace([change(f.events()[0])]); f.commit(LEDGER);
    assert.equal(f.cli('validate', '--base', f.base).status, 1);
  });
}
test('invalid UTF-8 ledger fails rather than comparing replacement characters', (t) => {
  const f = fixture(t); f.write(LEDGER, Buffer.from([0x7b, 0x22, 0xff, 0x22, 0x7d])); f.commit(LEDGER);
  f.fail(/encoded data was not valid/, 'validate', '--base', f.base);
});
test('missing or malformed completion fingerprints cannot provide coverage', (t) => {
  const f = fixture(t); f.start(); f.write('file.txt', 'changed'); finish(f);
  const original = f.events();
  for (const snapshot of [{}, { 'file.txt': { blob: 'wrong', mode: '100644' } }, { 'file.txt': { blob: null, mode: '100644' } }]) {
    f.replace([original[0], { ...original[1], snapshot }]); f.commit(LEDGER);
    assert.equal(f.cli('validate', '--base', f.base).status, 1);
  }
});
test('non-UTF-8 filename cannot masquerade as a replacement-character deletion', (t) => {
  const f = fixture(t); f.start('T-1', ['\ufffd.txt']);
  writeFileSync(Buffer.concat([Buffer.from(f.cwd + '/'), Buffer.from([0xff]), Buffer.from('.txt')]), 'uncovered content');
  f.git('add', '-A'); f.complete(); f.commit(LEDGER);
  f.fail(/unsupported non-UTF-8/, 'validate', '--base', f.base);
});
test('file replaced by directory covers former file deletion and each new child', (t) => {
  const f = fixture(t); const paths = ['file.txt', 'file.txt/child']; f.start('T-1', paths);
  f.git('rm', 'file.txt'); f.write('file.txt/child', 'child'); finish(f, paths); check(f);
  assert.deepEqual(f.events()[1].snapshot['file.txt'], { mode: null, blob: null });
});
test('directory replaced by file covers removed children and new file', (t) => {
  const f = fixture(t, { 'dir/child': 'base' }); const paths = ['dir/child', 'dir']; f.start('T-1', paths);
  f.git('rm', 'dir/child'); f.write('dir', 'file'); finish(f, paths); check(f);
  assert.deepEqual(f.events()[1].snapshot['dir/child'], { mode: null, blob: null });
});
test('recording a directory does not cover unlisted new child paths', (t) => {
  const f = fixture(t); f.start('T-1', ['file.txt']); f.git('rm', 'file.txt'); f.write('file.txt/child', 'child');
  f.stage('file.txt/child'); f.complete(); f.commit(LEDGER); reject(f);
});
for (const configuration of ['gitmodules', 'local']) {
  test(`submodule ignore settings in ${configuration} cannot hide changed gitlinks`, (t) => {
    const initial = { 'file.txt': 'base' };
    if (configuration === 'gitmodules') initial['.gitmodules'] = '[submodule "mod"]\n\tpath = mod\n\turl = https://example.invalid/mod.git\n\tignore = all\n';
    const f = fixture(t, initial); f.git('update-index', '--add', '--cacheinfo', `160000,${f.base},mod`);
    f.git('commit', '-qm', 'initial gitlink'); f.base = f.git('rev-parse', 'HEAD');
    f.git('update-index', '--cacheinfo', `160000,${f.base},mod`); f.git('commit', '-qm', 'changed gitlink');
    if (configuration === 'local') f.git('config', 'diff.ignoreSubmodules', 'all');
    reject(f);
  });
}
test('changed gitlink with current completion evidence passes', (t) => {
  const f = fixture(t); mkdirSync(join(f.cwd, 'mod')); f.git('update-index', '--add', '--cacheinfo', `160000,${f.base},mod`);
  f.git('commit', '-qm', 'initial gitlink'); f.base = f.git('rev-parse', 'HEAD');
  f.start('T-1', ['mod']); f.git('update-index', '--cacheinfo', `160000,${f.base},mod`); f.complete(); f.commit(LEDGER); check(f);
  assert.equal(f.events()[1].snapshot.mod.mode, '160000');
});
test('deleted child behind a replacement symlink parent records a deletion', (t) => {
  const f = fixture(t, { 'dir/child': 'base', 'target/child': 'target' }); f.start('T-1', ['dir/child', 'dir']);
  f.git('rm', 'dir/child'); symlinkSync('target', join(f.cwd, 'dir')); f.stage('dir'); f.complete(); f.commit(LEDGER); check(f);
  assert.deepEqual(f.events()[1].snapshot['dir/child'], { mode: null, blob: null });
});
test('ignored new files still require staging before completion', (t) => {
  const f = fixture(t, { 'file.txt': 'base', '.gitignore': 'ignored.txt\n' }); f.start('T-1', ['ignored.txt']); f.write('ignored.txt', 'not staged');
  f.fail(/stage the new file/, 'complete', '--id', 'T-1', '--summary', 'Done', '--next', 'Review', '--validation', 'passed');
});
