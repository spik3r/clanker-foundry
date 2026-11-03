# Partial orchestration sequence

This is a declared review view, not an executable state machine or a complete account of the evaluator.
REFUSED is an outcome marker, not a runtime state. Every invented refusal leaves its supplied execution unchanged; no receipt or effect list is returned.

COMPLETED does not release resources: acceptance retires the attempt and emits only a revocation intent. Explicit CLEANUP_CONFIRMED releases the retired attempt and resources. Nonfinal acceptance advances to PENDING while cleanup is still pending.

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
8. **RECORD_VALIDATION**: VALIDATING → VALIDATING (return).
   - **TR-VALIDATION-PASS**: An independent registered validator records a current PASS for the submitted subject without changing its validation deadline. Fixtures: validation_pass.
9. **RECORD_VALIDATION**: VALIDATING → REFUSED (throw).
   - **TR-VALIDATION-SAME-WORKER**: Validation by the assigned worker refuses even when its check registration is supplied. Fixtures: validation_same_worker.
   - **TR-VALIDATION-UNTRUSTED-CHECK**: Validation from a validator without the matching check registration refuses. Fixtures: validation_untrusted_check.
   - **TR-VALIDATION-WRONG-SUBJECT**: Validation bound to a different submitted subject refuses. Fixtures: validation_wrong_subject.
   - **TR-VALIDATION-PRESUBMIT-TIME**: Validation dated before the recorded submission refuses. Fixtures: validation_presubmit_time.
   - **TR-VALIDATION-DUPLICATE**: A second receipt for the same check refuses. Fixtures: validation_duplicate.
10. **RECORD_APPROVAL**: VALIDATING → VALIDATING (return).
   - **TR-APPROVAL-APPROVE**: A registered approver records an unexpired approval bound to the current execution and subject. Fixtures: approval_approve.
11. **RECORD_APPROVAL**: VALIDATING → REFUSED (throw).
   - **TR-APPROVAL-WRONG-SUBJECT**: An approval bound to a different submitted subject refuses. Fixtures: approval_wrong_subject.
   - **TR-APPROVAL-EXPIRED**: Recording an approval that expires at the command time refuses. Fixtures: approval_expired.
12. **ACCEPT_STAGE**: VALIDATING → COMPLETED (return).
   - **TR-ACCEPT-FINAL-COMPLETED**: Owner acceptance completes the final stage and returns a revocation intent; the retired attempt, pending cleanup and held resources remain until confirmation. Fixtures: accept_final_completed.
13. **ACCEPT_STAGE**: VALIDATING → PENDING (return).
   - **TR-ACCEPT-NONFINAL-PENDING**: Owner acceptance of a nonfinal stage advances to PENDING, retains the accepted subject and pending cleanup, and clears result, validations and approvals. Fixtures: accept_nonfinal_pending.
14. **ACCEPT_STAGE**: VALIDATING → REFUSED (throw).
   - **TR-ACCEPT-MISSING-CHECK**: Acceptance without the required passing check refuses. Fixtures: accept_missing_check.
   - **TR-ACCEPT-REVOKED-APPROVAL**: Acceptance with only a revoked approval refuses. Fixtures: accept_revoked_approval.
   - **TR-ACCEPT-UNREGISTERED-APPROVER**: Acceptance without a currently registered approver refuses. Fixtures: accept_unregistered_approver.
   - **TR-ACCEPT-UNAUTHORIZED-OWNER**: Stage acceptance by an actor other than the accountable owner refuses. Fixtures: accept_unauthorized_owner.
15. **CLEANUP_CONFIRMED**: COMPLETED → COMPLETED (return).
   - **TR-CLEANUP-COMPLETED-RELEASED**: Owner confirmation with the retired attempt token releases resources and clears the attempt while preserving COMPLETED. Fixtures: cleanup_completed_released.
16. **CLEANUP_CONFIRMED**: COMPLETED → REFUSED (throw).
   - **TR-CLEANUP-STALE-FENCE**: Cleanup confirmation with the advanced live fence rather than the retired attempt token refuses. Fixtures: cleanup_stale_fence.
   - **TR-CLEANUP-NOT-PENDING**: A repeated cleanup confirmation after cleanup has cleared refuses. Fixtures: cleanup_not_pending.

Commands not modeled: REVOKE_APPROVAL, REQUEST_CHANGES, RETRYABLE_ERROR, TIMEOUT, RETRY_TIMER, BLOCK, RESOLVE, FAIL, CANCEL.

States not modeled: RETRY_WAIT, BLOCKED, FAILED, CANCELLED.

Partial entries do not claim all guards, outcomes, transitions, effects or ordering have been modeled.
