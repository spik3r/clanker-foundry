#!/usr/bin/env node
import { randomUUID } from 'node:crypto';
import { relative, resolve, sep } from 'node:path';
import {
  LEGACY, TASK, REPO, fail, claimSections, makeSource, sourcePath,
  eventPath, filesList, buildGraph,
} from './lib/state.mjs';
import { createGit, localMemory, loadMemory, writeImmutable, validateCommitted } from './lib/git.mjs';

const [command = 'status', ...argv] = process.argv.slice(2);
const allowed = {
  register: ['id', 'source-repo', 'backlog', 'ref', 'record', 'source-repository', 'backlog-repository', 'claim'],
  status: ['id', 'stale-days'],
  start: ['id', 'source', 'claim', 'repository', 'branch', 'pr', 'files', 'summary', 'next'],
  reconcile: ['id', 'source', 'claim', 'repository', 'branch', 'pr', 'files', 'summary', 'next'],
  checkpoint: ['id', 'pr', 'files', 'summary', 'next'],
  resume: ['id', 'pr', 'files', 'summary', 'next'],
  complete: ['id', 'pr', 'files', 'summary', 'next', 'validation'],
  abandon: ['id', 'summary', 'next'],
  validate: ['base', 'repository', 'branch', 'source-repository', 'backlog-repository'],
};
function parseArgs() {
  const result = new Map();
  for (let i = 0; i < argv.length; i += 1) {
    if (!argv[i].startsWith('--') || argv[i] === '--') fail(`unexpected argument: ${argv[i]}`);
    const key = argv[i].slice(2);
    if (!allowed[command].includes(key)) fail(`unknown option for ${command}: --${key}`);
    if (result.has(key)) fail(`duplicate option: --${key}`);
    const values = [];
    while (argv[i + 1] !== undefined && !argv[i + 1].startsWith('--')) values.push(argv[++i]);
    if (!values.length) fail(`--${key} needs a value`);
    result.set(key, values);
  }
  return result;
}
function githubRepository(remote) {
  const match = /^(?:https:\/\/github\.com\/|git@github\.com:)([A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+?)(?:\.git)?\/?$/.exec(remote);
  return match?.[1];
}
try {
  if (!Object.hasOwn(allowed, command)) fail(`unknown command: ${command}`);
  const args = parseArgs();
  const value = (key, required = false) => {
    const result = (args.get(key) || []).join(' ').trim();
    if (required && !result) fail(`--${key} is required`);
    if (args.has(key) && !result) fail(`--${key} must not be empty`);
    return result;
  };
  const alias = (preferred, old, required = false) => {
    if (args.has(preferred) && args.has(old)) fail(`use only --${preferred}, not both aliases`);
    const result = value(preferred) || value(old);
    if (required && !result) fail(`--${preferred} is required`);
    return result;
  };
  const repo = createGit();
  const observedIdentity = (git) => {
    try { return githubRepository(git.git('remote', 'get-url', 'origin').trim()); }
    catch { return undefined; }
  };
  const codeRepository = () => {
    const observed = observedIdentity(repo); const identity = value('repository') || observed;
    if (!identity || !REPO.test(identity)) fail('provide --repository when the code checkout has no recognized GitHub remote');
    if (observed && identity !== observed) fail('code repository disagrees with the local origin remote');
    return identity;
  };
  if (command === 'register') {
    const id = value('id', true); const recordPath = value('record', true);
    if (!TASK.test(id)) fail('task ID must use a canonical project prefix and safe alphanumeric slug (3-80 characters)');
    const backlog = createGit(resolve(repo.root, alias('source-repo', 'backlog', true)));
    const revision = backlog.commit(value('ref', true));
    let observedRepository;
    try { observedRepository = githubRepository(backlog.git('remote', 'get-url', 'origin').trim()); } catch { /* An explicit identity is required below. */ }
    const repository = alias('source-repository', 'backlog-repository') || observedRepository;
    if (!repository || !REPO.test(repository)) fail('provide --source-repository when the local source has no recognized GitHub remote');
    if (observedRepository && observedRepository !== repository) fail('backlog repository disagrees with the local source remote');
    const records = new Map();
    for (const [path, item] of backlog.tree(revision, 'tasks')) {
      if (!path.endsWith('.md')) continue;
      if (item.type !== 'blob' || !['100644', '100755'].includes(item.mode)) fail(`canonical task must be a regular Markdown file: ${path}`);
      records.set(path, { text: backlog.readBlob(item.blob), blob: item.blob, mode: item.mode });
    }
    const source = makeSource({ repository, revision, id, path: recordPath, records });
    const claims = claimSections(source.record);
    if (!claims.size) fail('record the bounded claim in the canonical task before importing a task-source snapshot');
    if (value('claim') && !claims.has(value('claim'))) fail('requested claim section does not exist in the pinned task record');
    const path = sourcePath(source);
    const created = writeImmutable(repo.root, path, source, { identicalOK: true });
    console.log(JSON.stringify({ id, source: path, revision, claims: [...claims.keys()], created,
      warning: 'Pinned reviewed source, not live ownership, authorization or authentication. Independently verify against the canonical repository.' }, null, 2));
  } else if (command === 'validate') {
    const sourceRepository = alias('source-repository', 'backlog-repository', true);
    if (!REPO.test(sourceRepository)) fail('invalid --source-repository identity');
    const result = validateCommitted(repo, { base: value('base', true), repository: codeRepository(),
      backlogRepository: sourceRepository, branch: value('branch') || undefined });
    console.log(`project-memory: ${result.paths} changed paths covered; ${result.newEvents} new events; committed HEAD ${result.head.slice(0, 12)}${result.integrationBranch ? `; integration branch ${JSON.stringify(result.integrationBranch)}` : ''}`);
  } else {
    const entries = localMemory(repo.root); const memory = loadMemory(entries);
    const id = value('id', command !== 'status');
    if (id && !TASK.test(id)) fail('task ID must use a canonical project prefix and safe alphanumeric slug (3-80 characters)');
    if (command === 'status') {
      const days = args.has('stale-days') ? Number(value('stale-days')) : 7;
      if (!Number.isFinite(days) || days < 0) fail('--stale-days must be nonnegative');
      let legacy = null;
      if (entries.has(LEGACY)) {
        const rows = entries.get(LEGACY).text.split('\n').filter((line) => line.trim());
        legacy = { path: LEGACY, records: rows.length, readOnly: true, providesV2Coverage: false };
      }
      const tasks = [...memory.graph.tasks].filter(([taskId]) => !id || taskId === id).map(([taskId, state]) => ({
        id: taskId, conflict: state.heads.length !== 1,
        heads: state.heads.map((event) => ({ eventId: event.eventId, runId: event.runId, kind: event.kind,
          files: event.files, next: event.next, at: event.at, branch: event.branch, pullRequest: event.pullRequest,
          stale: ['start', 'checkpoint'].includes(event.kind) && Date.now() - Date.parse(event.at) > days * 86400000,
          sourceRevision: event.task.revision, claim: event.task.claim.heading })),
      }));
      console.log(JSON.stringify({ tasks, legacy, sourceMeaning: 'Reviewed pinned source only; confirm current canonical ownership separately.' }, null, 2));
    } else {
      const heads = memory.graph.tasks.get(id)?.heads || [];
      const beginning = command === 'start' || command === 'reconcile';
      let event;
      const providedFiles = () => args.get('files')?.map((path) => {
        const normalized = relative(repo.root, resolve(repo.root, path));
        return sep === '\\' ? normalized.split('\\').join('/') : normalized;
      });
      if (beginning) {
        if (command === 'reconcile' && heads.length < 2) fail('reconcile requires multiple unresolved heads; use start or resume');
        if (command === 'start' && heads.length > 1) fail('unresolved task frontier; use reconcile with the current intended scope');
        if (command === 'start' && heads.length === 1 && ['start', 'checkpoint'].includes(heads[0].kind)) fail('task is active; use resume/checkpoint or explicitly abandon it first');
        let candidates = [...memory.sources.values()].filter((source) => source.id === id);
        if (value('source')) candidates = candidates.filter((source) => source.revision === value('source'));
        if (value('claim')) candidates = candidates.filter((source) => claimSections(source.record).has(value('claim')));
        if (candidates.length !== 1) fail('select one registered source with --source <pinned revision>; no latest-source authority is guessed');
        const source = candidates[0]; const claims = claimSections(source.record);
        const claim = value('claim') ? claims.get(value('claim')) : claims.size === 1 ? [...claims.values()][0] : null;
        if (!claim) fail('select an existing canonical claim section with --claim <heading>');
        const files = providedFiles(); filesList(files);
        event = { formatVersion: 2, eventId: randomUUID(), runId: randomUUID(), kind: 'start',
          parents: heads.map((head) => head.eventId).sort(), sequence: 0,
          task: { repository: source.repository, id, path: source.path, revision: source.revision, blob: source.blob, claim },
          repository: codeRepository(), branch: value('branch') || repo.branch(),
          pullRequest: value('pr') || null, at: new Date().toISOString(),
          summary: value('summary', true), next: value('next', true), files };
      } else {
        if (heads.length !== 1) fail('select a task with one active head; unresolved frontiers need explicit reconcile');
        const parent = heads[0];
        if (!['start', 'checkpoint'].includes(parent.kind)) fail('task is complete or abandoned; start a new run');
        const files = providedFiles() || parent.files; filesList(files);
        event = { formatVersion: 2, eventId: randomUUID(), runId: parent.runId,
          kind: command === 'resume' ? 'checkpoint' : command, parents: [parent.eventId], sequence: parent.sequence + 1,
          task: parent.task, repository: parent.repository, branch: parent.branch,
          pullRequest: value('pr') || parent.pullRequest, at: new Date().toISOString(),
          summary: value('summary', true), next: value('next', true), files };
        if (command === 'complete') {
          event.validation = value('validation', true); event.snapshot = repo.snapshotIndex(files);
        }
      }
      // This validates structure while intentionally retaining historical forks.
      buildGraph([...memory.events, event], memory.sources);
      const path = eventPath(event); writeImmutable(repo.root, path, event);
      console.log(JSON.stringify({ id, command, event: path, runId: event.runId, eventId: event.eventId }, null, 2));
    }
  }
} catch (error) {
  console.error(`project-memory: ${error.message}`);
  process.exitCode = 1;
}
