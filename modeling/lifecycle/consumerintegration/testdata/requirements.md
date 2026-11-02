# Invented consumer candidate review

This public-safe offline fixture models a simulated consumer candidate, not any
real consumer repository, production policy, work assignment or approval. Actor
and reviewer labels are supplied assertions, not authenticated identities.

- A candidate begins in Candidate and submits to Reviewing, assigning a missing
  owner to the submitting actor. A present owner must match that actor.
- Nonblank outcome, nonempty acceptance and exact command revision are required.
- In Reviewing, a distinct nonblank reviewer may accept into terminal Accepted
  or return into Candidate using the corresponding explicit review decision.
- Review task ID, revision, model digest, task-content digest and acting reviewer
  must all match. Rework advances revision and invalidates old review bindings.
- Every accepted transition advances revision exactly once. Refusals leave the
  input unchanged; invalid input returns no task. Revision overflow refuses.
- Revision 1 permits any nonblank submit actor. Revision 2 changes only
  R_submitActor and G_submitActor: submission requires the literal maker actor.
  Thus a guest submission succeeds only in revision 1; maker still succeeds.

## Origin of literal expectations

A separate assistant wrote the models and literal suite results from these rules
before any custom generated reducer, scenario runner or acceptance replay ran.
Existing strict model decoding and canonical hashing were used only to validate
IR and bind the frozen artifacts. Task-content hashes were calculated separately
from the documented fixed-order JSON contract, not obtained from a reducer.
This is separation between assistants, not independent-person verification or
proof that the authored rules are correct. No expected result may be refreshed
from observed execution. Review both model revisions and their literals.

The fixtures cover success, self-review refusal, terminal refusal, rework and stale
review, same-revision changed content, a foreign-model review, unknown commands,
uint64 overflow and the deliberately changed submit-actor rule. Suite and manifest
identities remain revision-specific; cross-wiring them must stop before replay.
