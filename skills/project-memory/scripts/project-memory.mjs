#!/usr/bin/env node
import { appendFileSync, existsSync, mkdirSync, readFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { resolve, relative } from 'node:path';

const root = execFileSync('git', ['rev-parse', '--show-toplevel'], { encoding: 'utf8' }).trim();
const ledger = resolve(root, '.project-memory/tasks.jsonl');
const [command = 'status', ...argv] = process.argv.slice(2);
const args = new Map();
for (let i = 0; i < argv.length; i += 1) {
  if (argv[i].startsWith('--')) {
    const key = argv[i].slice(2);
    const values = [];
    while (argv[i + 1] && !argv[i + 1].startsWith('--')) values.push(argv[++i]);
    args.set(key, values);
  }
}
const value = (key, required = false) => {
  const text = (args.get(key) || []).join(' ').trim();
  if (required && !text) throw new Error(`--${key} is required`);
  return text;
};
const files = () => (args.get('files') || []).map((file) => relative(root, resolve(root, file))).filter(Boolean);
const validPath = (file) => file && !file.startsWith('../') && !file.startsWith('.project-memory/');
const readEvents = () => !existsSync(ledger) ? [] : readFileSync(ledger, 'utf8').trim().split('\n').filter(Boolean).map((line, index) => {
  try { return JSON.parse(line); } catch { throw new Error(`invalid JSONL at ${ledger}:${index + 1}`); }
});
const append = (event) => {
  mkdirSync(resolve(root, '.project-memory'), { recursive: true });
  appendFileSync(ledger, `${JSON.stringify(event)}\n`);
};
const changedPaths = (base) => execFileSync('git', ['diff', '--name-only', `${base}...HEAD`], { cwd: root, encoding: 'utf8' })
  .split('\n').filter(Boolean).filter((path) => !path.startsWith('.project-memory/'));
const assertEvent = (event, index) => {
  if (!event || typeof event !== 'object') throw new Error(`event ${index + 1} must be an object`);
  if (!event.id || !event.summary || !event.next) throw new Error(`event ${index + 1} needs id, summary, and next`);
  if (!['start', 'complete'].includes(event.event)) throw new Error(`event ${index + 1} has invalid event type`);
  if (Number.isNaN(Date.parse(event.at))) throw new Error(`event ${index + 1} has invalid timestamp`);
  if (!Array.isArray(event.files) || event.files.length === 0 || event.files.some((file) => !validPath(file))) {
    throw new Error(`event ${index + 1} needs repository-relative non-memory files`);
  }
};

try {
  if (command === 'start' || command === 'complete') {
    const id = value('id', true);
    const event = { id, event: command, at: new Date().toISOString(), summary: value('summary', true), next: value('next', true), files: files() };
    assertEvent(event, 0);
    append(event);
    console.log(`${command}: ${id}`);
  } else if (command === 'status') {
    const events = readEvents();
    const latest = new Map();
    for (const event of events) latest.set(event.id, event);
    console.log(JSON.stringify([...latest.values()].sort((a, b) => a.at.localeCompare(b.at)), null, 2));
  } else if (command === 'validate') {
    const base = value('base', true);
    const events = readEvents();
    const grouped = new Map();
    for (const [index, event] of events.entries()) {
      assertEvent(event, index);
      grouped.set(event.id, [...(grouped.get(event.id) || []), event]);
    }
    const covered = new Set();
    for (const [id, taskEvents] of grouped) {
      const startIndex = taskEvents.findIndex((event) => event.event === 'start');
      const completed = taskEvents.slice(startIndex + 1).filter((event) => event.event === 'complete');
      if (startIndex === -1 || completed.length === 0) continue;
      for (const event of completed) for (const file of event.files || []) covered.add(file);
    }
    const missing = changedPaths(base).filter((path) => !covered.has(path));
    if (missing.length) throw new Error(`changed paths lack started-and-completed task evidence:\n${missing.join('\n')}`);
    console.log(`project-memory: ${changedPaths(base).length} changed paths covered by ${[...grouped.keys()].length} task(s)`);
  } else {
    throw new Error(`unknown command: ${command}`);
  }
} catch (error) {
  console.error(`project-memory: ${error.message}`);
  process.exitCode = 1;
}
