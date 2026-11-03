# Intake and requirements review worksheet

Copy this worksheet for a particular review. Fill it from the actual source bytes
and the existing intake and trace reports. It is a human note, not a generated
report or an approval. Do not infer decision-to-requirement links; identify them
as explicit reviewer statements. Record uncertainty and omitted material.

## Inputs and identity

| Input or identity | Value | How obtained / limits |
|---|---|---|
| Intake request SHA-256, exact bytes | | |
| Requirements source SHA-256, exact bytes | | |
| Requirements normalized digest from trace | | Not the source-byte hash. |
| Model source SHA-256, exact bytes | | |
| Canonical model digest from trace | | Not the source-byte hash. |
| Acceptance catalog SHA-256, exact bytes | | Not a scenario-suite identity. |
| Scenario suite source SHA-256, exact bytes, if supplied | | |
| Tool versions and command context | | |

For every intake artifact, copy its declaration, unverified author label,
observed hash or absence, and binding status. List the three dependency
assertions and their statuses separately. A declaration match is consistency,
not proof of origin, currentness or a passing check. The trace's normalized
requirements digest cannot be compared to the intake's exact source-byte hash.
Do not equate the trace acceptance catalog with the intake scenario suite.

## Decisions, questions and declared links

| Intake ID and text | Requirement ID and statement | Reviewer-declared relationship and reason | Uncertainty |
|---|---|---|---|
| | | | |

List every relevant unresolved intake question explicitly, even when no link is
known. Mark any decision or requirement without a reviewed relationship as
unlinked; an empty table is not a finding of completeness. These relationships
are human statements; neither tool derives or verifies their meaning.

## Trace and acceptance declarations

| Requirement ID | Model links | Acceptance case IDs and catalog-declared status | Trace gaps | Reviewer question |
|---|---|---|---|---|
| | | | | |

Include pending and unlinked cases and gaps. `executable` is a catalog label,
not a test result. The trace checks reference existence and kinds, not whether
any case actually verifies a requirement. Record unlinked model elements and
cases when relevant; do not treat a partial selection as full trace coverage.

## Checks and next decisions

| Check | Status: passed / failed / not run / unknown | Exact command, artifact identity and result, or reason not run |
|---|---|---|
| Intake report | | |
| Requirements trace | | |
| Scenario or independent acceptance execution | | |
| Human review of declared relationships | | |

List outstanding questions, owner and next action as applicable. Keep the exact
inputs and observations available to the authorized reviewer. Hashes identify
bytes; they do not authenticate them. Completion of this worksheet, a linked case,
or a matching hash does not constitute review, independent authorship,
approval, execution authority or proof of correctness.
