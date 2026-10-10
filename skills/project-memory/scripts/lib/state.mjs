import { createHash } from 'node:crypto';

export const ROOT = '.project-memory/tasks';
export const LEGACY = '.project-memory/tasks.jsonl';
export const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
export const SHA = /^(?:[a-f0-9]{40}|[a-f0-9]{64})$/;
export const TASK_PATTERN = '[A-Z][A-Z0-9]{0,15}-[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*';
export const TASK = new RegExp(`^(?=.{3,80}$)${TASK_PATTERN}$`);
export const REPO = /^[A-Za-z0-9][A-Za-z0-9_.-]*\/[A-Za-z0-9][A-Za-z0-9_.-]*$/;
export const MODES = new Set(['100644', '100755', '120000', '160000']);
export const fail = (message) => { throw new Error(message); };
export const nonempty = (value) => typeof value === 'string' && value.trim().length > 0 && value.isWellFormed();
export const digest = (text) => createHash('sha256').update(text).digest('hex');
export const blobHash = (text, format = 'sha1') => createHash(format).update(`blob ${Buffer.byteLength(text)}\0`).update(text).digest('hex');
export const object = (value) => value && typeof value === 'object' && !Array.isArray(value);
export function keys(value, allowed, label) {
  if (!object(value) || Object.keys(value).some((key) => !allowed.includes(key))) fail(`${label}: invalid object fields`);
}
export const canonicalPath = (path) => nonempty(path) && !path.includes('\0') && !path.includes('\\')
  && !path.startsWith('/') && !/^[A-Za-z]:/.test(path)
  && path.split('/').every((part) => part && part !== '.' && part !== '..');
