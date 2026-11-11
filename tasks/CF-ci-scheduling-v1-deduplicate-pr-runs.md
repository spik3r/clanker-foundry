---
id: CF-ci-scheduling-v1
status: review
owner: Foundry CI scheduling contributor
repositories: [spik3r/clanker-foundry]
---

# Avoid duplicate CI runs for pull request branches

## Implementation claim

Owner: Foundry CI scheduling contributor. The integration lead assigned this
bounded maintenance lane on `fix/ci-scheduling-20261111`, from verified main
`30756603f16f3e4fe9216c40389d6d0df024218a`. Fresh GitHub reads found only held
PR101 and PR110 open, neither covering these workflow headers. Existing Project
Memory frontiers were complete; no active CI scheduling overlap was found.
This local source does not publish a reservation or grant runtime authority.
Recheck main and ownership before publication. Do not modify or retry the held PRs.

Exact scope:

- `.github/workflows/json-validator-go.yml`
- `.github/workflows/model-authoring-contract.yml`
- `.github/workflows/model-lifecycle-go.yml`
- `.github/workflows/orchestration.yml`
- `.github/workflows/project-memory-conformance.yml`
- `.github/workflows/tui-model-go.yml`
- `.github/workflows/validate.yml`
- `tests/project-memory-workflow.test.mjs`
- `AGENTS.md`, `docs/getting-started.md` and `skills/project-memory/README.md`,
  only wording about when existing CI runs
- This task and its new immutable Project Memory source/event files

Keep each unfiltered `pull_request` trigger and its fork-PR coverage unchanged.
Limit push triggers to `main` and all tags, including slash-containing tag names.
Preserve each existing `workflow_dispatch` entrypoint; do not add one to
`project-memory-conformance`. Feature-branch pushes without an open PR will no
longer run CI automatically; opening or updating the PR runs the existing checks.
No branch-only push cohort is needed in addition to that PR cohort.

Add workflow-level concurrency: PR runs group by workflow and PR number, with
cancellation enabled only for `pull_request`. All non-PR groups include run ID
and run attempt, so separate main/tag/manual runs cannot replace pending or
running peers. Do not add a queueing policy. Preserve job names, complete job
bodies, permissions, tests and security steps byte-for-byte against base. Do not
add path filters, alter required checks or branch protection, skip checks, bump
actions/dependencies, change application code, invoke providers, activate runtime,
deploy, or change another task or any existing history.

Fresh branch inventory covered 110 branches in two pages: main is the default
branch, every branch was unprotected, and the ruleset list was empty. The main
branch API reported protection disabled and no required status checks. No release
branch or documented release-branch CI policy was found. This supports the
explicit main-and-all-tags scope, not authority to change repository rules.

Publish a qualified draft only after local validation. Fresh independent exact
patch review and hosted checks on the exact head remain separate integration
gates. The integration lead owns any Ready transition and merge.

## Frozen expectations and validation plan

Fresh run history for PR111 head `f8184a80b021350d4061c9119d3335e9eb38df1f`
contains fourteen successful workflows: seven push and seven pull-request runs.
The seven tracked workflow headers have unfiltered push and PR triggers and no
concurrency group. This is duplicate scheduling evidence, not measured savings.

Review this public-safe source separately before committing or registering it.
Create and fetch its actual GitHub source commit before registration, verify the
claim bytes, and record a genuine v2 start before test or workflow edits. Freeze
regression tests before changing workflow headers and observe failures on the
old headers. Keep the frozen tests unchanged while implementing the schedule.

Extend the existing dependency-free workflow regression suite to cover all seven
trigger/concurrency headers and scenario expectations. Preserve every existing
job, security, validation-shell and helper assertion. Test PR workflow/number
separation and same-PR supersession, main/tag/manual noncancellation and run/attempt
separation, main push acceptance, other-branch push exclusion, slash tags, fork
PRs and the existing manual-trigger inventory. Scenario simulations prove the
configured contract, not observed GitHub scheduler execution.

Validate YAML with a real parser using YAML 1.2-compatible scalar treatment;
compare all job bodies and permission declarations against base. Use task-local
review tooling only, with no maintained dependencies. Run Project Memory and
workflow regressions, frozen graph conformance, repository and pack-validator
regressions, shared-base regressions, the relevant existing Go lifecycle checks
that read AGENTS.md, whitespace checks, and full committed Project Memory
validation. Run additional existing suites where practical. Report unavailable
GNU Stow or optional skills-ref honestly; hosted CI remains a separate gate.
Do not claim live tag, manual, fork or supersession execution without observing it.

## Observed local result

The reviewed source was created as `dd1cb32d9940dcf382b94e28910ffaa262ece3b3`,
fetched and verified before registration and the genuine start. No claim-only
branch was pushed. The original workflow suite passed 10 tests. The expanded
suite was frozen before workflow edits at SHA-256
`f63b4a988543b7474fcb0d521007b3d08f0e5486136a1279dc716375c52a6316`:
28 scheduling cases then failed on the old headers; all 38 now pass unchanged.
The existing job, security, shell-command and helper assertions remain present.

All seven workflows now retain unfiltered PR coverage and their original manual
entrypoints, run push CI on main and all tags, and group only PR runs for
cancellation. Non-PR groups include run ID and attempt. Generic PR concurrency
also includes manual reruns of older commits; it does not order runs by newest
commit. Static scenario tests are not live scheduler evidence.

PyYAML 6.0.3 BaseLoader with duplicate-key refusal parsed all seven workflows.
All nine job bodies and all permission bytes match base. A separate parsed
trigger model passed 63 cases; nine disposable schedule mutations were caught.
Combined memory/workflow/conformance tests passed 262/262. Runtime tests passed
321/321 after exporting the task-local npm cache for nested npm commands; the
initial run failed only because nested npm pack used the unavailable default
cache. Authoring tests passed 2/2. Repository validation, 47 model-map cases,
92 path cases, shared-base regressions and installer/wiring preflights passed.
The preflight seams do not establish GNU Stow acceptance.

The current AGENTS schedule wording exposed a frozen historical builder-test
input dependency. The separately reviewed and sourced
[fixture prerequisite](CF-builder-fixture-inputs-v1-historical-instructions.md)
is included without changing its production rules or old hashes. Its focused
race tests pass on the final instruction bytes. Broad lifecycle race validation
was attempted with available Go 1.27.4 and refused exact Go 1.27.3 requirements;
Go vet and formatting passed, while modelc validate also refused that version
and the subsequent drift check was not reached. The unchanged hosted job must
supply exact Go 1.27.3 acceptance. GNU Stow, ShellCheck, actionlint and optional
skills-ref were unavailable locally and are not claimed as passed.

Fresh prepublication checks still found main
`30756603f16f3e4fe9216c40389d6d0df024218a` and only nonoverlapping held PR101 and
PR110. The optional branch-rules API route was unsupported by the connector;
it was not retried. Branch inventory, main protection details and rulesets were
read through supported routes. No repository settings changed.

Editing is finished. Stage only the scoped paths, record both genuine
completions, validate the full committed diff and publish a qualified draft.
Independent exact-source review and exact-head hosted CI remain separate gates.
No live tag, manual, fork or supersession run, runner-time saving, Ready
transition, merge, runtime activation or provider execution is claimed.
