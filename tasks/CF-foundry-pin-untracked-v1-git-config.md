---
id: CF-foundry-pin-untracked-v1
status: completed
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

## Integration result

The coordinating agent reviewed `ca02d73d52ef8c9e00d6c203f246f2f134563cff`;
all ten hosted workflows passed on that head. [PR #140](https://github.com/spik3r/clanker-foundry/pull/140)
merged on 14 November 2026 as `365caf35e4bca0aa1ee759ebe2b7104f8e42aa8a`.
This bounded implementation claim is complete and reserves no further work.
