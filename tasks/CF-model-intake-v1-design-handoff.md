---
id: CF-model-intake-v1
status: in-progress
owner: Codex design-intake contributor
repositories: [spik3r/clanker-foundry]
---

# Design intake and artifact identity handoff

## Implementation claim

On branch `codex/model-intake-v1` from current main, implement a bounded,
strict, offline intake record for original intent, explicit decisions, unresolved
questions and separately declared requirements, model and scenario-suite
identities. Compare declarations with caller-supplied exact bytes, reporting
missing declarations, missing observations and digest mismatches honestly.
Author labels and SHA-256 matches are assertions and consistency checks, never
proof of independence, execution, review, approval, identity or authority.

Scope: `modeling/lifecycle/intake/`, `modeling/lifecycle/cmd/modelintake/`,
this task and its own new immutable Project Memory source/events. No writes to
existing model, traceability, authoring schema, modelpatch, scenario, diagnostics,
reviewpack, qualification, export, workflowacceptance, runtime or workflow files.
The held stack contains an independent workflow acceptance test and a design
change exporter; this upstream record neither executes or evaluates suites nor
packages artifacts or supplies their execution evidence. The other evidence
handoff lane is not transferred. Main and visible PRs and claims were inspected;
this local claim is not a shared lock. Recheck before publication.

## Acceptance and handoff

Preserve original intent and explicit decisions/questions; use strict versioned
JSON, bounded fields and input, exact-key/null discipline, deterministic ordering,
separate declared and observed identity, and a no-write stdin CLI. Include
independently stated positive and negative cases for malformed input, duplicates,
missing and mismatched artifacts, exact-byte sensitivity and deterministic output.
Keep actual artifacts private and do not claim format validity or executed tests.
Run the pinned Go and repository checks, record genuine Project Memory evidence,
and prepare one draft for review if permitted. No deployment or activation.
