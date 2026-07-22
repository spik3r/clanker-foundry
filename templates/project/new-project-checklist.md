# New project checklist

Work top to bottom. Skip items with a recorded reason, not silently.

## Foundations

- [ ] Repo created with a licence and `.gitignore` for the stack
- [ ] `README.md` from `templates/project/project-readme.md`
- [ ] `AGENTS.md` from `templates/project/project-agents.md`
- [ ] Tool-specific pointer files if the team uses them: `CLAUDE.md`, `GEMINI.md` (both point at `AGENTS.md` — one source of truth)
- [ ] Dependency pinning chosen (lockfile committed)

## Validation loop

- [ ] One command runs the tests; one command runs the linter
- [ ] CI runs both on every pull request
- [ ] A failing build blocks merge

## Change process

- [ ] PR template installed (`templates/pr/pull_request_template.md` → `.github/pull_request_template.md` or the host's equivalent)
- [ ] Branch protection on the default branch
- [ ] ADR directory created (`docs/adr/`) with ADR-000 recording the initial stack choice (`templates/adr/adr-template.md`)

## Security baseline

- [ ] Secrets never enter the repo; secret store chosen and documented
- [ ] Dependency scanning enabled (Dependabot, Renovate, or equivalent)
- [ ] `checklists/security.md` walked once before the first external deploy

## Operations

- [ ] Logging, metrics, and alerting decided before the first real user
- [ ] Rollback path known: <how a bad deploy is undone>
- [ ] Ownership recorded: team, on-call, escalation path
