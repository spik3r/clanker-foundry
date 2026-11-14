---
id: CF-validator-percent-path-v1
status: completed
owner: Codex validator maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Resolve percent-encoded relative link paths

## Implementation claim

The coordinator assigned this bounded repair after isolated fixtures showed
that the pack validator rejects existing percent-encoded destinations and can
accept an encoded-spelling decoy while the decoded destination is missing.
Reviewed main at `44db6e4`, current claims and open PRs #159 and #160; their
record closure and Project Memory repair do not overlap this scope.

Limit changes to scripts/validate.sh, its existing installer validator
regressions in tests/test-install.sh, a concise clarification in
docs/getting-started.md if needed, this record and new immutable evidence.
Decode valid percent triplets exactly once after splitting the raw fragment.
Keep plus signs, literal backslashes, encoded hash filename bytes and decoded
trailing newlines unchanged. Reject decoded NUL bytes. Leave malformed escapes
literal, preserving existing literal paths. Do not add dependencies or a generic
Markdown parser, or change external-link and reference-link handling.

Prove the regressions fail before repair. Cover present/missing destinations,
literal encoded decoys, percent/double encoding, Unicode, space, hash, lowercase
hex, plus, backslash, trailing newline, NUL and malformed-escape controls.
Run Bash 3.2 syntax, full installer and model-map regressions, repository
validator, whitespace and committed Project Memory coverage. Publish for
independent review; the coordinator owns merge.

## Integration result

The independent reviewer found no actionable issue at
`664700822d52c47a39933bc12b0d2e5656631ce5`; all ten hosted workflows passed
on that exact head. [PR #161](https://github.com/spik3r/clanker-foundry/pull/161)
merged on 14 November 2026 as `0f63234b1e687cff2dc91d40d41780e0751e0237`.
Git confirms the reviewed head is the merge's second parent. Project Memory has
one nonconflicting completed frontier. This bounded claim is complete and
reserves no further work.
