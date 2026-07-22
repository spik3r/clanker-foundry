# Starter AGENTS.md for a new project

Copy this file to the root of a new repository as `AGENTS.md` and fill in every
placeholder. Keep it short: stack, commands, and local conventions. Everything generic
(writing rules, correctness-first behaviour) belongs in the global base — see
`global-agents/` in this repo.

Delete this guidance section after copying.

---

# <Project name>

<One sentence: what this repo is.>

## Stack

- Language / runtime: <e.g. Go 1.24, Node 22>
- Frameworks: <e.g. Rails 8, Next.js 15>
- Data stores: <e.g. Postgres 16, Redis>
- Infrastructure: <e.g. AWS, Terraform, Kubernetes>

## Commands

```bash
<install deps>
<run the app locally>
<run tests>            # e.g. go test ./...
<run a single test>
<lint / typecheck>
<build>
```

An agent must be able to validate its own work with these commands. If a command needs
services running, say how to start them.

## Layout

- `<dir>/` — <what lives here>
- `<dir>/` — <what lives here>

## Conventions

- <naming, error handling, logging, test placement — only what differs from the language default>
- <branching / commit conventions>
- ADRs live in <path>; use one for decisions that are hard to reverse.

## Boundaries

- Do not edit: <generated code, vendored deps, migrations already applied>
- Never commit: <secrets, credentials, local config>
- <Anything the agent must ask before doing, e.g. force-pushing, schema changes>
