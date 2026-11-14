---
id: CF-wiring-parent-preflight-v1
status: completed
owner: Codex wiring maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Reject blocked wiring directories before Stow runs

## Implementation claim

Owner: Codex wiring maintenance contributor, delegated by the coordinator during
the owner's continued maintenance window. Base: origin/main at `ccad85b`.
The coordinator checked active tasks and assigned these paths exclusively.

The wiring helper's `--check` examines final instruction files but accepts a
regular file at a required client directory. In an isolated scratch setup with
a file at CODEX_HOME, preflight returned zero and the installer created 188
shared symlinks before failing at mkdir. All four client-directory blockers
passed the old preflight. This defeats the declared ordering for known wiring
conflicts; it is not a claim of transactional installation.

Check required canonical and client directory components before any writes.
Reject existing non-directories and dangling directory symlinks, including
blocked ancestors of a missing directory. Preserve directory symlinks that
resolve to real directories. Do not add an ancestor-symlink ban or promise to
preflight permissions, concurrent changes or other filesystem failures.

Scope: `global-agents/wire-global-agents.sh`,
`tests/test-wire-preflight.sh`, `tests/test-stow-preflight.sh`,
`docs/getting-started.md`, this task and new immutable Project Memory files.
Test four client parents, nested and canonical parent blockers, read-only
preflight, zero Stow calls on refusal, blocker preservation and valid directory
symlinks. Every installation test uses invented scratch HOME, CODEX_HOME and
XDG_CONFIG_HOME; never install into the real home.

Checks: prove regressions fail before the fix; Bash syntax, focused wiring and
installer preflight tests, full installer regression tests, repository validator,
whitespace and committed Project Memory validation. Commit, push and open a PR;
independent review and merge belong to the coordinator.

## Integration result

The coordinating agent reviewed `0c02fa0679e5d0ee47f5befb318ca30fa79bfd51`;
all ten hosted workflows passed on that head. [PR #150](https://github.com/spik3r/clanker-foundry/pull/150)
merged on 14 November 2026 as `96a271d66a155eac51d7cf7090aab68100f16f40`.
This bounded implementation claim is complete and reserves no further work.
