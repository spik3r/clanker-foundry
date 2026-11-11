---
id: CF-flexgo-tui-v1
status: in-progress
owner: Foundry Flexgo implementation
repositories: [spik3r/clanker-foundry]
---

# Compile a bounded record-browser model to Flexgo

## Implementation claim

The Foundry Flexgo implementation owns `modeling/tui/`, its dedicated CI workflow,
and a link from the current workflow guide. Build a strict semantic profile for a
read-only record browser, deterministic Go source generation using the existing
public responsive-browser proposal, and independent acceptance checks. The model
selects typed record fields, identity, summary and ordered details, labels, and
explicit responsive breakpoints. Applications supply records; models cannot run
commands, access files or inject source code. Preserve selection, focus, scrolling
and captive help across resize and data replacement.

The first result must compile and exercise a generated browser, including a model
change without hand-editing generated output. This is a bounded prototype, not a
general UI language or a released Flexgo component API. Do not change private
application sources, Flexgo core, existing lifecycle profiles, or unrelated work.

## Checks

Run focused model/compiler tests, generated behavior acceptance tests, formatting,
Go vet, repository validation and committed project-memory validation. Keep
unrun terminal visual checks explicit. Publish a draft PR; do not merge or deploy.

## Coordinator closeout claim

Owner: Foundry maintenance coordinator, for this integration record only. The
implementation owner and claim above remain historical provenance. Reviewed base:
`6fbc4a6b3accd558188dd2444e47e29d21894211`. Local branch:
`fix/tui-reload-offsets-20261111`.

Record the already merged PR107 outcome and release only its completed bounded
implementation scope. Writable paths are this task record and new immutable
source/event files for `CF-flexgo-tui-v1`. Do not change code, tests, other task
records, previous snapshots/events or legacy ledger bytes in this closeout run.
Independently review this claim before committing its source, register that exact
source, and record a genuine start before changing status or adding the result.
Check current main, PR107 integration and visible competing work, then run pack,
whitespace and committed Project Memory validation. Any follow-on maintenance
needs a separate reviewed claim and start; this closeout grants no runtime,
provider, production or deployment authority.
