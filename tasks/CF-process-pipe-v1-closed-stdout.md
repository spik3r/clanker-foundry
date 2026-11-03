---
id: CF-process-pipe-v1
status: in-progress
owner: Foundry process CLI repair contributor
repositories: [spik3r/clanker-foundry]
---

# Return the declared process-view status on a closed pipe

## Implementation claim

Owner: Foundry process CLI repair contributor, coordinated by the repository
owner's lead agent. The owner requested repairing PR20, completing its visual
review and merging after review. This supersedes the earlier publication-only
scope for this repair; the coordinator owns visual inspection and integration.
Branch: codex/lifecycle-process-view, original head 51822945. Current main
at a981b3ed is incorporated by an ordinary merge before this claim.

Write only modeling/lifecycle/cmd/modelprocess/main.go and main_test.go, this
record and its new immutable Project Memory source/events. Preserve the original
process-view task, frozen expectations, renderer and all other product paths.
No render or merge action belongs to this builder assignment. The existing PR20
is the matching active lane; no competing implementation is created.

## Acceptance and validation

Reproduce SIGPIPE termination with an actual compiled binary and closed stdout
before repair. Match existing CLI policy by ignoring SIGPIPE in main so a failed
write produces exit 3 and the fixed diagnostic. Add compiled-binary regressions
for both json and svg with closed stdout and complete successful output. Run
Go 1.27.3 full race tests, vet, modelc check, pack validation, installer checks,
whitespace checks and fresh committed Project Memory validation. Obtain an
independent source review and publish to PR20 for the coordinator's visual review
and merge decision. Record observed results and any unavailable checks.
