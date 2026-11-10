---
id: CF-installer-special-files-v1
status: in-progress
owner: astra-foundry-installer-special-files
repositories: [spik3r/clanker-foundry]
---

# Reject special-file instruction targets before installation

## Implementation claim

Owner: `astra-foundry-installer-special-files`. This bounded lane follows the
coordinator's assignment of independent Foundry improvements. Main was verified
as `7778beb4038f081ca0f36415a286811e8627fac6`; open PRs 101 and 102 do not overlap
this scope. The completed installer adoption remains attributed to its original
owner. Its existing records, source pins and events must remain unchanged.

The wiring preflight accepts a FIFO as Claude's instruction target. The wiring
operation then blocks on it after creating a Codex link. Stow's canonical-file
migration check likewise blocks when comparing the source with a FIFO. Reject
existing nonregular instruction files before reads, writes or Stow calls.

Writable product scope:

- `global-agents/wire-global-agents.sh`
- `scripts/stow.sh`
- `tests/test-wire-preflight.sh` (new focused regression suite)
- `tests/test-stow-preflight.sh`
- `tests/test-install.sh` (only invoke the new focused suite)
- This canonical task at `tasks/CF-installer-special-files-v1-preflight.md`
- New immutable source/event files under
  `.project-memory/tasks/CF-installer-special-files-v1/`

Keep absent targets and regular-file migration/import behavior. Continue to
reject Claude symlinks and directories. Preserve accepted canonical symlink
handling and known backup conflict behavior. Do not read from or overwrite FIFO,
socket or device targets while deciding whether they are safe. This is an
existing-target preflight, not a race-free transaction or filesystem confinement.
No permissions, dependency, workflow, profile selector, setup-helper, runtime,
provider, client-launch, global installation, PR 101/102 state, merge or deployment
change is included. Existing denied actions and the shared heavy-test hold remain.

## Validation and delivery

Independently review this public-safe source, commit it, verify its pinned bytes,
register it and record a genuine v2 start before product edits. Inspect live
claims before edits and preserve every existing lifecycle record and legacy byte.

Use isolated temporary homes and invented file contents. Cover FIFO rejection,
no new links or Stow calls on preflight failure, nonregular canonical targets,
regular-file append/migration and rerun compatibility, and existing symlink and
directory rules. Add deterministic read/write seams or bounded subprocess
timeouts so a regression cannot hang CI. Run the new tests against the unchanged
baseline as a negative control, preserve all existing assertions, and run the
existing focused installer-ordering suite, shell syntax, pack validation and
whitespace checks. Run real GNU Stow acceptance only if already available; do not
install tools or treat a seam as real Stow acceptance. Keep heavy aggregate tests
on hold, and disclose ShellCheck, native-platform or hosted checks not run.

Review the exact final diff independently, record completion only with observed
checks, commit the new evidence with the final implementation, and validate the
full committed diff against the verified main. New publication may proceed only
within the current user authorization and as a draft. Verify remote ancestry and
tree and current exact-head CI separately before claiming hosted qualification.
