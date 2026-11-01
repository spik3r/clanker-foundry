# An external consumer with its own generated model

This test shows a separate Go module importing Foundry's scenario and acceptance
libraries while using its own generated reducers. It does not replace Foundry's
`generated` package, copy its source into a consumer, add a business-rule
interpreter, or select an engine from request data.

The [invented requirements](testdata/requirements.md) describe Candidate,
Reviewing and Accepted states with submission, acceptance and return triggers.
The [first model](testdata/model-v1.json) permits any nonblank submit actor; the
[second](testdata/model-v2.json) changes one requirement and one guard so only the
literal `maker` actor may submit. Neither model is a real consumer policy. This
is not Heisentick adoption, task admission, deployment or execution authority.

## Run the proof

From `modeling/lifecycle`, with installed Go 1.27.3 and a writable local cache:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 -v ./consumerintegration
```

The test creates a disposable external module under Go's test temporary directory.
A local `replace` directive points its ordinary Foundry dependency to this checkout;
that is not a replacement of Foundry's generated reducer. The unchanged
`compiler.Generate` creates two distinct custom reducer packages and their derived
structural tests. A [field-only adapter template](testdata/adapter.go.txt) converts
between those generated types and Foundry's scenario transport types. Its actual
transition call goes exclusively to the custom generated `Transition` function.

The external module runs with offline module resolution, local toolchain selection
and workspace/configuration mode disabled; test builds omit VCS stamping so a
temporary directory's unrelated Git ancestor cannot contaminate the proof. It
downloads or installs nothing. A three-minute deadline requests cancellation of
each Go subprocess; a two-second WaitDelay limits lingering pipe waits. This is
not a guarantee of descendant termination, cleanup, or callback cancellation.
The test uses the same installed Go executable as its parent.

## Evidence and its limits

A separate assistant authored both models and the literal results in the
[first suite](testdata/suite-v1.json) and [second suite](testdata/suite-v2.json)
before generated replay. The [manifests](testdata/manifest-v1.json) and
[second manifest](testdata/manifest-v2.json) bind exact suite source bytes to the
respective canonical models. The [expected inventory](testdata/expected.json)
pins those identities and the documented scenario digest, without deriving
expected transitions from actual output. This is separation between assistants,
not independent-person verification or proof that the chosen requirements are
correct. The test pins the frozen inventory's SHA-256 and never rewrites fixtures.

Each suite has seven cases and fourteen steps. They exercise successful review,
self-review refusal, terminal-state refusal, rework and stale review, changed
same-revision content, foreign-model review, unknown input, revision overflow and
the changed submit-actor rule. The test checks:

- Actual custom-model scenario replay and opt-in acceptance replay match all
  fourteen literal steps, with exact model/source/scenario identities.
- Cross-model suites refuse before callback invocation. Cross-model acceptance,
  stale source bytes and check-only inputs retain their non-execution statuses.
- Whole-suite preflight rejects a malformed final expectation before any callback.
- An intentionally incorrect in-memory expectation stops on the first comparison
  mismatch. Neither frozen suite is altered.
- A callback returning a wrong model digest causes a provider-protocol error after
  one invocation; acceptance does not hide it as ordinary mismatch or not-run.
- Formatting changes affect source identity but not normalized scenario/model
  identity. Existing default-bound APIs do not silently adopt the custom model.
- Adapter field inventories fail if Task, Command, Review or Result gains a field.
  Direct sentinel tests cover every field, maximum uint64 values, nil versus empty
  lists/pointers, refusal order, copied ownership and unchanged result envelopes.
- Two mutations in disposable generated code—always refusing, and removing the
  submit-actor refusal—must be rejected by the frozen oracle's actual comparison.
  A build error does not count. The restored reducer must pass again.

The explicit binding is caller-supplied Go composition, not authenticated compiled
code provenance. Bound reports label it `caller-supplied-go-binding`, not
`compiledModelDigest`. Actor names, matching hashes and matched literals do not
establish real authorship, approval, current deployment or executed-source trust.
Callbacks remain trusted application code: this library does not sandbox, cancel,
roll back or constrain their effects or allocations. A hostile or incorrect
adapter can lie; this test checks the adapter supplied here, not arbitrary ones.
