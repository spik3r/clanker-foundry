# Independently stated first-design review cases

These literals were authored by the separate assistant lane
`/root/review_foundry_pilot` for the prospective claim
`CF-first-design-review-bridge-v1`, source commit
`a1f6828dfe2a52f55636dce8627b90cbbf801d52`. The bridge implementer is a different
assistant lane. This is contributor separation, not independent-person
verification, authenticated authorship, human approval or proof of correctness.

## Inputs and derivation

The base inputs are the existing invented `skills/model-driven-design/examples`
brief, intake, model, suite and manifest at main
`4fe0611d032ea7060e9804b23634d1ec13c481a4`. Their unchanged SHA-256 values are:

- `brief.md`: `ab3d47aaacdcdc8d57c2cb8a9ab432c3f5c22c42b7df11b707b8af03f32eb5de`
- `intake.json`: `9f4d049dc8ccf5de939cc22c802207dcad7bd075cbd5edb8a6b8ffd6592535af`
- `model.json`: `3e7b8c44c10940a7524161dd306d022019e8fc41e793a66a11771d6dd8621239`
- `suite.json`: `184115eead4e94d37dd369d7f07a36ed6cb84080e8a4c9f27be3dd5562e2e7dd`
- `manifest.json`: `68cf7227f055ab2079d9b3b4d7159b6dec8bbd19d8246f33eef8a036d96838ae`

The upstream model/suite/manifest already existed before this composition. The
fixture author read their literal bytes and the documented intake, intake-model,
acceptance-authoring, proposal-check and first-design handoff contracts. The
existing scenario expectations were not authored again and were not executed.
Original inputs are unchanged. The preserved open question `Q_use` blocks real
use under the invented brief; a valid check does not answer or remove it.

No Foundry validator, model generator, reducer, proposal checker or prospective
review bridge was invoked to derive these outcomes. A small Python construction
only reads UTF-8 bytes, builds the specified JSON request strings, checks byte
hash declarations, and serializes the hand-stated expected fields. It does not
interpret the lifecycle or calculate expected transition behavior. The source
hashes above identify byte observations, not execution or provenance.

## Meaning of the cases

Each `request` is the exact outer request text as one UTF-8 string. Its two nested
request fields are also text. Valid nested requests embed original model, suite,
manifest and requirement bytes without parsing/reserializing those inner files.
The intake carrier is assembled from the original declaration with explicitly
supplied observations; scenario and manifest literals are not refreshed.
Whitespace variants append one LF only. Duplicate-key and malformed cases are
constructed as raw strings rather than normalized through a JSON parser.

For valid cases, all expected observations are literal contract predictions:

- Cross-request comparison is exact decoded text equality, not semantic equality.
  An extra model LF therefore mismatches cross-request bytes but preserves valid
  model semantics and proposal preflight. Intake source identity remains stale.
- An extra suite LF invalidates the manifest's raw suite-byte identity. Proposal
  preflight is `stale`, suite validation is `not-checked`, and replay is `not-run`.
  If intake and proposal both carry that changed suite, their cross comparison
  still matches; neither native stale declaration is repaired or hidden.
- Null observations are `missing`. Empty model text is present and `rejected`.
  An intake observation containing `{}` produces a valid native intake-model
  report with a rejected model, whereas a proposal containing that same invalid
  model makes the whole composition invalid with no partial result.
- Intake suite text is opaque to intake-model inspection. A malformed-looking
  observed suite can be compared as text without becoming a proposal validation.
- A missing declaration does not become supplied merely because observed bytes
  match the proposal. The native intake binding stays `undeclared`.
- Original unresolved questions remain visible. The extra-question case verifies
  all question IDs in the intake contract's sorted order, not a ready decision.
- Strict malformed/duplicate/unknown/missing/null/type/Unicode/trailing-input
  refusals cover the outer wrapper and both native request boundaries. A malformed
  proposal suite is not promoted to a valid stale report merely because its raw
  hash differs.

`valid: true` means the composition can produce the two native reports and its
separate byte-comparison observations. It never means an accepted model, current
bindings, review approval, scenario execution or authorization. Every valid case
predicts `preflightReplay: not-run`; this fixture author ran no candidate.

These are independently stated observation vectors, not full expected native
report byte blobs. Implementation tests must separately preserve both complete
native reports, compare input ownership and repeatability, enforce no execution,
exercise resource/output limits and keep the no-authority envelope. The fixtures
cannot by themselves prove those implementation properties or validate arbitrary
real-world requirements. Preserve a failed observation and investigate it rather
than rewriting these expectations to fit a run.
