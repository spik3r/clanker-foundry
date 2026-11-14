# Second consumer change by another operator

A protocol for milestone 3 in the [current direction](../../current-direction.md):
a change to the consumer made by someone other than the session that built the
generators. Nothing here has been run yet. Fill in
[result-template.md](result-template.md) once per task.

## Why

Every consumer change so far was made or supervised by the session that wrote
the generators. That session knew where everything was. This trial asks
whether the documents are enough for someone who does not, and how long it
takes. No effort baseline was recorded before the first trial, so the timings
here are the first ones.

## Operator

A person, or an agent session started without this repository's history,
who did not write the generators. Record which. The owner chooses who.

The operator may read the consumer's `design/README.md`, its `foundry.json`
and models, the Foundry [modeling index](../../../modeling/README.md), the
profile and cases documents linked from the
[profile registry](../../../modeling/profiles.json), and the generator README
for the profile in use. Note it in the record if they read anything else.

Nobody from the session that built the tools answers questions during the
task. If the operator is stuck for 15 minutes, they record where and stop
that task.

## Setup

1. A clean checkout of the consumer at its current main. Record the commit.
2. A checkout of Foundry at exactly the revision in the consumer's
   `foundry.json`.
3. Go 1.27.4 or later, and Node 24 if the operator wants the schema checks.

## Tasks

Do task A first. Each task is one pull request in the consumer.

### Task A: model-only change

On the correlated leads screen (`design/correlated-leads-browser.json`):
show the next step first in the detail pane, and stop matching the filter on
the evidence text.

Expected: only the model and the regenerated output change. If the operator
finds they must change hand-written code, that is a finding, not a failure.

### Task B: change that crosses the seam

On the API catalog screen (`design/api-catalog-browser.json`): add a field
that is already in the saved records but not shown today, show it in the
details, and make the filter match it.

Expected: the model changes, the output is regenerated, and the hand-written
adapter that fills records gains one field. Nothing under
`internal/tui/gen/` is edited by hand.

## Record for each task

- Start and end time, and time spent stuck.
- Every file read, in order, and which document answered each question.
- Every command run, and each failure with its message.
- Each place the documents were wrong, missing or misleading.
- Lines changed by hand, split into model, adapter, tests and other.
- Whether `foundrygen check` and the consumer's tests passed before merge.
- Whether the operator looked at the changed screen in a real terminal.

## What counts

- **Done:** the change merged with passing checks, and nothing generated was
  edited by hand.
- **Stopped:** the operator was stuck for 15 minutes, or needed help from the
  tool authors. Record where.
- **Gap:** the requirement could not be expressed in the model. Write a
  [gap report](../../../skills/model-driven-design/assets/gap-report.md).

Whether the measured effort is worth it is the owner's decision. This
protocol does not set a threshold.

## Terminal check already done

On 14 November 2026 the coordinating session ran the consumer's
`mcp-inspect --tui` binary in a 110x30 `tmux` terminal with a synthetic
inventory file and captured each screen. This is not a person looking at a
terminal: it checks layout, keys and colours as text.

- Worked: selection, pane switch, help capturing keys, filter open, typing,
  `q` typed as text in the filter, backspace, keep, cancel with selection
  kept, no matches, digit keys ignored, quit with exit 0. The page background
  was the application's own colour and the generator's default did not
  appear.
- Found: at 60x20 the footer cut off the filter and quit hints; at 20x5 the
  page showed only clipped chrome with no resize message. A fix is in
  progress in the consumer.
- Note: a terminal reads `esc` followed within a few milliseconds by another
  key as one alt-key press. A scripted run must pause after `esc`; a person
  typing does not hit this.
