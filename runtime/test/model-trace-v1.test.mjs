import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve, dirname } from 'node:path';
import { evaluate } from '../src/state.mjs';
import { verify, render, load, checkViews } from '../model-trace-v1/check.mjs';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
const read = (p) => readFileSync(resolve(root,p), 'utf8');
const path = 'runtime/model-trace-v1/';
const rawOracle = read(`${path}testdata/oracle.json`);
const oracle = JSON.parse(rawOracle);
const manifest = JSON.parse(read(`${path}manifest.json`));
const files = load();
const pinned = '09b152afafec087c7101e71b26ff89a9a531288a45cafa1382d0063c4c0f411b';

// The oracle was independently stated and committed before this test. These
// selected assertions are not exhaustive evaluator behavior or human review.
test('frozen shadow trace replays through the one existing evaluator', () => {
  assert.equal(createHash('sha256').update(rawOracle).digest('hex'), pinned);
  assert.deepEqual(oracle.cases.map((c) => c.id), [
    'admit_read_only_pending', 'assign_dependencies_missing', 'assign_dependencies_false',
    'assign_resources_missing', 'assign_resources_false',
  ]);
  for (const c of oracle.cases) {
    const {current, command, context} = structuredClone(c.input);
    const before = structuredClone({current, command, context});
    if (c.expected.completion === 'return') {
      const result = evaluate(current, command, context);
      for (const [key,value] of Object.entries(c.expected.executionSelected)) assert.deepEqual(result.execution[key], value, `${c.id}: ${key}`);
      assert.deepEqual(result.execution.spec, command.payload.spec);
      assert.notStrictEqual(result.execution.spec, command.payload.spec);
      assert.deepEqual(result.event, c.expected.event);
      assert.deepEqual(result.effects, c.expected.effects);
    } else {
      assert.equal(c.expected.completion, 'throw');
      let returned = false;
      assert.throws(() => { evaluate(current, command, context); returned = true; }, (error) => {
        assert.equal(error.code,c.expected.error.code, c.id);
        assert.equal(error.message,c.expected.error.message,c.id);
        return true;
      });
      assert.equal(returned,false);
      assert.equal(c.expected.resultReturned,false);
      assert.equal(c.expected.eventReturned,false);
      assert.equal(c.expected.effectsReturned,false);
      for (const [key,value] of Object.entries(c.expected.currentSelectedAfter)) assert.deepEqual(current[key],value,`${c.id}: ${key}`);
    }
    assert.deepEqual({current, command, context}, before, `${c.id}: inputs mutated`);
  }
});

test('source-pinned trace covers the schema inventory and generated views', () => {
  assert.equal(verify(manifest, files),true);
  for (const [name, expected] of Object.entries(render(manifest))) assert.equal(read(name),expected,`${name} is stale`);
});

test('trace controls reject drift, omissions and dangling evidence', () => {
  assert.throws(() => verify(manifest, {...files, 'runtime/src/state.mjs': `${files['runtime/src/state.mjs']}\n`}), /source pin/);
  const mutant = (change, expected) => { const copy = structuredClone(manifest); change(copy); assert.throws(() => verify(copy,files),expected); };
  mutant((m) => m.commands.pop(), /command inventory/);
  mutant((m) => m.states.push({id:'MADE_UP',coverage:'unmapped'}), /state inventory/);
  mutant((m) => {m.commands.find((x) => x.id === 'REVOKE_APPROVAL').coverage = 'partial';}, /command coverage/);
  mutant((m) => {m.requirements[0].schemaPointer = '#/$defs/Command/oneOf/1';}, /schema reference/);
  mutant((m) => {m.requirements[0].anchor = 'no such evaluator source';}, /source anchor/);
  mutant((m) => {m.requirements[0].caseIds = ['not_a_case'];}, /case reference/);
  mutant((m) => {m.requirements[0].existingTest = 'not a test';}, /test reference/);
  mutant((m) => {m.requirements[1].caseIds = ['assign_resources_missing'];}, /case outcome/);
  mutant((m) => {m.flows[1].requirementIds = ['TR-ADMIT'];}, /flow reference/);
  const views = render(manifest);
  assert.equal(checkViews(manifest, views),true);
  assert.throws(() => checkViews(manifest, {...views, [`${path}views/sequence.md`]: `${views[`${path}views/sequence.md`]}tamper`}), /stale view/);
  assert.throws(() => checkViews(manifest, {...views, [`${path}views/state.dot`]: 'corrupted'}), /stale view/);
});

