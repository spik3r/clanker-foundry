# Partial orchestration sequence

This is a declared review view, not an executable state machine or a complete account of the evaluator.
REFUSED is an outcome marker, not a runtime state. In the invented refusal cases the input remains PENDING at revision 1.

1. **ADMIT**: NONE → PENDING (return).
   - **TR-ADMIT**: A valid invented read-only admission produces a PENDING record without effects. Fixtures: admit_read_only_pending.
2. **ASSIGN**: PENDING → REFUSED (throw).
   - **TR-DEPENDENCY**: Assignment refuses when the dependency assertion is not true. Fixtures: assign_dependencies_missing, assign_dependencies_false.
   - **TR-RESOURCE**: After the dependency guard passes, assignment refuses when the resource assertion is not true. Fixtures: assign_resources_missing, assign_resources_false.

Commands not modeled: ACKNOWLEDGE, SUBMIT, RECORD_VALIDATION, RECORD_APPROVAL, REVOKE_APPROVAL, ACCEPT_STAGE, REQUEST_CHANGES, RETRYABLE_ERROR, TIMEOUT, RETRY_TIMER, BLOCK, RESOLVE, FAIL, CANCEL, CLEANUP_CONFIRMED.

States not modeled: ASSIGNED, EXECUTING, VALIDATING, RETRY_WAIT, BLOCKED, COMPLETED, FAILED, CANCELLED.

Partial entries do not claim all guards, outcomes, transitions, effects or ordering have been modeled.
