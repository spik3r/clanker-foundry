# Bounded orchestration review trace

This is a first, **partial** machine-readable map of the existing shared
orchestration evaluator. It connects a small read-only admission and refusal
sequence to the v1 command schema, exact evaluator source, named existing tests,
and invented frozen replay cases. It does not replace, generate or interpret
runtime behavior. It calls the existing evaluator only in tests; no dispatcher,
provider, worker, store or candidate model is started.

The manifest deliberately marks ADMIT and ASSIGN as `partial`, PENDING as
`partial`, and every other schema command and state as `unmapped`. Partial is
not a coverage percentage or a claim that every guard or outcome is represented.
In particular, successful assignment, cancellation, cleanup, fencing, effects
and every other transition remain outside this slice. The known ASSIGN sequence
only records that missing/false dependency assertions refuse first, and that
missing/false resource assertions refuse after dependencies are asserted true.
These flags are caller-supplied assertions, not evidence of real resources or
authorization. No `LAUNCH_ATTEMPT` effect is returned by these refusal cases. REFUSED in the
views is an outcome marker, not a runtime state; the invented input remains
PENDING at revision 1 after refusal.

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

The [frozen observations](testdata/README.md) were written by a separate assistant
from static source and contract reading before this trace implementation. The
replay test calls `evaluate` with those original invented inputs and checks the
selected execution fields, exact event, effects, refusal codes and unchanged
inputs. The separately stated expected cases were not derived from a prior
run. This is separation between assistant contributors, not independent-person
review or an oracle independent of the evaluator's intended semantics.

Pins bind this reviewed snapshot to exact bytes and make drift visible; they do
not authenticate source or show that a linked test ran. The test-name and source
anchors are reviewed declarations and literal references, not an analysis that
proves all guard logic is mapped. Only the replay tests actually run selected
examples. A valid check, hash, view, or matching fixture does not prove model
completeness, correctness, real identity, permission, runtime admission or
execution. No worker is launched, and no integration or deployment is implied.
