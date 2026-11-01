---
id: CF-validation-evidence-time
status: in-progress
owner: Astra validation evidence fix
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Preserve validation freshness across intervening commands

## Problem and scope

The shared evaluator compares validation evidence time with the execution's
latest update time. Recording another check or approval advances that timestamp
without changing the submitted subject. This can reject a valid delayed receipt
observed after SUBMIT and before its recording command.

An invented in-memory trace reproduces the defect: SUBMIT at second 3, check A
observed at second 4 and recorded at second 6, then check B observed at second 5
and recorded at second 7. B fails with INVALID_EVIDENCE_TIME. The identical B
receipt succeeds without intervening A. An approval at second 6 causes the same
failure. Commands use current revisions and evidence binds the same live attempt,
subject, specification, repository and registered checker.

This behavior predates the shadow consumer facade. No production runner failure
or operational impact has been demonstrated.

## Implementation claim

Astra validation evidence fix owns this coordinator-reviewed bounded local lane on
`codex/validation-evidence-time`, based on exact merged main
`b4e5e75511bc271ed43ba2ca850ba2b9c230f0f2`.

The active packaging and adaptation lead explicitly confirms no active edits
or claim on `runtime/src/state.mjs` or `runtime/test/state.test.mjs`. Its merged
consumer-surface implementation reserves no further edits. Consumer adaptation
will wait for a fixed, reviewed producer pin rather than duplicate this guard
or add a consumer workaround.

The coordinator accepted this scope and deadline-based design. Current main,
open branches, local changes and canonical ownership were rechecked before
implementation. No competing runtime claim was found. This local claim is not
a published shared reservation.

Exact maintained write scope:

- `runtime/src/state.mjs`
- `runtime/test/state.test.mjs`
- `tasks/CF-validation-evidence-time-fix.md`
- New immutable helper-generated Project Memory source and event files for
  `CF-validation-evidence-time`.

Commit this reviewed canonical task source locally, register its exact revision,
verify its identity and sole claim, and record a genuine start before editing
implementation or regression tests. Preserve the source commit and chronology.

Do not change contracts, schemas, package metadata, the shadow facade, store,
runner, Project Memory helper, dependencies, workflows, other task records or
existing source/event bytes.

## Bounded outcome

Use the fixed validation deadline minus the immutable specification's
validation duration as the SUBMIT-time lower bound for evidence freshness.
Do not introduce a new schema field or a second implementation of the guard.

Preserve command-time monotonicity, current revision checks, future-evidence
refusal, validation expiry, duplicate-check refusal and exact binding to the
attempt, subject, specification, repository, workflow and trusted checker.
Intervening checks or approvals must not move the freshness lower bound or
extend the validation deadline.

## Acceptance

- Multiple independently timed check receipts remain admissible when each was
  observed after SUBMIT, even when another receipt was recorded in between.
- A valid intervening approval does not invalidate otherwise fresh evidence.
- Evidence before SUBMIT and after its recording command is refused.
- Evidence exactly at SUBMIT or exactly at its recording command is accepted
  when all other guards pass.
- Recording evidence exactly at the validation deadline remains expired.
- The validation deadline remains unchanged across checks and approvals.
- Existing identity, subject, revision, independence and expiry refusals remain.
- Accepted and rejected transitions preserve input immutability and deterministic
  replay through the same evaluator.

## Validation

Use invented fixtures and the existing pinned dependency graph. An installed
Node v24.15.0 executable has been verified by execution, and the original
in-memory reproduction confirms the defect on that runtime. Verify the actual
runtime version again for final checks; a directory name is not version evidence.
Do not install or change software as part of this lane.

First run the new regression cases against unchanged evaluator code and record
the genuine before-fix failures. Then run them against the correction.

Run focused state, contract, importer and shadow-facade regressions, then applicable
runtime and Project Memory suites using existing dependencies and synthetic
fixtures. Include independently timed receipts, intervening approval, pre-SUBMIT
and future evidence, millisecond boundaries, expiry and unchanged deadline
assertions. Do not count excluded or unavailable checks as passes.

Run the repository validator, whitespace checks and genuine committed full-diff
Project Memory validation against the reviewed base. Record actual commands,
results and limits in new completion evidence. Obtain independent review of
the exact final diff before handoff.

## Integration and exclusions

The coordinator owns integration and any subsequent publication decision.
This lane prepares a local correction and review evidence only.

No publication, pull request, formal approval, merge, deployment, manual CI,
settings change, runtime activation, operating namespace or database change,
live provider, browser or Docker execution is included. Do not copy private
consumer records, conversations, absolute scratch paths, raw logs or credentials
into this task, fixtures or evidence.

## Evidence and next action

The defect has been reproduced through the public shadow preview and directly
through the shared evaluator using invented inputs, including a verified
Node v24.15.0 run. Producer coordination clears the two declared runtime paths.
The coordinator reviewed this proposal and authorized local implementation.

Next: pin this canonical source, register and verify it, record the genuine start,
then implement and validate only the declared scope. Return the final diff and
exact local evidence identities for independent review before any publication.
