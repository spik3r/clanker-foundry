# Skill catalog

## Included skills

Short summaries below; each skill directory has its own `README.md` with detail
on modes, behaviour, output, and example invocations.

| Skill | Purpose | Default behaviour |
|---|---|---|
| [`engineer`](../skills/engineer/README.md) | End-to-end investigation, planning, implementation and validation | May edit in implementation or repair mode |
| [`review-change`](../skills/review-change/README.md) | Independent PR, branch, commit or working-tree review | Read-only |
| [`investigate-codebase`](../skills/investigate-codebase/README.md) | Trace an unfamiliar code path or subsystem | Read-only |
| [`architecture-review`](../skills/architecture-review/README.md) | Assess boundaries, coupling, data, reliability, security and evolution | Read-only unless asked for a plan |
| [`performance-investigation`](../skills/performance-investigation/README.md) | Measure and isolate performance bottlenecks | Evidence-first; edits only for scoped experiments |
| [`safe-refactor`](../skills/safe-refactor/README.md) | Improve structure while preserving observable behaviour | Small edits with continuous validation |
| [`adr`](../skills/adr/README.md) | Create or update an Architecture Decision Record | Writes documentation only |
| [`checkpoint`](../skills/checkpoint/README.md) | Create, refresh, or resume durable handoffs (`checkpoints/`) across sessions and agents | Writes documentation only |
| [`context-offload`](../skills/context-offload/README.md) | Create a compact, replaceable `context.md` handoff for one active task | Writes documentation only |
| [`task-management`](../skills/task-management/README.md) | Scope, track and close work in the existing task store | Writes task records within the requested scope |
| [`agent-coordination`](../skills/agent-coordination/README.md) | Check shared claims, avoid overlap and transfer bounded work | Coordinates ownership; does not launch workers |
| [`backlog-maintenance`](../skills/backlog-maintenance/README.md) | Reconcile stale, duplicate or completed tasks with evidence | Reviews stay read-only; upkeep edits the agreed records |
| [`project-memory`](../skills/project-memory/README.md) | Record immutable v2 task/run history and verify fresh committed-change evidence | Records local implementation evidence; keeps ownership in canonical task records |
| [`model-driven-design`](../skills/model-driven-design/README.md) | Turn an idea into a strict model proposal and independent acceptance handoff | Clarifies by default; authorized drafts and review preparation only |
| [`braindump-distiller`](../skills/braindump-distiller/README.md) | Turn unstructured ideas into a phased plan or interactive checklist | Read-only unless asked to save a plan |
| [`knowledge-note`](../skills/knowledge-note/README.md) | Create or update connected notes in a knowledge base or vault | Edits only the named vault |
| [`orchestrator`](../skills/orchestrator/README.md) | Coordinate scoped subagents, current plans, ownership, and independent review | Delegates only when the task benefits from orchestration |
| [`ui-verification`](../skills/ui-verification/README.md) | Verify visible browser behaviour and relevant repeated or interrupted flows | Verifies within the approved environment; reports browser checks separately from fallbacks |
| [`persistent-iteration`](../skills/persistent-iteration/README.md) | Continue bounded work or monitoring until the requested outcome or time boundary | Continues only within authorised scope |

## Suggested usage

Examples:

```text
Use engineer in implementation mode to add ...
Use engineer in validation mode on my current changes.
Use review-change to review this branch against main. Do not edit files.
Use investigate-codebase to trace how refunds flow from the API to Stripe.
Use architecture-review on the authentication and authorisation boundaries.
Use performance-investigation to diagnose the slow portfolio endpoint.
Use safe-refactor to split this service without changing behaviour.
Use adr to record our decision to use managed identity.
Use checkpoint to write a handoff before I switch models.
Use orchestrator to coordinate this feature across research, planning, implementation and review.
Use ui-verification to check the changed browser flows and narrow-screen layout.
Use persistent-iteration to watch this check run until it finishes.
```

`templates/prompts/` holds fill-in-the-blank versions of the most common requests.
It includes an [`explain-code`](../templates/prompts/explain-code.md) prompt for tracing
unfamiliar code without editing it. [`workflows/feature-delivery.md`](../workflows/feature-delivery.md)
shows how to run the four agent profiles as a feature delivery pipeline.
For reusable subagent profiles, see [agents/](../agents/README.md):
explorer-researcher, architect-planner, builder-developer, and independent
reviewer-verifier.

## Recommended setup

Install these globally:

- `engineer`
- `review-change`
- `investigate-codebase`
- `safe-refactor`
- `adr`
- `checkpoint`

Consider keeping these global but invoke them manually when needed:

- `architecture-review`
- `performance-investigation`
- `orchestrator`
- `ui-verification`
- `persistent-iteration`

Keep stack-specific workflows, commands and architecture rules in project-level skills or `AGENTS.md`.

For a machine-wide behaviour base shared by every project, see
[`global-agents/`](../global-agents/README.md).

## External skills

See [useful external skills](../docs/external-skills.md) for the official OpenAI curated
`security-threat-model`, Caveman, and complementary skills from
[mattpocock/skills](https://github.com/mattpocock/skills).

## Design notes

- Each skill is self-contained. No skill depends on another skill being installed.
- Review and investigation skills default to read-only.
- The descriptions include explicit trigger boundaries so agents are less likely to invoke the wrong workflow.
- Skills use portable instructions; optional helpers have their own prerequisites.
- The skills complement `AGENTS.md`: global facts and standing rules belong in `AGENTS.md`; repeatable procedures belong in skills and workflows; reusable output shapes belong in templates; verification lists belong in checklists.