const dispatchRaw = read(`${path}testdata/dispatch-oracle.json`);
const dispatchOracle = JSON.parse(dispatchRaw);
const dispatchPin = '4afde4fc00e01c2c54ce060a7d118883a6480b7e50856a11717b3b151e27edd3';

function assertDispatchCases(apply, cases = dispatchOracle.cases) {
  for (const c of cases) {
    const input = structuredClone(c.input);
    const before = structuredClone(input);
    const invoke = () => apply(input.current, input.command, input.context);
    if (c.expected.completion === 'return') {
      let result;
      assert.doesNotThrow(() => { result = invoke(); }, `${c.id}: successful return expected`);
      assert.deepEqual(result.execution, c.expected.execution, `${c.id}: execution`);
      assert.deepEqual(result.event, c.expected.event, `${c.id}: event`);
      assert.deepEqual(result.effects, c.expected.effects, `${c.id}: effects`);
    } else {
      assert.throws(invoke, (error) => {
        assert.equal(error.code, c.expected.error.code, `${c.id}: refusal code`);
        assert.equal(error.message, c.expected.error.message, `${c.id}: refusal message`);
        return true;
      }, `${c.id}: refusal expected`);
      assert.deepEqual(input.current, c.expected.currentAfter, `${c.id}: refused current`);
      assert.equal(c.expected.resultReturned, false);
      assert.equal(c.expected.eventReturned, false);
      assert.equal(c.expected.effectsReturned, false);
    }
    assert.deepEqual(input, before, `${c.id}: inputs changed`);
  }
}

test('frozen dispatch observations compare full execution, event, effects and unchanged inputs', () => {
  assert.equal(createHash('sha256').update(dispatchRaw).digest('hex'), dispatchPin);
  assert.deepEqual(dispatchOracle.cases.map((c) => c.id), [
    'assign_success', 'ack_success', 'submit_success', 'ack_wrong_worker',
    'ack_stale_fence', 'ack_expired_lease', 'submit_wrong_worker',
    'submit_stale_fence', 'submit_expired_lease', 'submit_wrong_repository_subject',
    'submit_wrong_artifact_digest', 'submit_wrong_spec_digest',
  ]);
  assertDispatchCases(evaluate);
});

test('dispatch trace controls reject reference, outcome and coverage drift', () => {
  const mutant = (change, expected) => {
    const copy = structuredClone(manifest); change(copy);
    assert.throws(() => verify(copy, files), expected);
  };
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-ACK-WRONG-WORKER').caseIds = ['submit_wrong_worker']; }, /case command/);
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-SUBMIT-SUCCESS').expectedCode = 'UNAUTHORIZED'; }, /requirement outcome/);
  mutant((m) => { m.flows.find((f) => f.id === 'submission').to = 'COMPLETED'; }, /flow reference/);
  mutant((m) => { m.states.find((s) => s.id === 'VALIDATING').coverage = 'unmapped'; }, /state coverage/);
  const drift = `${files[`${path}testdata/dispatch-oracle.json`]}\n`;
  assert.throws(() => verify(manifest, {...files, [`${path}testdata/dispatch-oracle.json`]: drift}), /source pin/);
});

