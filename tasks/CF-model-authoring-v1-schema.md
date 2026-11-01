---
id: CF-model-authoring-v1
status: in-progress
owner: Codex authoring contract contributor
repositories: [spik3r/clanker-foundry]
---

# Lifecycle model authoring contract

## Implementation claim

Implement an offline, versioned JSON Schema authoring aid for the existing
`lifecycle-json/v1` format and `lifecycle/v1` profile, with examples, independently
specified syntax and semantic cases, and automated checks against the existing Go
adapter. Scope: `modeling/lifecycle/authoring/`, a new authoring-only CI workflow,
this record, and its new Project Memory evidence. Do not change the model,
compiler, generated artifacts, scenario/diff, traceability, patch, diagnostics,
review-pack, qualification-exporter, or runtime implementation lanes. The schema
is an editor aid, not semantic authority or a claim of UML/SysML conformance.

The parent coordinated this task. Current main, published branches and open PRs
were inspected; no competing authoring schema claim was found. This prospective
local record is not itself a shared lock. Recheck before publication.

## Validation

Use pinned schema tooling for tests only, run the existing Go validator on the same
cases, exercise malformed and semantically invalid examples, validate deterministic
files and document schema limitations, then run repository and lifecycle checks.
Record actual outcomes and leave review and merge to the orchestrator. No deployment
or agent activation.
