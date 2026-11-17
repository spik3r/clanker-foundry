---
id: CF-solid-render-acceptance-v1
status: completed
owner: Codex Solid render acceptance contributor
repositories: [spik3r/clanker-foundry]
---

# Compile and render generated Solid dashboards in CI

## Implementation claim

The repository owner approved a Foundry-only qualification check for the
generated `dashboard/v1` Solid target. The existing JSX parser test is optional
and skipped when esbuild is absent; no CI job currently compiles or renders the
generated component with Solid.

Scope: add a self-contained verification package under
`modeling/diagram/verify/solid/` with a pinned npm lockfile and Node 24 test
runner; generate fresh Solid dashboard output from repository-owned models,
compile it with the Solid JSX compiler, and render it through Solid's server
renderer. Check literal expected stat, table, chart and hostile-text behavior,
including a negative control showing the acceptance assertions fail when the
expected output is wrong. Wire the required check into the diagram generator CI
workflow and update `modeling/diagram/DASHBOARD.md` to state what this check
does and does not establish. Also add this task and its immutable Project Memory
source/events.

Keep generated model behavior and emitted files unchanged. Do not read or copy
consumer application code or data. Do not claim browser, visual, human or
consumer acceptance from server-rendered output.

Checks: the Solid verification package tests; `go test -race -count=1 ./...`
and `go vet ./...` in `modeling/diagram`; `scripts/validate.sh`; whitespace
checks; and committed Project Memory validation.
