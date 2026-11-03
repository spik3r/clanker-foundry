# Partial orchestration sequence

This is a declared review view, not an executable state machine or a complete account of the evaluator.
REFUSED is an outcome marker, not a runtime state. Every invented refusal leaves its supplied execution unchanged; no receipt or effect list is returned.

1. **ADMIT**: NONE → PENDING (return).
   - **TR-ADMIT**: A valid invented read-only admission produces a PENDING record without effects. Fixtures: admit_read_only_pending.
2. **ASSIGN**: PENDING → REFUSED (throw).
   - **TR-DEPENDENCY**: Assignment refuses when the dependency assertion is not true. Fixtures: assign_dependencies_missing, assign_dependencies_false.
   - **TR-RESOURCE**: After the dependency guard passes, assignment refuses when the resource assertion is not true. Fixtures: assign_resources_missing, assign_resources_false.
3. **ASSIGN**: PENDING → ASSIGNED (return).
   - **TR-ASSIGN-SUCCESS**: A valid invented read-only assignment records the supplied worker attempt, advances the fence and returns a LAUNCH_ATTEMPT intent. Fixtures: assign_success.
4. **ACKNOWLEDGE**: ASSIGNED → EXECUTING (return).
   - **TR-ACK-SUCCESS**: The matching worker acknowledges its live attempt before the lease deadline without returning an effect. Fixtures: ack_success.
5. **ACKNOWLEDGE**: ASSIGNED → REFUSED (throw).
   - **TR-ACK-WRONG-WORKER**: Acknowledgement by a different worker refuses. Fixtures: ack_wrong_worker.
   - **TR-ACK-STALE-FENCE**: Acknowledgement with a stale fence refuses. Fixtures: ack_stale_fence.
   - **TR-ACK-EXPIRED-LEASE**: Acknowledgement at the exact lease deadline refuses. Fixtures: ack_expired_lease.
6. **SUBMIT**: EXECUTING → VALIDATING (return).
   - **TR-SUBMIT-SUCCESS**: The matching live worker submits a consistent successful artifact manifest; validation starts and a RUN_CHECKS intent names the stage checks. Fixtures: submit_success.
7. **SUBMIT**: EXECUTING → REFUSED (throw).
   - **TR-SUBMIT-WRONG-WORKER**: Submission by a different worker refuses. Fixtures: submit_wrong_worker.
   - **TR-SUBMIT-STALE-FENCE**: Submission with a stale fence refuses. Fixtures: submit_stale_fence.
   - **TR-SUBMIT-EXPIRED-LEASE**: Submission at the exact lease deadline refuses. Fixtures: submit_expired_lease.
   - **TR-SUBMIT-WRONG-REPOSITORY-SUBJECT**: Submission whose artifact names a different repository revision refuses. Fixtures: submit_wrong_repository_subject.
   - **TR-SUBMIT-WRONG-ARTIFACT-DIGEST**: Submission whose subject digest differs from its artifact manifest digest refuses. Fixtures: submit_wrong_artifact_digest.
   - **TR-SUBMIT-WRONG-SPEC-DIGEST**: Submission whose specification digest differs from the current specification refuses. Fixtures: submit_wrong_spec_digest.

Commands not modeled: RECORD_VALIDATION, RECORD_APPROVAL, REVOKE_APPROVAL, ACCEPT_STAGE, REQUEST_CHANGES, RETRYABLE_ERROR, TIMEOUT, RETRY_TIMER, BLOCK, RESOLVE, FAIL, CANCEL, CLEANUP_CONFIRMED.

States not modeled: RETRY_WAIT, BLOCKED, COMPLETED, FAILED, CANCELLED.

Partial entries do not claim all guards, outcomes, transitions, effects or ordering have been modeled.