// Mutants run only trusted local source in memory. The original source and
// snapshot pins remain unchanged, so rejection comes from behavioral assertions.
test('dispatch observations detect compiling worker, fence, expiry and digest guard mutants', async () => {
  const source = read('runtime/src/state.mjs');
  const contractURL = new URL('../src/contract.mjs', import.meta.url).href;
  const mutations = [
    ["authorized(context.role === 'worker' && context.actor_id === execution.attempt.worker_id);", 'authorized(true);', 'ack_wrong_worker'],
    ["input.fencing_token === execution.attempt.fencing_token, 'STALE_ATTEMPT'", "true, 'STALE_ATTEMPT'", 'ack_stale_fence'],
    ["now < millis(deadline), 'ATTEMPT_EXPIRED'", "now <= millis(deadline), 'ATTEMPT_EXPIRED'", 'ack_expired_lease'],
    ["result.subject_digest === digest(result.artifacts), 'ARTIFACT_DIGEST_MISMATCH'", "true, 'ARTIFACT_DIGEST_MISMATCH'", 'submit_wrong_artifact_digest'],
  ];
  for (const [from, to, id] of mutations) {
    assert.ok(source.includes(from), `${id}: mutation anchor`);
    let changed = source.replaceAll(from, to);
    if (id === 'ack_stale_fence') changed = changed.replaceAll('input.fencing_token === execution.fence', 'true');
    const mutated = changed.replace("'./contract.mjs'", JSON.stringify(contractURL));
    const module = await import(`data:text/javascript;base64,${Buffer.from(mutated).toString('base64')}`);
    const cases = dispatchOracle.cases.filter((c) => c.id === id);
    assert.equal(cases.length, 1);
    assert.throws(() => assertDispatchCases(module.evaluate, cases), {code: 'ERR_ASSERTION'}, `${id}: mutant escaped`);
  }
});

const completionRaw = read(`${path}testdata/completion-oracle.json`);
const completionOracle = JSON.parse(completionRaw);
const completionPin = 'd81eb8f2417e4c5fcd5df284eabc5aac8bc4a48f9110795020b2a430b487c3e8';

test('frozen completion observations compare full outcomes and preserve inputs', () => {
  assert.equal(createHash('sha256').update(completionRaw).digest('hex'), completionPin);
  assert.deepEqual(completionOracle.cases.map((c) => c.id), [
    'validation_pass', 'approval_approve', 'accept_final_completed',
    'cleanup_completed_released', 'accept_nonfinal_pending',
    'validation_same_worker', 'validation_untrusted_check', 'validation_wrong_subject',
    'validation_presubmit_time', 'validation_duplicate', 'approval_wrong_subject',
    'approval_expired', 'accept_missing_check', 'accept_revoked_approval',
    'accept_unregistered_approver', 'accept_unauthorized_owner',
    'cleanup_stale_fence', 'cleanup_not_pending',
  ]);
  assertDispatchCases(evaluate, completionOracle.cases);
  // Verify the supplied single-stage observations form a connected path from
  // the original submitted execution; no observation is synthesized by replay.
  let prior = dispatchOracle.cases.find((c) => c.id === 'submit_success').expected.execution;
  for (const id of ['validation_pass', 'approval_approve', 'accept_final_completed', 'cleanup_completed_released']) {
    const c = completionOracle.cases.find((c) => c.id === id);
    assert.deepEqual(c.input.current, prior, `${id}: disconnected frozen input`);
    prior = c.expected.execution;
  }
});

