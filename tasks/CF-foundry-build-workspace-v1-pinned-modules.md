---
id: CF-foundry-build-workspace-v1
status: in-progress
owner: Codex pipeline maintenance coordinator
repositories: [spik3r/clanker-foundry]
---

# Build pinned generators outside ambient Go workspaces

## Implementation claim

The owner authorized several hours of bounded maintenance, independent subagent
review, publication and merging after checks. Main at `c87356a`, live Project
Memory frontiers and open PRs were reviewed before this claim. The filter Unicode
repair is independent; the output-root and class-layout repairs remain queued.

`foundrygen` builds generators with inherited environment and read-only module
mode. A parent `go.work` or explicit `GOWORK` can replace a dependency with source
outside the pinned checkout. A finite local probe built the same module twice:
with ambient workspace discovery it printed the replacement dependency's value;
with `GOWORK=off` it printed the dependency selected by the module's own go.mod.
Read-only module mode does not prevent that substitution.

Disable workspace mode for generator builds so each module's committed go.mod
and go.sum select its dependencies. Keep the caller's cache, proxy and toolchain
settings and the empty environment used to invoke built generators. This does
not claim that builds are sandboxed, offline or independent of the installed Go
toolchain. The `--unpinned` development flag still skips only Git pin checks.

Scope: `modeling/foundrygen/exec.go`, new `modeling/foundrygen/exec_test.go`,
`modeling/foundrygen/README.md`, this task and new immutable Project Memory files.
Tests use fixed local fixture modules and a local replacement dependency, with
both automatic parent-workspace discovery and explicit GOWORK. No network or
consumer edits; no generator or runtime changes.

Checks: prove regressions fail before the fix; gofmt, go vet and full race tests
in modeling/foundrygen; repository validation, whitespace and committed Project
Memory validation; independent final-commit review and hosted PR checks.
