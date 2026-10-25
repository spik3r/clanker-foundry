# Cloudflare `security-audit` skill

External reference. This pack does not vendor, copy or wrap it.

- Source: <https://github.com/cloudflare/security-audit-skill> (MIT licence)
- Skill directory upstream: `skills/security-audit/`
- Facts below come from the upstream README and `SKILL.md` as read on 2026-10-25.
  Upstream may change; read its current `SKILL.md` before relying on this page.
  Lines marked **Suggestion** are this repo's opinion, not upstream behaviour.

## What it does

It turns a coding agent into a source-code security auditor. In full audit mode the
parent agent coordinates isolated sub-agents through six phases:

1. Reconnaissance: architecture, trust boundaries, input surfaces, initial coverage ledger.
2. Coverage-led hunting: isolated hunters take ledger units; critics look for gaps.
3. Candidate validation: a fresh verifier tries to disprove each candidate.
4. Structured output: `confirmed`, `needs_validation` and `rejected` records in
   `findings.json`, checked against `report-schema.json`.
5. Independent record verification: fresh agents re-check final source claims.
6. Reporting: `REPORT.md`, `FINDINGS-DETAIL.md`, `NEEDS-VALIDATION.md`.

Two zero-dependency Node validators (`validate-coverage-ledger.cjs`,
`validate-findings.cjs`) check the ledger and findings files.

Upstream design rules worth knowing: only established boundary failures count as
confirmed; the agent that checks a finding is never the one that found it; severity
is likelihood times impact; defence-in-depth gaps are hardening notes, not
vulnerabilities.

## Install

Upstream's documented route uses the `skills` CLI (needs Node.js):

```sh
# Project-level install, run inside the repo you want to audit
npx skills add https://github.com/cloudflare/security-audit-skill --skill security-audit

# User-level install, available in every project
npx skills add https://github.com/cloudflare/security-audit-skill --skill security-audit --global
```

`npx skills --help` lists agent-selection and non-interactive options.

Manual alternative (**Suggestion**), same symlink pattern this pack uses for its own
skills. Keep the clone outside any repo you audit:

```sh
git clone https://github.com/cloudflare/security-audit-skill ~/src/security-audit-skill
ln -s ~/src/security-audit-skill/skills/security-audit ~/.claude/skills/security-audit   # Claude Code
ln -s ~/src/security-audit-skill/skills/security-audit ~/.agents/skills/security-audit   # Codex
```

Update with `git -C ~/src/security-audit-skill pull`. Pin a commit if you want
repeatable audits (**Suggestion**).

Requirements (upstream):

- an agent whose model supports tool use and parallel sub-agents;
- Node.js, for the validators;
- an OS-enforced sandbox if the audit should build, test or run target code. Without
  one, the skill reads source only and reports blocked leads as `needs_validation`.

## Usage

Start the agent in, or pointed at, the target codebase. Example prompts from upstream:

- `security audit this codebase`
- `find security vulnerabilities in ./src`
- `do a security review, output to ~/audits/my-project`

Modes (upstream):

| Mode | Triggered by | Behaviour |
|---|---|---|
| Guidance (default) | Questions, focused reviews, triage | Uses relevant parts of the skill. Creates no output directories or audit files. |
| Full audit | Explicit audit or pen-test request, "full/comprehensive/end-to-end" review, or a request for report artifacts | Runs all six phases and writes files. |

If a request fits both, the skill asks one clarifying question.

Profiles and limits (upstream): `quick` is one hunter wave plus one critic pass with
partial coverage; `standard` is the full workflow; `deep` uses finer units, critic
waves until clean, and a second pass over covered units. Name paths or subsystems to
scope a run; everything else is recorded as `out_of_scope`, not covered. A maximum
agent count can be set; if it runs short the run is marked incomplete and the report
says so.

Output (upstream): by default `~/security-audit-skill/<repo-name>/run-<N>`, outside
the target. A directory inside the target is used only if you choose it and version
control ignores the whole directory. Repeated runs are additive: earlier ledgers and
findings steer gap-targeting and revalidation of changed source.

## Tips

Upstream-backed:

- Say which mode you want. A full audit writes files and spends many agent calls.
- Set profile, scope and agent budget in the prompt, for example
  `security audit ./services/auth, quick profile, at most 20 agents`.
- Expect partial coverage from `quick` and scoped runs. Only confirmed findings get
  a severity.
- Treat `needs_validation` as open leads, not clean results.
- Use dummy secrets and principals. Do not point it at deployed systems, shared
  infrastructure or production identities.

Suggestions for this pack's users:

- **Commit or stash first.** The run records the source ref and whether the worktree
  is dirty; a clean tree makes the report reproducible.
- **Keep audit output out of the repo.** Accept the default output directory, or add
  the chosen directory to `.gitignore` before the run. Audit reports can describe
  unfixed vulnerabilities; do not commit or publish them before triage.
- **Triage before fixing.** Read `REPORT.md` and `NEEDS-VALIDATION.md`, then fix
  confirmed items through `workflows/bug-investigation.md` or the `engineer` skill.
  Do not let the audit run edit code; it is a review pass.
- **Walk the findings against a checklist.** Cross-check confirmed items with
  [`checklists/security.md`](../checklists/security.md) and keep evidence per item.
- **Pair with threat modelling.** Run
  [`security-threat-model`](external-skills.md) first to get trust boundaries and
  abuse paths, then run the audit scoped to the riskiest boundary.
- **Re-run after fixes.** Additive runs revalidate changed source, which gives a
  cheap regression check on the same output directory.
- **Be careful with private code.** A full audit sends source to the model provider
  through your agent. Check that your policy allows this before auditing
  confidential repositories.
- **Do not install it as a dependency of another skill.** Pack skills must be
  self-contained, so reference this one from docs and prompts only.

## Overlap with this pack

| This pack | Relation |
|---|---|
| `review-change`, `checklists/code-review.md` | Diff-scoped review. The audit skill is whole-codebase and report-oriented. |
| `checklists/security.md` | Item-by-item verification list. Use it to check the audit's output, not replace it. |
| `security-threat-model` (external) | Design-level modelling. Complements source-level hunting. |
