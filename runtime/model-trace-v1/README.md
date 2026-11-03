# Bounded orchestration review trace

This is a first, **partial** machine-readable map of the existing shared
orchestration evaluator. It connects a small read-only admission and refusal
sequence to the v1 command schema, exact evaluator source, named existing tests,
and invented frozen replay cases. It does not replace, generate or interpret
runtime behavior. It calls the existing evaluator only in tests; no dispatcher,
provider, worker, store or candidate model is started.

The manifest marks ADMIT, ASSIGN, ACKNOWLEDGE and SUBMIT as `partial`, alongside
PENDING, ASSIGNED, EXECUTING and VALIDATING. Every other schema command and state
remains `unmapped`. Partial is not a coverage percentage or a claim that every
guard or outcome is represented. Cancellation, cleanup, retries, validation,
approval and acceptance remain outside this trace.

The original five cases cover admission and missing/false dependency and resource
assertions. The additive [dispatch observations](testdata/dispatch-oracle.json)
cover successful read-only assignment, acknowledgement and successful submission,
plus wrong-worker, stale-fence, exact-lease-deadline and artifact/specification
binding refusals. Successful assignment returns a `LAUNCH_ATTEMPT` intent and
successful submission returns a `RUN_CHECKS` intent. Tests compare those intents;
they never deliver them. Context flags are caller assertions, not evidence of
real resources, authenticated identities or authority. REFUSED in the views is
an outcome marker, not a runtime state; every refusal preserves its supplied
execution and returns no receipt, event or effect list.

## Review and check

On a trusted single-writer checkout with Node 24 and the runtime's already
installed locked dependencies, from the repository root:

```sh
node runtime/model-trace-v1/check.mjs --check
node --test runtime/test/model-trace-v1.test.mjs
```

The checker reads only a fixed list of tracked source paths. It verifies their
exact SHA-256 pins, obtains command kinds and states from the actual v1 schema,
checks that the manifest inventories each exactly once, validates declared
schema/fixture/test/source references, and compares generated views byte for
byte. The normal runtime `npm test` picks up the replay test, so this check is
also exercised by the existing test workflow. No CI configuration is changed.

The [DOT source](views/state.dot) and [sequence view](views/sequence.md) are
generated from the manifest's limited declared flow; they are review aids, not
an executable model, complete graph or UML, SysML, BPMN, UTP or fUML conformance.
When changing the manifest, deliberately review both source and references,
then regenerate the fixed views in the same trusted checkout:

```sh
node runtime/model-trace-v1/check.mjs --write-views
node runtime/model-trace-v1/check.mjs --check
```

The script assumes a trusted local checkout and is not a sandbox or defense
against hostile concurrent file changes or symlinks. The checker does not parse
arbitrary untrusted input and has no network or subprocess behavior. A failed
check is not authority to update pins blindly. Review changed upstream sources,
contract and fixtures before deliberately revising this trace.

## What the evidence means

The [original frozen observations](testdata/README.md) and additive dispatch
observations were written by separate assistant contributors from static source
and contract reading before the corresponding trace implementation. The original
five-case oracle is unchanged. The twelve dispatch cases compare the full
execution, exact event and effects or refusal code/message, plus unchanged inputs.
The expected cases were not derived from prior runs. This is separate-assistant
authorship, not independent-person review or an oracle independent of the
evaluator's intended semantics.

Four in-memory source mutations weaken worker identity, both fence checks, lease
expiry and artifact-digest checks. The frozen observations must detect each by
behavioral disagreement; those tests bypass source-pin checking and do not write
the evaluator file. The wrong-repository fixture also has a mismatched manifest
digest: its expected subject refusal tests the earlier guard's precedence. These
finite examples do not cover every refusal or prove the absence of other defects.

Pins bind this reviewed snapshot to exact bytes and make drift visible; they do
not authenticate source or show that a linked test ran. The test-name and source
anchors are reviewed declarations and literal references, not an analysis that
proves all guard logic is mapped. Only the replay tests actually run selected
examples. A valid check, hash, view, or matching fixture does not prove model
completeness, correctness, real identity, permission, runtime admission or
execution. No worker is launched, and no integration or deployment is implied.
