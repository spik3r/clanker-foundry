# Invented first-design brief: submit a draft

This is a fictional worked proposal, not a transcript of a real human decision or
an approved product. It exercises only a pure lifecycle function. A real user must
supply their own choices; copying this example does not answer their questions.

## Intent and bounded outcome

Allow a draft to be marked Ready when a named actor submits a nonblank outcome and
at least one acceptance criterion against the current revision. Nothing is saved,
scheduled, sent, deployed or authenticated. Ready ends this tiny example; it does
not mean the draft passed real review, was implemented, or is ready for release.

## Explicit example decisions

- D_scope: model only Draft -> Ready via submit, not the surrounding application.
- D_actor: a nonblank supplied actor is required. This string is an assertion,
  not authentication; no permission or role check is claimed.
- D_owner: submission preserves owner exactly, including nil. It does not assign
  ownership or require the actor to equal the owner.
- D_terminal: Ready is terminal; no edit, reopen or subsequent submit is modeled.
- D_failures: a refused command preserves task contents and revision. All failed
  modeled guards are reported, sorted by their IDs.

These choices are invented inputs to the example, not policies selected by a tool.

## Requirements for an independent expectation author

For runtime-valid task/command values, the sole declared transition T_submit goes
from Draft to Ready when command type is submit and all four conditions hold:

| Requirement | Guard identity | Required condition |
|---|---|---|
| R_actor | G_actor | Actor contains non-whitespace text. |
| R_criteria | G_criteria | Task acceptance contains at least one item. |
| R_outcome | G_outcome | Task outcome contains non-whitespace text. |
| R_revision | G_revision | Command expected revision equals task revision exactly. |

Successful submission changes only state to Ready and increments revision once.
Task ID, outcome, ordered acceptance and owner are preserved. Refusal returns the
unchanged task and sorted failing guard IDs. A submit in terminal Ready returns
ILLEGAL_TRANSITION with no transition ID. The profile's existing input validation
and revision-overflow rules still apply; this model does not override them.

Stable interface: model ID FirstDesignDraft, initial state Draft, terminal state
Ready, command submit, transition T_submit, four requirement/guard pairs above.
Example Task values may use task ID item-1, outcome Describe the change, acceptance
[Acceptance stated], owner null, actor author-1 and initial revision 0. Use distinct
suite/case/step IDs. Test positive submission, each missing guard input, multiple
failed guards and terminal refusal with full literal expected tasks/results.

## Unknowns and human gates

- Q_use: Who is authorized to accept this proposal for a real use, and what review
  evidence must that person see? Unanswered; blocks real integration/execution.
- Model meaning and literal expectations need separate independent review.
- Generated view syntax or table consistency is not visual inspection. A person
  must inspect the actual diagram/table against the requirements before accepting.
- Persistence, authentication, UI, concurrency and product integration remain
  outside this profile and outside this handoff.

No original human answer, visual approval, test execution, integration approval or
production authorization is asserted by this example.
