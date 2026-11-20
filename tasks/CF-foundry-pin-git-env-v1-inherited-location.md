---
id: CF-foundry-pin-git-env-v1
status: in-progress
owner: opencode coordinator
repositories: [spik3r/clanker-foundry]
---

# Check the pinned checkout when Git location variables are inherited

## Implementation claim

The owner requested latest main and continued implementation in a separate
worktree. The coordinator reviewed main at `357ef62`, the current direction,
existing claims and a live empty open-PR list on 21 November 2026. Existing
unfinished claims retain their owners. This closes the finding queued in
CF-tuicases-output-v1: inherited `GIT_DIR`/`GIT_WORK_TREE` can redirect
`foundrygen` pin checks away from the checkout selected by `--foundry`.

Source inspection found that `session.git` runs Git with the caller's full
environment. Git's repository-location variables override `-C`, so the
revision, tracked-file, status and index checks can inspect a different
repository. A read-only probe reproduced it: with `GIT_DIR` and
`GIT_WORK_TREE` pointing at a second checkout, `git -C CLEAN status` reported
the other checkout's modeling file.

Scope: `modeling/foundrygen/exec.go`, a new
`modeling/foundrygen/pin_env_test.go`, `modeling/foundrygen/README.md`, this
task and new immutable Project Memory source/events. Run the pin's Git commands
without variables that select another repository, worktree, index or object
store. Keep every other inherited Git and Go setting unchanged.

Exclusions: generator behavior, the build environment, model or project file
formats, ignore and index-flag policy, consumer code and runtime work.

Checks: demonstrate the redirect against the original code; prove a clean
checkout still passes and that revision mismatch, modified, untracked, ignored
and index-flag refusals are unchanged; run gofmt, `go vet` and
`go test -race -count=1 ./...` in `modeling/foundrygen`; run
`scripts/validate.sh`, whitespace checks and Project Memory validation.
Publication is authorized; merge is outside this claim.
