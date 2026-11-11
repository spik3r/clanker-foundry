import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, readFileSync, writeFileSync, rmSync, chmodSync, readdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import { runInNewContext } from 'node:vm';

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

const workflowNames = [
  'json-validator-go', 'model-authoring-contract', 'model-lifecycle-go',
  'orchestration', 'project-memory-conformance', 'tui-model-go', 'validate',
];
const groupExpression = "${{ github.workflow }}-${{ github.event_name == 'pull_request' && format('pr-{0}', github.event.pull_request.number) || format('run-{0}-{1}', github.run_id, github.run_attempt) }}";
const cancelExpression = "${{ github.event_name == 'pull_request' }}";

// Assert the maintained header's exact bounded syntax, not a general YAML parser.
// YAML syntax and unchanged job bodies are also checked during workflow review.
const topLevelBlock = (source, key) => {
  const lines = source.split('\n');
  const starts = lines.flatMap((line, i) => line.startsWith(`${key}:`) ? [i] : []);
  assert.equal(starts.length, 1, `expected one ${key} block`);
  const end = lines.findIndex((line, i) => i > starts[0] && /^\S/.test(line));
  return lines.slice(starts[0], end === -1 ? undefined : end).join('\n').trimEnd();
};
const triggers = (source, manual) => {
  const block = topLevelBlock(source, 'on');
  assert.equal(block, `on:
  push:
    branches: [main]
    tags: ['**']
  pull_request:${manual ? '\n  workflow_dispatch:' : ''}`);
  return block;
};
const concurrency = (source) => {
  const block = topLevelBlock(source, 'concurrency');
  assert.equal(block, `concurrency:\n  group: ${groupExpression}\n  cancel-in-progress: ${cancelExpression}`);
  return {
    group: block.match(/^  group: (.+)$/m)[1],
    cancel: block.match(/^  cancel-in-progress: (.+)$/m)[1],
  };
};
// These fixed expressions use the common boolean/format subset of Actions and
// JavaScript. This checks configured values, not GitHub's scheduler execution.
const evaluate = (expression, github) => expression.replace(/\$\{\{ (.*?) \}\}/g, (_, body) =>
  String(runInNewContext(body, {
    github,
    format: (pattern, ...args) => pattern.replace(/\{(\d+)\}/g, (_, i) => String(args[Number(i)])),
  }, { timeout: 1000 })));
const context = (workflow, event_name, overrides = {}) => ({
  workflow, event_name, run_id: 100, run_attempt: 1, ref: 'refs/heads/main',
  event: { pull_request: { number: 17, head: { repo: { fork: false } } } },
  ...overrides,
});
const scheduled = (block, event) => {
  const events = [...block.matchAll(/^  (\w+):$/gm)].map((match) => match[1]);
  if (!events.includes(event.event_name)) return false;
  if (event.event_name !== 'push') return true;
  const branches = block.match(/^    branches: \[([^\]]+)\]$/m)[1].split(',').map((s) => s.trim());
  const tags = block.match(/^    tags: \['([^']+)'\]$/m)[1];
  return event.ref.startsWith('refs/heads/') ? branches.includes(event.ref.slice(11)) :
    event.ref.startsWith('refs/tags/') && tags === '**';
};

