---
id: CF-foundry-pin-untracked-v1
status: in-progress
owner: Codex pin maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Refuse untracked generator source when Git hides untracked files

## Implementation claim

Owner: Codex pin maintenance contributor. The owner requested the next logical
independent task, implementation in a separate worktree, push and a pull request.
Main at `4a161d1` and open PR #139 were reviewed; #139 changes TUI hint IDs and
does not overlap this claim.

The pinned Foundry check runs `git status --porcelain -- modeling`. Git's
`status.showUntrackedFiles=no` setting suppresses untracked generator source,
allowing a checkout with extra build inputs to pass the clean-tree check.
Explicitly request all untracked files. Keep ignored-file policy, revision
matching and the development-only `--unpinned` escape unchanged.

Scope: `modeling/foundrygen/exec.go`, `modeling/foundrygen/run_test.go`,
`modeling/foundrygen/README.md`, this task and new immutable Project Memory
source/events. Add a regression with repository-local Git configuration,
covering both top-level and nested untracked modeling files and an untracked
file outside modeling. No consumer edits, other generators or runtime work.

Checks: demonstrate the regression fails before the fix; gofmt, go vet and
`go test -race -count=1 ./...` in modeling/foundrygen; repository validator;
whitespace and committed Project Memory validation. Publication is authorized;
merge and deployment are outside this claim.
