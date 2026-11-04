import test from 'node:test';
import assert from 'node:assert/strict';
import { EventEmitter } from 'node:events';
import { PassThrough } from 'node:stream';
import { chmod, mkdtemp, realpath, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createCodexCliProvider, createResponsesProvider } from '../src/providers.mjs';

// Independent, hand-declared acceptance contract frozen before production.
// These invented transports do not run a CLI, call a provider, or use credentials.
const request = { task: 'invented reply boundary' };
const start = { type: 'turn.started' }, end = { type: 'turn.completed' };
const message = (text) => ({ type: 'item.completed', item: { type: 'agent_message', text } });
const diagnostic = 'Code Mode is unavailable because code-mode host is disabled. Code mode will fail closed; enable `features.code_mode_host` and install `codex-code-mode-host`.';
const jsonl = (events) => Buffer.from(events.map((x) => JSON.stringify(x)).join('\n') + '\n');
function spawnBytes(chunks) {
  return () => {
    const child = new EventEmitter();
    for (const name of ['stdin', 'stdout', 'stderr']) child[name] = new PassThrough();
    child.kill = () => true;
    queueMicrotask(() => {
      for (const chunk of chunks) child.stdout.write(chunk);
      child.stdout.end(); child.stderr.end(); child.emit('close', 0, null);
    });
    return child;
  };
}
async function cli(t, chunks) {
  const directory = await realpath(await mkdtemp(join(tmpdir(), 'foundry-reply-oracle-')));
  await chmod(directory, 0o700);
  t.after(() => rm(directory, { recursive: true, force: true }));
  return createCodexCliProvider({ cwd: directory, spawnImpl: spawnBytes(chunks) })(request);
}
const apiBytes = (text) => Buffer.from(JSON.stringify({ status: 'completed', output: [{ type: 'message', role: 'assistant', content: [{ type: 'output_text', text }] }] }));
function api(chunks) {
  return createResponsesProvider({ model: 'invented-model', apiKey: 'invented-key', fetchImpl: async () => new Response(new ReadableStream({
    start(controller) { for (const chunk of chunks) controller.enqueue(chunk); controller.close(); },
  })) })(request);
}
async function rejects(promise, code) {
  await assert.rejects(promise, (error) => {
    assert.equal(error.code, code);
    assert.doesNotMatch(error.message, /private-payload|invented-key/);
    return true;
  });
}

test('CLI accepts one completed turn without requiring item.started or IDs', async (t) => {
  assert.equal(await cli(t, [jsonl([start, message('first'), message('final'), end])]), 'final');
});
test('CLI preserves optional thread, reasoning, metadata and the exact active diagnostic', async (t) => {
  const metadata = { type: 'rate_limits.updated' };
  const events = [metadata, { type: 'thread.started', thread_id: 'invented' }, start,
    { type: 'item.started', item: { type: 'reasoning' } },
    { type: 'item.completed', item: { type: 'reasoning', text: 'ignored' } },
    { type: 'item.completed', item: { type: 'error', message: diagnostic } }, message('valid'), end, metadata];
  assert.equal(await cli(t, [jsonl(events)]), 'valid');
});
test('CLI rejects invalid event phase order rather than borrowing earlier completion', async (t) => {
  const cases = {
    'no start': [message('private-payload'), end],
    'completion before start': [end, start, message('private-payload')],
    'second incomplete turn': [start, message('first'), end, start, message('private-payload')],
    'item after terminal': [start, message('first'), end, message('private-payload')],
    'only late message': [start, end, message('private-payload')],
    'repeated terminal': [start, message('first'), end, end],
    'nested start': [start, start, message('private-payload'), end],
    'thread after start': [start, { type: 'thread.started' }, message('private-payload'), end],
    'repeated thread': [{ type: 'thread.started' }, { type: 'thread.started' }, start, message('private-payload'), end],
    'diagnostic after terminal': [start, message('first'), end, { type: 'item.completed', item: { type: 'error', message: diagnostic } }],
  };
  for (const [name, events] of Object.entries(cases)) await t.test(name, async (s) => rejects(cli(s, [jsonl(events)]), 'PROVIDER_CLI_INVALID_OUTPUT'));
});
test('CLI EOF without a completed assistant turn remains incomplete', async (t) => {
  for (const events of [[], [start], [start, message('partial')], [start, end]]) {
    await rejects(cli(t, [jsonl(events)]), 'PROVIDER_INCOMPLETE');
  }
});
test('CLI still refuses tools and explicit failed turns', async (t) => {
  await rejects(cli(t, [jsonl([start, { type: 'item.completed', item: { type: 'command_execution', command: 'private-payload' } }, end])]), 'PROVIDER_TOOL_ACTIVITY');
  await rejects(cli(t, [jsonl([start, { type: 'turn.failed', error: 'private-payload' }])]), 'PROVIDER_CLI_FAILED');
});
for (const kind of ['CLI', 'Responses']) {
  const invoke = (t, chunks) => kind === 'CLI' ? cli(t, chunks) : api(chunks);
  const bytes = (text) => kind === 'CLI' ? jsonl([start, message(text), end]) : apiBytes(text);
  const invalid = kind === 'CLI' ? 'PROVIDER_CLI_INVALID_OUTPUT' : 'PROVIDER_INVALID_RESPONSE';
  test(`${kind} preserves valid split UTF8 and literal replacement characters`, async (t) => {
    const source = bytes('valid é 😀 �');
    // Split inside a multibyte codepoint; decoding must apply to the full bytes.
    const at = source.indexOf(Buffer.from('é')) + 1;
    assert.equal(await invoke(t, [source.subarray(0, at), source.subarray(at)]), 'valid é 😀 �');
  });
  test(`${kind} rejects malformed raw UTF8 without replacement`, async (t) => {
    const valid = bytes('MARKER');
    const at = valid.indexOf('MARKER');
    const bad = Buffer.concat([valid.subarray(0, at), Buffer.from([0xc3, 0x28]), valid.subarray(at + 6)]);
    await rejects(invoke(t, [bad.subarray(0, at + 1), bad.subarray(at + 1)]), invalid);
    await rejects(invoke(t, [Buffer.from([0xef, 0xbb, 0xbf]), bytes('valid')]), invalid);
  });
  test(`${kind} rejects escaped unpaired assistant surrogates`, async (t) => {
    for (const text of ['\ud800', '\udc00', 'private-payload\ud800']) await rejects(invoke(t, [bytes(text)]), invalid);
  });
}
test('Responses text-only fallback remains supported with decoded text validation', async () => {
  const fallback = (text) => createResponsesProvider({ model: 'invented-model', apiKey: 'invented-key', fetchImpl: async () => ({ ok: true, text: async () => text }) })(request);
  assert.equal(await fallback(apiBytes('fallback é 😀 �').toString('utf8')), 'fallback é 😀 �');
  await rejects(fallback(apiBytes('\ud800').toString('utf8')), 'PROVIDER_INVALID_RESPONSE');
});
