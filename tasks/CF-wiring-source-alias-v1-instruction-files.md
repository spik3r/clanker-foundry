---
id: CF-wiring-source-alias-v1
status: completed
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

## Observed implementation result

The helper now compares Claude memory with the actual canonical instructions
after its existing leaf checks. The installer separately compares that memory
with its planned repository source before any Stow call. Both use file identity,
not equal contents or a generic link-count restriction. No automatic repair or
instruction-source edit was added.

The focused regression scripts were frozen before production edits. On the old
scripts, canonical hardlink preflight returned 0 instead of 1, and the planned
source fixture reached the recording Stow seam instead of refusing. The same
regressions now pass; their SHA-256 values remain:

- Wiring tests: `b69c6d27ca4bc804c3653ac14e451bdd2ecac74c10d19cfdb96a9fef76767e98`.
- Stow tests: `d75c516471ea14984720b255423e01d4de947cafac20226db2abc9a36ac9b8ba`.

Observed local checks:

- Both focused suites passed with Bash 3.2.57.
- Separate Bash syntax checks passed for all four changed shell scripts.
- ShellCheck passed for those four scripts.
- Full `tests/test-install.sh` passed, including real GNU Stow installation and
  rerun controls against scratch homes, plus existing resolver tests.
- A real GNU Stow replay in an owned copied pack refused the planned-source
  alias before creating canonical or shared targets. Source bytes, inode, mode
  and mtime remained unchanged.
- `scripts/validate.sh` passed; optional `skills-ref` was not installed.
- Whitespace checks passed. The tracked instruction source still matches HEAD.

Pinned registration and the genuine start used Node 24.15.0 before the new tests
or code edits. This bounded implementation claim is complete; independent final
review, completion evidence, committed validation and PR publication remain next.
No real-home installation, merge or runtime activation is claimed.
