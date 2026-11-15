---
id: CF-wiring-source-alias-v1
status: in-progress
owner: Codex wiring source-alias contributor
repositories: [spik3r/clanker-foundry]
---

# Refuse instruction-source aliases before wiring

## Implementation claim

Owner: Codex wiring source-alias contributor, delegated by the coordinator
under the repository owner's continued reviewed-maintenance request. Reviewed
main: `667fe7d8ebef15ae7a4a7b3371b004e0c7cff370`. No PR or active Project Memory
frontier competes with this scope. Earlier installer and parent-preflight runs
are complete and merged in PRs 103 and 150; do not reopen or edit those records.

Owned fixtures reproduced the wiring helper accepting a regular Claude memory
file that is a hardlink to canonical instructions. Wiring then appends the
canonical import into that same file and exits successfully. With a Stow-style
canonical symlink, this changes the owned repository source. A real GNU Stow
fixture also reproduced the problem when the canonical path was initially
absent: the planned source became canonical only after Stow ran.

Add two narrow identity checks. The shared wiring preflight must refuse Claude
memory that is the same file as the existing canonical instructions, before
mkdir, content tools or client links. Stow must also refuse Claude memory that
is the same file as its planned repository instruction source, before any Stow
call. Follow canonical aliases for this identity check without banning unrelated
hardlinks, directory aliases or supported canonical symlinks.

Refuse without automatic unlinking, copying or repair. Preserve independent
regular memory, equal-but-distinct files, append/newline/rerun behavior, existing
target and parent refusals, migration and backup rules. These are ordinary
preflight observations, not a transaction or a concurrent-replacement guard.

Scope: `global-agents/wire-global-agents.sh`, `scripts/stow.sh`,
`tests/test-wire-preflight.sh`, `tests/test-stow-preflight.sh`,
`global-agents/README.md`, this task and new immutable Project Memory files.
No instruction-source edits, real-home installation, generator, dependency,
workflow, consumer, runtime or historical evidence changes. Add no CLI flag.

## Validation

Freeze regressions and show failures on unchanged scripts before repairing
them. Use only owned scratch homes and copied fixture repositories. Never
hardlink a test target to the real tracked `global-agents/AGENTS.md`.

Test both direct check and execution with regular canonical files and canonical
symlinks, including already-present imports. Require refusal, unchanged source
bytes/mode/mtime and original links, no content-tool calls and no client links.
For Stow, cover an absent canonical path and a distinct existing canonical copy
when Claude aliases the owned planned source. Require zero Stow calls and no
new shared targets, canonical link or backup. Keep equal-but-distinct files and
ordinary installation controls successful.

Run Bash syntax, both focused suites, full installer regressions, pack validation,
ShellCheck if installed, whitespace and fresh committed Project Memory validation.
Use real GNU Stow acceptance only when already available, always against scratch
homes. Report optional or platform-dependent skips. Publish for independent
exact-head review and hosted checks; the coordinator alone owns merge.
