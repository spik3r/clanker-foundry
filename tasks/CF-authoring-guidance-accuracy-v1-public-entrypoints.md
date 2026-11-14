---
id: CF-authoring-guidance-accuracy-v1
status: completed
owner: Codex authoring guidance contributor
repositories: [spik3r/clanker-foundry]
---

# Correct public authoring guidance and current route pointers

## Implementation claim

Owner: Codex authoring guidance contributor, delegated by the coordinating
agent during the owner's continued maintenance request. Main at `7b5a2d6`,
current direction and active work were reviewed. The coordinator separately
owns enum uniqueness schema work; this claim excludes schemas and code.

The JSON validator registry and profile exclude non-string constants despite
supporting boolean const. The TUI cases command block omits its module working
directory and says Go is unnecessary beside a go-run command. The pipeline
introduction is a dated 11 November snapshot that still presents the historical
builder route beside generation without pointing readers to the owner's newer
generated-code-only decision.

Scope: `modeling/profiles.json`, `modeling/jsonvalidator/PROFILE.md`,
`modeling/tui/CASES.md`, `docs/model-driven-pipeline.md`, this task and new
immutable Project Memory files. Correct only those claims: retain boolean
constants, show module cwd and distinguish prebuilt validate from go run, and
add a dated update linking the accepted generator pipeline direction. Preserve
historical pipeline sections and the number-spelling note. No source, schema,
version, generated output or consumer changes, and no broader plan rewrite.

Checks: validate a boolean-constant model, run the documented TUI validate
command from its stated cwd, run existing registry and authoring schema checks,
repository validation, whitespace and committed Project Memory validation.
Push and open a PR for independent coordinator review; merging belongs to the
coordinator.

## Integration result

The coordinating agent reviewed `f243eec5f8282d48af83db0112ffa70d11a78309`;
all ten hosted workflows passed on that head. [PR #145](https://github.com/spik3r/clanker-foundry/pull/145)
merged on 14 November 2026 as `658a7a464760b83a93dcb3268cd2e1d1e7dcb020`.
This bounded implementation claim is complete and reserves no further work.
