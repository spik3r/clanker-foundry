---
name: braindump-distiller
description: Distil an unstructured brain-dump, pasted requirements, transcript, or collection of ideas into a clear, actionable plan. Use when the user asks to organise thoughts, make a plan, break work into a checklist, decide where to start, or turn scattered material into phased tasks and open questions. Produce a concise Markdown plan by default, or an interactive HTML checklist when requested.
---

# Brain-dump Distiller

## Purpose

Turn scattered thoughts into one outcome-focused plan with small tasks, explicit
assumptions, and questions that reduce risk before work begins.

## Use when

- the user supplies a wall of text, fragments, transcript, screenshots, or rough spec;
- they need an actionable plan, checklist, prioritisation, or first step;
- ambiguity should be turned into explicit decisions rather than hidden guesses.

## Do not use when

- the task already has a clear, scoped implementation request: use `engineer`;
- the user needs a durable technical decision record: use `adr`;
- the task is to investigate repository behaviour: use `investigate-codebase`.

## Default access and modes

Read-only unless the user asks to save a plan. **Markdown** is the default output.
**Interactive HTML** copies `assets/plan_template.html` and replaces only its `PLAN`
object. Save generated files in a user-approved location; if none is given, use
`~/plans/YYYY-MM-DD-<slug>.html` and do not overwrite an existing plan.

## Workflow

1. Read all supplied material, including images. State when an implied attachment is
   unavailable rather than inventing its content.
2. Identify one primary goal. Put secondary goals, constraints, dependencies, risks,
   and assumptions in considerations.
3. Order two to five phases to resolve decisions and unknowns before irreversible or
   polishing work. Make each task independently completable in one sitting.
4. Put questions that affect scope or success before the plan; attach local questions
   to the task they affect. Do not stall on non-blocking uncertainty.
5. Check that the plan has a clear first action, a proportionate scope, and no filler.

## Validation

Check that each task advances the stated goal, phase order reduces risk, assumptions
are explicit, and no claim about attached material lacks evidence. For HTML output,
preserve the template rendering code and validate the edited `PLAN` object is valid
JavaScript.

## Output contract

Return:

- **Summary** — primary goal and first action;
- **Plan** — considerations, phased tasks, and questions;
- **Risks** — assumptions, dependencies, and blockers;
- **Next action** — the smallest useful start.

For saved output, state the file path and what remains unresolved.

## Stop or hand back

Ask one focused question only when missing information blocks a usable plan. Hand a
settled implementation task to `engineer`, architecture choices to `architecture-review`,
and durable decisions to `adr`.

## Safeguards

Distil rather than transcribe. Keep project management proportional to the work, do
not convert assumptions into requirements, and do not overwrite an existing plan.
