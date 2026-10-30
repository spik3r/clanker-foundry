import test from 'node:test';
import assert from 'node:assert/strict';
import { EventEmitter } from 'node:events';
import { PassThrough } from 'node:stream';
import { chmod, mkdtemp, mkdir, realpath, rm, symlink, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import {
  createCodexCliProvider, createResponsesProvider,
} from '../src/providers.mjs';

const request = Object.freeze({ task: { task_id: 'CF-provider-smoke' }, files: [] });
const response = (output, status = 'completed') => new Response(JSON.stringify({ status, output }), {
  status: 200,
  headers: { 'content-type': 'application/json' },
});
const responseMessage = (text) => ({
  type: 'message', role: 'assistant', status: 'completed',
  content: [{ type: 'output_text', text }],
});

test('Responses provider makes a fixed text-only, non-stored request', async () => {
  let called;
  const invoke = createResponsesProvider({
    model: 'gpt-test', apiKey: 'test-key', fetchImpl: async (...args) => {
      called = args;
      return response([responseMessage('provider-smoke-ok')]);
    },
  });
  assert.equal(await invoke(request), 'provider-smoke-ok');
  assert.equal(called[0], 'https://api.openai.com/v1/responses');
  const body = JSON.parse(called[1].body);
  assert.equal(body.model, 'gpt-test');
  assert.deepEqual(body.tools, []);
  assert.equal(body.store, false);
  assert.equal(Object.hasOwn(body, 'previous_response_id'), false);
  assert.equal(called[1].headers.authorization, 'Bearer test-key');
  assert.equal(called[1].signal.aborted, false);
  assert.match(body.input[0].content[0].text, /CF-provider-smoke/);
});

test('Responses provider rejects missing credentials without exposing their value', () => {
  assert.throws(() => createResponsesProvider({ model: 'gpt-test', apiKey: '' }), { code: 'PROVIDER_AUTH_MISSING' });
});

test('Responses provider reports only status for HTTP errors', async () => {
  const invoke = createResponsesProvider({
    model: 'gpt-test', apiKey: 'do-not-print-this',
    fetchImpl: async () => new Response('secret provider detail', { status: 429 }),
  });
  await assert.rejects(invoke(request), (error) => {
    assert.equal(error.code, 'PROVIDER_HTTP_ERROR');
    assert.equal(error.message, 'Provider returned HTTP 429');
    assert.doesNotMatch(error.message, /secret|do-not-print/);
    return true;
  });
});

test('Responses provider rejects incomplete, tool, refusal, invalid and oversized output', async (t) => {
  const cases = [
    [{ status: 'incomplete', output: [responseMessage('partial')] }, 'PROVIDER_INCOMPLETE'],
    [{ status: 'completed', output: [{ type: 'function_call', name: 'x' }] }, 'PROVIDER_TOOL_ACTIVITY'],
    [{ status: 'completed', output: [{ type: 'message', role: 'assistant', content: [{ type: 'refusal', refusal: 'no' }] }] }, 'PROVIDER_REFUSAL'],
    [{ status: 'completed', output: [responseMessage('a'.repeat(20))] }, 'OUTPUT_TOO_LARGE'],
  ];
  for (const [payload, code] of cases) {
    await t.test(code, async () => {
      const invoke = createResponsesProvider({ model: 'gpt-test', apiKey: 'x', maxOutputBytes: 8,
        fetchImpl: async () => response(payload.output, payload.status) });
      await assert.rejects(invoke(request), { code });
    });
  }
  const missingStatus = createResponsesProvider({ model: 'gpt-test', apiKey: 'x', fetchImpl: async () => new Response(JSON.stringify({ output: [responseMessage('no status')] })) });
  await assert.rejects(missingStatus(request), { code: 'PROVIDER_INCOMPLETE' });
  const invalid = createResponsesProvider({ model: 'gpt-test', apiKey: 'x', fetchImpl: async () => new Response('{bad') });
  await assert.rejects(invalid(request), { code: 'PROVIDER_INVALID_RESPONSE' });
});

test('Responses provider distinguishes cancellation, timeout and transport failure', async (t) => {
  await t.test('cancellation', async () => {
    const controller = new AbortController();
    const invoke = createResponsesProvider({ model: 'gpt-test', apiKey: 'x', fetchImpl: (_url, options) => new Promise((_resolve, reject) => {
      options.signal.addEventListener('abort', () => reject(new Error('aborted')), { once: true });
      controller.abort();
    }) });
    await assert.rejects(invoke(request, { signal: controller.signal }), { code: 'ATTEMPT_ABORTED' });
  });
  await t.test('timeout', async () => {
    const invoke = createResponsesProvider({ model: 'gpt-test', apiKey: 'x', timeoutMs: 10,
      fetchImpl: (_url, options) => new Promise((_resolve, reject) => options.signal.addEventListener('abort', () => reject(new Error('aborted')), { once: true })) });
    await assert.rejects(invoke(request), { code: 'PROVIDER_TIMEOUT' });
  });
  await t.test('network failure is redacted', async () => {
    const invoke = createResponsesProvider({ model: 'gpt-test', apiKey: 'hidden', fetchImpl: async () => { throw new Error('hidden endpoint and token'); } });
    await assert.rejects(invoke(request), (error) => error.code === 'PROVIDER_NETWORK_ERROR' && !/hidden/.test(error.message));
  });
  await t.test('transport error codes and messages are redacted', async () => {
    const invoke = createResponsesProvider({ model: 'gpt-test', apiKey: 'hidden', fetchImpl: async () => {
      throw Object.assign(new Error('hidden endpoint and token'), { code: 'ECONNRESET' });
    } });
    await assert.rejects(invoke(request), (error) => error.code === 'PROVIDER_NETWORK_ERROR' && !/hidden|ECONNRESET/.test(error.message));
  });
});

async function privateDirectory(t) {
  const root = await mkdtemp(path.join(os.tmpdir(), 'foundry-provider-'));
  await chmod(root, 0o700);
  t.after(() => rm(root, { recursive: true, force: true }));
  return realpath(root);
}

function cliEventOutput(text, { extra = [], includeEnd = true } = {}) {
  const events = [
    { type: 'thread.started', thread_id: 'opaque' },
    { type: 'turn.started' },
    { type: 'item.started', item: { type: 'agent_message' } },
    ...extra,
    { type: 'item.completed', item: { type: 'agent_message', text } },
  ];
  if (includeEnd) events.push({ type: 'turn.completed' });
  return `${events.map((event) => JSON.stringify(event)).join('\n')}\n`;
}

function fakeSpawn({ output, code = 0, beforeClose } = {}) {
  let called;
  const spawnImpl = (executable, args, options) => {
    called = { executable, args, options };
    const child = new EventEmitter();
    child.stdout = new PassThrough();
    child.stderr = new PassThrough();
    child.stdin = new PassThrough();
    child.kill = (signal) => {
      beforeClose?.(signal);
      queueMicrotask(() => child.emit('close', signal === 'SIGKILL' ? null : code, signal === 'SIGKILL' ? signal : null));
      return true;
    };
    queueMicrotask(() => {
      if (output !== undefined) child.stdout.write(output);
      child.stdout.end();
      child.stderr.write('sensitive stderr must not escape');
      child.stderr.end();
      child.emit('close', code, null);
    });
    return child;
  };
  return { spawnImpl, get called() { return called; } };
}

test('Codex CLI provider returns only the completed assistant message from a private empty cwd', async (t) => {
  const cwd = await privateDirectory(t);
  const fake = fakeSpawn({ output: cliEventOutput('cli-smoke-ok') });
  const invoke = createCodexCliProvider({ cwd, model: 'gpt-test', executable: '/usr/bin/codex', spawnImpl: fake.spawnImpl });
  assert.equal(await invoke(request), 'cli-smoke-ok');
  assert.equal(fake.called.executable, '/usr/bin/codex');
  assert.equal(fake.called.options.cwd, cwd);
  assert.equal(fake.called.options.shell, false);
  assert.deepEqual(fake.called.options.stdio, ['pipe', 'pipe', 'pipe']);
  assert.ok(!Object.hasOwn(fake.called.options.env, 'OPENAI_API_KEY'));
  assert.ok(!Object.hasOwn(fake.called.options.env, 'ANTHROPIC_API_KEY'));
  assert.ok(fake.called.args.includes('--model') && fake.called.args.includes('gpt-test'));
  assert.ok(fake.called.args.includes('--no-daemon'));
  for (const flag of ['shell_tool', 'multi_agent', 'apps', 'browser_use', 'computer_use', 'code_mode_host', 'skill_search', 'hooks', 'plugins']) {
    assert.ok(fake.called.args.includes(flag), `missing disable for ${flag}`);
  }
  assert.ok(fake.called.args.includes('mcp_servers={}'));
  assert.ok(fake.called.args.includes('plugins={}'));
  assert.ok(fake.called.args.includes('web_search="disabled"'));
  assert.ok(fake.called.args.includes('--ephemeral'));
  assert.ok(fake.called.args.includes('--ignore-user-config'));
  assert.ok(fake.called.args.includes('--ignore-rules'));
  assert.ok(fake.called.args.includes('--skip-git-repo-check'));
  assert.ok(fake.called.args.includes('--sandbox') && fake.called.args.includes('read-only'));
  assert.ok(fake.called.args.includes('--ask-for-approval') && fake.called.args.includes('never'));
  assert.ok(!fake.called.args.some((arg) => arg.includes('danger-full-access') || arg.includes('bypass')));
});

test('Codex CLI provider rejects tool events, unknown events, incomplete turns and output overflow', async (t) => {
  const cwd = await privateDirectory(t);
  const cases = [
    [cliEventOutput('x', { extra: [{ type: 'item.completed', item: { type: 'command_execution', command: 'secret' } }] }), 'PROVIDER_TOOL_ACTIVITY'],
    [cliEventOutput('x', { extra: [{ type: 'unknown.event' }] }), 'PROVIDER_TOOL_ACTIVITY'],
    [cliEventOutput('x', { extra: [{ type: 'item.completed', item: { type: 'error', message: 'unexpected error' } }] }), 'PROVIDER_TOOL_ACTIVITY'],
    [cliEventOutput('x', { includeEnd: false }), 'PROVIDER_INCOMPLETE'],
    [cliEventOutput('x'.repeat(16)), 'OUTPUT_TOO_LARGE'],
  ];
  for (const [output, code] of cases) await t.test(code, async () => {
    const fake = fakeSpawn({ output });
    const invoke = createCodexCliProvider({ cwd, model: 'gpt-test', maxOutputBytes: 8, spawnImpl: fake.spawnImpl });
    await assert.rejects(invoke(request), { code });
  });
});

test('Codex CLI provider accepts only the exact fail-closed code-mode-disabled diagnostic', async (t) => {
  const cwd = await privateDirectory(t);
  const diagnostic = 'Code Mode is unavailable because code-mode host is disabled. Code mode will fail closed; enable `features.code_mode_host` and install `codex-code-mode-host`.';
  const fake = fakeSpawn({ output: cliEventOutput('cli-smoke-ok', {
    extra: [{ type: 'item.completed', item: { type: 'error', message: diagnostic } }],
  }) });
  const invoke = createCodexCliProvider({ cwd, spawnImpl: fake.spawnImpl });
  assert.equal(await invoke(request), 'cli-smoke-ok');
});

test('Codex CLI provider redacts process errors and enforces cwd safety', async (t) => {
  const cwd = await privateDirectory(t);
  const failed = fakeSpawn({ output: 'secret', code: 7 });
  const invoke = createCodexCliProvider({ cwd, spawnImpl: failed.spawnImpl });
  await assert.rejects(invoke(request), (error) => error.code === 'PROVIDER_CLI_EXIT' && !/secret|sensitive/.test(error.message));

  const nonempty = path.join(cwd, 'not-empty');
  await mkdir(nonempty, { mode: 0o700 });
  await writeFile(path.join(nonempty, 'input.txt'), 'bounded test data');
  const bad = createCodexCliProvider({ cwd: nonempty, spawnImpl: failed.spawnImpl });
  await assert.rejects(bad(request), { code: 'UNSAFE_CWD' });

  const link = path.join(cwd, 'link');
  await symlink(nonempty, link);
  const linked = createCodexCliProvider({ cwd: link, spawnImpl: failed.spawnImpl });
  await assert.rejects(linked(request), { code: 'UNSAFE_CWD' });
});

test('Codex CLI provider terminates on timeout or cancellation', async (t) => {
  const cwd = await privateDirectory(t);
  const waitingSpawn = (_executable, _args, _options) => {
    const child = new EventEmitter();
    child.stdout = new PassThrough(); child.stderr = new PassThrough(); child.stdin = new PassThrough();
    child.kill = (signal) => { queueMicrotask(() => child.emit('close', null, signal)); return true; };
    return child;
  };
  const timed = createCodexCliProvider({ cwd, timeoutMs: 10, spawnImpl: waitingSpawn });
  await assert.rejects(timed(request), { code: 'PROVIDER_TIMEOUT' });
  const controller = new AbortController();
  const cancelled = createCodexCliProvider({ cwd, timeoutMs: 1000, spawnImpl: waitingSpawn });
  const pending = cancelled(request, { signal: controller.signal });
  controller.abort();
  await assert.rejects(pending, { code: 'ATTEMPT_ABORTED' });
});
