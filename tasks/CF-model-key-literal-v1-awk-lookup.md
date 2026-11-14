---
id: CF-model-key-literal-v1
status: completed
owner: Codex model-key maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Compare model lookup keys as literal data

## Implementation claim

Owner: Codex model-key maintenance contributor, delegated by the coordinating
agent under the owner's continued maintenance request. Main at `3037fb8`,
current direction, prior resolver claims and active work were reviewed. Earlier
resolver fixes are complete; other current work does not own these paths.

An invented profile with tier text backslash-146-ast selects the fast mapping,
and ba-backslash-154-anced selects balanced, though neither literal key exists.
AWK interprets backslash escapes in the lookup key passed through -v. A normal
unknown tier correctly refuses. Model-map names and values are documented data;
lookup must not reinterpret an absent key into a different configured key.

Scope: `scripts/agent-model.sh`, `tests/test-agent-model.sh`, this task and new
immutable Project Memory files. Pass the requested key through a command-scoped
environment variable and read it with AWK ENVIRON, preserving literal bytes.
Add octal-alias refusal regressions and literal backslash-bearing model-value
controls. Keep existing tier grammar, defaults, map validation, local override
precedence, basename selection and CRLF behavior unchanged. Do not edit shared
model names or real local/user configuration, launch agents or contact providers.

Checks: prove the new alias regressions fail before repair; Bash syntax, full
resolver and model-map regression suites, full installer regressions,
repository validation, whitespace and committed Project Memory validation.
Use only invented scratch profiles/maps and isolated test homes. Commit, push
and open a PR for independent coordinator review and hosted checks; no merge.

## Integration result

The coordinating agent confirmed independent review of
`cee42b380ddd87785cf1ccc3faefd9940c9b20c2`; all ten hosted workflows passed on that exact head.
[PR #158](https://github.com/spik3r/clanker-foundry/pull/158) merged on 14 November 2026 as
`44db6e4fc01e0c542070c0c27102465cdba650ba`. Git confirms the reviewed head is the merge's second parent.
This bounded implementation claim is complete and reserves no further work.
