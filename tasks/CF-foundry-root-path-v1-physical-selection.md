---
id: CF-foundry-root-path-v1
status: completed
owner: Codex Foundry root-path contributor
repositories: [spik3r/clanker-foundry]
---

# Preserve the filesystem meaning of CLI root paths

## Implementation claim

Owner: Codex Foundry root-path contributor, delegated by the coordinator under
the repository owner's request for reviewed, tested and merged Foundry work.
Base: main at `5a476742e61e5a4980d3e4cefa36ee59aa421ff1`.

An owned fixture reproduced `foundrygen` selecting a different directory when
a CLI root contains a symlink followed by `..`. The root helper cleans the path
before consulting the filesystem. A separate CLI fixture confirmed that `plan`
loaded the lexical decoy's project document instead of the physically named one.
This establishes wrong-root selection, not demonstrated output corruption.

Resolve the original `--project` and `--foundry` root spellings before lexical
cleanup. Keep accepted terminal and ancestor root aliases, current-directory and
empty project-root behavior. Refuse missing intermediate components and
non-directory components instead of cleaning them into a different valid root.
Keep the existing restrictions on model and output paths within the selected
root. Resolution is an ordinary observation, not a lock or a hostile-race guard.

Scope: `modeling/foundrygen/run.go`, new
`modeling/foundrygen/root_path_test.go`, `modeling/foundrygen/README.md`, this
task record and new immutable Project Memory source/event files. No generator,
template, profile, version, runtime, installer or existing task-history changes.
No consumer changes or runtime activation.

The coordinator checked current PRs and Project Memory frontiers: no competing
work is active. The earlier generation-plan and path-alias changes are merged;
their records remain unchanged. This claim covers root selection only.

## Validation

Check actual directory identity for absolute and relative interior-alias paths,
ordinary aliases, spaced and Unicode paths, and current-directory spellings.
Exercise all three CLI verbs with distinct invalid project and registry markers
to prove which roots are read before any generator executes. Check refusal,
empty stdout and preserved fixture bytes. Include missing and file-component
controls with valid lexical decoys. Use an existing public generator fixture for
a successful direct-versus-aliased integration check.

Run full foundrygen race tests and vet, the repository validator, whitespace
checks and committed Project Memory validation. Record observed commands and
results, including any platform-dependent test skips. Publish a PR for independent
review; only the coordinating agent may merge after checking the final head.

## Observed implementation result

The helper now resolves the original root through the filesystem before making
it absolute. Empty project-root input retains current-directory behavior. New
fixtures first failed on the unchanged helper for physical identity, unresolved
components and actual project/registry selection. With the repair, the focused
race suite passed, including a real public JSON validator fixture. Direct and
aliased project/checkout spellings give the same plan, check and unchanged
generation results; internal model symlinks still refuse.

Observed locally with Go 1.27.4 and cached offline modules:

- `go test -race -count=1 ./...` in `modeling/foundrygen`: passed (51.479s).
- `go vet ./...` in that module: passed.
- `scripts/validate.sh`: passed; optional `skills-ref` was not installed.
- Whitespace checks: passed.

Project Memory registration and the genuine start used Node 24.15.0 and pinned
source `f8600a87400bd1d771d8be907bf60c5154f80bd6` before implementation edits.
The bounded implementation claim is complete. Independent final review,
completion evidence, committed-diff validation and PR publication remain next;
this record does not claim a merge or runtime activation.