test('completion trace preserves partial inventories, references and cleanup distinctions', () => {
  const mutant = (change, expected) => {
    const copy = structuredClone(manifest); change(copy);
    assert.throws(() => verify(copy, files), expected);
  };
  mutant((m) => { m.commands.find((c) => c.id === 'REQUEST_CHANGES').coverage = 'partial'; }, /command coverage/);
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-VALIDATION-PASS').caseIds = ['approval_approve']; }, /case command/);
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-ACCEPT-UNAUTHORIZED-OWNER').anchor = "authorized(context.role === 'worker' && context.actor_id === execution.attempt.worker_id);"; }, /guard outcome reference/);
  mutant((m) => { m.flows.find((f) => f.id === 'nonfinal-acceptance').to = 'COMPLETED'; }, /flow reference/);
  const views = render(manifest);
  assert.match(views[`${path}views/sequence.md`], /COMPLETED does not release resources/);
  assert.match(views[`${path}views/state.dot`], /COMPLETED \(partial; cleanup may remain\)/);
  const drift = `${files[`${path}testdata/completion-oracle.json`]}\n`;
  assert.throws(() => verify(manifest, {...files, [`${path}testdata/completion-oracle.json`]: drift}), /source pin/);
});

// These behavioral mutations bypass source pins and never change disk files or
// deliver effect intents. Each must fail the frozen comparison after compiling.
test('completion observations detect independence, check, approval and cleanup guard mutants', async () => {
  const source = read('runtime/src/state.mjs');
  const contractURL = new URL('../src/contract.mjs', import.meta.url).href;
  const mutations = [
    ["requireGuard(evidence.runner_id !== execution.attempt.worker_id, 'REVIEW_NOT_INDEPENDENT');", "requireGuard(true, 'REVIEW_NOT_INDEPENDENT');", 'validation_same_worker'],
    ["requireGuard(checks().every((check) => execution.validations.some((item) => item.check_id === check && item.verdict === 'PASS')), 'CHECKS_NOT_PASSED');", "requireGuard(true, 'CHECKS_NOT_PASSED');", 'accept_missing_check'],
    ["item.decision === 'APPROVE' && !item.revoked &&", "item.decision === 'APPROVE' &&", 'accept_revoked_approval'],
    ['matchingAttempt(payload, false);', 'void payload;', 'cleanup_stale_fence'],
  ];
  for (const [from, to, id] of mutations) {
    assert.equal(source.split(from).length, 2, `${id}: unique mutation anchor`);
    const mutated = source.replace(from, to).replace("'./contract.mjs'", JSON.stringify(contractURL));
    const module = await import(`data:text/javascript;base64,${Buffer.from(mutated).toString('base64')}`);
    const cases = completionOracle.cases.filter((c) => c.id === id);
    assert.equal(cases.length, 1);
    assert.throws(() => assertDispatchCases(module.evaluate, cases), {code: 'ERR_ASSERTION'}, `${id}: mutant escaped`);
  }
});

const cancellationRaw = read(`${path}testdata/cancellation-oracle.json`);
const cancellationOracle = JSON.parse(cancellationRaw);
const cancellationPin = 'f2555364a4b53001c5e6dcfadf5822b2014c0e3fd1e26d14432eb714b681a044';

test('frozen cancellation observations compare full outcomes and preserve inputs', () => {
  assert.equal(createHash('sha256').update(cancellationRaw).digest('hex'), cancellationPin);
  assert.deepEqual(cancellationOracle.cases.map((c) => c.id), [
    'cancel_pending_without_attempt', 'cancel_active_retains_cleanup',
    'cleanup_cancelled_released', 'cancel_worker_unauthorized',
    'cancel_wrong_owner_unauthorized', 'cleanup_cancelled_stale_fence',
    'submit_cancelled_retired_token', 'submit_cancelled_current_fence',
  ]);
  assertDispatchCases(evaluate, cancellationOracle.cases);
  const byId = (id) => cancellationOracle.cases.find((c) => c.id === id);
  // Connections compare independently frozen literals, never replay output.
  assert.deepEqual(byId('cancel_pending_without_attempt').input.current,
    dispatchOracle.cases.find((c) => c.id === 'assign_success').input.current);
  assert.deepEqual(byId('cancel_active_retains_cleanup').input.current,
    dispatchOracle.cases.find((c) => c.id === 'ack_success').expected.execution);
  for (const id of ['cleanup_cancelled_released', 'cleanup_cancelled_stale_fence',
    'submit_cancelled_retired_token', 'submit_cancelled_current_fence']) {
    assert.deepEqual(byId(id).input.current, byId('cancel_active_retains_cleanup').expected.execution,
      `${id}: disconnected frozen input`);
  }
});

