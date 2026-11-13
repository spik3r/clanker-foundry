# Authoring trial, 13 November 2026

A test of the step 2 completion rule in the
[generator pipeline plan](../../generator-pipeline-plan.md): can a fresh agent
session, given written briefs and the profile registry, produce a valid model
or a gap report without reading generator source?

## Setup

- Foundry revision: main at `8184f3c` for the documents the author read; the
  validators and the cases runner were built from `50b50b4`.
- Author: one agent session on the lowest-cost model tier available, started
  with no prior context.
- The author was told to start from the design skill's
  [profiles reference](../../../skills/model-driven-design/references/foundry-profiles.md)
  and could read only the registry, `modeling/README.md`, the contract, and the
  profile document, schema, examples and cases documents the registry names,
  plus the gap-report template. It was told not to read any Go, template or
  test file and not to run any command. This limit was an instruction; nothing
  enforced it and nobody audited which files it opened.
- [briefs.md](briefs.md) holds the five briefs. They were written by the
  coordinating session. Briefs 4 and 5 ask for things the coordinator knew no
  profile could express.

## What the author produced

The `briefN.*` files in this directory are its output, unedited. The cases
files still carry `UNKNOWN` where the model digest goes, because the author
could not run a tool to compute it.

## What the tools said

Run by the coordinating session after the author finished. Each model was
checked once, with no retry and no correction.

| Brief | Profile chosen | `validate` | Cases | Gap reports |
| --- | --- | --- | --- | --- |
| 1. Branch browser | `record-browser/v1` | accepted | 5 of 5 passed | none |
| 2. Expense approval | `process-diagram/v1` | accepted | profile has none | the amount condition on a route |
| 3. Build dashboard | `dashboard/v1` | accepted | profile has none | stacked bars, if "split" means stacked |
| 4. Manifest check | `json-boolean/v1` | accepted | profile has none | string enum; forbidding extra properties |
| 5. Service list | `record-browser/v1` | accepted | 5 of 5 passed | `/` filter; `r` reload from a service |

For briefs 1 and 5 the coordinator replaced `UNKNOWN` with the digest that
`tuic validate` printed, then ran `tuicases check`. Nothing else in the cases
was changed.

The coordinator generated and looked at the expense diagram and the dashboard.
The diagram shows the process as briefed. Nobody compared any output with a
second, independently written set of expectations.

## What this does and does not show

- The registry, profile documents, schemas and examples were enough for this
  author to write five accepted models and ten passing cases at the first
  attempt.
- Where a brief went outside a profile, the author wrote a gap report and left
  the requirement out of the model. In brief 4 it omitted `kind` rather than
  weakening it to any string.
- This is one run of five briefs by one author. It is not a measure of how
  often an agent gets this right, and the briefs were written by someone who
  knew the profiles.
- The author wrote both the model and its cases for briefs 1 and 5. That is
  not independent expectation authorship, which the design skill asks for.
- An accepted model is well formed. Whether each model is the right design for
  its brief was judged only by the coordinator reading it.

## Document problems the author reported

From its notes files:

- The record browser profile does not say plainly whether falling below the
  wide width alone, with enough height, gives one pane.
- A cases file needs the model digest, which an author without tools cannot
  compute. The cases document should say what to write in that situation.
