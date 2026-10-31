import { createHash } from 'node:crypto';
import { chmod, mkdtemp, realpath, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { createCodexCliProvider, createResponsesProvider } from '../src/providers.mjs';

function parseArgs(argv) {
  const result = {};
  for (let index = 0; index < argv.length; index += 1) {
    const key = argv[index];
    if (!['--provider', '--model'].includes(key) || result[key.slice(2)] !== undefined) {
      throw new Error('Use --provider codex|responses and optional --model MODEL');
    }
    const value = argv[++index];
    if (!value || value.startsWith('--')) throw new Error('A value is required for each option');
    result[key.slice(2)] = value;
  }
  if (!['codex', 'responses'].includes(result.provider)) throw new Error('Provider must be codex or responses');
  if (result.provider === 'responses' && !result.model) throw new Error('Responses provider requires --model');
  return result;
}

let cwd;
try {
  const { provider, model } = parseArgs(process.argv.slice(2));
  let invoke;
  if (provider === 'responses') {
    invoke = createResponsesProvider({ model });
  } else {
    const created = await mkdtemp(path.join(os.tmpdir(), 'foundry-provider-live-'));
    await chmod(created, 0o700);
    cwd = await realpath(created);
    invoke = createCodexCliProvider({ model, cwd });
  }
  const output = await invoke({
    task: {
      task_id: 'CF-provider-live-smoke',
      prompt: 'Reply with exactly this text and nothing else: foundry-live-provider-ok',
    },
    attempt: { attempt_id: 'provider-live-smoke-1' },
    files: [],
    inputSubjects: [],
  });
  const expected = 'foundry-live-provider-ok';
  if (output.trim() !== expected) throw Object.assign(new Error('Provider response did not match the smoke check'), { code: 'LIVE_SMOKE_MISMATCH' });
  const digest = createHash('sha256').update(output, 'utf8').digest('hex');
  process.stdout.write(`${JSON.stringify({ provider, model: model ?? 'configured', status: 'PASS', output_bytes: Buffer.byteLength(output, 'utf8'), output_sha256: digest })}\n`);
} catch (error) {
  const code = typeof error?.code === 'string' ? error.code : 'LIVE_SMOKE_FAILED';
  process.stderr.write(`${JSON.stringify({ status: 'FAIL', code })}\n`);
  process.exitCode = 1;
} finally {
  if (cwd) await rm(cwd, { recursive: true, force: true });
}
