# Useful external skills

Install external skills selectively. This pack owns its engineering workflows; external
skills should fill a gap rather than duplicate an existing workflow.

## Recommended

| Need | Skill | Notes |
|---|---|---|
| Repository-grounded AppSec threat model | [OpenAI curated `security-threat-model`](https://github.com/openai/skills/tree/main/skills/.curated/security-threat-model) | Use when the user explicitly asks to model threats, abuse paths, trust boundaries, or mitigations. It is upstream-curated, so install it by name with `$skill-installer security-threat-model` rather than vendoring a copy here. |
| Concise response style | [Caveman](https://github.com/JuliusBrussee/caveman) | A communication layer, not an engineering workflow. Keep evidence, validation, risks, and blockers visible. |
| Requirement discovery and shared language | [Matt Pocock: `grill-with-docs`](https://github.com/mattpocock/skills) | Useful before a task has a usable scope. |
| Specs, tickets, prototypes, and test-first work | [Matt Pocock's skill collection](https://github.com/mattpocock/skills) | Select `to-spec`, `to-tickets`, `prototype`, or `tdd` only when their workflow fits the project. |

## Deliberate overlaps

Do not duplicate these without a repeated, demonstrated need:

- `diagnosing-bugs` overlaps `workflows/bug-investigation.md` and `engineer` repair mode.
- `code-review` overlaps `review-change` and `checklists/code-review.md`.
- `handoff` overlaps `checkpoint` and `context-offload`.
- `codebase-design` overlaps `architecture-review` and `safe-refactor`.

Read a candidate `SKILL.md` before installing it. Do not allow an external workflow to
override repository state, validation evidence, or explicit user constraints.
