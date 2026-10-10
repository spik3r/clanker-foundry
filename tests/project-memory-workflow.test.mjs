import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync, chmodSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const workflow = readFileSync(join(root, '.github/workflows/validate.yml'), 'utf8');
const job = (name) => {
  const marker = `  ${name}:\n`; const start = workflow.indexOf(marker);
  assert.notEqual(start, -1, `missing job: ${name}`);
  const tail = workflow.slice(start); const next = tail.slice(marker.length).search(/^  [a-z][a-z-]*:\n/m);
  return (next === -1 ? tail : tail.slice(0, marker.length + next)).replace(/\n  #.*\n?$/, '\n').trimEnd();
};
const memory = job('project-memory');
const step = (name) => {
  const marker = `      - name: ${name}\n`; const start = memory.indexOf(marker);
  assert.notEqual(start, -1, `missing memory step: ${name}`);
  const next = memory.indexOf('\n      - ', start + marker.length);
  return memory.slice(start, next === -1 ? memory.length : next);
};

test('workflow retains push, pull request and manual triggers', () => {
  assert.equal(workflow.slice(0, workflow.indexOf('jobs:')).trim(), 'name: validate\n\non:\n  push:\n  pull_request:\n  workflow_dispatch:');
  assert.doesNotMatch(workflow, /pull_request_target:|workflow_run:/);
});
test('pack validator and installer job remain unchanged', () => {
  assert.equal(job('validate'), `  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install GNU Stow
        run: sudo apt-get update && sudo apt-get install -y stow
      - name: Run pack validator
        run: bash scripts/validate.sh
      - name: Run installer regression tests
        run: bash tests/test-install.sh`);
});
test('advisory shellcheck job remains unchanged', () => {
  assert.equal(job('shellcheck'), `  shellcheck:
    runs-on: ubuntu-latest
    continue-on-error: true
    steps:
      - uses: actions/checkout@v4
      - name: Install shellcheck
        run: sudo apt-get update && sudo apt-get install -y shellcheck
      - name: Check scripts
        run: shellcheck scripts/*.sh global-agents/*.sh tests/*.sh`);
});
test('memory job preserves read-only permission, full history and no saved checkout credentials', () => {
  assert.match(memory, /permissions:\n      contents: read\n/);
  assert.match(memory, /uses: actions\/checkout@v4\n        with:\n          fetch-depth: 0\n          persist-credentials: false/);
  assert.doesNotMatch(memory, /contents: write|id-token:|pull-requests:|secrets\.|continue-on-error/);
});
test('memory tests use Node 24 and both actual test files on every trigger', () => {
  assert.match(memory, /uses: actions\/setup-node@v4\n        with:\n          node-version: '24'/);
  const tests = step('Test Project Memory lifecycle rules');
  assert.match(tests, /run: node --test tests\/project-memory\.test\.mjs tests\/project-memory-workflow\.test\.mjs/);
  assert.doesNotMatch(tests, /\n\s+if:|\|\|\s*true/);
  for (const path of ['tests/project-memory.test.mjs', 'tests/project-memory-workflow.test.mjs']) assert.ok(readFileSync(join(root, path), 'utf8').length);
});
test('committed validation uses the actual PR base and explicit repository identities', () => {
  const validate = step('Verify committed pull-request lifecycle evidence');
  assert.match(validate, /if: github\.event_name == 'pull_request'/);
  assert.match(validate, /BASE_SHA: \$\{\{ github\.event\.pull_request\.base\.sha \}\}/);
  assert.match(validate, /CODE_REPOSITORY: \$\{\{ github\.repository \}\}/);
  assert.match(validate, /TASK_SOURCE_REPOSITORY: \$\{\{ github\.repository \}\}/);
  assert.match(validate, /PR_HEAD_REF: \$\{\{ github\.event\.pull_request\.head\.ref \}\}/);
  assert.match(validate, /validate --base "\$BASE_SHA" --repository "\$CODE_REPOSITORY" --source-repository "\$TASK_SOURCE_REPOSITORY" --branch "\$PR_HEAD_REF"/);
  assert.doesNotMatch(validate, /continue-on-error|\|\|\s*true|HEAD\^|git fetch|--backlog-repository/);
  assert.equal((workflow.match(/name: Verify committed pull-request lifecycle evidence/g) || []).length, 1);
});

for (const exitCode of [0, 1]) {
  test(`actual validation shell command preserves quoted context and exit status ${exitCode}`, (t) => {
    const cwd = mkdtempSync(join(tmpdir(), 'memory-workflow-')); t.after(() => rmSync(cwd, { recursive: true, force: true }));
    const bin = join(cwd, 'bin'); mkdirSync(bin);
    // A synthetic node captures argv without running any repository code.
    const capture = join(cwd, 'argv.json'); const injected = 'branch $(touch SHOULD_NOT_EXIST); trailing text';
    const executable = join(bin, 'node');
    writeFileSync(executable, `#!${process.execPath}\nrequire('node:fs').writeFileSync(process.env.CAPTURE, JSON.stringify(process.argv.slice(2))); process.exit(Number(process.env.EXIT_CODE));\n`);
    chmodSync(executable, 0o755);
    const command = step('Verify committed pull-request lifecycle evidence').match(/^        run: (.+)$/m)?.[1];
    assert.ok(command);
    const result = spawnSync('bash', ['-e', '-c', command], { cwd, encoding: 'utf8', env: {
      ...process.env, PATH: `${bin}:${process.env.PATH}`, CAPTURE: capture, EXIT_CODE: String(exitCode),
      BASE_SHA: 'a'.repeat(40), CODE_REPOSITORY: 'example/application', TASK_SOURCE_REPOSITORY: 'example/tasks', PR_HEAD_REF: injected,
    } });
    assert.equal(result.status, exitCode, result.stderr);
    assert.deepEqual(JSON.parse(readFileSync(capture, 'utf8')), [
      'skills/project-memory/scripts/project-memory.mjs', 'validate', '--base', 'a'.repeat(40),
      '--repository', 'example/application', '--source-repository', 'example/tasks', '--branch', injected,
    ]);
    assert.throws(() => readFileSync(join(cwd, 'SHOULD_NOT_EXIST')), /ENOENT/);
  });
}
test('helper uses offline Git protections without a runtime provenance default', () => {
  const adapter = readFileSync(join(root, 'skills/project-memory/scripts/lib/git.mjs'), 'utf8');
  const cli = readFileSync(join(root, 'skills/project-memory/scripts/project-memory.mjs'), 'utf8');
  assert.match(adapter, /GIT_ALLOW_PROTOCOL: ''/); assert.match(adapter, /GIT_NO_LAZY_FETCH: '1'/);
  assert.match(adapter, /--no-replace-objects/); assert.match(adapter, /GIT_NO_REPLACE_OBJECTS: '1'/);
  assert.doesNotMatch(cli, /UPSTREAM\.json|provenance\.local|fetch\(|https\.request/);
});
test('skill documents read-only legacy continuity and explicit validation context', () => {
  const skill = readFileSync(join(root, 'skills/project-memory/SKILL.md'), 'utf8');
  assert.match(skill, /Node 24 and Git 2\.43 or later/);
  assert.match(skill, /byte-preserved read-only history/);
  assert.match(skill, /--source-repository example\/tasks/);
  assert.match(skill, /do not backdate events/);
  assert.doesNotMatch(skill, /version 1; historical|New events use schema version 1/);
});
