---
id: CF-002
category: tooling
status: in-progress
owner: Foundry implementation
repositories: [spik3r/clanker-foundry]
depends_on: [CF-001]
---

# Preserve literal pointer validation without a false ShellCheck warning

## Implementation claim

Owner: Foundry implementation. The repository owner approved correcting SC2016
in the pack validator and publishing the reviewed fix to existing draft PR7.

Scope: scripts/validate.sh, this task record, and its immutable v2 source/event
metadata. The backticks in the GEMINI.md pointer are Markdown characters and must
remain literal. Make the smallest targeted correction with an explanation; do not
enable command substitution, broadly suppress diagnostics, alter workflows or
change unrelated validation behavior. No merge, release or consumer rollout.

## Acceptance and validation

- Keep the existing pointer match and failure behavior.
- Document any intentional line-scoped ShellCheck exception.
- Check positive and negative pointer fixtures and prove backticks do not execute.
- Run shell syntax, pack validation, whitespace and existing memory regressions.
- Independently review the diff, publish to the existing draft and verify fresh
  CI, including the ShellCheck job and installer regressions.
- Record actual results and any local tool availability limits without treating
  unrun checks as passed.
