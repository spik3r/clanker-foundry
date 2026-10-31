import test from 'node:test';
import assert from 'node:assert/strict';
import { makeSource, blobHash } from '../../skills/project-memory/scripts/lib/state.mjs';
import { importReviewedTask, exportTaskView } from '../src/import-export.mjs';
import { taskSpec, taskExecution } from './fixtures.mjs';

const revision = 'a'.repeat(40);
function source(status = 'in-progress') {
  const record = `---\nid: ${taskSpec.task_id}\nstatus: ${status}\nowner: operator\nrepositories: [${taskSpec.repository.repository_id}]\n---\n\n# Inspect the pilot\n\n## Implementation claim\n\nOwner operator; read-only review.\n`;
  const path = `tasks/${taskSpec.task_id}-example.md`;
  return makeSource({
    repository: 'example/tasks', revision, id: taskSpec.task_id, path,
    records: new Map([[path, { text: record, blob: blobHash(record) }]]),
  });
}
const args = () => ({ source: source(), spec: taskSpec, reviewedBy: 'operator',
  expectedRepository: 'example/tasks', expectedRevision: revision });

test('a reviewed pinned source produces a proposal without copying the backlog or claim text', () => {
  const input = args();
  const before = structuredClone(input);
  const result = importReviewedTask(input);
  assert.equal(result.mode, 'PROPOSAL');
  assert.equal(result.source.revision, revision);
  assert.equal(result.source.claim.heading, 'Implementation claim');
  assert.equal(Object.hasOwn(result.source, 'record'), false);
  assert.deepEqual(input, before);
});

test('import rejects stale provenance, mismatched identity, missing review and terminal sources', () => {
  for (const changes of [
    { expectedRevision: 'b'.repeat(40) }, { expectedRepository: 'other/tasks' },
    { reviewedBy: 'other' }, { source: source('done') },
    { spec: { ...taskSpec, task_id: 'CF-different' } },
    { spec: { ...taskSpec, repository: { ...taskSpec.repository, repository_id: 'other/code' } } },
  ]) assert.throws(() => importReviewedTask({ ...args(), ...changes }));
  const forged = args();
  forged.source.record += '\nchanged bytes';
  assert.throws(() => importReviewedTask(forged));
});

test('export labels revision and authority while remaining outside the canonical task format', () => {
  const view = exportTaskView(taskExecution);
  assert.match(view, /Read-only projection/);
  assert.match(view, /Revision: 0/);
  assert.equal(view.startsWith('---'), false);
  assert.throws(() => exportTaskView({ ...taskExecution, spec_digest: 'f'.repeat(64) }));
});
