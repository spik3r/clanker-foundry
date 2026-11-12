---
id: CF-generation-plan-v1
status: in-progress
owner: Foundry implementation worker
repositories: [spik3r/clanker-foundry]
---

# Preview project generation without replacing outputs

## Implementation claim

Implement a `foundrygen plan` command using the same model-validation and
output-ownership preflight as `generate`. Return a versioned JSON report of
create, replace or unchanged actions and validated model identities in declared
output order. The command must not create project output directories or invoke
generation. Keep pin, path and ownership checks and empty stdout on failure.

Scope: modeling/foundrygen/run.go, modeling/foundrygen/plan_test.go,
modeling/foundrygen/README.md and this task. No PR publication, other generators,
application source, automatic approval or saved-plan execution.

## Validation

Run the foundrygen module tests and race tests, go vet, repository validator and
project-memory committed-diff validation. Exercise real generators with missing,
unchanged, drifted, invalid and unowned outputs, checking project bytes and mtimes.