export function filesList(files, label = 'event') {
  if (!Array.isArray(files) || files.length === 0 || new Set(files).size !== files.length
      || files.some((file) => !canonicalPath(file) || metadataPath(file) || file === LEGACY)) {
    fail(`${label}: files must be unique exact project paths, excluding recognized memory metadata`);
  }
}
export function metadataPath(path) {
  const parts = path.split('/');
  if (parts.length !== 5 || parts[0] !== '.project-memory' || parts[1] !== 'tasks' || !TASK.test(parts[2])) return false;
  if (parts[3] === 'sources') return /^(?:[a-f0-9]{40}|[a-f0-9]{64})\.json$/.test(parts[4]);
  return UUID.test(parts[3]) && UUID.test(parts[4].replace(/\.json$/, '')) && parts[4].endsWith('.json');
}
export function metadataContainer(path) {
  const parts = path.split('/');
  return parts[0] === '.project-memory' && parts[1] === 'tasks' && parts.length >= 2 && parts.length <= 4
    && (parts.length < 3 || TASK.test(parts[2]))
    && (parts.length < 4 || parts[3] === 'sources' || UUID.test(parts[3]));
}
export function parseJSON(text, label) {
  try { return JSON.parse(text); } catch { fail(`${label}: invalid JSON`); }
}
export function frontmatter(record) {
  if (!nonempty(record) || record.includes('\0')) fail('task record must be UTF-8 text');
  const match = /^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/.exec(record);
  if (!match) return null;
  const result = Object.create(null);
  for (const line of match[1].split(/\r?\n/)) {
    if (!line.trim() || line.trimStart().startsWith('#')) continue;
    const field = /^([a-z_]+):\s*(.*)$/.exec(line);
    if (!field) fail('unsupported task frontmatter; use the canonical task template');
    if (Object.hasOwn(result, field[1])) fail(`duplicate task frontmatter key: ${field[1]}`);
    let value = field[2].trim();
    if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) value = value.slice(1, -1);
    result[field[1]] = value;
  }
  return result;
}
export function repositories(record) {
  const raw = frontmatter(record)?.repositories;
  if (!raw || !raw.startsWith('[') || !raw.endsWith(']')) fail('task repositories must be an inline list');
  const values = raw.slice(1, -1).split(',').map((value) => value.trim().replace(/^(['"])(.*)\1$/, '$2')).filter(Boolean);
  if (!values.length || values.some((value) => !REPO.test(value))) fail('invalid task repository list');
  return values;
}
export function claimSections(record) {
  const headings = [];
  const headerLength = /^---\r?\n[\s\S]*?\r?\n---(?:\r?\n|$)/.exec(record)?.[0].length || 0;
  let offset = 0; let fence = null; let comment = false;
  for (const raw of record.match(/[^\n]*(?:\n|$)/g) || []) {
    if (!raw) continue;
    const index = offset; offset += raw.length;
    if (index < headerLength) continue;
    const line = raw.replace(/\r?\n$/, '');
    if (fence) {
      if (new RegExp(`^ {0,3}${fence.character}{${fence.length},}[ \\t]*$`).test(line)) fence = null;
      continue;
    }
    let visible = ''; let position = 0;
    while (position < line.length) {
      if (comment) {
        const end = line.indexOf('-->', position);
        if (end === -1) { visible += ' '.repeat(line.length - position); break; }
        visible += ' '.repeat(end + 3 - position); position = end + 3; comment = false;
      } else {
        const start = line.indexOf('<!--', position);
        if (start === -1) { visible += line.slice(position); break; }
        visible += line.slice(position, start); position = start; comment = true;
      }
    }
    const opening = /^ {0,3}(`{3,}|~{3,})(.*)$/.exec(visible);
    if (opening && !(opening[1][0] === '`' && opening[2].includes('`'))) {
      fence = { character: opening[1][0], length: opening[1].length }; continue;
    }
    const match = /^ {0,3}(#{1,6})[ \t]+(.+?)\s*$/.exec(visible);
    if (match && /^ {0,3}#/.test(line)) headings.push({ level: match[1].length,
      name: match[2].replace(/[ \t]+#+[ \t]*$/, '').trim(), index, bodyStart: offset });
  }
  const claims = new Map();
  for (const [index, heading] of headings.entries()) {
    if (heading.level < 2 || !/\b(?:claim|subclaim)\b/i.test(heading.name)) continue;
    const next = headings.slice(index + 1).find((item) => item.level <= heading.level);
    if (claims.has(heading.name)) fail(`duplicate claim heading: ${heading.name}`);
    const text = record.slice(heading.index, next?.index ?? record.length);
    if (!record.slice(heading.bodyStart, next?.index ?? record.length).trim()) fail(`empty claim section: ${heading.name}`);
    claims.set(heading.name, { heading: heading.name, digest: digest(text) });
  }
  return claims;
}
function declaredRedirectTarget(record, path, repository) {
  if (!/^# .*noncanonical redirect\s*$/mi.test(record)) fail(`${path}: missing canonical identity; not an explicit redirect`);
  const link = /canonical record is\s+\[[^\]]+\]\(([^)]+)\)/i.exec(record)?.[1];
  if (!link) fail(`${path}: redirect must name its exact canonical record`);
  let target;
  if (link.startsWith('https://')) {
    const escaped = repository.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    target = new RegExp(`^https://github\\.com/${escaped}/blob/[^/]+/(tasks/[^#?]+)$`).exec(link)?.[1];
  } else target = link.startsWith('tasks/') ? link : `tasks/${link}`;
  if (!canonicalPath(target) || !target.startsWith('tasks/')) fail(`${path}: redirect target is not an exact task path`);
  return target;
}
function redirectTarget(record, path, records, repository) {
  const target = declaredRedirectTarget(record, path, repository);
  if (!records.has(target)) fail(`${path}: redirect target is not an existing exact task record`);
  const meta = frontmatter(records.get(target).text);
  if (!TASK.test(meta?.id || '') || !target.startsWith(`tasks/${meta.id}-`)) fail(`${path}: redirect target is not canonical`);
  return { targetPath: target, targetId: meta.id, targetBlob: records.get(target).blob };
}
// Import-time evidence, not an offline proof that the source is still authoritative.
export function makeSource({ repository, revision, id, path, records }) {
  if (!REPO.test(repository) || !SHA.test(revision) || !TASK.test(id) || !path.startsWith(`tasks/${id}-`) || !path.endsWith('.md')) fail('invalid canonical task source identity');
  const canonicals = [];
  const redirects = [];
  for (const [name, entry] of records) {
    const meta = frontmatter(entry.text);
    if (meta && Object.hasOwn(meta, 'id') && !TASK.test(meta.id)) fail(`${name}: unsupported canonical task ID syntax`);
    if (meta?.id === id) {
      if (!name.startsWith(`tasks/${id}-`)) fail(`${name}: task ID disagrees with canonical path`);
      canonicals.push({ path: name, id, blob: entry.blob, status: meta.status || null });
    } else if (name.startsWith(`tasks/${id}-`)) {
      // A longer generic ID may share this prefix. Its own canonical identity
      // must agree with its filename before it can be treated as another task.
      if (TASK.test(meta?.id || '') && name.startsWith(`tasks/${meta.id}-`)) continue;
      if (meta && ['id', 'owner', 'status'].some((key) => Object.hasOwn(meta, key))) fail(`${name}: conflicting canonical task identity`);
      redirects.push({ path: name, blob: entry.blob, record: entry.text, ...redirectTarget(entry.text, name, records, repository) });
    }
  }
  if (canonicals.length !== 1 || canonicals[0].path !== path) fail(`canonical ID ${id} is ambiguous or missing: ${canonicals.map((entry) => entry.path).join(', ')}`);
  const entry = records.get(path);
  const source = { sourceVersion: 1, repository, revision, id, path, blob: entry.blob, record: entry.text,
    identityEvidence: { canonicals, redirects } };
  validateSource(source);
  return source;
}
export const sourcePath = (source) => `${ROOT}/${source.id}/sources/${source.revision}.json`;
export function validateSource(source, path = sourcePath(source)) {
  keys(source, ['sourceVersion', 'repository', 'revision', 'id', 'path', 'blob', 'record', 'identityEvidence'], 'source');
  if (source.sourceVersion !== 1 || ['repository', 'revision', 'id', 'path', 'blob'].some((key) => typeof source[key] !== 'string')
      || !REPO.test(source.repository) || !SHA.test(source.revision) || !TASK.test(source.id)
      || !canonicalPath(source.path) || !source.path.startsWith(`tasks/${source.id}-`) || !source.path.endsWith('.md')
      || !SHA.test(source.blob) || path !== sourcePath(source)) fail('invalid source path or identity');
  const meta = frontmatter(source.record);
  if (meta?.id !== source.id || blobHash(source.record, source.blob.length === 64 ? 'sha256' : 'sha1') !== source.blob) fail('source task ID or record blob does not match');
  repositories(source.record);
  keys(source.identityEvidence, ['canonicals', 'redirects'], 'source identity evidence');
  const { canonicals, redirects } = source.identityEvidence;
  if (!Array.isArray(canonicals) || canonicals.length !== 1 || !Array.isArray(redirects)) fail('source needs one canonical identity across every status');
  const canonical = canonicals[0];
  keys(canonical, ['path', 'id', 'blob', 'status'], 'canonical identity');
  if (canonical.path !== source.path || canonical.id !== source.id || canonical.blob !== source.blob || canonical.status !== (meta.status || null)) fail('canonical identity evidence disagrees with source');
  const seen = new Set([source.path]);
  for (const redirect of redirects) {
    keys(redirect, ['path', 'blob', 'record', 'targetPath', 'targetId', 'targetBlob'], 'redirect');
    if (['path', 'blob', 'record', 'targetPath', 'targetId', 'targetBlob'].some((key) => typeof redirect[key] !== 'string')) {
      fail('invalid explicit redirect evidence: identity fields must be strings');
    }
    const redirectMeta = frontmatter(redirect.record);
    if (!canonicalPath(redirect.path) || !redirect.path.startsWith(`tasks/${source.id}-`) || seen.has(redirect.path)
        || !SHA.test(redirect.blob) || blobHash(redirect.record, redirect.blob.length === 64 ? 'sha256' : 'sha1') !== redirect.blob
        || !/^# .*noncanonical redirect\s*$/mi.test(redirect.record)
        || (redirectMeta && ['id', 'owner', 'status'].some((key) => Object.hasOwn(redirectMeta, key)))
        || !TASK.test(redirect.targetId) || !canonicalPath(redirect.targetPath)
        || !redirect.targetPath.startsWith(`tasks/${redirect.targetId}-`) || !SHA.test(redirect.targetBlob)
        || declaredRedirectTarget(redirect.record, redirect.path, source.repository) !== redirect.targetPath) fail('invalid explicit redirect evidence');
    seen.add(redirect.path);
  }
  return source;
}
export const eventPath = (event) => `${ROOT}/${event.task.id}/${event.runId}/${event.eventId}.json`;
export function validateEvent(event, sources, path = eventPath(event)) {
  keys(event, ['formatVersion', 'eventId', 'runId', 'kind', 'parents', 'sequence', 'task', 'repository', 'branch', 'pullRequest', 'at', 'summary', 'next', 'files', 'validation', 'snapshot'], 'event');
  if (event.formatVersion !== 2 || typeof event.eventId !== 'string' || typeof event.runId !== 'string' || !UUID.test(event.eventId) || !UUID.test(event.runId)
      || !['start', 'checkpoint', 'complete', 'abandon'].includes(event.kind)
      || !Number.isSafeInteger(event.sequence) || event.sequence < 0
      || !Array.isArray(event.parents) || new Set(event.parents).size !== event.parents.length || event.parents.some((id) => typeof id !== 'string' || !UUID.test(id))) fail('invalid event identity or causal fields');
  keys(event.task, ['repository', 'id', 'path', 'revision', 'blob', 'claim'], 'event task');
  keys(event.task.claim, ['heading', 'digest'], 'event claim');
  const source = sources.get(`${ROOT}/${event.task.id}/sources/${event.task.revision}.json`);
  if (!source || path !== eventPath(event) || ['repository', 'id', 'path', 'revision', 'blob'].some((key) => event.task[key] !== source[key])) fail('event task/source identity mismatch');
  const claim = claimSections(source.record).get(event.task.claim.heading);
  if (!claim || claim.digest !== event.task.claim.digest) fail('event claim is not the pinned source section');
  if (!REPO.test(event.repository) || !repositories(source.record).includes(event.repository)
      || !nonempty(event.branch) || /[\0\r\n]/.test(event.branch)) fail('invalid code repository or branch');
  if (event.pullRequest !== null && event.pullRequest !== `https://github.com/${event.repository}/pull/${event.pullRequest?.split('/').at(-1)}`) fail('PR must link to the code repository');
  if (event.pullRequest !== null && !/\/pull\/[1-9][0-9]*$/.test(event.pullRequest)) fail('invalid PR URL');
  if (!nonempty(event.summary) || !nonempty(event.next) || typeof event.at !== 'string'
      || !Number.isFinite(Date.parse(event.at)) || new Date(event.at).toISOString() !== event.at) fail('invalid event summary, next action or timestamp');
  filesList(event.files);
  if (event.kind === 'complete') {
    if (!nonempty(event.validation) || !object(event.snapshot) || Object.keys(event.snapshot).length !== event.files.length) fail('completion needs declared validation and exact snapshots');
    for (const file of event.files) {
      if (!Object.hasOwn(event.snapshot, file)) fail(`missing completion snapshot: ${JSON.stringify(file)}`);
      const item = event.snapshot[file];
      keys(item, ['blob', 'mode'], 'file snapshot');
      if (Object.keys(item).length !== 2 || !((item.blob === null && item.mode === null)
          || (typeof item.blob === 'string' && SHA.test(item.blob) && MODES.has(item.mode)))) fail('invalid file blob/mode snapshot');
    }
  } else if (event.validation !== undefined || event.snapshot !== undefined) fail('only completion events carry validation/snapshots');
  return event;
}
const sameRunIdentity = (a, b) => ['repository', 'id', 'path', 'revision', 'blob'].every((key) => a.task[key] === b.task[key])
  && a.task.claim.heading === b.task.claim.heading && a.task.claim.digest === b.task.claim.digest
  && a.repository === b.repository && a.branch === b.branch
  && (a.pullRequest === null || a.pullRequest === b.pullRequest);
export function buildGraph(events, sources) {
  const byId = new Map();
  const runs = new Map();
  const identities = new Map();
  for (const event of events) {
    validateEvent(event, sources);
    if (byId.has(event.eventId)) fail(`duplicate event ID: ${event.eventId}`);
    byId.set(event.eventId, event);
    const key = `${event.task.repository}:${event.task.id}`;
    const identity = identities.get(key);
    if (identity && identity !== event.task.path) fail(`canonical task path changed for ${event.task.id}`);
    identities.set(key, event.task.path);
    const run = runs.get(event.runId) || [];
    run.push(event); runs.set(event.runId, run);
  }
  const children = new Map([...byId.keys()].map((id) => [id, []]));
  for (const event of events) {
    if (event.kind === 'start' ? event.sequence !== 0 : event.parents.length !== 1) fail('invalid start or ordinary-event parent count');
    for (const parentId of event.parents) {
      const parent = byId.get(parentId);
      if (!parent) fail(`missing parent: ${parentId}`);
      if (parent.task.id !== event.task.id || parent.task.repository !== event.task.repository || parent.task.path !== event.task.path) fail('cross-task causal parent');
      if (event.kind === 'start') {
        if (event.runId === parent.runId) fail('a new start must use a new run');
      } else {
        if (event.runId !== parent.runId || event.sequence !== parent.sequence + 1
            || !['start', 'checkpoint'].includes(parent.kind) || !sameRunIdentity(parent, event)) fail('illegal run transition or changed run identity');
        if (event.kind !== 'checkpoint' && event.files.some((file) => !parent.files.includes(file))) fail('completion/abandon scope exceeds its parent; checkpoint first');
      }
      children.get(parentId).push(event.eventId);
    }
  }
  for (const run of runs.values()) if (run.filter((event) => event.kind === 'start').length !== 1) fail('each run needs exactly one start');
  const visiting = new Set(); const visited = new Set();
  const visit = (id) => {
    if (visiting.has(id)) fail('causal cycle');
    if (visited.has(id)) return;
    visiting.add(id); for (const parent of byId.get(id).parents) visit(parent);
    visiting.delete(id); visited.add(id);
  };
  for (const id of byId.keys()) visit(id);
  for (const event of events) {
    if (event.kind !== 'start') continue;
    const parents = new Set(event.parents);
    for (const parent of event.parents) {
      const pending = [...byId.get(parent).parents]; const seen = new Set();
      while (pending.length) {
        const ancestor = pending.pop();
        if (parents.has(ancestor)) fail('reconciliation parents must be competing heads, not ancestors of one another');
        if (seen.has(ancestor)) continue;
        seen.add(ancestor); pending.push(...byId.get(ancestor).parents);
      }
    }
    if (event.parents.length === 1
        && ['start', 'checkpoint'].includes(byId.get(event.parents[0]).kind)) {
      fail('a single active head must be resumed or abandoned before a new run');
    }
  }
  const tasks = new Map();
  for (const event of events) {
    const task = tasks.get(event.task.id) || { events: [], heads: [] };
    task.events.push(event);
    if (!children.get(event.eventId).length) task.heads.push(event);
    tasks.set(event.task.id, task);
  }
  for (const task of tasks.values()) task.heads.sort((a, b) => a.eventId.localeCompare(b.eventId));
  return { byId, runs, children, tasks };
}
export function completionHeads(graph, freshIds) {
  const result = [];
  for (const [id, task] of graph.tasks) {
    if (task.heads.length !== 1) fail(`unresolved frontier for ${id}: ${task.heads.map((event) => event.eventId).join(', ')}`);
    const event = task.heads[0];
    if (event.kind === 'complete' && freshIds.has(event.eventId)) result.push(event);
  }
  return result;
}