test('all maintained workflows have scheduling coverage', () => {
  assert.deepEqual(readdirSync(join(root, '.github/workflows')).filter((name) => /\.ya?ml$/.test(name)).sort(),
    workflowNames.map((name) => `${name}.yml`).sort());
});
for (const name of workflowNames) {
  const source = readFileSync(join(root, '.github/workflows', `${name}.yml`), 'utf8');
  const manual = name !== 'project-memory-conformance';
  test(`${name}: scoped push, unfiltered PR and unchanged manual entrypoint`, () => {
    assert.equal(topLevelBlock(source, 'name'), `name: ${name}`);
    triggers(source, manual);
    assert.doesNotMatch(source, /pull_request_target:|workflow_run:/);
  });
  test(`${name}: trigger scenarios retain main, all tags, forks and manual inventory`, () => {
    const block = triggers(source, manual);
    const cases = [
      ['main push', context(name, 'push'), true],
      ['feature push', context(name, 'push', { ref: 'refs/heads/feature/change' }), false],
      ['release branch push', context(name, 'push', { ref: 'refs/heads/release/v1' }), false],
      ['tag push', context(name, 'push', { ref: 'refs/tags/v1.2.3' }), true],
      ['slash tag push', context(name, 'push', { ref: 'refs/tags/release/v1.2.3' }), true],
      ['PR', context(name, 'pull_request', { ref: 'refs/pull/17/merge' }), true],
      ['fork PR', context(name, 'pull_request', { event: { pull_request: { number: 18, head: { repo: { fork: true } } } } }), true],
      ['manual main', context(name, 'workflow_dispatch'), manual],
      ['manual feature', context(name, 'workflow_dispatch', { ref: 'refs/heads/feature/change' }), manual],
      ['manual tag', context(name, 'workflow_dispatch', { ref: 'refs/tags/release/v1' }), manual],
      ['privileged PR event', context(name, 'pull_request_target'), false],
    ];
    for (const [label, event, expected] of cases) assert.equal(scheduled(block, event), expected, label);
  });
  test(`${name}: only same-workflow, same-PR runs share a cancelling group`, () => {
    const { group, cancel } = concurrency(source);
    const pr = context(name, 'pull_request', { sha: 'newer-commit' });
    const firstGroup = evaluate(group, pr);
    assert.equal(firstGroup, `${name}-pr-17`);
    assert.equal(evaluate(cancel, pr), 'true');
    for (const later of [
      { run_id: 101 }, { run_attempt: 2 },
      // Rerunning an older commit joins this same group too. This policy does
      // not promise newest-commit ordering or distinguish reruns by SHA.
      { run_id: 99, run_attempt: 2, sha: 'older-commit' },
    ]) assert.equal(evaluate(group, { ...pr, ...later }), firstGroup);
    assert.notEqual(evaluate(group, context(name, 'pull_request', { event: { pull_request: { number: 18 } } })), firstGroup);
    for (const other of workflowNames.filter((candidate) => candidate !== name)) {
      assert.notEqual(evaluate(group, { ...pr, workflow: other }), firstGroup);
    }
  });
  test(`${name}: main, tag and manual runs neither cancel nor share pending groups`, () => {
    const { group, cancel } = concurrency(source);
    const groups = new Set();
    const events = [
      context(name, 'push'),
      context(name, 'push', { ref: 'refs/tags/release/v1' }),
      context(name, 'workflow_dispatch'),
      context(name, 'workflow_dispatch', { ref: 'refs/heads/feature/change' }),
    ];
    for (const [index, event] of events.entries()) {
      for (const attempt of [1, 2]) {
        const run = { ...event, run_id: 100 + index, run_attempt: attempt };
        assert.equal(evaluate(cancel, run), 'false');
        const actual = evaluate(group, run);
        assert.equal(actual, `${name}-run-${100 + index}-${attempt}`);
        assert.ok(!groups.has(actual), `pending-run collision: ${actual}`);
        groups.add(actual);
        assert.notEqual(actual, evaluate(group, context(name, 'pull_request')));
      }
    }
  });
}
test('pack validator and installer checks remain and include pack and project setup regressions', () => {
  assert.equal(job('validate'), `  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install GNU Stow
        run: sudo apt-get update && sudo apt-get install -y stow
      - name: Run pack validator
        run: bash scripts/validate.sh
      - name: Run pack validator regression tests
        run: bash tests/test-validate.sh
      - name: Run installer regression tests
        run: bash tests/test-install.sh
      - name: Run project shared-base regression tests
        run: bash tests/test-project-setup.sh`);
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