test('cancellation trace rejects mapping drift and preserves terminal cleanup distinctions', () => {
  const mutant = (change, expected) => {
    const copy = structuredClone(manifest); change(copy);
    assert.throws(() => verify(copy, files), expected);
  };
  mutant((m) => { m.commands.find((c) => c.id === 'CANCEL').coverage = 'unmapped'; }, /command coverage/);
  mutant((m) => { m.states.find((s) => s.id === 'CANCELLED').coverage = 'unmapped'; }, /state coverage/);
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-CANCEL-ACTIVE').caseIds = ['cleanup_cancelled_released']; }, /case command/);
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-CANCEL-WORKER').anchor = "authorized(context.role === 'worker' && context.actor_id === execution.attempt.worker_id);"; }, /guard outcome reference/);
  mutant((m) => { m.requirements.find((r) => r.id === 'TR-SUBMIT-CANCELLED-RETIRED-TOKEN').expectedCode = 'STALE_ATTEMPT'; }, /requirement outcome/);
  mutant((m) => { m.flows.find((f) => f.id === 'cancelled-cleanup').to = 'PENDING'; }, /flow reference/);
  const views = render(manifest);
  assert.match(views[`${path}views/sequence.md`], /confirmation uses the retired attempt token, not the advanced execution fence/);
  assert.match(views[`${path}views/state.dot`], /CANCELLED \(partial; cleanup may remain\)/);
  const drift = `${files[`${path}testdata/cancellation-oracle.json`]}\n`;
  assert.throws(() => verify(manifest, {...files, [`${path}testdata/cancellation-oracle.json`]: drift}), /source pin/);
});

// These controls compile altered trusted source only in memory. They bypass pin
// checking, preserve every producer byte and never deliver returned intents.
test('cancellation observations detect authorization, retirement, release, cleanup and terminal mutants', async () => {
  const source = read('runtime/src/state.mjs');
  const contractURL = new URL('../src/contract.mjs', import.meta.url).href;
  const mutations = [
    ["authorized(); retire(); execution.state = kind === 'FAIL'", "authorized(true); retire(); execution.state = kind === 'FAIL'", 'cancel_worker_unauthorized'],
    ["execution.fence += 1;\n        execution.cleanup_pending = true;", "execution.fence += 0;\n        execution.cleanup_pending = true;", 'cancel_active_retains_cleanup'],
    ["execution.cleanup_pending = true;", "execution.cleanup_pending = true; execution.resources_held = false;", 'cancel_active_retains_cleanup'],
    ['matchingAttempt(payload, false);', 'matchingAttempt(payload);', 'cleanup_cancelled_released'],
    ['matchingAttempt(payload, false);', 'void payload;', 'cleanup_cancelled_stale_fence'],
    ["!TERMINAL_STATES.includes(current.state) || kind === 'CLEANUP_CONFIRMED'", "current.state === 'CANCELLED' || !TERMINAL_STATES.includes(current.state) || kind === 'CLEANUP_CONFIRMED'", 'submit_cancelled_retired_token'],
  ];
  for (const [from, to, id] of mutations) {
    assert.equal(source.split(from).length, 2, `${id}: unique mutation anchor`);
    const mutated = source.replace(from, to).replace("'./contract.mjs'", JSON.stringify(contractURL));
    const module = await import(`data:text/javascript;base64,${Buffer.from(mutated).toString('base64')}`);
    const cases = cancellationOracle.cases.filter((c) => c.id === id);
    assert.equal(cases.length, 1);
    assert.throws(() => assertDispatchCases(module.evaluate, cases), {code: 'ERR_ASSERTION'}, `${id}: mutant escaped`);
  }
});
