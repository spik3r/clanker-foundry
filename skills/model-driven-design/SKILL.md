---
name: model-driven-design
description: Turn an idea into a reviewable model proposal and an independent acceptance handoff. Use for first-design model-driven work with any registered Foundry profile (lifecycle/v1, record-browser/v1, json-boolean/v1, process-diagram/v1 and others in the checkout's registry); expose missing human choices and profile limits before generation. Not a promise to generate an arbitrary application or approval to execute it.
---

# Model-driven design

Connect the person's intent to a model that can be inspected, and keep intended
outcomes separate from what generated code later does. AI drafting is not
claimed to be deterministic; neither a valid model nor repeatable output proves
that the design is right.

## Choose the requested boundary

- **Clarify:** an idea alone calls for questions and a proposed brief, read-only.
- **Draft:** when asked to save a model proposal, write only the agreed artifacts
  in a named writable destination; preserve existing work and repository claims.
- **Prepare review:** when local validation/generation is authorized, use existing
  trusted tools to prepare evidence and draft views. Do not run candidate behavior
  or publish, integrate, deploy, or claim approval from this step.

A request can authorize several boundaries together. Do not ask again about
permissions already supplied, or treat a proposed future step as authorization.
Follow the target repository's actual instructions and lifecycle before edits.

## 1. Make the important choices visible

Capture the outcome, actors, inputs, successful and refused examples, constraints,
non-goals and decision owner. Separate supplied facts from assumptions and unanswered
questions. Use [the brief template](assets/design-brief.md) when a saved brief helps.

Ask the smallest consequential question first: for example, who may act, what is
preserved on refusal, whether terminal work can reopen, or what counts as success.
Do not invent authentication, ownership, retries, timing or approval policy to
fill a gap. Mark less consequential unknowns and continue work they do not block.
A missing answer stays missing; removing a question is not evidence it was resolved.

## 2. Check what the chosen model can express

Read the actual versioned profile, schema and semantic validator, not just a
sample diagram. State which requirements fit, which need surrounding engineering,
and which block the proposal. Do not turn a whole application into an apparently
complete small state machine by silently omitting its unsupported parts.

To find and choose among the registered Foundry profiles, use
[the Foundry profiles reference](references/foundry-profiles.md). On the generated
path your output is data: a model, expectation literals or a gap report, never
feature code. When a requirement does not fit any profile, the outcome for it is a
[gap report](assets/gap-report.md), not code or a prose workaround.

For Foundry's current lifecycle route, use
[the Foundry handoff reference](references/foundry-handoff.md). Locate an existing
trusted checkout explicitly; an installed skill does not imply the modeling module,
toolchain or other skills are installed. If unavailable, deliver the brief and
identify the missing tool context rather than inventing successful checks. Do not
clone, install dependencies or select another provider without applicable authority.

## 3. Draft two separate accounts of the intended behavior

One account is the model proposal: stable semantic IDs, states, transitions,
guards and ordered effects, with each requirement traced to its intended rule.
Preserve original source bytes and distinguish their hash from normalized model
identity. Do not edit generated implementation as the source of design truth.

The other account is literal acceptance expectations from the requirements.
Arrange a different author or independent reviewer with the brief, public profile
contract and necessary stable interface IDs—not generated implementation or its
observed results. Record who actually did what and the limits of independence.
A label, hash, separate file or second prompt alone does not prove independence.
If independent checking is unavailable, call the cases proposals and leave that
gate pending. Do not manufacture a reviewer or human decision.

Freeze reviewed expectation bytes before candidate behavior is run. A failure is
a discrepancy to investigate, not permission to regenerate expected values,
replace a model binding, weaken a guard or silently amend the requirement.
Substantive corrections need a stated reason and renewed independent review.

## 4. Prepare evidence without promoting it to approval

Validate original bytes with the authoritative strict model tool. JSON Schema is
an authoring aid, not a substitute. Check identities, suite structure/binding and
remaining questions separately; a tool may return a valid report that contains gaps.
Keep malformed input, stale evidence, mismatches, unrun checks and passed checks
separate. Do not infer success from an exit code without reading its meaning.

Generate the supported diagram/table from that same model in an isolated owned
workspace. Review it against the source and brief, not as a second source of truth.
Programmatic parsing is not visual inspection. If a view cannot be inspected through
an allowed method, retain the visual gate as pending; do not work around a denial
or substitute another artifact and claim the blocked review happened. A UML-style
picture does not establish UML, SysML, BPMN, UTP or fUML conformance.

Use [the handoff template](assets/review-handoff.md) to present the original brief,
model and expectation identities; decision/question status; view/code identities;
actual checks and unrun steps; mismatches; and the next named decision. Keep sensitive
input and reports private. Existing repository records remain authoritative.

## 5. Stop at the real handoff

Ask the responsible person to review meaning and unresolved choices, and obtain
any separate authorization needed for candidate execution/integration. An AI review,
a model-valid status, a complete trace, a matching hash, a no-change comparison,
an exported archive or a passing generated structural test supplies none of that.

For later authorized work, identify the actual generated package/runner binding
and separately frozen expectations. Do not send a custom model's suite to a
built-in example runner and treat its stale refusal as a behavioral result.
Report a concrete next action and owner, not an aggregate "ready" claim that hides
pending work. Implementation outside the supported profile remains ordinary
engineering with its own scope and tests.
